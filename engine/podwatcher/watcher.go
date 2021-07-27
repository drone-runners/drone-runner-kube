// Copyright 2021 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package podwatcher

import (
	"context"
	"sync"
	"time"

	"github.com/drone-runners/drone-runner-kube/internal/docker/image"

	"github.com/sirupsen/logrus"
)

// PodWatcher is used to monitor status of a Kubernetes pod and containers inside of it.
// It is started with the Start method. Prior to waiting for a container state change,
// the container should be registered with a call to AddContainer.
type PodWatcher struct {
	// podName holds name of the pod. It's used mainly for logging.
	podName string

	// containerMap holds all containers in the pod, for each it holds current image and its state.
	containerMap map[string]*containerInfo

	// state represents PodWatcher state and can be: "init", "started" or "done".
	state watcherState

	// stop channel is used to finish the PodWatcher.
	// It prevents adding new wait clients and resolves all that are waiting.
	stop chan struct{}

	// errDone is used to store an eventual error from container watcher.
	errDone error

	// containerRegCh is a channel through which new containers are registered.
	containerRegCh chan containerInfo

	// clientCh is a channel through which new wait clients are added.
	clientCh chan *waitClient

	// clientList is an array of wait clients that are currently waiting for an event.
	clientList []*waitClient
}

type watcherState byte

const (
	stateInit watcherState = iota
	stateStarted
	stateDone
)

