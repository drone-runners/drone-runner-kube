---
date: 2000-01-01T00:00:00+00:00
title: Install on Linux
author: bradrydzewski
weight: 1
toc: true
description: |
  Install the runner using Docker.
---

This article explains how to install the Docker runner on Linux. The Docker runner is packaged as a minimal Docker image distributed on [DockerHub](https://hub.docker.com/r/drone/drone-runner-ssh), and is available for the following architectures:

* amd64
* arm
* arm64

# Step 1 - Download

Install Docker and pull the public image:

```
$ docker pull drone/drone-runner-docker:1
```

# Step 2 - Configure

The Docker runner is configured using environment variables. This article references the below configuration options. See [Configuration]({{< relref "reference" >}}) for a complete list of configuration options.

`DRONE_RPC_HOST`
: provides the hostname (and optional port) of your Drone server. The runner connects to the server at the host address to receive pipelines for execution.

`DRONE_RPC_PROTO`
: provides the protocol used to connect to your Drone server. The value must be either http or https.

`DRONE_RPC_SECRET`
: provides the shared secret used to authenticate with your Drone server. This must match the secret defined in your Drone server configuration.

# Step 3 - Install

The below command creates the a container and start the Docker runner. _Remember to replace the environment variables below with your Drone server details._

```
$ docker run -d \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DRONE_RPC_PROTO=https \
  -e DRONE_RPC_HOST=drone.company.com \
  -e DRONE_RPC_SECRET=super-duper-secret \
  -e DRONE_RUNNER_CAPACITY=2 \
  -e DRONE_RUNNER_NAME=${HOSTNAME} \
  -p 3000:3000 \
  --restart always \
  --name runner \
  drone/drone-runner-docker:1
```

# Step 4 - Verify

Use the `docker logs` command to view the logs and verify the runner successfully established a connection with the Drone server.

```
$ docker logs runner

INFO[0000] starting the server
INFO[0000] successfully pinged the remote server 
```
