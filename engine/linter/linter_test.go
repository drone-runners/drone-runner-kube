// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package linter

import (
	"path"
	"testing"

	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/manifest"
)

func TestLint(t *testing.T) {
	tests := []struct {
		path     string
		trusted  bool
		invalid  bool
		message  string
		repo     string
		patterns map[string][]string
	}{
		{
			path:    "testdata/simple.yml",
			trusted: false,
			invalid: false,
		},
		{
			path:    "testdata/missing_image.yml",
			invalid: true,
			message: "linter: invalid or missing image",
		},
		// user should not use reserved volume names.
		{
			path:    "testdata/volume_missing_name.yml",
			trusted: false,
			invalid: true,
			message: "linter: missing volume name",
		},
		{
			path:    "testdata/volume_invalid_name.yml",
			trusted: false,
			invalid: true,
			message: "linter: invalid volume name: _workspace",
		},
		// user should not be able to mount host path
		// volumes unless the repository is trusted.
		{
			path:    "testdata/volume_host_path.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot mount host volumes",
		},
		{
			path:    "testdata/volume_host_path.yml",
			trusted: true,
			invalid: false,
		},
		// user should be able to mount emptyDir volumes
		// where no medium is specified.
		{
			path:    "testdata/volume_empty_dir.yml",
			trusted: false,
			invalid: false,
		},
		// user should not be able to mount in-memory
		// emptyDir volumes unless the repository is
		// trusted.
		{
			path:    "testdata/volume_empty_dir_memory.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot mount in-memory volumes",
		},
		{
			path:    "testdata/volume_empty_dir_memory.yml",
			trusted: true,
			invalid: false,
		},
		// user should not be trying to mount internal or restricted
		// volume paths.
		{
			path:    "testdata/volume_restricted.yml",
			trusted: false,
			invalid: true,
			message: "linter: cannot mount volume at /run/drone",
		},
		// user should not be able to set the securityContext
		// unless the repository is trusted.
		{
			path:    "testdata/pipeline_privileged.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot enable privileged mode",
		},
		{
			path:    "testdata/pipeline_privileged.yml",
			trusted: true,
			invalid: false,
		},
		// linter should verify whether or not a repository can
		// use a target namespace
		{
			path:     "testdata/simple.yml",
			trusted:  false,
			invalid:  false,
			patterns: map[string][]string{"default": []string{"octocat/*"}},
			repo:     "octocat/hello-world",
		},
		{
			path:     "testdata/simple.yml",
			trusted:  false,
			invalid:  false,
			patterns: map[string][]string{"default": []string{"*/*"}},
			repo:     "octocat/hello-world",
		},
		{
			path:     "testdata/simple.yml",
			trusted:  false,
			invalid:  false,
			patterns: map[string][]string{"default": []string{"**"}},
			repo:     "octocat/hello-world",
		},
		{
			path:     "testdata/simple_ns.yml",
			trusted:  false,
			invalid:  false,
			patterns: map[string][]string{"default": []string{"octocat/*"}},
			repo:     "octocat/hello-world",
		},
		// no matching pattern, ok
		{
			path:     "testdata/simple_ns.yml",
			trusted:  false,
			invalid:  false,
			patterns: map[string][]string{"unknown": []string{"octocat/*"}},
			repo:     "octocat/hello-world",
		},
		{
			path:     "testdata/simple_ns.yml",
			trusted:  false,
			invalid:  true,
			patterns: map[string][]string{"default": []string{"octocat/*"}},
			repo:     "spaceghost/hello-world",
			message:  "linter: pipeline restricted from using configured namespace",
		},

		//
		// The below checks were moved to the parser, however, we
		// should decide where we want this logic to live.
		//

		// // user should not be able to use duplicate names
		// // for steps or services.
		// {
		// 	path:    "testdata/duplicate_step.yml",
		// 	invalid: true,
		// 	message: "linter: duplicate step names",
		// },
		// {
		// 	path:    "testdata/duplicate_step_service.yml",
		// 	invalid: true,
		// 	message: "linter: duplicate step names",
		// },
		// {
		// 	path:    "testdata/missing_name.yml",
		// 	invalid: true,
		// 	message: "linter: invalid or missing name",
		// },
	}
	for _, test := range tests {
		name := path.Base(test.path)
		if test.trusted {
			name = name + "/trusted"
		}
		t.Run(name, func(t *testing.T) {
			resources, err := manifest.ParseFile(test.path)
			if err != nil {
				t.Logf("yaml: %s", test.path)
				t.Logf("trusted: %v", test.trusted)
				t.Error(err)
				return
			}

			lint := New(test.patterns)
			repo := &drone.Repo{Trusted: test.trusted, Slug: test.repo}
			err = lint.Lint(resources.Resources[0].(*resource.Pipeline), repo)
			if err == nil && test.invalid == true {
				t.Logf("yaml: %s", test.path)
				t.Logf("trusted: %v", test.trusted)
				t.Errorf("Expect lint error")
				return
			}

			if err != nil && test.invalid == false {
				t.Logf("yaml: %s", test.path)
				t.Logf("trusted: %v", test.trusted)
				t.Errorf("Expect lint error is nil, got %s", err)
				return
			}

			if err == nil {
				return
			}

			if got, want := err.Error(), test.message; got != want {
				t.Logf("yaml: %s", test.path)
				t.Logf("trusted: %v", test.trusted)
				t.Errorf("Want message %q, got %q", want, got)
				return
			}
		})
	}
}
