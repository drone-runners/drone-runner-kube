---
date: 2000-01-01T00:00:00+00:00
title: DRONE_LIMIT_TRUSTED
author: bradrydzewski
weight: 1
---

Optional boolean value. Configures the runner to only process trusted repositories. This provides an extra layer of security and can stop untrusted repositories from executing pipelines with this runner.

```
DRONE_LIMIT_TRUSTED=true
```