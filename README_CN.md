[English](./README.md) | 中文

<div align="center">
	<img src="assets/logo.png" alt="logo" width="450px">
</div>

# 介绍

**优雅关闭** 是大多数服务的常见需求。当服务接收到终止信号时，优雅地关闭服务被认为是最佳实践。优雅关闭的过程通常包括以下步骤：

1. 创建一个`TerminateSignal`实例并注册所需的终止信号。
2. 注册在服务终止时需要关闭的资源。
3. 使用`WaitForAsync`、`WaitForSync`或`WaitForForceSync`方法等待`TerminateSignal`实例优雅地关闭。

> [!IMPORTANT]
> 强烈建议使用最新的通用版本**v0.1.3**的库。之前的版本存在重要的逻辑和控制问题，不再推荐使用。

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

### 方法

**创建**

-   `NewTerminateSignal`：创建一个新的 `TerminateSignal` 实例。
-   `NewTerminateSignalWithContext`：创建一个带有上下文的新的 `TerminateSignal` 实例。

> [!TIP]
> 可以使用 `InfinityTerminateTimeout` 值将超时信号设置为无限大，这意味着只有在调用 `Close` 方法并关闭注册的资源后，`TerminateSignal` 实例才会关闭。

**TerminateSignal**

-   `RegisterCancelCallback`：注册需要在服务终止时关闭的资源。
-   `GetStopContext`：获取 `TerminateSignal` 实例的上下文。
-   `Close`：异步关闭 `TerminateSignal` 实例。
-   `SyncClose`：同步关闭 `TerminateSignal` 实例。

**等待**

-   `WaitForAsync`：异步等待 `TerminateSignal` 实例优雅关闭。
-   `WaitForSync`：同步等待 `TerminateSignal` 实例优雅关闭。
-   `WaitForForceSync`：严格同步等待 `TerminateSignal` 实例优雅关闭。

> [!NOTE]
>
> -   使用 `WaitForAsync` 异步等待所有注册的资源关闭。
> -   使用 `WaitForForceSync` 严格同步等待所有注册的资源按照注册顺序关闭。执行顺序取决于注册的顺序。首次注册的函数将首先执行，然后是第二次注册的函数，依此类推，直到所有函数都执行完毕。
> -   使用 `WaitForSync` 同步等待所有注册的资源关闭。它逐个执行注册的函数，但在执行函数时，通过 `RegisterCancelCallback` 注册的内部函数是异步执行的。

> [!IMPORTANT] > `WaitingForGracefulShutdown` 方法自 v0.1.3 版本起已弃用。建议使用 `WaitForAsync`、`WaitForSync` 或 `WaitForForceSync` 方法代替。

### 示例

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

**执行结果**

```bash
$ go run demo.go
testTerminateSignal3.Terminate()
testTerminateSignal.Close()
testTerminateSignal2.Shutdown()
shutdown gracefully
```
