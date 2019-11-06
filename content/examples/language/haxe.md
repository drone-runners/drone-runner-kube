---
date: 2000-01-01T00:00:00+00:00
title: Haxe
title_in_header: Example Haxe Pipeline
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for Haxe projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `npm install` and `npm test` commands. These commands are executed inside the haxe Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
name: default

steps:
- name: test
  image: haxe
  commands:
  - haxelib install build.hxml
  - haxe build.hxml
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official haxe [images](https://hub.docker.com/r/_/haxe/), or your can bring your own.

# Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of Haxe. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
name: haxe4

steps:
- name: test
  image: haxe:4.0
  commands:
  - haxelib install build.hxml
  - haxe build.hxml

---
kind: pipeline
name: haxe3

steps:
- name: test
  image: haxe:4.0
  commands:
  - haxelib install build.hxml
  - haxe build.hxml

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
        "haxelib install build.hxml",
        "haxe build.hxml"
      ]
    }
  ]
};

[
  Pipeline("haxe4", "haxe:4.0"),
  Pipeline("haxe3", "haxe:3.0"),
]
```

# Test Multiple Architectures

You can use Drone's multi-pipeline feature to concurrently test your code on multiple architectures and operating systems.

```
---
kind: pipeline
name: test-on-amd64

platform:
  arch: amd64

steps:
- name: test
  image: haxe
  commands:
  - haxelib install build.hxml
  - haxe build.hxml

---
kind: pipeline
name: test-on-amd64

platform:
  arch: amd64

steps:
- name: test
  image: haxe
  commands:
  - haxelib install build.hxml
  - haxe build.hxml

---
kind: pipeline
name: test-on-arm64

platform:
  arch: arm64

steps:
- name: test
  image: haxe
  commands:
  - haxelib install build.hxml
  - haxe build.hxml

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.
