---
date: 2000-01-01T00:00:00+00:00
title: Setup Logging
title_in_header: Setup Logging
author: bradrydzewski
weight: 1

description: |
  Configure trace and debug logging.
---

The Kubernetes runner writes logs to stdout and stderr. You can access the logs using the web [Dashboard]({{< relref "dashboard.md" >}}) or the docker command line utility.

```
$ docker logs <container name>
```

The Kubernetes runner is configured to log runtime events. You can enable debug or trace level logs to get detailed information on the flow through the system.

```
DRONE_DEBUG=true
DRONE_TRACE=true
```

You can enable http request logging to get detailed http communication details between the runner and the server. _This generates significant output and should only be enabled to troubleshoot communication issues._

```
DRONE_RPC_DUMP_HTTP=true
DRONE_RPC_DUMP_HTTP_BODY=true
```