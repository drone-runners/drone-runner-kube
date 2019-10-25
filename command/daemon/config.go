// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package daemon

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config stores the system configuration.
type Config struct {
	Debug bool `envconfig:"DRONE_DEBUG"`
	Trace bool `envconfig:"DRONE_TRACE"`

	Client struct {
		Address    string `ignored:"true"`
		Proto      string `envconfig:"DRONE_RPC_PROTO"  default:"http"`
		Host       string `envconfig:"DRONE_RPC_HOST"   required:"true"`
		Secret     string `envconfig:"DRONE_RPC_SECRET" required:"true"`
		SkipVerify bool   `envconfig:"DRONE_RPC_SKIP_VERIFY"`
		Dump       bool   `envconfig:"DRONE_RPC_DUMP_HTTP"`
		DumpBody   bool   `envconfig:"DRONE_RPC_DUMP_HTTP_BODY"`
	}

	Dashboard struct {
		Disabled bool   `envconfig:"DRONE_UI_DISABLE"`
		Username string `envconfig:"DRONE_UI_USERNAME"`
		Password string `envconfig:"DRONE_UI_PASSWORD"`
		Realm    string `envconfig:"DRONE_UI_REALM" default:"MyRealm"`
	}

	Server struct {
		Proto string `envconfig:"DRONE_SERVER_PROTO"`
		Host  string `envconfig:"DRONE_SERVER_HOST"`
		Port  string `envconfig:"DRONE_SERVER_PORT" default:":3000"`
		Acme  bool   `envconfig:"DRONE_SERVER_ACME"`
	}

	Keypair struct {
		Public  string `envconfig:"DRONE_PUBLIC_KEY_FILE"`
		Private string `envconfig:"DRONE_PRIVATE_KEY_FILE"`
	}

	Runner struct {
		Name       string            `envconfig:"DRONE_RUNNER_NAME"`
		Capacity   int               `envconfig:"DRONE_RUNNER_CAPACITY" default:"2"`
		Procs      int64             `envconfig:"DRONE_RUNNER_MAX_PROCS"`
		Environ    map[string]string `envconfig:"DRONE_RUNNER_ENVIRON"`
		EnvFile    string            `envconfig:"DRONE_RUNNER_ENV_FILE"`
		Secrets    map[string]string `envconfig:"DRONE_RUNNER_SECRETS"`
		Labels     map[string]string `envconfig:"DRONE_RUNNER_LABELS"`
		Volumes    map[string]string `envconfig:"DRONE_RUNNER_VOLUMES"`
		Devices    []string          `envconfig:"DRONE_RUNNER_DEVICES"`
		Networks   []string          `envconfig:"DRONE_RUNNER_NETWORKS"`
		Privileged []string          `envconfig:"DRONE_RUNNER_PRIVILEGED_IMAGES"`
	}

	Platform struct {
		OS      string `envconfig:"DRONE_PLATFORM_OS"    default:"linux"`
		Arch    string `envconfig:"DRONE_PLATFORM_ARCH"  default:"amd64"`
		Kernel  string `envconfig:"DRONE_PLATFORM_KERNEL"`
		Variant string `envconfig:"DRONE_PLATFORM_VARIANT"`
	}

	Limit struct {
		Repos   []string `envconfig:"DRONE_LIMIT_REPOS"`
		Events  []string `envconfig:"DRONE_LIMIT_EVENTS"`
		Trusted bool     `envconfig:"DRONE_LIMIT_TRUSTED"`
	}

	Secret struct {
		Endpoint   string `envconfig:"DRONE_SECRET_PLUGIN_ENDPOINT"`
		Token      string `envconfig:"DRONE_SECRET_PLUGIN_TOKEN"`
		SkipVerify bool   `envconfig:"DRONE_SECRET_PLUGIN_SKIP_VERIFY"`
	}

	Registry struct {
		Endpoint   string `envconfig:"DRONE_REGISTRY_PLUGIN_ENDPOINT"`
		Token      string `envconfig:"DRONE_REGISTRY_PLUGIN_SECRET"`
		SkipVerify bool   `envconfig:"DRONE_REGISTRY_PLUGIN_SKIP_VERIFY"`
	}

	Docker struct {
		Config string `envconfig:"DRONE_DOCKER_CONFIG"`
	}
}

func fromEnviron() (Config, error) {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		return config, err
	}
	if config.Runner.Name == "" {
		config.Runner.Name, _ = os.Hostname()
	}
	if config.Dashboard.Password == "" {
		config.Dashboard.Disabled = true
	}
	config.Client.Address = fmt.Sprintf(
		"%s://%s",
		config.Client.Proto,
		config.Client.Host,
	)

	// environment variables can be sourced from a separate
	// file. These variables are loaded and appended to the
	// environment list.
	if file := config.Runner.EnvFile; file != "" {
		envs, err := godotenv.Read(file)
		if err != nil {
			return config, err
		}
		for k, v := range envs {
			config.Runner.Environ[k] = v
		}
	}

	return config, nil
}
