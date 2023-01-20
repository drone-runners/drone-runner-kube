// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

// +build !windows

package compiler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var nocontext = context.Background()

// dummy function that returns a non-random string for testing.
// it is used in place of the random function.
func notRandom() string {
	return "random"
}

// This test verifies the pipeline dependency graph. When no
// dependency graph is defined, a default dependency graph is
// automatically defined to run steps serially.
func TestCompile_Serial(t *testing.T) {
	testCompile(t, "testdata/serial.yml", "testdata/serial.json")
}

// This test verifies the pipeline services.
func TestCompile_Services(t *testing.T) {
	testCompile(t, "testdata/service.yml", "testdata/service.json")
}

// This test verifies a detached service.
func TestCompile_DetachedService(t *testing.T) {
	testCompile(t, "testdata/detached_service.yml", "testdata/detached_service.json")
}

// This test verifies the pipeline dependency graph. It also
// verifies that pipeline steps with no dependencies depend on
// the initial clone step.
func TestCompile_Graph(t *testing.T) {
	testCompile(t, "testdata/graph.yml", "testdata/graph.json")
}

// This test verifies no clone step exists in the pipeline if
// cloning is disabled.
func TestCompile_CloneDisabled_Serial(t *testing.T) {
	testCompile(t, "testdata/noclone_serial.yml", "testdata/noclone_serial.json")
}

// This test verifies no clone step exists in the pipeline if
// cloning is disabled. It also verifies no pipeline steps
// depend on a clone step.
func TestCompile_CloneDisabled_Graph(t *testing.T) {
	testCompile(t, "testdata/noclone_graph.yml", "testdata/noclone_graph.json")
}

// This test verifies that steps are disabled if conditions
// defined in the when block are not satisfied.
func TestCompile_Match(t *testing.T) {
	ir := testCompile(t, "testdata/match.yml", "testdata/match.json")
	if ir.Steps[0].RunPolicy != runtime.RunOnSuccess {
		t.Errorf("Expect run on success")
	}
	if ir.Steps[1].RunPolicy != runtime.RunNever {
		t.Errorf("Expect run never")
	}
}

// This test verifies that steps configured to run on both
// success or failure are configured to always run.
func TestCompile_RunAlways(t *testing.T) {
	ir := testCompile(t, "testdata/run_always.yml", "testdata/run_always.json")
	if ir.Steps[0].RunPolicy != runtime.RunAlways {
		t.Errorf("Expect run always")
	}
}

// This test verifies that steps configured to run on failure
// are configured to run on failure.
func TestCompile_RunFailure(t *testing.T) {
	ir := testCompile(t, "testdata/run_failure.yml", "testdata/run_failure.json")
	if ir.Steps[0].RunPolicy != runtime.RunOnFailure {
		t.Errorf("Expect run on failure")
	}
}

func TestPlatformNodeSelector(t *testing.T) {
	testCompile(t, "testdata/arm.yml", "testdata/arm.json")
}

// This test verifies that secrets defined in the yaml are
// requested and stored in the intermediate representation
// at compile time.
func TestCompile_Secrets(t *testing.T) {
	manifest, _ := manifest.ParseFile("testdata/secret.yml")

	compiler := &Compiler{
		Environ:  provider.Static(nil),
		Registry: registry.Static(nil),
		Secret: secret.StaticVars(map[string]string{
			"token":       "3DA541559918A808C2402BBA5012F6C60B27661C",
			"password":    "password",
			"my_username": "octocat",
		}),
	}
	args := runtime.CompilerArgs{
		Repo:     &drone.Repo{},
		Build:    &drone.Build{},
		Stage:    &drone.Stage{},
		System:   &drone.System{},
		Netrc:    &drone.Netrc{},
		Manifest: manifest,
		Pipeline: manifest.Resources[0].(*resource.Pipeline),
		Secret:   secret.Static(nil),
	}

	ir := compiler.Compile(nocontext, args).(*engine.Spec)

	// verify the reference to the secrets are stored in the
	// pipeline step.
	{
		got := ir.Steps[0].Secrets
		// sort to avoid changed order
		sort.Slice(got, func(i, j int) bool {
			return got[i].Name < got[j].Name
		})

		want := []*engine.SecretVar{
			{Name: "my_password", Env: "PASSWORD"},
			{Name: "my_username", Env: "USERNAME"},
		}
		if diff := cmp.Diff(got, want); len(diff) != 0 {
			t.Error(diff)
		}
	}

	// verify the secret are stored in the pipeline
	// specification with the expected value.
	{
		got := ir.Secrets
		want := map[string]*engine.Secret{
			"my_password": {
				Name: "my_password",
				Data: "", // secret not found, data empty
				Mask: false,
			},
			"my_username": {
				Name: "my_username",
				Data: "octocat", // secret found
				Mask: true,
			},
		}
		if diff := cmp.Diff(got, want); len(diff) != 0 {
			t.Error(diff)
		}
	}
}

