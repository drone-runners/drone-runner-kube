---
date: 2000-01-01T00:00:00+00:00
title: Platform
author: bradrydzewski
weight: 2
toc: false
description: |
  Configure the target operating system and architecture.
---

Use the `platform` section to configure the target operating system and architecture and routes the pipeline to the appropriate runner.

Example linux arm64 pipeline:

{{< highlight text "linenos=table,hl_lines=5-7" >}}
kind: pipeline
type: kubernetes
name: default

platform:
  os: linux
  arch: arm64

steps:
- name: build
  image: golang
  commands:
  - go build
  - go test
{{< / highlight >}}
