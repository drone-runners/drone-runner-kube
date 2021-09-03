// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"context"
	"io"
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
)

// Kubernetes implements a Kubernetes pipeline engine.
type Kubernetes struct {
	client    kubernetes.Interface
	watchers  *sync.Map
	launchers *sync.Map
}

// New returns a new engine with the provided kubernetes client
func New(client kubernetes.Interface) runtime.Engine {
	return &Kubernetes{
		client:    client,
		watchers:  &sync.Map{},
		launchers: &sync.Map{},
	}
}

// Setup the pipeline environment.
func (k *Kubernetes) Setup(ctx context.Context, specv runtime.Spec) (err error) {
	spec := specv.(*Spec)

	log := logger.FromContext(ctx).
		WithField("pod", spec.PodSpec.Name).
		WithField("namespace", spec.PodSpec.Namespace)

	if spec.Namespace != "" {
		namespace := toNamespace(spec.Namespace, spec.PodSpec.Labels)
		_, err = k.client.CoreV1().Namespaces().Create(namespace)
		if err != nil {
			log.WithError(err).Error("failed to create namespace")
			return err
		}
		log.Trace("created namespace")
	}

	if spec.PullSecret != nil {
		pullSecret := toDockerConfigSecret(spec)
		_, err = k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Create(pullSecret)
		if err != nil {
			log.WithError(err).Error("failed to create pull secret")
			return err
		}
		log.Trace("created pull secret")
	}

	secret := toSecret(spec)
	_, err = k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Create(secret)
	if err != nil {
		log.WithError(err).Error("failed to create secret")
		return err
	}
	log.Trace("created secret")

	_, err = k.client.CoreV1().Pods(spec.PodSpec.Namespace).Create(toPod(spec))
	if err != nil {
		log.WithError(err).Error("failed to create pod")
		return err
	}
	log.Trace("created pod")

	return nil
}

// Destroy the pipeline environment.
func (k *Kubernetes) Destroy(ctx context.Context, specv runtime.Spec) error {
	// HACK: this timeout delays deleting the Pod to ensure
	// there is enough time to stream the logs.
	time.Sleep(time.Second * 5)

	spec := specv.(*Spec)

	log := logger.FromContext(ctx).
		WithField("pod", spec.PodSpec.Name).
		WithField("namespace", spec.PodSpec.Namespace)

	if spec.PullSecret != nil {
		if err := k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Delete(spec.PullSecret.Name, &metav1.DeleteOptions{}); err != nil {
			log.WithError(err).Error("failed to delete pull secret")
		} else {
			log.Trace("deleted pull secret")
		}
	}

	if err := k.client.CoreV1().Secrets(spec.PodSpec.Namespace).Delete(spec.PodSpec.Name, &metav1.DeleteOptions{}); err != nil {
		log.WithError(err).Error("failed to delete secret")
	} else {
		log.Trace("deleted secret")
	}

	if err := k.client.CoreV1().Pods(spec.PodSpec.Namespace).Delete(spec.PodSpec.Name, &metav1.DeleteOptions{}); err != nil {
		log.WithError(err).Error("failed to delete pod")
	} else {
		log.Trace("deleted pod")
	}

	if spec.Namespace != "" {
		if err := k.client.CoreV1().Namespaces().Delete(spec.Namespace, &metav1.DeleteOptions{}); err != nil {
			log.WithError(err).Error("failed to delete namespace")
		} else {
			log.Trace("deleted namespace")
		}
	}

	if _l, loaded := k.launchers.LoadAndDelete(spec.PodSpec.Name); loaded {
		l := _l.(*launcher.Launcher)
		l.Stop()
	}

	if w, loaded := k.watchers.LoadAndDelete(spec.PodSpec.Name); loaded {
		watcher := w.(*podwatcher.PodWatcher)
		if err := watcher.WaitPodDeleted(); err != nil && err != context.Canceled {
			log.WithError(err).Error("PodWatcher terminated with error")
		} else {
			log.Trace("PodWatcher terminated")
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

	log := logger.FromContext(ctx).
		WithField("pod", podId).
		WithField("namespace", podNamespace).
		WithField("image", containerImage).
		WithField("placeholder", containerPlaceholder).
		WithField("container", containerId).
		WithField("step", stepName)

	w, loaded := k.watchers.LoadOrStore(podId, &podwatcher.PodWatcher{})
	watcher := w.(*podwatcher.PodWatcher)
	if !loaded {
		watcher.Start(context.Background(), &podwatcher.KubernetesWatcher{
			PodNamespace: podNamespace,
			PodName:      podId,
			KubeClient:   k.client,
			Period:       20 * time.Second,
		})

		log.Trace("PodWatcher started")
	}

	err = watcher.AddContainer(step.ID, step.Placeholder)
	if err != nil {
		return
	}

	log.Debug("Engine: Starting step")

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
	case <-time.After(8 * time.Minute):
		err = podwatcher.StartTimeoutContainerError{Container: containerId}
		log.WithError(err).Error("Engine: Container start timeout")
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
		logger.FromContext(ctx).
			WithError(err).
			WithField("pod", spec.PodSpec.Name).
			WithField("namespace", spec.PodSpec.Namespace).
			WithField("container", step.ID).
			WithField("step", step.Name).
			Error("failed to stream logs")
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
