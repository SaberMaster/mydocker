package main

import (
	"fmt"
	"github.com/3i2bgod/mydocker/cgroups/subsystems"
	"github.com/3i2bgod/mydocker/command"
	"github.com/3i2bgod/mydocker/container"
	"github.com/3i2bgod/mydocker/network"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "create a container ie: mydocker run -ti [image] [command]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},

		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},

		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},

		cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment",
		},

		cli.StringFlag{
			Name:  "net",
			Usage: "container network",
		},

		cli.StringSliceFlag{
			Name:  "p",
			Usage: "port mapping",
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

		//get image name
		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]

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

		envSlice := ctx.StringSlice("e")

		volume := ctx.String("v")

		network := ctx.String("net")
		portMapping := ctx.StringSlice("p")

		command.RunContainer(tty, cmdArray, resConf, volume, containerName, envSlice, imageName, network, portMapping)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "init container process run user's process in container. Do not call it outside",
	Action: func(ctx *cli.Context) error {
		log.Infof("init come on")
		err := container.RunContainerInitProcess()
		return err
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}
		containerName := ctx.Args().Get(0)
		imageUrl := ctx.Args().Get(1)
		command.CommitContainer(containerName, imageUrl)
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all containers info",
	Action: func(ctx *cli.Context) error {
		command.ListContainers()
		return nil
	},
}

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}

		containerName := ctx.Args().Get(0)
		command.LogContainer(containerName)
		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(ctx *cli.Context) error {

		if "" != os.Getenv(command.ENV_EXEC_PID) {
			log.Infof("pid callback pid %s", os.Getpid())
			return nil
		}

		if len(ctx.Args()) < 2 {
			return fmt.Errorf("Missing container name or cmd")
		}

		containerName := ctx.Args().Get(0)

		var cmdArray []string
		for _, arg := range ctx.Args().Tail() {
			cmdArray = append(cmdArray, arg)
		}

		command.ExecContainer(containerName, cmdArray)
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}

		containerName := ctx.Args().Get(0)
		command.StopContainer(containerName)
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove a container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}

		containerName := ctx.Args().Get(0)
		command.RemoveContainer(containerName)
		return nil
	},
}

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "create a container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "subnet cidr",
				},
			},
			Action: func(ctx *cli.Context) error {
				if len(ctx.Args()) < 1 {
					return fmt.Errorf("Missing network name")
				}
				network.Init()
				err := network.CreateNetwork(ctx.String("driver"),
					ctx.String("subnet"),
					ctx.Args()[0])
				if nil != err {
					return fmt.Errorf("create network error: %+v", err)
				}
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "list container network",
			Action: func(ctx *cli.Context) error {
				network.Init()
				network.ListNetwork()
				return nil
			},
		},

		{
			Name:  "remove",
			Usage: "remove a container network",
			Action: func(ctx *cli.Context) error {
				if len(ctx.Args()) < 1 {
					return fmt.Errorf("Missing network name")
				}
				network.Init()
				err := network.DeleteNetwork(ctx.Args()[0])
				if nil != err {
					return fmt.Errorf("remove network error: %+v", err)
				}
				return nil
			},
		},
	},
}
