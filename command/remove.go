package command

import (
	"github.com/3i2bgod/mydocker/container"
	"github.com/Sirupsen/logrus"
)

func RemoveContainer(containerName string)  {
	containerInfo, err := container.GetContainerInfo(containerName)
	if nil != err {
		logrus.Errorf("Get container: %s err: %v", containerName, err)
		return
	}

	// only remove the container which is stop
	if container.STOP != containerInfo.Status {
		logrus.Errorf("Couldn't remove running container")
		return
	}

	container.RemoveContainerDefaultDir(containerName)
}
