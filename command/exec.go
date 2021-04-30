// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/drone-runners/drone-runner-kube/command/internal"
	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/compiler"
	"github.com/drone-runners/drone-runner-kube/engine/linter"
	"github.com/drone-runners/drone-runner-kube/engine/policy"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone/drone-go/drone"
	"github.com/drone/envsubst"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/logger"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/pipeline/streamer/console"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"
	"github.com/drone/signal"

	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type execCommand struct {
	*internal.Flags

	Source     *os.File
	Include    []string
	Exclude    []string
	Privileged []string
	Volumes    map[string]string
	Environ    map[string]string
	Labels     map[string]string
	Secrets    map[string]string
	Namespace  string
	Config     string
	Policy     string
	Tmate      compiler.Tmate
	Clone      bool
	Pretty     bool
	Procs      int64
	Debug      bool
	Trace      bool
	Dump       bool
}

func (c *execCommand) run(*kingpin.ParseContext) error {
	rawsource, err := ioutil.ReadAll(c.Source)
	if err != nil {
		return err
	}

	kubeconfig := c.Config
	if kubeconfig == "" {
		dir, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(dir, ".kube", "config")
	}

	envs := environ.Combine(
		c.Environ,
		environ.System(c.System),
		environ.Repo(c.Repo),
		environ.Build(c.Build),
		environ.Stage(c.Stage),
		environ.Link(c.Repo, c.Build, c.System),
		c.Build.Params,
	)

	// parse the policy file
	var policies []*policy.Policy
	if c.Policy != "" {
		policies, err = policy.ParseFile(c.Policy)
		if err != nil {
			return err
		}
	}

	// string substitution function ensures that string
	// replacement variables are escaped and quoted if they
	// contain newlines.
	subf := func(k string) string {
		v := envs[k]
		if strings.Contains(v, "\n") {
			v = fmt.Sprintf("%q", v)
		}
		return v
	}

	// evaluates string replacement expressions and returns an
	// update configuration.
	config, err := envsubst.Eval(string(rawsource), subf)
	if err != nil {
		return err
	}

	// parse and lint the configuration.
	manifest, err := manifest.ParseString(config)
	if err != nil {
		return err
	}

	// a configuration can contain multiple pipelines.
	// get a specific pipeline resource for execution.
	resource, err := resource.Lookup(c.Stage.Name, manifest)
	if err != nil {
		return err
	}

	// lint the pipeline and return an error if any
	// linting rules are broken
	lint := linter.New(nil)
	err = lint.Lint(resource, c.Repo)
	if err != nil {
		return err
	}

	// compile the pipeline to an intermediate representation.
	comp := &compiler.Compiler{
		Environ:    provider.Static(c.Environ),
		Labels:     c.Labels,
		Tmate:      c.Tmate,
		Privileged: append(c.Privileged, compiler.Privileged...),
		Volumes:    c.Volumes,
		Secret:     secret.StaticVars(c.Secrets),
		Registry:   registry.Combine(),
		Namespace:  c.Namespace,
		Policies:   policies,
	}

	args := runtime.CompilerArgs{
		Pipeline: resource,
		Manifest: manifest,
		Build:    c.Build,
		Netrc:    c.Netrc,
		Repo:     c.Repo,
		Stage:    c.Stage,
		System:   c.System,
		Secret:   secret.StaticVars(c.Secrets),
	}
	spec := comp.Compile(nocontext, args).(*engine.Spec)

	// include only steps that are in the include list,
	// if the list in non-empty.
	if len(c.Include) > 0 {
	I:
		for _, step := range spec.Steps {
			if step.Name == "clone" {
				continue
			}
			for _, name := range c.Include {
				if step.Name == name {
					continue I
				}
			}
			step.RunPolicy = runtime.RunNever
		}
	}

	// exclude steps that are in the exclude list,
	// if the list in non-empty.
	if len(c.Exclude) > 0 {
	E:
		for _, step := range spec.Steps {
			if step.Name == "clone" {
				continue
			}
			for _, name := range c.Exclude {
				if step.Name == name {
					step.RunPolicy = runtime.RunNever
					continue E
				}
			}
		}
	}

	// create a step object for each pipeline step.
	for _, step := range spec.Steps {
		if step.RunPolicy == runtime.RunNever {
			continue
		}
		c.Stage.Steps = append(c.Stage.Steps, &drone.Step{
			StageID:   c.Stage.ID,
			Number:    len(c.Stage.Steps) + 1,
			Name:      step.Name,
			Status:    drone.StatusPending,
			ErrIgnore: step.ErrPolicy == runtime.ErrIgnore,
		})
	}

	// configures the pipeline timeout.
	timeout := time.Duration(c.Repo.Timeout) * time.Minute
	ctx, cancel := context.WithTimeout(nocontext, timeout)
	defer cancel()

	// listen for operating system signals and cancel execution
	// when received.
	ctx = signal.WithContextFunc(ctx, func() {
		println("received signal, terminating process")
		cancel()
	})

	state := &pipeline.State{
		Build:  c.Build,
		Stage:  c.Stage,
		Repo:   c.Repo,
		System: c.System,
	}
	state.Build.Status = drone.StatusRunning
	state.Stage.Status = drone.StatusRunning

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

	// change to out-of-cluster for local testing
	engine, err := engine.NewFromConfig(kubeconfig)
	if err != nil {
		return err
	}

	err = runtime.NewExecer(
		pipeline.NopReporter(),
		console.New(c.Pretty),
		engine,
		c.Procs,
	).Exec(ctx, spec, state)

	if c.Dump {
		dump(state)
	}
	if err != nil {
		return err
	}
	switch state.Stage.Status {
	case drone.StatusError, drone.StatusFailing, drone.StatusKilled:
		os.Exit(1)
	}
	return nil
}

