// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/policy"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone-runners/drone-runner-kube/internal/docker/image"

	"github.com/drone/runner-go/clone"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/labels"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline/runtime"
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

	// Compiler compiles the Yaml configuration file to an
	// intermediate representation optimized for simple execution.
	Compiler struct {
		// Environ provides a set of environment variables that
		// should be added to each pipeline step by default.
		Environ provider.Provider

		// Labels provides a set of labels that should be added
		// to each container by default.
		Labels map[string]string

		// Annotations provides a set of annotations that should be added
		// to each container by default.
		Annotations map[string]string

		// Privileged provides a list of docker images that
		// are always privileged.
		Privileged []string

		// PullPolicy provides a docker image pull policy
		// to use instead of pipeline's ones.
		PullPolicy string

		// Volumes provides a set of volumes that should be
		// mounted to each pipeline container.
		Volumes map[string]string

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

		// NodeSelector provides the default kubernetes node selector.
		NodeSelector map[string]string

		// Policy provides a set of policies used to set defaults
		// based on matching logic.
		Policies []*policy.Policy
	}
)

// Compile compiles the configuration file.
func (c *Compiler) Compile(ctx context.Context, args runtime.CompilerArgs) runtime.Spec {
	pipeline := args.Pipeline.(*resource.Pipeline)
	os := pipeline.Platform.OS
	arch := pipeline.Platform.Arch

	// create the workspace paths
	workspace := createWorkspace(pipeline)

	// create labels
	podLabels := labels.Combine(
		c.Labels,
		pipeline.Metadata.Labels,
	)

	// create annotations
	podAnnotations := labels.Combine(
		c.Annotations,
		labels.FromRepo(args.Repo),
		labels.FromBuild(args.Build),
		labels.FromStage(args.Stage),
		labels.FromSystem(args.System),
		labels.WithTimeout(args.Repo),
		pipeline.Metadata.Annotations,
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
			Namespace:          pipeline.Metadata.Namespace,
			Labels:             podLabels,
			Annotations:        podAnnotations,
			NodeName:           pipeline.NodeName,
			NodeSelector:       pipeline.NodeSelector,
			ServiceAccountName: pipeline.ServiceAccountName,
		},
		Platform: engine.Platform{
			OS:      pipeline.Platform.OS,
			Arch:    pipeline.Platform.Arch,
			Variant: pipeline.Platform.Variant,
			Version: pipeline.Platform.Version,
		},
		Secrets: map[string]*engine.Secret{},
		Volumes: []*engine.Volume{workVolume, statusVolume},
	}

	// set default namespace
	if spec.PodSpec.Namespace == "" {
		spec.PodSpec.Namespace = c.Namespace
	}

	// the runner can be configured to create random namespaces
	// that is created before the pipeline executes, and destroyed
	// after the pipeline completes.
	if spec.PodSpec.Namespace == "drone-" {
		namespace := random()
		spec.PodSpec.Namespace = namespace
		spec.Namespace = namespace
	}

	// ensure maps are non-nil
	if spec.PodSpec.Labels == nil {
		spec.PodSpec.Labels = map[string]string{}
	}
	if spec.PodSpec.Annotations == nil {
		spec.PodSpec.Annotations = map[string]string{}
	}
	if spec.PodSpec.NodeSelector == nil && c.NodeSelector != nil {
		spec.PodSpec.NodeSelector = c.NodeSelector
	}
	// set default service account
	if spec.PodSpec.ServiceAccountName == "" {
		spec.PodSpec.ServiceAccountName = c.ServiceAccount
	}
	// add dns_config
	if len(pipeline.DnsConfig.Nameservers) > 0 {
		spec.PodSpec.DnsConfig.Nameservers = pipeline.DnsConfig.Nameservers
	}

	if len(pipeline.DnsConfig.Searches) > 0 {
		spec.PodSpec.DnsConfig.Searches = pipeline.DnsConfig.Searches
	}

	for _, dnsconfig := range pipeline.DnsConfig.Options {
		spec.PodSpec.DnsConfig.Options = append(spec.PodSpec.DnsConfig.Options, engine.DNSConfigOptions{
			Name:  dnsconfig.Name,
			Value: dnsconfig.Value,
		})
	}
	// add tolerations
	for _, toleration := range pipeline.Tolerations {
		spec.PodSpec.Tolerations = append(spec.PodSpec.Tolerations, engine.Toleration{
			Key:               toleration.Key,
			Operator:          toleration.Operator,
			Effect:            toleration.Effect,
			TolerationSeconds: toleration.TolerationSeconds,
			Value:             toleration.Value,
		})
	}

	// list the global environment variables
	globals, _ := c.Environ.List(ctx, &provider.Request{
		Build: args.Build,
		Repo:  args.Repo,
	})

	// create the default environment variables.
	envs := environ.Combine(
		provider.ToMap(globals),
		args.Build.Params,
		pipeline.Environment,
		environ.Proxy(),
		environ.System(args.System),
		environ.Repo(args.Repo),
		environ.Build(args.Build),
		environ.Stage(args.Stage),
		environ.Link(args.Repo, args.Build, args.System),
		clone.Environ(clone.Config{
			SkipVerify: pipeline.Clone.SkipVerify,
			Trace:      pipeline.Clone.Trace,
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

	// set platform if needed
	if arch == "arm" || arch == "arm64" {
		spec.PodSpec.Labels["kubernetes.io/arch"] = arch
	}

	// set drone labels
	spec.PodSpec.Labels["io.drone"] = "true"
	spec.PodSpec.Labels["io.drone.name"] = spec.PodSpec.Name
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
	if pipeline.Clone.Disable == false {
		step := createClone(pipeline)
		step.ID = random()
		step.Envs = environ.Combine(envs, step.Envs)
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
	for _, src := range pipeline.Services {
		dst := createStep(pipeline, src)
		dst.Detach = true
		dst.Envs = environ.Combine(envs, dst.Envs)
		dst.Volumes = append(dst.Volumes, workMount, statusMount)
		setupScript(src, dst, os)
		setupWorkdir(src, dst, workspace)
		spec.Steps = append(spec.Steps, dst)

		// if the pipeline step has unmet conditions the step is
		// automatically skipped.
		if !src.When.Match(match) {
			dst.RunPolicy = runtime.RunNever
		}

		// override default placeholder image.
		if c.Placeholder != "" {
			dst.Placeholder = c.Placeholder
		}

		if len(validation.IsDNS1123Subdomain(src.Name)) == 0 {
			hostnames = append(hostnames, src.Name)
		}

		if c.isPrivileged(src) {
			dst.Privileged = true
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
	for _, src := range pipeline.Steps {
		dst := createStep(pipeline, src)
		dst.Envs = environ.Combine(envs, dst.Envs)
		dst.Volumes = append(dst.Volumes, workMount, statusMount)
		setupScript(src, dst, os)
		setupWorkdir(src, dst, workspace)
		spec.Steps = append(spec.Steps, dst)

		// if the pipeline step has unmet conditions the step is
		// automatically skipped.
		if !src.When.Match(match) {
			dst.RunPolicy = runtime.RunNever
		}

		// override default placeholder image.
		if c.Placeholder != "" {
			dst.Placeholder = c.Placeholder
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
	} else if pipeline.Clone.Disable == false {
		configureCloneDeps(spec)
	} else if pipeline.Clone.Disable == true {
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
				s := &engine.Secret{
					Name: s.Name,
					Data: secret,
					Mask: true,
				}
				spec.Secrets[s.Name] = s
				step.SpecSecrets = append(step.SpecSecrets, s)
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
	for _, name := range pipeline.PullSecrets {
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

	// append global volumes to the steps.
	for k, v := range c.Volumes {
		id := random()
		ro := strings.HasSuffix(v, ":ro")
		v = strings.TrimSuffix(v, ":ro")
		volume := &engine.Volume{
			HostPath: &engine.VolumeHostPath{
				ID:   id,
				Name: id,
				Path: k,
			},
		}
		spec.Volumes = append(spec.Volumes, volume)
		for _, step := range spec.Steps {
			mount := &engine.VolumeMount{
				Name:     id,
				Path:     v,
				ReadOnly: ro,
			}
			step.Volumes = append(step.Volumes, mount)
		}
	}

	// append volumes
	for _, v := range pipeline.Volumes {
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
		} else if v.Claim != nil {
			src.Claim = &engine.VolumeClaim{
				ID:        id,
				Name:      v.Name,
				ClaimName: v.Claim.ClaimName,
				ReadOnly:  v.Claim.ReadOnly,
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

	// apply default policy
	if m := policy.Match(match, c.Policies); m != nil {
		m.Apply(spec)
	}

	if c.PullPolicy != "" {
		pullPolicy := convertPullPolicy(c.PullPolicy)

		for _, step := range spec.Steps {
			step.Pull = pullPolicy
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
func (c *Compiler) findSecret(ctx context.Context, args runtime.CompilerArgs, name string) (s string, ok bool) {
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
