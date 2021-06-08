// Copyright 2021 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package podwatcher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/drone-runners/drone-runner-kube/internal/docker/image"
	"github.com/sirupsen/logrus"
)

var (
	// ErrUnknownContainer is an error that wait functions return when an unregistered container name is provided.
	ErrUnknownContainer = errors.New("unknown container")

	// ErrPodTerminated is an error that wait functions return when the pod is already terminated.
	ErrPodTerminated = errors.New("pod is terminated")
)

// PodWatcher is used to monitor status of a Kubernetes pod and containers inside of it.
// Containers should be registered with a call to AddContainer and then the watcher is started with
// the Start method. There are several methods for waiting for a status of a container.
type PodWatcher struct {
	// containerMap holds all containers in the pod, for each it holds current image and its state.
	containerMap map[string]*containerInfo

	// state represents PodWatcher state and can be: "init", "started" or "done".
	// Only during the "init" new containers can be added.
	// Only during "started" new wait clients can be added.
	// A watcher in "done" state can't be used.
	state watcherState

	// stop channel is used to prevent adding new wait clients if the watcher is finished.
	stop chan struct{}

	// errDone is used to store an eventual error from container watcher.
	errDone error

	// clientCh is a channel through which new wait clients are added.
	clientCh chan *client

	// clientList is an array of wait clients that are currently waiting for an event.
	clientList []*client
}

type watcherState byte

const (
	stateInit watcherState = iota
	stateStarted
	stateDone
)

// AddContainer registers a container for state tracking.
func (pw *PodWatcher) AddContainer(idx int, id, image, placeholder string) {
	if pw.state != stateInit {
		panic("AddContainer must be called before Start")
	}

	if pw.containerMap == nil {
		pw.containerMap = make(map[string]*containerInfo)
	}

	pw.containerMap[id] = &containerInfo{
		idx:         idx,
		id:          id,
		image:       image,
		placeholder: placeholder,
	}
}

