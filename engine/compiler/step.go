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
		DNS:          src.DNS,
		DNSSearch:    src.DNSSearch,
		Envs:         convertStaticEnv(src.Environment),
		ExtraHosts:   src.ExtraHosts,
		IgnoreErr:    strings.EqualFold(src.Failure, "ignore"),
		IgnoreStderr: false,
		IgnoreStdout: false,
		Privileged:   src.Privileged,
		Pull:         convertPullPolicy(src.Pull),
		User:         src.User,
		Secrets:      convertSecretEnv(src.Environment),
		WorkingDir:   src.WorkingDir,

		Volumes: nil, // set below
		Devices: nil, // see below
	}

	// appends the volumes to the container def.
	for _, vol := range src.Volumes {
		dst.Volumes = append(dst.Volumes, &engine.VolumeMount{
			Name: vol.Name,
			Path: vol.MountPath,
		})
	}

	// appends the devices to the container def.
	for _, vol := range src.Devices {
		dst.Devices = append(dst.Devices, &engine.VolumeDevice{
			Name:       vol.Name,
			DevicePath: vol.DevicePath,
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
		// secret we create a secret enviornment variable.
		if value.Secret != "" {
			dst.Secrets = append(dst.Secrets, &engine.Secret{
				Name: value.Secret,
				Mask: true,
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
		dst.RunPolicy = engine.RunAlways
	} else if isRunOnFailure(src) {
		dst.RunPolicy = engine.RunOnFailure
	}

	return dst
}
