---
date: 2000-01-01T00:00:00+00:00
title: Postgres
title_in_header: Example Postgres Configuration
author: bradrydzewski
weight: 1
toc: true
source: https://github.com/drone-demos/drone-demo-postgres
aliases: [/postgres-example/]
---

This guide covers configuring continuous integration pipelines for projects that have a Postgres dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a Postgres service container. The database server will be available at `database:5432`, where the hostname matches the service container name.

{{< highlight yaml "hl_lines=10-16" >}}
kind: pipeline
name: default

steps:
- name: test
  image: postgres:9-alpine
  commands:
  - psql -U postgres -d test -h database

services:
- name: database
  image: postgres:9-alpine
  ports:
  - 5432
  environment:
    POSTGRES_USER: postgres
    POSTGRES_DB: test
{{< / highlight >}}

# Database Settings

The official Postgres image provides environment variables used at startup
to create the default username, password, database and more. Please see the
official image [documentation](https://hub.docker.com/_/postgres/) for more details.

{{< highlight yaml "hl_lines=6-8" >}}
services:
- name: database
  image: postgres
  ports:
  - 5432
  environment:
    POSTGRES_USER: postgres
    POSTGRES_DB: test
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
  image: postgres
  commands:
  - sleep 15
  - psql -U postgres -d test -h database
{{< / highlight >}}

## Incorrect Hostname

You cannot use `127.0.0.1` or `localhost` to connect with the Postgres container. If you are unable to connect to the Postgres container please verify you are using the correct hostname, corresponding with the name of the postgres service container. 

Bad:

```
steps:
- name: test
  image: postgres
  commands:
  - sleep 15
  - psql -U postgres -d test -h localhost
```

Good:

```
steps:
- name: test
  image: postgres
  commands:
  - sleep 15
  - psql -U postgres -d test -h database
```

