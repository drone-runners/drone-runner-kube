#!/bin/sh

set -e
set -x

docker buildx build --file docker/placeholder/Dockerfile --platform linux/amd64,linux/arm64,linux/arm/v6,linux/arm/v7 -t drone/placeholder:latest --push .
