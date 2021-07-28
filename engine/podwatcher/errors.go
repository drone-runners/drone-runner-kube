// Copyright 2021 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package podwatcher

import "fmt"

// UnknownContainerError is an error that wait functions return when an unregistered container name is provided.
type UnknownContainerError struct {
	container string
}

func (e UnknownContainerError) Error() string {
	return "unknown container: " + e.container
}

// PodTerminatedError is an error that container wait functions return when the pod is already terminated.
type PodTerminatedError struct{}

func (e PodTerminatedError) Error() string {
	return "pod is terminated"
}

// FailedContainerError is returned as an error when placeholder container terminates abnormally.
// The correct container image failed to load. Usually happens when image doesn't exist.
type FailedContainerError struct {
	container string
	exitCode  int32
	reason    string
}

func (e FailedContainerError) Error() string {
	return fmt.Sprintf(
		"kubernetes error: container failed to start: id=%s exitcode=%d reason=%s",
		e.container, e.exitCode, e.reason)
}

// AbortedContainerError is an error returned when a container, that was earlier started successfully,
// suddenly reverted image back to the placeholder image and terminated.
type AbortedContainerError struct {
	container string
	state     containerState
	exitCode  int32
	reason    string
}

func (e AbortedContainerError) Error() string {
	return fmt.Sprintf(
		"kubernetes error: container failed to start and reverted back to placeholder image: id=%s state=%s exitCode=%d reason=%s",
		e.container, e.state, e.exitCode, e.reason)
}

// StartTimeoutContainerError is returned as an error when a container fails to run after some predefined time.
type StartTimeoutContainerError struct {
	Container string
}

func (e StartTimeoutContainerError) Error() string {
	return fmt.Sprintf(
		"kubernetes error: container failed to start in timely manner: id=%s",
		e.Container)
}

// OtherContainerError is returned as an error by wait function when some other container
// in the same pod fails with a kubernetes error.
type OtherContainerError struct {
	Err error
}

func (e OtherContainerError) Error() string {
	return fmt.Sprintf(
		"kubernetes error: aborting due to error: %s",
		e.Err)
}
