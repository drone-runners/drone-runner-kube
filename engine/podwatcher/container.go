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
	containerId    string
	containerState containerState
	resolveCh      chan error
}

// containerInfo is used by the PodWatcher to track state of each container inside a pod.
type containerInfo struct {
	id          string
	state       containerState
	stateInfo   string
	placeholder string
	image       string
	exitCode    int32

	// failAt is used by PodWatcher to recover from invalid Kubernetes events.
	_failAt time.Time
}

func (info *containerInfo) diff(old *containerInfo) (m map[string]interface{}) {
	if old == nil {
		return
	}

	m = make(map[string]interface{})

	if info.state != old.state {
		m["state"] = old.state.String() + "->" + info.state.String()
	}
	if info.stateInfo != old.stateInfo {
		if old.stateInfo == "" {
			m["stateInfo"] = info.stateInfo
		} else {
			m["stateInfo"] = old.stateInfo + "->" + info.stateInfo
		}
	}
	if info.image != old.image {
		m["image"] = old.image + "->" + info.image
	}
	if info.exitCode != old.exitCode {
		m["exitCode"] = info.exitCode
	}

	return
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
