package command

import (
	"fmt"
	"github.com/3i2bgod/mydocker/container"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	// we must include this file manual, or this file will be skipped
	_ "github.com/3i2bgod/mydocker/nsenter"
)

const ENV_EXEC_PID =  "mydocker_pid"
const ENV_EXEC_CMD =  "mydocker_cmd"

// because go will enter multi-thread when start a new process
// and Mount Namespace cannot be entered by `setns` when a process
// is in multi-thread status if we use golang
// so we use cgo to call native C code
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

	// as the forked process's parents process is not the actual container
	// is the host, so we need to get the container env vars and set in the forked process
	cmd.Env = append(os.Environ(), getEnvsByPid(pid)...)

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

func getEnvsByPid(pid string) []string {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := ioutil.ReadFile(path)
	if nil != err {
		logrus.Errorf("Read file: %s err: %v", path, err)
		return nil
	}

	return strings.Split(string(contentBytes), "\u0000")
}