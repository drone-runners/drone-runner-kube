---
date: 2000-01-01T00:00:00+00:00
title: Go (with Gopath)
title_in_header: Example Go Pipeline
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for Go projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `go get` and `go test` commands. These commands are executed inside the golang Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
name: default

workspace:
  base: /go
  path: src/github.com/octocat/hello-world

steps:
- name: test
  image: golang
  commands:
  - go get
  - go test
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official golang [images](https://hub.docker.com/r/_/golang/), or your can bring your own.

# GOPATH

If you are not using Go modules you will need to override the default project workspace to ensure your code is cloned to the correct location in the `GOPATH`.

{{< highlight yaml "hl_lines=4-6" >}}
kind: pipeline
name: default

workspace:
  base: /go
  path: src/github.com/octocat/hello-world

steps:
- name: test
  image: golang
  commands:
  - go get
  - go test
{{< / highlight >}}

# Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of Go. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
name: go-1-11

steps:
- name: test
  image: golang:1.11
  commands:
  - go get
  - go test

---
kind: pipeline
name: go-1-10

steps:
- name: test
  image: golang:1.10
  commands:
  - go get
  - go test

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
        "go get",
        "go test"
      ]
    }
  ]
};

[
  Pipeline("go-1-11", "golang:1.11"),
  Pipeline("go-1-12", "golang:1.12"),
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
  image: golang
  commands:
  - go get
  - go test

---
kind: pipeline
name: test-on-arm64

platform:
  arch: arm64

steps:
- name: test
  image: golang
  commands:
  - go get
  - go test

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.