func dump(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func registerExec(app *kingpin.Application) {
	c := new(execCommand)
	c.Environ = map[string]string{}
	c.Secrets = map[string]string{}
	c.Labels = map[string]string{}
	c.Volumes = map[string]string{}

	cmd := app.Command("exec", "executes a pipeline").
		Action(c.run)

	cmd.Arg("source", "source file location").
		Default(".drone.yml").
		FileVar(&c.Source)

	cmd.Flag("clone", "enable cloning").
		BoolVar(&c.Clone)

	cmd.Flag("secrets", "secret parameters").
		StringMapVar(&c.Secrets)

	cmd.Flag("include", "include pipeline steps").
		StringsVar(&c.Include)

	cmd.Flag("exclude", "exclude pipeline steps").
		StringsVar(&c.Exclude)

	cmd.Flag("environ", "environment variables").
		StringMapVar(&c.Environ)

	cmd.Flag("labels", "container labels").
		StringMapVar(&c.Labels)

	cmd.Flag("volumes", "container volumes").
		StringMapVar(&c.Volumes)

	cmd.Flag("privileged", "privileged docker images").
		StringsVar(&c.Privileged)

	cmd.Flag("kubeconfig", "path to the kubernetes config file").
		StringVar(&c.Config)

	cmd.Flag("policy", "path to the pipeline policy file").
		StringVar(&c.Policy)

	cmd.Flag("namespace", "default kubernetes namespace").
		Default("default").
		StringVar(&c.Namespace)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.Debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.Trace)

	cmd.Flag("dump", "dump the pipeline state to stdout").
		BoolVar(&c.Dump)

	cmd.Flag("pretty", "pretty print the output").
		Default(
			fmt.Sprint(
				isatty.IsTerminal(
					os.Stdout.Fd(),
				),
			),
		).BoolVar(&c.Pretty)

	cmd.Flag("tmate-image", "tmate docker image").
		Default("drone/drone-runner-docker:latest").
		StringVar(&c.Tmate.Image)

	cmd.Flag("tmate-enabled", "tmate enabled").
		BoolVar(&c.Tmate.Enabled)

	cmd.Flag("tmate-server-host", "tmate server host").
		StringVar(&c.Tmate.Server)

	cmd.Flag("tmate-server-port", "tmate server port").
		StringVar(&c.Tmate.Port)

	cmd.Flag("tmate-server-rsa-fingerprint", "tmate server rsa fingerprint").
		StringVar(&c.Tmate.RSA)

	cmd.Flag("tmate-server-ed25519-fingerprint", "tmate server rsa fingerprint").
		StringVar(&c.Tmate.ED25519)

	// shared pipeline flags
	c.Flags = internal.ParseFlags(cmd)
}
