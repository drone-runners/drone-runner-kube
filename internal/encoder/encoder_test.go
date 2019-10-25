// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package encoder

import "testing"

func TestEncode(t *testing.T) {
	testdatum := []struct {
		data interface{}
		text string
	}{
		{
			data: "foo",
			text: "foo",
		},
		{
			data: true,
			text: "true",
		},
		{
			data: 42,
			text: "42",
		},
		{
			data: float64(42.424242),
			text: "42.424242",
		},
		{
			data: []interface{}{"foo", "bar", "baz"},
			text: "foo,bar,baz",
		},
		{
			data: []interface{}{1, 1, 2, 3, 5, 8},
			text: "1,1,2,3,5,8",
		},
		{
			data: []byte("foo"),
			text: "Zm9v",
		},
		{
			data: []interface{}{
				struct {
					Name string `json:"name"`
				}{
					Name: "john",
				},
			},
			text: `[{"name":"john"}]`,
		},
		{
			data: map[interface{}]interface{}{"foo": "bar"},
			text: `{"foo":"bar"}`,
		},
	}

	for _, testdata := range testdatum {
		if got, want := Encode(testdata.data), testdata.text; got != want {
			t.Errorf("Want interface{} encoded to %q, got %q", want, got)
		}
	}
}
