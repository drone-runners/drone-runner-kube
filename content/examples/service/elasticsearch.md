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

In the below example we demonstrate a pipeline that launches a Elasticsearch service container. The elastic server will be available at `database:9200`, where the hostname matches the service container name.

{{< highlight yaml "hl_lines=18-21" >}}
---
kind: pipeline
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
  - curl http://database:9200

services:
- name: database
  image: elasticsearch:5-alpine
  ports:
  - 9200

...
{{< / highlight >}}

# Common Problems

## Initialization

If you are unable to connect to the Elastic container please make sure you
are giving the instance adequate time to initialize and begin accepting
connections.

{{< highlight yaml "hl_lines=9" >}}
kind: pipeline
name: default

steps:
- name: test
  image: alpine:3.8
  commands:
  - apk add curl
  - sleep 45
  - curl http://database:9200
{{< / highlight >}}

## Incorrect Hostname

You cannot use `127.0.0.1` or `localhost` to connect with the database. If you are unable to connect to the database please verify you are using the correct hostname, corresponding with the name of the container. 

Bad:

```
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
  ports:
  - 9200
```

Good:

```
steps:
- name: test
  image: alpine:3.8
  commands:
  - apk add curl
  - sleep 45
  - curl http://database:9200

services:
- name: database
  image: elasticsearch:5-alpine
  ports:
  - 9200
```



