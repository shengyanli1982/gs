<div align="center">
	<img src="assets/logo.png" alt="logo" width="450px">
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/shengyanli1982/gs)](https://goreportcard.com/report/github.com/shengyanli1982/gs)
[![Build Status](https://github.com/shengyanli1982/gs/actions/workflows/test.yaml/badge.svg)](https://github.com/shengyanli1982/gs/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/shengyanli1982/gs.svg)](https://pkg.go.dev/github.com/shengyanli1982/gs)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shengyanli1982/gs)](https://golang.org/doc/go1.19)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/shengyanli1982/gs)

# GS - Graceful Shutdown

**GS** is a lightweight Go library for implementing graceful shutdown in your services. It provides a clean API to handle OS termination signals and coordinate the shutdown of multiple resources.

> [!IMPORTANT]
>
> This library is recommended for production use. Previous versions (< v0.2.0) had significant logic and concurrency issues and are no longer maintained.

## Features

- **Simple API** - Minimal setup required
- **Zero dependencies** - No external packages needed
- **Concurrent-safe** - Thread-safe handler registration and execution
- **Context support** - Integrate with Go's context propagation
- **Multiple signals** - Handle SIGINT, SIGTERM, and SIGQUIT
- **Flexible shutdown modes** - Async, sync, and force-sync execution

## Installation

```bash
go get github.com/shengyanli1982/gs
```

## Quick Start

```go
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/shengyanli1982/gs"
)

func main() {
	sig := gs.NewTerminateSignal()

	// Register shutdown handlers
	sig.RegisterCancelHandles(
		func() { fmt.Println("Closing database connection...") },
		func() { fmt.Println("Stopping HTTP server...") },
	)

	// Start HTTP server
	server := &http.Server{Addr: ":8080"}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	fmt.Println("Server started. Press Ctrl+C to shutdown.")

	// Wait for shutdown signal (SIGINT, SIGTERM, SIGQUIT)
	gs.WaitForAsync(sig)

	// Graceful HTTP shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	fmt.Println("Server shutdown gracefully.")
	os.Exit(0)
}
```

## API Reference

### Creating Termination Signals

```go
// Create a new termination signal
sig := gs.NewTerminateSignal()

// Create with context support
sig := gs.NewTerminateSignalWithContext(ctx)
```

### Registering Shutdown Handlers

```go
// Register one or more cleanup functions
sig.RegisterCancelHandles(
	cleanupFunc1,
	cleanupFunc2,
	cleanupFunc3,
)
```

**Note:** Handlers are executed in the order they were registered. The registration is concurrent-safe.

### Shutdown Modes

GS provides three shutdown modes with different execution semantics:

#### 1. Async Mode (`WaitForAsync`)

- **Between signals:** Parallel execution
- **Within each signal:** Concurrent execution (each handler in a goroutine)
- **Use case:** Maximum throughput, order doesn't matter

```go
sig1 := gs.NewTerminateSignal()
sig2 := gs.NewTerminateSignal()

sig1.RegisterCancelHandles(handler1a, handler1b)
sig2.RegisterCancelHandles(handler2a, handler2b)

// All handlers execute concurrently
gs.WaitForAsync(sig1, sig2)
```

#### 2. Sync Mode (`WaitForSync`)

- **Between signals:** Sequential execution
- **Within each signal:** Concurrent execution (each handler in a goroutine)
- **Use case:** Signals need ordered shutdown, handlers can run concurrently

```go
sig1 := gs.NewTerminateSignal()
sig2 := gs.NewTerminateSignal()

// sig1 shuts down completely, then sig2
gs.WaitForSync(sig1, sig2)
```

#### 3. Force Sync Mode (`WaitForForceSync`)

- **Between signals:** Sequential execution
- **Within each signal:** Sequential execution (handlers run one by one)
- **Use case:** Strict ordering required at all levels

```go
sig1 := gs.NewTerminateSignal()
sig2 := gs.NewTerminateSignal()

// sig1 handlers run sequentially, then sig2 handlers
gs.WaitForForceSync(sig1, sig2)
```

### Manual Control

You can also trigger shutdown programmatically:

```go
// Async close
sig.Close(nil)

// Sync close
sig.SyncClose(nil)

// With wait group for coordination
var wg sync.WaitGroup
wg.Add(1)
sig.Close(&wg)
wg.Wait()
```

### Context Integration

Access the shutdown context to coordinate with other goroutines:

```go
sig := gs.NewTerminateSignal()
ctx := sig.GetStopContext()

// Watch for shutdown in your workers
go func() {
	select {
	case <-ctx.Done():
		fmt.Println("Shutdown signal received")
	case <-time.After(10 * time.Second):
		fmt.Println("Timeout")
	}
}()

gs.WaitForAsync(sig)
```

## Advanced Examples

### HTTP Server with Database Connection

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/shengyanli1982/gs"
)

type Database struct {
	connected bool
}

func (db *Database) Close() {
	fmt.Println("Closing database connection...")
	time.Sleep(500 * time.Millisecond)
	db.connected = false
	fmt.Println("Database closed.")
}

func main() {
	sig := gs.NewTerminateSignal()
	db := &Database{connected: true}

	// Register handlers in dependency order
	sig.RegisterCancelHandles(
		func() {
			fmt.Println("Stopping HTTP server...")
			time.Sleep(100 * time.Millisecond)
			fmt.Println("HTTP server stopped.")
		},
		db.Close,
	)

	server := &http.Server{Addr: ":8080"}
	go server.ListenAndServe()

	fmt.Println("Server running on :8080")
	gs.WaitForSync(sig)

	// Graceful HTTP shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	fmt.Println("All services shutdown gracefully.")
}
```

### Multiple Service Shutdown with Coordination

```go
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/shengyanli1982/gs"
)

type Service struct {
	name string
}

func (s *Service) Stop() {
	fmt.Printf("Stopping %s...\n", s.name)
	time.Sleep(300 * time.Millisecond)
	fmt.Printf("%s stopped.\n", s.name)
}

func main() {
	// Create signals for different service groups
	webSig := gs.NewTerminateSignal()
	dbSig := gs.NewTerminateSignal()

	web := &Service{name: "WebService"}
	db := &Service{name: "Database"}

	webSig.RegisterCancelHandles(web.Stop)
	dbSig.RegisterCancelHandles(db.Stop)

	// Synchronous shutdown: web services first, then database
	fmt.Println("Waiting for shutdown signal...")
	gs.WaitForSync(webSig, dbSig)

	fmt.Println("All services shutdown completed.")
}
```

## Behavior Details

### Shutdown Flow

1. **Signal received:** Library catches OS signal (SIGINT, SIGTERM, SIGQUIT)
2. **Context canceled:** All goroutines holding the signal's context are notified
3. **Handlers executed:** Registered handlers run according to the shutdown mode
4. **Wait for completion:** `WaitFor*` functions block until all handlers complete

### Concurrent Safety

- **Handler registration** is thread-safe and can be called from multiple goroutines
- **Shutdown execution** is protected by `sync.Once` to prevent duplicate shutdowns
- **Handler list** is protected by mutex during registration and execution

### Signal Handling

The library listens for these OS signals:

- `SIGINT` (Ctrl+C)
- `SIGTERM` (kill command)
- `SIGQUIT` (Ctrl+\ on Unix)

## Migration from v0.1.x

> [!WARNING]
>
> The `WaitingForGracefulShutdown` method is deprecated since v0.2.0. Use `WaitForAsync`, `WaitForSync`, or `WaitForForceSync` instead.

```go
// Old (deprecated)
gs.WaitingForGracefulShutdown(sig)

// New
gs.WaitForAsync(sig)  // or WaitForSync, WaitForForceSync
```

## Platform Notes

- **Windows:** Use with console applications. The library uses Windows console control events.
- **Unix/Linux/macOS:** Full support for SIGINT, SIGTERM, and SIGQUIT.

## Benchmarks

```
BenchmarkNewTerminateSignal-10         	26520704	87.51 ns/op	192 B/op	3 allocs/op
BenchmarkRegisterCancelHandles_10-10   	 6428458	363.0 ns/op	440 B/op	8 allocs/op
BenchmarkClose_Async_10Handlers-10     	 679658	4432 ns/op	760 B/op	19 allocs/op
BenchmarkClose_Sync_10Handlers-10      	2443123	976.1 ns/op	520 B/op	9 allocs/op
```

## License

This library is distributed under the MIT license. See [LICENSE](LICENSE) for details.
