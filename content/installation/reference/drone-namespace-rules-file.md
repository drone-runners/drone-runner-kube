---
date: 2000-01-01T00:00:00+00:00
title: DRONE_NAMESPACE_RULES_FILE
author: bradrydzewski
weight: 1
---

Optional string value. Provides the path to a file that defines namespace usage rules. These rules prevent an organization or repository from using a restricted namespace.

```
DRONE_NAMESPACE_RULES_FILE=/path/to/rules.yml
```

The namespace rules are defined using yaml syntax. The key is the namespace name, and the value is an array of matching patterns.


```
default:
- cisco/*

development:
- cisco/openh264
- Webex/*
```

The rules in the above example prevent a pipeline from using the `default` namespace unless the repository name matches `cisco/*`.
