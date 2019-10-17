package main

import (
	"github.com/3i2bgod/mydocker/container"
	"github.com/Sirupsen/logrus"
	"os"
	"strings"
)

func Run(tty bool, cmdArray []string)  {
	parent, writePipe := container.NewParentProcess(tty)

	if nil == parent {
		logrus.Error("new parent process error")
		return
	}
	if err := parent.Start(); nil != err  {
		logrus.Error(err)
	}
	sendInitCommand(cmdArray, writePipe)
	parent.Wait()
	os.Exit(0)
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	logrus.Infof("command all is [%s]", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
