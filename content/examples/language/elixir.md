---
date: 2000-01-01T00:00:00+00:00
title: Elixir
title_in_header: Example Elixir Pipeline
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for Elixir projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes a series of `mix` commands. These commands are executed inside the Elixir Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
name: default

steps:
- name: test
  image: elixir:1.5
  commands:
  - mix local.rebar --force
  - mix local.hex --force
  - mix deps.get
  - mix test
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official Elixir [images](https://hub.docker.com/r/_/elixir/), or your can bring your own.

# Dependencies

If you decide to split your pipeline into multiple steps you need to make sure each step has access to project dependencies. Dependencies are downloaded to `/root/.mix` which is outside the shared workspace. Create a named volume to share this directory with all pipeline steps:

```
kind: pipeline
name: default

steps:
- name: install
  image: elixir:1.5
  volumes:
  - name: mix
    path: /root/.mix
  commands:
  - mix local.rebar --force
  - mix local.hex --force
  - mix deps.get

- name: test
  image: elixir:1.5
  volumes:
  - name: mix
    path: /root/.mix
  commands:
  - mix test

volumes:
- name: mix
  temp: {}
```

# Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of Elixir. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
name: elixir-1-5

steps:
- name: test
  image: elixir:1.5
  commands:
  - mix local.rebar --force
  - mix local.hex --force
  - mix deps.get
  - mix test

---
kind: pipeline
name: elixir-1-4

steps:
- name: test
  image: elixir:1.4
  commands:
  - mix local.rebar --force
  - mix local.hex --force
  - mix deps.get
  - mix test

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
        "mix local.rebar --force",
        "mix local.hex --force",
        "mix deps.get",
        "mix test",
      ]
    }
  ]
};

[
  Pipeline("elixir-1-5", "elixir:1.5"),
  Pipeline("elixir-1-4", "elixir:1.4"),
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
  image: elixir
  commands:
  - mix local.rebar --force
  - mix local.hex --force
  - mix deps.get
  - mix test

---
kind: pipeline
name: test-on-arm64

platform:
  arch: arm64

steps:
- name: test
  image: elixir
  commands:
  - mix local.rebar --force
  - mix local.hex --force
  - mix deps.get
  - mix test

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.
