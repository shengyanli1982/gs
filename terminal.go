package gs

import (
	"context"
	"sync"
	"sync/atomic"
)

// TerminateSignal 结构体包含了一个 context、一个取消函数、一个等待组、一个函数切片、一个完成通道、一个互斥锁和一个原子关闭标记
// The TerminateSignal struct contains a context, a cancel function, a wait group, a function slice, a done channel, a mutex, and an atomic closed flag
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

	// done 是一个在关闭流程完成后被关闭的通道，用于向重复或并发调用 Close 的调用者广播关闭已完成
	// done is a channel closed after the shutdown sequence completes, used to broadcast completion to repeated or concurrent Close callers
	done chan struct{}

	// mu 是一个 sync.Mutex 实例，用于保护 handles 切片的并发读写和 closed 检查的原子性
	// mu is a sync.Mutex instance, used to protect concurrent read/write of the handles slice and the atomicity of the closed check
	mu sync.Mutex

	// closed 是一个 atomic.Bool 实例，用于标记 TerminateSignal 是否已经关闭
	// closed is an atomic.Bool instance, used to mark whether the TerminateSignal is closed
	closed atomic.Bool
}

// NewTerminateSignalWithContext 创建一个带有父上下文的 TerminateSignal 实例
// NewTerminateSignalWithContext creates a TerminateSignal instance with a parent context
func NewTerminateSignalWithContext(ctx context.Context) *TerminateSignal {
	// 初始化 TerminateSignal 结构体；wg、handles、closed 省略显式初始化，
	// 它们的零值（空 WaitGroup、nil 切片、false）即为正确初始状态，可减少冗余分配
	// Initialize the TerminateSignal struct; explicit initialization of wg, handles and closed is omitted,
	// since their zero values (empty WaitGroup, nil slice, false) are already the correct initial state, reducing redundant allocations
	t := TerminateSignal{
		// done 是一个在关闭流程完成后被关闭的通道，用于向重复或并发调用 Close 的调用者广播关闭已完成
		// done is a channel closed after the shutdown sequence completes, used to broadcast completion to repeated or concurrent Close callers
		done: make(chan struct{}),
	}

	// 使用 context.WithCancel 创建一个新的 context 和取消函数
	// Use context.WithCancel to create a new context and cancel function
	t.ctx, t.cancel = context.WithCancel(ctx)

	// 返回 TerminateSignal 实例的指针
	// Return the pointer to the TerminateSignal instance
	return &t
}

