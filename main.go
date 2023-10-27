package main

import (
	"log"
	"os"

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
	app.Commands = []*cli.Command{netrcCommand}
	app.Flags = globalFlags

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
			Clone:    c.String("remote"),
			CloneSSH: c.String("remote-ssh"),
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
			UseSSH:          c.Bool("use-ssh"),
			SSHKey:          c.String("ssh-key"),
		},
		Backoff: Backoff{
			Attempts: c.Int("backoff-attempts"),
			Duration: c.Duration("backoff"),
		},
	}

	SetDefaults(c, &plugin)

	return plugin.Exec()
}
