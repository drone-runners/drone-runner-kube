// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package policy

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/buildkite/yaml"
)

// ParseFile parses a policy file.
func ParseFile(f string) ([]*Policy, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

// Parse parses a policy.
func Parse(b []byte) ([]*Policy, error) {
	buf := bytes.NewBuffer(b)
	res := []*Policy{}
	dec := yaml.NewDecoder(buf)
	for {
		out := new(Policy)
		err := dec.Decode(out)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		res = append(res, out)
	}
	return res, nil
}
