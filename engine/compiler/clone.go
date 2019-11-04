// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"strconv"

	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone/runner-go/manifest"
)

// default name of the clone step.
const cloneStepName = "clone"

// helper function returns the clone image based on the
// target operating system.
func cloneImage(platform manifest.Platform) string {
	switch platform.OS {
	case "windows":
		return "drone/git:latest"
	default:
		return "drone/git:latest"
	}
}

// helper function configures the clone depth parameter,
// specific to the clone plugin.
func cloneParams(src manifest.Clone) map[string]string {
	dst := map[string]string{}
	if depth := src.Depth; depth > 0 {
		dst["PLUGIN_DEPTH"] = strconv.Itoa(depth)
	}
	if skipVerify := src.SkipVerify; skipVerify {
		dst["GIT_SSL_NO_VERIFY"] = "true"
		dst["PLUGIN_SKIP_VERIFY"] = "true"
	}
	return dst
}

// helper function creates a default container configuration
// for the clone stage. The clone stage is automatically
// added to each pipeline.
func createClone(src *resource.Pipeline) *engine.Step {
	return &engine.Step{
		Name:        cloneStepName,
		Image:       cloneImage(src.Platform),
		Placeholder: placeholderImage,
		RunPolicy:   engine.RunAlways,
		Envs:        cloneParams(src.Clone),
	}
}
