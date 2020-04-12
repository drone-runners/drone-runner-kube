// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"bytes"
	"encoding/json"
	"testing"
)

//
// pull policy unit tests.
//

func TestPullPolicy_Marshal(t *testing.T) {
	tests := []struct {
		policy PullPolicy
		data   string
	}{
		{
			policy: PullAlways,
			data:   `"always"`,
		},
		{
			policy: PullDefault,
			data:   `"default"`,
		},
		{
			policy: PullIfNotExists,
			data:   `"if-not-exists"`,
		},
		{
			policy: PullNever,
			data:   `"never"`,
		},
	}
	for _, test := range tests {
		data, err := json.Marshal(&test.policy)
		if err != nil {
			t.Error(err)
			return
		}
		if bytes.Equal([]byte(test.data), data) == false {
			t.Errorf("Failed to marshal policy %s", test.policy)
		}
	}
}

func TestPullPolicy_Unmarshal(t *testing.T) {
	tests := []struct {
		policy PullPolicy
		data   string
	}{
		{
			policy: PullAlways,
			data:   `"always"`,
		},
		{
			policy: PullDefault,
			data:   `"default"`,
		},
		{
			policy: PullIfNotExists,
			data:   `"if-not-exists"`,
		},
		{
			policy: PullNever,
			data:   `"never"`,
		},
		{
			// no policy should default to on-success
			policy: PullDefault,
			data:   `""`,
		},
	}
	for _, test := range tests {
		var policy PullPolicy
		err := json.Unmarshal([]byte(test.data), &policy)
		if err != nil {
			t.Error(err)
			return
		}
		if got, want := policy, test.policy; got != want {
			t.Errorf("Want policy %q, got %q", want, got)
		}
	}
}

func TestPullPolicy_UnmarshalTypeError(t *testing.T) {
	var policy PullPolicy
	err := json.Unmarshal([]byte("[]"), &policy)
	if _, ok := err.(*json.UnmarshalTypeError); !ok {
		t.Errorf("Expect unmarshal error return when JSON invalid")
	}
}

func TestPullPolicy_String(t *testing.T) {
	tests := []struct {
		policy PullPolicy
		value  string
	}{
		{
			policy: PullAlways,
			value:  "always",
		},
		{
			policy: PullDefault,
			value:  "default",
		},
		{
			policy: PullIfNotExists,
			value:  "if-not-exists",
		},
		{
			policy: PullNever,
			value:  "never",
		},
	}
	for _, test := range tests {
		if got, want := test.policy.String(), test.value; got != want {
			t.Errorf("Want policy string %q, got %q", want, got)
		}
	}
}
