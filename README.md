<div align="center">
	<h1>G.S</h1>
	<img src="assets/logo.png" alt="logo" width="300px">
    <h4>A lightweight generic graceful shutdown component</h4>
</div>

# Introduction

**Graceful Shutdown** is a common requirement for most services. It is a good practice to shutdown the service gracefully when the service receives a signal to terminate. The graceful shutdown process usually includes the following steps:

1. Create `TerminateSignal` instance, and register the signal which want to be received.
2. Register the resources which want to be closed when the service is terminated.
3. Use `WaitingForGracefulShutdown` method to wait for the `TerminateSignal` instance to shutdown gracefully.

# Advantage

-   Simple and easy to use
-   No third-party dependencies
-   Low memory usage
-   Support timeout signal
-   Support context
-   Support multiple signals

# Installation

```bash
go get github.com/shengyanli1982/gs
```

# Quick Start

`GS` is very simple, less code and easy to use. Just create `TerminateSignal` instances, register the resources which want to be closed when the service is terminated, and use `WaitingForGracefulShutdown` method to wait for the `TerminateSignal` instances to shutdown gracefully.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shengyanli1982/gs"
)

type testTerminateSignal struct{}

func (t *testTerminateSignal) Close() {
	time.Sleep(5 * time.Second)
}

func main() {
	// Create TerminateSignal instance
	s := gs.NewDefaultTerminateSignal()

	// create a resource which want to be closed when the service is terminated
	tts := &testTerminateSignal{}

	// Register the close method of the resource which want to be closed when the service is terminated
	s.CancelCallbacksRegistry(tts.Close)

	// Create a goroutine to send a signal to the process after 2 seconds
	go func() {
		time.Sleep(2*time.Second)
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			assert.Fail(t, err.Error())
		}
		err = p.Signal(os.Interrupt)
		if err != nil {
			assert.Fail(t, err.Error())
		}
	}()

	// Use WaitingForGracefulShutdown method to wait for the TerminateSignal instance to shutdown gracefully
	WaitingForGracefulShutdown(s)

	fmt.Println("shutdown gracefully")
```

# Features

`GS` provides features not many but enough for most services.

## Timeout Signal

`TerminateSignal` instance can be created with a timeout signal. When the timeout signal is received, the `TerminateSignal` instance will be closed not waiting for resources registered in the `TerminateSignal` instance will be closed.

> [!TIP]
> The **Timeout** can fix the problem that the service cannot be closed due to the resource cannot be closed. But it is not recommended to use timeout signal, because it may cause the resource to be closed abnormally.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shengyanli1982/gs"
)

type testTerminateSignal struct{}

func (t *testTerminateSignal) Close() {
	time.Sleep(5 * time.Second)
}

func main() {
	s := gs.NewTerminateSignal(time.Second)  // timeout signal is set to 1 second

	tts := &testTerminateSignal{}

	s.CancelCallbacksRegistry(tts.Close)

	go func() {
		time.Sleep(2*time.Second)
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

	fmt.Println("shutdown gracefully")
}
```
