package gs

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestTerminateSignal struct {
	name string
}

func (t *TestTerminateSignal) Close() {
	fmt.Println(">>>>: " + t.name + " -> TestTerminateSignal.Close()")
}

func NewTestTerminateSignal(name string) *TestTerminateSignal {
	return &TestTerminateSignal{name: name}
}

func TestTerminateSignal_Standard(t *testing.T) {
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")
	tts := NewTestTerminateSignal("test")
	sig.RegisterCancelHandles(tts.Close)
	wg := sync.WaitGroup{}
	wg.Add(1)
	sig.Close(&wg)
	wg.Wait()
}

func TestTerminateSignal_WithContext(t *testing.T) {
	ctx := context.Background()
	sig := NewTerminateSignalWithContext(ctx)
	assert.NotNil(t, sig, "signal is nil")
	tts := NewTestTerminateSignal("test")
	sig.RegisterCancelHandles(tts.Close)
	wg := sync.WaitGroup{}
	wg.Add(1)
	sig.Close(&wg)
	wg.Wait()
}

func TestTerminateSignal_MultiRegisters(t *testing.T) {
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")
	assert.Equal(t, sig.GetStopContext().Err(), nil)
	for i := 0; i < 11; i++ {
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
	}
	sig.Close(nil)
}

func TestTerminateSignal_MultiRegisters_Sync(t *testing.T) {
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")
	assert.Equal(t, sig.GetStopContext().Err(), nil)
	for i := 0; i < 10; i++ {
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
	}
	sig.SyncClose(nil)
}

// TestTerminateSignal_ConcurrentRegister_Race 验证并发调用 RegisterCancelHandles 不存在数据竞争
// TestTerminateSignal_ConcurrentRegister_Race verifies that concurrent calls to RegisterCancelHandles have no data race
func TestTerminateSignal_ConcurrentRegister_Race(t *testing.T) {
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")

	var wg sync.WaitGroup
	const n = 100

	wg.Add(n)
	for i := 0; i < n; i++ {
		// Go 1.22+ 循环变量按迭代独立，闭包可直接捕获 i，无需显式传参
		// Go 1.22+ loop variables are per-iteration, closures can capture i directly without explicit parameters
		go func() {
			defer wg.Done()
			sig.RegisterCancelHandles(func() {
				_ = i // capture to ensure each handler is unique
			})
		}()
	}
	wg.Wait()

	sig.Close(nil)

	ctx := sig.GetStopContext()
	assert.NotNil(t, ctx)
	assert.Error(t, ctx.Err(), "context should be canceled after Close")
}

// TestTerminateSignal_ConcurrentRegisterAndClose_Race 验证并发注册与关闭操作同时进行时不存在数据竞争
// TestTerminateSignal_ConcurrentRegisterAndClose_Race verifies no data race when concurrent registration and close happen simultaneously
func TestTerminateSignal_ConcurrentRegisterAndClose_Race(t *testing.T) {
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")

	var wg sync.WaitGroup
	const n = 100

	wg.Add(n + 1)
	for i := 0; i < n; i++ {
		// Go 1.22+ 循环变量按迭代独立，闭包可直接捕获 i，无需显式传参
		// Go 1.22+ loop variables are per-iteration, closures can capture i directly without explicit parameters
		go func() {
			defer wg.Done()
			sig.RegisterCancelHandles(func() {
				_ = i
			})
		}()
	}
	go func() {
		defer wg.Done()
		sig.Close(nil)
	}()
	wg.Wait()

	ctx := sig.GetStopContext()
	assert.NotNil(t, ctx)
	assert.Error(t, ctx.Err(), "context should be canceled after Close")
}

// TestTerminateSignal_CancelBeforeHandlers 验证 cancel() 在 handler 执行之前被调用，且 handler 确实被执行
// TestTerminateSignal_CancelBeforeHandlers verifies that cancel() is called before handler execution and that the handler actually runs
func TestTerminateSignal_CancelBeforeHandlers(t *testing.T) {
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")

	// 用原子布尔记录 handler 是否真正执行过
	// Use an atomic bool to record whether the handler actually executed
	var handlerExecuted atomic.Bool

	sig.RegisterCancelHandles(func() {
		// 在 handler 执行时，context 应该已经被取消
		// When the handler executes, the context should already be canceled
		assert.Error(t, sig.GetStopContext().Err(), "context should be canceled before handlers execute")
		handlerExecuted.Store(true)
	})

	sig.Close(nil)

	// cancel 已被调用（context 已取消）
	// cancel has been called (the context is canceled)
	assert.Error(t, sig.GetStopContext().Err(), "context should be canceled")

	// handler 确实被执行过
	// The handler was actually executed
	assert.True(t, handlerExecuted.Load(), "handler should have been executed")
}

