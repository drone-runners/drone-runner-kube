---
date: 2000-01-01T00:00:00+00:00
title: Rust
title_in_header: Example Rust Pipeline
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for Rust projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `cargo build` and `cargo test` commands. These commands are executed inside the rust Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
name: default

steps:
- name: test
  image: rust:1.30
  commands:
  - cargo build --verbose --all
  - cargo test --verbose --all
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official rust [images](https://hub.docker.com/r/_/rust/), or your can bring your own.

# Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test multiple versions of Rust. This is equivalent to matrix capabilities found in other continuous integration systems.

```
kind: pipeline
name: rust-1-30

steps:
- name: test
  image: rust:1.30
  commands:
  - cargo build --verbose --all
  - cargo test --verbose --all

---
kind: pipeline
name: rust-1-29

steps:
- name: test
  image: rust:1.29
  commands:
  - cargo build --verbose --all
  - cargo test --verbose --all
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
        "cargo build --verbose --all",
        "cargo test --verbose --all"
      ]
    }
  ]
};

[
  Pipeline("rust-1-29", "rust:1.29"),
  Pipeline("rust-1-30", "rust:1.30"),
]
```

# Test Multiple Architectures

You can use Drone's multi-pipeline feature to concurrently test your code on multiple architectures and operating systems.

```
kind: pipeline
name: test-on-amd64

platform:
  arch: amd64

steps:
- name: test
  image: rust:1.30
  commands:
  - cargo build --verbose --all
  - cargo test --verbose --all

---
kind: pipeline
name: test-on-arm64

platform:
  arch: arm64

steps:
- name: test
  image: rust:1.30
  commands:
  - cargo build --verbose --all
  - cargo test --verbose --all
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.
