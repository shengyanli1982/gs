package gs

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type CloseType int8

const (
	ASyncClose CloseType = iota
	SyncClose
	ForceSyncClose
)

func waiting(mode CloseType, sigs ...*TerminateSignal) {
	// 创建一个 os.Signal 类型的通道，用于接收系统信号
	// Create a channel of type os.Signal to receive system signals
	quit := make(chan os.Signal, 1)

	// 注册我们关心的系统信号，当这些信号发生时，会发送到 quit 通道
	// Register the system signals we care about, when these signals occur, they will be sent to the quit channel
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

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

		case SyncClose:
			// 对每一个 TerminateSignal，同步进行关闭操作
			// For each TerminateSignal, perform the close operation synchronously
			for _, ts := range sigs {
				ts.Close(nil)
			}

		case ForceSyncClose:
			// 对每一个 TerminateSignal，强制同步进行关闭操作
			// For each TerminateSignal, forcibly perform the close operation synchronously
			for _, ts := range sigs {
				ts.SyncClose(nil)
			}

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
