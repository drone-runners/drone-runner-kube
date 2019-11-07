---
date: 2000-01-01T00:00:00+00:00
title: Overview
author: bradrydzewski
weight: -2
toc: true
description: |
  Overview of Kubernetes pipelines.
---

A `kubernetes` pipeline executes pipeline steps as containers inside Kubernetes pods. Containers provide isolation allowing safe execution of concurrent pipelines on the same machine.

A major benefit of container-based pipelines is the ability to bring your own build environment, in the form of a Docker image. Drone automatically downloads docker images at runtime.

{{< alert warn >}}
Please note Kubernetes pipelines are not a drop-in replacement for Docker pipelines. The configuration and runtime behaviors may differ.
{{< / alert >}}

Example pipeline configuration:

{{< highlight text "linenos=table" >}}
---
kind: pipeline
type: kubernetes
name: default

steps:
- name: greeting
  image: golang:1.12
  commands:
  - go build
  - go test

...
{{< / highlight >}}

The kind and type attributes define a Kubernetes pipeline.

{{< highlight text "linenos=table" >}}
---
kind: pipline
type: kubernetes
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

# Kubernetes vs Docker Pipelines

A Kubernetes pipeline and Docker pipeline share many similarities, but they should not be considered drop-in replacements for one another. There are a few notable differences in the configuration syntax and runtime behavior.

- Kubernetes pipelines are scheduled to execute in the same Pod and therefore share the same network. This means services are accessible at a `localhost` address vs a custom hostname.
- Kubernetes pipelines are scheduled by Kubernetes which provides advanced affinity options. The Kubernetes runner exposes Node Selector capabilities to the pipeline using the `node_selector` attribute.
- Kubernetes containers automatically mount service account credentials to `/var/run/secrets/kubernetes.io/serviceaccount`. This may have security implications and may impact plugins that integrate with Kubernetes.

# Known Issues

Kubernetes pipelines are considered experimental and may not be suitable for production use yet. You may experience unexpected issues, some of which are detailed below.

- The pipeline status is not correctly passed to containers, impacting plugins that rely on this value. This primarily impacts notification plugins, such as Slack, which will always report the pipeline status as success.
- The command line utility does not support linting, formatting or exection of Kubernetes pipelines.