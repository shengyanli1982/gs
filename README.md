<div align="center">
	<img src="assets/logo.png" alt="logo" width="450px">
</div>

# Introduction

**Graceful Shutdown** is a common requirement for most services. It is considered a best practice to gracefully shut down a service when it receives a termination signal. The process of graceful shutdown typically involves the following steps:

1. Create a `TerminateSignal` instance and register the desired termination signal.
2. Register the resources that need to be closed when the service is terminated.
3. Use the `WaitForAsync`, `WaitForSync`, or `WaitForForceSync` method to wait for the `TerminateSignal` instance to gracefully shut down.

> [!IMPORTANT]
> It is strongly recommended to use the latest general version **v0.1.3** of the library. Previous versions have significant logic and control issues and are no longer recommended.

# Advantages

-   Simple and user-friendly
-   No external dependencies required
-   Low memory footprint
-   Supports timeout signals
-   Supports context
-   Supports multiple signals

# Installation

```bash
go get github.com/shengyanli1982/gs
```

# Quick Start

`GS` is a simple and easy-to-use library for graceful shutdown. To use it, follow these steps:

1. Create a `TerminateSignal` instance.
2. Register the resources that need to be closed when the service is terminated.
3. Use the appropriate waiting method (`WaitForAsync`, `WaitForSync`, or `WaitForForceSync`) to wait for the `TerminateSignal` instance to gracefully shut down.

### Methods

**Create**

-   `NewTerminateSignal`: Create a new `TerminateSignal` instance.
-   `NewTerminateSignalWithContext`: Create a new `TerminateSignal` instance with context.

> [!TIP]
> The `InfinityTerminateTimeout` value can be used to set the timeout signal to infinity, meaning that the `TerminateSignal` instance will not be closed until the `Close` method is called and the registered resources are closed.

**TerminateSignal**

-   `RegisterCancelCallback`: Register the resources that need to be closed when the service is terminated.
-   `GetStopContext`: Get the context of the `TerminateSignal` instance.
-   `Close`: Close the `TerminateSignal` instance asynchronously.
-   `SyncClose`: Close the `TerminateSignal` instance synchronously.

**Waiting**

-   `WaitForAsync`: Wait for the `TerminateSignal` instance to gracefully shut down asynchronously.
-   `WaitForSync`: Wait for the `TerminateSignal` instance to gracefully shut down synchronously.
-   `WaitForForceSync`: Wait for the `TerminateSignal` instance to gracefully shut down strict synchronously.

> [!NOTE]
>
> -   Use `WaitForAsync` to asynchronously wait for all registered resources to be closed.
> -   Use `WaitForForceSync` to synchronously wait for all registered resources to be closed in strict order. The execution order depends on the order of registration. The functions registered with `RegisterCancelCallback` in the first registration will be executed first, followed by the second registration, and so on, until all functions have been executed.
> -   Use `WaitForSync` to synchronously wait for all registered resources to be closed. It executes the registered functions one by one, but when executing a function, the internal functions registered through `RegisterCancelCallback` are executed asynchronously.

> [!IMPORTANT]
> The `WaitingForGracefulShutdown` method is deprecated since v0.1.3. It is recommended to use the `WaitForAsync`, `WaitForSync`, or `WaitForForceSync` methods instead.

### Example

```go
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

	// 注册需要在终止信号发生时执行的回调函数
	// Register the callback functions to be executed when the termination signal occurs
	s.RegisterCancelCallback(t1.Close, t2.Shutdown, t3.Terminate)

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

		// 向当前进程发送中断信号
		// Send an interrupt signal to the current process
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
```

**Result**

```bash
$ go run demo.go
testTerminateSignal3.Terminate()
testTerminateSignal.Close()
testTerminateSignal2.Shutdown()
shutdown gracefully
```
