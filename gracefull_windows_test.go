//go:build windows

package gs

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/windows"
)

var procGenerateConsoleCtrlEvent = windows.NewLazyDLL("kernel32.dll").NewProc("GenerateConsoleCtrlEvent")

func GenerateConsoleCtrlEvent(ctrlEvent uint32, processGroupID uint32) error {
	ret, _, err := procGenerateConsoleCtrlEvent.Call(
		uintptr(ctrlEvent),
		uintptr(processGroupID),
	)
	if ret == 0 {
		return err
	}
	return nil
}

func TestWaitForAsync_Signal(t *testing.T) {
	sig := NewTerminateSignal()

	for i := 0; i < 10; i++ {
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelCallback(tts.Close)
	}

	go func() {
		time.Sleep(time.Second)
		err := GenerateConsoleCtrlEvent(syscall.CTRL_C_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
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
		err := GenerateConsoleCtrlEvent(syscall.CTRL_C_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
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
		err := GenerateConsoleCtrlEvent(syscall.CTRL_C_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
	}()

	WaitForSync(sig)
}

func TestWaitForSync_Wait(t *testing.T) {
	sigs := make([]*TerminateSignal, 0)

	for i := 0; i < 10; i++ {
		sig := NewTerminateSignal()
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelCallback(tts.Close)
		sigs = append(sigs, sig)
	}

	go func() {
		err := GenerateConsoleCtrlEvent(syscall.CTRL_C_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
	}()

	WaitForSync(sigs...)
}

func TestWaitForForceSync_Signal(t *testing.T) {
	sig := NewTerminateSignal()

	for i := 0; i < 10; i++ {
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelCallback(tts.Close)
	}

	go func() {
		err := GenerateConsoleCtrlEvent(syscall.CTRL_BREAK_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
	}()

	WaitForForceSync(sig)
}

func TestWaitForForceSync_Wait(t *testing.T) {
	sigs := make([]*TerminateSignal, 0)

	for i := 0; i < 10; i++ {
		sig := NewTerminateSignal()
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelCallback(tts.Close)
		sigs = append(sigs, sig)
	}

	go func() {
		err := GenerateConsoleCtrlEvent(syscall.CTRL_BREAK_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
	}()

	WaitForForceSync(sigs...)
}
