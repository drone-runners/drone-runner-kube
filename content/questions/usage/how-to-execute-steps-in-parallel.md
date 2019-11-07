---
date: 2000-01-01T00:00:00+00:00
title: How to execute steps in parallel?
author: bradrydzewski
weight: 1
draft: true
---



```
kind: pipeline
type: kubernetes
name: greeting

steps:
- name: one
  commands:
  - echo one

- name: two
  commands:
  - echo two

- name: three
  commands:
  - echo three
  depends_on:
  - one
  - two

- name: four
  commands:
  - echo four
  depends_on:
  - three
```