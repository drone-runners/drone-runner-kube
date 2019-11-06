---
date: 2000-01-01T00:00:00+00:00
title: How to share files between steps?
author: bradrydzewski
weight: 1
draft: true
---

{{< highlight text "linenos=table,hl_lines=7" >}}
kind: pipeline
type: docker
name: default

steps:
- name: write
  image: alpine
  commands:
  - echo "hello" > greeting.txt

- name: read
  image: alpine
  commands:
  - cat greeting.txt

{{< / highlight >}}
