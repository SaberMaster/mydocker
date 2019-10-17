package container

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if nil == cmdArray || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command error, cmdArray is nil")
	}
	logrus.Infof("container user command [%s]", strings.Join(cmdArray, " "))

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// mount proc, or `ps` will search parents proc
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	// support env var
	path, err := exec.LookPath(cmdArray[0])
	if nil != err {
		logrus.Errorf("Exec look path error %v", err)
		return err
	}
	logrus.Infof("Find exec path: %s", path)

	// replace init proc with command
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); nil != err {
		logrus.Errorf(err.Error())
	}
	return nil
}

// read user command from the pipe we created
func readUserCommand() []string {
	// 3 is read pipe (the fourth pipe is the the cmd extraFiles)
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if nil != err {
		logrus.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}