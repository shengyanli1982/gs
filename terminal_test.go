package gs

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestTerminateSignal_SyncClose_Order 验证 SyncClose 按注册顺序执行 handler（README 承诺的注册顺序语义）
// TestTerminateSignal_SyncClose_Order verifies that SyncClose executes handlers in registration order (the registration-order semantics promised by the README)
func TestTerminateSignal_SyncClose_Order(t *testing.T) {
	sig := NewTerminateSignal()
	require.NotNil(t, sig)

	// SyncClose 在当前 goroutine 中顺序执行，append 共享切片不存在并发访问，无需加锁
	// SyncClose executes sequentially in the current goroutine, so appending to the shared slice has no concurrent access and needs no lock
	var order []int
	for i := 0; i < 5; i++ {
		// Go 1.22+ 循环变量按迭代独立，闭包可直接捕获 i
		// Go 1.22+ loop variables are per-iteration, closures can capture i directly
		sig.RegisterCancelHandles(func() {
			order = append(order, i)
		})
	}

	sig.SyncClose(nil)

	// handler 必须严格按注册顺序 0,1,2,3,4 执行
	// Handlers must run strictly in registration order 0,1,2,3,4
	assert.Equal(t, []int{0, 1, 2, 3, 4}, order)
}

// TestTerminateSignal_NilHandler_Skipped 验证两种关闭模式下 nil handler 均被跳过：合法 handler 正常执行、不 panic、外部等待组被 Done
// TestTerminateSignal_NilHandler_Skipped verifies that nil handlers are skipped in both close modes: valid handlers still run, no panic occurs, and the external wait group is Done
func TestTerminateSignal_NilHandler_Skipped(t *testing.T) {
	// 两种模式各验一次 nil 跳过路径：ASyncClose（Close）与 SyncClose
	// Exercise the nil-skip path once per mode: ASyncClose (Close) and SyncClose
	modes := []struct {
		name  string
		close func(sig *TerminateSignal, wg *sync.WaitGroup)
	}{
		{"Close", (*TerminateSignal).Close},
		{"SyncClose", (*TerminateSignal).SyncClose},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			sig := NewTerminateSignal()
			require.NotNil(t, sig)

			// close 返回前所有 handler 均已执行完毕（ASync 模式在 close 内部 wg.Wait），原子计数保证读取无竞态
			// All handlers have completed before close returns (the ASync mode waits via wg.Wait inside close); the atomic counter guarantees a race-free read
			var executed atomic.Int32
			valid := func() { executed.Add(1) }

			// 注册 [valid, nil, valid]，nil 位于中间，必须被跳过且不引发 panic
			// Register [valid, nil, valid]; the nil sits in the middle and must be skipped without panicking
			sig.RegisterCancelHandles(valid, nil, valid)

			var wg sync.WaitGroup
			wg.Add(1)
			require.NotPanics(t, func() { mode.close(sig, &wg) })

			// 两个合法 handler 都必须执行，nil 被静默跳过
			// Both valid handlers must run; the nil is silently skipped
			assert.Equal(t, int32(2), executed.Load(), "both valid handlers should have run")

			// 外部等待组必须被 Done 恰好一次（沿用文件内既有超时模式，避免永挂）
			// The external wait group must be Done exactly once (reusing the existing timeout pattern in this file to avoid hanging forever)
			waitGroupDone(t, &wg)
			assert.Error(t, sig.GetStopContext().Err(), "context should be canceled after close")
		})
	}
}

// TestTerminateSignal_LoserBlocksUntilWinnerDone 验证输家 Close 会阻塞到赢家完成整个关闭流程后才返回
// TestTerminateSignal_LoserBlocksUntilWinnerDone verifies that a loser Close blocks until the winner has finished the entire shutdown sequence
func TestTerminateSignal_LoserBlocksUntilWinnerDone(t *testing.T) {
	sig := NewTerminateSignal()
	require.NotNil(t, sig)

	// started 通告赢家已进入 handler；release 控制 handler 何时结束（用 channel 编排时序，避免裸阻塞）
	// started announces that the winner has entered the handler; release controls when the handler finishes (channels orchestrate the timing, avoiding raw blocking)
	started := make(chan struct{})
	release := make(chan struct{})
	sig.RegisterCancelHandles(func() {
		close(started)
		<-release
	})

	// goroutine A 先调用 Close 成为赢家，并卡在慢 handler 中
	// Goroutine A calls Close first, becomes the winner, and gets stuck in the slow handler
	winnerDone := make(chan struct{})
	go func() {
		defer close(winnerDone)
		sig.Close(nil)
	}()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("winner handler did not start within timeout")
	}

	// goroutine B 后调用 Close 成为输家，必须阻塞在 <-s.done 上等待赢家完成
	// Goroutine B calls Close afterwards, becomes the loser, and must block on <-s.done waiting for the winner
	loserDone := make(chan struct{})
	go func() {
		defer close(loserDone)
		sig.Close(nil)
	}()

	// 赢家尚未完成时，B 不得提前返回（100ms 观察窗口）
	// While the winner has not finished, B must not return early (100ms observation window)
	select {
	case <-loserDone:
		t.Fatal("loser returned before the winner finished")
	case <-time.After(100 * time.Millisecond):
	}

	// 释放慢 handler：赢家关闭 done 通道后，B 应被唤醒并在超时内返回
	// Release the slow handler: once the winner closes the done channel, B should be woken up and return within the timeout
	close(release)

	select {
	case <-winnerDone:
	case <-time.After(3 * time.Second):
		t.Fatal("winner did not finish within timeout")
	}
	select {
	case <-loserDone:
	case <-time.After(3 * time.Second):
		t.Fatal("loser did not finish within timeout")
	}
}
