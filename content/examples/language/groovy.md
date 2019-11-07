---
date: 2000-01-01T00:00:00+00:00
title: Groovy
title_in_header: Example Groovy Pipeline
author: bradrydzewski
weight: 1
toc: false
---

This guide covers configuring continuous integration pipelines for Groovy projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `./gradlew assemble` and `./gradlew check` commands. These commands are executed inside a Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: gradle:2.5-jdk8
  commands:
  - ./gradlew assemble
  - ./gradlew check
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official groovy [images](https://hub.docker.com/r/_/groovy/), or your can bring your own.
