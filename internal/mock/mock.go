// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package mock

//go:generate mockgen -package=mock -destination=mock_engine_gen.go github.com/drone-runners/drone-runner-kube/engine Engine
//go:generate mockgen -package=mock -destination=mock_execer_gen.go github.com/drone-runners/drone-runner-kube/runtime Execer
