// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package policy

import (
	"reflect"
	"testing"

	"github.com/drone-runners/drone-runner-kube/engine"
)

func TestPodSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description string
		policy      *Policy
		spec        engine.PodSpec
		want        engine.PodSpec
	}{
		{
			description: "test override tolerations",
			policy: &Policy{
				Tolerations: []Toleration{{Key: "drone"}},
			},
			spec: engine.PodSpec{
				Tolerations: []engine.Toleration{{Key: "memory-optimized"}},
			},
			want: engine.PodSpec{
				Tolerations: []engine.Toleration{{Key: "drone"}},
			},
		},
		{
			description: "test append tolerations",
			policy: &Policy{
				AppendTolerations: true,
				Tolerations:       []Toleration{{Key: "drone"}},
			},
			spec: engine.PodSpec{
				Tolerations: []engine.Toleration{{Key: "memory-optimized"}},
			},
			want: engine.PodSpec{
				Tolerations: []engine.Toleration{
					{Key: "memory-optimized"},
					{Key: "drone"},
				},
			},
		},
		{
			description: "test override node_selector",
			policy: &Policy{
				NodeSelector: map[string]string{"instancegroup": "drone"},
			},
			spec: engine.PodSpec{
				NodeSelector: map[string]string{"instanceclass": "memory-optimized"},
			},
			want: engine.PodSpec{
				NodeSelector: map[string]string{"instancegroup": "drone"},
			},
		},
		{
			description: "test merge node_selector",
			policy: &Policy{
				MergeNodeSelector: true,
				NodeSelector:      map[string]string{"instancegroup": "drone"},
			},
			spec: engine.PodSpec{
				NodeSelector: map[string]string{"instanceclass": "memory-optimized"},
			},
			want: engine.PodSpec{
				NodeSelector: map[string]string{
					"instancegroup": "drone",
					"instanceclass": "memory-optimized",
				},
			},
		},
	}

	for _, test := range tests {
		got := &engine.Spec{PodSpec: test.spec}
		test.policy.Apply(got)

		if !reflect.DeepEqual(test.want, got.PodSpec) {
			t.Errorf("description: %s\nexpected: %#v\ngot: %#v", test.description, test.want, got.PodSpec)
		}
	}
}
