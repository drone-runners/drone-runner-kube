---
date: 2000-01-01T00:00:00+00:00
title: MongoDB
title_in_header: Example MongoDB Configuration
author: bradrydzewski
weight: 1
toc: true
aliases: [/mongodb-example/]
source: https://github.com/drone-demos/drone-demo-mongodb
---

This guide covers configuring continuous integration pipelines for projects that have a MongoDB dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a Mongo service container. The server will be available at `localhost:27017`.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: ping
  image: mongo:4
  commands:
  - sleep 5
  - mongo --eval "db.version()"

services:
- name: mongo
  image: mongo:4
  command: [ --smallfiles ]
```

# Common Problems

If you are unable to connect to the Mongo container please make sure you
are giving the instance adequate time to initialize and begin accepting
connections.

{{< highlight yaml "linenos=table,hl_lines=9" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: ping
  image: mongo:4
  commands:
  - sleep 5
  - mongo --eval "db.version()"
{{< / highlight >}}
