// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"strings"

	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
)

const (
	workspacePath     = "/drone/src"
	workspaceName     = "workspace"
	workspaceHostName = "host"
)

func createWorkspace(from *resource.Pipeline) string {
	path := workspacePath
	if from.Workspace.Path != "" {
		path = from.Workspace.Path
	}
	if from.Platform.OS == "windows" {
		path = toWindowsDrive(path)
	}
	return path
}

func setupWorkdir(src *resource.Step, dst *engine.Step, path string) {
	// if the working directory is already set
	// do not alter.
	if dst.WorkingDir != "" {
		return
	}
	// if the user is running the container as a
	// service (detached mode) with no commands, we
	// should use the default working directory.
	if dst.Detach && len(src.Commands) == 0 {
		return
	}
	// else set the working directory.
	dst.WorkingDir = path
}

// helper function converts the path to a valid windows
// path, including the default C drive.
func toWindowsDrive(s string) string {
	return "c:" + toWindowsPath(s)
}

// helper function converts the path to a valid windows
// path, replacing backslashes with forward slashes.
func toWindowsPath(s string) string {
	return strings.Replace(s, "/", "\\", -1)
}
