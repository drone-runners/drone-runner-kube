---
date: 2000-01-01T00:00:00+00:00
title: Steps
author: bradrydzewski
weight: 4
toc: true
description: |
  Configure pipeline steps.
---

Pipeline steps are defined as a series of shell commands. The commands are executed inside the root directory of your git repository. The root of your git repository, also called the workspace, is shared by all steps in your pipeline.

Example configuration:

{{< highlight text "linenos=table" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: backend
  image: golang
  commands:
  - go build
  - go test

- name: frontend
  image: node
  commands:
  - npm install
  - npm test
{{< / highlight >}}

# Commands

The commands are executed inside the root directory of your git repository. The root of your git repository, also called the workspace, is shared by all steps in your pipeline. This allows file artifacts to persist between steps.

{{< highlight text "linenos=table,linenostart=5" >}}
steps:
- name: backend
  image: golang
  commands:
  - go build
  - go test
{{< / highlight >}}

The above commands are converted to a simple shell script. The commands in the above example are roughly converted to the below script:

{{< highlight text "linenos=table" >}}
#!/bin/sh
set -e
set -x

go build
go test
{{< / highlight >}}

The above shell script is then executed as the docker entrypoint. The below docker command is an (incomplete) example of how the script is executed:

```
docker run --entrypoint=build.sh golang
```

The container exit code is used to determine whether the step is passing or failing. If a command returns a non-zero exit code, the step is marked as failing. The overall pipeline status is also marked as failing, and remaining pipeline steps are skipped (_unless explicitly configured to run on failure_).

# Environment

The environment section provides the ability to define environment variables scoped to individual pipeline steps.

{{< highlight text "linenos=table,hl_lines=4-6,linenostart=5" >}}
steps:
- name: backend
  image: golang
  environment:
    GOOS: linux
    GOARCH: amd64
  commands:
  - go build
  - go test
{{< / highlight >}}

See the Environment article for additional details:

{{< link "/configuration/environment/overview" >}}

# Plugins

Plugins are docker containers that encapsulate commands, and can be shared and re-used in your pipeline. Examples of plugins include sending Slack notifications, building and publishing Docker images, and uploading artifacts to S3.

Example Slack plugin:

{{< highlight text "linenos=table,hl_lines=5-9,linenostart=15" >}}
- name: notify
  image: plugins/slack
  settings:
    webhook: https://hooks.slack.com/services/...
{{< / highlight >}}

The great thing about plugins is they are just Docker containers. This means you can easily encapsulate logic, bundle in a Docker container, and share your plugin with your organization or with the broader community.

{{< link-to "http://plugins.drone.io" >}}
Plugin Registry
{{< / link-to >}}

# Conditions

The when section provides the ability to conditionally limit the execution of steps at runtime. The below example limits step execution by branch, however, you can limit execution by event, reference, status and more.

{{< highlight text "linenos=table,hl_lines=7-9,linenostart=5" >}}
steps:
- name: backend
  image: golang
  commands:
  - go build
  - go test
  when:
    branch:
    - master
{{< / highlight >}}

Use the status condition to override the default runtime behavior and execute steps even when the pipeline status is failure:

{{< highlight text "linenos=table,hl_lines=5-9,linenostart=15" >}}
- name: notify
  image: plugins/slack
  settings:
    webhook: https://hooks.slack.com/services/...
  when:
    status:
    - failure
    - success
{{< / highlight >}}

See the Conditions article for additional details:

{{< link "conditions" >}}

# Failure

The failure attribute lets you customize how the system handles failure of an individual step. This can be useful if you want to allow a step to fail without failing the overall pipeline.

{{< highlight text "linenos=table,hl_lines=4,linenostart=5" >}}
steps:
- name: backend
  image: golang
  failure: ignore
  commands:
  - go build
  - go test
{{< / highlight >}}

# Detach

The detach attribute lets execute the pipeline step in the background. The runner starts the step, detaches and runs in the background, and immediately proceeds to the next step.

The target use case for this feature is to start a service or daemon, and then execute unit tests against the service in subsequent steps.

{{< alert "info" >}}
Note that a detached step cannot fail the pipeline. The runner may ignore the exit code.
{{< / alert >}}

{{< highlight text "linenos=table,hl_lines=4,linenostart=5" >}}
steps:
- name: backend
  image: golang
  detach: true
  commands:
  - go build
  - go test
  - go run main.go -http=:3000
{{< / highlight >}}

# Privileged Mode

The privileged attribute runs the container with escalated privileges. This is the equivalent of running a container with the `--privileged` flag.

{{< alert "security" >}}
This setting is only available to trusted repositories. Privileged mode effectively grants the container root access to your host machine. Please use with caution.
{{< / alert >}}

{{< highlight text "linenos=table,hl_lines=4,linenostart=5" >}}
steps:
- name: backend
  image: golang
  privileged: true
  commands:
  - go build
  - go test
  - go run main.go -http=:3000
{{< / highlight >}}