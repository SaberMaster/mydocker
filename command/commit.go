package command

import (
	log "github.com/Sirupsen/logrus"
	"os/exec"
)

func CommitContainer(imageUrl string)  {
	mntURL := "/root/mnt/"

	if _, err := exec.Command("tar", "-czf", imageUrl, "-C", mntURL, ".").CombinedOutput(); nil != err {
		log.Errorf("Tar folder %s error %v", mntURL, err)
	}
}