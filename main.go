package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

var version = "0.0.0+0"

func main() {
	app := cli.NewApp()
	app.Name = "git plugin"
	app.Usage = "git plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "remote",
			Usage:   "git remote url",
			EnvVars: []string{"PLUGIN_REMOTE", "CI_REPO_REMOTE", "CI_REMOTE_URL"},
		},
		&cli.StringFlag{
			Name:    "path",
			Usage:   "git clone path",
			EnvVars: []string{"PLUGIN_PATH", "CI_WORKSPACE"},
		},
		&cli.StringFlag{
			Name:    "sha",
			Usage:   "git commit sha",
			EnvVars: []string{"PLUGIN_SHA", "CI_COMMIT_SHA"},
		},
		&cli.StringFlag{
			Name:    "ref",
			Value:   "refs/heads/master",
			Usage:   "git commit ref",
			EnvVars: []string{"PLUGIN_REF", "CI_COMMIT_REF"},
		},
		&cli.StringFlag{
			Name:    "event",
			Value:   "push",
			Usage:   "build event",
			EnvVars: []string{"CI_BUILD_EVENT"},
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
			Usage:   "Define safe directories",
			EnvVars: []string{"PLUGIN_SAFE_DIRECTORY"},
			Value:   "*",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if c.String("env-file") != "" {
		_ = godotenv.Load(c.String("env-file"))
	}

	plugin := Plugin{
		Repo: Repo{
			Clone: c.String("remote"),
		},
		Build: Build{
			Commit: c.String("sha"),
			Event:  c.String("event"),
			Path:   c.String("path"),
			Ref:    c.String("ref"),
		},
		Netrc: Netrc{
			Login:    c.String("netrc.username"),
			Machine:  c.String("netrc.machine"),
			Password: c.String("netrc.password"),
		},
		Config: Config{
			Depth:           c.Int("depth"),
			Tags:            c.Bool("tags"),
			Recursive:       c.Bool("recursive"),
			SkipVerify:      c.Bool("skip-verify"),
			CustomCert:      c.String("custom-cert"),
			SubmoduleRemote: c.Bool("submodule-update-remote"),
			Submodules:      c.Generic("submodule-override").(*MapFlag).Get(),
			Lfs:             c.Bool("lfs"),
			Branch:          c.String("branch"),
			Partial:         c.Bool("partial"),
			Home:            c.String("home"),
			SafeDirectory:   c.String("safe-directory"),
		},
		Backoff: Backoff{
			Attempts: c.Int("backoff-attempts"),
			Duration: c.Duration("backoff"),
		},
	}

	SetDefaults(c, &plugin)

	return plugin.Exec()
}
