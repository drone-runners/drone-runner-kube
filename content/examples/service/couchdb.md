---
date: 2000-01-01T00:00:00+00:00
title: CouchDB
title_in_header: Example CouchDB Configuration
author: bradrydzewski
weight: 1
toc: true
source: https://github.com/drone-demos/drone-demo-couchdb
---

This guide covers configuring continuous integration pipelines for projects that have a CouchDB dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a CouchDB service container. The server will be available at `localhost:5984`.

```
---
kind: pipeline
type: kubernetes
name: default

platform:
  os: linux
  arch: amd64

steps:
- name: test
  image: couchdb:2.2
  commands:
  - sleep 15
  - curl http://localhost:5984

services:
- name: database
  image: couchdb:2.2

...
```

# Common Problems

If you are unable to connect to the CouchDB container please make sure you
are giving the instance adequate time to initialize and begin accepting
connections.

{{< highlight yaml "linenos=table,hl_lines=9" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: coucdb:2.2
  commands:
  - sleep 15
  - curl http://localhost:5984
{{< / highlight >}}
