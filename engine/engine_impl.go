// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"context"
	"io"
)

// Kubernetes implements a Kubernetes pipeline engine.
type Kubernetes struct {
}

// New returns a new engine.
func New() *Kubernetes {
	return nil
}

// NewEnv returns a new Engine from the environment.
func NewEnv() (*Kubernetes, error) {
	return nil, nil
}

// Setup the pipeline environment.
func (e *Kubernetes) Setup(ctx context.Context, spec *Spec) error {
	return nil
}

// Destroy the pipeline environment.
func (e *Kubernetes) Destroy(ctx context.Context, spec *Spec) error {
	return nil
}

// Run runs the pipeline step.
func (e *Kubernetes) Run(ctx context.Context, spec *Spec, step *Step, output io.Writer) (*State, error) {
	return nil, nil
}
