package command

import (
	"encoding/json"
	"github.com/3i2bgod/mydocker/cgroups"
	"github.com/3i2bgod/mydocker/cgroups/subsystems"
	"github.com/3i2bgod/mydocker/container"
	"github.com/3i2bgod/mydocker/misc"
	"github.com/3i2bgod/mydocker/network"
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func RunContainer(tty bool, cmdArray []string, res *subsystems.ResourceConfig, volume string, containerName string, envSlice []string, imageName string, network string, portMapping []string) {
	containerId := misc.RandomStringBytes(10)
	if "" == containerName {
		containerName = containerId
	}

	parent, writePipe := container.NewParentProcess(tty, volume, containerName, envSlice, imageName)

	if nil == parent {
		logrus.Error("new parent process error")
		return
	}

	if err := parent.Start(); nil != err {
		logrus.Error(err)
	}

	containerInfo, err := recordContainerInfo(parent.Process.Pid, cmdArray, containerName, containerId, volume, network, portMapping)

	if nil != err {
		logrus.Errorf("Record container info error: %v", err)
		return
	}

	setCgroupAndNetworkAndWaitParentProcess(tty, res, parent, cmdArray, writePipe, containerInfo)
	//removeWorkSpace(volume)
	//os.Exit(0)

	if tty {
		container.RemoveContainerDefaultDir(containerInfo.Name)
		container.RemoveWorkSpace(containerInfo.Name, volume)
	}
	os.Exit(0)
}

func setCgroupAndNetworkAndWaitParentProcess(tty bool, res *subsystems.ResourceConfig, parent *exec.Cmd, cmdArray []string, writePipe *os.File, containerInfo *container.ContainerInfo) {
	// use docker-cgroup as cgroup name
	cgroupManager := cgroups.NewCgroupManager(containerInfo.Id)
	if tty {
		defer cgroupManager.Destory()
	}
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	// set network
	if "" != containerInfo.Network {
		network.Init()
		if err := network.Connect(containerInfo.Network, containerInfo); nil != err {
			logrus.Errorf("Connect Network error: %v", err)
			return
		}
	}
	sendInitCommand(cmdArray, writePipe)
	if tty {
		parent.Wait()
	}
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	logrus.Infof("command all is [%s]", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

func recordContainerInfo(containerPID int, commandArray []string, containerName string, containerId string, volume string, nw string, portMapping []string) (*container.ContainerInfo, error) {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")

	containerInfo := &container.ContainerInfo{
		Pid:         strconv.Itoa(containerPID),
		Id:          containerId,
		Name:        containerName,
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Volume:      volume,
		Network:     nw,
		PortMapping: portMapping,
	}

	jsonBytes, err := json.Marshal(containerInfo)

	if nil != err {
		logrus.Errorf("Record container info error: %v", err)
		return nil, err
	}

	jsonStr := string(jsonBytes)

	containerDefaultLocation := container.GetContainerDefaultFilePath(containerName)

	if err := os.MkdirAll(containerDefaultLocation, 0622); nil != err {
		logrus.Errorf("Mkdir dir: %s error: %v", containerDefaultLocation, err)
		return nil, err
	}
	infoFileName := containerDefaultLocation + container.CONFIG_FILE_NAME

	file, err := os.Create(infoFileName)

	if nil != err {
		logrus.Errorf("Create file %s error: %v", infoFileName, err)
		return nil, err
	}

	if _, err := file.WriteString(jsonStr); nil != err {
		logrus.Errorf("File %s write string error: %v", infoFileName, err)
		return nil, err
	}

	return containerInfo, err
}
