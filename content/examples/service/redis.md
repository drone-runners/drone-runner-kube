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

In the below example we demonstrate a pipeline that launches a Redis service container. The database server will be available at `redis:6379`, where the hostname matches the service container name.

{{< highlight yaml "hl_lines=13-15" >}}
kind: pipeline
name: default

steps:
- name: test
  image: redis
  commands:
  - sleep 5
  - redis-cli -h redis ping
  - redis-cli -h redis set FOO bar
  - redis-cli -h redis get FOO

services:
- name: redis
  image: redis
  ports:
  - 6379
{{< / highlight >}}

# Common Problems

## Initialization

If you are unable to connect to the Postgres container please make sure you
are giving Postgres adequate time to initialize and begin accepting
connections.

{{< highlight yaml "hl_lines=8" >}}
kind: pipeline
name: default

steps:
- name: test
  image: redis
  commands:
  - sleep 15
  - redis-cli -h redis ping
{{< / highlight >}}

## Incorrect Hostname

You cannot use `127.0.0.1` or `localhost` to connect with the Redis container. If you are unable to connect to the Redis container please verify you are using the correct hostname, corresponding with the name of the redis service container. 

Bad:

```
steps:
- name: test
  image: redis
  commands:
  - sleep 15
  - redis-cli -h localhost ping
```

Good:

```
steps:
- name: test
  image: redis
  commands:
  - sleep 15
  - redis-cli -h redis ping
```

