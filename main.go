package main

import (
	"context"
	"github.com/fanghongbo/log-agent/common/g"
	"github.com/fanghongbo/log-agent/task"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var (
		quit   chan os.Signal
		ctx    context.Context
		cancel context.CancelFunc
		err    error
	)

	if err = g.Init(); err != nil {
		g.AppLog.Fatal(err)
	}

	task.Start()

	// 等待中断信号以优雅地关闭服务（设置 5 秒的超时时间）
	quit = make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	<-quit

	g.AppLog.Info("ready to close service..")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		task.Stop()
	}()

	if err = g.Destroy(ctx); err != nil {
		g.AppLog.Fatalf("service close err: %v", err)
	} else {
		g.AppLog.Info("service close ..")
	}
}
