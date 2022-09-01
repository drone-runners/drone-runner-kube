// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package policy

import "github.com/drone/runner-go/manifest"

// Match returns the matching Policy. If there is no matching
// Policy, but a default Policy is defined, the default Policy
// is returned. Otherwise a nil Policy is returned.
func Match(match manifest.Match, policy []*Policy) *Policy {
	for _, p := range policy {
		if p.Conditions.Match(match) {
			println("MATCHED POLICY " + p.Name)
			return p
		}
	}
	for _, p := range policy {
		if p.Name == "default" && !p.Exclude.Match(match) {
			println("MATCHED DEFAULT POLICY!!")
			return p
		}
	}
	println("!!! NO POLICY WAS MATCHED !!!")
	return nil
}
