// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/drone-runners/drone-runner-kube/engine/launcher"
	"github.com/drone-runners/drone-runner-kube/engine/podwatcher"

	"github.com/drone/runner-go/logger"
	"github.com/drone/runner-go/pipeline/runtime"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Kubernetes implements a Kubernetes pipeline engine.
type Kubernetes struct {
	client    *kubernetes.Clientset
	watchers  *sync.Map
	launchers *sync.Map
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
		client:    clientset,
		watchers:  &sync.Map{},
		launchers: &sync.Map{},
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
		client:    clientset,
		watchers:  &sync.Map{},
		launchers: &sync.Map{},
	}

	return
}

// Setup the pipeline environment.
func (k *Kubernetes) Setup(ctx context.Context, specv runtime.Spec) (err error) {
	spec := specv.(*Spec)

	if spec.Namespace != "" {
		_, err = k.client.CoreV1().Namespaces().Create(toNamespace(spec.Namespace, spec.PodSpec.Labels))
		if err != nil {
			return err
		}
	}

	if spec.PullSecret != nil {
		_, err = k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Create(toDockerConfigSecret(spec))
		if err != nil {
			return err
		}
	}

	_, err = k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Create(toSecret(spec))
	if err != nil {
		return err
	}

	_, err = k.client.CoreV1().Pods(spec.PodSpec.Namespace).Create(toPod(spec))
	if err != nil {
		return err
	}

	return nil
}

// Destroy the pipeline environment.
func (k *Kubernetes) Destroy(ctx context.Context, specv runtime.Spec) error {
	// HACK: this timeout delays deleting the Pod to ensure
	// there is enough time to stream the logs.
	time.Sleep(time.Second * 5)

	spec := specv.(*Spec)

	if spec.PullSecret != nil {
		if err := k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Delete(spec.PullSecret.Name, &metav1.DeleteOptions{}); err != nil {
			logger.FromContext(ctx).
				WithError(err).
				WithField("pull-secret", spec.PullSecret.Name).
				WithField("namespace", spec.PodSpec.Namespace).
				Error("failed to delete pull secret")
		}
	}

	if err := k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Delete(spec.PodSpec.Name, &metav1.DeleteOptions{}); err != nil {
		logger.FromContext(ctx).
			WithError(err).
			WithField("pod", spec.PodSpec.Name).
			WithField("namespace", spec.PodSpec.Namespace).
			Error("failed to delete secrets")
	}

	if err := k.client.CoreV1().Pods(spec.PodSpec.Namespace).Delete(spec.PodSpec.Name, &metav1.DeleteOptions{}); err != nil {
		logger.FromContext(ctx).
			WithError(err).
			WithField("pod", spec.PodSpec.Name).
			WithField("namespace", spec.PodSpec.Namespace).
			Error("failed to delete pod")
	}

	if spec.Namespace != "" {
		if err := k.client.CoreV1().Namespaces().Delete(spec.Namespace, &metav1.DeleteOptions{}); err != nil {
			logger.FromContext(ctx).
				WithError(err).
				WithField("namespace", spec.PodSpec.Namespace).
				Error("failed to delete namespace")
		}
	}

	if _l, loaded := k.launchers.LoadAndDelete(spec.PodSpec.Name); loaded {
		l := _l.(*launcher.Launcher)
		l.Stop()
	}

	if w, loaded := k.watchers.LoadAndDelete(spec.PodSpec.Name); loaded {
		watcher := w.(*podwatcher.PodWatcher)
		if err := watcher.WaitPodDeleted(); err != nil && err != context.Canceled {
			logger.FromContext(ctx).
				WithError(err).
				WithField("pod", spec.PodSpec.Name).
				WithField("namespace", spec.PodSpec.Namespace).
				Error("PodWatcher terminated with error")
		} else {
			logger.FromContext(ctx).
				WithField("pod", spec.PodSpec.Name).
				WithField("namespace", spec.PodSpec.Namespace).
				Debug("PodWatcher terminated")
		}
	}

	return nil
}

// Run runs the pipeline step.
func (k *Kubernetes) Run(ctx context.Context, specv runtime.Spec, stepv runtime.Step, output io.Writer) (state *runtime.State, err error) {
	spec := specv.(*Spec)
	step := stepv.(*Step)

	podId := spec.PodSpec.Name
	podNamespace := spec.PodSpec.Namespace
	stepName := step.Name
	containerId := step.ID
	containerImage := step.Image
	containerPlaceholder := step.Placeholder

	w, loaded := k.watchers.LoadOrStore(podId, &podwatcher.PodWatcher{})
	watcher := w.(*podwatcher.PodWatcher)
	if !loaded {
		watcher.Start(context.Background(), &podwatcher.KubernetesWatcher{
			PodNamespace: podNamespace,
			PodName:      podId,
			Clientset:    k.client,
			Period:       20 * time.Second,
		})

		logger.FromContext(ctx).
			WithField("pod", podId).
			WithField("step", stepName).
			Debug("PodWatcher started")
	}

	err = watcher.AddContainer(step.ID, step.Placeholder)
	if err != nil {
		return
	}

	logger.FromContext(ctx).
		WithField("pod", podId).
		WithField("container", containerId).
		WithField("image", containerImage).
		WithField("placeholder", containerPlaceholder).
		WithField("step", stepName).
		Debugf("Engine: Starting step")

	err = <-k.startContainer(ctx, spec, step)
	if err != nil {
		return
	}

	chErrStart := make(chan error)
	go func() {
		chErrStart <- watcher.WaitContainerStart(containerId)
	}()

	select {
	case err = <-chErrStart:
	case <-time.After(3 * time.Minute):
		err = podwatcher.StartTimeoutContainerError{Container: containerId}
		logger.FromContext(ctx).
			WithError(err).
			WithField("pod", podId).
			WithField("container", containerId).
			WithField("image", containerImage).
			WithField("placeholder", containerPlaceholder).
			WithField("step", stepName).
			Warnf("Engine: Container start timeout")
	}
	if err != nil {
		return
	}

	err = k.fetchLogs(ctx, spec, step, output)
	if err != nil {
		return
	}

	code, err := watcher.WaitContainerTerminated(containerId)
	if err != nil {
		return
	}

	state = &runtime.State{
		ExitCode:  code,
		Exited:    true,
		OOMKilled: false,
	}

	return
}

func (k *Kubernetes) fetchLogs(ctx context.Context, spec *Spec, step *Step, output io.Writer) error {
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

func (k *Kubernetes) startContainer(ctx context.Context, spec *Spec, step *Step) <-chan error {
	podName := spec.PodSpec.Name
	podNamespace := spec.PodSpec.Namespace
	containerName := step.ID
	containerImage := step.Image

	_l, loaded := k.launchers.LoadOrStore(podName, launcher.New(podName, podNamespace, k.client, &spec.podUpdateMutex))
	l := _l.(*launcher.Launcher)
	if !loaded {
		l.Start(ctx)
	}

	statusEnvs := make(map[string]string)
	for _, env := range statusesWhiteList {
		statusEnvs[env] = step.Envs[env]
	}

	return l.Launch(containerName, containerImage, statusEnvs)
}
