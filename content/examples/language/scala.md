---
date: 2000-01-01T00:00:00+00:00
title: Scala
author: bradrydzewski
weight: 1
toc: true
draft: true
---

# Scala Example

This guide covers configuring continuous integration pipelines for Scala projects. If you're new to Drone please read our Tutorial and build configuration guides first.

## Build and Test

In the below example we demonstrate a pipeline that executes `sbt` commands. These command are executed inside the openjdk Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: UNKNOWN
  commands:
  - sbt compile
  - sbt test
```

Please note that you can use any valid Docker image in your pipeline from any Docker registry. You can use the official openjdk images, or your can bring your own.

## Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of the JDK. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
type: kubernetes
name: scala211

steps:
- name: test
  image: UNKNOWN
  commands:
  - sbt compile
  - sbt test

---
kind: pipeline
type: kubernetes
name: scala210

steps:
- name: test
  image: UNKNOWN
  commands:
  - sbt compile
  - sbt test

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.

```
local Pipeline(name, image) = {
  kind: "pipeline",
  type: "kubernetes",
  name: name,
  steps: [
    {
      name: "test",
      image: image,
      commands: [
        "sbt compile",
        "sbt test"
      ]
    }
  ]
};

[
  Pipeline("scala211", "UNKNOWN:2.11"),
  Pipeline("scala210", "UNKNOWN:1.10"),
]
```
