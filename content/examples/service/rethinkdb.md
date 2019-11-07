---
date: 2000-01-01T00:00:00+00:00
title: RethinkDB
title_in_header: Example RethinkDB Configuration
author: bradrydzewski
weight: 1
toc: true
source: https://github.com/drone-demos/drone-demo-rethinkdb
---

This guide covers configuring continuous integration pipelines for projects that have a rethinkdb dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a rethinkdb service container. The server can be reached at `localhost:28015`.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: node:9
  commands:
  - npm install -s -g recli
  - recli -h database -j 'r.db("localhost").table("stats")'

services:
- name: database
  image: rethinkdb:2
```
