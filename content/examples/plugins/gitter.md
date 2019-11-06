---
date: 2000-01-01T00:00:00+00:00
title: Gitter
author: bradrydzewski
weight: 1
toc: false
draft: true
---

# Gitter Plugin Example

The easiest way to send a Gitter notification is to use the Gitter plugin. Please see the official plugin documentation for more details.

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

- name: notify
  image: plugins/gitter
  settings:
    webhook:
      from_secret: webhook
...
```
