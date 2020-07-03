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
	LimitCPU      int64
	LimitMemory   int64
	RequestCPU    int64
	RequestMemory int64
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
			Limits: compiler.ResourceObject{
				CPU:    c.LimitCPU,
				Memory: c.LimitMemory,
			},
			Requests: compiler.ResourceObject{
				CPU:    c.RequestCPU,
				Memory: c.RequestMemory,
			},
		},
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

	cmd.Flag("limit-cpu", "limit container cpu").
		Int64Var(&c.LimitCPU)

	cmd.Flag("limit-memory", "limit container memory").
		Int64Var(&c.LimitMemory)

	cmd.Flag("request-cpu", "request container cpu").
		Int64Var(&c.RequestCPU)

	cmd.Flag("request-memory", "request container memory").
		Int64Var(&c.RequestMemory)

		// shared pipeline flags
	c.Flags = internal.ParseFlags(cmd)
}
