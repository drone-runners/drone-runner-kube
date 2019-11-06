---
date: 2000-01-01T00:00:00+00:00
title: MariaDB
title_in_header: Example MariaDB Configuration
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for projects that have a MariaDB dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a MariaDB service container. The database server will be available at `database:3306`, where the hostname matches the service container name.

{{< highlight yaml "hl_lines=10-16" >}}
kind: pipeline
name: default

steps:
- name: test
  image: mariadb
  commands:
  - sleep 15
  - mysql -u root -h database --execute="SELECT VERSION();"

services:
- name: database
  image: mariadb
  ports:
  - 3306
  environment:
    MYSQL_ALLOW_EMPTY_PASSWORD: 'yes'
    MYSQL_DATABASE: test
{{< / highlight >}}

# Database Settings

The official MariaDB image provides environment variables used at startup
to create the default username, password, database and more. Please see the
official image [documentation](https://hub.docker.com/_/mariadb/) for more details.

{{< highlight yaml "hl_lines=4-6" >}}
services:
- name: database
  image: mariadb
  ports:
  - 3306
  environment:
    MYSQL_DATABASE: test
    MYSQL_ALLOW_EMPTY_PASSWORD: 'yes'
{{< / highlight >}}

# Common Problems

## Initialization

If you are unable to connect to the MariaDB container please make sure you
are giving MariaDB adequate time to initialize and begin accepting
connections.

{{< highlight yaml "hl_lines=8" >}}
kind: pipeline
name: default

steps:
- name: test
  image: mariadb
  ports:
  - 3306
  commands:
  - sleep 15
  - mysql -u root -h database
{{< / highlight >}}

## Incorrect Hostname

You cannot use `127.0.0.1` or `localhost` to connect with the MariaDB container. If you are unable to connect to MariaDB please verify you are using the correct hostname, corresponding with the name of the container. 

Bad:

```
steps:
- name: test
  image: mariadb
  commands:
  - sleep 15
  - mysql -u root -h localhost
```

Good:

```
steps:
- name: test
  image: mariadb
  commands:
  - sleep 15
  - mysql -u root -h database

services:
- name: database
  image: mariadb
```
