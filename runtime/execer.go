// Code generated automatically. DO NOT EDIT.

// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package runtime

import (
	"context"
	"sync"

	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/replacer"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/logger"
	"github.com/drone/runner-go/pipeline"

	"github.com/hashicorp/go-multierror"
	"github.com/natessilva/dag"
	"golang.org/x/sync/semaphore"
)

// Execer is the execution context for executing the intermediate
// representation of a pipeline.
type Execer interface {
	Exec(context.Context, *engine.Spec, *pipeline.State) error
}

type execer struct {
	mu       sync.Mutex
	engine   engine.Engine
	reporter pipeline.Reporter
	streamer pipeline.Streamer
	sem      *semaphore.Weighted
}

// NewExecer returns a new execer used
func NewExecer(
	reporter pipeline.Reporter,
	streamer pipeline.Streamer,
	engine engine.Engine,
	procs int64,
) Execer {
	exec := &execer{
		reporter: reporter,
		streamer: streamer,
		engine:   engine,
	}
	if procs > 0 {
		// optional semaphor that limits the number of steps
		// that can execute concurrently.
		exec.sem = semaphore.NewWeighted(procs)
	}
	return exec
}

// Exec executes the intermediate representation of the pipeline
// and returns an error if execution fails.
func (e *execer) Exec(ctx context.Context, spec *engine.Spec, state *pipeline.State) error {
	defer e.engine.Destroy(noContext, spec)

	if err := e.engine.Setup(noContext, spec); err != nil {
		state.FailAll(err)
		return e.reporter.ReportStage(noContext, state)
	}

	// create a directed graph, where each vertex in the graph
	// is a pipeline step.
	var d dag.Runner
	for _, s := range spec.Steps {
		step := s
		d.AddVertex(step.Name, func() error {
			return e.exec(ctx, state, spec, step)
		})
	}

	// create the vertex edges from the values configured in the
	// depends_on attribute.
	for _, s := range spec.Steps {
		for _, dep := range s.DependsOn {
			d.AddEdge(dep, s.Name)
		}
	}

	var result error
	if err := d.Run(); err != nil {
		multierror.Append(result, err)
	}

	// once pipeline execution completes, notify the state
	// manageer that all steps are finished.
	state.FinishAll()
	if err := e.reporter.ReportStage(noContext, state); err != nil {
		multierror.Append(result, err)
	}
	return result
}

func (e *execer) exec(ctx context.Context, state *pipeline.State, spec *engine.Spec, step *engine.Step) error {
	var result error

	select {
	case <-ctx.Done():
		state.Cancel()
		return nil
	default:
	}

	log := logger.FromContext(ctx)
	log = log.WithField("step.name", step.Name)
	ctx = logger.WithContext(ctx, log)

	if e.sem != nil {
		// the semaphore limits the number of steps that can run
		// concurrently. acquire the semaphore and release when
		// the pipeline completes.
		if err := e.sem.Acquire(ctx, 1); err != nil {
			return nil
		}

		defer func() {
			// recover from a panic to ensure the semaphore is
			// released to prevent deadlock. we do not expect a
			// panic, however, we are being overly cautious.
			if r := recover(); r != nil {
				// TODO(bradrydzewsi) log the panic.
			}
			// release the semaphore
			e.sem.Release(1)
		}()
	}

	switch {
	case state.Skipped():
		return nil
	case state.Cancelled():
		return nil
	case step.RunPolicy == engine.RunNever:
		return nil
	case step.RunPolicy == engine.RunAlways:
		break
	case step.RunPolicy == engine.RunOnFailure && state.Failed() == false:
		state.Skip(step.Name)
		return e.reporter.ReportStep(noContext, state, step.Name)
	case step.RunPolicy == engine.RunOnSuccess && state.Failed():
		state.Skip(step.Name)
		return e.reporter.ReportStep(noContext, state, step.Name)
	}

	state.Start(step.Name)
	err := e.reporter.ReportStep(noContext, state, step.Name)
	if err != nil {
		return err
	}

	copy := cloneStep(step)

	// the pipeline environment variables need to be updated to
	// reflect the current state of the build and stage.
	state.Lock()
	copy.Envs = environ.Combine(
		copy.Envs,
		environ.Build(state.Build),
		environ.Stage(state.Stage),
		environ.Step(findStep(state, step.Name)),
	)
	state.Unlock()

	// writer used to stream build logs.
	wc := e.streamer.Stream(noContext, state, step.Name)
	wc = replacer.New(wc, step.Secrets)

	// if the step is configured as a daemon, it is detached
	// from the main process and executed separately.
	// todo(bradrydzewski) this code is still experimental.
	if step.Detach {
		go func() {
			e.engine.Run(ctx, spec, copy, wc)
			wc.Close()
		}()
		return nil
	}

	exited, err := e.engine.Run(ctx, spec, copy, wc)

	// close the stream. If the session is a remote session, the
	// full log buffer is uploaded to the remote server.
	if err := wc.Close(); err != nil {
		multierror.Append(result, err)
	}

	if exited != nil {
		state.Finish(step.Name, exited.ExitCode)
		err := e.reporter.ReportStep(noContext, state, step.Name)
		if err != nil {
			multierror.Append(result, err)
		}
		// if the exit code is 78 the system will skip all
		// subsequent pending steps in the pipeline.
		if exited.ExitCode == 78 {
			state.SkipAll()
		}
		return result
	}

	switch err {
	case context.Canceled, context.DeadlineExceeded:
		state.Cancel()
		return nil
	}

	// if the step failed with an internal error (as oppsed to a
	// runtime error) the step is failed.
	state.Fail(step.Name, err)
	err = e.reporter.ReportStep(noContext, state, step.Name)
	if err != nil {
		multierror.Append(result, err)
	}
	return result
}

// helper function to clone a step. The runner mutates a step to
// update the environment variables to reflect the current
// pipeline state.
func cloneStep(src *engine.Step) *engine.Step {
	dst := new(engine.Step)
	*dst = *src
	dst.Envs = environ.Combine(src.Envs)
	return dst
}

// helper function returns the named step from the state.
func findStep(state *pipeline.State, name string) *drone.Step {
	for _, step := range state.Stage.Steps {
		if step.Name == name {
			return step
		}
	}
	panic("step not found: " + name)
}
