package main

import (
	"encoding/json"
	"fmt"
	"github.com/3i2bgod/mydocker/cgroups"
	"github.com/3i2bgod/mydocker/cgroups/subsystems"
	"github.com/3i2bgod/mydocker/container"
	"github.com/3i2bgod/mydocker/misc"
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func Run(tty bool, cmdArray []string, res *subsystems.ResourceConfig, volume string, containerName string) {
	parent, writePipe := container.NewParentProcess(tty, volume)

	if nil == parent {
		logrus.Error("new parent process error")
		return
	}

	if err := parent.Start(); nil != err  {
		logrus.Error(err)
	}

	containerName, err := recordContainerInfo(parent.Process.Pid, cmdArray, containerName)

	if nil != err {
		logrus.Errorf("Record container info error: %v", err)
		return
	}

	setCgroupAndWaitParentProcess(tty, res, parent, cmdArray, writePipe)
	//removeWorkSpace(volume)
	//os.Exit(0)

	if tty {
		deleteContainerInfo(containerName)
	}
}


func setCgroupAndWaitParentProcess(tty bool, res *subsystems.ResourceConfig, parent *exec.Cmd, cmdArray []string, writePipe *os.File) {
	// use docker-cgroup as cgroup name
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	if tty {
		defer cgroupManager.Destory()
	}
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	sendInitCommand(cmdArray, writePipe)
	if tty {
		parent.Wait()
	}
}

func removeWorkSpace(volume string)  {
	// remove workspace
	mntURL := "/root/mnt/"
	tempDirRoot := "/ramdisk/mydocker/tmp/"
	container.DeleteWorkSpace(tempDirRoot, mntURL, volume)
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	logrus.Infof("command all is [%s]", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

func recordContainerInfo(containerPID int, commandArray []string, containerName string) (string, error) {
	id := misc.RandomStringBytes(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")

	if "" == containerName {
		containerName = id
	}

	containerInfo := &container.ContainerInfo{
		Pid:         strconv.Itoa(containerPID),
		Id:          id,
		Name:        containerName,
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
	}

	jsonBytes, err := json.Marshal(containerInfo)

	if nil != err {
		logrus.Errorf("Record container info error: %v", err)
		return "", err
	}

	jsonStr := string(jsonBytes)

	containerDefaultLocation := fmt.Sprintf(container.DEFAULT_INFO_LOCATION, containerName)

	if err := os.MkdirAll(containerDefaultLocation, 0622); nil != err {
		logrus.Errorf("Mkdir dir: %s error: %v", containerDefaultLocation, err)
		return "", err
	}
	infoFileName := containerDefaultLocation + container.CONFIG_NAME

	file, err := os.Create(infoFileName)

	if nil != err {
		logrus.Errorf("Create file %s error: %v", infoFileName, err)
		return "", err
	}

	if _, err := file.WriteString(jsonStr); nil != err {
		logrus.Errorf("File %s write string error: %v", infoFileName, err)
		return "", err
	}

	return containerName, err
}

func deleteContainerInfo(containerName string) {
	containerDefaultLocation := fmt.Sprintf(container.DEFAULT_INFO_LOCATION, containerName)

	if err := os.RemoveAll(containerDefaultLocation); nil != err {
		logrus.Errorf("Remove dir %s error: %v", containerDefaultLocation, err)
	}
}
