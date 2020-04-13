package container

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

func NewParentProcess(tty bool, volume string, containerName string, envSlice []string, imageName string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := os.Pipe()
	if nil != err {
		logrus.Errorf("new pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	logrus.Info("init parent process cmd: [init]")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	cmd.Env = append(os.Environ(), envSlice...)
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		containerDefaultLocation := GetContainerDefaultFilePath(containerName)
		if err := os.MkdirAll(containerDefaultLocation, 0622); nil != err {
			logrus.Errorf("Mkdir dir: %s error: %v", containerDefaultLocation, err)
			return nil, nil
		}

		stdLogFilePath := containerDefaultLocation + LOG_FILE_NAME

		stdLogFile, err := os.Create(stdLogFilePath)
		if nil != err {
			logrus.Errorf("Create file %s error: %v", stdLogFile, err)
			return nil, nil
		}

		cmd.Stdout = stdLogFile
	}

	NewWorkSpace(imageName, containerName, volume)
	//change current dir to mntURL
	cmd.Dir = fmt.Sprintf(MNT_URL, containerName)
	// attach the readPipe to the cmd
	cmd.ExtraFiles = []*os.File{readPipe}
	// return writePipe to send user cmd
	return cmd, writePipe
}


