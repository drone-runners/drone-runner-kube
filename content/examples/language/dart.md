---
date: 2000-01-01T00:00:00+00:00
title: Dart
title_in_header: Example Dart Pipeline
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for Dart projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `pub get` and `pub run test` commands. These commands are executed inside the dart Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
name: default

steps:
- name: test
  image: google/dart
  commands:
  - pub get
  - pub run test
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official Dart [images](https://hub.docker.com/r/google/dart/), or your can bring your own.

# Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of Dart. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
name: dart2

steps:
- name: test
  image: google/dart:2
  commands:
  - pub get
  - pub run test

---
kind: pipeline
name: dart1

steps:
- name: test
  image: google/dart:1
  commands:
  - pub get
  - pub run test

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
        "pub get",
        "pub run test"
      ]
    }
  ]
};

[
  Pipeline("dart1", "dart:2"),
  Pipeline("dart2", "dart:1"),
]
```
