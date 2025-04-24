package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/zboyco/s3ctl/internal/cmd"
)

func main() {
	// 创建一个可取消的context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号监听
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGINT,  // 中断信号
		syscall.SIGTERM, // 终止信号
	)

	// 在后台goroutine中监听信号
	go func() {
		sig := <-signalChan
		fmt.Printf("\n接收到信号: %s，程序退出...\n\n", sig)
		cancel() // 取消context
	}()

	if err := cmd.Execute(ctx); err != nil {
		os.Exit(1)
	}
}
