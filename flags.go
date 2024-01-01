package main

import (
	"time"

	"github.com/urfave/cli/v2"
)

var globalFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "remote",
		Usage:   "git remote url",
		EnvVars: []string{"PLUGIN_REMOTE", "CI_REPO_CLONE_URL"},
	},
	&cli.StringFlag{
		Name:    "remote-ssh",
		Usage:   "git clone ssh url",
		EnvVars: []string{"PLUGIN_REMOTE_SSH", "CI_REPO_CLONE_SSH_URL"},
	},
	&cli.StringFlag{
		Name:    "path",
		Usage:   "git clone path",
		EnvVars: []string{"PLUGIN_PATH", "CI_WORKSPACE"},
	},
	&cli.StringFlag{
		Name:    "sha",
		Usage:   "git commit sha",
		EnvVars: []string{"CI_COMMIT_SHA"},
	},
	&cli.StringFlag{
		Name:    "ref",
		Value:   "refs/heads/main",
		Usage:   "git commit ref",
		EnvVars: []string{"PLUGIN_REF"},
	},
	&cli.StringFlag{
		Name:    "event",
		Value:   "push",
		Usage:   "pipeline event",
		EnvVars: []string{"CI_PIPELINE_EVENT"},
	},
	&cli.StringFlag{
		Name:    "netrc.machine",
		Usage:   "netrc machine",
		EnvVars: []string{"CI_NETRC_MACHINE"},
	},
	&cli.StringFlag{
		Name:    "netrc.username",
		Usage:   "netrc username",
		EnvVars: []string{"CI_NETRC_USERNAME"},
	},
	&cli.StringFlag{
		Name:    "netrc.password",
		Usage:   "netrc password",
		EnvVars: []string{"CI_NETRC_PASSWORD"},
	},
	&cli.IntFlag{
		Name:    "depth",
		Usage:   "clone depth",
		EnvVars: []string{"PLUGIN_DEPTH"},
	},
	&cli.BoolFlag{
		Name:    "recursive",
		Usage:   "clone submodules",
		EnvVars: []string{"PLUGIN_RECURSIVE"},
		Value:   true,
	},
	&cli.BoolFlag{
		Name:    "tags",
		Usage:   "clone tags, if not explicitly set and event is tag its default is true else false",
		EnvVars: []string{"PLUGIN_TAGS"},
	},
	&cli.BoolFlag{
		Name:    "skip-verify",
		Usage:   "skip tls verification",
		EnvVars: []string{"PLUGIN_SKIP_VERIFY"},
	},
	&cli.StringFlag{
		Name:    "custom-cert",
		Usage:   "path or url to custom cert",
		EnvVars: []string{"PLUGIN_CUSTOM_SSL_PATH", "PLUGIN_CUSTOM_SSL_URL"},
	},
	&cli.BoolFlag{
		Name:    "submodule-update-remote",
		Usage:   "update remote submodules",
		EnvVars: []string{"PLUGIN_SUBMODULES_UPDATE_REMOTE", "PLUGIN_SUBMODULE_UPDATE_REMOTE"},
	},
	&cli.GenericFlag{
		Name:    "submodule-override",
		Usage:   "json map of submodule overrides",
		EnvVars: []string{"PLUGIN_SUBMODULE_OVERRIDE"},
		Value:   &MapFlag{},
	},
	&cli.BoolFlag{
		Name:    "submodule-partial",
		Usage:   "update submodules via partial clone (depth=1) (default)",
		EnvVars: []string{"PLUGIN_SUBMODULES_PARTIAL", "PLUGIN_SUBMODULE_PARTIAL"},
		Value:   true,
	},
	&cli.DurationFlag{
		Name:    "backoff",
		Usage:   "backoff duration",
		EnvVars: []string{"PLUGIN_BACKOFF"},
		Value:   5 * time.Second,
	},
	&cli.IntFlag{
		Name:    "backoff-attempts",
		Usage:   "backoff attempts",
		EnvVars: []string{"PLUGIN_ATTEMPTS"},
		Value:   5,
	},
	&cli.BoolFlag{
		Name:    "lfs",
		Usage:   "whether to retrieve LFS content if available",
		EnvVars: []string{"PLUGIN_LFS"},
		Value:   true,
	},
	&cli.StringFlag{
		Name:  "env-file",
		Usage: "source env file",
	},
	&cli.StringFlag{
		Name:    "branch",
		Usage:   "Change branch name",
		EnvVars: []string{"PLUGIN_BRANCH", "CI_COMMIT_BRANCH", "CI_REPO_DEFAULT_BRANCH"},
	},
	&cli.BoolFlag{
		Name:    "partial",
		Usage:   "Enable/Disable Partial clone",
		EnvVars: []string{"PLUGIN_PARTIAL"},
		Value:   true,
	},
	&cli.StringFlag{
		Name:    "home",
		Usage:   "Change home directory",
		EnvVars: []string{"PLUGIN_HOME"},
	},
	&cli.StringFlag{
		Name:    "safe-directory",
		Usage:   "Define/replace safe directories",
		EnvVars: []string{"PLUGIN_SAFE_DIRECTORY", "CI_WORKSPACE"},
	},
	&cli.BoolFlag{
		Name:    "use-ssh",
		Usage:   "Using ssh for git clone",
		EnvVars: []string{"PLUGIN_USE_SSH"},
		Value:   false,
	},
	&cli.StringFlag{
		Name:    "ssh-key",
		Usage:   "SSH key for ssh clone",
		EnvVars: []string{"PLUGIN_SSH_KEY"},
	},
}
