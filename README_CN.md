[English](./README.md) | 中文

<div align="center">
	<img src="assets/logo.png" alt="logo" width="450px">
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/shengyanli1982/gs)](https://goreportcard.com/report/github.com/shengyanli1982/gs)
[![Build Status](https://github.com/shengyanli1982/gs/actions/workflows/test.yaml/badge.svg)](https://github.com/shengyanli1982/gs/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/shengyanli1982/gs.svg)](https://pkg.go.dev/github.com/shengyanli1982/gs)

# 介绍

**优雅关闭** 是大多数服务的常见需求。当服务接收到终止信号时，优雅地关闭服务被认为是一种最佳实践。优雅关闭的过程通常包括以下步骤：

1. 创建一个 `TerminateSignal` 实例并注册所需的终止信号。
2. 注册在服务终止时需要关闭的资源。
3. 使用 `WaitForAsync`、`WaitForSync` 或 `WaitForForceSync` 方法等待 `TerminateSignal` 实例优雅地关闭。

> [!IMPORTANT]
>
> 强烈建议使用 **v0.1.3** 之后最新稳定的版本。之前的版本存在重要的逻辑和控制问题，不再推荐使用。

# 优势

-   简单易用
-   无需外部依赖
-   内存占用低
-   支持超时信号
-   支持上下文
-   支持多个信号

# 安装

```bash
go get github.com/shengyanli1982/gs
```

# 快速入门

`GS` 是一个简单易用的优雅关闭库。使用它的步骤如下：

1. 创建一个 `TerminateSignal` 实例。
2. 注册需要在服务终止时关闭的资源。
3. 使用适当的等待方法（`WaitForAsync`、`WaitForSync` 或 `WaitForForceSync`）等待 `TerminateSignal` 实例优雅关闭。

> [!IMPORTANT]
>
> 如果您在 `Windows` 上使用 `GS`，只能使用于 `console` 应用程序。

### 方法

**创建实例**

-   `NewTerminateSignal`：创建一个新的 `TerminateSignal` 实例。
-   `NewTerminateSignalWithContext`：创建一个带有上下文的新的 `TerminateSignal` 实例。

**终结信号**

-   `RegisterCancelHandles`：注册需要在服务终止时关闭的资源。
-   `GetStopContext`：获取 `TerminateSignal` 实例的上下文。
-   `Close`：异步关闭 `TerminateSignal` 实例。
-   `SyncClose`：同步关闭 `TerminateSignal` 实例。

**等待**

-   `WaitForAsync`：异步等待 `TerminateSignal` 实例优雅关闭。
-   `WaitForSync`：同步等待 `TerminateSignal` 实例优雅关闭。
-   `WaitForForceSync`：严格同步等待 `TerminateSignal` 实例优雅关闭。

> [!NOTE]
>
> **`同步关闭 (SyncClose)` 和 `严格同步关闭 (ForceSyncClose)` 的区别**
>
> ```go
> // SyncClose 表示同步关闭，即在不同的 TerminateSignal 中同步执行关闭操作, eg: t1.Close() then t2.Close() >then t3.Close()
> // 在每个 TerminateSignal 中，是异步执行的
> // SyncClose represents synchronous closure, i.e., the closure operation is performed synchronously >in different TerminateSignal, eg: t1.Close() then t2.Close() then t3.Close()
> // In each TerminateSignal, it is asynchronous
> SyncClose
>
> // ForceSyncClose 表示强制同步关闭，即在不同的 TerminateSignal 中同步执行关闭操作, eg: t1.Close() then t2.>Close() then t3.Close()
> // 在每个 TerminateSignal 中，是完全同步执行的
> // ForceSyncClose represents forced synchronous closure, i.e., the closure operation is performed >synchronously in different TerminateSignal, eg: t1.Close() then t2.Close() then t3.Close()
> // In each TerminateSignal, it is completely synchronous
> ForceSyncClose
> ```
>
> `ForceSyncClose` 是完全同步的，而 `SyncClose` 在每个 `TerminateSignal` 中是异步的。

> [!IMPORTANT]
>
> 自 v0.1.3 版本起 `WaitingForGracefulShutdown` 方法已弃用。建议使用 `WaitForAsync`、`WaitForSync` 或 `WaitForForceSync` 方法代替。

### 示例

> [!TIP]
>
> `os.Process.Signal()` 方法只适用于 Linux 和 MacOS，对于 Windows 平台无效。如果您想在 Windows 上测试代码，请参考 `gracefull_windows_test.go` 文件。

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
```

**执行结果**

```bash
$ go run demo.go
testTerminateSignal3.Terminate()
testTerminateSignal.Close()
testTerminateSignal2.Shutdown()
shutdown gracefully
```
