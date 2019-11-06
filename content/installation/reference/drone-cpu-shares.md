---
date: 2000-01-01T00:00:00+00:00
title: DRONE_CPU_SHARES
author: bradrydzewski
weight: 1
---

Optional integer value. Set this flag to a value greater or less than the default of 1024 to increase or reduce a pipeline container's weight, and give it access to a greater or lesser proportion of the host machine's CPU cycles.

```
DRONE_CPU_SHARES=1024
```