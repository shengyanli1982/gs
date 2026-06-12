package gs

import (
	"context"
	"fmt"
	"sync"
	"testing"

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
		go func(id int) {
			defer wg.Done()
			sig.RegisterCancelHandles(func() {
				_ = id // capture to ensure each handler is unique
			})
		}(i)
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
		go func(id int) {
			defer wg.Done()
			sig.RegisterCancelHandles(func() {
				_ = id
			})
		}(i)
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

// TestTerminateSignal_CancelBeforeHandlers 验证 cancel() 在 handler 执行之前被调用
// TestTerminateSignal_CancelBeforeHandlers verifies that cancel() is called before handler execution
func TestTerminateSignal_CancelBeforeHandlers(t *testing.T) {
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")

	var cancelTime, handlerTime int64
	var mu sync.Mutex

	sig.RegisterCancelHandles(func() {
		mu.Lock()
		handlerTime = 1
		// 在 handler 执行时，context 应该已经被取消
		assert.Error(t, sig.GetStopContext().Err(), "context should be canceled before handlers execute")
		mu.Unlock()
	})

	sig.Close(nil)

	// cancel 已被调用（context 已取消）
	ctx := sig.GetStopContext()
	assert.Error(t, ctx.Err(), "context should be canceled")

	mu.Lock()
	cancelTime = 1
	mu.Unlock()

	assert.Equal(t, int64(1), cancelTime)
	assert.Equal(t, int64(1), handlerTime)
}

// TestTerminateSignal_RegisterAfterClose 验证 Close 后 RegisterCancelHandles 不会注册新的 handler
// TestTerminateSignal_RegisterAfterClose verifies that RegisterCancelHandles doesn't add new handlers after Close
func TestTerminateSignal_RegisterAfterClose(t *testing.T) {
	sig := NewTerminateSignal()
	assert.NotNil(t, sig, "signal is nil")

	sig.Close(nil)

	called := false
	sig.RegisterCancelHandles(func() {
		called = true
	})

	assert.False(t, called, "handler registered after Close should not be called")
}
