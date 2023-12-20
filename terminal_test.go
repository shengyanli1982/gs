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
	signal := NewTerminateSignal(-1)
	if signal == nil {
		assert.Fail(t, "signal is nil")
	}
	assert.Equal(t, signal.GetStopCtx().Err(), nil)
	tts := &testTerminateSignal{}
	signal.CancelCallbacksRegistry(tts.Close)
	wg := sync.WaitGroup{}
	wg.Add(1)
	signal.Close(&wg)
	wg.Wait()
}

func TestTerminateSignal_WithTimeout(t *testing.T) {
	timeout := time.Second
	signal := NewTerminateSignal(timeout)
	if signal == nil {
		assert.Fail(t, "signal is nil")
	}
	assert.Equal(t, signal.GetStopCtx().Err(), nil)
	tts := &testTerminateSignal{}
	signal.CancelCallbacksRegistry(tts.Close)
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
	assert.Equal(t, signal.GetStopCtx().Err(), nil)
	tts := &testTerminateSignal{}
	signal.CancelCallbacksRegistry(tts.Close)
	wg := sync.WaitGroup{}
	wg.Add(1)
	signal.Close(&wg)
	wg.Wait()
}
