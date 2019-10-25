// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package match

import (
	"path/filepath"

	"github.com/drone/drone-go/drone"
)

// NOTE most runners do not require match capabilities. This is
// provided as a defense in depth mechanism given the sensitive
// nature of this runner executing code directly on the host.
// The matching function is a last line of defence to prevent
// unauthorized code from running on the host machine.

// Func returns a new match function that returns true if the
// repository and build do not match the allowd repository names
// and build events.
func Func(repos, events []string, trusted bool) func(*drone.Repo, *drone.Build) bool {
	return func(repo *drone.Repo, build *drone.Build) bool {
		// if trusted mode is enabled, only match repositories
		// that are trusted.
		if trusted && repo.Trusted == false {
			return false
		}
		if match(repo.Slug, repos) == false {
			return false
		}
		if match(build.Event, events) == false {
			return false
		}
		return true
	}
}

func match(s string, patterns []string) bool {
	// if no matching patterns are defined the string
	// is always considered a match.
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		if match, _ := filepath.Match(pattern, s); match {
			return true
		}
	}
	return false
}