// helper function parses and compiles the source file and then
// compares to a golden json file.
func testCompile(t *testing.T, source, golden string) *engine.Spec {
	// replace the default random function with one that
	// is deterministic, for testing purposes.
	random = notRandom

	// restore the default random function and the previously
	// specified temporary directory
	defer func() {
		random = uniuri.New
	}()

	manifest, err := manifest.ParseFile(source)
	if err != nil {
		t.Error(err)
		return nil
	}

	compiler := &Compiler{
		Environ:  provider.Static(nil),
		Registry: registry.Static(nil),
		Secret: secret.StaticVars(map[string]string{
			"token":       "3DA541559918A808C2402BBA5012F6C60B27661C",
			"password":    "password",
			"my_username": "octocat",
		}),
	}
	args := runtime.CompilerArgs{
		Repo:     &drone.Repo{},
		Build:    &drone.Build{Target: "master"},
		Stage:    &drone.Stage{},
		System:   &drone.System{},
		Netrc:    &drone.Netrc{Machine: "github.com", Login: "octocat", Password: "correct-horse-battery-staple"},
		Manifest: manifest,
		Pipeline: manifest.Resources[0].(*resource.Pipeline),
		Secret:   secret.Static(nil),
	}

	got := compiler.Compile(nocontext, args)

	raw, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Error(err)
	}

	want := new(engine.Spec)
	err = json.Unmarshal(raw, want)
	if err != nil {
		t.Error(err)
	}

	opts := cmp.Options{
		cmpopts.IgnoreUnexported(engine.Spec{}),
		cmpopts.IgnoreFields(engine.Spec{}, "Secrets"),
		cmpopts.IgnoreFields(engine.Step{}, "Envs", "Secrets", "SpecSecrets"),
		cmpopts.IgnoreFields(engine.PodSpec{}, "Annotations", "Labels"),
	}
	if diff := cmp.Diff(got, want, opts...); len(diff) != 0 {
		t.Error(diff)
	}

	return got.(*engine.Spec)
}

func TestCompile_Netrc(t *testing.T) {
	testCompileNetrc(t, "testdata/netrc.yml", "testdata/netrc.json", false)
}

func TestCompile_NetrcOnly(t *testing.T) {
	testCompileNetrc(t, "testdata/netrc.yml", "testdata/netrc_only_clone.json", true)
}

// This test verifies that netrc values are properly defined as kubernetes secrets.
func testCompileNetrc(t *testing.T, source, golden string, netRcCloneOnly bool) {
	// replace the default random function with one that
	// is deterministic, for testing purposes.
	random = notRandom

	// restore the default random function and the previously
	// specified temporary directory
	defer func() {
		random = uniuri.New
	}()

	manifest, err := manifest.ParseFile(source)
	if err != nil {
		t.Error(err)
		return
	}

	compiler := &Compiler{
		Environ:  provider.Static(nil),
		Registry: registry.Static(nil),
		Secret: secret.StaticVars(map[string]string{
			"token":       "3DA541559918A808C2402BBA5012F6C60B27661C",
			"password":    "password",
			"my_username": "octocat",
		}),
		NetrcCloneOnly: netRcCloneOnly,
	}
	args := runtime.CompilerArgs{
		Repo:     &drone.Repo{},
		Build:    &drone.Build{Target: "master"},
		Stage:    &drone.Stage{},
		System:   &drone.System{},
		Netrc:    &drone.Netrc{Machine: "github.com", Login: "octocat", Password: "correct-horse-battery-staple"},
		Manifest: manifest,
		Pipeline: manifest.Resources[0].(*resource.Pipeline),
		Secret:   secret.Static(nil),
	}

	got := compiler.Compile(nocontext, args)

	raw, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Error(err)
	}

	want := new(engine.Spec)
	err = json.Unmarshal(raw, want)
	if err != nil {
		t.Error(err)
	}

	opts := cmp.Options{
		cmpopts.IgnoreUnexported(engine.Spec{}),
		cmpopts.IgnoreFields(engine.Step{}, "Envs"),
		cmpopts.IgnoreFields(engine.PodSpec{}, "Annotations", "Labels"),
	}
	if diff := cmp.Diff(got, want, opts...); len(diff) != 0 {
		t.Error(diff)
	}
}

func dump(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func TestDivideIntEqually(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		count    int
		unit     int64
		expected []int64
	}{
		{
			name:     "one part, no splitting",
			amount:   10,
			count:    1,
			unit:     1,
			expected: []int64{10},
		},
		{
			name:     "two halves",
			amount:   10,
			count:    2,
			unit:     1,
			expected: []int64{5, 5},
		},
		{
			name:     "one part, convert to unit multiple",
			amount:   10,
			count:    1,
			unit:     3,
			expected: []int64{12}, // must be multiple of unit
		},
		{
			name:     "split 1 to two parts",
			amount:   1,
			count:    2,
			unit:     1,
			expected: []int64{1, 0},
		},
		{
			name:     "split 12/5, unit 1",
			amount:   12,
			count:    5,
			unit:     1,
			expected: []int64{3, 3, 2, 2, 2},
		},
		{
			name:     "split 12/5, unit 2",
			amount:   12,
			count:    5,
			unit:     2,
			expected: []int64{4, 2, 2, 2, 2},
		},
		{
			name:     "split 12/5, unit 3",
			amount:   12,
			count:    5,
			unit:     3,
			expected: []int64{3, 3, 3, 3, 0},
		},
	}

	for _, test := range tests {
		result := divideIntEqually(test.amount, test.count, test.unit)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("test %q failed, amount=%d, count=%d, unit=%d, expected=%v, got=%v",
				test.name, test.amount, test.count, test.unit, test.expected, result)
		}
	}
}
