---
date: 2000-01-01T00:00:00+00:00
title: MySQL
title_in_header: Example MySQL Configuration
author: bradrydzewski
weight: 1
source: https://github.com/drone-demos/drone-demo-mysql
aliases: [/mysql-example/]
toc: true
---

This guide covers configuring continuous integration pipelines for projects that have a MySQL dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a MySQL service container. The database server will be available at `database:3306`, where the hostname matches the service container name.

{{< highlight yaml "hl_lines=10-16" >}}
kind: pipeline
name: default

steps:
- name: test
  image: mysql
  commands:
  - sleep 15
  - mysql -u root -h database --execute="SELECT VERSION();"

services:
- name: database
  image: mysql
  ports:
  - 3306
  environment:
    MYSQL_ALLOW_EMPTY_PASSWORD: 'yes'
    MYSQL_DATABASE: test
{{< / highlight >}}

# Database Options

If you need to start the mysql container with additional runtime options you can override the entrypoint and command arguments.

```
services:
- name: database
  image: mysql
  ports:
  - 3306
  entrypoint: [ "mysqld" ]
  command: [ "--character-set-server=utf8mb4" ]
```

# Database Settings

The official MySQL image provides environment variables used at startup
to create the default username, password, database and more. Please see the
official image [documentation](https://hub.docker.com/_/mysql/) for more details.

{{< highlight yaml "hl_lines=4-6" >}}
services:
- name: database
  image: mysql
  ports:
  - 3306
  environment:
    MYSQL_DATABASE: test
    MYSQL_ALLOW_EMPTY_PASSWORD: 'yes'
{{< / highlight >}}

# Common Problems

## Initialization

If you are unable to connect to the MySQL container please make sure you
are giving MySQL adequate time to initialize and begin accepting
connections.

{{< highlight yaml "hl_lines=8" >}}
kind: pipeline
name: default

steps:
- name: test
  image: mysql
  commands:
  - sleep 15
  - mysql -u root -h database
{{< / highlight >}}

## Incorrect Hostname

You cannot use `127.0.0.1` or `localhost` to connect with the Mysql container. If you are unable to connect to Mysql please verify you are using the correct hostname, corresponding with the name of the mysql service container. 

Bad:

```
steps:
- name: test
  image: mysql
  commands:
  - sleep 15
  - mysql -u root -h localhost
```

Good:

```
steps:
- name: test
  image: mysql
  commands:
  - sleep 15
  - mysql -u root -h database

services:
- name: database
  image: mysql
```

