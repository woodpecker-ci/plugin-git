package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func SetDefaults(c *cli.Context, p *Plugin) {
	if p.Build.Event == "tag" && !c.IsSet("tags") {
		// tags clone not explicit set but pipeline is triggered by a tag
		// auto set tags cloning to true
		p.Config.Tags = true
	}

	if c.IsSet("tags") && c.IsSet("partial") {
		fmt.Println("WARNING: ignore partial clone as tags are fetched")
	}

	if p.Config.Tags && p.Config.Partial {
		// if tag fetching is enabled per event or setting, disable partial clone
		p.Config.Partial = false
	}

	if p.Config.Partial {
		p.Config.Depth = 1
		p.Config.filter = "tree:0"
	}
}
