package gs

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// CloseType 是一个 int8 类型的别名，用于表示关闭类型
// CloseType is an alias for int8, used to represent the type of closure
type CloseType int8

// 定义了三种关闭类型：异步关闭、同步关闭和强制同步关闭
// Three types of closure are defined: asynchronous closure, synchronous closure, and forced synchronous closure
const (
	// ASyncClose 表示异步关闭，即在新的 goroutine 中执行关闭操作
	// ASyncClose represents asynchronous closure, i.e., the closure operation is performed in a new goroutine
	ASyncClose CloseType = iota

	// SyncClose 表示同步关闭，即在不同的 TerminateSignal 中同步执行关闭操作, eg: t1.Close() then t2.Close() then t3.Close()
	// 在每个 TerminateSignal 中，是异步执行的
	// SyncClose represents synchronous closure, i.e., the closure operation is performed synchronously in different TerminateSignal, eg: t1.Close() then t2.Close() then t3.Close()
	// In each TerminateSignal, it is asynchronous
	SyncClose

	// ForceSyncClose 表示强制同步关闭，即在不同的 TerminateSignal 中同步执行关闭操作, eg: t1.Close() then t2.Close() then t3.Close()
	// 在每个 TerminateSignal 中，是完全同步执行的
	// ForceSyncClose represents forced synchronous closure, i.e., the closure operation is performed synchronously in different TerminateSignal, eg: t1.Close() then t2.Close() then t3.Close()
	// In each TerminateSignal, it is completely synchronous
	ForceSyncClose
)

// waiting 函数用于等待系统信号，并根据关闭模式和 TerminateSignal 进行不同的处理
// The waiting function waits for system signals and handles them differently according to the close mode and TerminateSignal
func waiting(mode CloseType, sigs ...*TerminateSignal) {
	// 创建一个 os.Signal 类型的通道，用于接收系统信号
	// Create a channel of type os.Signal to receive system signals
	quit := make(chan os.Signal, 1)

	// 注册我们关心的系统信号，当这些信号发生时，会发送到 quit 通道
	// Register the system signals we care about, when these signals occur, they will be sent to the quit channel
	signal.Notify(quit, syscall.SIGINT, syscall.SIGINT, syscall.SIGQUIT)

	// 阻塞等待任何系统信号
	// Block and wait for any system signal
	<-quit

	// 停止接收更多的系统信号
	// Stop receiving more system signals
	signal.Stop(quit)

	// 关闭 quit 通道
	// Close the quit channel
	close(quit)

	// 如果有提供 TerminateSignal，那么就等待它们全部关闭
	// If TerminateSignal is provided, then wait for all of them to close
	if len(sigs) > 0 {
		// 根据关闭模式进行不同的处理
		// Handle differently according to the close mode
		switch mode {
		// ASyncClose 表示异步关闭
		// ASyncClose indicates asynchronous close
		case ASyncClose:
			// 创建一个 WaitGroup，用于等待所有的 TerminateSignal 关闭
			// Create a WaitGroup to wait for all TerminateSignal to close
			wg := sync.WaitGroup{}

			// 添加等待的数量
			// Add the number of waits
			wg.Add(len(sigs))

			// 对每一个 TerminateSignal，启动一个 goroutine 进行关闭操作
			// For each TerminateSignal, start a goroutine to perform the close operation
			for _, ts := range sigs {
				go ts.Close(&wg)
			}

			// 等待所有的 TerminateSignal 都关闭
			// Wait for all TerminateSignal to close
			wg.Wait()

		// SyncClose 表示同步关闭
		// SyncClose indicates synchronous close
		case SyncClose:
			// 对每一个 TerminateSignal，同步进行关闭操作
			// For each TerminateSignal, perform the close operation synchronously
			for _, ts := range sigs {
				ts.Close(nil)
			}

		// ForceSyncClose 表示强制同步关闭
		// ForceSyncClose indicates forced synchronous close
		case ForceSyncClose:
			// 对每一个 TerminateSignal，强制同步进行关闭操作
			// For each TerminateSignal, forcibly perform the close operation synchronously
			for _, ts := range sigs {
				ts.SyncClose(nil)
			}

		// 默认行为
		// Default behavior
		default:
			// 默认情况下，不进行任何操作
			// By default, do nothing
		}
	}
}

// WaitForAsync 函数等待所有的异步关闭信号
// The WaitForAsync function waits for all asynchronous shutdown signals
func WaitForAsync(sigs ...*TerminateSignal) {
	// 调用 waiting 函数，传入 ASyncClose 作为关闭模式和 sigs 作为关闭信号
	// Call the waiting function, passing in ASyncClose as the close mode and sigs as the close signals
	waiting(ASyncClose, sigs...)
}

// WaitForSync 函数等待所有的同步关闭信号
// The WaitForSync function waits for all synchronous shutdown signals
func WaitForSync(sigs ...*TerminateSignal) {
	// 调用 waiting 函数，传入 SyncClose 作为关闭模式和 sigs 作为关闭信号
	// Call the waiting function, passing in SyncClose as the close mode and sigs as the close signals
	waiting(SyncClose, sigs...)
}

// WaitForForceSync 函数等待所有的强制同步关闭信号
// The WaitForForceSync function waits for all forced synchronous shutdown signals
func WaitForForceSync(sigs ...*TerminateSignal) {
	// 调用 waiting 函数，传入 ForceSyncClose 作为关闭模式和 sigs 作为关闭信号
	// Call the waiting function, passing in ForceSyncClose as the close mode and sigs as the close signals
	waiting(ForceSyncClose, sigs...)
}
