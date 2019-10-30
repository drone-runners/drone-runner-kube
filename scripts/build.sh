#!/bin/sh

# disable go modules
export GOPATH=""

# disable cgo
export CGO_ENABLED=0

set -e
set -x

# linux
GOOS=linux GOARCH=amd64 go build -o release/linux/amd64/drone-runner-kube cmd/runner/main.go
GOOS=linux GOARCH=arm64 go build -o release/linux/arm64/drone-runner-kube cmd/runner/main.go
GOOS=linux GOARCH=arm   go build -o release/linux/arm/drone-runner-kube cmd/runner/main.go

GOOS=linux GOARCH=amd64 go build -o release/linux/amd64/placeholder cmd/placeholder/main.go
GOOS=linux GOARCH=arm64 go build -o release/linux/arm64/placeholder cmd/placeholder/main.go
GOOS=linux GOARCH=arm   go build -o release/linux/arm/placeholder cmd/placeholder/main.go
