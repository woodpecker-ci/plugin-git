package main

import (
	"time"

	"github.com/urfave/cli/v3"
)

var globalFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "remote",
		Usage:   "git remote url",
		Sources: cli.EnvVars("PLUGIN_REMOTE", "CI_REPO_CLONE_URL"),
	},
	&cli.StringFlag{
		Name:    "remote-ssh",
		Usage:   "git clone ssh url",
		Sources: cli.EnvVars("PLUGIN_REMOTE_SSH", "CI_REPO_CLONE_SSH_URL"),
	},
	&cli.StringFlag{
		Name:    "object-format",
		Usage:   "specify the object format (hash) to be used on init. if not set it is autodetect by the commit sha.",
		Sources: cli.EnvVars("PLUGIN_OBJECT_FORMAT"),
	},
	&cli.StringFlag{
		Name:    "path",
		Usage:   "git clone path",
		Sources: cli.EnvVars("PLUGIN_PATH", "CI_WORKSPACE"),
	},
	&cli.StringFlag{
		Name:    "sha",
		Usage:   "git commit sha",
		Sources: cli.EnvVars("PLUGIN_SHA", "CI_COMMIT_SHA"),
	},
	&cli.StringFlag{
		Name:    "ref",
		Usage:   "git commit ref",
		Sources: cli.EnvVars("PLUGIN_REF"),
	},
	&cli.StringFlag{
		Name:    "event",
		Value:   "push",
		Usage:   "pipeline event",
		Sources: cli.EnvVars("CI_PIPELINE_EVENT"),
	},
	&cli.StringFlag{
		Name:    "netrc.machine",
		Usage:   "netrc machine",
		Sources: cli.EnvVars("CI_NETRC_MACHINE"),
	},
	&cli.StringFlag{
		Name:    "netrc.username",
		Usage:   "netrc username",
		Sources: cli.EnvVars("CI_NETRC_USERNAME"),
	},
	&cli.StringFlag{
		Name:    "netrc.password",
		Usage:   "netrc password",
		Sources: cli.EnvVars("CI_NETRC_PASSWORD"),
	},
	&cli.IntFlag{
		Name:    "depth",
		Usage:   "clone depth",
		Sources: cli.EnvVars("PLUGIN_DEPTH"),
	},
	&cli.BoolFlag{
		Name:    "recursive",
		Usage:   "clone submodules",
		Sources: cli.EnvVars("PLUGIN_RECURSIVE"),
		Value:   true,
	},
	&cli.BoolFlag{
		Name:    "tags",
		Usage:   "clone tags, if not explicitly set and event is tag its default is true else false",
		Sources: cli.EnvVars("PLUGIN_TAGS"),
	},
	&cli.BoolFlag{
		Name:    "skip-verify",
		Usage:   "skip tls verification",
		Sources: cli.EnvVars("PLUGIN_SKIP_VERIFY"),
	},
	&cli.StringFlag{
		Name:    "custom-cert",
		Usage:   "path or url to custom cert",
		Sources: cli.EnvVars("PLUGIN_CUSTOM_SSL_PATH", "PLUGIN_CUSTOM_SSL_URL"),
	},
	&cli.BoolFlag{
		Name:    "submodule-update-remote",
		Usage:   "update remote submodules",
		Sources: cli.EnvVars("PLUGIN_SUBMODULES_UPDATE_REMOTE", "PLUGIN_SUBMODULE_UPDATE_REMOTE"),
	},
	&cli.StringFlag{
		Name:    "submodule-override",
		Usage:   "json map of submodule overrides",
		Sources: cli.EnvVars("PLUGIN_SUBMODULE_OVERRIDE"),
	},
	&cli.BoolFlag{
		Name:    "submodule-partial",
		Usage:   "update submodules via partial clone (depth=1) (default)",
		Sources: cli.EnvVars("PLUGIN_SUBMODULES_PARTIAL", "PLUGIN_SUBMODULE_PARTIAL"),
		Value:   true,
	},
	&cli.DurationFlag{
		Name:    "backoff",
		Usage:   "backoff duration",
		Sources: cli.EnvVars("PLUGIN_BACKOFF"),
		Value:   5 * time.Second,
	},
	&cli.IntFlag{
		Name:    "backoff-attempts",
		Usage:   "backoff attempts",
		Sources: cli.EnvVars("PLUGIN_ATTEMPTS"),
		Value:   5,
	},
	&cli.BoolFlag{
		Name:    "lfs",
		Usage:   "whether to retrieve LFS content if available",
		Sources: cli.EnvVars("PLUGIN_LFS"),
		Value:   true,
	},
	&cli.StringFlag{
		Name:  "env-file",
		Usage: "source env file",
	},
	&cli.StringFlag{
		Name:    "branch",
		Usage:   "Change branch name",
		Sources: cli.EnvVars("PLUGIN_BRANCH", "CI_COMMIT_BRANCH", "CI_REPO_DEFAULT_BRANCH"),
	},
	&cli.BoolFlag{
		Name:    "partial",
		Usage:   "Enable/Disable Partial clone",
		Sources: cli.EnvVars("PLUGIN_PARTIAL"),
		Value:   true,
	},
	&cli.StringFlag{
		Name:    "home",
		Usage:   "Change home directory",
		Sources: cli.EnvVars("PLUGIN_HOME"),
	},
	&cli.StringFlag{
		Name:    "safe-directory",
		Usage:   "Define/replace safe directories",
		Sources: cli.EnvVars("PLUGIN_SAFE_DIRECTORY", "CI_WORKSPACE"),
	},
	&cli.BoolFlag{
		Name:    "use-ssh",
		Usage:   "Using ssh for git clone",
		Sources: cli.EnvVars("PLUGIN_USE_SSH"),
		Value:   false,
	},
	&cli.StringFlag{
		Name:    "ssh-key",
		Usage:   "SSH key for ssh clone",
		Sources: cli.EnvVars("PLUGIN_SSH_KEY"),
	},
	&cli.BoolFlag{
		Name:    "merge-pull-request",
		Usage:   "Merge pull reguest with target branch",
		Sources: cli.EnvVars("PLUGIN_MERGE_PULL_REQUEST"),
	},
	&cli.StringFlag{
		Name:    "target-branch",
		Usage:   "Target branch when merging pull request",
		Sources: cli.EnvVars("PLUGIN_TARGET_BRANCH", "CI_COMMIT_TARGET_BRANCH"),
	},
}
