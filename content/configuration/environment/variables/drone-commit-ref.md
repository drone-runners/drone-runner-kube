---
date: 2000-01-01T00:00:00+00:00
title: DRONE_COMMIT_REF
author: bradrydzewski
---

Provides the git reference for the current running build.

```
DRONE_COMMIT_REF=refs/heads/master
```

Example tag reference:

```
DRONE_COMMIT_REF=refs/tags/v1.0.0
```

Example pull request reference (GitHub):
```
DRONE_COMMIT_REF=refs/pull/42/head
```