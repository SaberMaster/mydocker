package main

import (
	"github.com/3i2bgod/mydocker/cgroups"
	"github.com/3i2bgod/mydocker/cgroups/subsystems"
	"github.com/3i2bgod/mydocker/container"
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

func Run(tty bool, cmdArray []string, res *subsystems.ResourceConfig)  {
	parent, writePipe := container.NewParentProcess(tty)

	if nil == parent {
		logrus.Error("new parent process error")
		return
	}

	if err := parent.Start(); nil != err  {
		logrus.Error(err)
	}

	setCgroupAndWaitParentProcess(res, parent, cmdArray, writePipe)
	os.Exit(0)
}

func setCgroupAndWaitParentProcess(res *subsystems.ResourceConfig, parent *exec.Cmd, cmdArray []string, writePipe *os.File) {
	// use docker-cgroup as cgroup name
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	sendInitCommand(cmdArray, writePipe)
	parent.Wait()
	// remove workspace
	mntURL := "/root/mnt/"
	rootURL := "/ramdisk/"
	container.DeleteWorkSpace(rootURL, mntURL)
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	logrus.Infof("command all is [%s]", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
