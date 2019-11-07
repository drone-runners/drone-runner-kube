---
date: 2000-01-01T00:00:00+00:00
title: Conditions
author: bradrydzewski
weight: 7
toc: true
description: |
  Configure pipeline steps to execute conditionally.
---

Conditions can be used to limit pipeline step execution at runtime. For example, you may want to limit step execution by branch:

{{< highlight text "linenos=table,hl_lines=11-14" >}}
kind: pipeline
type: kubernetes
name: default

steps:
- name: build
  image: golang
  commands:
  - go build
  - go test
  when:
    branch:
    - master
    - feature/*
{{< / highlight >}}

You can use wildcard matching in your conditions. _Note that conditions use [glob](https://golang.org/pkg/path/#Match) pattern matching, not regular expressions._

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  ref:
  - refs/heads/master
  - refs/heads/**
  - refs/pull/*/head
{{< / highlight >}}

You can also combine multiple conditions. _Note that all conditions must evaluate to true when combining multiple conditions._

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  branch:
  - master
  event:
  - push
{{< / highlight >}}

# By Branch

The branch condition limits step execution based on the git branch. Please note that the target branch is evaluated for pull requests; and branch names are not available for tag events.

{{< alert "warn" >}}
Note that you cannot use branch conditions with tags. A tag is not associated with the source branch from which it was created.
{{< / alert >}}

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  branch:
  - master
  - feature/*
{{< / highlight >}}

Example include syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  branch:
    include:
    - master
    - feature/*
{{< / highlight >}}

Example exclude syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  branch:
    exclude:
    - master
    - feature/*
{{< / highlight >}}

# By Event

The event condition limits step execution based on the drone event type. This can be helpful when you want to limit steps based on push, pull request, tag and more.

{{< alert "warn" >}}
Note that you cannot use branch conditions with tag events. A tag is not associated with the source branch from which it was created.
{{< / alert >}}

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  event:
  - push
  - pull_request
  - tag
  - promote
  - rollback
{{< / highlight >}}

Example include syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  event:
    include:
    - push
    - pull_request
{{< / highlight >}}

Example exclude syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  event:
    exclude:
    - pull_request
{{< / highlight >}}

# By Reference

The reference condition limits step execution based on the git reference name. This can be helpful when you want to glob match branch or tag names.

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  ref:
  - refs/heads/feature-*
  - refs/tags/*
{{< / highlight >}}

Example include syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  ref:
    include:
    - refs/heads/feature-*
    - refs/pull/**
    - refs/tags/**
{{< / highlight >}}

Example exclude syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  ref:
    exclude:
    - refs/heads/feature-*
    - refs/pull/**
    - refs/tags/**
{{< / highlight >}}

# By Repository

The repository condition limits step execution based on repository name. This can be useful when Drone is enabled for a repository and its forks, and you want to limit execution accordingly.


{{< highlight text "linenos=table,linenostart=11" >}}
when:
  repo:
  - octocat/hello-world
{{< / highlight >}}

Example include syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  repo:
    include:
    - octocat/hello-world
    - spacebhost/hello-world
{{< / highlight >}}

Example exclude syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  repo:
    exclude:
    - octocat/hello-world
    - spacebhost/hello-world
{{< / highlight >}}

Example using wildcard matching:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  repo:
    include:
    - octocat/*
{{< / highlight >}}

# By Instance

The instance condition limits step execution based on the Drone instance hostname. This can be useful if you have multiple Drone instances configured for a single repository, sharing the same yaml file, and want to limit steps by instance.

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  instance:
  - drone.instance1.com
  - drone.instance2.com
{{< / highlight >}}

Example include syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  instance:
    include:
    - drone.instance1.com
    - drone.instance2.com
{{< / highlight >}}

Example exclude syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  instance:
    exclude:
    - drone.instance1.com
    - drone.instance2.com
{{< / highlight >}}

Example using wildcard matching:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  instance:
    include:
    - *.company.com
{{< / highlight >}}

# By Status

The status condition limits step execution based on the pipeline status. For example, you may want to configure Slack notification only on failure.

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  status:
  - failure
{{< / highlight >}}

Execute a step on failure:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  status:
  - failure
{{< / highlight >}}

Execute a step on success or failure:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  status:
  - success
  - failure
{{< / highlight >}}

The following configuration is redundant. The default behavior is for pipeline steps to only execute when the pipeline is in a passing state.

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  status:
  - success
{{< / highlight >}}

# By Target

The target condition limits step execution based on the target deployment environment. This only applies to promotion and rollback events.

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  target:
  - production
{{< / highlight >}}

Example include syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  target:
    include:
    - staging
    - production
{{< / highlight >}}

Example exclude syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  target:
    exclude:
    - production
{{< / highlight >}}


# By Cron

The cron condition limits step execution based on the cron name that triggered the pipeline. This only applies to cron events.

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  event:
  - cron
  cron:
  - nightly
{{< / highlight >}}

Example include syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  event:
  - cron
  target:
    include:
    - weekly
    - nightly
{{< / highlight >}}

Example exclude syntax:

{{< highlight text "linenos=table,linenostart=11" >}}
when:
  event:
  - cron
  target:
    exclude:
    - nightly
{{< / highlight >}}