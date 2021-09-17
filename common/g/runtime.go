package g

import (
	"fmt"
	"github.com/fanghongbo/log-agent/utils"
	"runtime"
	"time"
)

var (
	Pwd         string
	LocalIp     string
)

func initRuntime() error {
	var (
		osType string
		err    error
	)

	osType = runtime.GOOS

	if osType != "linux" && osType != "darwin" {
		return fmt.Errorf("current system %s is not supported", osType)
	}

	// 内存监控
	go memMonitor()

	// 当前主机默认的ip地址
	if LocalIp, err = utils.GetLocalDefaultIp(); err != nil {
		return err
	}

	// 当前程序运行路径
	if Pwd, err = utils.GetPwd(); err != nil {
		return err
	}

	return nil
}

func memMonitor() {
	var (
		nowMemUsedMB uint64
		maxMemMB     uint64
		rate         uint64
	)

	for {
		time.Sleep(time.Second * 10)

		nowMemUsedMB = utils.GetMemUsedMB()
		maxMemMB = uint64(utils.GetSysMemLimit(1))
		rate = (nowMemUsedMB * 100) / maxMemMB

		// 若超50%限制，打印 warning
		if rate > 50 {
			AppLog.Warnf("heap memory used rate, current: %d%%", rate)
		}

		// 超过100%，就退出了
		if rate > 100 {
			// 堆内存已超过限制，退出进程
			AppLog.Fatalf("heap memory size over limit. quit process.[used:%dMB][limit:%dMB][rate:%d]", nowMemUsedMB, maxMemMB, rate)
		}
	}
}
