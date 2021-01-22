// Copyright 2020 Drone.IO Inc.
// Copyright 2019 The Tekton Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package entrypoint

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

// ErrSkip is returned from the wait function when the step
// should be skipped.
var ErrSkip = errors.New("Skip the step")

type entrypointCommand struct {
	name     string
	args     []string
	waitfile string
	donefile string
	interval time.Duration
}

func (c *entrypointCommand) run(*kingpin.ParseContext) error {
	if len(c.waitfile) != 0 {
		err := c.wait(c.waitfile, false)
		if err == ErrSkip {
			return nil
		} else if err != nil {
			return err
		}
	}

	cmd := exec.Command(c.name, c.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return err
	}

	signals := make(chan os.Signal, 1)
	defer close(signals)
	signal.Notify(signals)
	defer signal.Reset()

	go func() {
		for s := range signals {
			// Forward signal to main process and all children
			if s != syscall.SIGCHLD {
				_ = syscall.Kill(-cmd.Process.Pid, s.(syscall.Signal))
			}
		}
	}()

	err := cmd.Wait()

	if len(c.donefile) != 0 {
		os.Create(c.donefile)
	}
	if exit, ok := err.(*exec.ExitError); ok {
		os.Exit(exit.ExitCode())
	}
	return err
}

func (c *entrypointCommand) wait(file string, expectContent bool) error {
	for ; ; time.Sleep(c.interval) {
		if _, err := os.Stat(file); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
		if _, err := os.Stat(file + ".err"); err == nil {
			return ErrSkip
		}
	}
}

// Register registers the entrypoint command.
func registerEntrypoint(app *kingpin.Application) {
	c := new(entrypointCommand)

	cmd := app.Command("entrypoint", "entrypoint override").
		Hidden().
		Action(c.run)

	cmd.Flag("name", "command name").
		StringVar(&c.name)

	cmd.Flag("arg", "command arg").
		StringsVar(&c.args)

	cmd.Flag("waitfile", "file to watch to start the command").
		Default("/tmp/drone/wait"). // etc/drone/%d
		StringVar(&c.waitfile)

	cmd.Flag("donefile", "file to write to when the command is done").
		StringVar(&c.donefile)

	cmd.Flag("interval", "interval to poll the wait file").
		Default("1s").
		DurationVar(&c.interval)
}
