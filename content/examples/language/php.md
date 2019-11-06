---
date: 2000-01-01T00:00:00+00:00
title: PHP
title_in_header: Example PHP Pipeline
author: bradrydzewski
weight: 1
---

This guide covers configuring continuous integration pipelines for PHP projects. If you're new to Drone please read our Tutorial and build configuration guides first.

# Build and Test

In the below example we demonstrate a pipeline that installs the project dependnecies using composer, and then executes the project unit tests. These commands are executed inside a Docker container, downloaded at runtime from DockerHub.

```
kind: pipeline
name: default

steps:
- name: install
  image: composer
  commands:
  - composer install

- name: test
  image: php:7
  commands:
  - vendor/bin/phpunit --configuration config.xml
```

This example assumes phpunit is a dev dependency in `composer.json`

```
{
    "require-dev": {
        "phpunit/phpunit": "3.7.*"
    }
}
```

Please note you can use any Docker image in your pipeline from any Docker registry. You can use the official [php](https://hub.docker.com/r/_/php/) or [composer](https://hub.docker.com/r/_/composer/) images, or your can bring your own.



<!-- $HOME/.composer -->

<!-- 
## Test Multiple Versions

You can use Drone's multi-pipeline feature to concurrently test against multiple versions of PHP. This is equivalent to matrix capabilities found in other continuous integration systems.

```
---
kind: pipeline
name: php5

steps:
- name: test
  image: php:5
  commands:
  - composer install
  - phpunit

---
kind: pipeline
name: php6

steps:
- name: test
  image: php:6
  commands:
  - composer install
  - phpunit

...
```

If you find this syntax too verbose we recommend using jsonnet. If you are unfamiliar with jsonnet please read our guide. -->
