---
date: 2000-01-01T00:00:00+00:00
title: C
title_in_header: Example C Pipeline
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for C projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `make` and `make test` commands. These commands are executed inside the gcc Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: gcc
  commands:
  - ./configure
  - make
  - make test
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official gcc [images](https://hub.docker.com/r/_/gcc/), or your can bring your own.

# Test Multiple Architectures

You can use Drone's multi-pipeline feature to concurrently test your code on multiple architectures and operating systems.

```
---
kind: pipeline
type: kubernetes
name: test-on-amd64

platform:
  arch: amd64

steps:
- name: test
  image: gcc
  commands:
  - ./configure
  - make
  - make test

---
kind: pipeline
type: kubernetes
name: test-on-arm64

platform:
  arch: arm64

steps:
- name: test
  image: gcc
  commands:
  - ./configure
  - make
  - make test

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.

```
local Pipeline(version, arch) = {
  kind: "pipeline",
  type: "kubernetes",
  name: "test-on-"+arch,
  platform: {
    arch: arch,
  }
  steps: [
    {
      name: "test",
      image: "gcc",
      commands: [
        "./configure",
        "make",
        "make test"
      ]
    }
  ]
};

[
  Pipeline("arm"),
  Pipeline("arm64"),
  Pipeline("amd64"),
]
```
