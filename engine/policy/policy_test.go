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
		desc string
		p    *Policy
		spec engine.PodSpec
		want engine.PodSpec
	}{
		{
			desc: "test override tolerations",
			p: &Policy{
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
			desc: "test append tolerations",
			p: &Policy{
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
			desc: "test override node_selector",
			p: &Policy{
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
			desc: "test merge node_selector",
			p: &Policy{
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
		test.p.Apply(got)

		if !reflect.DeepEqual(test.want, got.PodSpec) {
			t.Errorf("desc: %s\nexpected: %#v\ngot: %#v", test.desc, test.want, got.PodSpec)
		}
	}
}
