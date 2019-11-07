---
date: 2000-01-01T00:00:00+00:00
title: Swift
title_in_header: Example Swift Pipeline
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for Swift projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes the project unit tests with the `swift build` and `swift test` commands. These commands are executed inside a Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: swift:4
  commands:
  - swift build
  - swift test
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official swift [images](https://hub.docker.com/r/_/swift/), or your can bring your own.


# Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of Swift. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
type: kubernetes
name: swift3

steps:
- name: test
  image: swift:3
  commands:
  - swift build
  - swift test

---
kind: pipeline
type: kubernetes
name: swift4

steps:
- name: test
  image: swift:4
  commands:
  - swift build
  - swift test

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.

```
local Pipeline(version) = {
  kind: "pipeline",
  type: "kubernetes",
  name: "swift"+version,
  steps: [
    {
      name: "test",
      image: "swift:"+version,
      commands: [
        "swift build",
        "swift test"
      ]
    }
  ]
};

[
  Pipeline("3"),
  Pipeline("4"),
]
```
