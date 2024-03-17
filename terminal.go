package gs

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// TerminateSignal 结构体包含了一个 context，一个取消函数，一个等待组，一个函数切片和一个 sync.Once 实例
// The TerminateSignal struct contains a context, a cancel function, a wait group, a function slice, and a sync.Once instance
type TerminateSignal struct {
	// ctx 是一个 context.Context 实例，用于传递取消信号
	// ctx is a context.Context instance, used to pass cancellation signals
	ctx context.Context

	// cancel 是一个 context.CancelFunc 函数，用于取消 context
	// cancel is a context.CancelFunc function, used to cancel the context
	cancel context.CancelFunc

	// wg 是一个 sync.WaitGroup 实例，用于等待所有的 goroutine 完成
	// wg is a sync.WaitGroup instance, used to wait for all goroutines to complete
	wg sync.WaitGroup

	// exec 是一个函数切片，包含了所有需要在终止信号发生时执行的回调函数
	// exec is a function slice, containing all callback functions that need to be executed when the termination signal occurs
	exec []func()

	// once 是一个 sync.Once 实例，用于确保某个操作只执行一次
	// once is a sync.Once instance, used to ensure that an operation is only performed once
	once sync.Once

	// closed 是一个 atomic.Bool 实例，用于标记 TerminateSignal 是否已经关闭
	// closed is an atomic.Bool instance, used to mark whether the TerminateSignal is closed
	closed atomic.Bool
}

// NewTerminateSignalWithContext 创建一个带有上下文和超时的 TerminateSignal 实例
// NewTerminateSignalWithContext creates a TerminateSignal instance with context and timeout
func NewTerminateSignalWithContext(ctx context.Context) *TerminateSignal {
	// 初始化 TerminateSignal 结构体
	// Initialize the TerminateSignal struct
	t := TerminateSignal{
		// wg 是一个 sync.WaitGroup 实例，用于等待所有的 goroutine 完成
		// wg is a sync.WaitGroup instance, used to wait for all goroutines to complete
		wg: sync.WaitGroup{},

		// exec 是一个函数切片，包含了所有需要在终止信号发生时执行的回调函数
		// exec is a function slice, containing all callback functions that need to be executed when the termination signal occurs
		exec: make([]func(), 0),

		// once 是一个 sync.Once 实例，用于确保某个操作只执行一次
		// once is a sync.Once instance, used to ensure that an operation is only performed once
		once: sync.Once{},

		// closed 是一个 atomic.Bool 实例，用于标记 TerminateSignal 是否已经关闭
		// closed is an atomic.Bool instance, used to mark whether the TerminateSignal is closed
		closed: atomic.Bool{},
	}

	// 将 closed 的值设置为 false，表示 TerminateSignal 还没有关闭
	// Set the value of closed to false, indicating that the TerminateSignal is not closed yet
	t.closed.Store(false)

	t.ctx, t.cancel = context.WithCancel(ctx)

	// 返回 TerminateSignal 实例的指针
	// Return the pointer to the TerminateSignal instance
	return &t
}

// NewTerminateSignal 创建一个带有超时的 TerminateSignal 实例
// NewTerminateSignal creates a TerminateSignal instance with a timeout
func NewTerminateSignal() *TerminateSignal {
	// 使用 context.Background() 作为父 context，并设置超时时间
	// Use context.Background() as the parent context and set the timeout
	return NewTerminateSignalWithContext(context.Background())
}

// RegisterCancelCallback 注册需要取消的回调函数
// RegisterCancelCallback registers the callback functions to be canceled
func (s *TerminateSignal) RegisterCancelCallback(callbacks ...func()) {
	// 如果 TerminateSignal 已经关闭，那么直接返回
	// If the TerminateSignal is already closed, then return directly
	if s.closed.Load() {
		return
	}

	// 将回调函数添加到 s.exec 切片中
	// Add the callback functions to the s.exec slice
	s.exec = append(s.exec, callbacks...)
}

// GetStopContext 获取停止信号的 Context
// GetStopContext gets the Context of the stop signal
func (s *TerminateSignal) GetStopContext() context.Context {
	// 返回 s.ctx，即停止信号的 Context
	// Return s.ctx, which is the Context of the stop signal
	return s.ctx
}

// worker 是一个执行回调函数的方法
// worker is a method that executes the callback function
func (s *TerminateSignal) worker(fn func()) {
	// 在函数返回时，调用 Done 方法
	// Call the Done method when the function returns
	defer s.wg.Done()

	// 如果 s.ctx 已经超时的话，那么直接返回
	// If s.ctx has already timed out, then return directly
	if err := s.ctx.Err(); err != nil {
		if !errors.Is(err, context.Canceled) {
			return
		}
	}

	// 执行注册待执行的函数
	// Execute the registered function
	fn()
}

// close 关闭 TerminateSignal 实例
// close the TerminateSignal instance
func (s *TerminateSignal) close(closeMode CloseType, wg *sync.WaitGroup) {
	// 使用 sync.Once 确保 Close 只被执行一次
	// Use sync.Once to ensure Close is only executed once
	s.once.Do(func() {
		// 将 closed 的值设置为 true，表示 TerminateSignal 已经关闭
		// Set the value of closed to true, indicating that the TerminateSignal is closed
		s.closed.Store(true)

		// 遍历所有的回调函数
		// Iterate over all callback functions
		for _, fn := range s.exec {
			// 如果回调函数不为空
			// If the callback function is not null
			if fn != nil {
				// 增加等待组的计数，表示有一个新的任务需要等待完成
				// Increase the count of the wait group, indicating that there is a new task to wait for completion
				s.wg.Add(1)

				// 根据关闭模式进行不同的处理
				// Handle differently according to the close mode
				switch closeMode {
				case ASyncClose:
					// 在新的 goroutine 中执行 worker 函数，这样可以并发执行多个任务
					// Execute the worker function in a new goroutine, so that multiple tasks can be executed concurrently
					go s.worker(fn)

				case SyncClose:
					// 在当前 goroutine 中执行 worker 函数，这样可以保证任务按顺序执行
					// Execute the worker function in the current goroutine, so that tasks can be executed in order
					s.worker(fn)
				}
			}
		}

		// 取消 context
		// Cancel the context
		s.cancel()

		// 等待所有的 worker 完成
		// Wait for all workers to complete
		s.wg.Wait()

		// 如果外部的等待组不为空，调用 Done 方法
		// If the external wait group is not null, call the Done method
		if wg != nil {
			wg.Done()
		}
	})
}

// Close 方法异步关闭 TerminateSignal 实例
// The Close method asynchronously closes the TerminateSignal instance
func (s *TerminateSignal) Close(wg *sync.WaitGroup) {
	// 调用 close 方法，传入 ASyncClose 作为关闭模式和 wg 作为等待组
	// Call the close method, passing in ASyncClose as the close mode and wg as the wait group
	s.close(ASyncClose, wg)
}

// SyncClose 方法同步关闭 TerminateSignal 实例
// The SyncClose method synchronously closes the TerminateSignal instance
func (s *TerminateSignal) SyncClose(wg *sync.WaitGroup) {
	// 调用 close 方法，传入 SyncClose 作为关闭模式和 wg 作为等待组
	// Call the close method, passing in SyncClose as the close mode and wg as the wait group
	s.close(SyncClose, wg)
}
