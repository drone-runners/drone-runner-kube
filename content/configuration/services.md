---
date: 2000-01-01T00:00:00+00:00
title: Services
author: bradrydzewski
weight: 5
toc: true
description: |
  Configure service containers.
---

Drone supports launching detached service containers as part of your pipeline. The typical use case for services is when your unit tests require a running redis server, for example:

{{< highlight text "linenos=table,hl_lines=5-7" >}}
kind: pipeline
type: kubernetes
name: default

services:
- name: cache
  image: redis
{{< / highlight >}}

Service containers share the same network as your pipeline steps and can be access at the localhost address. In our previous example, the redis container can be accessed from the pipeline at `tcp://127.0.0.1:6379`

{{< highlight text "linenos=table,hl_lines=9" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: ping
  image: redis
  commands:
  - redis-cli -h 127.0.0.1 ping

services:
- name: cache
  image: redis
{{< / highlight >}}

It is important to note the service container exit code is ignored, and a non-zero exit code does not fail the overall pipeline. Drone expects service containers to exit with a non-zero exit code, since they often need to be killed after the pipeline completes.

# Detached Steps

Services can also be defined directly in the pipeline, as detached pipeline steps. This can be useful when you need direct control over when the service is started, relative to other steps in your pipeline.

{{< highlight text "linenos=table,hl_lines=8" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: cache
  image: redis
  detach: true

- name: ping
  image: redis
  commands:
  - redis-cli ping
{{< / highlight >}}

# Common Problems

This section highlights some common problems that users encounter when configuring services. If you continue to experience issues please also check the faq. You might also want to compare your yaml to our example service configurations.

## Initialization

It is important to remember that after a container is started, the software running inside the container (e.g. redis) takes time to initialize and begin accepting connections.

Be sure to give the service adequate time to initialize before attempting to connect. A naive solution is to use the sleep command.

{{< highlight patch >}}
kind: pipeline
type: kubernetes
name: default

steps:
  - name: ping
    image: redis
    commands:
+   - sleep 5
    - redis-cli -h cache ping

services:
  - name: cache
    image: redis
{{< / highlight >}}
