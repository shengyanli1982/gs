package gs

import (
	"context"
	"sync"
	"testing"
)

func BenchmarkNewTerminateSignal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewTerminateSignal()
	}
}

func BenchmarkNewTerminateSignalWithContext(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTerminateSignalWithContext(ctx)
	}
}

func BenchmarkRegisterCancelHandles_Single(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		sig.RegisterCancelHandles(noop)
	}
}

func BenchmarkRegisterCancelHandles_10(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		for j := 0; j < 10; j++ {
			sig.RegisterCancelHandles(noop)
		}
	}
}

func BenchmarkRegisterCancelHandles_100(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		for j := 0; j < 100; j++ {
			sig.RegisterCancelHandles(noop)
		}
	}
}

func BenchmarkRegisterCancelHandles_Batch10(b *testing.B) {
	handles := make([]func(), 10)
	for i := range handles {
		handles[i] = func() {}
	}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		sig.RegisterCancelHandles(handles...)
	}
}

func BenchmarkClose_Async_1Handler(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		sig.RegisterCancelHandles(noop)
		sig.Close(nil)
	}
}

func BenchmarkClose_Async_10Handlers(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		for j := 0; j < 10; j++ {
			sig.RegisterCancelHandles(noop)
		}
		sig.Close(nil)
	}
}

func BenchmarkClose_Async_100Handlers(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		for j := 0; j < 100; j++ {
			sig.RegisterCancelHandles(noop)
		}
		sig.Close(nil)
	}
}

func BenchmarkClose_Sync_1Handler(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		sig.RegisterCancelHandles(noop)
		sig.SyncClose(nil)
	}
}

func BenchmarkClose_Sync_10Handlers(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		for j := 0; j < 10; j++ {
			sig.RegisterCancelHandles(noop)
		}
		sig.SyncClose(nil)
	}
}

func BenchmarkClose_Sync_100Handlers(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		for j := 0; j < 100; j++ {
			sig.RegisterCancelHandles(noop)
		}
		sig.SyncClose(nil)
	}
}

func BenchmarkClose_Async_ExternalWG(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		sig.RegisterCancelHandles(noop)
		var wg sync.WaitGroup
		wg.Add(1)
		sig.Close(&wg)
		wg.Wait()
	}
}

func BenchmarkClose_Sync_ExternalWG(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		sig.RegisterCancelHandles(noop)
		var wg sync.WaitGroup
		wg.Add(1)
		sig.SyncClose(&wg)
		wg.Wait()
	}
}

func BenchmarkConcurrentRegister(b *testing.B) {
	noop := func() {}
	sig := NewTerminateSignal()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sig.RegisterCancelHandles(noop)
		}
	})
}

func BenchmarkFullLifecycle_Async10(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		for j := 0; j < 10; j++ {
			sig.RegisterCancelHandles(noop)
		}
		sig.Close(nil)
	}
}

func BenchmarkFullLifecycle_Sync10(b *testing.B) {
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		for j := 0; j < 10; j++ {
			sig.RegisterCancelHandles(noop)
		}
		sig.SyncClose(nil)
	}
}
