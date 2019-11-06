---
date: 2000-01-01T00:00:00+00:00
title: Vault
title_in_header: Example Vault Configuration
author: bradrydzewski
weight: 1
toc: false
source: https://github.com/drone-demos/drone-demo-vault
---

This guide covers configuring continuous integration pipelines for projects that have a vault dependency. If you're new to Drone please read our Tutorial and build configuration guides first.

# Example

In the below example we demonstrate a pipeline that launches a vault service container. The vault server will be available at `vault:8200`, where the hostname matches the service container name.

```
kind: pipeline
step: default

steps:
- name: test
  image: vault:1.0.0-beta2
  environment:
    VAULT_ADDR: http://vault:8200
    VAULT_TOKEN: dummy
  commands:
  - sleep 5
  - vault kv put secret/my-secret my-value=s3cr3t
  - vault kv get secret/my-secret

services:
- name: vault
  image: vault:1.0.0-beta2
  environment:
    VAULT_DEV_ROOT_TOKEN_ID: dummy
  ports:
  - 8200
```