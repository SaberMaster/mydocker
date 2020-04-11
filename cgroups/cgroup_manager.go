package cgroups

import (
	"github.com/3i2bgod/mydocker/cgroups/subsystems"
	"github.com/Sirupsen/logrus"
)

type CgroupManager struct {
	// the path cgroup in hierarchy, the relative path of cgroup to root_cgroup
	Path string
	// source config
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubsystemIns {
		if err := subSysIns.Apply(c.Path, pid); nil != err {
			logrus.Warnf("apply cgroup fail :%v", err)
		}
	}
	return nil
}

func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubsystemIns {
		if err := subSysIns.Set(c.Path, res); nil != err {
			logrus.Warnf("set cgroup fail :%v", err)
		}
	}
	return nil
}

func (c *CgroupManager) Destory() error {
	// sometimes we cannot del cgroup dir, use instead cgdelete cpu:mydocker-cgroup
	// if we want to del cgroup dir, make sure that tasks file is empty
	for _, subSysIns := range subsystems.SubsystemIns {
		if err := subSysIns.Remove(c.Path); nil != err {
			logrus.Warnf("remove cgroup fail :%v", err)
		}
	}
	return nil
}
