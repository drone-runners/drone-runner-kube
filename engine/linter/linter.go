// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package linter

import (
	"errors"
	"fmt"

	"github.com/drone-runners/drone-runner-kube/engine/resource"
)

// ErrDuplicateStepName is returned when two Pipeline steps
// have the same name.
var ErrDuplicateStepName = errors.New("linter: duplicate step names")

// ErrMissingDependency is returned when a Pipeline step
// defines dependencies that are invlid or unknown.
var ErrMissingDependency = errors.New("linter: invalid or unknown step dependency")

// ErrCyclicalDependency is returned when a Pipeline step
// defines a cyclical dependency, which would result in an
// infinite execution loop.
var ErrCyclicalDependency = errors.New("linter: cyclical step dependency detected")

// Opts provides linting options.
type Opts struct {
	Trusted bool
}

// Linter evaluates the pipeline against a set of
// rules and returns an error if one or more of the
// rules are broken.
type Linter struct{}

// New returns a new Linter.
func New() *Linter {
	return new(Linter)
}

// Lint executes the linting rules for the pipeline
// configuration.
func (l *Linter) Lint(pipeline *resource.Pipeline, opts Opts) error {
	return checkPipeline(pipeline, opts.Trusted)
}

func checkPipeline(pipeline *resource.Pipeline, trusted bool) error {
	// if err := checkNames(pipeline); err != nil {
	// 	return err
	// }
	if err := checkSteps(pipeline, trusted); err != nil {
		return err
	}
	if err := checkVolumes(pipeline, trusted); err != nil {
		return err
	}
	return nil
}

// func checkNames(pipeline *resource.Pipeline) error {
// 	names := map[string]struct{}{}
// 	if !pipeline.Clone.Disable {
// 		names["clone"] = struct{}{}
// 	}
// 	steps := append(pipeline.Services, pipeline.Steps...)
// 	for _, step := range steps {
// 		_, ok := names[step.Name]
// 		if ok {
// 			return ErrDuplicateStepName
// 		}
// 		names[step.Name] = struct{}{}
// 	}
// 	return nil
// }

func checkSteps(pipeline *resource.Pipeline, trusted bool) error {
	steps := append(pipeline.Services, pipeline.Steps...)
	for _, step := range steps {
		if step == nil {
			return errors.New("linter: nil step")
		}
		if err := checkStep(step, trusted); err != nil {
			return err
		}
	}
	return nil
}

func checkStep(step *resource.Step, trusted bool) error {
	if step.Image == "" {
		return errors.New("linter: invalid or missing image")
	}
	// if step.Name == "" {
	// 	return errors.New("linter: invalid or missing name")
	// }
	// if len(step.Name) > 100 {
	// 	return errors.New("linter: name exceeds maximum length")
	// }
	if trusted == false && step.Privileged {
		return errors.New("linter: untrusted repositories cannot enable privileged mode")
	}
	if trusted == false && len(step.Devices) > 0 {
		return errors.New("linter: untrusted repositories cannot mount devices")
	}
	if trusted == false && len(step.DNS) > 0 {
		return errors.New("linter: untrusted repositories cannot configure dns")
	}
	if trusted == false && len(step.DNSSearch) > 0 {
		return errors.New("linter: untrusted repositories cannot configure dns_search")
	}
	if trusted == false && len(step.ExtraHosts) > 0 {
		return errors.New("linter: untrusted repositories cannot configure extra_hosts")
	}
	if trusted == false && len(step.Network) > 0 {
		return errors.New("linter: untrusted repositories cannot configure network_mode")
	}
	for _, mount := range step.Volumes {
		switch mount.Name {
		case "workspace", "_workspace", "_docker_socket":
			return fmt.Errorf("linter: invalid volume name: %s", mount.Name)
		}
	}
	return nil
}

func checkVolumes(pipeline *resource.Pipeline, trusted bool) error {
	for _, volume := range pipeline.Volumes {
		if volume.EmptyDir != nil {
			err := checkEmptyDirVolume(volume.EmptyDir, trusted)
			if err != nil {
				return err
			}
		}
		if volume.HostPath != nil {
			err := checkHostPathVolume(volume.HostPath, trusted)
			if err != nil {
				return err
			}
		}
		switch volume.Name {
		case "":
			return fmt.Errorf("linter: missing volume name")
		case "workspace", "_workspace", "_docker_socket":
			return fmt.Errorf("linter: invalid volume name: %s", volume.Name)
		}
	}
	return nil
}

func checkHostPathVolume(volume *resource.VolumeHostPath, trusted bool) error {
	if trusted == false {
		return errors.New("linter: untrusted repositories cannot mount host volumes")
	}
	return nil
}

func checkEmptyDirVolume(volume *resource.VolumeEmptyDir, trusted bool) error {
	if trusted == false && volume.Medium == "memory" {
		return errors.New("linter: untrusted repositories cannot mount in-memory volumes")
	}
	return nil
}
