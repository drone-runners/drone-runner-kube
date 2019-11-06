---
date: 2000-01-01T00:00:00+00:00
title: Docker (dind)
title_in_header: Example Docker-in-Docker Configuration
author: bradrydzewski
weight: 1
toc: false
---

This guide covers configuring continuous integration pipelines for projects that have a Docker dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a Docker service container (Docker-in-Docker). The service container is run in privileged mode. For security reasons, only trusted repositories can enable privileged mode.

```
---
kind: pipeline
name: default

steps:
- name: test
  image: docker:dind
  volumes:
  - name: dockersock
    path: /var/run
  commands:
  - sleep 5 # give docker enough time to start
  - docker ps -a

services:
- name: docker
  image: docker:dind
  privileged: true
  volumes:
  - name: dockersock
    path: /var/run

volumes:
- name: dockersock
  temp: {}
```
