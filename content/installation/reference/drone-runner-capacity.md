---
date: 2000-01-01T00:00:00+00:00
title: DRONE_RUNNER_CAPACITY
author: bradrydzewski
weight: 1
---

Optional number value. Limits the number of concurrent pipelines that a runner can execute. This does _not_ limit the number of concurrent pipelines that can execute on a single remote instance.

```
DRONE_RUNNER_CAPACITY=10
```