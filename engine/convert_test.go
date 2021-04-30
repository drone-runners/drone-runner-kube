// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestSecurityContext(t *testing.T) {
	test := &Step{
		Privileged: true,
		User:       int64ptr(1000),
		Group:      int64ptr(1000),
	}

	want := v1.SecurityContext{
		Privileged: boolptr(true),
		RunAsUser:  int64ptr(1000),
		RunAsGroup: int64ptr(1000),
	}

	got := toSecurityContext(test)

	failed := *want.Privileged != *got.Privileged || *want.RunAsUser != *got.RunAsUser || *want.RunAsGroup != *got.RunAsGroup

	if failed {
		t.Error("security context was not converted to expected values")
	}
}
