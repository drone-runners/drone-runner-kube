---
date: 2000-01-01T00:00:00+00:00
title: DRONE_NAMESPACE_RULES
author: bradrydzewski
weight: 1
---

Optional string map. Defines linting rules to prevent an organization or repository from using a restricted namespace.

```
DRONE_NAMESPACE_RULES=default:cisco/*,development:Webex/*
```

For example, the above rule prevents a pipeline from using the `development` namespace unless the repository name matches `Webex/*`
