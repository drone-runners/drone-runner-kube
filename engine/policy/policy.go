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
		Conditions        manifest.Conditions `yaml:"match"`
		Name              string
		Metadata          Metadata
		Resources         Resources
		MergeNodeSelector bool              `yaml:"merge_node_selector"`
		NodeSelector      map[string]string `yaml:"node_selector"`
		ServiceAccount    string            `yaml:"service_account"`
		AppendTolerations bool              `yaml:"append_tolerations"`
		Tolerations       []Toleration
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
		// the runner can be configured to create random namespaces
		// that is created before the pipeline executes, and destroyed
		// after the pipeline completes.
		if spec.PodSpec.Namespace == "drone-" {
			namespace := random()
			spec.PodSpec.Namespace = namespace
			spec.Namespace = namespace
		}
	}

	// apply labels.
	// note that labels are appended as opposed to replaced
	// to ensure they do not remove Drone internal defaults.
	if v := p.Metadata.Labels; v != nil {
		spec.PodSpec.Labels = environ.Combine(spec.PodSpec.Labels, v)
	}

	// apply annotations.
	// note that annotations are appended as opposed to replaced
	// to ensure they do not remove Drone internal defaults.
	if v := p.Metadata.Annotations; v != nil {
		spec.PodSpec.Annotations =
			environ.Combine(spec.PodSpec.Annotations, v)
	}

	// apply resource requests
	if v := p.Resources.Request.CPU; v != 0 {
		spec.Resources.Requests.CPU = v
	}
	if v := p.Resources.Request.Memory; v != 0 {
		spec.Resources.Requests.Memory = int64(v)
	}
	if v := p.Resources.Limit.CPU; v != 0 {
		spec.Resources.Limits.CPU = v
	}
	if v := p.Resources.Limit.Memory; v != 0 {
		spec.Resources.Limits.Memory = int64(v)
	}

	// apply the default nodeselector.
	switch ns := p.NodeSelector; ns != nil {
	case p.MergeNodeSelector:
		for k, v := range ns {
			if spec.PodSpec.NodeSelector[k] == "" {
				spec.PodSpec.NodeSelector[k] = v
			}
		}
	default:
		spec.PodSpec.NodeSelector = ns
	}

	// apply the default service account.
	if v := p.ServiceAccount; v != "" {
		spec.PodSpec.ServiceAccountName = v
	}

	// apply (and override) the default tolerations.
	if v := p.Tolerations; len(v) != 0 {
		var dst []engine.Toleration

		if p.AppendTolerations {
			dst = spec.PodSpec.Tolerations
		}

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
