// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package resource

import (
	"testing"

	"github.com/drone/runner-go/manifest"
)

func TestLookup(t *testing.T) {
	want := &Pipeline{Name: "default"}
	m := &manifest.Manifest{
		Resources: []manifest.Resource{want},
	}
	got, err := Lookup("default", m)
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("Expect resource not found error")
	}
}

func TestLookupNotFound(t *testing.T) {
	m := &manifest.Manifest{
		Resources: []manifest.Resource{
			&manifest.Secret{
				Kind: "secret",
				Name: "password",
			},
			// matches name, but is not of kind pipeline
			&manifest.Secret{
				Kind: "secret",
				Name: "default",
			},
		},
	}
	_, err := Lookup("default", m)
	if err == nil {
		t.Errorf("Expect resource not found error")
	}
}
