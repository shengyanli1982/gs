package gs

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitingForGracefulShutdown(t *testing.T) {
	tts := &testTerminateSignal{}
	s := NewDefaultTerminateSignal()
	s.CancelCallbacksRegistry(tts.Close)

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

	WaitingForGracefulShutdown(s)
}
