// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"context"
	"fmt"

	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone-runners/drone-runner-kube/internal/docker/image"

	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/clone"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/labels"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/registry/auths"
	"github.com/drone/runner-go/secret"

	"github.com/dchest/uniuri"
)

// random generator function
var random = func() string {
	return uniuri.NewLenChars(20, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))
}

// Privileged provides a list of plugins that execute
// with privileged capabilities in order to run Docker
// in Docker.
var Privileged = []string{
	"plugins/docker",
	"plugins/acr",
	"plugins/ecr",
	"plugins/gcr",
	"plugins/heroku",
}

// Args provides compiler arguments.
type Args struct {
	// Manifest provides the parsed manifest.
	Manifest *manifest.Manifest

	// Pipeline provides the parsed pipeline. This pipeline is
	// the compiler source and is converted to the intermediate
	// representation by the Compile method.
	Pipeline *resource.Pipeline

	// Build provides the compiler with stage information that
	// is converted to environment variable format and passed to
	// each pipeline step. It is also used to clone the commit.
	Build *drone.Build

	// Stage provides the compiler with stage information that
	// is converted to environment variable format and passed to
	// each pipeline step.
	Stage *drone.Stage

	// Repo provides the compiler with repo information. This
	// repo information is converted to environment variable
	// format and passed to each pipeline step. It is also used
	// to clone the repository.
	Repo *drone.Repo

	// System provides the compiler with system information that
	// is converted to environment variable format and passed to
	// each pipeline step.
	System *drone.System

	// Netrc provides netrc parameters that can be used by the
	// default clone step to authenticate to the remote
	// repository.
	Netrc *drone.Netrc

	// Secret returns a named secret value that can be injected
	// into the pipeline step.
	Secret secret.Provider
}

// Compiler compiles the Yaml configuration file to an
// intermediate representation optimized for simple execution.
type Compiler struct {

	// Environ provides a set of environment variables that
	// should be added to each pipeline step by default.
	Environ map[string]string

	// Labels provides a set of labels that should be added
	// to each container by default.
	Labels map[string]string

	// Privileged provides a list of docker images that
	// are always privileged.
	Privileged []string

	// Secret returns a named secret value that can be injected
	// into the pipeline step.
	Secret secret.Provider

	// Registry returns a list of registry credentials that can be
	// used to pull private container images.
	Registry registry.Provider

	// Cloner provides an option to override the default clone
	// image used to clone the repository when the pipeline
	// initializes.
	Cloner string
}

