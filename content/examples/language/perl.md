---
date: 2000-01-01T00:00:00+00:00
title: Perl
title_in_header: Example Perl Pipeline
author: bradrydzewski
weight: 1
---

This guide covers configuring continuous integration pipelines for Perl projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that installs the project dependnecies using `cpanm`, and then executes the project unit tests. These commands are executed inside a Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: perl
  commands:
  - cpanm --quiet --installdeps --notest .
  - perl Build.PL
  - ./Build test
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official perl [images](https://hub.docker.com/r/_/perl/), or your can bring your own.
