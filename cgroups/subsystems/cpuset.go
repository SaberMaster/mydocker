package subsystems

type CpusetSubSystem struct {
}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}

func (s *CpusetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	//// init mems and cpus value, or will occur "fail write /sys/fs/cgroup/cpuset/mydocker-cgroup/tasks: no space left on device" error
	if err := SetCgroupResourceConfig(s.Name(), "cpuset.mems", cgroupPath, "0"); nil != err {
		return err
	}
	if "" == res.CpuSet {
		res.CpuSet = "0"
	}
	return SetCgroupResourceConfig(s.Name(), "cpuset.cpus", cgroupPath, res.CpuSet)
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	return ApplyCgroupResourceConfig(s.Name(), cgroupPath, pid)
}

func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	return RemoveCgroupResourceConfig(s.Name(), cgroupPath)
}
