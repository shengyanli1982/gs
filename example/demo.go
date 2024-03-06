package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shengyanli1982/gs"
)

// testTerminateSignal 是一个模拟服务的结构体
// testTerminateSignal is a struct simulating a service
type testTerminateSignal struct{}

// Close 是 testTerminateSignal 的一个方法，用于关闭服务
// Close is a method of testTerminateSignal for closing the service
func (t *testTerminateSignal) Close() {
	fmt.Println("testTerminateSignal.Close()")
}

// testTerminateSignal2 是另一个模拟服务的结构体
// testTerminateSignal2 is another struct simulating a service
type testTerminateSignal2 struct{}

// Shutdown 是 testTerminateSignal2 的一个方法，用于关闭服务
// Shutdown is a method of testTerminateSignal2 for shutting down the service
func (t *testTerminateSignal2) Shutdown() {
	fmt.Println("testTerminateSignal2.Shutdown()")
}

// testTerminateSignal3 是第三个模拟服务的结构体
// testTerminateSignal3 is the third struct simulating a service
type testTerminateSignal3 struct{}

// Terminate 是 testTerminateSignal3 的一个方法，用于终止服务
// Terminate is a method of testTerminateSignal3 for terminating the service
func (t *testTerminateSignal3) Terminate() {
	fmt.Println("testTerminateSignal3.Terminate()")
}

// main 函数是程序的入口点
// The main function is the entry point of the program
func main() {
	// 创建 TerminateSignal 实例
	// Create TerminateSignal instance
	s := gs.NewDefaultTerminateSignal()

	// 创建希望在服务终止时关闭的资源
	// Create resources which want to be closed when the service is terminated
	t1 := &testTerminateSignal{}
	t2 := &testTerminateSignal2{}
	t3 := &testTerminateSignal3{}

	// 注册希望在服务终止时关闭的资源的关闭方法
	// Register the close method of the resource which want to be closed when the service is terminated
	s.RegisterCancelCallback(t1.Close, t2.Shutdown, t3.Terminate)

	// 创建一个 goroutine，在 2 秒后向进程发送一个信号
	// Create a goroutine to send a signal to the process after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			fmt.Println(err.Error())
		}
		err = p.Signal(os.Interrupt)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	// 使用 WaitingForGracefulShutdown 方法等待 TerminateSignal 实例优雅地关闭
	// Use WaitingForGracefulShutdown method to wait for the TerminateSignal instance to shutdown gracefully
	gs.WaitingForGracefulShutdown(s)

	// "优雅地关闭"
	// "Shutdown gracefully"
	fmt.Println("shutdown gracefully")
}
