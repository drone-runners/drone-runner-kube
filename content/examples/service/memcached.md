---
date: 2000-01-01T00:00:00+00:00
title: Memcached
title_in_header: Example Memcached Configuration
author: bradrydzewski
weight: 1
---

This guide covers configuring continuous integration pipelines for projects that have a Memcached dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a Memcached service container. The memecache server will be available at `cache:11211`, where the hostname matches the service container name.

```
kind: pipeline
name: default

steps:
- name: test
  image: ubuntu
  commands:
  - apt-get update -qq
  - apt-get install -y -qq telnet > /dev/null
  - (sleep 1; echo "stats"; sleep 2; echo "quit";) | telnet cache 11211 || true

services:
- name: cache
  image: memcached:alpine
  command: [ -vv ]
  ports:
  - 11211
```
