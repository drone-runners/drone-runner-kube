// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package command

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/client"
	"github.com/drone/runner-go/logger"
	"github.com/drone/runner-go/pipeline"
	"github.com/drone/runner-go/pipeline/reporter/remote"
	"github.com/drone/runner-go/pipeline/runtime"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Config stores the system configuration.
type ClientConfig struct {
	Address    string `json:"address"`
	Proto      string `json:"proto"`
	Host       string `json:"host"`
	Secret     string `json:"secret"`
	SkipVerify bool   `json:"skip_verify"`
	Dump       bool   `json:"dump_http"`
	DumpBody   bool   `json:"dump_http_body"`
}

type exec2Command struct {
	Debug bool
	Trace bool
}

func (c *exec2Command) run(*kingpin.ParseContext) error {
	//
	//
	//
	//

	out, err := ioutil.ReadFile("/etc/drone/spec.json")
	if err != nil {
		return err
	}

	spec := new(engine.Spec)
	if err := json.Unmarshal(out, spec); err != nil {
		return err
	}

	//
	//
	//
	//

	out, err = ioutil.ReadFile("/etc/drone/spec.json")
	if err != nil {
		return err
	}

	state := new(pipeline.State)
	if err := json.Unmarshal(out, spec); err != nil {
		return err
	}

	//
	//
	//
	//

	config := new(ClientConfig)
	if err := json.Unmarshal(out, config); err != nil {
		return err
	}

	out, err = ioutil.ReadFile("/etc/drone/client.json")
	if err != nil {
		return err
	}

	//
	//
	//
	//

	// enable debug logging
	logrus.SetLevel(logrus.WarnLevel)
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if c.Trace {
		logrus.SetLevel(logrus.TraceLevel)
	}
	logger.Default = logger.Logrus(
		logrus.NewEntry(
			logrus.StandardLogger(),
		),
	)

	//
	//
	//

	log := logger.FromContext(context.Background()).
		WithField("repo.id", state.Repo.ID).
		WithField("repo.name", state.Repo.ID).
		WithField("stage.id", state.Stage.ID).
		WithField("stage.name", state.Stage.Name).
		WithField("stage.number", state.Stage.Number).
		WithField("repo.namespace", state.Repo.Namespace).
		WithField("repo.name", state.Repo.Name).
		WithField("build.id", state.Build.ID).
		WithField("build.number", state.Build.Number)

	//
	//
	//
	//

	engine, err := engine.NewInCluster()
	if err != nil {
		return err
	}

	//
	//
	//

	cli := client.New(
		config.Address,
		config.Secret,
		config.SkipVerify,
	)

	remote := remote.New(cli)

	//
	//
	//

	ctxdone, cancel := context.WithCancel(context.Background())
	defer cancel()

	timeout := time.Duration(state.Repo.Timeout) * time.Minute
	ctxtimeout, cancel := context.WithTimeout(ctxdone, timeout)
	defer cancel()

	ctxcancel, cancel := context.WithCancel(ctxtimeout)
	defer cancel()

	// next we opens a connection to the server to watch for
	// cancellation requests. If a build is cancelled the running
	// stage should also be cancelled.
	go func() {
		done, _ := cli.Watch(ctxdone, state.Build.ID)
		if done {
			cancel()
			log.Debugln("received cancellation")
		} else {
			log.Debugln("done listening for cancellations")
		}
	}()

	state.Stage.Started = time.Now().Unix()
	state.Stage.Status = drone.StatusRunning
	if err := cli.Update(ctxdone, state.Stage); err != nil {
		log.WithError(err).Error("cannot update stage")
		return err
	}

	log.Debug("updated stage to running")

	ctxcancel = logger.WithContext(ctxcancel, logger.Default)
	return runtime.NewExecer(remote, remote, engine, 0).
		Exec(ctxcancel, spec, state)
}

func registerExec2(app *kingpin.Application) {
	c := new(exec2Command)

	cmd := app.Command("controller", "executes a pipeline").
		Hidden().
		Action(c.run)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.Debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.Trace)
}
