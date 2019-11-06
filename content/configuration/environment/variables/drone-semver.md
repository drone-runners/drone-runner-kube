---
date: 2000-01-01T00:00:00+00:00
title: DRONE_SEMVER
author: bradrydzewski
---

If the git tag is a valid semver string, provides the tag as a semver string.

```
DRONE_SEMVER=1.2.3-alpha.1
```

The semver string is also available in the following formats:

```
DRONE_SEMVER_SHORT=1.2.3
DRONE_SEMVER_PATCH=3
DRONE_SEMVER_MINOR=2
DRONE_SEMVER_MAJOR=1
DRONE_SEMVER_PRERELEASE=alpha.1
```
