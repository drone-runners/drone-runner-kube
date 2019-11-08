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

	"github.com/asaskevich/govalidator"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/clone"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/labels"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/registry/auths"
	"github.com/drone/runner-go/secret"

	"github.com/dchest/uniuri"
	"github.com/gosimple/slug"
)

// random generator function
var random = func() string {
	return "drone-" + uniuri.NewLenChars(20, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))
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

type (
	// Resources describes the compute resource requirements.
	Resources struct {
		Limits   ResourceObject
		Requests ResourceObject
	}

	// ResourceObject describes compute resource requirements.
	ResourceObject struct {
		CPU    int64
		Memory int64
	}

	// Args provides compiler arguments.
	Args struct {
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
	Compiler struct {

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

		// Resources defines resource limits that are applied by
		// default to all pipeline containers if none exist.
		Resources Resources

		// Cloner provides an option to override the default clone
		// image used to clone the repository when the pipeline
		// initializes.
		Cloner string

		// Placeholder provides the default placeholder image
		// used to sleep the pipeline container until it is ready
		// for execution.
		Placeholder string

		// Namespace provides the default kubernetes namespace
		// when no namespace is provided.
		Namespace string

		// ServiceAccount provides the default kubernetes Service Account
		// when no Service Account is provided.
		ServiceAccount string
	}
)

// Compile compiles the configuration file.
func (c *Compiler) Compile(ctx context.Context, args Args) *engine.Spec {
	os := args.Pipeline.Platform.OS
	arch := args.Pipeline.Platform.Arch

	// create the workspace paths
	workspace := createWorkspace(args.Pipeline)

	// create system labels
	annotations := labels.Combine(
		c.Labels,
		labels.FromRepo(args.Repo),
		labels.FromBuild(args.Build),
		labels.FromStage(args.Stage),
		labels.FromSystem(args.System),
		labels.WithTimeout(args.Repo),
		args.Pipeline.Metadata.Labels,
	)

	// create the workspace mount
	workMount := &engine.VolumeMount{
		Name: "_workspace",
		Path: workspace,
	}

	// create the workspace volume
	workVolume := &engine.Volume{
		EmptyDir: &engine.VolumeEmptyDir{
			ID:   random(),
			Name: workMount.Name,
		},
	}

	// create the statuses volume
	statusMount := &engine.VolumeMount{
		Name: "_status",
		Path: "/run/drone",
	}

	// create the statuses DownwardAPI volume
	statusVolume := &engine.Volume{
		DownwardAPI: &engine.VolumeDownwardAPI{
			ID:   random(),
			Name: statusMount.Name,
			Items: []engine.VolumeDownwardAPIItem{
				{
					Path:      "env",
					FieldPath: "metadata.annotations",
				},
			},
		},
	}

	spec := &engine.Spec{
		PodSpec: engine.PodSpec{
			Name:               random(),
			Namespace:          args.Pipeline.Metadata.Namespace,
			Labels:             args.Pipeline.Metadata.Labels,
			Annotations:        labels.Combine(args.Pipeline.Metadata.Annotations, annotations),
			NodeSelector:       args.Pipeline.NodeSelector,
			ServiceAccountName: args.Pipeline.ServiceAccountName,
		},
		Platform: engine.Platform{
			OS:      args.Pipeline.Platform.OS,
			Arch:    args.Pipeline.Platform.Arch,
			Variant: args.Pipeline.Platform.Variant,
			Version: args.Pipeline.Platform.Version,
		},
		Secrets: map[string]*engine.Secret{},
		Envs:    map[string]string{},
		Volumes: []*engine.Volume{workVolume, statusVolume},
	}

	// set default namespace and ensure maps are non-nil
	if spec.PodSpec.Namespace == "" {
		spec.PodSpec.Namespace = c.Namespace
	}
	if spec.PodSpec.Labels == nil {
		spec.PodSpec.Labels = map[string]string{}
	}
	if spec.PodSpec.Annotations == nil {
		spec.PodSpec.Annotations = map[string]string{}
	}
	// set default service account
	if spec.PodSpec.ServiceAccountName == "" {
		spec.PodSpec.ServiceAccountName = c.ServiceAccount
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
	if workVolume.EmptyDir != nil {
		envs["DRONE_DOCKER_VOLUME_ID"] = workVolume.EmptyDir.ID
	} else {
		envs["DRONE_DOCKER_VOLUME_PATH"] = workVolume.HostPath.Path
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

	spec.Envs = envs

	// set platform if needed
	if arch == "arm" || arch == "arm64" {
		spec.PodSpec.Labels["kubernetes.io/arch"] = arch
	}

	// set drone labels
	spec.PodSpec.Labels["io.drone"] = "true"
	spec.PodSpec.Labels["io.drone.repo.namespace"] = slug.Make(args.Repo.Namespace)
	spec.PodSpec.Labels["io.drone.repo.name"] = slug.Make(args.Repo.Name)
	spec.PodSpec.Labels["io.drone.build.number"] = fmt.Sprint(args.Build.Number)
	spec.PodSpec.Labels["io.drone.build.event"] = slug.Make(args.Build.Event)

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
		step.WorkingDir = workspace
		step.Volumes = append(step.Volumes, workMount, statusMount)
		spec.Steps = append(spec.Steps, step)

		// override default clone image.
		if c.Cloner != "" {
			step.Image = c.Cloner
		}

		// override default placeholder image.
		if c.Placeholder != "" {
			step.Placeholder = c.Placeholder
		}
	}

	var hostnames []string

	// create steps
	for _, src := range args.Pipeline.Services {
		dst := createStep(args.Pipeline, src)
		dst.Detach = true
		dst.Volumes = append(dst.Volumes, workMount, statusMount)
		setupScript(src, dst, os)
		setupWorkdir(src, dst, workspace)
		spec.Steps = append(spec.Steps, dst)

		// if the pipeline step has unmet conditions the step is
		// automatically skipped.
		if !src.When.Match(match) {
			dst.RunPolicy = engine.RunNever
		}

		// override default placeholder image.
		if c.Placeholder != "" {
			dst.Placeholder = c.Placeholder
		}

		if govalidator.IsHost(src.Name) {
			hostnames = append(hostnames, src.Name)
		}
	}

	if len(hostnames) > 0 {
		spec.PodSpec.HostAliases = []engine.HostAlias{
			{
				IP:        "127.0.0.1",
				Hostnames: hostnames,
			},
		}
	}

	// create steps
	for _, src := range args.Pipeline.Steps {
		dst := createStep(args.Pipeline, src)
		dst.Volumes = append(dst.Volumes, workMount, statusMount)
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
			// if the secret was already fetched and stored in the
			// secret map it can be skipped.
			if _, ok := spec.Secrets[s.Name]; ok {
				continue
			}
			secret, ok := c.findSecret(ctx, args, s.Name)
			if ok {
				spec.Secrets[s.Name] = &engine.Secret{
					Name: s.Name,
					Data: secret,
					Mask: true,
				}
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

	// if registry credentials provides, encode the credentials
	// as a docker-config-json file and create secret.
	if len(creds) != 0 {
		spec.PullSecret = &engine.Secret{
			Name: random(),
			Data: auths.Encode(creds...),
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

	// apply default resources limits
	for _, v := range spec.Steps {
		if v.Resources.Requests.CPU == 0 {
			v.Resources.Requests.CPU = c.Resources.Requests.CPU
		}
		if v.Resources.Requests.Memory == 0 {
			v.Resources.Requests.Memory = c.Resources.Requests.Memory
		}
		if v.Resources.Limits.CPU == 0 {
			v.Resources.Limits.CPU = c.Resources.Limits.CPU
		}
		if v.Resources.Limits.Memory == 0 {
			v.Resources.Limits.Memory = c.Resources.Limits.Memory
		}
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
