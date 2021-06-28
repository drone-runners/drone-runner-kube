// Copyright 2021 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package podwatcher

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

const placeholder = "placeholder"

type testContainerWatcher struct {
	containers []containerInfo
	finish     chan struct{}
	event      chan containerInfo
	tick       chan struct{}
}

func makeContainerWatcher(containers []string) *testContainerWatcher {
	cw := &testContainerWatcher{
		containers: make([]containerInfo, len(containers)),
		finish:     make(chan struct{}),
		event:      make(chan containerInfo),
		tick:       make(chan struct{}),
	}

	// init state: each step has a placeholder container
	for i, containerId := range containers {
		cw.containers[i] = containerInfo{
			id:          containerId,
			state:       stateWaiting,
			stateInfo:   "",
			placeholder: placeholder,
			image:       placeholder,
			exitCode:    0,
		}
	}

	return cw
}

func (w *testContainerWatcher) Name() string { return "Test" }

func (w *testContainerWatcher) Watch(ctx context.Context, containers chan<- []containerInfo) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-w.finish:
			return nil
		case update := <-w.event:
			cc := make([]containerInfo, len(w.containers))
			for i, c := range w.containers {
				if update.id == c.id {
					cc[i] = update
				} else {
					cc[i] = c
				}
			}
			containers <- cc
		}
	}
}

func (w *testContainerWatcher) PeriodicCheck(ctx context.Context, containers chan<- []containerInfo, stop <-chan struct{}) error {
	return nil
}

