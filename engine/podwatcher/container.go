// Copyright 2021 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package podwatcher

import (
	"context"
	"fmt"
)

type ContainerWatcher interface {
	// Name returns and of the pod that contains the containers
	Name() string

	// Watch waits for the container updates and puts the new status to the channel.
	// The function should close the containers channel when it finishes.
	// It should finish either when the context is done or when no more events are expected.
	Watch(ctx context.Context, containers chan<- []containerInfo) error

	// PeriodicCheck should periodically put the current state of the containers to the channel.
	// The function should close the containers channel when it finishes.
	// It should finish either when the context is done or when the stop channel is closed.
	// To disable the feature, the function should be a no op, and the containers channel should remain open.
	PeriodicCheck(ctx context.Context, containers chan<- []containerInfo, stop <-chan struct{}) error
}

// client represents a process that waits for a container state change.
// Wait is resolved by writing an error value to the resolveCh channel.
// If containerId is an empty string, the process waits for whole the pod to finish.
type client struct {
	containerId    string
	containerState containerState
	resolveCh      chan error
}

// containerInfo is used by the PodWatcher to track state of each container inside of a pod.
type containerInfo struct {
	id          string
	state       containerState
	stateInfo   string
	placeholder string
	image       string
	exitCode    int32
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

// exitCodeError is used to return exit code of a terminated container if the exit code is not zero.
type exitCodeError int

func (e exitCodeError) Error() string {
	return fmt.Sprintf("exitCode=%d", e)
}
