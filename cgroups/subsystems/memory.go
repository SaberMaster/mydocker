package subsystems

type MemorySubSystem struct {
}

func (s *MemorySubSystem) Name() string {
	return "memory"
}

func (s *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	return SetCgroupResourceConfig(s.Name(), "memory.limit_in_bytes", cgroupPath, res.MemoryLimit)
}

func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	return ApplyCgroupResourceConfig(s.Name(), cgroupPath, pid)
}

func (s *MemorySubSystem) Remove(cgroupPath string) error {
	return RemoveCgroupResourceConfig(s.Name(), cgroupPath)
}



