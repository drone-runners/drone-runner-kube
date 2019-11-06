---
date: 2000-01-01T00:00:00+00:00
title: Host Volumes
author: bradrydzewski
weight: 19
toc: false
description: |
  Mount host volumes to access the host filesystem in your pipeline.
---

Host mounts allow you to mount an absolute path on the host machine into a pipeline step. This setting is only available to trusted repositories.

{{< alert "security" >}}
This setting is only available to trusted repositories, since mounting host machine volumes is a security risk.
{{< / alert >}}

{{< highlight text "linenos=table,hl_lines=8-10 15-18" >}}
kind: pipeline
type: docker
name: default

steps:
- name: build
  image: node
  volumes:
  - name: cache
    path: /tmp/cache
  commands:
  - npm install
  - npm test

volumes:
- name: cache
  host:
    path: /var/lib/cache
{{< / highlight >}}

The first step is to define the host machine volume path. The host volume path must be an absolute path.

{{< highlight text "linenos=table,linenostart=15" >}}
volumes:
- name: cache
  host:
    path: /var/lib/cache
{{< / highlight >}}

The next step is to configure your pipeline step to mount the named host path into your container. The container path must also be an absolute path.

{{< highlight text "linenos=table,linenostart=5" >}}
steps:
- name: build
  image: node
  volumes:
  - name: cache
    path: /tmp/cache
{{< / highlight >}}
