package gs

import (
	"context"
	"sync"
	"time"
)

// InfinityTerminateTimeout 是一个常量，表示无限大的超时时间。
// InfinityTerminateTimeout is a constant that represents an infinite timeout.
const InfinityTerminateTimeout = time.Duration(-1)

// TerminateSignal 是一个结构体，包含了上下文、取消函数、等待组和一组执行函数。
// TerminateSignal is a struct that includes a context, a cancel function, a wait group, and a set of execution functions.
type TerminateSignal struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	exec   []func()
	once   sync.Once
}

// NewTerminateSignalWithContext 是一个函数，它创建一个带有上下文和超时的 TerminateSignal 实例。
// NewTerminateSignalWithContext is a function that creates a TerminateSignal instance with a context and a timeout.
func NewTerminateSignalWithContext(ctx context.Context, timeout time.Duration) *TerminateSignal {
	t := TerminateSignal{
		wg:   sync.WaitGroup{},
		exec: []func(){},
		once: sync.Once{},
	}

	// 如果超时时间小于等于 0，则不设置超时时间。
	// If the timeout is less than or equal to 0, the timeout is not set.
	if int64(timeout) <= 0 {
		t.ctx, t.cancel = context.WithCancel(ctx) // 不设置超时时间 (No timeout)
	} else {
		t.ctx, t.cancel = context.WithTimeout(ctx, timeout) // 设置超时时间 (Set timeout)
	}

	return &t
}

// NewTerminateSignal 是一个函数，它创建一个带有超时的 TerminateSignal 实例。
// NewTerminateSignal is a function that creates a TerminateSignal instance with a timeout.
func NewTerminateSignal(timeout time.Duration) *TerminateSignal {
	return NewTerminateSignalWithContext(context.Background(), timeout)
}

// NewDefaultTerminateSignal 是一个函数，它创建一个默认的 TerminateSignal 实例，超时时间为无限大。
// NewDefaultTerminateSignal is a function that creates a default TerminateSignal instance with an infinite timeout.
func NewDefaultTerminateSignal() *TerminateSignal {
	return NewTerminateSignalWithContext(context.Background(), InfinityTerminateTimeout)
}

// RegisterCancelCallback 是一个方法，它注册需要取消的回调函数。
// RegisterCancelCallback is a method that registers the callback functions to be canceled.
func (s *TerminateSignal) RegisterCancelCallback(callbacks ...func()) {
	s.exec = append(s.exec, callbacks...)
}

// GetStopContext 是一个方法，它获取停止信号的 Context。
// GetStopContext is a method that gets the Context of the stop signal.
func (s *TerminateSignal) GetStopContext() context.Context {
	return s.ctx
}

// Close 是一个方法，它关闭 TerminateSignal 实例，并执行所有注册的回调函数。
// Close is a method that closes the TerminateSignal instance and executes all registered callback functions.
func (s *TerminateSignal) Close(wg *sync.WaitGroup) {
	// 使用 sync.Once 确保 Close 方法只被执行一次。
	// Use sync.Once to ensure that the Close method is only executed once.
	s.once.Do(func() {
		// 遍历并执行所有注册的回调函数。
		// Iterate over and execute all registered callback functions.
		for _, callback := range s.exec {
			if callback != nil {
				s.wg.Add(1)
				// 在新的 goroutine 中执行回调函数，以实现并发。
				// Execute the callback function in a new goroutine to achieve concurrency.
				go s.worker(callback)
			}
		}
		// 发送取消信号，使得所有使用该 Context 的 goroutine 都可以接收到取消信号。
		// Send the cancel signal so that all goroutines using this Context can receive the cancel signal.
		s.cancel()
		// 等待所有的回调函数都执行完毕。
		// Wait for all callback functions to finish executing.
		s.wg.Wait()
		// 如果 wg 不为 nil，那么调用 wg.Done() 方法，表示一个操作已经完成。
		// If wg is not nil, call the wg.Done() method to indicate that an operation has been completed.
		if wg != nil {
			wg.Done()
		}
	})
}

// worker 是一个方法，它在接收到关闭信号后执行回调函数。
// worker is a method that executes the callback function after receiving the shutdown signal.
func (s *TerminateSignal) worker(callback func()) {
	// 使用 defer 语句确保在函数结束时调用 wg.Done() 方法，表示一个操作已经完成。
	// Use the defer statement to ensure that the wg.Done() method is called when the function ends, indicating that an operation has been completed.
	defer s.wg.Done()
	select {
	// 当接收到关闭信号时，执行回调函数。
	// When the shutdown signal is received, execute the callback function.
	case <-s.ctx.Done():
		callback()
	default:
	}
}
