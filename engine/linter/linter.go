// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package linter

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/manifest"
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
	Trusted   bool
	Namespace string
	Name      string
	Slug      string
}

// Linter evaluates the pipeline against a set of
// rules and returns an error if one or more of the
// rules are broken.
type Linter struct {
	patterns map[string][]string
}

// New returns a new Linter.
func New(patterns map[string][]string) *Linter {
	return &Linter{patterns: patterns}
}

// Lint executes the linting rules for the pipeline
// configuration.
func (l *Linter) Lint(pm manifest.Resource, repo *drone.Repo) error {
	pipeline := pm.(*resource.Pipeline)
	if err := checkStageResources(pipeline); err != nil {
		return err
	}
	if err := checkSteps(pipeline, repo.Trusted); err != nil {
		return err
	}
	if err := checkVolumes(pipeline, repo.Trusted); err != nil {
		return err
	}
	if err := checkNamespace(pipeline.Metadata.Namespace, repo.Slug, l.patterns); err != nil {
		return err
	}
	return nil
}

func checkStageResources(pipeline *resource.Pipeline) error {
	if pipeline.Resources.Limits.CPU != 0 {
		return errors.New("linter: cpu limit cannot be applied at stage level")
	}
	if pipeline.Resources.Limits.Memory != 0 {
		return errors.New("linter: memory limit cannot be applied at stage level")
	}
	return nil
}

func checkSteps(pipeline *resource.Pipeline, trusted bool) error {
	steps := append(pipeline.Services, pipeline.Steps...)

	names := map[string]struct{}{}
	if !pipeline.Clone.Disable {
		names["clone"] = struct{}{}
	}

	for _, step := range steps {
		if step == nil {
			return errors.New("linter: nil step")
		}

		// unique list of names
		_, ok := names[step.Name]
		if ok {
			return ErrDuplicateStepName
		}
		names[step.Name] = struct{}{}

		if err := checkStep(step, trusted); err != nil {
			return err
		}
		if err := checkDeps(step, names); err != nil {
			return err
		}
	}
	return nil
}

func checkStep(step *resource.Step, trusted bool) error {
	if step.Image == "" {
		return errors.New("linter: invalid or missing image")
	}
	if trusted == false && step.Privileged {
		return errors.New("linter: untrusted repositories cannot enable privileged mode")
	}
	for _, mount := range step.Volumes {
		switch mount.Name {
		case "workspace", "_workspace", "_docker_socket", "_status":
			return fmt.Errorf("linter: invalid volume name: %s", mount.Name)
		}
		if strings.HasPrefix(filepath.Clean(mount.MountPath), "/run/drone") {
			return fmt.Errorf("linter: cannot mount volume at /run/drone")
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
		if volume.Claim != nil {
			err := checkClaimVolume(volume.Claim, trusted)
			if err != nil {
				return err
			}
		}
		switch volume.Name {
		case "":
			return fmt.Errorf("linter: missing volume name")
		case "workspace", "_workspace", "_docker_socket", "_status", "_addons":
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

func checkClaimVolume(volume *resource.VolumeClaim, trusted bool) error {
	if trusted == false {
		return errors.New("linter: untrusted repositories cannot mount PVC")
	}
	return nil
}

func checkEmptyDirVolume(volume *resource.VolumeEmptyDir, trusted bool) error {
	if trusted == false && volume.Medium == "memory" {
		return errors.New("linter: untrusted repositories cannot mount in-memory volumes")
	}
	return nil
}

func checkNamespace(namespace, name string, mapping map[string][]string) error {
	if len(mapping) == 0 {
		return nil
	}
	if len(namespace) == 0 {
		return nil
	}
	patterns, ok := mapping[namespace]
	if !ok {
		return nil
	}
	for _, pattern := range patterns {
		if match, _ := doublestar.Match(pattern, name); match {
			return nil
		}
	}
	return errors.New("linter: pipeline restricted from using configured namespace")
}

func checkDeps(step *resource.Step, deps map[string]struct{}) error {
	for _, dep := range step.DependsOn {
		_, ok := deps[dep]
		if !ok {
			return fmt.Errorf("linter: unknown step dependency detected: %s references %s", step.Name, dep)
		}
		if step.Name == dep {
			return fmt.Errorf("linter: cyclical step dependency detected: %s", dep)
		}
	}
	return nil
}
