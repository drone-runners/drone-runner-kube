---
date: 2000-01-01T00:00:00+00:00
title: Gradle
title_in_header: Example Gradle Pipeline
author: bradrydzewski
weight: 1
toc: false
---

# Gradle Example

This guide covers configuring continuous integration pipelines for Gradle projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `gradle assemble` and `gradle check` commands. These commands are executed inside a Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: gradle:jdk10
  commands:
  - gradle assemble
  - gradle check
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official gradle [images](https://hub.docker.com/r/_/gradle/), or your can bring your own.



