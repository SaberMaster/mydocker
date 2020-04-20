package container

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"path"
	"strings"
)

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

func NewWorkSpace(imageName string, containerName string, volume string) {
	//// as I test the project in docker
	//// the fileSystem is overlay
	//// but overlay fs can't be overlay upperDir and workDir
	//// so i mount ram to a folder
	//// mount -t tmpfs tmpfs /ramdisk/
	readonlyLayer := CreateReadOnlyLayer4Image(imageName, containerName)
	if "" == readonlyLayer {
		logrus.Errorf("Create readonly layer failed for image: %s", imageName)
		return
	}

	writeLayer := CreateWriteLayer(fmt.Sprintf(OVERLAY_TMP_URL,  containerName, "container"))

	mntURL := fmt.Sprintf(MNT_URL, containerName)
	makeDir(mntURL)
	CreateMountPoint(readonlyLayer, writeLayer, containerName, mntURL, "container")

	MountVolumeIfNeed(containerName, volume)
}

func RemoveWorkSpace(containerName string, volume string)  {
	// remove workspace
	DeleteWorkSpace(containerName, volume)
}

func MountVolumeIfNeed(containerName string, volume string) {
	if "" == volume {
		return
	}

	volumeURLs := volumeUrlExtract(volume)
	length := len(volumeURLs)

	if 2 == length &&
		"" != volumeURLs[0] &&
		"" != volumeURLs[1] {
		tmpMountDir := fmt.Sprintf(OVERLAY_TMP_URL, containerName, "mounts")
		makeDir(tmpMountDir)
		MountVolume(tmpMountDir, containerName, volumeURLs)
		logrus.Infof("%q", volumeURLs)
	} else {
		logrus.Infof("Volume parameter input is not correct.")
	}
}

func MountVolume(tmpDirRoot string, containerName string, volumeURLs []string) {
	parentURL := volumeURLs[0]
	logrus.Infof("make parent dir %s", parentURL)
	makeDir(parentURL)


	logrus.Infof("make container dir %s", parentURL)
	containerURL := volumeURLs[1]
	containerMountUrlPrefix := fmt.Sprintf(MNT_URL, containerName)
	containerVolumeURL := path.Join(containerMountUrlPrefix, containerURL)
	makeDir(containerVolumeURL)

	readonlyLayer := CreateReadOnlyLayer4MountPoint(tmpDirRoot)
	// here use overlay to mount, as i test in docker, so the parentURL can't by overlay
	CreateMountPoint(readonlyLayer, parentURL, containerName, containerVolumeURL,"mounts")
}

func volumeUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
}

func unzipImage(unTarFolderUri string, imageUri string) error {
	exist, err := PathExists(unTarFolderUri)

	if nil != err {
		logrus.Infof("Fail to judge whether dir %s exists. %v", unTarFolderUri, err)
		return err
	}

	if false == exist {
		err := makeDir(unTarFolderUri)
		if nil != err {
			return err
		}
		if _, err := exec.Command("tar", "-xvf", imageUri, "-C", unTarFolderUri).CombinedOutput(); nil != err {
			logrus.Errorf("Untar file %s to dir %s error: %v", imageUri, unTarFolderUri, err)
			return err
		}
	}
	return nil
}


func CreateReadOnlyLayer4Image(imageName string, containerName string) string {
	unTarFolderUri := fmt.Sprintf(OVERLAY_TMP_URL, containerName, "image")
	imageUri := ROOT_URL + "/" + imageName + ".tar"

	if err := unzipImage(unTarFolderUri, imageUri); nil != err {
		return ""
	}

	return unTarFolderUri
}

func makeDir(URL string) error {
	logrus.Infof("Mkdir %s", URL)
	if _, err := os.Stat(URL); nil != err && os.IsNotExist(err){
		if err := os.MkdirAll(URL, 0777); nil != err {
			logrus.Errorf("Mkdir %s error. %v", URL, err)
			return err
		}
	}
	return nil
}

func CreateReadOnlyLayer4MountPoint(tmpDirRoot string) string{
	readOnlyLayer := tmpDirRoot + "/" + "readOnlyLayer/"
	makeDir(readOnlyLayer)
	return readOnlyLayer
}

func CreateWriteLayer(tmpDirRoot string) string{
	writeURL := tmpDirRoot + "/" + "writeLayer"
	makeDir(writeURL)
	return writeURL
}

func CreateWorkDir4Overlay(tmpDirRoot string) string{
	// workDir is using for overlay work
	workURL := tmpDirRoot + "/" + "work"
	makeDir(workURL)
	return workURL
}

func CreateMountPoint(readonlyLayer string, writeLayer string, containerName string, mntUrl string, mode string) {
	workDir := CreateWorkDir4Overlay(fmt.Sprintf(OVERLAY_TMP_URL, containerName, mode))
	var dirs string
	dirs = "lowerdir=" + readonlyLayer + ",upperdir=" + writeLayer + ",workdir=" + workDir;
	logrus.Infof("overlay mount options: %s ,mounts on dir: %s", dirs, mntUrl)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", mntUrl,  "-o", dirs)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); nil != err {
		logrus.Errorf("Create mount point err:%v", err)
	}
}

func DeleteWorkSpace(containerName string, volume string)  {
	containerMountUrlPrefix := fmt.Sprintf(MNT_URL, containerName)
	if "" != volume {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)

		if 2 == length &&
			"" != volumeURLs[0] &&
			"" != volumeURLs[1] {
			containerVolumeURL := path.Join(containerMountUrlPrefix, volumeURLs[1])
			logrus.Infof("unmount volume %q", volumeURLs)
			tmpMountDir := fmt.Sprintf(OVERLAY_TMP_URL, containerName, "mounts")
			DeleteMountPoint(tmpMountDir, containerVolumeURL)
			DeleteReadOnlyLayer(tmpMountDir)
		} else {
			logrus.Infof("Volume parameter input is not correct.")
		}
	}

	DeleteMountPoint(fmt.Sprintf(OVERLAY_TMP_URL,  containerName, "container"), containerMountUrlPrefix)
	DeleteWriteLayer(fmt.Sprintf(OVERLAY_TMP_URL,  containerName, "container"))
	DeleteContainerReadOnlyLayer(containerName)

	// remove overlay tmp file
	containerOverlayTmp := fmt.Sprintf(OVERLAY_TMP_URL, containerName, "")
	if err := os.RemoveAll(containerOverlayTmp); nil != err {
		logrus.Errorf("Remove dir %s error: %v", containerOverlayTmp, err)
	}
}

func DeleteMountPoint(tempDirRoot string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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

func DeleteContainerReadOnlyLayer(containerName string) {
	unTarFolderUri := fmt.Sprintf(OVERLAY_TMP_URL, containerName, "image")

	if err := os.RemoveAll(unTarFolderUri); nil != err {
		logrus.Errorf("Remove dir %s error: %v", unTarFolderUri, err)
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
