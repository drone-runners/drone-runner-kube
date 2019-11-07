---
date: 2000-01-01T00:00:00+00:00
title: Nats
title_in_header: Example Nats Configuration
author: bradrydzewski
weight: 1
toc: false
---

This guide covers configuring continuous integration pipelines for projects that have a Nats dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Basic Example

In the below example we demonstrate a pipeline that launches a Nats service container. The service will be available at `localhost:4222`.

```
kind: pipeline
name: default

steps:
- name: test
  image: ruby:2
  commands:
  - gem install nats
  - nats-pub greeting 'hello'
  - nats-pub greeting 'world'

services:
- name: nats
  image: nats:1.3.0
```
