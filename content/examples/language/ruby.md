---
date: 2000-01-01T00:00:00+00:00
title: Ruby
title_in_header: Example Ruby Pipeline
author: bradrydzewski
weight: 1
toc: true
---

This guide covers configuring continuous integration pipelines for Ruby projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that executes `bundle install` and `rake` commands. These commands are executed inside the ruby Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: test
  image: ruby
  commands:
  - bundle install --jobs=3 --retry=3
  - rake
```

Please note that you can use any Docker image in your pipeline from any Docker registry. You can use the official ruby [images](https://hub.docker.com/r/_/ruby/), or your can bring your own.

# Dependencies

If you decide to split your pipeline into multiple steps you need to make sure each step has access to project dependencies. Dependencies are downloaded to `/usr/local/bundle` which is outside of the shared workspace. Create a named volume to share this directory with all pipeline steps:

```
kind: pipeline
type: kubernetes
name: default

steps:
- name: install
  image: ruby
  volumes:
  - name: bundle
    path: /usr/local/bundle
  commands:
  - bundle install --jobs=3 --retry=3

- name: test
  image: ruby
  volumes:
  - name: bundle
    path: /usr/local/bundle
  commands:
  - rake

volumes:
- name: bundle
  temp: {}
```

# Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of Ruby. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
type: kubernetes
name: ruby-2-4

steps:
- name: test
  image: ruby:2.4
  commands:
  - bundle install --jobs=3 --retry=3
  - rake

---
kind: pipeline
type: kubernetes
name: ruby-2-3

steps:
- name: test
  image: ruby:2.3
  commands:
  - bundle install --jobs=3 --retry=3
  - rake

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.

```
local Pipeline(name, image) = {
  kind: "pipeline",
  type: "kubernetes",
  name: name,
  steps: [
    {
      name: "test",
      image: image,
      commands: [
        "bundle install --jobs=3 --retry=3",
        "rake"
      ]
    }
  ]
};

[
  Pipeline("ruby23", "ruby:2.3"),
  Pipeline("ruby24", "ruby:2.4"),
]
```

# Test Multiple Architectures

You can use Drone's multi-pipeline feature to concurrently test your code on multiple architectures and operating systems.

```
---
kind: pipeline
type: kubernetes
name: test-on-amd64

platform:
  arch: amd64

steps:
- name: test
  image: ruby
  commands:
  - bundle install --jobs=3 --retry=3
  - rake

---
kind: pipeline
type: kubernetes
name: test-on-arm64

platform:
  arch: arm64

steps:
- name: test
  image: ruby
  commands:
  - bundle install --jobs=3 --retry=3
  - rake

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide.
