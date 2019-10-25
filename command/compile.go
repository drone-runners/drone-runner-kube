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
	"github.com/drone-runners/drone-runner-kube/engine/compiler"
	"github.com/drone-runners/drone-runner-kube/engine/linter"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone/envsubst"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"

	"gopkg.in/alecthomas/kingpin.v2"
)

type compileCommand struct {
	*internal.Flags

	Source     *os.File
	Privileged []string
	Networks   []string
	Volumes    map[string]string
	Environ    map[string]string
	Labels     map[string]string
	Secrets    map[string]string
	Resources  compiler.Resources
	Clone      bool
	Config     string
}

func (c *compileCommand) run(*kingpin.ParseContext) error {
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
	lint := linter.New()
	opts := linter.Opts{Trusted: c.Repo.Trusted}
	err = lint.Lint(resource, opts)
	if err != nil {
		return err
	}

	// compile the pipeline to an intermediate representation.
	comp := &compiler.Compiler{
		Environ:    c.Environ,
		Labels:     c.Labels,
		Resources:  c.Resources,
		Privileged: append(c.Privileged, compiler.Privileged...),
		Networks:   c.Networks,
		Volumes:    c.Volumes,
		Secret:     secret.StaticVars(c.Secrets),
		Registry: registry.Combine(
			registry.File(c.Config),
		),
	}

	// when running a build locally cloning is always
	// disabled in favor of mounting the source code
	// from the current working directory.
	if c.Clone == false {
		comp.Mount, _ = os.Getwd()
	}

	args := compiler.Args{
		Pipeline: resource,
		Manifest: manifest,
		Build:    c.Build,
		Netrc:    c.Netrc,
		Repo:     c.Repo,
		Stage:    c.Stage,
		System:   c.System,
	}
	spec := comp.Compile(nocontext, args)

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

	cmd.Flag("networks", "container networks").
		StringsVar(&c.Networks)

	cmd.Flag("volumes", "container volumes").
		StringMapVar(&c.Volumes)

	cmd.Flag("privileged", "privileged docker images").
		StringsVar(&c.Privileged)

	cmd.Flag("cpu-period", "container cpu period").
		Int64Var(&c.Resources.CPUPeriod)

	cmd.Flag("cpu-quota", "container cpu quota").
		Int64Var(&c.Resources.CPUQuota)

	cmd.Flag("cpu-set", "container cpu set").
		StringsVar(&c.Resources.CPUSet)

	cmd.Flag("cpu-shares", "container cpu shares").
		Int64Var(&c.Resources.CPUShares)

	cmd.Flag("memory", "container memory limit").
		Int64Var(&c.Resources.MemLimit)

	cmd.Flag("memory-swap", "container memory swap limit").
		Int64Var(&c.Resources.MemSwapLimit)

	cmd.Flag("shmsize", "container shm size").
		Int64Var(&c.Resources.ShmSize)

	cmd.Flag("docker-config", "path to the docker config file").
		StringVar(&c.Config)

	// shared pipeline flags
	c.Flags = internal.ParseFlags(cmd)
}
