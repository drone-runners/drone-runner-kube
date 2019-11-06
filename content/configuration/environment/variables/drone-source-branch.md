---
date: 2000-01-01T00:00:00+00:00
title: DRONE_SOURCE_BRANCH
author: bradrydzewski
---

Provides the source branch for the pull request. This value may be empty for certain source control management providers.

```
DRONE_SOURCE_BRANCH=feature/develop
```

This environment variable can be used in conjunction with the target branch variable to get the pull request base and head branch.

```
DRONE_SOURCE_BRANCH=feature/develop
DRONE_TARGET_BRANCH=master
```
