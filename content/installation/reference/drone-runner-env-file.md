---
date: 2000-01-01T00:00:00+00:00
title: DRONE_RUNNER_ENV_FILE
author: bradrydzewski
weight: 1
---

Optional string value. Provides the path to an environment variable file used to define global environment variables injected into all pipeline steps. The environment file format is documented [here](https://github.com/joho/godotenv/blob/master/README.md).

```
DRONE_RUNNER_ENV_FILE=/etc/drone.conf
```

Remember to mount this file from the host machine into the Docker container. _Note the configuration path defined above should point to the file path inside the container._

```
docker run \
--volume /etc/drone.conf:/etc/drone.conf
```
