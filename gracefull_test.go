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
		sig.RegisterCancelCallback(tts.Close)
	}

	go func() {
		time.Sleep(time.Second)
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			assert.Fail(t, err.Error())
		}
		err = p.Signal(os.Interrupt)
		if err != nil {
			assert.Fail(t, err.Error())
		}
	}()

	WaitForAsync(sig)
}

func TestWaitForAsync_Wait(t *testing.T) {
	sigs := make([]*TerminateSignal, 0)

	for i := 0; i < 10; i++ {
		sig := NewTerminateSignal()
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelCallback(tts.Close)
		sigs = append(sigs, sig)
	}

	go func() {
		time.Sleep(time.Second)
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			assert.Fail(t, err.Error())
		}
		err = p.Signal(os.Interrupt)
		if err != nil {
			assert.Fail(t, err.Error())
		}
	}()

	WaitForAsync(sigs...)
}

func TestWaitForSync_Signal(t *testing.T) {
	sig := NewTerminateSignal()

	for i := 0; i < 10; i++ {
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelCallback(tts.Close)
	}

	go func() {
		time.Sleep(time.Second)
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			assert.Fail(t, err.Error())
		}
		err = p.Signal(os.Interrupt)
		if err != nil {
			assert.Fail(t, err.Error())
		}
	}()

	WaitForSync(sig)
}
