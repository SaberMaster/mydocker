package container

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	CreatedTime string `json:"created_time"`
	Status      string `json:"status"`
}

var (
	RUNNING               = "running"
	STOP                  = "stppped"
	EXIT                  = "exited"
	DEFAULT_INFO_LOCATION = "/var/run/mydocker/%s/"
	CONFIG_FILE_NAME      = "config.json"
	LOG_FILE_NAME         = "container.log"
)

func GetContainerDefaultFilePath(containerName string) string {
	return fmt.Sprintf(DEFAULT_INFO_LOCATION, containerName)
}

func GetContainerInfo(containerName string) (*ContainerInfo, error)  {
	containersDefaultPath := GetContainerDefaultFilePath(containerName)
	configFilePath := containersDefaultPath + CONFIG_FILE_NAME

	content, err := ioutil.ReadFile(configFilePath)

	if nil != err {
		logrus.Errorf("Read file: %s err: %v", configFilePath, err)
		return nil, err
	}

	var containerInfo ContainerInfo

	if err := json.Unmarshal(content, &containerInfo); nil != err {
		logrus.Errorf("Json unmarshal error: %v", err)
		return nil, err
	}

	return &containerInfo, nil
}


func RemoveContainerDefaultDir(containerName string) {
	containerDefaultLocation := GetContainerDefaultFilePath(containerName)

	if err := os.RemoveAll(containerDefaultLocation); nil != err {
		logrus.Errorf("Remove dir %s error: %v", containerDefaultLocation, err)
	}
}
