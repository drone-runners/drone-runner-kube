---
date: 2000-01-01T00:00:00+00:00
title: DRONE_TARGET_BRANCH
author: bradrydzewski
---

Provides the target branch for the push or pull request. This value may be empty for tag events.

```
DRONE_TARGET_BRANCH=master
```

This environment variable can be used in conjunction with the source branch variable to get the pull request base and head branch.

```
DRONE_SOURCE_BRANCH=feature/develop
DRONE_TARGET_BRANCH=master
```
