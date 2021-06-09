// Copyright 2021 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package podwatcher

import (
	"context"
	"testing"
	"time"
)

type testContainerWatcher struct {
	containers []containerInfo
	finish     chan struct{}
	event      chan containerInfo
	tick       chan struct{}
}

func makeContainerWatcher() *testContainerWatcher {
	return &testContainerWatcher{
		containers: make([]containerInfo, 0, 4),
		finish:     make(chan struct{}),
		event:      make(chan containerInfo),
		tick:       make(chan struct{}),
	}
}

func (w *testContainerWatcher) Name() string { return "Test" }

func (w *testContainerWatcher) Watch(ctx context.Context, containers chan<- []containerInfo) error {
	defer close(containers)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-w.finish:
			return nil
		case c := <-w.event:
			var found bool

			for i := 0; i < len(w.containers); i++ {
				if w.containers[i].id == c.id {
					w.containers[i] = c
					found = true
					break
				}
			}

			if !found {
				w.containers = append(w.containers, c)
			}

			containers <- w.containers
		}
	}
}

func (w *testContainerWatcher) PeriodicCheck(ctx context.Context, containers chan<- []containerInfo, stop <-chan struct{}) error {
	return nil
}

func TestPodWatcher(t *testing.T) {
	//logrus.SetLevel(logrus.TraceLevel)
	//logrus.SetOutput(os.Stdout)

	type step struct {
		id         string
		state      containerState
		skipUpdate bool
		exitCode   int
		waitFor    containerState
		expected   error
	}

	// agent is a hack to preform something to disrupt the normal test flow
	type agent struct {
		atStep int
		action string
	}

	tests := []struct {
		name       string
		containers []string
		steps      []step
		agent      *agent // performed before specific step
		agentPod   string // performed before termination
		expected   error
	}{
		{
			name:       "one container, wait for running state",
			containers: []string{"A"},
			steps: []step{
				{id: "A", state: stateRunning, waitFor: stateRunning, expected: nil},
			},
			agentPod: "finish",
		},
		{
			name:       "two containers, wait for terminated state of the second",
			containers: []string{"A", "B"},
			steps: []step{
				{id: "B", state: stateTerminated, waitFor: stateTerminated, expected: nil},
			},
			agentPod: "finish",
		},
		{
			name:       "three containers, wait running, get terminated",
			containers: []string{"A", "B", "C"},
			steps: []step{
				{id: "C", state: stateTerminated, waitFor: stateRunning, expected: nil},
			},
			agentPod: "finish",
		},
		{
			name:       "wait running, wait terminated",
			containers: []string{"A"},
			steps: []step{
				{id: "A", state: stateRunning, waitFor: stateRunning, expected: nil},
				{id: "A", state: stateTerminated, waitFor: stateTerminated, expected: nil},
			},
			agentPod: "finish",
		},
		{
			name:       "wait terminated, exit code = 1",
			containers: []string{"A"},
			steps: []step{
				{id: "A", state: stateTerminated, waitFor: stateTerminated, exitCode: 1, expected: nil},
			},
			agentPod: "finish",
		},
		{
			name:       "wait running, cancel context",
			containers: []string{"A"},
			steps: []step{
				{id: "A", state: stateRunning, waitFor: stateRunning, expected: nil},
				{id: "A", state: stateTerminated, waitFor: stateTerminated, expected: context.Canceled},
			},
			agent:    &agent{atStep: 1, action: "ctx"},
			agentPod: "", // an agent canceled context at step 1 (the line above), so no need to do anything
			expected: context.Canceled,
		},
		{
			name:       "wait running, finish watcher",
			containers: []string{"A"},
			steps: []step{
				{id: "A", state: stateRunning, waitFor: stateRunning, expected: nil},
				{id: "A", state: stateTerminated, waitFor: stateTerminated, expected: ErrPodTerminated},
			},
			agent:    &agent{atStep: 1, action: "finish"},
			agentPod: "", // an agent finished watcher at step 1 (the line above), so no need to do anything
			expected: ErrPodTerminated,
		},
		{
			name:       "wait terminate, but already terminated",
			containers: []string{"A"},
			steps: []step{
				{id: "A", state: stateRunning, waitFor: stateRunning, expected: nil},
				{id: "A", skipUpdate: true, waitFor: stateTerminated, expected: nil},
			},
			agent:    &agent{atStep: 1, action: "terminate0"},
			agentPod: "finish",
		},
	}

	for _, test := range tests {
		//logrus.Infoln("-------------")

		func() {
			pw := &PodWatcher{}

			ctx, cancelFunc := context.WithCancel(context.Background())
			defer cancelFunc()

			cw := makeContainerWatcher()

			pw.Start(ctx, cw)

			for stepIdx, s := range test.steps {
				var err error
				var exitCode int

				pw.AddContainer(s.id, "")

				if test.agent != nil && test.agent.atStep == stepIdx {
					switch test.agent.action {
					case "ctx":
						cancelFunc()
					case "finish":
						close(cw.finish)
					case "terminate0":
						cw.event <- containerInfo{
							id:          test.containers[0],
							state:       stateTerminated,
							stateInfo:   "",
							placeholder: "",
							image:       test.containers[0],
							exitCode:    0,
						}
						time.Sleep(10 * time.Millisecond)
					}
				}

				if !s.skipUpdate {
					go func() {
						time.Sleep(10 * time.Millisecond)
						cw.event <- containerInfo{
							id:          s.id,
							state:       s.state,
							stateInfo:   "",
							placeholder: "",
							image:       s.id,
							exitCode:    int32(s.exitCode),
						}
					}()
				}

				if s.waitFor == stateRunning {
					err = pw.WaitContainerStart(s.id)
				} else {
					exitCode, err = pw.WaitContainerTerminated(s.id)
				}

				if err != nil && s.expected == nil {
					t.Errorf("test %q, step=%d failed: expected no error but got %v", test.name, stepIdx, err)
				} else if s.expected != nil && err != s.expected {
					t.Errorf("test %q, step=%d failed: expected error but got %v", test.name, stepIdx, err)
				}

				if exitCode != s.exitCode {
					t.Errorf("test %q, step=%d failed: expected exit code %d, but got %v", test.name, stepIdx, s.exitCode, exitCode)
				}
			}

			switch test.agentPod {
			case "ctx":
				cancelFunc()
				time.Sleep(10 * time.Millisecond)
			case "finish":
				go func() {
					time.Sleep(10 * time.Millisecond)
					close(cw.finish)
				}()
			}

			err := pw.WaitPodDeleted()
			if err != nil && test.expected == nil {
				t.Errorf("test %q failed: expected no error but got %v", test.name, err)
			} else if test.expected != nil && err != test.expected {
				t.Errorf("test %q failed: expected error but got %v", test.name, err)
			}
		}()
	}
}
