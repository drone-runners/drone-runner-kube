---
date: 2000-01-01T00:00:00+00:00
title: Redis
title_in_header: Example Redis Configuration
author: bradrydzewski
weight: 1
toc: true
source: https://github.com/drone-demos/drone-demo-redis
aliases: [/redis-example/]
---

This guide covers configuring continuous integration pipelines for projects that have a Redis dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a Redis service container. The server can be reached at `localhost:6379`.

{{< highlight yaml "linenos=table,hl_lines=13-16" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: redis
  commands:
  - sleep 5
  - redis-cli -h localhost ping
  - redis-cli -h localhost set FOO bar
  - redis-cli -h localhost get FOO

services:
- name: redis
  image: redis
{{< / highlight >}}

# Common Problems

If you are unable to connect to the Redis container please make sure you
are giving Redis adequate time to initialize and begin accepting
connections.

{{< highlight yaml "linenos=table,hl_lines=9" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: redis
  commands:
  - sleep 15
  - redis-cli -h localhost ping
{{< / highlight >}}
