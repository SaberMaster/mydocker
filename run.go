package main

import (
	"github.com/3i2bgod/mydocker/container"
	"github.com/Sirupsen/logrus"
	"os"
)

func Run(tty bool, command string)  {
	parent := container.NewParentProcess(tty, command)

	if err := parent.Start(); nil != err  {
		logrus.Error(err)
	}
	parent.Wait()
	os.Exit(-1)
}
