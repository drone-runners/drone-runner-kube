// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package policy

import (
	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/manifest"
)

type (
	// Policy defines pipeline defaults.
	Policy struct {
		Conditions     manifest.Conditions `yaml:"match"`
		Name           string
		Metadata       Metadata
		Resources      Resources
		NodeSelector   map[string]string `yaml:"node_selector"`
		ServiceAccount string            `yaml:"service_account"`
		Tolerations    []Toleration
	}

	// Metadata defines resource metadata.
	Metadata struct {
		Namespace   string
		Labels      map[string]string
		Annotations map[string]string
	}

	// Resources defines limits and requests for resource
	// memory and cpu.
	Resources struct {
		Request Resource
		Limit   Resource
	}

	// Resource defines resource memory and cpu.
	Resource struct {
		CPU    int64
		Memory manifest.BytesSize
	}

	// Toleration defines pod tolerations.
	Toleration struct {
		Effect            string
		Key               string
		Operator          string
		TolerationSeconds *int `yaml:"toleration_seconds"`
		Value             string
	}
)

// Apply applies the policy to the pipeline.
func (p *Policy) Apply(spec *engine.Spec) {
	// apply the default namspace
	if v := p.Metadata.Namespace; v != "" {
		spec.PodSpec.Namespace = v
		// TODO figure out how to handle temporary namespace (e.g. drone-)
	}

	// apply labels.
	// note that labels are appended as opposed to replaced
	// to ensure they do not remove Drone internal defaults.
	if v := p.Metadata.Labels; v != nil {
		p.Metadata.Labels = environ.Combine(p.Metadata.Labels, v)
	}

	// apply annotations.
	// note that annotations are appended as opposed to replaced
	// to ensure they do not remove Drone internal defaults.
	if v := p.Metadata.Annotations; v != nil {
		p.Metadata.Annotations = environ.Combine(p.Metadata.Annotations, v)
	}

	// apply resource requests
	if v := p.Resources.Request.CPU; v != 0 {
		for _, s := range spec.Steps {
			s.Resources.Requests.CPU = v
		}
	}
	if v := p.Resources.Request.Memory; v != 0 {
		for _, s := range spec.Steps {
			s.Resources.Requests.Memory = int64(v)
		}
	}

	// apply resource limits
	if v := p.Resources.Limit.CPU; v != 0 {
		for _, s := range spec.Steps {
			s.Resources.Limits.CPU = v
		}
	}
	if v := p.Resources.Limit.Memory; v != 0 {
		for _, s := range spec.Steps {
			s.Resources.Limits.Memory = int64(v)
		}
	}

	// apply the default nodeselector.
	if v := p.NodeSelector; v != nil {
		spec.PodSpec.NodeSelector = v
	}

	// apply the default service account.
	if v := p.ServiceAccount; v != "" {
		spec.PodSpec.ServiceAccountName = v
	}

	// apply (and override) the default tolerations.
	if v := p.Tolerations; len(v) != 0 {
		var dst []engine.Toleration
		for _, src := range v {
			dst = append(dst, engine.Toleration{
				Effect:            src.Effect,
				Key:               src.Key,
				Operator:          src.Operator,
				TolerationSeconds: src.TolerationSeconds,
				Value:             src.Value,
			})
		}
		spec.PodSpec.Tolerations = dst
	}
}
