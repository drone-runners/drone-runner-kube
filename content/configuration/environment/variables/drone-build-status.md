---
date: 2000-01-01T00:00:00+00:00
title: DRONE_BUILD_STATUS
author: bradrydzewski
---

Provides the status for the current running build. If build pipelines and build steps are passing, the build status defaults to success.

```
DRONE_BUILD_STATUS=success
DRONE_BUILD_STATUS=failure
```

_Please note this is point in time snapshot. This value may not accurately reflect the overall build status when multiple pipelines are running in parallel_.
