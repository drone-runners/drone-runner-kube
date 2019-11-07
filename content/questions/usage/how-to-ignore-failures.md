---
date: 2000-01-01T00:00:00+00:00
title: How to ignore steps that fail?
author: bradrydzewski
draft: true
weight: 1
---

Normally when a pipeline step fails, it fails the overall pipeline. You can allow a step to fail without failing the overall pipeline with the following configuration:

{{< highlight text "linenos=table,hl_lines=8" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: foo
  image: alpine
  failure: ignore
  commands:
  - echo foo
  - exit 1

- name: bar
  image: alpine
  commands:
  - echo bar
{{< / highlight >}}
