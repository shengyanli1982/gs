package gs

import (
	"context"
	"sync"
	"time"
)

// InfinityTerminateTimeout 无限大的超时时间 (Infinite timeout)
const InfinityTerminateTimeout = time.Duration(-1)

type TerminateSignal struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	exec   []func()
	once   sync.Once
}

// 创建一个带有上下文和超时的 TerminateSignal 实例
// Create a TerminateSignal instance with context and timeout
func NewTerminateSignalWithContext(ctx context.Context, timeout time.Duration) *TerminateSignal {
	t := TerminateSignal{
		wg:   sync.WaitGroup{},
		exec: []func(){},
		once: sync.Once{},
	}

	// 如果超时时间小于等于 0，则不设置超时时间 (If the timeout is less than or equal to 0, the timeout is not set)
	if int64(timeout) <= 0 {
		t.ctx, t.cancel = context.WithCancel(ctx) // 不设置超时时间 (No timeout)
	} else {
		t.ctx, t.cancel = context.WithTimeout(ctx, timeout) // 设置超时时间 (Set timeout)
	}

	return &t
}

// 创建一个带有超时的 TerminateSignal 实例
// Create a TerminateSignal instance with timeout
func NewTerminateSignal(timeout time.Duration) *TerminateSignal {
	return NewTerminateSignalWithContext(context.Background(), timeout)
}

// 创建一个默认的 TerminateSignal 实例，超时时间为无限大
// Create a default TerminateSignal instance with infinite timeout
func NewDefaultTerminateSignal() *TerminateSignal {
	return NewTerminateSignalWithContext(context.Background(), InfinityTerminateTimeout)
}

// 注册需要取消的回调函数
// Register the callback function to be canceled
func (s *TerminateSignal) RegisterCancelCallback(callbacks ...func()) {
	s.exec = append(s.exec, callbacks...)
}

// 获取停止信号的 Context
// Get the Context of the stop signal
func (s *TerminateSignal) GetStopContext() context.Context {
	return s.ctx
}

// Close 关闭 TerminateSignal 实例
// Close the TerminateSignal instance
func (s *TerminateSignal) Close(wg *sync.WaitGroup) {
	s.once.Do(func() {
		// 执行回调函数 (Execute the callback function)
		for _, callback := range s.exec {
			if callback != nil {
				s.wg.Add(1)
				go s.worker(callback) // 执行回调函数 (Execute the callback function)
			}
		}
		s.cancel()  // 发送关闭信号 (Send the shutdown signal)
		s.wg.Wait() // 等待关闭完成 (Wait for the shutdown is complete)
		if wg != nil {
			wg.Done() // 通知关闭完成 (Notify the shutdown is complete)
		}
	})
}

// worker 执行回调函数
func (s *TerminateSignal) worker(callback func()) {
	defer s.wg.Done() // 通知关闭完成 (Notify the shutdown is complete)
	select {
	case <-s.ctx.Done(): // 等待关闭信号 (Wait for the shutdown signal)
		callback() // 执行回调函数 (Execute the callback function)
	default:
	}
}
