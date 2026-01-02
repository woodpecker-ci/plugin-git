# plugin-git

<p align="center">
  <a href="https://ci.woodpecker-ci.org/repos/5586" title="Build Status">
    <img src="https://ci.woodpecker-ci.org/api/badges/5586/status.svg" alt="Build Status">
  </a>
  <a href="https://discord.gg/fcMQqSMXJy" title="Discord chat">
    <img src="https://img.shields.io/discord/838698813463724034.svg" alt="Discord chat">
  </a>
  <a href="https://goreportcard.com/report/github.com/woodpecker-ci/plugin-git" title="Go Report Card">
    <img src="https://goreportcard.com/badge/github.com/woodpecker-ci/plugin-git" alt="Go Report Card">
  </a>
  <a href="https://godoc.org/github.com/woodpecker-ci/plugin-git" title="GoDoc">
    <img src="https://godoc.org/github.com/woodpecker-ci/plugin-git?status.svg" alt="GoDoc">
  </a>
  <a href="https://hub.docker.com/r/woodpeckerci/plugin-git" title="Docker pulls">
    <img src="https://img.shields.io/docker/pulls/woodpeckerci/plugin-git" alt="Docker pulls">
  </a>
  <a href="https://opensource.org/licenses/Apache-2.0" title="License: Apache-2.0">
    <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License: Apache-2.0">
  </a>
</p>

Woodpecker plugin to clone `git` repositories. For the usage information and a listing of the available options please take a look at [the docs](https://woodpecker-ci.org/plugins/git-clone).
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
docker buildx build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --platform linux/amd64 --output type=docker \
  --file docker/Dockerfile.multiarch --tag woodpeckerci/plugin-git .
```

*The platform linux/amd64 should be replaced by the correct platform.*

This will build the image and load it into docker so the image can be used locally.
[More information on the output formats can be found in docker buildx doc](https://docs.docker.com/engine/reference/commandline/buildx_build/#output).

## Usage

Clone a commit:

```console
docker run --rm \
  -e CI_REPO_CLONE_URL=https://github.com/garyburd/redigo.git \
  -e CI_WORKSPACE=/go/src/github.com/garyburd/redigo \
  -e CI_PIPELINE_EVENT=push \
  -e CI_COMMIT_SHA=d8dbe4d94f15fe89232e0402c6e8a0ddf21af3ab \
  -e PLUGIN_REF=refs/heads/master \
  woodpeckerci/plugin-git
```

Clone a pull request:

```console
docker run --rm \
  -e CI_REPO_CLONE_URL=https://github.com/garyburd/redigo.git \
  -e CI_WORKSPACE=/go/src/github.com/garyburd/redigo \
  -e CI_PIPELINE_EVENT=pull_request \
  -e CI_COMMIT_SHA=3b4642018d177bf5fecc5907e7f341a2b5c12b8a \
  -e PLUGIN_REF=refs/pull/74/head \
  woodpeckerci/plugin-git
```

Clone a pull request and attempt merging it with the target branch:

```console
docker run --rm \
  -e CI_PIPELINE_EVENT=pull_request \
  -e CI_REPO_CLONE_URL=https://codeberg.org/johanvdw/test-git-plugin.git \
  -e CI_COMMIT_SHA=d02eaf69b920b19fd7b14ad3aee622dd97413fbc \
  -e CI_COMMIT_TARGET_BRANCH=main \
  -e PLUGIN_REF=refs/pull/1/head \
  -e PLUGIN_MERGE_PULL_REQUEST=true \
  -e PLUGIN_GIT_USERNAME=ci \
  -e PLUGIN_GIT_USEREMAIL=ci@mydomain.ci \
   release/plugin-git
```

Clone a tag:

```console
docker run --rm \
  -e CI_REPO_CLONE_URL=https://github.com/garyburd/redigo.git \
  -e CI_WORKSPACE=/go/src/github.com/garyburd/redigo \
  -e CI_PIPELINE_EVENT=tag \
  -e CI_COMMIT_SHA=3b4642018d177bf5fecc5907e7f341a2b5c12b8a \
  -e PLUGIN_REF=refs/tags/74/head \
  woodpeckerci/plugin-git
```

## Build arguments

### HOME

The docker image can be build using `--build-arg HOME=<custom home>`.
This will create the directory for the custom home and set the custom home as the default value for the `home` plugin setting (see [the plugin docs](./docs.md) for more information about this setting).
