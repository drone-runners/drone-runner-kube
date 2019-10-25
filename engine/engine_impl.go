// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"context"
	"io"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Kubernetes implements a Kubernetes pipeline engine.
type Kubernetes struct {
	client *kubernetes.Clientset
}

// NewFromConfig returns a new out-of-cluster engine.
func NewFromConfig(path string) (*Kubernetes, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Kubernetes{clientset}, nil
}

// NewInCluster returns a new in-cluster engine.
func NewInCluster() (*Kubernetes, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Kubernetes{clientset}, nil
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