func (pw *PodWatcher) Start(ctx context.Context, w ContainerWatcher) {
	if pw.state != stateInit {
		panic("Start can be called only once")
	}

	pw.state = stateStarted
	pw.stop = make(chan struct{})    // stop channel, close the channel to terminate the PodWatcher
	pw.clientCh = make(chan *client) // a channel for accepting new wait clients

	logrus.Debugf("Pod=%s watcher started", w.Name())

	wg := &sync.WaitGroup{}
	wg.Add(3)

	// Listening container events related to the pod.
	chEvents := make(chan []containerCurrentStatus)
	go func() {
		defer wg.Done()
		pw.errDone = w.Watch(ctx, chEvents)
	}()

	// Periodic scanning of containers. This should help in case an event was missed.
	chPeriodic := make(chan []containerCurrentStatus)
	go func() {
		defer wg.Done()
		_ = w.PeriodicCheck(ctx, chPeriodic, pw.stop)
	}()

	go func() {
		defer func() {
			wg.Done()
			pw.terminate()
		}()

		for {
			select {
			case <-ctx.Done():
				pw.errDone = ctx.Err()
				return

			case containers, ok := <-chEvents:
				if !ok {
					return
				}

				pw.updateContainers(containers)

			case containers, ok := <-chPeriodic:
				if !ok {
					return
				}

				pw.updateContainers(containers)

			case cl := <-pw.clientCh: // a new client is waiting for a container status
				if cl.containerId == "" {
					// The client is not asking for a container status, but the status of the whole pod.
					// Put the client to the list of unresolved clients.
					pw.clientList = append(pw.clientList, cl)
					break
				}

				c, ok := pw.containerMap[cl.containerId]
				if !ok {
					// The client is asking for an unknown container.
					// Resolve the client with the ErrUnknownContainer error.
					cl.resolveCh <- ErrUnknownContainer
					break
				}

				// Try to resolve the client right now...
				if !_tryResolveWaitClient(cl, c) {
					// ... if can't, put the client to the list of unresolved clients.
					pw.clientList = append(pw.clientList, cl)
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		logrus.Debugf("Pod=%s watcher terminated", w.Name())
	}()
}

func (pw *PodWatcher) terminate() {
	// stop accepting new wait clients
	// and stop periodic container scanner
	close(pw.stop)

	pw.state = stateDone

	// Tell all existing wait clients that it's over...
	// If an error happened during the accepting pod events, fail all with that error
	pw.notifyClientsPodTerminated(pw.errDone)
}

// updateContainers examines all containers in a pod and if any changes are detected it executes
// the method notifyClientsContainerChange for each changed container.
func (pw *PodWatcher) updateContainers(containers []containerCurrentStatus) {
	for _, cs := range containers {
		c, ok := pw.containerMap[cs.id]
		if !ok {
			continue // unknown container
		}

		c.currentImage = cs.currentImage // update container's current image

		var (
			newState     containerState
			newStateInfo string
			exitCode     int32
		)

		if image.Match(c.currentImage, c.placeholder) {
			// If container's current image is a placeholder image,
			// then it's state is Pending, whatever the actual container state is in.
			newState = statePending
			newStateInfo = ""
		} else {
			newState = cs.currentState
			newStateInfo = cs.currentStateInfo
			exitCode = cs.exitCode
		}

		if c.currentState != newState || c.currentStateInfo != newStateInfo {
			c.currentState = newState
			c.currentStateInfo = newStateInfo
			c.exitCode = exitCode

			logrus.Tracef("-> Container state changed: idx=%02d name=%s image=%s -> %s (%s)",
				c.idx, c.id, c.currentImage, c.currentState, c.currentStateInfo)

			pw.notifyClientsContainerChange(c)
		}
	}
}

// _tryResolveWaitClient will resolve the client if the container state is greater or equal to the requested state.
// For example: A container in TERMINATED state will resolve all clients waiting for it to enter RUNNING state
// and all clients waiting for it to enter TERMINATED state.
func _tryResolveWaitClient(cl *client, c *containerInfo) bool {
	if cl.containerId != c.id {
		return false
	}

	if c.currentState >= cl.containerState {
		if cl.containerState == stateTerminated && c.exitCode != 0 {
			// tell the client that the container failed with an exit code != 0
			cl.resolveCh <- exitCodeError(int(c.exitCode))
		} else {
			// tell the client to proceed
			cl.resolveCh <- nil
		}

		return true
	}

	return false
}

// notifyClientsContainerChange resolves all wait clients that wait for a specific state of a container.
func (pw *PodWatcher) notifyClientsContainerChange(c *containerInfo) {
	for clIdx := 0; clIdx < len(pw.clientList); {
		cl := pw.clientList[clIdx]

		if _tryResolveWaitClient(cl, c) {
			//  remove the client from the list (order is not preserved)
			pw.clientList[clIdx] = pw.clientList[len(pw.clientList)-1]
			pw.clientList[len(pw.clientList)-1] = nil
			pw.clientList = pw.clientList[:len(pw.clientList)-1]
		} else {
			clIdx++
		}
	}
}

// notifyClientsPodTerminated resolves all wait clients
func (pw *PodWatcher) notifyClientsPodTerminated(err error) {
	for _, cl := range pw.clientList {
		if err != nil {
			cl.resolveCh <- err
		} else if cl.containerId == "" {
			cl.resolveCh <- nil
		} else {
			cl.resolveCh <- ErrPodTerminated
		}
	}

	pw.clientList = nil
}

func (pw *PodWatcher) waitForEvent(containerId string, state containerState) (err error) {
	ch := make(chan error)

	logrus.Tracef("Wait: state=%s for container=%s", state.String(), containerId)
	t := time.Now()

	defer func() {
		logrus.Tracef("Wait finished: state=%s for container=%s duration=%.2fs err=%v\n",
			state.String(), containerId, time.Since(t).Seconds(), err)
	}()

	select {
	case pw.clientCh <- &client{containerId: containerId, containerState: state, resolveCh: ch}:
		err = <-ch

	case <-pw.stop:
		if pw.errDone != nil {
			err = pw.errDone
		} else {
			err = ErrPodTerminated
		}
	}

	return
}

// WaitContainerStart waits until a container in the pod starts.
func (pw *PodWatcher) WaitContainerStart(containerId string) error {
	return pw.waitForEvent(containerId, stateRunning)
}

// WaitContainerTerminated waits until a container in the pod is terminated.
func (pw *PodWatcher) WaitContainerTerminated(containerId string) (int, error) {
	err := pw.waitForEvent(containerId, stateTerminated)
	if code, ok := err.(exitCodeError); ok { // exit codes != 0 are masked as a exitCodeError
		return int(code), nil
	}

	return 0, err
}

// WaitPodDeleted waits until the pod is deleted.
func (pw *PodWatcher) WaitPodDeleted() (err error) {
	return pw.waitForEvent("", stateTerminated /* the state is unimportant */)
}
