package main

import (
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
		containerInfo, err := container.GetContainerInfo(file.Name())
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
