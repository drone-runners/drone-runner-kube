// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/drone-runners/drone-runner-kube/command/internal"
	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/compiler"
	"github.com/drone-runners/drone-runner-kube/engine/linter"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone/envsubst"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"

	"gopkg.in/alecthomas/kingpin.v2"
)

type compileCommand struct {
	*internal.Flags

	Source        *os.File
	Privileged    []string
	Volumes       map[string]string
	Environ       map[string]string
	Labels        map[string]string
	Secrets       map[string]string
	Clone         bool
	Spec          bool
	Config        string
	Resource      compiler.Resources
	StageRequests compiler.ResourceObject
	Tmate         compiler.Tmate
}

func (c *compileCommand) run(*kingpin.ParseContext) error {
	// resource memory amounts are provided in megabytes, so convert them to bytes.
	c.Resource.Limits.Memory *= 1024 * 1024
	c.Resource.MinRequests.Memory *= 1024 * 1024
	c.StageRequests.Memory *= 1024 * 1024

	rawsource, err := ioutil.ReadAll(c.Source)
	if err != nil {
		return err
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

	// parse and lint the configuration
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
		Privileged: append(c.Privileged, compiler.Privileged...),
		Volumes:    c.Volumes,
		Secret:     secret.Combine(),
		Registry:   registry.Combine(),
		Resources: compiler.Resources{
			Limits:      c.Resource.Limits,
			MinRequests: c.Resource.MinRequests,
		},
		StageRequests: c.StageRequests,
		Tmate:         c.Tmate,
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
	spec := comp.Compile(nocontext, args)

	if c.Spec {
		engine.Dump(os.Stdout, spec.(*engine.Spec))
		return nil
	}

	// encode the pipeline in json format and print to the
	// console for inspection.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(spec)
	return nil
}

func registerCompile(app *kingpin.Application) {
	c := new(compileCommand)
	c.Environ = map[string]string{}
	c.Secrets = map[string]string{}
	c.Labels = map[string]string{}
	c.Volumes = map[string]string{}

	cmd := app.Command("compile", "compile the yaml file").
		Action(c.run)

	cmd.Flag("source", "source file location").
		Default(".drone.yml").
		FileVar(&c.Source)

	cmd.Flag("clone", "enable cloning").
		BoolVar(&c.Clone)

	cmd.Flag("secrets", "secret parameters").
		StringMapVar(&c.Secrets)

	cmd.Flag("environ", "environment variables").
		StringMapVar(&c.Environ)

	cmd.Flag("labels", "container labels").
		StringMapVar(&c.Labels)

	cmd.Flag("volumes", "container volumes").
		StringMapVar(&c.Volumes)

	cmd.Flag("privileged", "privileged docker images").
		StringsVar(&c.Privileged)

	cmd.Flag("spec", "output the kubernetes spec").
		BoolVar(&c.Spec)

	cmd.Flag("limit-memory", "memory limit in MiB for containers").
		Int64Var(&c.Resource.Limits.Memory)

	cmd.Flag("limit-cpu", "cpu limit in millicores for containers").
		Int64Var(&c.Resource.Limits.CPU)

	cmd.Flag("request-memory", "memory in MiB for entire pod").
		Default("100"). // Default is 100MiB
		Int64Var(&c.StageRequests.Memory)

	cmd.Flag("request-cpu", "cpu in millicores for entire pod").
		Default("100").
		Int64Var(&c.StageRequests.CPU)

	cmd.Flag("min-request-memory", "min memory in MiB allocated to each container").
		Default("4"). // Default is 4MiB
		Int64Var(&c.Resource.MinRequests.Memory)

	cmd.Flag("min-request-cpu", "min cpu in millicores allocated to each container").
		Default("1").
		Int64Var(&c.Resource.MinRequests.CPU)

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
