---
date: 2000-01-01T00:00:00+00:00
title: How to execute steps on failure?
author: bradrydzewski
weight: 1
draft: true
---


```
kind: pipeline
type: kubernetes
name: default

steps:
- name: foo
  commands:
  - echo foo
  - exit 1

- name: bar
  commands:
  - echo bar
  when:
    status: [ failure ]
```