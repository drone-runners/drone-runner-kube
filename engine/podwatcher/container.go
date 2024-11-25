// Copyright 2021 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package podwatcher

import (
	"context"
	"fmt"
	"time"
)

type ContainerWatcher interface {
	// Name returns name of the pod that contains the containers
	Name() string

	// Watch waits for updates of the containers and puts the updated data to the channel passed as a parameter.
	// It should finish either when the context is done or when no more events are expected.
	Watch(ctx context.Context, containers chan<- []containerInfo) error

	// PeriodicCheck should periodically put the current state of the containers to the channel.
	// It should finish either when the context is done or when the stop channel is closed.
	// To disable the feature, the implementation should be an empty function.
	PeriodicCheck(ctx context.Context, containers chan<- []containerInfo, stop <-chan struct{}) error
}

// waitClient is a process that waits for state of a container (with id = containerId) to change to containerState.
// It is resolved by writing an error value to the resolveCh channel, or nil if no error occurred.
// If containerId is an empty string, the process waits for whole the pod to finish.
type waitClient struct {
	containerId  string
	waitForState stepState
	resolveCh    chan error
}

// containerRegInfo is used by the PodWatcher register new containers to watch.
type containerRegInfo struct {
	containerId string
	placeholder string
	image       string
}

// containerInfo is used by the ContainerWatcher to send info about a container to PodWatcher.
type containerInfo struct {
	id           string
	state        containerState
	image        string
	exitCode     int32
	reason       string
	restartCount int32
	ready        bool
}

func (info *containerInfo) stateToMap() (m map[string]interface{}) {
	m = make(map[string]interface{})
	m["state"] = info.state.String()
	m["image"] = info.image
	if info.exitCode != 0 {
		m["exitCode"] = info.exitCode
	}
	if info.reason != "" {
		m["reason"] = info.reason
	}
	if info.restartCount != 0 {
		m["restartCount"] = info.restartCount
	}
	if !info.ready {
		m["ready"] = "true"
	}
	return
}

// containerWatchInfo is used by the PodWatcher to track the state of each container inside a pod.
type containerWatchInfo struct {
	id          string
	image       string
	placeholder string

	stepState stepState

	exitCode     int32
	reason       string
	restartCount int

	addedAt  time.Time
	failedAt time.Time
}

type containerState int

const (
	stateWaiting containerState = iota
	stateRunning
	stateTerminated
)

func (s containerState) String() string {
	switch s {
	case stateWaiting:
		return "WAITING"
	case stateRunning:
		return "RUNNING"
	case stateTerminated:
		return "TERMINATED"
	default:
		panic("unsupported containerInfo state")
	}
}

// stepState is used to track the current state of a step.
type stepState int

const (
	stepStateWaiting = iota
	stepStateRunning
	stepStateFinished
	stepStatePlaceholderFailed
	stepStateFailed
)

func (s stepState) String() string {
	switch s {
	case stepStateWaiting:
		return "waiting"
	case stepStateRunning:
		return "running"
	case stepStateFinished:
		return "finished"
	case stepStatePlaceholderFailed:
		return "failed-p"
	case stepStateFailed:
		return "failed"
	default:
		panic("unsupported containerInfo state")
	}
}

// exitCodeError is used to return exit code of a terminated container if the exit code is not zero.
type exitCodeError int

func (e exitCodeError) Error() string {
	return fmt.Sprintf("exitCode=%d", e)
}
