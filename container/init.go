package container

import (
	"github.com/Sirupsen/logrus"
	"os"
	"syscall"
)

func RunContainerInitProcess(command string, args []string) error {
	logrus.Infof("container init command %s", command)

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// mount proc, or `ps` will search parents proc
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	argv := []string{command}
	// replace init proc with command
	if err := syscall.Exec(command, argv, os.Environ()); nil != err {
		logrus.Errorf(err.Error())
	}
	return nil
}