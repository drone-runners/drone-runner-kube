---
date: 2000-01-01T00:00:00+00:00
title: Node Selection
author: bradrydzewski
weight: 9
toc: false
description: |
  Route pipelines to specific servers.
---

The `node_selector` section can be used to route pipelines to specific Kubernetes nodes, or groups of nodes, that have matching labels. This can be useful when you need to route pipelines to runners with special configurations or hardware.

{{< highlight text "linenos=table,hl_lines=11-14" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: build
  image: golang
  commands:
  - go build
  - go test

node_selector:
  keyA: valueA
  keyB: valueB
{{< / highlight >}}
