---
date: 2000-01-01T00:00:00+00:00
title: Julia
title_in_header: Example Julia Pipeline
author: bradrydzewski
weight: 1
draft: true
---

This guide covers configuring continuous integration pipelines for Julia projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `pub get` and `pub run test` commands. These commands are executed inside the julia Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: julia:1
  commands:
  - julia deps/build.jl
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official julia [images](https://hub.docker.com/r/_/julia/), or your can bring your own.
