// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"sync"

	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/pipeline/runtime"
)

type (
	// Spec provides the pipeline spec. This provides the
	// required instructions for reproducible pipeline
	// execution.
	Spec struct {
		PodSpec    PodSpec            `json:"pod_spec,omitempty"`
		Platform   Platform           `json:"platform,omitempty"`
		Steps      []*Step            `json:"steps,omitempty"`
		Internal   []*Step            `json:"internal,omitempty"`
		Volumes    []*Volume          `json:"volumes,omitempty"`
		Secrets    map[string]*Secret `json:"secrets,omitempty"`
		PullSecret *Secret            `json:"pull_secrets,omitempty"`

		// Resources hold resource limit for each container and
		// resource request amount for the whole pod.
		// This must be present here so that a policy can override the values.
		Resources Resources

		// Runtime field to gate updating of the pod that this pipeline
		// is running on. Helps to avoid self-inflicted 409 Conflict
		// responses from the kubernetes api server.
		podUpdateMutex sync.Mutex

		// Namespace is an optional namespace that should be
		// created before the pipeline starts and executed after
		// the pipeline completes. WARNING this field should only
		// be set if you want custom per-pipeline namespaces.
		Namespace string `json:"namespace,omitempty"`

		// stop channel is created by the engine's Setup method, and closed by the Destroy method.
		// It's used to quickly bail out from the Run method if the pipeline is terminated or canceled.
		stop chan struct{}
	}

	// Step defines a pipeline step.
	Step struct {
		ID           string            `json:"id,omitempty"`
		Command      []string          `json:"args,omitempty"`
		Detach       bool              `json:"detach,omitempty"`
		DependsOn    []string          `json:"depends_on,omitempty"`
		Entrypoint   []string          `json:"entrypoint,omitempty"`
		Envs         map[string]string `json:"environment,omitempty"`
		ErrPolicy    runtime.ErrPolicy `json:"err_policy,omitempty"`
		IgnoreStdout bool              `json:"ignore_stderr,omitempty"`
		IgnoreStderr bool              `json:"ignore_stdout,omitempty"`
		Image        string            `json:"image,omitempty"`
		Name         string            `json:"name,omitempty"`
		Placeholder  string            `json:"placeholder,omitempty"`
		Privileged   bool              `json:"privileged,omitempty"`
		Resources    Resources         `json:"resources,omitempty"`
		Pull         PullPolicy        `json:"pull,omitempty"`
		RunPolicy    runtime.RunPolicy `json:"run_policy,omitempty"`
		Secrets      []*SecretVar      `json:"secrets,omitempty"`
		SpecSecrets  []*Secret         `json:"spec_secrets,omitempty"`
		User         *int64            `json:"user,omitempty"`
		Group        *int64            `json:"group,omitempty"`
		Volumes      []*VolumeMount    `json:"volumes,omitempty"`
		WorkingDir   string            `json:"working_dir,omitempty"`
	}

	// Platform defines the target platform.
	Platform struct {
		OS      string `json:"os,omitempty"`
		Arch    string `json:"arch,omitempty"`
		Variant string `json:"variant,omitempty"`
		Version string `json:"version,omitempty"`
	}

	// Secret represents a secret variable.
	Secret struct {
		Name string `json:"name,omitempty"`
		Data string `json:"data,omitempty"`
		Mask bool   `json:"mask,omitempty"`
	}

	// SecretVar represents an environment variable
	// sources from a secret.
	SecretVar struct {
		Name string `json:"name,omitempty"`
		Env  string `json:"env,omitempty"`
	}

	// State represents the process state.
	State struct {
		ExitCode  int  // Container exit code
		Exited    bool // Container exited
		OOMKilled bool // Container is oom killed
	}

	// Volume that can be mounted by containers.
	Volume struct {
		EmptyDir    *VolumeEmptyDir    `json:"temp,omitempty"`
		HostPath    *VolumeHostPath    `json:"host,omitempty"`
		DownwardAPI *VolumeDownwardAPI `json:"downward_api,omitempty"`
		Claim       *VolumeClaim       `json:"claim,omitempty"`
		ConfigMap   *VolumeConfigMap   `json:"config_map,omitempty"`
		Secret      *VolumeSecret      `json:"secret,omitempty"`
	}

	// VolumeMount describes a mounting of a Volume
	// within a container.
	VolumeMount struct {
		Name     string `json:"name,omitempty"`
		Path     string `json:"path,omitempty"`
		SubPath  string `json:"sub_path,omitempty"`
		ReadOnly bool   `json:"read_only,omitempty"`
	}

	// VolumeEmptyDir mounts a temporary directory from the
	// host node's filesystem into the container. This can
	// be used as a shared scratch space.
	VolumeEmptyDir struct {
		ID        string `json:"id,omitempty"`
		Name      string `json:"name,omitempty"`
		Medium    string `json:"medium,omitempty"`
		SizeLimit int64  `json:"size_limit,omitempty"`
	}

	// VolumeHostPath mounts a file or directory from the
	// host node's filesystem into your container.
	VolumeHostPath struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Path string `json:"path,omitempty"`
	}
	// VolumeDownwardAPI ...
	VolumeDownwardAPI struct {
		ID    string                  `json:"id,omitempty"`
		Name  string                  `json:"name,omitempty"`
		Items []VolumeDownwardAPIItem `json:"items,omitempty"`
	}
	// VolumeDownwardAPIItem ...
	VolumeDownwardAPIItem struct {
		Path      string `json:"path,omitempty"`
		FieldPath string `json:"field_path,omitempty"`
	}

	// VolumeClaim ...
	VolumeClaim struct {
		ID        string `json:"id,omitempty"`
		Name      string `json:"name,omitempty"`
		ClaimName string `json:"claim_name,omitempty"`
		ReadOnly  bool   `json:"read_only,omitempty"`
	}

	// VolumeConfigMap ...
	VolumeConfigMap struct {
		ID            string `json:"id,omitempty"`
		Name          string `json:"name,omitempty"`
		ConfigMapName string `json:"config_map_name,omitempty"`
		DefaultMode   int32  `json:"default_mode,omitempty"`
		Optional      bool   `json:"optional,omitempty"`
	}

	// VolumeSecret ...
	VolumeSecret struct {
		ID          string `json:"id,omitempty"`
		Name        string `json:"name,omitempty"`
		SecretName  string `json:"secret_name,omitempty"`
		DefaultMode int32  `json:"default_mode,omitempty"`
		Optional    bool   `json:"optional,omitempty"`
	}

	// Resources describes the compute resource requirements.
	Resources struct {
		Limits   ResourceObject `json:"limits,omitempty"`
		Requests ResourceObject `json:"requests,omitempty"`
	}

	// ResourceObject describes compute resource requirements.
	ResourceObject struct {
		CPU    int64 `json:"cpu"`
		Memory int64 `json:"memory"`
	}

	// PodSpec ...
	PodSpec struct {
		Name               string            `json:"name,omitempty"`
		Namespace          string            `json:"namespace,omitempty"`
		Annotations        map[string]string `json:"annotations,omitempty"`
		Labels             map[string]string `json:"labels,omitempty"`
		NodeName           string            `json:"node_name,omitempty"`
		NodeSelector       map[string]string `json:"node_selector,omitempty"`
		Tolerations        []Toleration      `json:"tolerations,omitempty"`
		ServiceAccountName string            `json:"service_account_name,omitempty"`
		HostAliases        []HostAlias       `json:"host_aliases,omitempty"`
		DnsConfig          DnsConfig         `json:"dns_config,omitempty"`
	}

	// HostAlias ...
	HostAlias struct {
		IP        string   `json:"ip,omitempty"`
		Hostnames []string `json:"hostnames,omitempty"`
	}

	// Toleration ...
	Toleration struct {
		Effect            string `json:"effect,omitempty"`
		Key               string `json:"key,omitempty"`
		Operator          string `json:"operator,omitempty"`
		TolerationSeconds *int   `json:"toleration_seconds,omitempty"`
		Value             string `json:"value,omitempty"`
	}
	// DnsConfig
	DnsConfig struct {
		Nameservers []string           `json:"nameservers,omitempty"`
		Searches    []string           `json:"searches,omitempty"`
		Options     []DNSConfigOptions `json:"options,omitempty"`
	}

	DNSConfigOptions struct {
		Name  string  `json:"name,omitempty"`
		Value *string `json:"value,omitempty"`
	}
)

