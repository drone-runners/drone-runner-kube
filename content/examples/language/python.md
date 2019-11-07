---
date: 2000-01-01T00:00:00+00:00
title: Python
author: bradrydzewski
weight: 1
toc: true
---

# Python Example

This guide covers configuring continuous integration pipelines for Python projects. If you're new to Drone please read our Tutorial and build configuration guides first.

## Build and Test

In the below example we demonstrate a pipeline that executes `pip install` and `pytest` commands. These commands are executed inside a Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: python
  commands:
  - pip install -r requirements.txt
  - pytest
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official python [images](https://hub.docker.com/r/_/python/), or your can bring your own.

## Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of Python. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
type: kubernetes
name: python-2

steps:
- name: test
  image: python:2
  commands:
  - pip install -r requirements.txt
  - pytest

---
kind: pipeline
type: kubernetes
name: python-3-3

steps:
- name: test
  image: python:3.3
  commands:
  - pip install -r requirements.txt
  - pytest

---
kind: pipeline
type: kubernetes
name: python-3-4

steps:
- name: test
  image: python:3.4
  commands:
  - pip install -r requirements.txt
  - pytest
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
        "pip install -r requirements.txt",
        "pytest"
      ]
    }
  ]
};

[
  Pipeline("python-2", "python:2"),
  Pipeline("python-3-3", "python:3.3"),
  Pipeline("python-3-4", "python:3.4"),
  Pipeline("python-3-5", "python:3.5"),
  Pipeline("python-3-6", "python:3.6"),
]
```

## Test Multiple Architectures

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
  image: python
  commands:
  - pip install -r requirements.txt
  - pytest

---
kind: pipeline
type: kubernetes
name: test-on-arm64

platform:
  arch: arm64

steps:
- name: test
  image: python
  commands:
  - pip install -r requirements.txt
  - pytest

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.
