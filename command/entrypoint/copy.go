// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package entrypoint

import (
	"io"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

type copyCommand struct {
	source string
	target string
}

func (c *copyCommand) run(*kingpin.ParseContext) error {
	return Copy(c.source, c.target)
}

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

// Register registers the copy command.
func registerCopy(app *kingpin.Application) {
	c := new(copyCommand)

	cmd := app.Command("copy", "entrypoint copy").
		Hidden().
		Action(c.run)

	cmd.Flag("source", "source binary path").
		Default("/bin/drone-runner-kube").
		StringVar(&c.source)

	cmd.Flag("target", "target binary path").
		Default("/usr/drone/drone-runner-kube").
		StringVar(&c.target)
}
