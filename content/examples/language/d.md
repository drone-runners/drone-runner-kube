---
date: 2000-01-01T00:00:00+00:00
title: D
title_in_header: Example D Pipeline
author: bradrydzewski
weight: 1
toc: false
---

This guide covers configuring continuous integration pipelines for D projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `dub test` command to compile and test your code. These commands are executed inside a Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: dlanguage/dmd
  commands:
  - dub test
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official dmd [images](https://hub.docker.com/r/dlanguage/dmd/), or your can bring your own.
