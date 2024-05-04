//go:build !windows

package gs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitForAsync_Signal(t *testing.T) {
	sig := NewTerminateSignal()

	for i := 0; i < 10; i++ {
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
	}

	go func() {
		time.Sleep(time.Second)
		p, err := os.FindProcess(os.Getpid())
		assert.NoError(t, err, "os.FindProcess failed")
		err = p.Signal(os.Interrupt)
		assert.NoError(t, err, "os.Signal failed")
	}()

	WaitForAsync(sig)
}

func TestWaitForAsync_Wait(t *testing.T) {
	sigs := make([]*TerminateSignal, 0)

	for i := 0; i < 10; i++ {
		sig := NewTerminateSignal()
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
		sigs = append(sigs, sig)
	}

	go func() {
		time.Sleep(time.Second)
		p, err := os.FindProcess(os.Getpid())
		assert.NoError(t, err, "os.FindProcess failed")
		err = p.Signal(os.Interrupt)
		assert.NoError(t, err, "os.Signal failed")
	}()

	WaitForAsync(sigs...)
}

func TestWaitForSync_Signal(t *testing.T) {
	sig := NewTerminateSignal()

	for i := 0; i < 10; i++ {
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
	}

	go func() {
		time.Sleep(time.Second)
		p, err := os.FindProcess(os.Getpid())
		assert.NoError(t, err, "os.FindProcess failed")
		err = p.Signal(os.Interrupt)
		assert.NoError(t, err, "os.Signal failed")
	}()

	WaitForSync(sig)
}

func TestWaitForSync_Wait(t *testing.T) {
	sigs := make([]*TerminateSignal, 0)

	for i := 0; i < 10; i++ {
		sig := NewTerminateSignal()
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
		sigs = append(sigs, sig)
	}

	go func() {
		time.Sleep(time.Second)
		p, err := os.FindProcess(os.Getpid())
		assert.NoError(t, err, "os.FindProcess failed")
		err = p.Signal(os.Interrupt)
		assert.NoError(t, err, "os.Signal failed")
	}()

	WaitForSync(sigs...)
}

func TestWaitForForceSync_Signal(t *testing.T) {
	sig := NewTerminateSignal()

	for i := 0; i < 10; i++ {
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
	}

	go func() {
		time.Sleep(time.Second)
		p, err := os.FindProcess(os.Getpid())
		assert.NoError(t, err, "os.FindProcess failed")
		err = p.Signal(os.Interrupt)
		assert.NoError(t, err, "os.Signal failed")
	}()

	WaitForForceSync(sig)
}

func TestWaitForForceSync_Wait(t *testing.T) {
	sigs := make([]*TerminateSignal, 0)

	for i := 0; i < 10; i++ {
		sig := NewTerminateSignal()
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
		sigs = append(sigs, sig)
	}

	go func() {
		time.Sleep(time.Second)
		p, err := os.FindProcess(os.Getpid())
		assert.NoError(t, err, "os.FindProcess failed")
		err = p.Signal(os.Interrupt)
		assert.NoError(t, err, "os.Signal failed")
	}()

	WaitForForceSync(sigs...)
}
