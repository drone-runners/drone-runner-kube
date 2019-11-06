---
date: 2000-01-01T00:00:00+00:00
title: Couchbase
author: bradrydzewski
weight: 37
draft: true
toc: false
---

This guide covers configuring continuous integration pipelines for projects that have a Couchbase dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a Couchbase service container. The database server will be available at `database:1234`, where the hostname matches the service container name.
