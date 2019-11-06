---
date: 2000-01-01T00:00:00+00:00
title: Go (with Modules)
title_in_header: Example Go Pipeline
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for Go projects that use Go Modules. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `go test` and `go build` commands. These commands are executed inside the golang Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
name: default

steps:
- name: test
  image: golang
  commands:
  - go test
  - go build
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official golang [images](https://hub.docker.com/r/_/golang/), or your can bring your own.

# Dependencies

If you decide to split your pipeline into multiple steps you need to make sure each step has access to project dependencies. Dependencies are downloaded to `/go` which is outside the shared workspace. Create a named volume to share this directory with all pipeline steps:

{{< highlight yaml "hl_lines=7-9 15-17 21-23" >}}
kind: pipeline
name: default

steps:
- name: test
  image: golang
  volumes:
  - name: deps
    path: /go
  commands:
  - go test

- name: build
  image: golang
  volumes:
  - name: deps
    path: /go
  commands:
  - go build

volumes:
- name: deps
  temp: {}
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
