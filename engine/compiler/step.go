// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"strings"

	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone-runners/drone-runner-kube/internal/docker/image"
	"github.com/drone-runners/drone-runner-kube/internal/encoder"

	"github.com/drone/runner-go/pipeline/runtime"
)

func createStep(spec *resource.Pipeline, src *resource.Step) *engine.Step {
	dst := &engine.Step{
		ID:           random(),
		Name:         src.Name,
		Image:        image.Expand(src.Image),
		Command:      src.Command,
		Entrypoint:   src.Entrypoint,
		Detach:       src.Detach,
		DependsOn:    src.DependsOn,
		Envs:         convertStaticEnv(src.Environment),
		IgnoreStderr: false,
		IgnoreStdout: false,
		Privileged:   src.Privileged,
		Pull:         convertPullPolicy(src.Pull),
		User:         src.User,
		Group:        src.Group,
		Resources:    convertResources(src.Resources),
		Secrets:      convertSecretEnv(src.Environment),
		WorkingDir:   src.WorkingDir,
	}

	// appends the volumes to the container def.
	for _, vol := range src.Volumes {
		dst.Volumes = append(dst.Volumes, &engine.VolumeMount{
			Name: vol.Name,
			Path: vol.MountPath,
		})
	}

	// appends the settings variables to the
	// container definition.
	for key, value := range src.Settings {
		// fix https://github.com/drone/drone-yaml/issues/13
		if value == nil {
			continue
		}
		// all settings are passed to the plugin env
		// variables, prefixed with PLUGIN_
		key = "PLUGIN_" + strings.ToUpper(key)

		// if the setting parameter is sources from the
		// secret we create a secret environment variable.
		if value.Secret != "" {
			dst.Secrets = append(dst.Secrets, &engine.SecretVar{
				Name: value.Secret,
				Env:  key,
			})
		} else {
			// else if the setting parameter is opaque
			// we inject as a string-encoded environment
			// variable.
			dst.Envs[key] = encoder.Encode(value.Value)
		}
	}

	// set the pipeline step run policy. steps run on
	// success by default, but may be optionally configured
	// to run on failure.
	if isRunAlways(src) {
		dst.RunPolicy = runtime.RunAlways
	} else if isRunOnFailure(src) {
		dst.RunPolicy = runtime.RunOnFailure
	}

	// set the pipeline failure policy. steps can choose
	// to ignore the failure, or fail fast.
	switch src.Failure {
	case "ignore":
		dst.ErrPolicy = runtime.ErrIgnore
	case "fast", "fast-fail", "fail-fast":
		dst.ErrPolicy = runtime.ErrFailFast
	}

	return dst
}
