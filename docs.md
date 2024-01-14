---
name: Git Clone
icon: https://raw.githubusercontent.com/woodpecker-ci/plugin-git/main/git.svg
description: This is the default plugin for the clone step.
author: Woodpecker Authors
tags: [git, clone]
containerImage: woodpeckerci/plugin-git
containerImageUrl: https://hub.docker.com/r/woodpeckerci/plugin-git
url: https://github.com/woodpecker-ci/plugin-git
---

# plugin-git

This plugin is automatically introduced into your pipeline as the first step.
Its purpose is to clone your Git repository.

## Features

- Git LFS support is enabled by default.
- Fetch tags when needed.
- Ajust submodules.

## Overriding Settings

You can manually define your `clone` step in order to change plugin or override some of the default settings.
Consult [the `clone` section of the pipeline documentation][workflowClone] for more information;
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

| Settings Name             | Default                             | Description                                                                                                                             |
| ------------------------- | ----------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `depth`                   | _none_                              | If specified, uses git's `--depth` option to create a shallow clone with a limited number of commits, overwritten by partial            |
| `lfs`                     | `true`                              | Set this to `false` to disable retrieval of LFS files                                                                                   |
| `recursive`               | `false`                             | Clones submodules recursively                                                                                                           |
| `skip_verify`             | `false`                             | Skips the SSL verification                                                                                                              |
| `tags`                    | `false` (except on tag event)       | Fetches tags when set to true, default is false if event is not tag else true                                                           |
| `submodule_overrides`     | _none_                              | Override submodule urls                                                                                                                 |
| `submodule_update_remote` | `false`                             | Pass the --remote flag to git submodule update                                                                                          |
| `submodule_partial`       | `true`                              | Update submodules via partial clone (depth=1)                                                                                           |
| `custom_ssl_path`         | _none_                              | Set path to custom cert                                                                                                                 |
| `custom_ssl_url`          | _none_                              | Set url to custom cert                                                                                                                  |
| `backoff`                 | `5sec`                              | Change backoff duration                                                                                                                 |
| `attempts`                | `5`                                 | Change backoff attempts                                                                                                                 |
| `branch`                  | $CI_COMMIT_BRANCH                   | Change branch name to checkout to                                                                                                       |
| `partial`                 | `true` (except if tags are fetched) | Only fetch the one commit and it's blob objects to resolve all files, overwrite depth with 1                                            |
| `home`                    |                                     | Change HOME var for commands executed, fail if it does not exist                                                                        |
| `remote`                  | $CI_REPO_CLONE_URL                  | Set the git remote url                                                                                                                  |
| `remote_ssh`              | $CI_REPO_CLONE_SSH_URL              | Set the git SSH remote url                                                                                                              |
| `sha`                     | $CI_COMMIT_SHA                      | git commit hash to retrieve (use `sha: ''` to clone the `ref`)                                                                          |
| `ref`                     | $CI_COMMIT_REF                      | Set the git reference to retrieve (use `ref: refs/heads/a_branch` and `sha: ''` to retrieve the head commit from the "a_branch" branch) |
| `path`                    | $CI_WORKSPACE                       | Set destination path to clone to                                                                                                        |
| `use_ssh`                 | `false`                             | Clone using SSH                                                                                                                         |
| `ssh_key`                 | _none_                              | SSH key for SSH clone                                                                                                                   |

[workflowClone]: https://woodpecker-ci.org/docs/usage/workflow-syntax#clone
