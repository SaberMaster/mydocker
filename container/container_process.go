package container

import (
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {
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
	// the fileSystem is overlay
	// but overlay fs can't be overlay upperDir and workDir
	// so i mount ram to a folder
	// mount -t tmpfs tmpfs /ramdisk/
	rootURL := "/ramdisk/"
	NewWorkSpace(rootURL, mntURL, volume)
	//change current dir to mntURL
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

func NewWorkSpace(rootURL string, mntURL string, volume string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL, "root")
	MountVolumeIfNeed(mntURL, volume)
}

func MountVolumeIfNeed(mntURL string, volume string) {
	if "" == volume {
		return
	}

	volumeURLs := volumeUrlExtract(volume)
	length := len(volumeURLs)

	if 2 == length &&
			"" != volumeURLs[0] &&
			"" != volumeURLs[1] {
		MountVolume(mntURL, volumeURLs)
		logrus.Infof("%q", volumeURLs)
	} else {
		logrus.Infof("Volume parameter input is not correct.")
	}
}

func MountVolume(mntURL string, volumeURLs []string) {
	parentURL := volumeURLs[0]
	logrus.Infof("make parent dir %s", parentURL)
	makeDir(parentURL)

	logrus.Infof("make container dir %s", parentURL)
	containerURL := volumeURLs[1]
	containerVolumeURL := path.Join(mntURL, containerURL)

	// here use overlay to mount, as i test in docker, so the parentURL can't by overlay
	CreateMountPoint(parentURL, containerVolumeURL, "mount")
}

func volumeUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
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
	if _, err := os.Stat(URL); nil != err && os.IsNotExist(err){
		if err := os.Mkdir(URL, 0777); nil != err {
			logrus.Errorf("Mkdir %s error. %v", URL, err)
		}
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

func CreateMountPoint(rootURL string, mntURL string, mountType string) {
	makeDir(mntURL)
	CreateWorkDir4Overlay(rootURL)
	var dirs string
	if "root" == mountType {
		dirs = "lowerdir=" + rootURL + "busybox,upperdir=" + rootURL + "writeLayer,workdir=" + rootURL + "work";
	} else {
		dirs = "lowerdir=" + rootURL+ "writeLayer,upperdir=" + rootURL + "writeLayer,workdir=" + rootURL + "work";
	}
	logrus.Infof("overlay mount options: %s", dirs)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", mntURL,  "-o", dirs)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); nil != err {
		logrus.Errorf("Create mount point err:%v", err)
	}
}

func DeleteWorkSpace(rootURL string, mntURL string, volume string)  {
	if "" != volume {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)

		if 2 == length &&
			"" != volumeURLs[0] &&
			"" != volumeURLs[1] {
			containerVolumeURL := path.Join(mntURL, volumeURLs[1])
			logrus.Infof("unmount volume %q", volumeURLs)
			DeleteMountPoint(volumeURLs[0], containerVolumeURL)
		} else {
			logrus.Infof("Volume parameter input is not correct.")
		}
	}

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
	logrus.Infof("Umount dir %s", mntURL)

	if err := os.RemoveAll(mntURL); nil != err {
		logrus.Errorf("Remove dir %s error: %v", mntURL, err)
	}
	logrus.Infof("Remove dir %s", mntURL)
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
