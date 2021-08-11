// Copyright 2021 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package launcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

const aggregateTimer = 400 * time.Millisecond

// Launcher is used to launch several containers at once. It uses a timer to
// collect potentially several container launch events.
// Than the containers are updates with a single call to Kubernetes API.
type Launcher struct {
	stop, stopped chan struct{}

	kubeClient  kubernetes.Interface
	podUpdateMx *sync.Mutex

	podNamespace string
	podName      string

	timer     *time.Timer
	requests  map[string]*request
	requestCh chan request
}

type request struct {
	containerID    string
	containerImage string
	chErr          chan error
	statusEnvs     map[string]string
	found          bool
}

// New creates a new Launcher.
func New(podName, podNamespace string, clientset kubernetes.Interface, podUpdateMx *sync.Mutex) *Launcher {
	l := &Launcher{
		stop:         make(chan struct{}),
		stopped:      make(chan struct{}),
		kubeClient:   clientset,
		podUpdateMx:  podUpdateMx,
		podNamespace: podNamespace,
		podName:      podName,
		timer:        nil,
		requests:     nil,
		requestCh:    make(chan request),
	}

	return l
}

// Stop terminates Launcher's main go routine.
func (l *Launcher) Stop() {
	close(l.stop)
	<-l.stopped
}

// Start starts Launcher's main go routine.
func (l *Launcher) Start(ctx context.Context) {
	if l.timer != nil {
		panic("timer already started")
	}

	t := time.NewTimer(time.Second)
	t.Stop()

	l.timer = t

	go func() {
		defer close(l.stopped)

		for {
			select {
			case <-ctx.Done():
				return

			case <-l.stop:
				return

			case <-l.timer.C:
				l.startContainers(l.requests)
				l.requests = nil

			case req := <-l.requestCh:
				if l.requests == nil {
					l.requests = make(map[string]*request)
				}
				l.requests[req.containerID] = &req

				if !l.timer.Stop() {
					select {
					default:
					case <-l.timer.C:
					}
				}

				l.timer.Reset(aggregateTimer)
			}
		}
	}()
}

// Launch schedules launch of a pod's container.
func (l *Launcher) Launch(containerID, containerImage string, statusEnvs map[string]string) <-chan error {
	chErr := make(chan error)
	l.requestCh <- request{
		containerID:    containerID,
		containerImage: containerImage,
		chErr:          chErr,
		statusEnvs:     statusEnvs,
		found:          false,
	}

	return chErr
}

func (l *Launcher) startContainers(requests map[string]*request) {
	var backoff = wait.Backoff{
		Steps:    15,
		Duration: 500 * time.Millisecond,
		Factor:   1.0,
		Jitter:   0.5,
	}

	t := time.Now()

	err := retry.RetryOnConflict(backoff, func() error {
		// We protect this read/modify/write with a mutex to reduce the
		// chance of a self-inflicted concurrent modification error
		// when a DAG in a pipeline is fanning out and we have a lot of
		// steps to Start at once.
		l.podUpdateMx.Lock()
		defer l.podUpdateMx.Unlock()

		pod, err := l.kubeClient.CoreV1().Pods(l.podNamespace).Get(l.podName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for i, container := range pod.Spec.Containers {
			req, ok := requests[container.Name]
			if !ok {
				continue
			}

			req.found = true

			pod.Spec.Containers[i].Image = req.containerImage

			if pod.ObjectMeta.Annotations == nil {
				pod.ObjectMeta.Annotations = map[string]string{}
			}
			for envName, envValue := range req.statusEnvs {
				pod.ObjectMeta.Annotations[envName] = envValue
			}
		}

		_, err = l.kubeClient.CoreV1().Pods(l.podNamespace).Update(pod)

		return err
	})

	var notFounds int
	for _, req := range requests {
		if !req.found {
			notFounds++
		}
	}

	logrus.
		WithField("count", len(requests)).
		WithField("failed", notFounds).
		WithField("success", len(requests)-notFounds).
		Debugf("Launched containers. Duration=%.2fs", time.Since(t).Seconds())

	for _, req := range requests {
		if !req.found {
			req.chErr <- fmt.Errorf("container %s not found in pod %s", req.containerID, l.podName)
			continue
		}

		req.chErr <- err
	}
}
