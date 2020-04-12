package command

import (
	"fmt"
	"github.com/3i2bgod/mydocker/container"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

func LogContainer(containerName string) {
	configFileDir := container.GetContainerDefaultFilePath(containerName)

	logFilePath := configFileDir + container.LOG_FILE_NAME

	file, err := os.Open(logFilePath)
	defer file.Close()
	if nil != err {
		logrus.Errorf("Open container log file %s error: %v", logFilePath, err)
		return
	}

	content, err := ioutil.ReadAll(file)

	if nil != err {
		logrus.Errorf("Read container log file: %s err: %v", logFilePath, err)
		return
	}

	fmt.Fprint(os.Stdout, string(content))
}
