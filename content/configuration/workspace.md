---
date: 2000-01-01T00:00:00+00:00
title: Workspace
author: bradrydzewski
weight: 2
toc: false
description: |
  Describes the pipeline workspace and directory structure.
---

Drone automatically creates a temporary volume, known as your workspace, where it clones your repository. The workspace is the current working directory for each step in your pipeline.

Because the workspace is a volume, filesystem changes are persisted between pipeline steps. In other words, individual steps can communicate and share state using the workspace filesystem.

Workspace path inside your pipeline containers:

```
/drone/src
```

{{< alert "warn" >}}
Note the workspace volume is ephemeral. It is created when the pipeline starts and destroyed after the pipeline completes.
{{< / alert >}}
