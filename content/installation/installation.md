---
date: 2000-01-01T00:00:00+00:00
title: Installation Guide
author: bradrydzewski
weight: 1
toc: true
description: |
  Install the runner with the Docker image.
---

This article explains how to install the Kubernetes runner on Linux. The Kubernetes runner is packaged as a minimal Docker image distributed on [DockerHub](https://hub.docker.com/r/drone/drone-runner-kube).

# Step 1 - Configure

The Kubernetes runner is configured using environment variables. This article references the below configuration options. See [Configuration]({{< relref "reference" >}}) for a complete list of configuration options.

- __DRONE_RPC_HOST__
  : provides the hostname (and optional port) of your Drone server. The runner connects to the server at the host address to receive pipelines for execution.
- __DRONE_RPC_PROTO__
  : provides the protocol used to connect to your Drone server. The value must be either http or https.
- __DRONE_RPC_SECRET__
  : provides the shared secret used to authenticate with your Drone server. This must match the secret defined in your Drone server configuration.

# Step 2 - Authenticate

The Kubernetes runner uses in-cluster authentication to communicate with the Kubernetes API. Please ensure the Kubernetes runner is associated with a service account when deployed to your cluster.

# Step 3 - Install

The following is a rudimentary manifest file used to configure and install the Kubernetes runner. _Remember to replace the environment variables below with the correct values._

Here are the sample rbac rules:

{{< highlight text "linenos=table" >}}
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: default
  name: drone
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  verbs:
  - get
  - create
  - delete
  - list
  - watch
  - update

---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: drone
  namespace: default
subjects:
- kind: ServiceAccount
  name: default
  namespace: default
roleRef:
  kind: Role
  name: drone
  apiGroup: rbac.authorization.k8s.io
{{< / highlight >}}

And here is the example deployment:

{{< highlight text "linenos=table,hl_lines=24 26 28" >}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: drone
  labels:
    app.kubernetes.io/name: drone
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: drone
  template:
    metadata:
      labels:
        app.kubernetes.io/name: drone
    spec:
      containers:
      - name: runner
        image: drone/drone-runner-kube:latest
        ports:
        - containerPort: 3000
        env:
        - name: DRONE_RPC_HOST
          value: drone.company.com
        - name: DRONE_RPC_PROTO
          value: http
        - name: DRONE_RPC_SECRET
          value: super-duper-secret
{{< / highlight >}}

# Step 4 - Verify

Use the `kubectl logs drone -c runner` command to view the logs and verify the runner successfully established a connection with the Drone server.

```
$ kubectl logs drone -c runner

INFO[0000] starting the server
INFO[0000] successfully pinged the remote server 
```

# Migration Issues

If you are migrating from the legacy Kubernetes runtime to the new Kubernetes runner please unset the following server environment variables:

```
DRONE_AGENTS_DISABLED
DRONE_KUBERNETES_ENABLED
```