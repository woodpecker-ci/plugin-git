---
name: Git Plugin
icon: https://woodpecker-ci.org/img/logo.svg
description: This is the default plugin for the clone step.
---

This plugin is automatically introduced into your pipeline as the first step.
Its purpose is to clone your Git repository.

## Features

- Git LFS support is enabled by default.


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

| Settings Name | Default | Description |
| ------------- | ------: | ----------- |
| `depth`       |  *none* | If specified, uses git's `--depth` option to create a shallow clone with a limited number of commits. |
| `lfs`         |  `true` | Set this to `false` to disable retrieval of LFS files. |


[pipelineClone]: https://woodpecker-ci.org/docs/usage/pipeline-syntax#clone