// TestTerminateSignal_RegisterAfterClose 验证 Close 后 RegisterCancelHandles 不会注册新的 handler
// TestTerminateSignal_RegisterAfterClose verifies that RegisterCancelHandles doesn't add new handlers after Close
func TestTerminateSignal_RegisterAfterClose(t *testing.T) {
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")

	sig.Close(nil)

	sig.RegisterCancelHandles(func() {})

	// Close 后注册必须被拒绝，handles 应保持为空（包内可直接访问）
	// Registration after Close must be rejected and handles should remain empty (directly accessible within the package)
	assert.Equal(t, 0, len(sig.handles), "handlers registered after Close should not be stored")
}

// waitGroupDone 在超时前等待外部等待组计数归零，超时则判定失败（用于把"永挂"转为可诊断的失败）
// waitGroupDone waits for the external wait group to reach zero within a timeout, failing otherwise (turns a "hang forever" into a diagnosable failure)
func waitGroupDone(t *testing.T, wg *sync.WaitGroup) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("wait group was not Done within timeout")
	}
}

// TestTerminateSignal_DoubleClose_MultipleWaitGroups 回归验证：多次（含并发）Close 时，每个调用者的外部等待组都会被 Done 恰好一次
// TestTerminateSignal_DoubleClose_MultipleWaitGroups is a regression test: on multiple (including concurrent) Close calls, the external wait group of every caller must be Done exactly once
func TestTerminateSignal_DoubleClose_MultipleWaitGroups(t *testing.T) {
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")
	sig.RegisterCancelHandles(func() {})

	// 第一次 Close 完成后再 Close，第二个等待组也必须被 Done（旧实现会在此处永挂）
	// After the first Close completes, a second Close must also Done the second wait group (the old implementation hung forever here)
	var wg1, wg2 sync.WaitGroup
	wg1.Add(1)
	sig.Close(&wg1)
	waitGroupDone(t, &wg1)

	wg2.Add(1)
	sig.Close(&wg2)
	waitGroupDone(t, &wg2)

	// 并发多个 goroutine 各自携带独立等待组调用 Close，全部都要被 Done
	// Multiple goroutines call Close concurrently with their own wait groups, all of them must be Done
	const n = 10
	wgs := make([]sync.WaitGroup, n)
	var wgAll sync.WaitGroup
	wgAll.Add(n)
	for i := 0; i < n; i++ {
		wgs[i].Add(1)
		go func(wg *sync.WaitGroup) {
			defer wgAll.Done()
			sig.Close(wg)
		}(&wgs[i])
	}
	wgAll.Wait()

	// Close 返回前必然已调用 wg.Done，因此每个等待组都会立即归零
	// wg.Done must have been called before Close returns, so every wait group reaches zero immediately
	for i := range wgs {
		waitGroupDone(t, &wgs[i])
	}
}

// TestTerminateSignal_CloseAfterPreclose 验证先 Close(nil) 预关闭后，携带等待组的 Close 仍能立即正常返回（模拟 WaitFor 遇到预关闭实例）
// TestTerminateSignal_CloseAfterPreclose verifies that after a Close(nil) pre-close, a Close carrying a wait group still returns immediately and correctly (simulating WaitFor meeting a pre-closed instance)
func TestTerminateSignal_CloseAfterPreclose(t *testing.T) {
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")
	sig.RegisterCancelHandles(func() {})

	// 模拟服务自行预关闭（例如程序里手动 Close 后又收到 SIGTERM）
	// Simulate the service pre-closing itself (e.g. a manual Close in code followed by a SIGTERM)
	sig.Close(nil)

	// Close 必须立即返回且 Done 等待组
	// Close must return immediately and Done the wait group
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	go func() {
		sig.Close(&wg)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Close should return immediately after pre-close")
	}
	waitGroupDone(t, &wg)
}

// TestTerminateSignal_Close_NoHandlers 验证零 handler 时两种关闭模式均正常完成
// TestTerminateSignal_Close_NoHandlers verifies that both close modes complete normally with zero handlers
func TestTerminateSignal_Close_NoHandlers(t *testing.T) {
	// 异步关闭模式：wg.Add(0) 与 close(done) 路径
	// Asynchronous close mode: the wg.Add(0) and close(done) paths
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")
	var wg sync.WaitGroup
	wg.Add(1)
	sig.Close(&wg)
	waitGroupDone(t, &wg)
	assert.Error(t, sig.GetStopContext().Err(), "context should be canceled after Close")

	// 同步关闭模式
	// Synchronous close mode
	sig2 := NewTerminateSignal()
	assert.NotNil(t, sig2, "signal is nil")
	var wg2 sync.WaitGroup
	wg2.Add(1)
	sig2.SyncClose(&wg2)
	waitGroupDone(t, &wg2)
	assert.Error(t, sig2.GetStopContext().Err(), "context should be canceled after SyncClose")
}
