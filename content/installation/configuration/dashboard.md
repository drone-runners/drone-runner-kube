---
date: 2000-01-01T00:00:00+00:00
title: Setup the Dashboard
title_in_header: Dashboard
author: bradrydzewski
weight: 2

description: |
  Configure the administrative dashboard.
---

The Docker runner features a user interface (web interface) for read only access to recent runtime information and log activity. The user interface simplifies troubleshooting and access to system information.

![runner dashboard](../../../screenshots/runner_dashboard.png)

_The above screenshot demonstrates information available for troubleshooting pipeline execution. This includes access to system logs associated with the pipeline._

# Activation

The user interface is disabled by default. To enable the user interface you must configure the runner with a username and password.

```
DRONE_UI_USERNAME=root
DRONE_UI_PASSWORD=root
```

# Access

The user interface can be accessed at the default port (3000) using the host machine IP address or hostname as the address. The browser will automatically prompt you for the username and password. _Please note you may need to configure firewall rules to access the user interface._
