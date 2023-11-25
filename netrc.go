package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/plugin-git/netrc"
)

var netrcCommand = &cli.Command{
	Name:  "netrc",
	Usage: "built-in credentials helper to read netrc",
	Flags: []cli.Flag{&cli.StringFlag{
		Name:    "home",
		Usage:   "Change home directory",
		EnvVars: []string{"PLUGIN_HOME"},
	}},
	Action: netrcGet,
}

func netrcGet(c *cli.Context) error {
	if c.Args().Len() == 0 {
		curExec, err := os.Executable()
		if err != nil {
			return err
		}
		fmt.Printf("built-in credentials helper to read netrc\n"+
			"exec \"git config --global credential.helper '%s netrc'\" to use it\n", curExec)
		return nil
	}

	// set custom home
	if c.IsSet("home") {
		os.Setenv("HOME", c.String("home"))
	}

	// implement custom git credentials helper
	// https://git-scm.com/docs/gitcredentials
	switch c.Args().First() {
	case "get":
		netRC, err := netrc.Read()
		if err != nil {
			return err
		}
		if netRC != nil {
			fmt.Printf("username=%s\n", netRC.Login)
			fmt.Printf("password=%s\n", netRC.Password)
			fmt.Println("quit=true")
		}
	case "store":
		// TODO: netrc.Save()
	case "erase":
		_, err := netrc.Delete()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("got unknown helper arg '%s'", c.Args().First())
	}

	return nil
}

func setNetRCHelper() *exec.Cmd {
	curExec, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	credHelper := fmt.Sprintf("%s netrc", curExec)

	return appendEnv(exec.Command("git", "config", "--global", "credential.helper", credHelper), defaultEnvVars...)
}
