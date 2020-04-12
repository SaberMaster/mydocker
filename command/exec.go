package command

import (
	"github.com/3i2bgod/mydocker/container"
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	// we must include this file manual, or this file will be skipped
	_ "github.com/3i2bgod/mydocker/nsenter"
)

const ENV_EXEC_PID =  "mydocker_pid"
const ENV_EXEC_CMD =  "mydocker_cmd"

func ExecContainer(containerName string, cmdArray []string) {
	pid, err := getContainerPidByName(containerName)

	if nil != err {
		logrus.Errorf("Exec container getContainerPidByName %s error %v", containerName, err)
		return
	}

	cmdStr := strings.Join(cmdArray, " ")
	logrus.Infof("container pid %s", pid)
	logrus.Infof("command %s", cmdStr)

	// run exec cmd in forked child process
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)

	if err := cmd.Run(); nil != err {
		logrus.Errorf("Exec container %s error %v", containerName, err)
	}
}

func getContainerPidByName(containerName string) (string, error) {
	containerInfo, err := container.GetContainerInfo(containerName)
	if nil != err {
		logrus.Errorf("Get container: %s err: %v", containerName, err)
		return "", err
	}

	return containerInfo.Pid, nil
}