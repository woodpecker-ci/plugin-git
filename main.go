package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
)

var version = "0.0.0+0"

func main() {
	app := cli.Command{}
	app.Name = "git plugin"
	app.Usage = "git plugin"
	app.Action = run
	app.Version = version
	app.Flags = globalFlags

	ctx := context.Background()

	if err := app.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, c *cli.Command) error {
	if c.String("env-file") != "" {
		_ = godotenv.Load(c.String("env-file"))
	}

	plugin := Plugin{
		Repo: Repo{
			Clone:        c.String("remote"),
			CloneSSH:     c.String("remote-ssh"),
			ObjectFormat: c.String("object-format"),
		},
		Pipeline: Pipeline{
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
			Depth:             c.Int("depth"),
			Tags:              c.Bool("tags"),
			Recursive:         c.Bool("recursive"),
			SkipVerify:        c.Bool("skip-verify"),
			CustomCert:        c.String("custom-cert"),
			SubmoduleRemote:   c.Bool("submodule-update-remote"),
			Submodules:        c.String("submodule-override"),
			SubmodulePartial:  c.Bool("submodule-partial"),
			Lfs:               c.Bool("lfs"),
			Branch:            c.String("branch"),
			Partial:           c.Bool("partial"),
			Home:              c.String("home"),
			SafeDirectory:     c.String("safe-directory"),
			UseSSH:            c.Bool("use-ssh"),
			SSHKey:            c.String("ssh-key"),
			MergePullRequest:  c.Bool("merge-pull-request"),
			TargetBranch:      c.String("target-branch"),
			FetchTargetBranch: c.Bool("fetch-target-branch"),
			Event:             c.String("event"),
			GitUserName:       c.String("git-user-name"),
			GitUserEmail:      c.String("git-user-email"),
		},
		Backoff: Backoff{
			Attempts: c.Int("backoff-attempts"),
			Duration: c.Duration("backoff"),
		},
	}

	SetDefaults(c, &plugin)

	return plugin.Exec()
}
