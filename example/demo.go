package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shengyanli1982/gs"
)

// 模拟一个服务
// Simulate a service
type testTerminateSignal struct{}

// Close 方法用于关闭 testTerminateSignal 服务
// The Close method is used to close the testTerminateSignal service
func (t *testTerminateSignal) Close() {
	fmt.Println("testTerminateSignal.Close()")
}

// 模拟一个服务
// Simulate a service
type testTerminateSignal2 struct{}

// Shutdown 方法用于关闭 testTerminateSignal2 服务
// The Shutdown method is used to close the testTerminateSignal2 service
func (t *testTerminateSignal2) Shutdown() {
	fmt.Println("testTerminateSignal2.Shutdown()")
}

// 模拟一个服务
// Simulate a service
type testTerminateSignal3 struct{}

// Terminate 方法用于关闭 testTerminateSignal3 服务
// The Terminate method is used to close the testTerminateSignal3 service
func (t *testTerminateSignal3) Terminate() {
	fmt.Println("testTerminateSignal3.Terminate()")
}

func main() {
	// 创建一个新的 TerminateSignal 实例
	// Create a new TerminateSignal instance
	s := gs.NewTerminateSignal()

	// 创建三个测试服务的实例
	// Create instances of three test services
	t1 := &testTerminateSignal{}
	t2 := &testTerminateSignal2{}
	t3 := &testTerminateSignal3{}

	// 注册需要在终止信号发生时执行的处理函数
	// Register the handle functions to be executed when the termination signal occurs
	s.RegisterCancelHandles(t1.Close, t2.Shutdown, t3.Terminate)

	// 在新的 goroutine 中执行一个函数
	// Execute a function in a new goroutine
	go func() {
		// 等待 2 秒
		// Wait for 2 seconds
		time.Sleep(2 * time.Second)

		// 查找当前进程
		// Find the current process
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			fmt.Println(err.Error())
		}

		// 向当前进程发送中断信号, os.Process.Signal() 对 Linux 和 MacOS 有效, Windows 无效
		// Send an interrupt signal to the current process, os.Process.Signal() is valid for Linux and MacOS, invalid for Windows
		err = p.Signal(os.Interrupt)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	// 等待所有的异步关闭信号
	// Wait for all asynchronous shutdown signals
	gs.WaitForAsync(s)

	// 打印一条消息，表示服务已经优雅地关闭
	// Print a message indicating that the service has been gracefully shut down
	fmt.Println("shutdown gracefully")
}
