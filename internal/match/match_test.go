// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package match

import (
	"testing"

	"github.com/drone/drone-go/drone"
)

func TestFunc(t *testing.T) {
	tests := []struct {
		repo    string
		event   string
		trusted bool
		match   bool
		matcher func(*drone.Repo, *drone.Build) bool
	}{
		//
		// Expect match true
		//

		// repository, event and trusted flag matching
		{
			repo:    "octocat/hello-world",
			event:   "push",
			trusted: true,
			match:   true,
			matcher: Func([]string{"spaceghost/*", "octocat/*"}, []string{"push"}, true),
		},
		// repoisitory matching
		{
			repo:    "octocat/hello-world",
			event:   "pull_request",
			trusted: false,
			match:   true,
			matcher: Func([]string{"spaceghost/*", "octocat/*"}, []string{}, false),
		},
		// event matching
		{
			repo:    "octocat/hello-world",
			event:   "pull_request",
			trusted: false,
			match:   true,
			matcher: Func([]string{}, []string{"pull_request"}, false),
		},
		// trusted flag matching
		{
			repo:    "octocat/hello-world",
			event:   "pull_request",
			trusted: true,
			match:   true,
			matcher: Func([]string{}, []string{}, true),
		},

		//
		// Expect match false
		//

		// repository matching
		{
			repo:    "spaceghost/hello-world",
			event:   "pull_request",
			trusted: false,
			match:   false,
			matcher: Func([]string{"octocat/*"}, []string{}, false),
		},
		// event matching
		{
			repo:    "octocat/hello-world",
			event:   "pull_request",
			trusted: false,
			match:   false,
			matcher: Func([]string{}, []string{"push"}, false),
		},
		// trusted flag matching
		{
			repo:    "octocat/hello-world",
			event:   "pull_request",
			trusted: false,
			match:   false,
			matcher: Func([]string{}, []string{}, true),
		},
		// does not match repository
		{
			repo:    "foo/hello-world",
			event:   "push",
			trusted: true,
			match:   false,
			matcher: Func([]string{"spaceghost/*", "octocat/*"}, []string{"push"}, true),
		},
		// does not match event
		{
			repo:    "octocat/hello-world",
			event:   "pull_request",
			trusted: true,
			match:   false,
			matcher: Func([]string{"spaceghost/*", "octocat/*"}, []string{"push"}, true),
		},
		// does not match trusted flag
		{
			repo:    "octocat/hello-world",
			event:   "push",
			trusted: false,
			match:   false,
			matcher: Func([]string{"spaceghost/*", "octocat/*"}, []string{"push"}, true),
		},
	}

	for i, test := range tests {
		repo := &drone.Repo{
			Slug:    test.repo,
			Trusted: test.trusted,
		}
		build := &drone.Build{
			Event: test.event,
		}
		match := test.matcher(repo, build)
		if match != test.match {
			t.Errorf("Expect match %v at index %d", test.match, i)
		}
	}
}