func TestPodWatcher(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)
	logrus.SetOutput(os.Stdout)

	type operation int

	const (
		opContAdd operation = iota
		opContSetStatePlaceholder
		opContSetState
		opWaitContainer
		opWaitPodTerm
		opCtxCancel
		opPodTerm
	)

	type step struct {
		op          operation
		containerId string
		state       containerState
		exitCode    int
		expected    error
	}

	tests := []struct {
		name       string
		containers []string
		steps      []step
	}{
		{
			name:       "one container, wait for running state",
			containers: []string{"A"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opWaitContainer, containerId: "A", state: stateRunning, expected: nil},
				{op: opContSetState, containerId: "A", state: stateRunning},
			},
		},
		{
			name:       "two containers, wait for terminated state of the second",
			containers: []string{"A", "B"},
			steps: []step{
				{op: opContAdd, containerId: "B"},
				{op: opWaitContainer, containerId: "B", state: stateTerminated, expected: nil},
				{op: opContSetState, containerId: "B", state: stateTerminated},
			},
		},
		{
			name:       "three containers, wait for running, but get terminated",
			containers: []string{"A", "B", "C"},
			steps: []step{
				{op: opContAdd, containerId: "B"},
				{op: opWaitContainer, containerId: "B", state: stateRunning, expected: nil},
				{op: opContSetState, containerId: "B", state: stateTerminated},
			},
		},
		{
			name:       "wait running, wait terminated",
			containers: []string{"A"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opWaitContainer, containerId: "A", state: stateRunning, expected: nil},
				{op: opContSetState, containerId: "A", state: stateRunning},
				{op: opWaitContainer, containerId: "A", state: stateTerminated, expected: nil},
				{op: opContSetState, containerId: "A", state: stateTerminated},
			},
		},
		{
			name:       "wait running, already running",
			containers: []string{"A"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opContSetState, containerId: "A", state: stateRunning},
				{op: opWaitContainer, containerId: "A", state: stateRunning, expected: nil},
			},
		},
		{
			name:       "wait terminated, already terminated",
			containers: []string{"A"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opContSetState, containerId: "A", state: stateTerminated, exitCode: 2},
				{op: opWaitContainer, containerId: "A", state: stateTerminated, exitCode: 2, expected: nil},
			},
		},
		{
			name:       "wait running, ignore placeholder",
			containers: []string{"A"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opWaitContainer, containerId: "A", state: stateRunning, expected: context.Canceled},
				{op: opContSetStatePlaceholder, containerId: "A", state: stateRunning},
				{op: opCtxCancel},
			},
		},
		{
			name:       "wait terminated, exit code = 1",
			containers: []string{"A"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opWaitContainer, containerId: "A", state: stateTerminated, exitCode: 1, expected: nil},
				{op: opContSetState, containerId: "A", state: stateTerminated, exitCode: 1},
			},
		},
		{
			name:       "wait running, exit code = 1",
			containers: []string{"A"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opWaitContainer, containerId: "A", state: stateRunning, exitCode: 1, expected: nil},
				{op: opContSetState, containerId: "A", state: stateTerminated, exitCode: 1},
			},
		},
		{
			name:       "unknown container",
			containers: []string{"A", "B"},
			steps: []step{
				{op: opWaitContainer, containerId: "A", state: stateRunning, expected: UnknownContainerError{}},
			},
		},
		{
			name:       "unknown image",
			containers: []string{"A"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opWaitContainer, containerId: "A", state: stateRunning, expected: FailedContainerError{}},
				{op: opContSetStatePlaceholder, containerId: "A", state: stateTerminated, exitCode: 2},
			},
		},
		{
			name:       "wait running, cancel context",
			containers: []string{"A"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opWaitContainer, containerId: "A", state: stateRunning, expected: context.Canceled},
				{op: opCtxCancel},
			},
		},
		{
			name:       "wait running, finish pod",
			containers: []string{"A"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opWaitContainer, containerId: "A", state: stateRunning, expected: PodTerminatedError{}},
				{op: opPodTerm},
			},
		},
		{
			name:       "finish pod, wait running",
			containers: []string{"A"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opPodTerm},
				{op: opWaitContainer, containerId: "A", state: stateRunning, expected: PodTerminatedError{}},
			},
		},
		{
			name:       "wait finish",
			containers: []string{"A"},
			steps: []step{
				{op: opWaitPodTerm, expected: nil},
				{op: opPodTerm},
			},
		},
		{
			name:       "wait finish, cancel context",
			containers: []string{"A"},
			steps: []step{
				{op: opWaitPodTerm, expected: context.Canceled},
				{op: opCtxCancel},
			},
		},
		{
			name:       "several wait clients",
			containers: []string{"A", "B", "C"},
			steps: []step{
				{op: opContAdd, containerId: "A"},
				{op: opWaitContainer, containerId: "A", state: stateTerminated, expected: nil},
				{op: opContAdd, containerId: "B"},
				{op: opWaitContainer, containerId: "A", state: stateTerminated, expected: nil},
				{op: opWaitContainer, containerId: "B", state: stateRunning, expected: PodTerminatedError{}},
				{op: opContAdd, containerId: "C"},
				{op: opWaitContainer, containerId: "C", state: stateTerminated, expected: PodTerminatedError{}},
				{op: opContSetState, containerId: "A", state: stateTerminated},
				{op: opPodTerm},
			},
		},
	}

	for _, test := range tests {
		logrus.Infof("--- test: %s ---", test.name)

		func() {
			ctx, cancelFunc := context.WithCancel(context.Background())
			defer cancelFunc()

			cw := makeContainerWatcher(test.containers)

			pw := &PodWatcher{}
			pw.Start(ctx, cw)

			wg := &sync.WaitGroup{}

			for stepIdx, s := range test.steps {
				if stepIdx > 0 {
					// to make sure those go routines below are able to reach the wait function inside
					time.Sleep(10 * time.Millisecond)
				}

				switch s.op {
				case opContAdd:
					pw.AddContainer(s.containerId, placeholder)

				case opContSetStatePlaceholder:
					cw.event <- containerInfo{
						id:          s.containerId,
						state:       s.state,
						stateInfo:   "",
						placeholder: placeholder,
						image:       placeholder,
						exitCode:    int32(s.exitCode),
					}

				case opContSetState:
					cw.event <- containerInfo{
						id:          s.containerId,
						state:       s.state,
						stateInfo:   "",
						placeholder: placeholder,
						image:       s.containerId,
						exitCode:    int32(s.exitCode),
					}

				case opWaitContainer:
					if s.state == stateRunning {
						wg.Add(1)
						go func(testName, containerId string, stepIdx int, expected error) {
							defer wg.Done()
							err := pw.WaitContainerStart(containerId)
							if err != nil && expected == nil {
								t.Errorf("test %q, step=%d failed: expected no error but got %v", testName, stepIdx, err)
							} else if expected != nil && (err == nil || reflect.TypeOf(err) != reflect.TypeOf(expected)) {
								t.Errorf("test %q, step=%d failed: expected error %v but got %v", testName, stepIdx, expected, err)
							}
						}(test.name, s.containerId, stepIdx, s.expected)
					} else if s.state == stateTerminated {
						wg.Add(1)
						go func(testName, containerId string, stepIdx int, expected error, expectedExitCode int) {
							defer wg.Done()
							exitCode, err := pw.WaitContainerTerminated(containerId)
							if err != nil && expected == nil {
								t.Errorf("test %q, step=%d failed: expected no error but got %v", testName, stepIdx, err)
							} else if expected != nil && (err == nil || reflect.TypeOf(err) != reflect.TypeOf(expected)) {
								t.Errorf("test %q, step=%d failed: expected error %v but got %v", testName, stepIdx, expected, err)
							}
							if exitCode != expectedExitCode {
								t.Errorf("test %q, step=%d failed: expected exit code %d, but got %v", testName, stepIdx, expectedExitCode, exitCode)
							}
						}(test.name, s.containerId, stepIdx, s.expected, s.exitCode)
					}

				case opWaitPodTerm:
					wg.Add(1)
					go func(testName string, stepIdx int, expected error) {
						defer wg.Done()
						err := pw.WaitPodDeleted()
						if err != nil && expected == nil {
							t.Errorf("test %q, step=%d failed: expected no error but got %v", testName, stepIdx, err)
						} else if expected != nil && (err == nil || reflect.TypeOf(err) != reflect.TypeOf(expected)) {
							t.Errorf("test %q, step=%d failed: expected error %v but got %v", testName, stepIdx, expected, err)
						}
					}(test.name, stepIdx, s.expected)

				case opCtxCancel:
					cancelFunc()

				case opPodTerm:
					close(cw.finish)
				}
			}

			wg.Wait()
		}()
	}
}
