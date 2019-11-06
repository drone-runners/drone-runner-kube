---
date: 2000-01-01T00:00:00+00:00
title: DRONE_RUNNER_MAX_PROCS
author: bradrydzewski
weight: 1
---

Optional umber value. Limits the number of concurrent steps that a runner can execute for a single pipeline. This is disabled by default. This can be useful if you need to throttle the maximum number of parallel steps to prevent resource exhaustion.

```
DRONE_RUNNER_MAX_PROCS=10
```