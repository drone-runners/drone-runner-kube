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

In the below example we demonstrate a pipeline that launches a MariaDB service container. The database server will be available at `localhost:3306`.

{{< highlight yaml "linenos=table,hl_lines=11-17" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: mariadb
  commands:
  - sleep 15
  - mysql -u root --execute="SELECT VERSION();"

services:
- name: database
  image: mariadb
  environment:
    MYSQL_ALLOW_EMPTY_PASSWORD: 'yes'
    MYSQL_DATABASE: test
{{< / highlight >}}

# Database Settings

The official MariaDB image provides environment variables used at startup
to create the default username, password, database and more. Please see the
official image [documentation](https://hub.docker.com/_/mariadb/) for more details.

{{< highlight yaml "linenos=table,hl_lines=4-6" >}}
services:
- name: database
  image: mariadb
  environment:
    MYSQL_DATABASE: test
    MYSQL_ALLOW_EMPTY_PASSWORD: 'yes'
{{< / highlight >}}

# Common Problems

If you are unable to connect to the MariaDB container please make sure you
are giving MariaDB adequate time to initialize and begin accepting
connections.

{{< highlight yaml "linenos=table,hl_lines=9" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: mariadb
  commands:
  - sleep 15
  - mysql -u root
{{< / highlight >}}
