---
date: 2000-01-01T00:00:00+00:00
title: Erlang
title_in_header: Example Erlang Pipeline
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for Erlang projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `rebar` commands. These commands are executed inside the Erlang Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: erlang:21
  commands:
  - rebar get-deps
  - rebar compile
  - rebar skip_deps=true eunit
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official Erlang [images](https://hub.docker.com/r/_/erlang/), or your can bring your own.

# Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of Erlang. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
type: kubernetes
name: erlang21

steps:
- name: test
  image: erlang:21
  commands:
  - rebar get-deps
  - rebar compile
  - rebar skip_deps=true eunit

---
kind: pipeline
type: kubernetes
name: erlang20

steps:
- name: test
  image: erlang:20
  commands:
  - rebar get-deps
  - rebar compile
  - rebar skip_deps=true eunit

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.

```
local Pipeline(version) = {
  kind: "pipeline",
  type: "kubernetes",
  name: "erlang"+version,
  steps: [
    {
      name: "test",
      image: "erlang:"+version,
      commands: [
        "rebar get-deps",
        "rebar compile",
        "rebar skip_deps=true eunit",
      ]
    }
  ]
};

[
  Pipeline("21"),
  Pipeline("20"),
  Pipeline("19"),
  Pipeline("18"),
]
```
