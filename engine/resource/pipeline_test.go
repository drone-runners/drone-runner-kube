// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package resource

import (
	"testing"

	"github.com/drone/runner-go/manifest"

	"github.com/google/go-cmp/cmp"
)

func TestGetStep(t *testing.T) {
	step1 := &Step{Name: "build"}
	step2 := &Step{Name: "test"}
	pipeline := &Pipeline{
		Steps: []*Step{step1, step2},
	}
	if pipeline.GetStep("build") != step1 {
		t.Errorf("Expected named step")
	}
	if pipeline.GetStep("deploy") != nil {
		t.Errorf("Expected nil step")
	}
}

func TestGetters(t *testing.T) {
	platform := manifest.Platform{
		OS:   "linux",
		Arch: "amd64",
	}
	trigger := manifest.Conditions{
		Branch: manifest.Condition{
			Include: []string{"master"},
		},
	}
	pipeline := &Pipeline{
		Version:  "1.0.0",
		Kind:     "pipeline",
		Type:     "docker",
		Name:     "default",
		Deps:     []string{"before"},
		Platform: platform,
		Trigger:  trigger,
	}
	if got, want := pipeline.GetVersion(), pipeline.Version; got != want {
		t.Errorf("Want Version %s, got %s", want, got)
	}
	if got, want := pipeline.GetKind(), pipeline.Kind; got != want {
		t.Errorf("Want Kind %s, got %s", want, got)
	}
	if got, want := pipeline.GetType(), pipeline.Type; got != want {
		t.Errorf("Want Type %s, got %s", want, got)
	}
	if got, want := pipeline.GetName(), pipeline.Name; got != want {
		t.Errorf("Want Name %s, got %s", want, got)
	}
	if diff := cmp.Diff(pipeline.GetDependsOn(), pipeline.Deps); diff != "" {
		t.Errorf("Unexpected Deps")
		t.Log(diff)
	}
	if diff := cmp.Diff(pipeline.GetTrigger(), pipeline.Trigger); diff != "" {
		t.Errorf("Unexpected Trigger")
		t.Log(diff)
	}
	if got, want := pipeline.GetPlatform(), pipeline.Platform; got != want {
		t.Errorf("Want Platform %s, got %s", want, got)
	}
}
