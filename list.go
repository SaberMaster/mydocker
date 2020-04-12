package main

import (
	"encoding/json"
	"fmt"
	"github.com/3i2bgod/mydocker/container"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"text/tabwriter"
)

func ListContainers()  {
	containersDefaultLocation := container.GetContainerDefaultFilePath("")

	files, err := ioutil.ReadDir(containersDefaultLocation)

	if nil != err {
		logrus.Errorf("Read dir: %s err: %v", containersDefaultLocation, err)
		return
	}

	var containerInfos []*container.ContainerInfo

	for _, file := range files {
		containerInfo, err := getContainerInfo(file)
		if nil != err {
			logrus.Errorf("Get container info err: %v", err)
			continue
		}

		containerInfos = append(containerInfos, containerInfo)
	}

	writer := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(writer, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, containerInfo := range containerInfos {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			containerInfo.Id,
			containerInfo.Name,
			containerInfo.Pid,
			containerInfo.Status,
			containerInfo.Command,
			containerInfo.CreatedTime)

	}
	if err := writer.Flush(); nil != err {
		logrus.Errorf("Flush error: %v", err)
		return
	}
}

func getContainerInfo(file os.FileInfo) (*container.ContainerInfo, error) {

	containerName := file.Name()

	configFileDir := container.GetContainerDefaultFilePath(containerName)

	configFilePath := configFileDir + container.CONFIG_FILE_NAME

	content, err := ioutil.ReadFile(configFilePath)

	if nil != err {
		logrus.Errorf("Read file: %s err: %v", configFilePath, err)
		return nil, err
	}

	var containerInfo container.ContainerInfo

	if err := json.Unmarshal(content, &containerInfo); nil != err {
		logrus.Errorf("Json unmarshal error: %v", err)
		return nil, err
	}

	return &containerInfo, nil
}

