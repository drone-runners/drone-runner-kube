---
date: 2000-01-01T00:00:00+00:00
title: DRONE_STAGE_FINISHED
author: bradrydzewski
---

Provides the unix timestamp for when the pipeline is finished. A running pipeline cannot have a finish timestamp, therefore, the system always sets this value to the current timestamp.

```
DRONE_STAGE_FINISHED=915148800
```
