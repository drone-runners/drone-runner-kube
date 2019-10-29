// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// helper function converts the engine pull policy
// to the kubernetes pull policy constant.
func toPullPolicy(from PullPolicy) v1.PullPolicy {
	switch from {
	case PullAlways:
		return v1.PullAlways
	case PullNever:
		return v1.PullNever
	case PullIfNotExists:
		return v1.PullIfNotPresent
	default:
		return v1.PullIfNotPresent
	}
}

// helper function returns a kubernetes namespace
// for the given specification.
func toNamespace(spec *Spec) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   spec.Metadata.Namespace,
			Labels: spec.Metadata.Labels,
		},
	}
}

// helper function converts environment variable
// string data to kubernetes variables.
func toEnvVars(spec *Spec, step *Step) []v1.EnvVar {
	var to []v1.EnvVar
	for k, v := range step.Envs {
		to = append(to, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	to = append(to, v1.EnvVar{
		Name: "KUBERNETES_NODE",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "spec.nodeName",
			},
		},
	})
	// secrets are set using environment variables. We may want
	// to consider passing secrets as kubernetes secrets.
	for _, secret := range step.Secrets {
		to = append(to, v1.EnvVar{
			Name:  secret.Env,
			Value: string(secret.Data),
		})
	}
	return to
}

func boolptr(v bool) *bool {
	return &v
}

func stringptr(v string) *string {
	return &v
}
