package container

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if nil == cmdArray || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command error, cmdArray is nil")
	}
	logrus.Infof("container user command [%s]", strings.Join(cmdArray, " "))

	setUpMount()

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

/**
	init mountPoint
 */
func setUpMount()  {
	pwd, err := os.Getwd()

	if nil != err {
		logrus.Errorf("Get current location error %v", err)
	}

	logrus.Infof("Current location is %s", pwd)

	// fix: support new version linux kernel, if remove this line, will crash
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")

	pivotRoot(pwd)
	// mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// mount proc, or `ps` will search parents proc
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID | syscall.MS_STRICTATIME, "mode=755")
}

func pivotRoot(root string) error {
	// todo bind mount operation
	//After this call the same contents is accessible in two places.  One can also remount a single file (on a single file). It's also possible to use the bind mount to create a mountpoint from a regular directory, for exam-
	//	ple:
	//mount --bind foo foo
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND | syscall.MS_REC, ""); nil != err {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}

	// create dir to mount old_root to new file_system
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); nil != err {
		return err
	}

	// pivotRoot: moves the root file system of the current process to the directory pivotDir
	// and makes root the new root file system
	// after pivot, we will use new filesystem, `root` will be /
	// so the old filesystem now location is /.pivot_root
	if err := syscall.PivotRoot(root, pivotDir); nil != err {
		return fmt.Errorf("pivot_root: %v", err)
	}

	if err := syscall.Chdir("/"); nil != err {
		return fmt.Errorf("chdir to / error: %v", err)
	}

	// unmount old fs
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); nil != err{
		return fmt.Errorf("unmount pivot_root dir error : %v", err)
	}

	// remove tmp pivotDir
	return os.Remove(pivotDir)
}