// NewTerminateSignal 创建一个 TerminateSignal 实例
// NewTerminateSignal creates a TerminateSignal instance
func NewTerminateSignal() *TerminateSignal {
	// 使用 context.Background() 作为父 context
	// Use context.Background() as the parent context
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

	// 首次注册时预分配底层数组，消灭逐个注册导致的阶梯式扩容；
	// 容量下限 8 覆盖典型 1-8 个 handler 的场景，批量注册超过 8 个时按 len 预留避免多余增长。
	// close() 会将 handles 置为 nil，但 close 后的注册会被上面的 closed 检查拒绝，
	// 因此 handles == nil 仅表示"从未注册过"
	// Preallocate the backing array on the first registration to eliminate staircase growth
	// caused by registering handlers one at a time; the minimum capacity of 8 covers the
	// typical 1-8 handler case, and for batches larger than 8 we reserve len(handles) to
	// avoid extra growth. close() swaps handles to nil, but registration after close is
	// rejected by the closed check above, so handles == nil only means "never registered"
	if s.handles == nil {
		c := len(handles)
		if c < 8 {
			c = 8
		}
		s.handles = make([]func(), 0, c)
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

// close 关闭 TerminateSignal 实例。通过原子 CAS 选出唯一的赢家执行实际关闭流程，
// 其余调用者阻塞等待关闭完成；每个调用者传入的外部等待组（不为 nil 时）都会被 Done 恰好一次
// close the TerminateSignal instance. An atomic CAS elects a single winner that performs the
// actual shutdown sequence while all other callers block until the shutdown completes; the
// external wait group of every caller (when non-nil) is Done exactly once
func (s *TerminateSignal) close(closeMode CloseType, wg *sync.WaitGroup) {
	// 使用 CAS 将 closed 从 false 置为 true，成功者即赢家，执行实际关闭流程
	// Use CAS to flip closed from false to true; the caller that succeeds is the winner and performs the actual shutdown sequence
	if s.closed.CompareAndSwap(false, true) {
		// 先取消 context，向持有 ctx 的后台工作通知关闭已开始
		// Cancel the context first to signal shutdown to workers holding ctx
		s.cancel()

		// 在锁保护下取走 handles（swap 而非 copy：零分配，且允许 GC 回收 handler 闭包）；
		// CAS 前已持 mu 的并发注册在释放锁前完成 append，因此仍会被包含
		// Take handles under lock (swap rather than copy: zero allocation and allows GC to reclaim handler closures);
		// concurrent registrations that held mu before the CAS finish their append before releasing the lock, so they are still included
		s.mu.Lock()
		handles := s.handles
		s.handles = nil
		s.mu.Unlock()

		// 根据关闭模式进行不同的处理
		// Handle differently according to the close mode
		switch closeMode {
		// ASyncClose 表示异步关闭
		// ASyncClose indicates asynchronous close
		case ASyncClose:
			// 统计非空回调数量，批量增加等待组计数
			// Count non-nil callbacks and batch add to wait group
			n := 0
			for _, fn := range handles {
				if fn != nil {
					n++
				}
			}
			s.wg.Add(n)

			// 在新的 goroutine 中并发执行每个回调函数
			// Execute each callback function concurrently in a new goroutine
			for _, fn := range handles {
				if fn != nil {
					go s.worker(fn)
				}
			}

			// 等待所有的 handler 完成（仅异步路径需要等待；同步路径不经过 s.wg，计数恒为 0，移出后行为不变）
			// Wait for all handlers to complete (only the async path needs to wait; the sync path bypasses s.wg so its counter is always 0, and moving the wait here changes no behavior)
			s.wg.Wait()

		// SyncClose 表示同步关闭
		// SyncClose indicates synchronous close
		case SyncClose:
			// 顺序执行无并发协调需求，绕过 wg：在当前 goroutine 中按注册顺序直接调用每个回调函数，
			// 消除等待组原子操作（Add/Done）在顺序执行下的纯开销（pprof 实证）
			// Sequential execution has no concurrency-coordination need, so bypass wg: call each callback
			// directly in the current goroutine in registration order, eliminating the pure overhead of
			// wait-group atomic operations (Add/Done) on the sequential path (confirmed by pprof)
			for _, fn := range handles {
				if fn != nil {
					fn()
				}
			}

		// 其余取值不可达：close 是内部函数，当前 2 个调用点均传入合法模式；
		// 此处显式 panic，消除"误传非法模式时静默丢弃全部 handler"的维护陷阱
		// Any other value is unreachable: close is an internal function and all 2 current call sites
		// pass a valid mode; panic explicitly here to eliminate the maintenance trap of silently
		// dropping all handlers if an invalid mode were ever passed
		default:
			panic("gs: unreachable close mode")
		}

		// 关闭 done 通道，广播关闭已完成，唤醒所有等待中的输家
		// Close the done channel to broadcast completion and wake up all waiting losers
		close(s.done)
	} else {
		// 输家：阻塞等待赢家完成关闭流程（替代 sync.Once 的阻塞语义）
		// Loser: block until the winner finishes the shutdown sequence (replaces the blocking semantics of sync.Once)
		<-s.done
	}

	// 每个调用者都对自己传入的外部等待组调用 Done（修复此前只有首个调用者的等待组会被 Done 的问题）
	// Every caller calls Done on its own external wait group (fixes the issue that only the first caller's wait group was Done)
	if wg != nil {
		wg.Done()
	}
}

// Close 方法异步关闭 TerminateSignal 实例。Close 是幂等的：首个调用者执行实际关闭流程，
// 并发或后续的调用会阻塞直到关闭完成后才返回；每个调用者传入的外部等待组（不为 nil 时）都会被 Done 恰好一次。
// 注意：在 handler 内部回调同一实例的 Close 或 SyncClose 属于不支持的用法，会导致死锁。
// 注意：handler panic 会沿调用栈传播（Close 不做 recover），此时 done 通道不会被关闭；
// 若外层自行 recover 吞掉 panic，后续或并发的 Close 调用将永久阻塞。
// The Close method asynchronously closes the TerminateSignal instance. Close is idempotent: the first
// caller performs the actual shutdown sequence, while concurrent or subsequent callers block until the
// shutdown completes; the external wait group of every caller (when non-nil) is Done exactly once.
// Note: calling Close or SyncClose on the same instance from inside a handler is unsupported and will deadlock.
// Note: a panic inside a handler propagates up the call stack (Close performs no recover) and the done
// channel is never closed; if the caller recovers the panic itself, subsequent or concurrent Close
// calls will block forever.
func (s *TerminateSignal) Close(wg *sync.WaitGroup) {
	// 调用 close 方法，传入 ASyncClose 作为关闭模式和 wg 作为等待组
	// Call the close method, passing in ASyncClose as the close mode and wg as the wait group
	s.close(ASyncClose, wg)
}

// SyncClose 方法同步关闭 TerminateSignal 实例（handler 在当前 goroutine 中按注册顺序执行）。
// SyncClose 是幂等的：首个调用者执行实际关闭流程，并发或后续的调用会阻塞直到关闭完成后才返回；
// 每个调用者传入的外部等待组（不为 nil 时）都会被 Done 恰好一次。
// 注意：在 handler 内部回调同一实例的 Close 或 SyncClose 属于不支持的用法，会导致死锁。
// 注意：handler panic 会沿调用栈传播（SyncClose 不做 recover），此时 done 通道不会被关闭；
// 若外层自行 recover 吞掉 panic，后续或并发的 Close/SyncClose 调用将永久阻塞。
// The SyncClose method synchronously closes the TerminateSignal instance (handlers run in registration
// order in the current goroutine). SyncClose is idempotent: the first caller performs the actual shutdown
// sequence, while concurrent or subsequent callers block until the shutdown completes; the external wait
// group of every caller (when non-nil) is Done exactly once.
// Note: calling Close or SyncClose on the same instance from inside a handler is unsupported and will deadlock.
// Note: a panic inside a handler propagates up the call stack (SyncClose performs no recover) and the done
// channel is never closed; if the caller recovers the panic itself, subsequent or concurrent
// Close/SyncClose calls will block forever.
func (s *TerminateSignal) SyncClose(wg *sync.WaitGroup) {
	// 调用 close 方法，传入 SyncClose 作为关闭模式和 wg 作为等待组
	// Call the close method, passing in SyncClose as the close mode and wg as the wait group
	s.close(SyncClose, wg)
}
