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

In the below example we demonstrate a pipeline that launches a Mongo service container. The database server will be available at `mongo:27017`, where the hostname matches the service container name.

```
kind: pipeline
name: default

steps:
- name: ping
  image: mongo:4
  commands:
  - sleep 5
  - mongo --host mongo --eval "db.version()"

services:
- name: mongo
  image: mongo:4
  command: [ --smallfiles ]
  ports:
  - 27017
```

# Common Problems

## Initialization

If you are unable to connect to the Mongo container please make sure you
are giving the instance adequate time to initialize and begin accepting
connections.

{{< highlight yaml "hl_lines=8" >}}
kind: pipeline
name: default

steps:
- name: ping
  image: mongo:4
  commands:
  - sleep 5
  - mongo --host localhost --eval "db.version()"
{{< / highlight >}}

## Incorrect Hostname

You cannot use `127.0.0.1` or `localhost` to connect with the Mongo container. If you are unable to connect to Mongo please verify you are using the correct hostname, corresponding with the name of the container. 

Bad:

```
steps:
- name: ping
  image: mongo:4
  commands:
  - sleep 5
  - mongo --host localhost --eval "db.version()"

services:
- name: mongo
  image: mongo:4
  command: [ --smallfiles ]
  ports:
  - 27017
```

Good:

```
steps:
- name: ping
  image: mongo:4
  commands:
  - sleep 5
  - mongo --host mongo --eval "db.version()"

services:
- name: mongo
  image: mongo:4
  command: [ --smallfiles ]
  ports:
  - 27017
```
