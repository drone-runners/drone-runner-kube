// Code generated automatically. DO NOT EDIT.

// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package replacer

import (
	"fmt"
	"io"
	"strings"

	"github.com/drone-runners/drone-runner-kube/engine"
)

const maskedf = "[secret:%s]"

// Replacer is an io.Writer that finds and masks sensitive data.
type Replacer struct {
	w io.WriteCloser
	r *strings.Replacer
}

// New returns a replacer that wraps writer w.
func New(w io.WriteCloser, secrets []*engine.Secret) io.WriteCloser {
	var oldnew []string
	for _, secret := range secrets {
		if len(secret.Data) == 0 || secret.Mask == false {
			continue
		}
		name := strings.ToLower(secret.Name)
		masked := fmt.Sprintf(maskedf, name)
		oldnew = append(oldnew, string(secret.Data))
		oldnew = append(oldnew, masked)
	}
	if len(oldnew) == 0 {
		return w
	}
	return &Replacer{
		w: w,
		r: strings.NewReplacer(oldnew...),
	}
}

// Write writes p to the base writer. The method scans for any
// sensitive data in p and masks before writing.
func (r *Replacer) Write(p []byte) (n int, err error) {
	_, err = r.w.Write([]byte(r.r.Replace(string(p))))
	return len(p), err
}

// Close closes the base writer.
func (r *Replacer) Close() error {
	return r.w.Close()
}
