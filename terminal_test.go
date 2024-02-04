package gs

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testTerminateSignal struct{}

func (t *testTerminateSignal) Close() {
	fmt.Println("testTerminateSignal.Close()")
}

func TestTerminateSignal_Standard(t *testing.T) {
	signal := NewTerminateSignal(InfinityTerminateTimeout)
	if signal == nil {
		assert.Fail(t, "signal is nil")
	}
	assert.Equal(t, signal.GetStopContext().Err(), nil)
	tts := &testTerminateSignal{}
	signal.RegisterCancelCallback(tts.Close)
	wg := sync.WaitGroup{}
	wg.Add(1)
	signal.Close(&wg)
	wg.Wait()
}

func TestTerminateSignal_WithTimeout(t *testing.T) {
	timeout := time.Millisecond * 500
	signal := NewTerminateSignal(timeout)
	if signal == nil {
		assert.Fail(t, "signal is nil")
	}
	assert.Equal(t, signal.GetStopContext().Err(), nil)
	tts := &testTerminateSignal{}
	signal.RegisterCancelCallback(tts.Close)
	time.Sleep(time.Second)
	wg := sync.WaitGroup{}
	wg.Add(1)
	signal.Close(&wg)
	wg.Wait()
}

func TestTerminateSignal_WithContext(t *testing.T) {
	timeout := time.Second
	ctx := context.Background()
	signal := NewTerminateSignalWithContext(ctx, timeout)
	if signal == nil {
		assert.Fail(t, "signal is nil")
	}
	assert.Equal(t, signal.GetStopContext().Err(), nil)
	tts := &testTerminateSignal{}
	signal.RegisterCancelCallback(tts.Close)
	wg := sync.WaitGroup{}
	wg.Add(1)
	signal.Close(&wg)
	wg.Wait()
}