func (pw *PodWatcher) Start(ctx context.Context, cw ContainerWatcher) {
	if pw.state != stateInit {
		panic("Start can be called only once")
	}

	podName := cw.Name()

	pw.podName = podName
	pw.state = stateStarted
	pw.stop = make(chan struct{}) // stop channel, close the channel to terminate the PodWatcher
	pw.containerRegCh = make(chan containerInfo)
	pw.clientCh = make(chan *waitClient) // a channel for accepting new wait clients

	errDone := make(chan error)

	wg := &sync.WaitGroup{}
	wg.Add(3)

	// Listening container events related to the pod.
	chEvents := make(chan []containerInfo)
	go func() {
		defer wg.Done()
		errDone <- cw.Watch(ctx, chEvents)
	}()

	// Periodic scanning of containers. This should help in case an event was missed.
	chPeriodic := make(chan []containerInfo)
	go func() {
		defer wg.Done()
		_ = cw.PeriodicCheck(ctx, chPeriodic, pw.stop)
	}()

	go func() {
		defer func() {
			close(pw.stop)
			pw.state = stateDone
			pw.notifyClientsPodTerminated(pw.errDone)

			wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				pw.errDone = ctx.Err()
				return

			case pw.errDone = <-errDone:
				return

			case containers := <-chEvents:
				pw.updateContainers(containers)

			case containers := <-chPeriodic:
				pw.updateContainers(containers)

			case c := <-pw.containerRegCh:
				if pw.containerMap == nil {
					pw.containerMap = make(map[string]*containerInfo)
				}

				pw.containerMap[c.id] = &c

			case cl := <-pw.clientCh: // a new waitClient is waiting for a container state
				if cl.containerId == "" {
					// The waitClient is not asking for a container status, but the status of the whole pod.
					// Put the waitClient to the list of unresolved clients.
					pw.clientList = append(pw.clientList, cl)
					break
				}

				c, ok := pw.containerMap[cl.containerId]
				if !ok {
					// The waitClient is asking for an unknown container.
					// Resolve the waitClient with the ErrUnknownContainer error.
					cl.resolveCh <- UnknownContainerError{container: cl.containerId}
					break
				}

				// Try to resolve the waitClient right now...
				if !_tryResolveWaitClient(cl, c, nil) {
					// ... if can't, put the waitClient to the list of unresolved clients.
					pw.clientList = append(pw.clientList, cl)
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		//...
	}()
}

// updateContainers examines all containers in a pod and if any changes are detected it executes
// the method notifyClientsContainerChange for each changed container.
func (pw *PodWatcher) updateContainers(containers []containerInfo) {
	for _, cs := range containers {
		c, ok := pw.containerMap[cs.id]
		if !ok {
			continue // unknown container
		}

		if c.image == cs.image && c.state == cs.state && c.stateInfo == cs.stateInfo {
			continue
		}

		if cs.image == c.placeholder && c.image != c.placeholder {
			err := AbortedContainerError{
				container: c.id,
				state:     cs.state,
				exitCode:  cs.exitCode,
				reason:    cs.stateInfo,
			}

			c.state = stateTerminated
			c.stateInfo = cs.stateInfo
			c.exitCode = cs.exitCode
			pw.notifyClientsContainerChange(c, err)

			continue
		}

		c.image = cs.image
		c.state = cs.state
		c.stateInfo = cs.stateInfo
		c.exitCode = cs.exitCode

		if c.image == c.placeholder {
			if c.state == stateTerminated {
				err := FailedContainerError{
					container: c.id,
					exitCode:  c.exitCode,
					reason:    c.stateInfo,
				}

				pw.notifyClientsContainerChange(c, err)
			}
		} else {
			logrus.
				WithField("pod", pw.podName).
				WithField("container", c.id).
				WithField("image", c.image).
				WithField("state", c.state).
				WithField("stateInfo", c.stateInfo).
				Debug("PodWatcher: Container state changed")

			pw.notifyClientsContainerChange(c, nil)
		}
	}
}

// _tryResolveWaitClient will resolve the waitClient if the container state is greater or equal to the requested state.
// For example: A container in TERMINATED state will resolve all clients waiting for it to enter RUNNING state
// and all clients waiting for it to enter TERMINATED state.
func _tryResolveWaitClient(cl *waitClient, c *containerInfo, err error) bool {
	if cl.containerId != c.id {
		return false
	}

	if err != nil {
		// tell the waitClient to fail
		cl.resolveCh <- err
		return true
	}

	if image.Match(c.image, c.placeholder) {
		return false
	}

	if c.state >= cl.containerState {
		if cl.containerState == stateTerminated && c.exitCode != 0 {
			// tell the waitClient that the container failed with an exit code != 0
			cl.resolveCh <- exitCodeError(int(c.exitCode))
		} else {
			// tell the waitClient to proceed
			cl.resolveCh <- nil
		}

		return true
	}

	return false
}

// notifyClientsContainerChange resolves all wait clients that wait for a specific state of a container.
func (pw *PodWatcher) notifyClientsContainerChange(c *containerInfo, err error) {
	if err != nil {
		_, isFailed := err.(FailedContainerError)
		_, isAborted := err.(AbortedContainerError)
		if isKubeError := isFailed || isAborted; isKubeError {
			for _, cl := range pw.clientList {
				if cl.containerId == c.id {
					cl.resolveCh <- err
				} else {
					cl.resolveCh <- OtherContainerError{Err: err}
				}
			}
			pw.clientList = nil
			return
		}
	}

	for clIdx := 0; clIdx < len(pw.clientList); {
		cl := pw.clientList[clIdx]

		if _tryResolveWaitClient(cl, c, err) {
			// remove the waitClient from the list (order is not preserved)
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
			cl.resolveCh <- PodTerminatedError{}
		}
	}

	pw.clientList = nil
}

func (pw *PodWatcher) waitForEvent(containerId string, state containerState) (err error) {
	ch := make(chan error)

	logrus.
		WithField("pod", pw.podName).
		WithField("container", containerId).
		WithField("state", state.String()).
		Trace("PodWatcher: Waiting...")

	defer func(t time.Time) {
		logrus.
			WithError(err).
			WithField("pod", pw.podName).
			WithField("container", containerId).
			WithField("state", state.String()).
			Debugf("PodWatcher: Wait finished. Duration=%.2fs", time.Since(t).Seconds())
	}(time.Now())

	select {
	case pw.clientCh <- &waitClient{containerId: containerId, containerState: state, resolveCh: ch}:
		err = <-ch

	case <-pw.stop:
		if pw.errDone != nil {
			err = pw.errDone
		} else if containerId != "" {
			err = PodTerminatedError{}
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
	// note: the state used below is unimportant, it's used only for logging
	return pw.waitForEvent("", stateTerminated)
}

// AddContainer registers a container for state tracking.
// Adding containers is necessary because PodWatcher
// must know name of the placeholder image for each container.
func (pw *PodWatcher) AddContainer(id string, placeholder string) error {
	select {
	case pw.containerRegCh <- containerInfo{
		id:          id,
		state:       stateWaiting,
		stateInfo:   "",
		image:       placeholder,
		placeholder: placeholder,
		exitCode:    0,
	}:
		return nil
	case <-pw.stop:
		return PodTerminatedError{}
	}
}

func (pw *PodWatcher) Name() string {
	return pw.podName
}
