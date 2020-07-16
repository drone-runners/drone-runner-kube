// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"strings"

	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/compiler/shell"
	"github.com/drone-runners/drone-runner-kube/engine/compiler/shell/powershell"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
)

// helper function configures the pipeline script for the
// target operating system.
func setupScript(src *resource.Step, dst *engine.Step, os string) {
	if len(src.Commands) > 0 {
		switch os {
		case "windows":
			setupScriptWindows(src.Commands, dst)
		default:
			setupScriptPosix(src.Commands, dst)
		}
	}

	if len(src.Entrypoint) > 0 {
		cmds := []string{
			strings.Join(append(src.Entrypoint, src.Command...), " "),
		}
		switch os {
		case "windows":
			setupScriptWindows(cmds, dst)
		default:
			setupScriptPosix(cmds, dst)
		}
	}
}

// helper function configures the pipeline script for the
// windows operating system.
func setupScriptWindows(commands []string, dst *engine.Step) {
	dst.Entrypoint = []string{"powershell", "-noprofile", "-noninteractive", "-command"}
	dst.Command = []string{"echo $Env:DRONE_SCRIPT | iex"}
	dst.Envs["DRONE_SCRIPT"] = powershell.Script(commands)
	dst.Envs["SHELL"] = "powershell.exe"
}

// helper function configures the pipeline script for the
// linux operating system.
func setupScriptPosix(commands []string, dst *engine.Step) {
	if dst.Shell == "" {
		dst.Entrypoint = []string{"/bin/sh", "-c"}
		dst.Command = []string{`echo "$DRONE_SCRIPT" | /bin/sh`}
	} else {
		dst.Entrypoint = []string{dst.Shell, "-c"}
		dst.Command = []string{`echo "$DRONE_SCRIPT" | ` + dst.Shell}
	}
	dst.Envs["DRONE_SCRIPT"] = shell.Script(commands)
}
