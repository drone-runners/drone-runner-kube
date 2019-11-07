---
date: 2000-01-01T00:00:00+00:00
title: Node
title_in_header: Node Example
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for Node projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `npm install` and `npm test` commands. These commands are executed inside the node Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: node
  commands:
  - npm install
  - npm test
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official node [images](https://hub.docker.com/r/_/node/), or your can bring your own.

# Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of Node. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
type: kubernetes
name: node6

steps:
- name: test
  image: node:6
  commands:
  - npm install
  - npm test

---
kind: pipeline
name: node8

steps:
- name: test
  image: node:8
  commands:
  - npm install
  - npm test

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
        "npm install",
        "npm test"
      ]
    }
  ]
};

[
  Pipeline("node6", "node:6"),
  Pipeline("node8", "node:8"),
]
```

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
  image: node
  commands:
  - npm install
  - npm test

---
kind: pipeline
type: kubernetes
name: test-on-arm64

platform:
  arch: arm64

steps:
- name: test
  image: node
  commands:
  - npm install
  - npm test

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.
