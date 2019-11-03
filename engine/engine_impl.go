// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"context"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/client-go/util/retry"
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
	// TODO(bradrydzewski) revisit how we want to pass sensitive data
	// to the pipeline contianers.
	// for _, step := range spec.Steps {
	// 	_, err := e.client.CoreV1().Secrets(spec.PodSpec.Namespace).Create(toSecret(step))
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	_, err := e.client.CoreV1().Pods(spec.PodSpec.Namespace).Create(toPod(spec))
	if err != nil {
		return err
	}

	return nil
}

// Destroy the pipeline environment.
func (e *Kubernetes) Destroy(ctx context.Context, spec *Spec) error {
	// TODO(bradrydzewski) revisit how we want to pass sensitive data
	// to the pipeline contianers.
	// for _, step := range spec.Steps {
	// 	err := e.client.CoreV1().Secrets(spec.PodSpec.Namespace).Delete(step.ID, &metav1.DeleteOptions{})
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	err := e.client.CoreV1().Pods(spec.PodSpec.Namespace).Delete(spec.PodSpec.Name, &metav1.DeleteOptions{})
	return err
}

// Run runs the pipeline step.
func (e *Kubernetes) Run(ctx context.Context, spec *Spec, step *Step, output io.Writer) (*State, error) {
	err := e.start(spec, step)
	if err != nil {
		return nil, err
	}

	err = e.waitForReady(ctx, spec, step)
	if err != nil {
		return nil, err
	}

	err = e.tail(ctx, spec, step, output)
	if err != nil {
		return nil, err
	}

	return e.waitForTerminated(ctx, spec, step)
}

func (e *Kubernetes) waitFor(ctx context.Context, spec *Spec, conditionFunc func(e watch.Event) (bool, error)) error {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return e.client.CoreV1().Pods(spec.PodSpec.Namespace).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return e.client.CoreV1().Pods(spec.PodSpec.Namespace).Watch(metav1.ListOptions{})
		},
	}

	preconditionFunc := func(store cache.Store) (bool, error) {
		_, exists, err := store.Get(&metav1.ObjectMeta{Namespace: spec.PodSpec.Namespace, Name: spec.PodSpec.Name})
		if err != nil {
			return true, err
		}
		if !exists {
			return true, err
		}
		return false, nil
	}

	_, err := watchtools.UntilWithSync(ctx, lw, &v1.Pod{}, preconditionFunc, conditionFunc)
	return err
}

func (e *Kubernetes) waitForReady(ctx context.Context, spec *Spec, step *Step) error {
	return e.waitFor(ctx, spec, func(e watch.Event) (bool, error) {
		switch t := e.Type; t {
		case watch.Added, watch.Modified:
			pod := e.Object.(*v1.Pod)
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Name == step.ID && cs.Image != placeHolderImage && (cs.State.Running != nil || cs.State.Terminated != nil) {
					return true, nil
				}
			}
		}
		return false, nil
	})
}

func (e *Kubernetes) waitForTerminated(ctx context.Context, spec *Spec, step *Step) (*State, error) {
	state := &State{
		Exited:    true,
		OOMKilled: false,
	}
	err := e.waitFor(ctx, spec, func(e watch.Event) (bool, error) {
		switch t := e.Type; t {
		case watch.Added, watch.Modified:
			pod := e.Object.(*v1.Pod)
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Name == step.ID && cs.Image != placeHolderImage && (cs.State.Terminated != nil) {
					state.ExitCode = int(cs.State.Terminated.ExitCode)
					return true, nil
				}
			}
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (e *Kubernetes) tail(ctx context.Context, spec *Spec, step *Step, output io.Writer) error {
	req := e.client.CoreV1().RESTClient().Get().
		Namespace(spec.PodSpec.Namespace).
		Name(spec.PodSpec.Name).
		Resource("pods").
		SubResource("log").
		Param("follow", "true").
		Param("container", step.ID)

	readCloser, err := req.Stream()
	if err != nil {
		return err
	}

	defer readCloser.Close()
	_, err = io.Copy(output, readCloser)
	return err
}

func (e *Kubernetes) start(spec *Spec, step *Step) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pod, err := e.client.CoreV1().Pods(spec.PodSpec.Namespace).Get(spec.PodSpec.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for i, container := range pod.Spec.Containers {
			if container.Name == step.ID {
				pod.Spec.Containers[i].Image = step.Image
			}
		}

		_, err = e.client.CoreV1().Pods(spec.PodSpec.Namespace).Update(pod)
		return err
	})

	return err
}
