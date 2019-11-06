---
date: 2000-01-01T00:00:00+00:00
title: Crystal
title_in_header: Example Crystal Pipeline
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for Crystal projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `shards install` and `crystal spec` commands. These commands are executed inside the crystal Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
name: default

steps:
- name: test
  image: crystallang/crystal
  commands:
  - shards install
  - crystal spec
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official crystal [images](https://hub.docker.com/r/crystallang/crystal/), or your can bring your own.

# Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of Crystal. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
name: nightly

steps:
- name: test
  image: crystallang/crystal:nightly
  pull: always
  commands:
  - shards install
  - crystal spec

---
kind: pipeline
name: latest

steps:
- name: test
  image: crystallang/crystal:latest
  commands:
  - shards install
  - crystal spec

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.

```
local Pipeline(name, image) = {
  kind: "pipeline",
  name: name,
  steps: [
    {
      name: "test",
      image: image,
      commands: [
        "shards install",
        "crystal spec"
      ]
    }
  ]
};

[
  Pipeline("nightly", "crystal:nightly"),
  Pipeline("latest", "crystal:latest"),
]
```
