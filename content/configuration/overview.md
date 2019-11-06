---
date: 2000-01-01T00:00:00+00:00
title: Overview
author: bradrydzewski
weight: -2
toc: false
description: |
  Overview of Docker pipelines.
---

A `docker` pipeline is a pipeline that executes shell commands inside Docker containers. Docker containers provide isolation allowing safe execution of concurrent pipelines on the same machine.

A major benefit of container pipelines is the ability to bring your own build environment, in the form of a Docker image. Drone automatically downloads docker images at runtime.

Example pipeline configuration:

{{< highlight text "linenos=table" >}}
---
kind: pipeline
type: docker
name: default

steps:
- name: greeting
  image: golang:1.12
  commands:
  - go build
  - go test

...
{{< / highlight >}}

The kind and type attributes define a Docker pipeline.

{{< highlight text "linenos=table" >}}
---
kind: pipline
type: docker
{{< / highlight >}}

The `steps` section defines a series of shell commands. These commands are executed inside the Docker container as the `Entrypoint`. If any command returns a non-zero exit code, the pipeline fails and exits.

{{< highlight text "linenos=table,linenostart=6" >}}
steps:
- name: greeting
  image: golang:1.12
  commands:
  - go build
  - go test
{{< / highlight >}}
