---
date: 2000-01-01T00:00:00+00:00
title: Java
title_in_header: Example Java Pipeline
author: bradrydzewski
weight: 1
toc: false
---

# Java Example

This guide covers configuring continuous integration pipelines for Java projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `mvn install` and `mvn test` commands. These commands are executed inside a Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: maven:3-jdk-10
  commands:
  - mvn install -DskipTests=true -Dmaven.javadoc.skip=true -B -V
  - mvn test -B
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official [maven](https://hub.docker.com/r/_/maven/) or [openjdk](https://hub.docker.com/r/_/openjdk/) images, or your can bring your own.
