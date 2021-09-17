package utils

import (
	"github.com/shirou/gopsutil/mem"
	"math"
	"os"
	"path/filepath"
	"runtime"
)

func GetMemUsedMB() uint64 {
	var (
		sts runtime.MemStats
		ret uint64
	)

	runtime.ReadMemStats(&sts)
	// 这里取了mem.Alloc
	ret = sts.HeapAlloc / 1024 / 1024
	return ret
}

func GetPwd() (string, error) {
	var (
		pwd string
		err error
	)

	if pwd, err = os.Executable(); err != nil {
		return "", err
	}

	return filepath.Dir(pwd), nil
}

func GetSysCPULimit(rate float64) int {
	var limit int

	limit = int(math.Ceil(float64(runtime.NumCPU()) * rate))
	if limit < 1 {
		limit = 1
	}

	return limit
}

func GetSysMemLimit(maxMemRate float64) int {
	var (
		m     *mem.VirtualMemoryStat
		total int
		limit int
		err   error
	)

	m, err = mem.VirtualMemory()
	if err != nil {
		limit = 512
	} else {
		total = int(m.Total / (1024 * 1024))
		limit = int(float64(total) * maxMemRate)
	}

	if limit < 512 {
		limit = 512
	}

	return limit
}
