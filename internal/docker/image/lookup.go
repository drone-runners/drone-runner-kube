// Copyright 2020 Drone.IO Inc.
// Copyright 2019 The Tekton Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package image

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// LookupFunc defines a function to Lookup image information
// from a remote registry.
type LookupFunc func(image, username, password string) (*Image, error)

// Image stores the image configuration.
type Image struct {
	Entrypoint []string
	Command    []string
}

// Lookup returns the image metadata from the remote docker
// registry, authenticating if needed.
func Lookup(image, username, password string) (*Image, error) {
	ref, err := name.ParseReference(image, name.WeakValidation)
	if err != nil {
		return nil, err
	}

	var img v1.Image
	if len(username) != 0 || len(password) != 0 {
		img, err = remote.Image(ref,
			remote.WithAuth(&authn.Basic{
				Username: username,
				Password: password,
			}),
		)
	} else {
		img, err = remote.Image(ref)
	}
	if err != nil {
		return nil, err
	}

	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, err
	}

	return &Image{
		Entrypoint: cfg.Config.Entrypoint,
		Command:    cfg.Config.Cmd,
	}, nil
}
