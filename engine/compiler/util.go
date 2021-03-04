// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"path/filepath"
	"strings"

	"github.com/drone-runners/drone-runner-kube/engine"
	"github.com/drone-runners/drone-runner-kube/engine/resource"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/manifest"
)

// helper function returns true if the step is configured to
// always run regardless of status.
func isRunAlways(step *resource.Step) bool {
	if len(step.When.Status.Include) == 0 &&
		len(step.When.Status.Exclude) == 0 {
		return false
	}
	return step.When.Status.Match(drone.StatusFailing) &&
		step.When.Status.Match(drone.StatusPassing)
}

// helper function returns true if the step is configured to
// only run on failure.
func isRunOnFailure(step *resource.Step) bool {
	if len(step.When.Status.Include) == 0 &&
		len(step.When.Status.Exclude) == 0 {
		return false
	}
	return step.When.Status.Match(drone.StatusFailing)
}

// helper function returns true if the pipeline specification
// manually defines an execution graph.
func isGraph(spec *engine.Spec) bool {
	for _, step := range spec.Steps {
		if len(step.DependsOn) > 0 {
			return true
		}
	}
	return false
}

// helper function creates the dependency graph for serial
// pipeline execution.
func configureSerial(spec *engine.Spec) {
	var prev *engine.Step
	for _, step := range spec.Steps {
		if prev != nil {
			step.DependsOn = []string{prev.Name}
		}
		prev = step
	}
}

// helper function converts the environment variables to a map,
// returning only inline environment variables not derived from
// a secret.
func convertStaticEnv(src map[string]*manifest.Variable) map[string]string {
	dst := map[string]string{}
	for k, v := range src {
		if v == nil {
			continue
		}
		if strings.TrimSpace(v.Secret) == "" {
			dst[k] = v.Value
		}
	}
	return dst
}

// helper function converts the environment variables to a map,
// returning only inline environment variables not derived from
// a secret.
func convertSecretEnv(src map[string]*manifest.Variable) []*engine.SecretVar {
	dst := []*engine.SecretVar{}
	for k, v := range src {
		if v == nil {
			continue
		}
		if strings.TrimSpace(v.Secret) != "" {
			dst = append(dst, &engine.SecretVar{
				Name: v.Secret,
				Env:  k,
			})
		}
	}
	return dst
}

// helper function converts the resource limits structure from the
// yaml package to the resource limit structure used by the engine.
func convertResources(src resource.Resources) engine.Resources {
	return engine.Resources{
		Limits: engine.ResourceObject{
			CPU:    src.Limits.CPU,
			Memory: int64(src.Limits.Memory),
		},
	}
}

// helper function modifies the pipeline dependency graph to
// account for the clone step.
func configureCloneDeps(spec *engine.Spec) {
	for _, step := range spec.Steps {
		if step.Name == "clone" {
			continue
		}
		if len(step.DependsOn) == 0 {
			step.DependsOn = []string{"clone"}
		}
	}
}

// helper function modifies the pipeline dependency graph to
// account for a disabled clone step.
func removeCloneDeps(spec *engine.Spec) {
	for _, step := range spec.Steps {
		if step.Name == "clone" {
			return
		}
	}
	for _, step := range spec.Steps {
		if len(step.DependsOn) == 1 &&
			step.DependsOn[0] == "clone" {
			step.DependsOn = []string{}
		}
	}
}

// helper function modifies the pipeline dependency graph to
// account for the clone step.
func convertPullPolicy(s string) engine.PullPolicy {
	switch strings.ToLower(s) {
	case "always":
		return engine.PullAlways
	case "if-not-exists":
		return engine.PullIfNotExists
	case "never":
		return engine.PullNever
	default:
		return engine.PullDefault
	}
}

// helper function returns true if mounting the volume
// is restricted for un-trusted containers.
func isRestrictedVolume(path string) bool {
	path, _ = filepath.Abs(path)
	path = strings.ToLower(path)
	switch {
	case path == "/":
	case path == "/var":
	case path == "/etc":
	case strings.HasPrefix(path, "/var/run") || strings.HasSuffix(path, "/var/run") || strings.Contains(path, "/var/run/"):
	case strings.HasPrefix(path, "/proc") || strings.HasSuffix(path, "/proc") || strings.Contains(path, "/proc/"):
	case strings.HasPrefix(path, "/mount") || strings.HasSuffix(path, "/mount") || strings.Contains(path, "/mount/"):
	case strings.HasPrefix(path, "/bin") || strings.HasSuffix(path, "/bin") || strings.Contains(path, "/bin/"):
	case strings.HasPrefix(path, "/usr/local/bin") || strings.HasSuffix(path, "/usr/local/bin") || strings.Contains(path, "/usr/local/bin/"):
	case strings.HasPrefix(path, "/usr/local/sbin") || strings.HasSuffix(path, "/usr/local/sbin") || strings.Contains(path, "/usr/local/sbin/"):
	case strings.HasPrefix(path, "/usr/bin") || strings.HasSuffix(path, "/usr/bin") || strings.Contains(path, "/usr/bin/"):
	case strings.HasPrefix(path, "/mnt") || strings.HasSuffix(path, "/mnt") || strings.Contains(path, "/mnt/"):
	case strings.HasPrefix(path, "/media") || strings.HasSuffix(path, "/media") || strings.Contains(path, "/media/"):
	case strings.HasPrefix(path, "/sys") || strings.HasSuffix(path, "/sys") || strings.Contains(path, "/sys/"):
	case strings.HasPrefix(path, "/dev") || strings.HasSuffix(path, "/dev") || strings.Contains(path, "/dev/"):
	case strings.HasPrefix(path, "/etc/docker") || strings.HasSuffix(path, "/etc/docker") || strings.Contains(path, "/etc/docker/"):
	default:
		return false
	}
	return true
}

// helper function returns true if the environment variable
// is restricted for internal-use only.
func isRestrictedVariable(env map[string]*manifest.Variable) bool {
	for _, name := range restrictedVars {
		if _, ok := env[name]; ok {
			return true
		}
	}
	return false
}

// Upper value for cpu and memory requests for step containers.
// It is same as stage resource request if set in pipeline. Otherwise, it defaults to
// runner environment variable values.
func getStepUpperRequestVal(stageResources resource.Resources,
	defaultRequests ResourceObject) ResourceObject {
	r := ResourceObject{
		CPU:    stageResources.Requests.CPU,
		Memory: int64(stageResources.Requests.Memory),
	}
	if r.CPU == 0 {
		r.CPU = defaultRequests.CPU
	}
	if r.Memory == 0 {
		r.Memory = defaultRequests.Memory
	}
	return r
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// list of restricted variables
var restrictedVars = []string{
	"XDG_RUNTIME_DIR",
	"DOCKER_OPTS",
	"DOCKER_HOST",
	"PATH",
	"HOME",
}
