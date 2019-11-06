---
date: 2000-01-01T00:00:00+00:00
title: Docker
author: bradrydzewski
weight: 1
toc: false
draft: true
---

# Docker Manifest Example

The easiest way to tag multi-architecture and multi-os images is with the Docker manifest plugin. This is typically used in conjunction with the Docker plugin to push a manifest for images created in the pipeline.

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
    tags: [ linux-amd64 ]
    username:
      from_secret: username
    password:
      from_secret: password

- name: manifest
  image: plugins/manifest:1
  settings:
    repo: octocat/hello-world
    tags: [ latest ]
    username:
      from_secret: username
    password:
      from_secret: password
...
```
