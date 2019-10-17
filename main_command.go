package main

import (
	"fmt"
	"github.com/3i2bgod/mydocker/container"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: "create a container",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:        "ti",
			Usage:       "enable tty",
		},
	},
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing container cmd")
		}

		var cmdArray []string
		for _, arg := range ctx.Args() {
			cmdArray = append(cmdArray, arg)
		}

		tty := ctx.Bool("ti")

		Run(tty, cmdArray)
		return nil
	},
}

var initCommand = cli.Command{
	Name:   "init",
	Usage:  "init container process run user's process in container. Do not call it outside",
	Action: func(ctx *cli.Context) error {
		log.Infof("init come on")
		err := container.RunContainerInitProcess()
		return err
	},
}