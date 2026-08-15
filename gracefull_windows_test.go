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

// procGetConsoleWindow 是 kernel32.dll 的 GetConsoleWindow 函数（x/sys/windows 未导出该函数）
// procGetConsoleWindow is the GetConsoleWindow function from kernel32.dll (not exported by x/sys/windows)
var procGetConsoleWindow = windows.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow")

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

// hasConsoleWindow 检测当前进程是否关联了控制台窗口（无控制台时返回值为 0）
// hasConsoleWindow detects whether the current process has an associated console window (returns 0 when there is none)
func hasConsoleWindow() bool {
	ret, _, _ := procGetConsoleWindow.Call()
	return ret != 0
}

// skipIfNoConsole 在无控制台窗口时跳过测试。在无控制台环境（如 CI 后台进程）中，
// GenerateConsoleCtrlEvent 会返回成功但事件不会被送达，signal.Notify 注册的处理器永远收不到信号，测试将永久阻塞
// skipIfNoConsole skips the test when there is no console window. In a console-less environment
// (e.g. a CI background process), GenerateConsoleCtrlEvent returns success but the event is never
// delivered, so handlers registered via signal.Notify never receive a signal and the test blocks forever
func skipIfNoConsole(t *testing.T) {
	t.Helper()
	if !hasConsoleWindow() {
		t.Skip("no console window detected, console control events cannot be delivered")
	}
}

func TestWaitForAsync_Signal(t *testing.T) {
	skipIfNoConsole(t)
	sig := NewTerminateSignal()

	for i := 0; i < 10; i++ {
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
	}

	go func() {
		time.Sleep(time.Second)
		err := GenerateConsoleCtrlEvent(syscall.CTRL_C_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
	}()

	WaitForAsync(sig)
}

func TestWaitForAsync_Wait(t *testing.T) {
	skipIfNoConsole(t)
	sigs := make([]*TerminateSignal, 0)

	for i := 0; i < 10; i++ {
		sig := NewTerminateSignal()
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
		sigs = append(sigs, sig)
	}

	go func() {
		// 等待一秒钟，确保 signal.Notify 已经注册，避免事件先于注册被丢弃
		// Wait for one second to ensure signal.Notify has registered, so the event is not dropped before registration
		time.Sleep(time.Second)
		err := GenerateConsoleCtrlEvent(syscall.CTRL_C_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
	}()

	WaitForAsync(sigs...)
}

func TestWaitForSync_Signal(t *testing.T) {
	skipIfNoConsole(t)
	sig := NewTerminateSignal()

	for i := 0; i < 10; i++ {
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
	}

	go func() {
		// 等待一秒钟，确保 signal.Notify 已经注册，避免事件先于注册被丢弃
		// Wait for one second to ensure signal.Notify has registered, so the event is not dropped before registration
		time.Sleep(time.Second)
		err := GenerateConsoleCtrlEvent(syscall.CTRL_C_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
	}()

	WaitForSync(sig)
}

func TestWaitForSync_Wait(t *testing.T) {
	skipIfNoConsole(t)
	sigs := make([]*TerminateSignal, 0)

	for i := 0; i < 10; i++ {
		sig := NewTerminateSignal()
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
		sigs = append(sigs, sig)
	}

	go func() {
		// 等待一秒钟，确保 signal.Notify 已经注册，避免事件先于注册被丢弃
		// Wait for one second to ensure signal.Notify has registered, so the event is not dropped before registration
		time.Sleep(time.Second)
		err := GenerateConsoleCtrlEvent(syscall.CTRL_C_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
	}()

	WaitForSync(sigs...)
}

func TestWaitForForceSync_Signal(t *testing.T) {
	skipIfNoConsole(t)
	sig := NewTerminateSignal()

	for i := 0; i < 10; i++ {
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
	}

	go func() {
		// 等待一秒钟，确保 signal.Notify 已经注册，避免事件先于注册被丢弃
		// Wait for one second to ensure signal.Notify has registered, so the event is not dropped before registration
		time.Sleep(time.Second)
		err := GenerateConsoleCtrlEvent(syscall.CTRL_BREAK_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
	}()

	WaitForForceSync(sig)
}

func TestWaitForForceSync_Wait(t *testing.T) {
	skipIfNoConsole(t)
	sigs := make([]*TerminateSignal, 0)

	for i := 0; i < 10; i++ {
		sig := NewTerminateSignal()
		tts := NewTestTerminateSignal(fmt.Sprintf("test-%d", i))
		sig.RegisterCancelHandles(tts.Close)
		sigs = append(sigs, sig)
	}

	go func() {
		// 等待一秒钟，确保 signal.Notify 已经注册，避免事件先于注册被丢弃
		// Wait for one second to ensure signal.Notify has registered, so the event is not dropped before registration
		time.Sleep(time.Second)
		err := GenerateConsoleCtrlEvent(syscall.CTRL_BREAK_EVENT, 0)
		assert.NoError(t, err, "GenerateConsoleCtrlEvent failed")
	}()

	WaitForForceSync(sigs...)
}
