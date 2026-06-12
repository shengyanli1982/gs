package gs

import (
	"context"
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

	// handles 是一个函数切片，包含了所有需要在终止信号发生时执行的处理函数
	// handles is a function slice, containing all handle functions that need to be executed when the termination signal occurs
	handles []func()

	// once 是一个 sync.Once 实例，用于确保某个操作只执行一次
	// once is a sync.Once instance, used to ensure that an operation is only performed once
	once sync.Once

	// mu 是一个 sync.Mutex 实例，用于保护 handles 切片的并发读写和 closed 检查的原子性
	// mu is a sync.Mutex instance, used to protect concurrent read/write of the handles slice and the atomicity of the closed check
	mu sync.Mutex

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

		// handles 是一个函数切片，包含了所有需要在终止信号发生时执行的处理函数
		// handles is a function slice, containing all handle functions that need to be executed when the termination signal occurs
		handles: make([]func(), 0),

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

	// 使用 context.WithCancel 创建一个新的 context 和取消函数
	// Use context.WithCancel to create a new context and cancel function
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

// RegisterCancelHandles 注册需要取消的处理函数
// RegisterCancelHandles registers the handle functions to be canceled
func (s *TerminateSignal) RegisterCancelHandles(handles ...func()) {
	// 加锁以确保 closed 检查和 handles append 的原子性，防止与 close() 产生竞态
	// Lock to ensure atomicity of closed check and handles append, preventing race with close()
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果 TerminateSignal 已经关闭，那么直接返回
	// If the TerminateSignal is already closed, then return directly
	if s.closed.Load() {
		return
	}

	// 将回调函数添加到 s.handles 切片中
	// Add the callback functions to the s.handles slice
	s.handles = append(s.handles, handles...)
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

	// 执行注册的回调函数（handler 是清理函数，应始终执行）
	// Execute the registered callback function (handlers are cleanup functions and should always execute)
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

		// 先取消 context，向持有 ctx 的后台工作通知关闭已开始
		// Cancel the context first to signal shutdown to workers holding ctx
		s.cancel()

		// 根据关闭模式进行不同的处理
		// Handle differently according to the close mode
		switch closeMode {
		// ASyncClose 表示异步关闭
		// ASyncClose indicates asynchronous close
		case ASyncClose:
			// 在 mutex 保护下直接遍历 handles 并分发 goroutine，无需拷贝快照
			// Iterate handles directly under mutex and dispatch goroutines, no snapshot copy needed
			s.mu.Lock()
			// 统计非空回调数量，批量增加等待组计数
			// Count non-nil callbacks and batch add to wait group
			n := 0
			for _, fn := range s.handles {
				if fn != nil {
					n++
				}
			}
			s.wg.Add(n)
			// 在新的 goroutine 中并发执行每个回调函数
			// Execute each callback function concurrently in a new goroutine
			for _, fn := range s.handles {
				if fn != nil {
					go s.worker(fn)
				}
			}
			s.mu.Unlock()

		// SyncClose 表示同步关闭
		// SyncClose indicates synchronous close
		case SyncClose:
			// 在 mutex 保护下获取 handles 快照，因为同步 handler 执行时间可能较长，不能持锁执行
			// Get a snapshot of handles under mutex, since sync handlers may take long and we cannot hold the lock
			s.mu.Lock()
			handles := make([]func(), len(s.handles))
			copy(handles, s.handles)
			s.mu.Unlock()

			// 统计非空回调数量，批量增加等待组计数
			// Count non-nil callbacks and batch add to wait group
			n := 0
			for _, fn := range handles {
				if fn != nil {
					n++
				}
			}
			s.wg.Add(n)
			// 在当前 goroutine 中按顺序执行每个回调函数
			// Execute each callback function in the current goroutine in order
			for _, fn := range handles {
				if fn != nil {
					s.worker(fn)
				}
			}
		}

		// 等待所有的 handler 完成
		// Wait for all handlers to complete
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
