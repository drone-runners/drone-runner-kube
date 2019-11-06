---
date: 2000-01-01T00:00:00+00:00
title: DRONE_BUILD_ACTION
author: bradrydzewski
---

Provides the action that triggered the pipeline execution. Use this value to differentiate between the pull request being opened vs synchronized.

```
DRONE_BUILD_ACTION=sync
DRONE_BUILD_ACTION=open
```
