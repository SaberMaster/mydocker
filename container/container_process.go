package container

import (
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
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

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	mntURL := "/root/mnt/"
	// as I test the project in docker
	// the fileSystem is already overlay
	// but overlay fs can't be overlay upperDir and workDir
	// so i mount ram to a folder
	// mount -t tmpfs tmpfs /ramdisk/
	rootURL := "/ramdisk/"
	NewWorkSpace(rootURL, mntURL)
	cmd.Dir = mntURL
	// attach the readPipe to the cmd
	cmd.ExtraFiles = []*os.File{readPipe}
	// return writePipe to send user cmd
	return cmd, writePipe
}


func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if nil == err {
		return true, err
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func NewWorkSpace(rootURL string, mntURL string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)
}

func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"

	exists, err := PathExists(busyboxURL)
	if nil != err {
		logrus.Infof("Fail to judge whether dir %s exists. %v", busyboxURL, err)
	}

	if false == exists {
		makeDir(busyboxURL)
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); nil != err {
			logrus.Errorf("Untar file %s to dir %s error: %v", busyboxTarURL, busyboxURL, err)
		}
	}
}

func makeDir(URL string) {
	if err := os.Mkdir(URL, 0777); nil != err {
		logrus.Errorf("Mkdir %s error. %v", URL, err)
	}
}

func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	makeDir(writeURL)
}

func CreateWorkDir4Overlay(rootURL string) {
	// workDir is using for overlay work
	workURL := rootURL + "work/"
	makeDir(workURL)
}

func CreateMountPoint(rootURL string, mntURL string) {
	makeDir(mntURL)
	CreateWorkDir4Overlay(rootURL)
	dirs := "lowerdir=" + rootURL + "busybox,upperdir=" + rootURL + "writeLayer,workdir=" + rootURL + "work";
	logrus.Infof("overlay mount options: %s", dirs)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", mntURL,  "-o", dirs)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); nil != err {
		logrus.Errorf("Create mount point err:%v", err)
	}
}

func DeleteWorkSpace(rootURL string, mntURL string)  {
	DeleteMountPoint(rootURL, mntURL)
	DeleteWriteLayer(rootURL)
}

func DeleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os .Stderr

	if err := cmd.Run(); nil != err {
		logrus.Errorf("umount error: %v", err)
	}

	if err := os.RemoveAll(mntURL); nil != err {
		logrus.Errorf("Remove dir %s error: %v", mntURL, err)
	}
	DeleteWorkDir4Overlay(rootURL)
}

func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"

	if err := os.RemoveAll(writeURL); nil != err {
		logrus.Errorf("Remove dir %s error: %v", rootURL, err)
	}
}

func DeleteWorkDir4Overlay(rootURL string) {
	workURL := rootURL + "work/"

	if err := os.RemoveAll(workURL); nil != err {
		logrus.Errorf("Remove dir %s error: %v", rootURL, err)
	}
}
