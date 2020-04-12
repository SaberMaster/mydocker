package container

import (
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

func NewParentProcess(tty bool, volume string, containerName string, envSlice []string) (*exec.Cmd, *os.File) {
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

	// as we need to detach child process, so the unmount and clean dir
	// is not impl after parent process exit
	//mntURL := "/root/mnt/"
	//// as I test the project in docker
	//// the fileSystem is overlay
	//// but overlay fs can't be overlay upperDir and workDir
	//// so i mount ram to a folder
	//// mount -t tmpfs tmpfs /ramdisk/
	//tmpDirRoot := "/ramdisk/mydocker/tmp/"
	//imageTarUrl := "/ramdisk/busybox.tar"
	//NewWorkSpace(tmpDirRoot, mntURL, imageTarUrl, volume)
	//change current dir to mntURL

	// if need to unzip busybox.tar to `mntURL` folder
	mntURL := "/root/busybox"
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

func NewWorkSpace(tmpDirRoot string, mntURL string, imageTarUrl string, volume string) {
	readonlyLayer := CreateReadOnlyLayer4Image(tmpDirRoot, imageTarUrl)
	writeLayer := CreateWriteLayer(tmpDirRoot)
	CreateMountPoint(tmpDirRoot, readonlyLayer, writeLayer, mntURL)
	MountVolumeIfNeed(tmpDirRoot, mntURL, volume)
}

func MountVolumeIfNeed(tmpDirRoot string, mntURL string, volume string) {
	if "" == volume {
		return
	}

	volumeURLs := volumeUrlExtract(volume)
	length := len(volumeURLs)

	if 2 == length &&
			"" != volumeURLs[0] &&
			"" != volumeURLs[1] {
		tmpMountDir := tmpDirRoot + "mounts/"
		makeDir(tmpMountDir)
		MountVolume(tmpMountDir, mntURL, volumeURLs)
		logrus.Infof("%q", volumeURLs)
	} else {
		logrus.Infof("Volume parameter input is not correct.")
	}
}

func MountVolume(tmpDirRoot string, mntURL string, volumeURLs []string) {
	parentURL := volumeURLs[0]
	logrus.Infof("make parent dir %s", parentURL)
	makeDir(parentURL)

	logrus.Infof("make container dir %s", parentURL)
	containerURL := volumeURLs[1]
	containerVolumeURL := path.Join(mntURL, containerURL)

	readonlyLayer := CreateReadOnlyLayer4MountPoint(tmpDirRoot)
	// here use overlay to mount, as i test in docker, so the parentURL can't by overlay
	//tmpDirRoot, lowerDir, upperDir, mntURL
	CreateMountPoint(tmpDirRoot, readonlyLayer, parentURL, containerVolumeURL)
}

func volumeUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
}


func CreateReadOnlyLayer4Image(tmpDirRoot string, imageTarUrl string) string{
	containerReadOnlyRoot := tmpDirRoot + "containerReadOnlyRoot/"

	exists, err := PathExists(containerReadOnlyRoot)
	if nil != err {
		logrus.Infof("Fail to judge whether dir %s exists. %v", containerReadOnlyRoot, err)
	}

	if false == exists {
		makeDir(containerReadOnlyRoot)
		if _, err := exec.Command("tar", "-xvf", imageTarUrl, "-C", containerReadOnlyRoot).CombinedOutput(); nil != err {
			logrus.Errorf("Untar file %s to dir %s error: %v", imageTarUrl, containerReadOnlyRoot, err)
		}
	}
	return containerReadOnlyRoot
}

func makeDir(URL string) {
	logrus.Infof("Mkdir %s", URL)
	if _, err := os.Stat(URL); nil != err && os.IsNotExist(err){
		if err := os.Mkdir(URL, 0777); nil != err {
			logrus.Errorf("Mkdir %s error. %v", URL, err)
		}
	}
}

func CreateReadOnlyLayer4MountPoint(tmpDirRoot string) string{
	readOnlyLayer := tmpDirRoot + "readOnlyLayer/"
	makeDir(readOnlyLayer)
	return readOnlyLayer
}

func CreateWriteLayer(tmpDirRoot string) string{
	writeURL := tmpDirRoot + "writeLayer/"
	makeDir(writeURL)
	return writeURL
}

func CreateWorkDir4Overlay(tmpDirRoot string) string{
	// workDir is using for overlay work
	workURL := tmpDirRoot + "work/"
	makeDir(workURL)
	return workURL
}

//tmpDirRoot, lowerDir, upperDir, mntURL
func CreateMountPoint(tempDirRoot string, lowerDir string, upperDir string, mntURL string) {
	makeDir(mntURL)
	workDir := CreateWorkDir4Overlay(tempDirRoot)
	var dirs string
	dirs = "lowerdir=" + lowerDir + ",upperdir=" + upperDir + ",workdir=" + workDir;
	logrus.Infof("overlay mount options: %s", dirs)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", mntURL,  "-o", dirs)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); nil != err {
		logrus.Errorf("Create mount point err:%v", err)
	}
}

func DeleteWorkSpace(tempDirRoot string, mntURL string, volume string)  {
	if "" != volume {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)

		if 2 == length &&
			"" != volumeURLs[0] &&
			"" != volumeURLs[1] {
			containerVolumeURL := path.Join(mntURL, volumeURLs[1])
			logrus.Infof("unmount volume %q", volumeURLs)
			tmpMountDir := tempDirRoot + "mounts/"
			DeleteMountPoint(tmpMountDir, containerVolumeURL)
			DeleteReadOnlyLayer(tmpMountDir)
		} else {
			logrus.Infof("Volume parameter input is not correct.")
		}
	}

	DeleteMountPoint(tempDirRoot, mntURL)
	DeleteWriteLayer(tempDirRoot)
	DeleteContainerReadOnlyLayer(tempDirRoot)
}

func DeleteMountPoint(tempDirRoot string, mntURL string) {
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
	DeleteWorkDir4Overlay(tempDirRoot)
}

func DeleteReadOnlyLayer(tempDirRoot string) {
	workURL := tempDirRoot + "readOnlyLayer/"

	if err := os.RemoveAll(workURL); nil != err {
		logrus.Errorf("Remove dir %s error: %v", tempDirRoot, err)
	}
}

func DeleteContainerReadOnlyLayer(tempDirRoot string) {
	workURL := tempDirRoot + "containerReadOnlyRoot/"

	if err := os.RemoveAll(workURL); nil != err {
		logrus.Errorf("Remove dir %s error: %v", tempDirRoot, err)
	}
}

func DeleteWriteLayer(tempDirRoot string) {
	writeURL := tempDirRoot + "writeLayer/"

	if err := os.RemoveAll(writeURL); nil != err {
		logrus.Errorf("Remove dir %s error: %v", tempDirRoot, err)
	}
}

func DeleteWorkDir4Overlay(tempDirRoot string) {
	workURL := tempDirRoot + "work/"

	if err := os.RemoveAll(workURL); nil != err {
		logrus.Errorf("Remove dir %s error: %v", tempDirRoot, err)
	}
}
