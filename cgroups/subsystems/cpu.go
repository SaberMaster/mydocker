package subsystems

type CpuSubSystem struct {
}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}

func (s *CpuSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	return SetCgroupResourceConfig(s.Name(), "cpu.shares", cgroupPath, res.CpuShare)
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	return ApplyCgroupResourceConfig(s.Name(), cgroupPath, pid)
}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	return RemoveCgroupResourceConfig(s.Name(), cgroupPath)
}
