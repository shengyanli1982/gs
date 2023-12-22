package gs

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// 等待所有关闭信号
// Wait for all shutdown signals
func WaitingForGracefulShutdown(sigs ...*TerminateSignal) {
	quit := make(chan os.Signal, 1)                                       // 创建一个接收信号的通道 (Create a channel to receive signals)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) // 注册要接收的信号 (Register the signals to receive)
	<-quit                                                                // 等待接收信号 (Wait for the signal to receive)
	signal.Stop(quit)                                                     // 停止接收信号 (Stop receiving signals)
	close(quit)                                                           // 关闭通道 (Close the channel)
	if len(sigs) > 0 {                                                    // 执行关闭动作 (Execute the shutdown action)
		wg := sync.WaitGroup{}
		wg.Add(len(sigs))
		// 批量执行 TerminateSignal 实例的关闭动作 (Batch execution of the shutdown action of each TerminateSignal instance)
		for _, s := range sigs {
			go s.Close(&wg) // 执行信号的关闭动作，wg 计数器减一 (Execute the shutdown action of each signal, wg counter minus one)
		}
		wg.Wait()
	}
}
