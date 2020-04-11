package subsystems

import (
	"bufio"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

func FindCgroupMountpoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if nil != err {
		return ""
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields) - 1], ",") {
			if opt == subsystem {
				return fields[4]
			}
		}
	}

	if err := scanner.Err(); nil != err {
		return ""
	}
	return ""
}

func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error)  {
	cgroupRoot := FindCgroupMountpoint(subsystem)
	cgroupFullPath := path.Join(cgroupRoot, cgroupPath)
	logrus.Infof("cgroup full path: %v", cgroupFullPath)
	if _, err := os.Stat(cgroupFullPath); nil == err || (autoCreate && os.IsNotExist(err)){
		if os.IsNotExist(err) {
			if err := os.Mkdir(cgroupFullPath, 0755); nil != err {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
			logrus.Infof("create dir: %s", cgroupFullPath)
		}
		return cgroupFullPath, nil
	} else {
		return "", fmt.Errorf("cgroup path error %v", err)
	}
}


func writeCgroupResourceConfig2File(resourceName string, cgroupPath string, resourceValue string, resourceCgroupFileName string) error {
	if "" == resourceValue {
		return nil
	}
	if subsysCgroupPath, err := GetCgroupPath(resourceName, cgroupPath, true); nil == err {
		resourceFile := path.Join(subsysCgroupPath, resourceCgroupFileName)
		logrus.Infof("set cgroup %s to path %s value: %s", resourceName, resourceFile, resourceValue)
		if err := ioutil.WriteFile(resourceFile,
			[]byte(resourceValue),
			0644);
			nil != err {
			return fmt.Errorf("set cgroup %s fail %v", resourceName, err)
		}
		return nil
	} else {

		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

func SetCgroupResourceConfig(resourceName string, resourceCgroupFileName string, cgroupPath string, resourceValue string) error {
	return writeCgroupResourceConfig2File(resourceName, cgroupPath, resourceValue, resourceCgroupFileName)
}

func ApplyCgroupResourceConfig(resourceName string, cgroupPath string, pid int) error {
	return writeCgroupResourceConfig2File(resourceName, cgroupPath, strconv.Itoa(pid), "tasks")
}

func RemoveCgroupResourceConfig(resourceName string, cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(resourceName, cgroupPath, true); nil == err {
		logrus.Infof("remove subsysCgroupPath:%s", subsysCgroupPath)
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}