//
// implements the Spec interface
//

func (s *Spec) StepLen() int              { return len(s.Steps) }
func (s *Spec) StepAt(i int) runtime.Step { return s.Steps[i] }

//
// implements the Secret interface
//

func (s *Secret) GetName() string  { return s.Name }
func (s *Secret) GetValue() string { return string(s.Data) }
func (s *Secret) IsMasked() bool   { return s.Mask }

//
// implements the Step interface
//

func (s *Step) GetName() string                  { return s.Name }
func (s *Step) GetDependencies() []string        { return s.DependsOn }
func (s *Step) GetEnviron() map[string]string    { return s.Envs }
func (s *Step) SetEnviron(env map[string]string) { s.Envs = env }
func (s *Step) GetErrPolicy() runtime.ErrPolicy  { return s.ErrPolicy }
func (s *Step) GetRunPolicy() runtime.RunPolicy  { return s.RunPolicy }
func (s *Step) GetSecretAt(i int) runtime.Secret { return s.SpecSecrets[i] }
func (s *Step) GetSecretLen() int                { return len(s.SpecSecrets) }
func (s *Step) IsDetached() bool                 { return s.Detach }
func (s *Step) Clone() runtime.Step {
	dst := new(Step)
	*dst = *s
	dst.Envs = environ.Combine(s.Envs)
	return dst
}

func (s *Step) GetImage() string {
	return s.Image
}
