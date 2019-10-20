package subsystems

import (
	"os"
	"path"
	"testing"
)

func TestMemoryCgroup(t *testing.T) {
	memorySubSystem := MemorySubSystem{}
	resConfig := ResourceConfig{
		MemoryLimit: "1000m",
	}

	testCgroup := "testmemlimit"

	if err := memorySubSystem.Set(testCgroup, &resConfig); nil != err {
		t.Fatalf("cgroup fail %v", err)
	}

	stat, _ := os.Stat(path.Join(FindCgroupMountpoint("memory"), testCgroup))

	t.Logf("cgroup stats: %+v", stat)

	if err := memorySubSystem.Apply(testCgroup, os.Getpid()); nil != err {
		t.Fatalf("cgroup Apply %v", err)
	}

	//move pid to cgroup root
	if err := memorySubSystem.Apply("", os.Getpid()); nil != err {
		t.Fatalf("cgroup apply %v", err)
	}

	if err := memorySubSystem.Remove(testCgroup); nil != err {
		t.Fatalf("cgroup remove %v", err)
	}
}
