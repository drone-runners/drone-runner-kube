---
date: 2000-01-01T00:00:00+00:00
title: DRONE_LIMIT_EVENTS
author: bradrydzewski
weight: 1
---

Optional comma-separated string value. Provides a white list of build events that can be processed by this runner. This provides an extra layer of security to limit the kind of workloads this runner can process.

```
DRONE_LIMIT_EVENTS=push,tag
```