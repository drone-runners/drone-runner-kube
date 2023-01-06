// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package policy

import (
	"reflect"
	"testing"

	"github.com/drone-runners/drone-runner-kube/engine"
)

func TestTolerations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		p    *Policy
		want []engine.Toleration
	}{
		{
			desc: "test override tolerations",
			p: &Policy{
				Tolerations: []Toleration{{Key: "drone"}},
			},
			want: []engine.Toleration{
				{Key: "drone"},
			},
		},
		{
			desc: "test append tolerations",
			p: &Policy{
				AppendTolerations: true,
				Tolerations:       []Toleration{{Key: "drone"}},
			},
			want: []engine.Toleration{
				{Key: "memory-optimized"},
				{Key: "drone"},
			},
		},
	}

	for _, test := range tests {
		spec := &engine.Spec{
			PodSpec: engine.PodSpec{
				Tolerations: []engine.Toleration{{Key: "memory-optimized"}},
			},
		}

		test.p.Apply(spec)

		if !reflect.DeepEqual(test.want, spec.PodSpec.Tolerations) {
			t.Errorf("tolerations are incorrect\ndesc: %s\nexpected: %#v\ngot: %#v", test.desc, test.want, spec.PodSpec.Tolerations)
		}
	}
}

func TestNodeSelectors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		p    *Policy
		want map[string]string
	}{
		{
			desc: "test override node_selector",
			p: &Policy{
				NodeSelector: map[string]string{"instancegroup": "drone"},
			},
			want: map[string]string{"instancegroup": "drone"},
		},
		{
			desc: "test merge node_selector",
			p: &Policy{
				MergeNodeSelector: true,
				NodeSelector:      map[string]string{"instancegroup": "drone"},
			},
			want: map[string]string{
				"instancegroup": "drone",
				"instanceclass": "memory-optimized",
			},
		},
	}

	for _, test := range tests {
		spec := &engine.Spec{
			PodSpec: engine.PodSpec{
				NodeSelector: map[string]string{"instanceclass": "memory-optimized"},
			},
		}

		test.p.Apply(spec)

		if !reflect.DeepEqual(test.want, spec.PodSpec.NodeSelector) {
			t.Errorf("node_selector is incorrect\ndesc: %s\nexpected: %#v\ngot: %#v", test.desc, test.want, spec.PodSpec.NodeSelector)
		}
	}
}
