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
type: docker
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

If you are running Docker pipelines on windows you must specify the operating system version number.

Example windows 1809 pipeline:

{{< highlight text "linenos=table,linenostart=5,hl_lines=4" >}}
platform:
  os: windows
  arch: amd64
  version: 1809
{{< / highlight >}}

# Supported Platforms

os          | arch    | version
------------|---------|---
`linux`     | `amd64` |
`linux`     | `arm`   |
`linux`     | `arm64` |
`windows`   | `amd64` | `1809`
`windows`   | `amd64` | `1903`
`windows`   | `amd64` | `1909 (coming soon)`
