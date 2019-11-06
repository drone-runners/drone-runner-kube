---
date: 2000-01-01T00:00:00+00:00
title: Substitution
author: bradrydzewski
weight: 2
toc: false
description: |
  Learn how Drone emulates bash string substitution.
---

Drone provides the ability to substitute repository and build metadata to facilitate dynamic pipeline configurations.

Example commit substitution:

{{< highlight text "linenos=table,hl_lines=9" >}}
kind: pipeline
type: docker
name: default

steps:
- name: publish
  image: plugins/docker
  settings:
    tags: ${DRONE_COMMIT}
    repo: octocat/hello-world
{{< / highlight >}}

Example tag substitution:

{{< highlight text "linenos=table,hl_lines=5,linenostart=5" >}}
steps:
- name: publish
  image: plugins/docker
  settings:
    tags: ${DRONE_TAG}
    repo: octocat/hello-world
{{< / highlight >}}

# String Operations

Drone provides partial emulation for bash string operations. This can be used to manipulate string values prior to substitution.

Example variable substitution with substring:

{{< highlight text "linenos=table,hl_lines=5,linenostart=5" >}}
steps:
- name: publish
  image: plugins/docker
  settings:
    tags: ${DRONE_COMMIT_SHA:0:8}
    repo: octocat/hello-world
{{< / highlight >}}

Example variable substitution strips v prefix from v1.0.0:

{{< highlight text "linenos=table,hl_lines=5,linenostart=5" >}}
steps:
- name: publish
  image: plugins/docker
  settings:
    tags: ${DRONE_TAG##v}
    repo: octocat/hello-world
{{< / highlight >}}

# String Operations Reference

List of emulated string operations:

OPERATION	        | DESC
--------------------|---
`${param}`          | parameter substitution
`${param,}`         | parameter substitution with lowercase first char
`${param,,}`        | parameter substitution with lowercase
`${param^}`         | parameter substitution with uppercase first char
`${param^^}`        | parameter substitution with uppercase
`${param:pos}`      | parameter substitution with substring
`${param:pos:len}`  | parameter substitution with substring and length
`${param=default}`  | parameter substitution with default
`${param##prefix}`  | parameter substitution with prefix removal
`${param%%suffix}`  | parameter substitution with suffix removal
`${param/old/new}`  | parameter substitution with find and replace