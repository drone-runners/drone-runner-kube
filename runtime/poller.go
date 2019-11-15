// Code generated automatically. DO NOT EDIT.

// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package runtime

import (
	"context"
	"sync"

	"github.com/drone/runner-go/client"
	"github.com/drone/runner-go/logger"
)

var noContext = context.Background()

// Poller polls the server for pending stages and dispatches
// for execution by the Runner.
type Poller struct {
	Client client.Client
	Filter *client.Filter
	Runner *Runner
}

// Poll opens N connections to the server to poll for pending
// stages for execution. Pending stages are dispatched to a
// Runner for execution.
func (p *Poller) Poll(ctx context.Context, n int) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			for {
				select {
				case <-ctx.Done():
					wg.Done()
					return
				default:
					p.poll(ctx, i+1)
				}
			}
		}(i)
	}

	wg.Wait()
}

// poll requests a stage for execution from the server, and then
// dispatches for execution.
func (p *Poller) poll(ctx context.Context, thread int) error {
	log := logger.FromContext(ctx).WithField("thread", thread)
	log.WithField("thread", thread).Debug("request stage from remote server")

	// request a new build stage for execution from the central
	// build server.
	stage, err := p.Client.Request(ctx, p.Filter)
	if err == context.Canceled || err == context.DeadlineExceeded {
		log.WithError(err).Trace("no stage returned")
		return nil
	}
	if err != nil {
		log.WithError(err).Error("cannot request stage")
		return err
	}

	// exit if a nil or empty stage is returned from the system
	// and allow the runner to retry.
	if stage == nil || stage.ID == 0 {
		return nil
	}

	return p.Runner.Run(
		logger.WithContext(noContext, log), stage)
}
