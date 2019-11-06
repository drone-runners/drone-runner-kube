---
date: 2000-01-01T00:00:00+00:00
title: Specification
author: bradrydzewski
weight: 4
toc: false
type: spec
hide_children: true

description: |
  Yaml specification document.
---

This document introduces the data structures that represent the _docker pipeline_. The Docker pipeline is a continuous integration pipeline that executes pipeline steps inside isolated Docker containers. 

# The `Resource` interface

The [`Resource`](#the-resource-interface) interface is implemented by all top-level objects, including the docker [`Pipeline`](#the-pipeline-object).

{{< highlight typescript "linenos=table" >}}
interface Resource {
  kind: string;
  type: string;
  name: string;
}
{{< / highlight >}}

## The `kind` attribute

Defines the kind of resource, used to identify the resource implementation. This attribute is of type `string` and is required.

## The `type` attribute

Defines the type of resource, used to identify the resource implementation. This attribute is of type `string` and is required.

## The `name` attribute

The name of the resource. This value is required and must match `[a-zA-Z0-9_-]`. This value is displayed in the user interface (non-normative) and is used to identify the pipeline (non-normative).

# The `Pipeline` object

The [`Pipeline`](#the-pipeline-object) is the top-level object used to represent the docker pipeline. The [`Pipeline`](#the-pipeline-object) object implements the [`Resource`](#the-resource-interface) interface.

{{< highlight typescript "linenos=table" >}}
class Pipeline : Resource {
  kind:     string;
  type:     string;
  name:     string;
  platform: Platform;
  clone:    Clone;
  steps:    Step[];
  volumes:  Volume[];
  node:     [string, string];
  trigger:  Conditions;

  image_pull_secrets: string[]
}
{{< / highlight >}}

## The `kind` attribute

The kind of resource. This value must be set to `pipeline`.

## The `type` attribute

The type of resource. This value must be set to `docker`.

## The `platform` section

The target operating system and architecture on which the pipeline must execute. This attribute is of type [`Platform`](#the-platform-object) and is recommended. If empty, the default operating system and architecture may be `linux` and `amd64` respectively.

## The `clone` section

Defines the pipeline clone behavior and can be used to disable automatic cloning. This attribute is of type [`Clone`](#the-clone-object) and is optional.

## The `steps` section

Defines the pipeline steps. This attribute is an array of type [`Step`](#the-step-object) and is required. The array must not be empty and the order of the array must be retained.

## The `node` attribute

Defines key value pairs used to route the pipeline to a specific runner or group of runners. This attribute is of type `[string, string]` and is optional.

## The `trigger` section

The conditions used to determine whether or not the pipeline should be skipped. This attribute is of type [`Conditions`](#the-conditions-object) and is optional.

## The `image_pull_secrets` attribute

The list of secrets used to pull private Docker images; This attribute is an array of type `string` and is optional.

# The `Platform` object

The [`Platform`](#the-platform-object) object defines the target os and architecture for the pipeline.

{{< highlight typescript "linenos=table" >}}
class Platform {
  os:      OS;
  arch:    Arch;
  variant: string;
  version: string;
}
{{< / highlight >}}

## The `os` attribute

Defines the target operating system. The attribute is an enumeration of type `OS` and is recommended. If empty the operating system may default to Linux.

## The `arch` attribute

Defines the target architecture. The attribute is an enumeration of type `Arch` and is recommended. If empty the architecture may default to amd64.

## The `variant` attribute

Defines the architecture variant. This is most commonly used in conjunction with the arm architecture (non-normative) and can be used to differentiate between armv7, armv8, and so on (non-normative).

## The `version` attribute

Defines the operating system version. This is most commonly used in conjunction with the windows operating system (non-normative) and can be used to differentiate between 1809, 1903, and so on (non-normative).

# The `Clone` object

The [`Clone`](#the-clone-object) object defines the clone behavior for the pipeline.

{{< highlight typescript "linenos=table" >}}
class Clone {
  depth:   number;
  disable: boolean;
}
{{< / highlight >}}

## The `depth` attribute

Configures the clone depth. This is an optional `number` value. If empty the full repository may be cloned (non-normative).

## The `disable` attribute

Disables cloning the repository. This is an optional `boolean` value. It can be useful when you need to disable implement your own custom clone logic (non-normative).

# The `Step` object

The `Step` object defines a pipeline step.

{{< highlight typescript "linenos=table" >}}
class Step {
  command:     string[];
  commands:    string[];
  detach:      boolean;
  entrypoint:  string[];
  environment: [string, string];
  failure:     Failure;
  image:       string;
  name:        string;
  network_mode string;
  privileged   boolean;
  pull:        Pull;
  user         string;
  volumes:     Volume[];
  when:        Conditions;
}
{{< / highlight >}}


## The `commands` attribute

Defines a list of shell commands executed inside the Docker container. The commands are executed using the default container shell (non-normative) as the container `ENTRYPOINT`. This attribute is an array of type `string` and is required.

## The `command` attribute

Overrides the image `COMMAND`. This should only be used with service containers and cannot not be used with the `commands` attribute. This attribute is an array of type `[string]` and is optional.

## The `detach` attribute

The detach attribute instructions the system to start the Docker container and then run in the background. This value is of type `boolean` and is optional.

## The `entrypoint` attribute

Overrides the image `ENTRYPOINT`. This should only be used with service containers and cannot not be used with the `commands` attribute. This attribute is an array of type `[string]` and is optional.

## The `environment` attribute

Defines a list of environment variables scoped to the pipeline step. This attribute is of type `[string, string]` and is optional.

## The `failure` attribute

Defines how the system handles failure. The default value is `always` indicating a failed step always fails the overall pipeline. A value of `ignore` indicates the failure is ignored. This attribute is of enumeration [`Failure`](#the-failure-enum) and is optional.

## The `image` attribute

The name of the docker image. The image name should include the tag and will default to the latest tag if unspecified. This value is of type `string` and is required.

## The `name` attribute

The name of the step. This value is required and must match [a-zA-Z0-9_-]. This value is displayed in the user interface (non-normative) and is used to identify the step (non-normative).

## The `network_mode` attribute

Overrides the default network to which the Docker container is attached. For example `host` or `bridge`. This attribute is of type `string` and is optional.

## The `privileged` attribute

Overrides the default Docker security policy and grants the container nearly full access to the host machine. This attribute is of type `boolean` and is optional.

## The `pull` attribute

Defines how and when the system should pull images. This attribute is of enumeration [`Pull`](#the-pull-enum) and is optional.

## The `user` attribute

Overrides the default username or uid used when executing the pipeline commands or entrypoint. This attribute is of type `string` and is optional.

## The `when` section

The conditions used to determine whether or not the step should be skipped. This attribute is of type [`Conditions`](#the-conditions-object) and is optional.

# The `Conditions` object

The [`Conditions`](#the-conditions-object) object defines a set of conditions. If any condition evaluates to true its parent object is skipped.

{{< highlight typescript "linenos=table" >}}
class Conditions {
  action:   Constraint | string[];
  branch:   Constraint | string[];
  cron:     Constraint | string[];
  event:    Constraint | Event[];
  instance: Constraint | string[];
  ref:      Constraint | string[];
  repo:     Constraint | string[];
  status:   Constraint | Status[];
  target:   Constraint | string[];
}
{{< / highlight >}}

## The `action` attribute

Defines matching criteria based on the build action. The build action is synonymous with a webhook action (non-normative). This attribute is of type [`Constraint`](#the-constraint-object) or an array of type `string` and is optional.

## The `branch` attribute

Defines matching criteria based on the git branch. This attribute is of type [`Constraint`](#the-constraint-object) or an array of type `string` and is optional.

## The `cron` attribute

Defines matching criteria based on the cron job that triggered the build. This attribute is of type [`Constraint`](#the-constraint-object) or an array of type `string` and is optional.

## The `event` attribute

Defines matching criteria based on the build event. The build event is synonymous with a webhook event (non-normative). This attribute is of type [`Constraint`](#the-constraint-object) or an array of type [`Event`](#the-event-enum) and is optional.

## The `instance` attribute

Defines matching criteria based on the instance hostname. This attribute is of type [`Constraint`](#the-constraint-object) or an array of type `string` and is optional.

## The `ref` attribute

Defines matching criteria based on the git reference. This attribute is of type [`Constraint`](#the-constraint-object) or an array of type `string` and is optional.

## The `repo` attribute

Defines matching criteria based on the repository name. This attribute is of type [`Constraint`](#the-constraint-object) or an array of type `string` and is optional.

## The `status` attribute

Defines matching criteria based on the pipeline status. This attribute is of type [`Constraint`](#the-constraint-object) or an array of type [`Status`](#the-status-enum) and is optional.

## The `target` attribute

Defines matching criteria based on the target environment. The target environment is typically defined by a promote or rollback event (non-normative). This attribute is of type [`Constraint`](#the-constraint-object) or an array of type `string` and is optional.

# The `Constraint` object

The [`Constraint`](#the-constraint-object) object defines pattern matching criteria. If the pattern matching evaluates to false, the parent object is skipped.

{{< highlight typescript "linenos=table" >}}
class Constraint {
  exclude: string[];
  include: string[];
}
{{< / highlight >}}

## The `include` attribute

List of matching patterns. If no pattern is a match, the parent object is skipped. This attribute is an array of type `string` and is optional.

## The `exclude` attribute

List of matching patterns. If any pattern is a match, the parent object is skipped. This attribute is an array of type `string` and is optional.

# The `Secret` object

The [`Secret`](#the-secret-object) defines the named source of a secret.

{{< highlight typescript "linenos=table" >}}
class Secret {
  from_secret: string;
}
{{< / highlight >}}

# Enums

## The `Event` enum

The `Event` enum provides a list of pipeline events. This value represents the event that triggered the pipeline.

{{< highlight typescript "linenos=table" >}}
enum Event {
  cron,
  promote,
  pull_request,
  push,
  rollback,
  tag,
}
{{< / highlight >}}

## The `Status` enum

The `Status` enum provides a list of pipeline statuses. The default pipeline state is `success`, even if the pipeline is still running.

{{< highlight typescript "linenos=table" >}}
enum Status {
  failure,
  success,
}
{{< / highlight >}}

## The `Pull` enum

The `Pull` enum defines if and when a docker image should be pull from the registry.

{{< highlight typescript "linenos=table" >}}
enum Failure {
  always,
  never,
  if-not-exists,
}
{{< / highlight >}}

## The `Failure` enum

The `Failure` enum defines a list of failure behaviors. The value `always` indicates a failure will fail the parent process. The value `ignore` indicates the failure is silently ignored.

{{< highlight typescript "linenos=table" >}}
enum Failure {
  always,
  ignore,
}
{{< / highlight >}}

## The `OS` enum

The `OS` enum provides a list of supported operating systems.

{{< highlight typescript "linenos=table" >}}
enum OS {
  darwin,
  dragonfly,
  freebsd,
  linux,
  netbsd,
  openbsd,
  solaris,
  windows,
}
{{< / highlight >}}

## The `Arch` enum

The `Arch` enum provides a list of supported chip architectures.

{{< highlight typescript "linenos=table" >}}
enum Arch {
  386,
  amd64,
  arm64,
  arm,
}
{{< / highlight >}}
