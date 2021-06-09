// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/drone-runners/drone-runner-kube/engine/podwatcher"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

var backoff = wait.Backoff{
	Steps:    15,
	Duration: 500 * time.Millisecond,
	Factor:   1.0,
	Jitter:   0.5,
}

// Kubernetes implements a Kubernetes pipeline engine.
type Kubernetes struct {
	client  *kubernetes.Clientset
	watcher *podwatcher.PodWatcher
}

// New returns a new engine. It tries first with in-cluster config, if it fails it will try with out-of-cluster config.
func New() (engine runtime.Engine, err error) {
	engine, err = NewInCluster()
	if err == nil {
		return
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir = filepath.Join(dir, ".kube", "config")

	engine, err = NewFromConfig(dir)
	if err != nil {
		return
	}

	return
}

// NewFromConfig returns a new out-of-cluster engine.
func NewFromConfig(path string) (engine runtime.Engine, err error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	engine = &Kubernetes{
		client: clientset,
	}

	return
}

// NewInCluster returns a new in-cluster engine.
func NewInCluster() (engine runtime.Engine, err error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	engine = &Kubernetes{
		client: clientset,
	}

	return
}

// Setup the pipeline environment.
func (k *Kubernetes) Setup(ctx context.Context, specv runtime.Spec) error {
	spec := specv.(*Spec)

	if spec.Namespace != "" {
		_, err := k.client.CoreV1().Namespaces().Create(toNamespace(spec.Namespace, spec.PodSpec.Labels))
		if err != nil {
			return err
		}
	}

	if spec.PullSecret != nil {
		_, err := k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Create(toDockerConfigSecret(spec))
		if err != nil {
			return err
		}
	}

	_, err := k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Create(toSecret(spec))
	if err != nil {
		return err
	}

	pod, err := k.client.CoreV1().Pods(spec.PodSpec.Namespace).Create(toPod(spec))
	if err != nil {
		return err
	}

	// Start a watcher that watches all k8s event related to the pod just created.
	// The watcher will finish when the pod is deleted.
	k.watcher = &podwatcher.PodWatcher{}
	k.watcher.Start(ctx, &podwatcher.KubernetesWatcher{
		PodNamespace: pod.Namespace,
		PodName:      pod.Name,
		Clientset:    k.client,
		Period:       time.Minute,
	})

	return nil
}

// Destroy the pipeline environment.
func (k *Kubernetes) Destroy(ctx context.Context, specv runtime.Spec) error {
	// HACK: this timeout delays deleting the Pod to ensure
	// there is enough time to stream the logs.
	time.Sleep(time.Second * 5)

	spec := specv.(*Spec)
	var result error

	if spec.PullSecret != nil {
		err := k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Delete(spec.PullSecret.Name, &metav1.DeleteOptions{})
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	err := k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Delete(spec.PodSpec.Name, &metav1.DeleteOptions{})
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = k.client.CoreV1().Pods(spec.PodSpec.Namespace).Delete(spec.PodSpec.Name, &metav1.DeleteOptions{})
	if err != nil {
		result = multierror.Append(result, err)
	}

	if spec.Namespace != "" {
		err := k.client.CoreV1().Namespaces().Delete(spec.Namespace, &metav1.DeleteOptions{})
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	err = k.watcher.WaitPodDeleted()
	if err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

// Run runs the pipeline step.
func (k *Kubernetes) Run(ctx context.Context, specv runtime.Spec, stepv runtime.Step, output io.Writer) (*runtime.State, error) {
	spec := specv.(*Spec)
	step := stepv.(*Step)

	// start tracking the container with this name
	k.watcher.AddContainer(step.ID, step.Placeholder)

	err := k.start(spec, step)
	if err != nil {
		return nil, err
	}

	logrus.Tracef("Step %q: %q (%s)\n", step.ID, step.Name, step.Image)

	err = k.watcher.WaitContainerStart(step.ID)
	if err != nil {
		return nil, err
	}

	err = k.tail(ctx, spec, step, output)
	// this feature flag retries fetching the logs if it fails on
	// the first attempt. This is meant to help triage the following
	// issue:
	//
	//    https://discourse.drone.io/t/kubernetes-runner-intermittently-fails-steps/7372
	//
	// BEGIN: FEATURE FLAG
	if err != nil {
		if os.Getenv("DRONE_FEATURE_FLAG_RETRY_LOGS") == "true" {
			<-time.After(time.Second * 5)
			err = k.tail(ctx, spec, step, output)
		}
		if err != nil {
			<-time.After(time.Second * 5)
			err = k.tail(ctx, spec, step, output)
		}
	}
	// END: FEATURE FAG
	if err != nil {
		return nil, err
	}

	code, err := k.watcher.WaitContainerTerminated(step.ID)
	if err != nil {
		return nil, err
	}

	state := &runtime.State{
		ExitCode:  code,
		Exited:    true,
		OOMKilled: false,
	}

	return state, nil
}

func (k *Kubernetes) tail(ctx context.Context, spec *Spec, step *Step, output io.Writer) error {
	opts := &v1.PodLogOptions{
		Follow:    true,
		Container: step.ID,
	}

	req := k.client.CoreV1().RESTClient().Get().
		Namespace(spec.PodSpec.Namespace).
		Name(spec.PodSpec.Name).
		Resource("pods").
		SubResource("log").
		VersionedParams(opts, scheme.ParameterCodec)

	readCloser, err := req.Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	return cancellableCopy(ctx, output, readCloser)
}

func (k *Kubernetes) start(spec *Spec, step *Step) error {
	err := retry.RetryOnConflict(backoff, func() error {
		// We protect this read/modify/write with a mutex to reduce the
		// chance of a self-inflicted concurrent modification error
		// when a DAG in a pipeline is fanning out and we have a lot of
		// steps to Start at once.
		spec.podUpdateMutex.Lock()
		defer spec.podUpdateMutex.Unlock()

		pod, err := k.client.CoreV1().Pods(spec.PodSpec.Namespace).Get(spec.PodSpec.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for i, container := range pod.Spec.Containers {
			if container.Name != step.ID {
				continue
			}

			pod.Spec.Containers[i].Image = step.Image

			if pod.ObjectMeta.Annotations == nil {
				pod.ObjectMeta.Annotations = map[string]string{}
			}
			for _, env := range statusesWhiteList {
				pod.ObjectMeta.Annotations[env] = step.Envs[env]
			}
		}

		_, err = k.client.CoreV1().Pods(spec.PodSpec.Namespace).Update(pod)
		return err
	})

	return err
}
