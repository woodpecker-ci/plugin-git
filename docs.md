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
- Adjust submodules.
- Sparse checkout for monorepos.

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

| Settings Name             | Default                             | Description                                                                                                                                                                |
| ------------------------- | ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `depth`                   | _none_                              | If specified, uses git's `--depth` option to create a shallow clone with a limited number of commits, overwritten by `partial`. Setting it to `0` disables shallow cloning |
| `lfs`                     | `true`                              | Set this to `false` to disable retrieval of LFS files                                                                                                                      |
| `recursive`               | `true`                              | Clones submodules                                                                                                                                                          |
| `skip-verify`             | `false`                             | Skips the SSL verification                                                                                                                                                 |
| `tags`                    | `false` (except on tag event)       | Fetches tags when set to true, default is false if event is not tag else true                                                                                              |
| `submodule-override`      | _none_                              | Override submodule urls                                                                                                                                                    |
| `submodule-update-remote` | `false`                             | Pass the --remote flag to git submodule update                                                                                                                             |
| `submodule-partial`       | `true`                              | Update submodules via partial clone (depth=1)                                                                                                                              |
| `custom-ssl-path`         | _none_                              | Set path to custom cert                                                                                                                                                    |
| `custom-ssl-url`          | _none_                              | Set url to custom cert                                                                                                                                                     |
| `backoff`                 | `5sec`                              | Change backoff duration                                                                                                                                                    |
| `attempts`                | `5`                                 | Change backoff attempts                                                                                                                                                    |
| `branch`                  | $CI_COMMIT_BRANCH                   | Change branch name to checkout to                                                                                                                                          |
| `partial`                 | `true` (except if tags are fetched) | Only fetch the one commit and it's blob objects to resolve all files, overwrite depth with 1                                                                               |
| `sparse`                  | _none_                              | Materialize only selected directories using Git sparse checkout                                                                                                           |
| `home`                    |                                     | Change HOME var for commands executed, fail if it does not exist                                                                                                           |
| `remote`                  | $CI_REPO_CLONE_URL                  | Set the git remote url                                                                                                                                                     |
| `remote-ssh`              | $CI_REPO_CLONE_SSH_URL              | Set the git SSH remote url                                                                                                                                                 |
| `object-format`           | detected from commit SHA            | Set the object format for Git initialization. Supported values: `sha1`, `sha256`.                                                                                          |
| `sha`                     | $CI_COMMIT_SHA                      | git commit hash to retrieve                                                                                                                                                |
| `ref`                     | _none_                              | Set the git reference to retrieve                                                                                                                                          |
| `path`                    | $CI_WORKSPACE                       | Set destination path to clone to                                                                                                                                           |
| `use-ssh`                 | `false`                             | Clone using SSH                                                                                                                                                            |
| `ssh-key`                 | _none_                              | path to SSH key for SSH clone                                                                                                                                              |
| `ssh-key-private`         | _none_                              | the content of the private key file. Is written into `ssh-key`                                                                                                             |
| `ssh-host-key`            | _none_                              | SSH host key for verification to prevent man-in-the-middle attacks. If no key is given here, verification is disabled. You can get it using `ssh-keyscan -t rsa <url>`     |
| `merge-pull-request`      | `false`                             | merge the pull request with the target branch (can fail on merge conflict)                                                                                                 |
| `fetch-target-branch`     | `false`                             | fetch the target branch without merging (useful for tools like nx affected that need both branches locally)                                                                |
| `target-branch`           | $CI_COMMIT_TARGET_BRANCH            | Target branch used when merging pull requests (`merge-pull-request`) or when fetching the target branch (`fetch-target-branch`)                                            |
| `git-user-name`           | _none_                              | Git username used when pull requests are used.                                                                                                                             |
| `git-user-email`          | _none_                              | Git email used when pull requests are used.                                                                                                                                |

## Sparse checkout

Sparse checkout limits which tracked paths Git materializes in the working tree. It is independent of partial clone: `partial` controls which Git objects are fetched in advance, while `sparse` controls which files appear in the workspace. The two settings can be used together or independently.

Sparse checkout uses Git cone mode and accepts repository-relative directories:

```yaml
clone:
  git:
    image: woodpeckerci/plugin-git
    settings:
      sparse:
        - src/backend
        - shared
```

Cone mode includes files directly in the repository root, such as `README.md`, and files directly in ancestor directories. Exact file and gitignore-style pattern selection are not supported.

The plugin requires Git 2.35 or newer. Cached workspaces are safe to reuse: every run replaces the previous sparse selection, and removing `sparse` disables a previously active sparse checkout and restores a full working tree.

`partial: true` remains the default and provides the largest network saving when combined with sparse checkout. With `partial: false`, the working tree is still sparse but Git may fetch all repository objects. Fetching tags disables partial clone as before, but does not disable sparse checkout.

When LFS is enabled, sparse checkout downloads LFS content while materializing selected files instead of performing a repository-wide LFS fetch. With `lfs: false`, selected LFS files remain pointer files. Submodule behavior is unchanged: the default `recursive: true` can initialize submodules outside the sparse selection. Use `recursive: false` when minimizing materialized content is more important than automatic submodule initialization.

Woodpecker transports primitive list settings as comma-separated environment values. Consequently, a sparse path or pattern containing a comma cannot be represented unambiguously through YAML settings.

[workflowClone]: https://woodpecker-ci.org/docs/usage/workflow-syntax#clone
