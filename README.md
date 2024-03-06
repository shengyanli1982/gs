<div align="center">
	<h1>GS</h1>
	<p>A lightweight generic graceful shutdown component<p>
	<img src="assets/logo.png" alt="logo" width="300px">
</div>

# Introduction

**Graceful Shutdown** is a common requirement for most services. It is a good practice to gracefully shutdown a service when it receives a termination signal. The process typically involves the following steps:

1. Create a `TerminateSignal` instance and register the desired termination signal.
2. Register the resources that need to be closed when the service is terminated.
3. Use the `WaitingForGracefulShutdown` method to wait for the `TerminateSignal` instance to gracefully shutdown.

# Advantage

-   Simple and user-friendly
-   No external dependencies required
-   Efficient memory usage
-   Supports timeout signals
-   Supports context
-   Handles multiple signals

# Installation

```bash
go get github.com/shengyanli1982/gs
```

# Quick Start

`GS` is a lightweight and user-friendly library for graceful shutdown in Go. With `TerminateSignal` instances, you can easily register resources to be closed when the service is terminated. Use the `WaitingForGracefulShutdown` method to wait for the `TerminateSignal` instances to gracefully shutdown.

### Methods

**Create**

-   `NewTerminateSignal`: Create a new `TerminateSignal` instance
-   `NewDefaultTerminateSignal`: Create a new `TerminateSignal` instance with default signals
-   `NewTerminateSignalWithContext`: Create a new `TerminateSignal` instance with context

> [!TIP]
> The `InfinityTerminateTimeout` value sets the timeout signal to infinity. This means that the `TerminateSignal` instance will not be closed until the `Close` method is called and the resources registered in the `TerminateSignal` instance are closed.

**TerminateSignal**

-   `RegisterCancelCallback`: Register resources to be closed when the service is terminated
-   `GetStopContext`: Get the context of the `TerminateSignal` instance
-   `Close`: Close the `TerminateSignal` instance

**Waiting**

-   `WaitingForGracefulShutdown`: Use this method to wait for all `TerminateSignal` instances to gracefully shutdown

### Example

```go
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
```

**Result**

```bash
$ go run demo.go
testTerminateSignal.Close()
testTerminateSignal2.Shutdown()
testTerminateSignal3.Terminate()
shutdown gracefully
```

# Features

`GS` provides a few but sufficient features for most services.

## Timeout Signal

A `TerminateSignal` instance can be created with a timeout signal. When the timeout signal is received, the `TerminateSignal` instance will be closed without waiting for the registered resources to be closed.

> [!TIP]
> Using a timeout signal can address the issue of a service not being able to close due to a resource that cannot be closed. However, it is not recommended to rely on timeout signals as they may result in abnormal closure of resources.

### Example

```go
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

// Close 是 testTerminateSignal 的一个方法，用于关闭服务，并且会等待5秒钟
// Close is a method of testTerminateSignal for closing the service, and it will wait for 5 seconds
func (t *testTerminateSignal) Close() {
	time.Sleep(5 * time.Second)
}

// main 函数是程序的入口点
// The main function is the entry point of the program
func main() {
	// 创建 TerminateSignal 实例，超时信号设置为1秒
	// Create TerminateSignal instance, the timeout signal is set to 1 second
	s := gs.NewTerminateSignal(time.Second)

	// 创建希望在服务终止时关闭的资源
	// Create a resource which want to be closed when the service is terminated
	t1 := &testTerminateSignal{}

	// 注册希望在服务终止时关闭的资源的关闭方法
	// Register the close method of the resource which want to be closed when the service is terminated
	s.RegisterCancelCallback(t1.Close)

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
```
