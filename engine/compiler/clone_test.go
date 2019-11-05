// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"testing"

	"github.com/dchest/uniuri"
	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestClone(t *testing.T) {
	random = notRandom
	defer func() {
		random = uniuri.New
	}()

	c := &Compiler{
		Registry: registry.Static(nil),
		Secret:   secret.Static(nil),
	}
	args := Args{
		Repo:     &drone.Repo{},
		Build:    &drone.Build{},
		Stage:    &drone.Stage{},
		System:   &drone.System{},
		Netrc:    &drone.Netrc{},
		Manifest: &manifest.Manifest{},
		Pipeline: &resource.Pipeline{},
	}
	want := []*engine.Step{
		{
			ID:          "random",
			Image:       "drone/git:latest",
			Placeholder: "drone/placeholder:1",
			Name:        "clone",
			Pull:        engine.PullDefault,
			RunPolicy:   engine.RunAlways,
			WorkingDir:  "/drone/src",
			Volumes: []*engine.VolumeMount{
				&engine.VolumeMount{
					Name: "_workspace",
					Path: "/drone/src",
				},
				&engine.VolumeMount{
					Name: "_status",
					Path: "/run/drone",
				},
			},
		},
	}
	got := c.Compile(nocontext, args)
	ignore := cmpopts.IgnoreFields(engine.Step{}, "Envs", "Labels")
	if diff := cmp.Diff(got.Steps, want, ignore); len(diff) != 0 {
		t.Errorf(diff)
	}
}

func TestCloneDisable(t *testing.T) {
	c := &Compiler{
		Registry: registry.Static(nil),
		Secret:   secret.Static(nil),
	}
	args := Args{
		Repo:     &drone.Repo{},
		Build:    &drone.Build{},
		Stage:    &drone.Stage{},
		System:   &drone.System{},
		Netrc:    &drone.Netrc{},
		Manifest: &manifest.Manifest{},
		Pipeline: &resource.Pipeline{Clone: manifest.Clone{Disable: true}},
	}
	got := c.Compile(nocontext, args)
	if len(got.Steps) != 0 {
		t.Errorf("Expect no clone step added when disabled")
	}
}

func TestCloneCreate(t *testing.T) {
	want := &engine.Step{
		Name:        "clone",
		Image:       "drone/git:latest",
		Placeholder: "drone/placeholder:1",
		RunPolicy:   engine.RunAlways,
		Envs:        map[string]string{"PLUGIN_DEPTH": "50"},
	}
	src := &resource.Pipeline{Clone: manifest.Clone{Depth: 50}}
	got := createClone(src)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

func TestCloneImage(t *testing.T) {
	tests := []struct {
		in  manifest.Platform
		out string
	}{
		{
			in:  manifest.Platform{},
			out: "drone/git:latest",
		},
		{
			in:  manifest.Platform{OS: "linux"},
			out: "drone/git:latest",
		},
		{
			in:  manifest.Platform{OS: "windows"},
			out: "drone/git:latest",
		},
	}
	for _, test := range tests {
		got, want := cloneImage(test.in), test.out
		if got != want {
			t.Errorf("Want clone image %q, got %q", want, got)
		}
	}
}

func TestCloneParams(t *testing.T) {
	params := cloneParams(manifest.Clone{})
	if len(params) != 0 {
		t.Errorf("Expect empty clone parameters")
	}
	params = cloneParams(manifest.Clone{Depth: 0})
	if len(params) != 0 {
		t.Errorf("Expect zero depth ignored")
	}
	params = cloneParams(manifest.Clone{Depth: 50, SkipVerify: true})
	if params["PLUGIN_DEPTH"] != "50" {
		t.Errorf("Expect clone depth 50")
	}
	if params["GIT_SSL_NO_VERIFY"] != "true" {
		t.Errorf("Expect GIT_SSL_NO_VERIFY is true")
	}
	if params["PLUGIN_SKIP_VERIFY"] != "true" {
		t.Errorf("Expect PLUGIN_SKIP_VERIFY is true")
	}
}
