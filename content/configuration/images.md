---
date: 2000-01-01T00:00:00+00:00
title: Images
author: bradrydzewski
weight: 6
toc: true
description: |
  Configure pipeline images.
---

Pipeline steps are defined as a series of Docker containers. Each step must therefore define the Docker image used to create the container.

{{< highlight text "linenos=table,hl_lines=7" >}}
kind: pipeline
type: docker
name: default

steps:
- name: build
  image: golang:1.12
  commands:
  - go build
  - go test
{{< / highlight >}}

Drone supports any valid Docker image from any Docker registry:

```
image: golang
image: golang:1.7
image: library/golang:1.7
image: index.docker.io/library/golang
image: index.docker.io/library/golang:1.7
image: docker.company.com/golang
```

# Pulling Images

If the image does not exist in the local cache, Drone instructs Docker to pull the image automatically. You will never need to manually download or install Docker images.

If the image is tagged with latest, Drone will always attempt to pull the latest version of the image. Configure the runner to only pull the image if not found in the local cache:

{{< highlight text "linenos=table,linenostart=5,hl_lines=3" >}}
steps:
- name: build
  pull: if-not-exists
  image: golang
{{< / highlight >}}

To always pull the latest version of the image:

{{< highlight text "linenos=table,linenostart=5,hl_lines=3" >}}
steps:
- name: build
  pull: always
  image: golang:1.12
{{< / highlight >}}

To never pull the image and always use the image in the local cache:

{{< highlight text "linenos=table,linenostart=5,hl_lines=3" >}}
steps:
- name: build
  pull: never
  image: golang:1.12
{{< / highlight >}}

# Pulling Private Images

If the image is private you need to provide Drone with docker credentials, sourced from a secret. You can manage secrets in your repository settings screen in the web interface.

First create a secret that includes your Docker credentials in the format of a Docker config.json file. This file may provide credentials for multiple registries.

```
{
    "auths": {
        "docker.io": {
            "auth": "4452D71687B6BC2C9389C3..."
        }
    }
}
```

Next, define which secrets should be used to pull private images using the image_pull_secrets attribute:

{{< highlight text "linenos=table,linenostart=5,hl_lines=8-9" >}}
steps:
- name: build
  image: registry.internal.company.com/golang:1.12
  commands:
  - go build
  - go test

image_pull_secrets:
- dockerconfig
{{< / highlight >}}

{{< alert "info" >}}
If you want to pull private images from Amazon Elastic Container Registry (ECR) you will need to install a registry credential plugin.
{{< / alert >}}