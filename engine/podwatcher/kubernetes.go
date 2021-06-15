// Copyright 2021 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package podwatcher

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
)

// KubernetesWatcher is a struct that implements ContainerWatcher interface by watching a kubernetes pod.
type KubernetesWatcher struct {
	PodNamespace string
	PodName      string
	Clientset    *kubernetes.Clientset
	Period       time.Duration
}

func (w *KubernetesWatcher) Name() string {
	return w.PodName
}

// Watch is a part of ContainerWatcher implementation for the KubernetesWatcher struct.
// It will create a Kubernetes watcher that watches all events coming from a specific pod.
// The method will run until the pod terminates and is deleted (until the "Deleted" event arrives).
func (w *KubernetesWatcher) Watch(ctx context.Context, containers chan<- []containerInfo) error {
	label := "io.drone.name=" + w.PodName

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (k8sruntime.Object, error) {
			options.LabelSelector = label
			return w.Clientset.CoreV1().Pods(w.PodNamespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.LabelSelector = label
			return w.Clientset.CoreV1().Pods(w.PodNamespace).Watch(options)
		},
	}

	preconditionFunc := func(store cache.Store) (stop bool, err error) {
		_, exists, err := store.Get(&metav1.ObjectMeta{Namespace: w.PodNamespace, Name: w.PodName})
		stop = err != nil || !exists
		return
	}

	_, err := watchtools.UntilWithSync(ctx, lw, &v1.Pod{}, preconditionFunc, func(event watch.Event) (bool, error) {
		pod, ok := event.Object.(*v1.Pod)
		if !ok || pod.Name != w.PodName {
			return false, nil
		}

		logrus.
			WithField("pod", pod.Name).
			WithField("event", event.Type).
			Trace("PodWatcher: Event")

		if event.Type == watch.Deleted {
			return true, nil // stop listening to further events
		}

		containers <- extractContainers(pod)

		return false, nil
	})

	return err
}

// PeriodicCheck is a part of ContainerWatcher implementation for the KubernetesWatcher struct.
func (w *KubernetesWatcher) PeriodicCheck(ctx context.Context, containers chan<- []containerInfo, stop <-chan struct{}) error {
	if w.Period == 0 {
		return nil
	}

	ticker := time.NewTicker(w.Period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-stop:
			return nil

		case <-ticker.C:
			pod, err := w.Clientset.CoreV1().Pods(w.PodNamespace).Get(w.PodName, metav1.GetOptions{})
			if err != nil {
				logrus.
					WithError(err).
					WithField("pod", w.PodName).
					Warn("PodWatcher: Failed to read pod")
				continue
			}

			logrus.
				WithField("pod", w.PodName).
				Trace("PodWatcher: Periodic container state check")

			containers <- extractContainers(pod)
		}
	}
}

func extractContainers(pod *v1.Pod) (result []containerInfo) {
	if pod == nil {
		return
	}

	result = make([]containerInfo, len(pod.Status.ContainerStatuses))

	for i, cs := range pod.Status.ContainerStatuses {
		var (
			state     containerState
			stateInfo string
			exitCode  int32
		)

		if cs.State.Terminated != nil {
			state, stateInfo = stateTerminated, cs.State.Terminated.Reason
			exitCode = cs.State.Terminated.ExitCode
		} else if cs.State.Running != nil {
			state, stateInfo = stateRunning, ""
		} else if cs.State.Waiting != nil {
			state, stateInfo = stateWaiting, cs.State.Waiting.Reason
		} else {
			// kubernetes doc explains that this situation should be treated as Waiting state
			state, stateInfo = stateWaiting, ""
		}

		result[i] = containerInfo{
			id:        cs.Name,
			state:     state,
			stateInfo: stateInfo,
			image:     cs.Image,
			exitCode:  exitCode,
		}
	}

	return
}
