// Copyright 2021 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package podwatcher

import (
	"context"
	"strconv"
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

	// containerWatchInfo holds info about all containers in the pod.
	containerMap map[string]*containerWatchInfo

	// state represents PodWatcher state and can be: "init", "started" or "done".
	state watcherState

	// stop channel is used to finish the PodWatcher.
	// It prevents adding new wait clients and resolves all that are waiting.
	stop chan struct{}

	// errDone is used to store an eventual error from container watcher.
	errDone error

	// containerRegCh is a channel through which new containers are registered.
	containerRegCh chan containerRegInfo

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
	pw.containerRegCh = make(chan containerRegInfo)
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
					pw.containerMap = make(map[string]*containerWatchInfo)
				}

				pw.containerMap[c.containerId] = &containerWatchInfo{
					id:          c.containerId,
					image:       c.image,
					placeholder: c.placeholder,
					stepState:   stepStateWaiting,
					addedAt:     time.Now(),
				}

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
				if !_tryResolveWaitClient(cl, c) {
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

		// We've already declared the container as finished or failed, so just notify all that it's terminated.
		if c.stepState >= stepStateFinished {
			if cs.state != stateTerminated {
				// Should not happen: A container that was marked as terminated is now running again... just log and proceed
				logrus.
					WithField("pod", pw.podName).
					WithField("container", c.id).
					WithFields(cs.stateToMap()).
					Warn("PodWatcher: Container zombie found...")
			}
			pw.notifyClients(c)
			continue
		}

		isPlaceholder := image.Match(cs.image, c.placeholder)

		// A running container with placeholder image (that we track, so present in pw.containerMap)
		// usually means that Kubernetes is downloading the real step image in background.
		if isPlaceholder && (cs.state == stateWaiting || cs.state == stateRunning) && cs.restartCount == 0 {
			continue
		}

		// Sometimes, kubernetes sends an event about a terminated container with:
		// Terminated.ExitCode=2 and Terminated.Reason="Error".
		// Often container the current image will revert back to the placeholder image.
		// Kubernetes is probably downloading an image for the container in the background.
		if cs.exitCode == 2 && cs.reason == "Error" {
			if cs.restartCount == 0 {
				continue
			}

			recoveryDuration := time.Minute

			if c.failedAt.IsZero() {
				logrus.
					WithField("pod", pw.podName).
					WithField("container", c.id).
					WithFields(cs.stateToMap()).
					Trace("PodWatcher: Container failed. Trying recovery...")
				c.failedAt = time.Now()
			} else if time.Since(c.failedAt) < recoveryDuration {
				logrus.
					WithField("pod", pw.podName).
					WithField("container", c.id).
					WithFields(cs.stateToMap()).
					Trace("PodWatcher: Container failed. Waiting to recover...")
			} else {
				logrus.
					WithField("pod", pw.podName).
					WithField("container", c.id).
					WithFields(cs.stateToMap()).
					Debug("PodWatcher: Container failed.")

				c.stepState = stepStateFailed
				c.exitCode = cs.exitCode
				c.reason = cs.reason

				pw.notifyClients(c)
			}

			continue

		} else {
			c.failedAt = time.Time{}
		}

		switch cs.state {
		case stateTerminated:
			if isPlaceholder {
				c.stepState = stepStatePlaceholderFailed

				logrus.
					WithField("pod", pw.podName).
					WithField("container", c.id).
					WithFields(cs.stateToMap()).
					Debug("PodWatcher: Container placeholder terminated.")
			} else {
				c.stepState = stepStateFinished

				logrus.
					WithField("pod", pw.podName).
					WithField("container", c.id).
					WithFields(cs.stateToMap()).
					Debug("PodWatcher: Container terminated.")
			}
			c.exitCode = cs.exitCode
			c.reason = cs.reason

			pw.notifyClients(c)

		case stateWaiting, stateRunning:
			var s stepState
			if isPlaceholder || cs.state == stateWaiting {
				s = stepStateWaiting
			} else {
				s = stepStateRunning
			}

			c.restartCount = int(cs.restartCount)

			if s == c.stepState && c.reason == cs.reason {
				continue // step state unchanged
			}

			c.stepState = s
			c.exitCode = 0
			c.reason = cs.reason

			logrus.
				WithField("pod", pw.podName).
				WithField("container", c.id).
				WithField("stepState", c.stepState).
				WithFields(cs.stateToMap()).
				Debug("PodWatcher: Container state changed")

			pw.notifyClients(c)
		}
	}
}

func (pw *PodWatcher) notifyClients(c *containerWatchInfo) {
	switch c.stepState {
	case stepStateWaiting:
	case stepStateRunning, stepStateFinished:
		pw.notifyClientsContainerChange(c)
	case stepStatePlaceholderFailed, stepStateFailed:
		pw.notifyClientsError(c, FailedContainerError{
			container: c.id,
			exitCode:  c.exitCode,
			reason:    c.reason,
			image:     c.image,
		})
	}
}

