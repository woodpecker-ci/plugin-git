# plugin-git

<p align="center">
  <a href="https://wp.laszlo.cloud/woodpecker-ci/plugin-git" title="Build Status">
    <img src="https://wp.laszlo.cloud/api/badges/woodpecker-ci/plugin-git/status.svg">
  </a>
  <a href="https://discord.gg/fcMQqSMXJy" title="Join the Discord chat at https://discord.gg/fcMQqSMXJy">
    <img src="https://img.shields.io/discord/838698813463724034.svg">
  </a>
  <a href="https://goreportcard.com/badge/github.com/woodpecker-ci/plugin-git" title="Go Report Card">
    <img src="https://goreportcard.com/badge/github.com/woodpecker-ci/plugin-git">
  </a>
  <a href="https://godoc.org/github.com/woodpecker-ci/plugin-git" title="GoDoc">
    <img src="https://godoc.org/github.com/woodpecker-ci/plugin-git?status.svg">
  </a>
  <a href="https://hub.docker.com/r/woodpeckerci/plugin-git" title="Docker pulls">
    <img src="https://img.shields.io/docker/pulls/woodpeckerci/plugin-git">
  </a>
  <a href="https://opensource.org/licenses/Apache-2.0" title="License: Apache-2.0">
    <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg">
  </a>
</p>

Woodpecker plugin to clone `git` repositories. For the usage information and a listing of the available options please take a look at [the docs](https://woodpecker-ci.org/plugins/plugin-git).
The docs are also available in [`docs.md` in this repository](docs.md).

## Build

Build the binary with the following command:

```console
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GO111MODULE=on

go build -v -a -tags netgo -o release/linux/amd64/plugin-git
```

## Docker

Build the Docker image with the following command:

```console
docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file docker/Dockerfile.linux.amd64 --tag woodpeckerci/plugin-git .
```

## Usage

Clone a commit:

```console
docker run --rm \
  -e CI_REPO_REMOTE=https://github.com/garyburd/redigo.git \
  -e CI_WORKSPACE=/go/src/github.com/garyburd/redigo \
  -e CI_BUILD_EVENT=push \
  -e CI_COMMIT_SHA=d8dbe4d94f15fe89232e0402c6e8a0ddf21af3ab \
  -e CI_COMMIT_REF=refs/heads/master \
  woodpeckerci/plugin-git
```

Clone a pull request:

```console
docker run --rm \
  -e CI_REPO_REMOTE=https://github.com/garyburd/redigo.git \
  -e CI_WORKSPACE=/go/src/github.com/garyburd/redigo \
  -e CI_BUILD_EVENT=pull_request \
  -e CI_COMMIT_SHA=3b4642018d177bf5fecc5907e7f341a2b5c12b8a \
  -e CI_COMMIT_REF=refs/pull/74/head \
  woodpeckerci/plugin-git
```

Clone a tag:

```console
docker run --rm \
  -e CI_REPO_REMOTE=https://github.com/garyburd/redigo.git \
  -e CI_WORKSPACE=/go/src/github.com/garyburd/redigo \
  -e CI_BUILD_EVENT=tag \
  -e CI_COMMIT_SHA=3b4642018d177bf5fecc5907e7f341a2b5c12b8a \
  -e CI_COMMIT_REF=refs/tags/74/head \
  woodpeckerci/plugin-git
```
