// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package entrypoint

import "gopkg.in/alecthomas/kingpin.v2"

// Register registers the entrypoint command.
func Register(app *kingpin.Application) {
	registerEntrypoint(app)
	registerCopy(app)
}