// notifyClientsError resolves wait clients with an error.
func (pw *PodWatcher) notifyClientsError(c *containerWatchInfo, err error) {
	_, isFailed := err.(FailedContainerError)
	isKubeError := isFailed

	for clIdx := 0; clIdx < len(pw.clientList); {
		cl := pw.clientList[clIdx]

		if isKubeError {
			if cl.containerId == "" {
				clIdx++
				continue
			} else if cl.containerId == c.id {
				cl.resolveCh <- err
			} else {
				cl.resolveCh <- OtherContainerError{Err: err}
			}
		} else if cl.containerId == c.id {
			cl.resolveCh <- err
		} else {
			clIdx++
			continue
		}

		// remove the waitClient from the list (order is not preserved)
		pw.clientList[clIdx] = pw.clientList[len(pw.clientList)-1]
		pw.clientList[len(pw.clientList)-1] = nil
		pw.clientList = pw.clientList[:len(pw.clientList)-1]
	}
}

// notifyClientsContainerChange resolves all wait clients that wait for a specific state of a container.
func (pw *PodWatcher) notifyClientsContainerChange(c *containerWatchInfo) {
	for clIdx := 0; clIdx < len(pw.clientList); {
		cl := pw.clientList[clIdx]

		if !_tryResolveWaitClient(cl, c) {
			clIdx++
			continue
		}

		// remove the waitClient from the list (order is not preserved)
		pw.clientList[clIdx] = pw.clientList[len(pw.clientList)-1]
		pw.clientList[len(pw.clientList)-1] = nil
		pw.clientList = pw.clientList[:len(pw.clientList)-1]
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

// _tryResolveWaitClient will resolve the waitClient if the step state is greater or equal to the requested state.
// For example: A container in "finished" state will resolve all clients waiting for it to enter either
// "running" or "finished" state.
func _tryResolveWaitClient(cl *waitClient, c *containerWatchInfo) bool {
	if cl.containerId != c.id {
		return false
	}

	if c.stepState < cl.waitForState {
		return false
	}

	if cl.waitForState == stepStateFinished && c.exitCode != 0 {
		// tell the waitClient that the container failed with an exit code != 0
		cl.resolveCh <- exitCodeError(int(c.exitCode))
	} else {
		// tell the waitClient to proceed
		cl.resolveCh <- nil
	}

	return true
}

func (pw *PodWatcher) waitForEvent(containerId string, stepState stepState) (err error) {
	ch := make(chan error)

	logrus.
		WithField("pod", pw.podName).
		WithField("container", containerId).
		WithField("stepState", stepState.String()).
		Debug("PodWatcher: Waiting...")

	defer func(t time.Time) {
		logrus.
			WithError(err).
			WithField("pod", pw.podName).
			WithField("container", containerId).
			WithField("stepState", stepState.String()).
			Debugf("PodWatcher: Wait finished. Duration=%.2fs", time.Since(t).Seconds())
	}(time.Now())

	select {
	case pw.clientCh <- &waitClient{containerId: containerId, waitForState: stepState, resolveCh: ch}:
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
	return pw.waitForEvent(containerId, stepStateRunning)
}

// WaitContainerReStart waits until a container in the pod restarts.
func (pw *PodWatcher) WaitContainerReStart(containerId string) bool {
	logrus.
		WithField("pod", pw.podName).
		WithField("container", containerId).
		Debug("PodWatcher: Waiting to be restated")
	retries := 0
	for retries < 60 {
		if pw.containerMap[containerId].stepState != stepStateRunning {
			return false
		}
		if pw.containerMap[containerId].restartCount > 0 {
			return true
		}
		retries++
		logrus.
			WithField("pod", pw.podName).
			WithField("container", containerId).
			WithField("restart count", strconv.Itoa(pw.containerMap[containerId].restartCount)).
			Debug("PodWatcher: Waiting to be restated")

		<-time.After(time.Second * 5)
	}
	return false
}

// WaitContainerTerminated waits until a container in the pod is terminated.
func (pw *PodWatcher) WaitContainerTerminated(containerId string) (int, error) {
	err := pw.waitForEvent(containerId, stepStateFinished)
	if code, ok := err.(exitCodeError); ok { // exit codes != 0 are masked as a exitCodeError
		return int(code), nil
	}

	return 0, err
}

// WaitPodDeleted waits until the pod is deleted.
func (pw *PodWatcher) WaitPodDeleted() (err error) {
	// note: the state used below is unimportant, it's used only for logging
	return pw.waitForEvent("", stepStateFinished)
}

// AddContainer registers a container for state tracking.
// Adding containers is necessary because PodWatcher
// must know name of the placeholder image for each container.
func (pw *PodWatcher) AddContainer(id, placeholder, image string) error {
	select {
	case pw.containerRegCh <- containerRegInfo{containerId: id, placeholder: placeholder, image: image}:
		return nil
	case <-pw.stop:
		return PodTerminatedError{}
	}
}

func (pw *PodWatcher) Name() string {
	return pw.podName
}
