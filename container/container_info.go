package container

import "fmt"

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