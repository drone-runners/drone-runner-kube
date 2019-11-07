---
date: 2000-01-01T00:00:00+00:00
title: Elasticsearch
title_in_header: Example Elasicsearch Configuration
author: bradrydzewski
weight: 1
toc: true
source: https://github.com/drone-demos/drone-demo-elasticsearch
---

This guide covers configuring continuous integration pipelines for projects that have a Elasticsearch dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a Elasticsearch service container. The server will be available at `localhost:9200`.

{{< highlight yaml "linenos=table,hl_lines=18-21" >}}
---
kind: pipeline
type: kubernetes
name: default

platform:
  os: linux
  arch: amd64

steps:
- name: test
  image: alpine:3.8
  commands:
  - apk add curl
  - sleep 45
  - curl http://localhost:9200

services:
- name: database
  image: elasticsearch:5-alpine

...
{{< / highlight >}}

# Common Problems

If you are unable to connect to the Elastic container please make sure you
are giving the instance adequate time to initialize and begin accepting
connections.

{{< highlight yaml "linenos=table,hl_lines=10" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: alpine:3.8
  commands:
  - apk add curl
  - sleep 45
  - curl http://localhost:9200
{{< / highlight >}}
