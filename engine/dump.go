// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"io"

	"github.com/ghodss/yaml"
)

const (
	documentBegin = "---\n"
	documentEnd   = "...\n"
)

// Dump encodes returns specification as a Kubernetes
// multi-document yaml configuration file, and writes
// to io.Writer w.
func Dump(w io.Writer, spec *Spec) {

	//
	// Secret Encoding.
	//

	{
		io.WriteString(w, documentBegin)
		res := toSecret(spec)
		res.Kind = "Secret"
		raw, _ := yaml.Marshal(res)
		w.Write(raw)
	}

	//
	// Pull Secret Encoding.
	//

	if spec.PullSecret != nil {
		io.WriteString(w, documentBegin)
		res := toDockerConfigSecret(spec)
		res.Kind = "Secret"
		raw, _ := yaml.Marshal(res)
		w.Write(raw)
	}

	//
	// Step Encoding.
	//

	{
		io.WriteString(w, documentBegin)
		res := toPod(spec)
		res.Kind = "Pod"
		raw, _ := yaml.Marshal(res)
		w.Write(raw)
	}

	io.WriteString(w, documentEnd)
}
