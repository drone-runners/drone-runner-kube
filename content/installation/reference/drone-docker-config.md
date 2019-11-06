---
date: 2000-01-01T00:00:00+00:00
title: DRONE_DOCKER_CONFIG
author: bradrydzewski
weight: 1
---

Optional string value. Provides the path to a Docker `config.json` file used to source registry credentials from third party system.

```
DRONE_DOCKER_CONFIG=/root/.docker/config.json
```

Note that you will typically need to mount this file from the host machine into your Docker container. _Note the configuration path defined above should point to the file path inside the container._

```
docker run \
--volume /root/.docker/config.json:/root/.docker/config.json
```
