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
