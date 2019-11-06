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

In the below example we demonstrate a pipeline that launches a CouchDB service container. The database server will be available at `database:5984`, where the hostname matches the service container name.

```
---
kind: pipeline
name: default

platform:
  os: linux
  arch: amd64

steps:
- name: test
  image: couchdb:2.2
  commands:
  - sleep 15
  - curl http://database:5984

services:
- name: database
  image: couchdb:2.2
  ports:
  - 5984

...
```

# Common Problems

## Initialization

If you are unable to connect to the CouchDB container please make sure you
are giving the instance adequate time to initialize and begin accepting
connections.

{{< highlight yaml "hl_lines=8" >}}
kind: pipeline
name: default

steps:
- name: test
  image: coucdb:2.2
  commands:
  - sleep 15
  - curl http://database:5984
{{< / highlight >}}

## Incorrect Hostname

You cannot use `127.0.0.1` or `localhost` to connect with the database. If you are unable to connect to the database please verify you are using the correct hostname, corresponding with the name of the container. 

Bad:

```
steps:
- name: test
  image: couchdb:2.2
  commands:
  - sleep 15
  - curl http://localhost:5984

services:
- name: database
  image: couchdb:2.2
  ports:
  - 5984
```

Good:

```
steps:
- name: test
  image: couchdb:2.2
  commands:
  - sleep 15
  - curl http://database:5984

services:
- name: database
  image: couchdb:2.2
  ports:
  - 5984
```
