---
date: 2000-01-01T00:00:00+00:00
title: Docker
author: bradrydzewski
weight: 1
toc: false
draft: true
---

# Docker Plugin Example

The easiest way to build and publish Docker images is to use the Docker plugin. This plugin uses Docker-in-Docker to build your Docker image in an isolated container, and push to the Docker registry. Please see the official plugin documentation for more details.

Example Configuration:

```
---
kind: pipeline
name: default

steps:
- name: test
  image: golang
  commands:
  - go test
  - go build

- name: publish
  image: plugins/docker
  settings:
    repo: octocat/hello-world
    tags: [ latest ]
    username:
      from_secret: username
    password:
      from_secret: password
...
```
