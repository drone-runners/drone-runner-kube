---
date: 2000-01-01T00:00:00+00:00
title: Tolerations
author: bradrydzewski
weight: 8
toc: false
description: |
  Prevent pipelines from being scheduled on inappropriate nodes.
---

The `tolerations` section can be used in conjunction with taints to ensure pipelines are not scheduled onto inappropriate nodes.

{{< link-to "https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/" >}}
Learn more about Taints and Tolerations
{{</ link-to >}}

Example configuration:

{{< highlight text "linenos=table,hl_lines=11-15" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: build
  image: golang
  commands:
  - go build
  - go test

tolerations:
- key: example-key
  operator: Exists
  effect: NoSchedule
{{< / highlight >}}
