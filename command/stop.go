package command

import (
	"encoding/json"
	"github.com/3i2bgod/mydocker/container"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"strconv"
	"syscall"
)

func StopContainer(containerName string)  {
	containerInfo, err := container.GetContainerInfo(containerName)
	if nil != err {
		logrus.Errorf("Get container: %s err: %v", containerName, err)
		return
	}
	pidInt, err := strconv.Atoi(containerInfo.Pid)

	if nil != err {
		logrus.Errorf("Convert pid from string to int error: %v", err)
		return
	}

	if err := syscall.Kill(pidInt, syscall.SIGTERM); nil != err {
		logrus.Errorf("Stop container: %s err: %v", containerName, err)
		return
	}

	containerInfo.Status = container.STOP
	containerInfo.Pid = " "

	newContainerBytes, err := json.Marshal(containerInfo)

	if nil != err {
		logrus.Errorf("Json marshal %s error: %v", containerName, err)
		return
	}

	containerConfigPath := container.GetContainerDefaultFilePath(containerName) + container.CONFIG_FILE_NAME

	if err := ioutil.WriteFile(containerConfigPath, newContainerBytes, 0622); nil != err {
		logrus.Errorf("Write file %s error", containerConfigPath)
	}
}
