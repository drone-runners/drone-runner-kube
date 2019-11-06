---
date: 2000-01-01T00:00:00+00:00
title: DRONE_RUNNER_PRIVILEGED_IMAGES
author: bradrydzewski
weight: 1
---

Optional comma separated list. Provides a list of Docker images that are started as privileged containers by default.

{{< alert "security" >}}
Privileged mode effectively grants the container root access to your host machine. Please use with caution.
{{< / alert >}}

```
DRONE_RUNNER_PRIVILEGED_IMAGES=plugins/docker,plugin/ecr
```