// Compile compiles the configuration file.
func (c *Compiler) Compile(ctx context.Context, args Args) *engine.Spec {
	os := args.Pipeline.Platform.OS

	// create the workspace paths
	workspace := createWorkspace(args.Pipeline)

	// create system labels
	labels := labels.Combine(
		c.Labels,
		labels.FromRepo(args.Repo),
		labels.FromBuild(args.Build),
		labels.FromStage(args.Stage),
		labels.FromSystem(args.System),
		labels.WithTimeout(args.Repo),
		args.Pipeline.Metadata.Labels,
	)

	// create the workspace mount
	mount := &engine.VolumeMount{
		Name: "_workspace",
		Path: workspace,
	}

	// create the workspace volume
	volume := &engine.Volume{
		EmptyDir: &engine.VolumeEmptyDir{
			ID:     random(),
			Name:   mount.Name,
			Labels: labels,
		},
	}

	spec := &engine.Spec{
		PodSpec: engine.PodSpec{
			Name:               random(),
			Namespace:          args.Pipeline.Metadata.Namespace,
			Labels:             args.Pipeline.Metadata.Labels,
			Annotations:        args.Pipeline.Metadata.Annotations,
			NodeSelector:       args.Pipeline.NodeSelector,
			ServiceAccountName: args.Pipeline.ServiceAccountName,
		},
		Platform: engine.Platform{
			OS:      args.Pipeline.Platform.OS,
			Arch:    args.Pipeline.Platform.Arch,
			Variant: args.Pipeline.Platform.Variant,
			Version: args.Pipeline.Platform.Version,
		},
		Volumes: []*engine.Volume{volume},
	}

	// add tolerations
	for _, toleration := range args.Pipeline.Tolerations {
		spec.PodSpec.Tolerations = append(spec.PodSpec.Tolerations, engine.Toleration{
			Operator:          toleration.Operator,
			Effect:            toleration.Effect,
			TolerationSeconds: toleration.TolerationSeconds,
			Value:             toleration.Value,
		})
	}

	// create the default environment variables.
	envs := environ.Combine(
		c.Environ,
		args.Build.Params,
		args.Pipeline.Environment,
		environ.Proxy(),
		environ.System(args.System),
		environ.Repo(args.Repo),
		environ.Build(args.Build),
		environ.Stage(args.Stage),
		environ.Link(args.Repo, args.Build, args.System),
		clone.Environ(clone.Config{
			SkipVerify: args.Pipeline.Clone.SkipVerify,
			Trace:      args.Pipeline.Clone.Trace,
			User: clone.User{
				Name:  args.Build.AuthorName,
				Email: args.Build.AuthorEmail,
			},
		}),
	)

	// create the workspace variables
	envs["DRONE_WORKSPACE"] = workspace

	// create volume reference variables
	if volume.EmptyDir != nil {
		envs["DRONE_DOCKER_VOLUME_ID"] = volume.EmptyDir.ID
	} else {
		envs["DRONE_DOCKER_VOLUME_PATH"] = volume.HostPath.Path
	}

	// create the netrc environment variables
	if args.Netrc != nil && args.Netrc.Machine != "" {
		envs["DRONE_NETRC_MACHINE"] = args.Netrc.Machine
		envs["DRONE_NETRC_USERNAME"] = args.Netrc.Login
		envs["DRONE_NETRC_PASSWORD"] = args.Netrc.Password
		envs["DRONE_NETRC_FILE"] = fmt.Sprintf(
			"machine %s login %s password %s",
			args.Netrc.Machine,
			args.Netrc.Login,
			args.Netrc.Password,
		)
	}

	match := manifest.Match{
		Action:   args.Build.Action,
		Cron:     args.Build.Cron,
		Ref:      args.Build.Ref,
		Repo:     args.Repo.Slug,
		Instance: args.System.Host,
		Target:   args.Build.Deploy,
		Event:    args.Build.Event,
		Branch:   args.Build.Target,
	}

	// create the clone step
	if args.Pipeline.Clone.Disable == false {
		step := createClone(args.Pipeline)
		step.ID = random()
		step.Envs = environ.Combine(envs, step.Envs)
		step.WorkingDir = workspace
		step.Labels = labels
		step.Volumes = append(step.Volumes, mount)
		spec.Steps = append(spec.Steps, step)

		// override default clone image.
		if c.Cloner != "" {
			step.Image = c.Cloner
		}
	}

	// create steps
	for _, src := range args.Pipeline.Services {
		dst := createStep(args.Pipeline, src)
		dst.Detach = true
		dst.Envs = environ.Combine(envs, dst.Envs)
		dst.Volumes = append(dst.Volumes, mount)
		dst.Labels = labels
		setupScript(src, dst, os)
		setupWorkdir(src, dst, workspace)
		spec.Steps = append(spec.Steps, dst)

		// if the pipeline step has unmet conditions the step is
		// automatically skipped.
		if !src.When.Match(match) {
			dst.RunPolicy = engine.RunNever
		}
	}

	// create steps
	for _, src := range args.Pipeline.Steps {
		dst := createStep(args.Pipeline, src)
		dst.Envs = environ.Combine(envs, dst.Envs)
		dst.Volumes = append(dst.Volumes, mount)
		dst.Labels = labels
		setupScript(src, dst, os)
		setupWorkdir(src, dst, workspace)
		spec.Steps = append(spec.Steps, dst)

		// if the pipeline step has unmet conditions the step is
		// automatically skipped.
		if !src.When.Match(match) {
			dst.RunPolicy = engine.RunNever
		}

		// if the pipeline step has an approved image, it is
		// automatically defaulted to run with escalalated
		// privileges.
		if c.isPrivileged(src) {
			dst.Privileged = true
		}
	}

	if isGraph(spec) == false {
		configureSerial(spec)
	} else if args.Pipeline.Clone.Disable == false {
		configureCloneDeps(spec)
	} else if args.Pipeline.Clone.Disable == true {
		removeCloneDeps(spec)
	}

	for _, step := range spec.Steps {
		for _, s := range step.Secrets {
			secret, ok := c.findSecret(ctx, args, s.Name)
			if ok {
				s.Data = []byte(secret)
			}
		}
	}

	// get registry credentials from registry plugins
	creds, err := c.Registry.List(ctx, &registry.Request{
		Repo:  args.Repo,
		Build: args.Build,
	})
	if err != nil {
		// TODO (bradrydzewski) return an error to the caller
		// if the provider returns an error.
	}

	// get registry credentials from secrets
	for _, name := range args.Pipeline.PullSecrets {
		secret, ok := c.findSecret(ctx, args, name)
		if ok {
			parsed, err := auths.ParseString(secret)
			if err == nil {
				creds = append(creds, parsed...)
			}
		}
	}

	for _, step := range spec.Steps {
	STEPS:
		for _, cred := range creds {
			if image.MatchHostname(step.Image, cred.Address) {
				step.Auth = &engine.Auth{
					Address:  cred.Address,
					Username: cred.Username,
					Password: cred.Password,
				}
				break STEPS
			}
		}
	}

	// append volumes
	for _, v := range args.Pipeline.Volumes {
		id := random()
		src := new(engine.Volume)
		if v.EmptyDir != nil {
			src.EmptyDir = &engine.VolumeEmptyDir{
				ID:        id,
				Name:      v.Name,
				Medium:    v.EmptyDir.Medium,
				SizeLimit: int64(v.EmptyDir.SizeLimit),
				Labels:    labels,
			}
		} else if v.HostPath != nil {
			src.HostPath = &engine.VolumeHostPath{
				ID:   id,
				Name: v.Name,
				Path: v.HostPath.Path,
			}
		} else {
			continue
		}
		spec.Volumes = append(spec.Volumes, src)
	}

	return spec
}

func (c *Compiler) isPrivileged(step *resource.Step) bool {
	// privileged-by-default containers are only
	// enabled for plugins steps that do not define
	// commands, command, or entrypoint.
	if len(step.Commands) > 0 {
		return false
	}
	if len(step.Command) > 0 {
		return false
	}
	if len(step.Entrypoint) > 0 {
		return false
	}
	// if the container image matches any image
	// in the whitelist, return true.
	for _, img := range c.Privileged {
		a := img
		b := step.Image
		if image.Match(a, b) {
			return true
		}
	}
	return false
}

// helper function attempts to find and return the named secret.
// from the secret provider.
func (c *Compiler) findSecret(ctx context.Context, args Args, name string) (s string, ok bool) {
	if name == "" {
		return
	}

	// source secrets from the global secret provider
	// and the repository secret provider.
	provider := secret.Combine(
		args.Secret,
		c.Secret,
	)

	// TODO return an error to the caller if the provider
	// returns an error.
	found, _ := provider.Find(ctx, &secret.Request{
		Name:  name,
		Build: args.Build,
		Repo:  args.Repo,
		Conf:  args.Manifest,
	})
	if found == nil {
		return
	}
	return found.Data, true
}
