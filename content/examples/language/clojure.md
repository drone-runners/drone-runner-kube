---
date: 2000-01-01T00:00:00+00:00
title: Clojure
title_in_header: Example Clojure Pipeline
author: bradrydzewski
weight: 1
---

This guide covers configuring continuous integration pipelines for Clojure projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes the `lein test` command. These commands are executed inside the clojure Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
name: default

steps:
- name: test
  image: clojure
  commands:
  - lein test
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official Clojure [images](https://hub.docker.com/r/_/clojure/), or your can bring your own.
