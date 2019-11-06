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

In the below example we demonstrate a pipeline that launches a Nats service container. The nats service will be available at `nats:4222`, where the hostname matches the service container name.

```
kind: pipeline
name: default

steps:
- name: test
  image: ruby:2
  commands:
  - gem install nats
  - nats-pub -s tcp://nats:4222 greeting 'hello'
  - nats-pub -s tcp://nats:4222 greeting 'world'

services:
- name: nats
  image: nats:1.3.0
  ports:
  - 4222
  - 8222
```
