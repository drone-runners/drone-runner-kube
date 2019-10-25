// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package resource

import (
	"testing"

	"github.com/drone/runner-go/manifest"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	got, err := manifest.ParseFile("testdata/manifest.yml")
	if err != nil {
		t.Error(err)
		return
	}

	want := []manifest.Resource{
		&manifest.Signature{
			Kind: "signature",
			Hmac: "a8842634682b78946a2",
		},
		&manifest.Secret{
			Kind: "secret",
			Type: "encrypted",
			Name: "token",
			Data: "f0e4c2f76c58916ec25",
		},
		&Pipeline{
			Kind:    "pipeline",
			Type:    "docker",
			Name:    "default",
			Version: "1",
			Workspace: Workspace{
				Path: "/drone/src",
			},
			Platform: manifest.Platform{
				OS:   "linux",
				Arch: "arm64",
			},
			Clone: manifest.Clone{
				Depth: 50,
			},
			PullSecrets: []string{"dockerconfigjson"},
			Trigger: manifest.Conditions{
				Branch: manifest.Condition{
					Include: []string{"master"},
				},
			},
			Services: []*Step{
				{
					Name:       "redis",
					Image:      "redis:latest",
					Entrypoint: []string{"/bin/redis-server"},
					Command:    []string{"--debug"},
				},
			},
			Steps: []*Step{
				{
					Name:      "build",
					Image:     "golang",
					Detach:    false,
					DependsOn: []string{"clone"},
					Commands: []string{
						"go build",
						"go test",
					},
					Environment: map[string]*manifest.Variable{
						"GOOS":   &manifest.Variable{Value: "linux"},
						"GOARCH": &manifest.Variable{Value: "arm64"},
					},
					Failure: "ignore",
					When: manifest.Conditions{
						Event: manifest.Condition{
							Include: []string{"push"},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(got.Resources, want); diff != "" {
		t.Errorf("Unexpected manifest")
		t.Log(diff)
	}
}

func TestParseErr(t *testing.T) {
	_, err := manifest.ParseFile("testdata/malformed.yml")
	if err == nil {
		t.Errorf("Expect error when malformed yaml")
	}
}

func TestParseLintErr(t *testing.T) {
	_, err := manifest.ParseFile("testdata/linterr.yml")
	if err == nil {
		t.Errorf("Expect linter returns error")
		return
	}
}

func TestParseLintNilStep(t *testing.T) {
	_, err := manifest.ParseFile("testdata/nilstep.yml")
	if err == nil {
		t.Errorf("Expect linter returns error")
		return
	}
}

func TestParseNoMatch(t *testing.T) {
	r := &manifest.RawResource{Kind: "pipeline", Type: "exec"}
	_, match, _ := parse(r)
	if match {
		t.Errorf("Expect no match")
	}
}

func TestMatch(t *testing.T) {
	r := &manifest.RawResource{
		Kind: "pipeline",
		Type: "docker",
	}
	if match(r) == false {
		t.Errorf("Expect match, got false")
	}

	r = &manifest.RawResource{
		Kind: "approval",
		Type: "docker",
	}
	if match(r) == true {
		t.Errorf("Expect kind mismatch, got true")
	}

	r = &manifest.RawResource{
		Kind: "pipeline",
		Type: "dummy",
	}
	if match(r) == true {
		t.Errorf("Expect type mismatch, got true")
	}

}

func TestLint(t *testing.T) {
	p := new(Pipeline)
	p.Steps = []*Step{{Name: "build"}, {Name: "test"}}
	if err := lint(p); err != nil {
		t.Errorf("Expect no lint error, got %s", err)
	}

	p.Steps = []*Step{{Name: "build"}, {Name: "build"}}
	if err := lint(p); err == nil {
		t.Errorf("Expect error when duplicate name")
	}

	p.Steps = []*Step{{Name: "build"}, {Name: ""}}
	if err := lint(p); err == nil {
		t.Errorf("Expect error when empty name")
	}
}
