---
date: 2000-01-01T00:00:00+00:00
title: DRONE_FAILED_STAGES
author: bradrydzewski
---

Provides a comma-separate list of failed pipeline stages for the current running build.

```
DRONE_FAILED_STAGES=build,test
```

_Please note this is point in time snapshot. This value may not accurately reflect the latest results when multiple pipelines are running in parallel_.
