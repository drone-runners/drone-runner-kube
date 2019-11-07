---
date: 2000-01-01T00:00:00+00:00
title: R
title_in_header: Example R Pipeline
author: bradrydzewski
weight: 1
toc: false
---

This guide covers configuring continuous integration pipelines for R projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `R` commands to install dependencies and compile code. These commands are executed inside the r-base Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: r-base
  commands:
  - R -e 'install.packages(c("package1","package2"))'
  - R CMD build .
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official r-base [images](https://hub.docker.com/r/_/r-base/), or your can bring your own.
