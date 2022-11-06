---
name: Git Plugin
icon: https://woodpecker-ci.org/img/logo.svg

description: This is the default plugin for the clone step.
authors: Woodpecker Authors
tags: [git, clone]
containerImage: woodpeckerci/plugin-git
containerImageUrl: https://hub.docker.com/r/woodpeckerci/plugin-git
url: https://github.com/woodpecker-ci/plugin-git
---

This plugin is automatically introduced into your pipeline as the first step.
Its purpose is to clone your Git repository.

## Features

- Git LFS support is enabled by default.
- Fetch tags when needed.
- Ajust submodules.

## Overriding Settings

You can manually define your `clone` step in order to change plugin or override some of the default settings.
Consult [the `clone` section of the pipeline documentation][pipelineClone] for more information;
this documentation page only describes this plugin.

```yaml
clone:
  git:
    image: woodpeckerci/plugin-git
    settings:
      depth: 50
      lfs: false
```

## Settings

| Settings Name             | Default           | Description
| --------------------------| ----------------- | --------------------------------------------
| `depth`                   | *none*            | If specified, uses git's `--depth` option to create a shallow clone with a limited number of commits
| `lfs`                     | `true`            | Set this to `false` to disable retrieval of LFS files
| `recursive`               | `false`           | Clones submodules recursively
| `skip_verify`             | `false`           | Skips the SSL verification
| `tags`                    | `false` (except on tag event) | Fetches tags when set to true, default is false if event is not tag else true
| `submodule_overrides`     | *none*            | Override submodule urls
| `submodule_update_remote` | `false`           | Pass the --remote flag to git submodule update
| `custom_ssl_path`         | *none*            | Set path to custom cert
| `custom_ssl_url`          | *none*            | Set url to custom cert
| `backoff`                 | `5sec`            | Change backoff duration
| `attempts`                | `5`               | Change backoff attempts
| `branch`                  | $CI_COMMIT_BRANCH | Change branch name to checkout to
| `home`                    |                   | Change HOME var for commands executed, fail if it does not exist

[pipelineClone]: https://woodpecker-ci.org/docs/usage/pipeline-syntax#clone
