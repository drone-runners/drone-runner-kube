// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package daemon

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/drone-runners/drone-runner-kube/engine/policy"

	"github.com/buildkite/yaml"
	"github.com/docker/go-units"
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

	Runner struct {
		Name       string            `envconfig:"DRONE_RUNNER_NAME"`
		Capacity   int               `envconfig:"DRONE_RUNNER_CAPACITY" default:"100"`
		Procs      int64             `envconfig:"DRONE_RUNNER_MAX_PROCS"`
		Environ    map[string]string `envconfig:"DRONE_RUNNER_ENVIRON"`
		EnvFile    string            `envconfig:"DRONE_RUNNER_ENV_FILE"`
		Secrets    map[string]string `envconfig:"DRONE_RUNNER_SECRETS"`
		Labels     map[string]string `envconfig:"DRONE_RUNNER_LABELS"`
		Volumes    map[string]string `envconfig:"DRONE_RUNNER_VOLUMES"`
		Privileged []string          `envconfig:"DRONE_RUNNER_PRIVILEGED_IMAGES"`
		Config     string            `envconfig:"DRONE_KUBERNETES_CONFIG"`
	}

	Limit struct {
		Repos   []string `envconfig:"DRONE_LIMIT_REPOS"`
		Events  []string `envconfig:"DRONE_LIMIT_EVENTS"`
		Trusted bool     `envconfig:"DRONE_LIMIT_TRUSTED"`
	}

	Resources struct {
		LimitCPU      int64     `envconfig:"DRONE_RESOURCE_LIMIT_CPU"`
		LimitMemory   BytesSize `envconfig:"DRONE_RESOURCE_LIMIT_MEMORY"`
		RequestCPU    int64     `envconfig:"DRONE_RESOURCE_REQUEST_CPU" default:"100"`
		RequestMemory BytesSize `envconfig:"DRONE_RESOURCE_REQUEST_MEMORY" default:"104857600"` // default 100MB
		// Defaults for min memory & cpu requests are set to ensure that default limitrange values are not used.
		// Default for MinRequestMemory is set to 4MB. Anything below 4MB fails with
		// "Error: Error response from daemon: Minimum memory limit allowed is 4MB" on gke
		MinRequestMemory BytesSize `envconfig:"DRONE_RESOURCE_MIN_REQUEST_MEMORY" default:"4194304"` // default 4MB
		MinRequestCPU    int64     `envconfig:"DRONE_RESOURCE_MIN_REQUEST_CPU" default:"1"`
	}

	Policy struct {
		Path   string           `envconfig:"DRONE_POLICY_FILE"`
		Parsed []*policy.Policy `envconfig:"-"`
	}

	Secret struct {
		Endpoint   string `envconfig:"DRONE_SECRET_PLUGIN_ENDPOINT"`
		Token      string `envconfig:"DRONE_SECRET_PLUGIN_TOKEN"`
		SkipVerify bool   `envconfig:"DRONE_SECRET_PLUGIN_SKIP_VERIFY"`
	}

	Netrc struct {
		CloneOnly bool `envconfig:"DRONE_NETRC_CLONE_ONLY"`
	}

	Registry struct {
		Endpoint   string `envconfig:"DRONE_REGISTRY_PLUGIN_ENDPOINT"`
		Token      string `envconfig:"DRONE_REGISTRY_PLUGIN_TOKEN"`
		SkipVerify bool   `envconfig:"DRONE_REGISTRY_PLUGIN_SKIP_VERIFY"`
	}

	Environ struct {
		Endpoint   string `envconfig:"DRONE_ENV_PLUGIN_ENDPOINT"`
		Token      string `envconfig:"DRONE_ENV_PLUGIN_TOKEN"`
		SkipVerify bool   `envconfig:"DRONE_ENV_PLUGIN_SKIP_VERIFY"`
	}

	Docker struct {
		Config string `envconfig:"DRONE_DOCKER_CONFIG"`
	}

	Images struct {
		Clone       string `envconfig:"DRONE_IMAGE_CLONE"`
		Placeholder string `envconfig:"DRONE_IMAGE_PLACEHOLDER"`
	}

	ServiceAccount struct {
		Default string `envconfig:"DRONE_SERVICE_ACCOUNT_DEFAULT"`
	}

	NodeSelector struct {
		Default map[string]string `envconfig:"DRONE_NODE_SELECTOR_DEFAULT"`
	}

	Annotations struct {
		Default map[string]string `envconfig:"DRONE_ANNOTATIONS_DEFAULT"`
	}

	Labels struct {
		Default map[string]string `envconfig:"DRONE_LABELS_DEFAULT"`
	}

	Namespace struct {
		Rules     map[string][]string `envconfig:"-"`
		RulesMap  map[string]string   `envconfig:"DRONE_NAMESPACE_RULES"`
		RulesFile string              `envconfig:"DRONE_NAMESPACE_RULES_FILE"`
		Default   string              `envconfig:"DRONE_NAMESPACE_DEFAULT" default:"default"`
	}

	Tmate struct {
		Enabled bool   `envconfig:"DRONE_TMATE_ENABLED" default:"true"`
		Image   string `envconfig:"DRONE_TMATE_IMAGE"   default:"drone/drone-runner-docker:1"`
		Server  string `envconfig:"DRONE_TMATE_HOST"`
		Port    string `envconfig:"DRONE_TMATE_PORT"`
		RSA     string `envconfig:"DRONE_TMATE_FINGERPRINT_RSA"`
		ED25519 string `envconfig:"DRONE_TMATE_FINGERPRINT_ED25519"`
	}
}

