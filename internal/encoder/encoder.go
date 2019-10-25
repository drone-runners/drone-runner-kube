// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package encoder

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/buildkite/yaml"
	json "github.com/ghodss/yaml"
)

// Encode encodes an interface value as a string. This function
// assumes all types were unmarshaled by the yaml.v2 library.
// The yaml.v2 package only supports a subset of primitive types.
func Encode(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case []byte:
		return base64.StdEncoding.EncodeToString(v)
	case []interface{}:
		return encodeSlice(v)
	default:
		return encodeMap(v)
	}
}

// helper function encodes a parameter in map format.
func encodeMap(v interface{}) string {
	yml, _ := yaml.Marshal(v)
	out, _ := json.YAMLToJSON(yml)
	return string(out)
}

// helper function encodes a parameter in slice format.
func encodeSlice(v interface{}) string {
	out, _ := yaml.Marshal(v)

	in := []string{}
	err := yaml.Unmarshal(out, &in)
	if err == nil {
		return strings.Join(in, ",")
	}
	out, _ = json.YAMLToJSON(out)
	return string(out)
}
