// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package daemon

import (
	"context"
	"time"

	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/compiler"
	"github.com/drone-runners/drone-runner-kube/engine/linter"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone-runners/drone-runner-kube/internal/match"
	"github.com/drone-runners/drone-runner-kube/runtime"

	"github.com/drone/runner-go/client"
	"github.com/drone/runner-go/handler/router"
	"github.com/drone/runner-go/logger"
	loghistory "github.com/drone/runner-go/logger/history"
	"github.com/drone/runner-go/pipeline/history"
	"github.com/drone/runner-go/pipeline/remote"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"
	"github.com/drone/runner-go/server"
	"github.com/drone/signal"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/alecthomas/kingpin.v2"
)

// empty context.
var nocontext = context.Background()

type daemonCommand struct {
	envfile string
}

func (c *daemonCommand) run(*kingpin.ParseContext) error {
	// load environment variables from file.
	godotenv.Load(c.envfile)

	// load the configuration from the environment
	config, err := fromEnviron()
	if err != nil {
		return err
	}

	// setup the global logrus logger.
	setupLogger(config)

	ctx, cancel := context.WithCancel(nocontext)
	defer cancel()

	// listen for termination signals to gracefully shutdown
	// the runner daemon.
	ctx = signal.WithContextFunc(ctx, func() {
		println("received signal, terminating process")
		cancel()
	})

	cli := client.New(
		config.Client.Address,
		config.Client.Secret,
		config.Client.SkipVerify,
	)
	if config.Client.Dump {
		cli.Dumper = logger.StandardDumper(
			config.Client.DumpBody,
		)
	}
	cli.Logger = logger.Logrus(
		logrus.NewEntry(
			logrus.StandardLogger(),
		),
	)

	engine, err := engine.NewEnv()
	if err != nil {
		logrus.WithError(err).
			Fatalln("cannot load the docker engine")
	}

	remote := remote.New(cli)
	tracer := history.New(remote)
	hook := loghistory.New()
	logrus.AddHook(hook)

	poller := &runtime.Poller{
		Client: cli,
		Runner: &runtime.Runner{
			Client:   cli,
			Machine:  config.Runner.Name,
			Reporter: tracer,
			Linter:   linter.New(),
			Match: match.Func(
				config.Limit.Repos,
				config.Limit.Events,
				config.Limit.Trusted,
			),
			Compiler: &compiler.Compiler{
				Environ:    config.Runner.Environ,
				Privileged: append(config.Runner.Privileged, compiler.Privileged...),
				Networks:   config.Runner.Networks,
				Volumes:    config.Runner.Volumes,
				// Resources:  nil,
				Registry: registry.Combine(
					registry.File(
						config.Docker.Config,
					),
					registry.External(
						config.Registry.Endpoint,
						config.Registry.Token,
						config.Registry.SkipVerify,
					),
				),
				Secret: secret.Combine(
					secret.StaticVars(
						config.Runner.Secrets,
					),
					secret.External(
						config.Secret.Endpoint,
						config.Secret.Token,
						config.Secret.SkipVerify,
					),
				),
			},
			Execer: runtime.NewExecer(
				tracer,
				remote,
				engine,
				config.Runner.Procs,
			),
		},
		Filter: &client.Filter{
			Kind:    resource.Kind,
			Type:    resource.Type,
			OS:      config.Platform.OS,
			Arch:    config.Platform.Arch,
			Variant: config.Platform.Variant,
			Kernel:  config.Platform.Kernel,
			Labels:  config.Runner.Labels,
		},
	}

	var g errgroup.Group
	server := server.Server{
		Addr: config.Server.Port,
		Handler: router.New(tracer, hook, router.Config{
			Username: config.Dashboard.Username,
			Password: config.Dashboard.Password,
			Realm:    config.Dashboard.Realm,
		}),
	}

	logrus.WithField("addr", config.Server.Port).
		Infoln("starting the server")

	g.Go(func() error {
		return server.ListenAndServe(ctx)
	})

	// Ping the server and block until a successful connection
	// to the server has been established.
	for {
		err := cli.Ping(ctx, config.Runner.Name)
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if ctx.Err() != nil {
			break
		}
		if err != nil {
			logrus.WithError(err).
				Errorln("cannot ping the remote server")
			time.Sleep(time.Second)
		} else {
			logrus.Infoln("successfully pinged the remote server")
			break
		}
	}

	g.Go(func() error {
		logrus.WithField("capacity", config.Runner.Capacity).
			WithField("endpoint", config.Client.Address).
			WithField("kind", resource.Kind).
			WithField("type", resource.Type).
			Infoln("polling the remote server")

		poller.Poll(ctx, config.Runner.Capacity)
		return nil
	})

	err = g.Wait()
	if err != nil {
		logrus.WithError(err).
			Errorln("shutting down the server")
	}
	return err
}

// helper function configures the global logger from
// the loaded configuration.
func setupLogger(config Config) {
	logger.Default = logger.Logrus(
		logrus.NewEntry(
			logrus.StandardLogger(),
		),
	)
	if config.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if config.Trace {
		logrus.SetLevel(logrus.TraceLevel)
	}
}

// Register the daemon command.
func Register(app *kingpin.Application) {
	c := new(daemonCommand)

	cmd := app.Command("daemon", "starts the runner daemon").
		Default().
		Action(c.run)

	cmd.Arg("envfile", "load the environment variable file").
		Default("").
		StringVar(&c.envfile)
}
