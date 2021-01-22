// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"bytes"
	"encoding/json"
)

var statusesWhiteList = []string{
	"DRONE_BUILD_STATUS",
	"DRONE_BUILD_FINISHED",
	"DRONE_STAGE_STATUS",
	"DRONE_STAGE_FINISHED",
	"DRONE_FAILED_STEPS",
	"DRONE_FAILED_STAGES",
}

// PullPolicy defines the container image pull policy.
type PullPolicy int

// PullPolicy enumeration.
const (
	PullDefault PullPolicy = iota
	PullAlways
	PullIfNotExists
	PullNever
)

func (p PullPolicy) String() string {
	return pullPolicyID[p]
}

var pullPolicyID = map[PullPolicy]string{
	PullDefault:     "default",
	PullAlways:      "always",
	PullIfNotExists: "if-not-exists",
	PullNever:       "never",
}

var pullPolicyName = map[string]PullPolicy{
	"":              PullDefault,
	"default":       PullDefault,
	"always":        PullAlways,
	"if-not-exists": PullIfNotExists,
	"never":         PullNever,
}

// MarshalJSON marshals the string representation of the
// pull type to JSON.
func (p *PullPolicy) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(pullPolicyID[*p])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshals the json representation of the
// pull type from a string value.
func (p *PullPolicy) UnmarshalJSON(b []byte) error {
	// unmarshal as string
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	// lookup value
	*p = pullPolicyName[s]
	return nil
}
