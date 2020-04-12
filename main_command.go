package main

import (
	"fmt"
	"github.com/3i2bgod/mydocker/cgroups/subsystems"
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
		cli.BoolFlag{
			Name:        "d",
			Usage:       "detach container",
		},
		cli.StringFlag{
			Name:        "m",
			Usage:       "memory limit",
		},
		cli.StringFlag{
			Name:        "cpushare",
			Usage:       "cpushare limit",
		},
		cli.StringFlag{
			Name:        "cpuset",
			Usage:       "cpuset limit",
		},

		cli.StringFlag{
			Name:        "v",
			Usage:       "volume",
		},

		cli.StringFlag{
			Name:        "name",
			Usage:       "container name",
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
		detach := ctx.Bool("d")

		if tty && detach {
			return fmt.Errorf("ti and p parameter can not both provided")
		}

		resConf := &subsystems.ResourceConfig{
			MemoryLimit: ctx.String("m"),
			CpuShare:    ctx.String("cpushare"),
			CpuSet:      ctx.String("cpuset"),
		}

		containerName := ctx.String("name")
		log.Infof("create tty %v", tty)
		Run(tty, cmdArray, resConf, "", containerName)
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

var commitCommand = cli.Command{
	Name:   "commit",
	Usage:  "commit a container into image",
	Action: func(ctx *cli.Context) error{
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}
		imageUrl := ctx.Args().Get(0)
		commitContainer(imageUrl)
		return nil
	},
}

var listCommand = cli.Command{
	Name:   "ps",
	Usage:  "list all containers info",
	Action: func(ctx *cli.Context) error{
		ListContainers()
		return nil
	},
}

var logCommand = cli.Command{
	Name:   "logs",
	Usage:  "print logs of a container",
	Action: func(ctx *cli.Context) error{
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}

		containerName := ctx.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}
