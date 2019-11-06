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

In the below example we demonstrate a pipeline that launches a rethinkdb service container. The database server will be available at `database:28015`, where the hostname matches the service container name.

```
kind: pipeline
name: default

steps:
- name: test
  image: node:9
  commands:
  - npm install -s -g recli
  - recli -h database -j 'r.db("rethinkdb").table("stats")'

services:
- name: database
  image: rethinkdb:2
  command: [ rethinkdb, --bind, all ]
  ports:
  - 28015
```

# Common Problems

## Incorrect Hostname

You cannot use `127.0.0.1` or `localhost` to connect with the database. If you are unable to connect please verify you are using the correct hostname, corresponding with the container name. 

Bad:

```
kind: pipeline
name: default

steps:
- name: test
  image: node:9
  commands:
  - npm install -s -g recli
  - recli -h localhost -j 'r.db("rethinkdb").table("stats")'

services:
- name: database
  image: rethinkdb:2
  command: [ rethinkdb, --bind, all ]
  ports:
  - 28015
```

Good:

```
kind: pipeline
name: default

steps:
- name: test
  image: node:9
  commands:
  - npm install -s -g recli
  - recli -h database -j 'r.db("rethinkdb").table("stats")'

services:
- name: database
  image: rethinkdb:2
  command: [ rethinkdb, --bind, all ]
  ports:
  - 28015
```
