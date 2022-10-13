# Changelog

## [v1.6.0](https://github.com/woodpecker-ci/plugin-git/releases/tag/v1.6.0) - 2022-10-13

* BUGFIXES
  * Handle git-lfs separately (#40)
* ENHANCEMENTS
  * if no branch info is set, fallback to default repo branch (#41)

## [v1.5.0](https://github.com/woodpecker-ci/plugin-git/releases/tag/v1.5.0) - 2022-10-06

* ENHANCEMENTS
  * Release binarys (#37)
  * Use ref to checkout if no commit sha is set (#36)
  * Fix tests (#35)
* MISC
  * Update urfave/cli to v2.17.1 (#38)
  * Use built-in log instead of logrus (#34)

## [v1.4.0](https://github.com/woodpecker-ci/plugin-git/releases/tag/v1.4.0) - 2022-08-30

* ENHANCEMENTS
  * Auto enable tags clone if it's ci event is 'tag' (#30)
  * Support more architectures (#29)

## [v1.3.0](https://github.com/woodpecker-ci/plugin-git/releases/tag/v1.3.0) - 2022-08-15

* FEATURES
  * Add option to Change branch name for checkout (#28)

## [v1.2.0](https://github.com/woodpecker-ci/plugin-git/releases/tag/v1.2.0) - 2022-05-25

* FEATURES
  * Add git-lfs (#21)
  * Custom ssl certs for git (#19)
* ENHANCEMENTS
  * Add an `lfs` setting which lets you disable Git LFS (#24)
* DOCUMENTATION
  * Add docs page (#23)

## [v1.1.2](https://github.com/woodpecker-ci/plugin-git/releases/tag/v1.1.2) - 2022-01-30

* BUGFIXES
  * Fix empty login/password in netrc (#20)

## [v1.1.1](https://github.com/woodpecker-ci/plugin-git/releases/tag/v1.1.1) - 2021-12-23

* BUGFIXES
  * Fix version info (#13)

## [v1.1.0](https://github.com/woodpecker-ci/plugin-git/releases/tag/v1.1.0) - 2021-12-18

* FEATURES
  * Add ppc64le support (#8)
* BUGFIXES
  * Regognize "CI_*" EnvVars (#6)
* ENHANCEMENTS
  * Multiarch build (#8)
* MISC
  * Upgrade urfave/cli to v2 (#5)