// legacy environment variables. the key is the legacy
// variable name, and the value is the new variable name.
var legacy = map[string]string{
	"DRONE_REGISTRY_ENDPOINT":      "DRONE_REGISTRY_PLUGIN_ENDPOINT",
	"DRONE_REGISTRY_SECRET":        "DRONE_REGISTRY_PLUGIN_TOKEN",
	"DRONE_REGISTRY_PLUGIN_SECRET": "DRONE_REGISTRY_PLUGIN_TOKEN",
}

func fromEnviron() (Config, error) {
	// loop through legacy environment variable and, if set
	// rewrite to the new variable name.
	for k, v := range legacy {
		if s, ok := os.LookupEnv(k); ok {
			os.Setenv(v, s)
		}
	}

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

	// namespace usage rules can be sourced from a separate
	// file. These variables are loaded and appended to the map.
	config.Namespace.Rules = map[string][]string{}
	if file := config.Namespace.RulesFile; file != "" {
		out, err := ioutil.ReadFile(file)
		if err != nil {
			return config, err
		}
		err = yaml.Unmarshal(out, &config.Namespace.Rules)
		if err != nil {
			return config, err
		}
	}
	// namespace usage rules can be sourced from a separate
	// file. These variables are loaded and appended to the map.
	for k, v := range config.Namespace.RulesMap {
		config.Namespace.Rules[k] = []string{v}
	}

	// environment variables can be sourced from a separate
	// file. These variables are loaded and appended to the
	// environment list.
	if file := config.Runner.EnvFile; file != "" {
		envs, err := godotenv.Read(file)
		if err != nil {
			return config, err
		}
		if config.Runner.Environ == nil {
			config.Runner.Environ = map[string]string{}
		}
		for k, v := range envs {
			config.Runner.Environ[k] = v
		}
	}

	// parse the policy file if defined
	if file := config.Policy.Path; file != "" {
		config.Policy.Parsed, err = policy.ParseFile(file)
		if err != nil {
			return config, err
		}
	}

	return config, nil
}

type BytesSize int64

func (b *BytesSize) Decode(value string) error {
	intType, err := units.RAMInBytes(value)
	if err != nil {
		return err
	}
	*b = BytesSize(intType)
	return nil
}
