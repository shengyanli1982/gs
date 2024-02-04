<div align="center">
	<h1>GS</h1>
	<p>A lightweight generic graceful shutdown component<p>
	<img src="assets/logo.png" alt="logo" width="300px">
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

### Methods

**Create**

-   `NewTerminateSignal` : Create a new `TerminateSignal` instance
-   `NewDefaultTerminateSignal` : Create a new `TerminateSignal` instance with default signals
-   `NewTerminateSignalWithContext` : Create a new `TerminateSignal` instance with context

> [!TIP]
> The `InfinityTerminateTimeout` value is used to set the timeout signal to infinity. It means that the `TerminateSignal` instance will not be closed until `Close` method is called and the resources registered in the `TerminateSignal` instance are closed.

**TerminateSignal**

-   `RegisterCancelCallback` : Register the resources which want to be closed when the service is terminated
-   `GetStopContext` : Get the context of the `TerminateSignal` instance
-   `Close` : Close the `TerminateSignal` instance

**Waiting**

-   `WaitingForGracefulShutdown` : Use this method to wait for all `TerminateSignal` instances to shutdown gracefully

### Example

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shengyanli1982/gs"
)

// simulate a service
type testTerminateSignal struct{}

func (t *testTerminateSignal) Close() {
	fmt.Println("testTerminateSignal.Close()")
}

// simulate a service
type testTerminateSignal2 struct{}

func (t *testTerminateSignal2) Shutdown() {
	fmt.Println("testTerminateSignal2.Shutdown()")
}

// simulate a service
type testTerminateSignal3 struct{}

func (t *testTerminateSignal3) Terminate() {
	fmt.Println("testTerminateSignal3.Terminate()")
}

func main() {
	// Create TerminateSignal instance
	s := gs.NewDefaultTerminateSignal()

	// create resources which want to be closed when the service is terminated
	t1 := &testTerminateSignal{}
	t2 := &testTerminateSignal2{}
	t3 := &testTerminateSignal3{}

	// Register the close method of the resource which want to be closed when the service is terminated
<<<<<<< HEAD
	s.CancelCallbacksRegistry(t1.Close, t2.Shutdown, t3.Terminate)
=======
	s.RegisterCancelCallback(tts.Close)
>>>>>>> main

	// Create a goroutine to send a signal to the process after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			fmt.Println(err.Error())
		}
		err = p.Signal(os.Interrupt)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	// Use WaitingForGracefulShutdown method to wait for the TerminateSignal instance to shutdown gracefully
	gs.WaitingForGracefulShutdown(s)

	fmt.Println("shutdown gracefully")
}
```

**Result**

```bash
# go run main.go
testTerminateSignal3.Terminate()
testTerminateSignal.Close()
testTerminateSignal2.Shutdown()
shutdown gracefully
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
	"fmt"
	"os"
	"time"

	"github.com/shengyanli1982/gs"
)

// simulate a service
type testTerminateSignal struct{}

func (t *testTerminateSignal) Close() {
	time.Sleep(5 * time.Second)
}

func main() {
	// Create TerminateSignal instance
	s := gs.NewTerminateSignal(time.Second)  // timeout signal is set to 1 second

	// create a resource which want to be closed when the service is terminated
	t1 := &testTerminateSignal{}

<<<<<<< HEAD
	// Register the close method of the resource which want to be closed when the service is terminated
	s.CancelCallbacksRegistry(t1.Close)
=======
	s.RegisterCancelCallback(tts.Close)
>>>>>>> main

	// Create a goroutine to send a signal to the process after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			fmt.Println(err.Error())
		}
		err = p.Signal(os.Interrupt)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	// Use WaitingForGracefulShutdown method to wait for the TerminateSignal instance to shutdown gracefully
	gs.WaitingForGracefulShutdown(s)

	fmt.Println("shutdown gracefully")
}
```
