package gs

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// WaitingForGracefulShutdown 是一个函数，它等待所有的终止信号。
// WaitingForGracefulShutdown is a function that waits for all termination signals.
func WaitingForGracefulShutdown(sigs ...*TerminateSignal) {
	// 创建一个可以接收 os.Signal 类型的通道，用于接收系统信号。
	// Create a channel that can receive os.Signal types, used to receive system signals.
	quit := make(chan os.Signal, 1)

	// signal.Notify 函数使得我们可以将输入的信号转发到 quit 通道，这里我们监听了 SIGINT, SIGTERM, SIGQUIT 三种信号。
	// The signal.Notify function allows us to forward the input signals to the quit channel. Here we listen for three signals: SIGINT, SIGTERM, SIGQUIT.
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// 阻塞程序执行，直到 quit 通道接收到信号。
	// Block program execution until the quit channel receives a signal.
	<-quit

	// signal.Stop 函数使 quit 通道停止接收所有信号。
	// The signal.Stop function causes the quit channel to stop receiving all signals.
	signal.Stop(quit)

	// 关闭 quit 通道。
	// Close the quit channel.
	close(quit)

	// 如果 sigs 切片的长度大于 0，那么我们将执行关闭动作。
	// If the length of the sigs slice is greater than 0, then we will execute the shutdown action.
	if len(sigs) > 0 {
		// 创建一个 WaitGroup，用于等待所有的关闭动作完成。
		// Create a WaitGroup to wait for all shutdown actions to complete.
		wg := sync.WaitGroup{}
		wg.Add(len(sigs))

		// 遍历 sigs 切片，对每一个终止信号执行关闭动作。
		// Iterate over the sigs slice, performing the shutdown action for each termination signal.
		for _, ts := range sigs {
			// 在一个新的 goroutine 中执行关闭动作，这样可以并发地关闭所有的终止信号。
			// Perform the shutdown action in a new goroutine, so that all termination signals can be shut down concurrently.
			go ts.Close(&wg)
		}

		// 等待所有的关闭动作完成。
		// Wait for all shutdown actions to complete.
		wg.Wait()
	}
}
