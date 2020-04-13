package command

import (
	"fmt"
	"github.com/3i2bgod/mydocker/container"
	log "github.com/Sirupsen/logrus"
	"os/exec"
)

func CommitContainer(containerName string, imageUrl string)  {
	mntURL := fmt.Sprintf(container.MNT_URL, containerName)

	if _, err := exec.Command("tar", "-czf", imageUrl, "-C", mntURL, ".").CombinedOutput(); nil != err {
		log.Errorf("Tar folder %s error %v", mntURL, err)
	}
}