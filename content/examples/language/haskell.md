---
date: 2000-01-01T00:00:00+00:00
title: Haskell
title_in_header: Example Haskell Pipeline
author: bradrydzewski
weight: 1
---

This guide covers configuring continuous integration pipelines for Haskell projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `cabal` commands. These commands are executed inside the Haskell Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: haskell
  commands:
  - cabal install --only-dependencies --enable-tests
  - cabal configure --enable-tests
  - cabal build
  - cabal test
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official Haskell [images](https://hub.docker.com/r/_/haskell/), or your can bring your own.

# Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of Haskell. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
type: kubernetes
name: haskell8

steps:
- name: test
  image: haskell:8
  commands:
  - cabal install --only-dependencies --enable-tests
  - cabal configure --enable-tests
  - cabal build
  - cabal test

---
kind: pipeline
type: kubernetes
name: haskell7

steps:
- name: test
  image: haskell:7
  commands:
  - cabal install --only-dependencies --enable-tests
  - cabal configure --enable-tests
  - cabal build
  - cabal test

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.

```
local Pipeline(version) = {
  kind: "pipeline",
  type: "kubernetes",
  name: "haskell"+version,
  steps: [
    {
      name: "test",
      image: "haskell:"+version,
      commands: [
        "cabal install --only-dependencies --enable-tests",
        "cabal configure --enable-tests",
        "cabal build",
        "cabal test"
      ]
    }
  ]
};

[
  Pipeline("7"),
  Pipeline("8"),
]
```
