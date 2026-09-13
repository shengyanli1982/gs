package gs

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

// benchSink 是包级汇（sink），用于承接构造函数基准的结果，防止编译器判定结果未被使用而消除被测分配与内联，导致少报
// benchSink is a package-level sink that receives the results of constructor benchmarks, preventing the compiler from treating the results as unused and eliminating the measured allocations and inlining, which would under-report
var benchSink *TerminateSignal

// 去伪影说明：原实现丢弃构造结果，编译器可消除分配；且 NewTerminateSignalWithContext 可被内联（cost 76≤80），
// 导致 t 假性留栈、少报 1 alloc/96B。赋值给 benchSink 后，本基准反映生产真实成本
// Artifact-removal note: the original implementation discarded the constructed value, letting the compiler eliminate allocations; also NewTerminateSignalWithContext was inlinable (cost 76≤80),
// causing t to falsely stay on the stack and under-reporting 1 alloc/96B. Assigning to benchSink makes this benchmark reflect the true production cost
func BenchmarkNewTerminateSignal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = NewTerminateSignal()
	}
}

// 去伪影说明：同上，赋值给 benchSink 防止内联与分配消除；
// 预期本基准从 3 allocs 变为 4 allocs——这是去除测量伪影，不是性能回归
// Artifact-removal note: same as above, assigning to benchSink prevents inlining and allocation elimination;
// this benchmark is expected to change from 3 allocs to 4 allocs — that is removing a measurement artifact, not a performance regression
func BenchmarkNewTerminateSignalWithContext(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSink = NewTerminateSignalWithContext(ctx)
	}
}

func BenchmarkRegisterCancelHandles_Single(b *testing.B) {
	b.ReportAllocs()
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		sig.RegisterCancelHandles(noop)
	}
}

func BenchmarkRegisterCancelHandles_10(b *testing.B) {
	b.ReportAllocs()
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		for j := 0; j < 10; j++ {
			sig.RegisterCancelHandles(noop)
		}
	}
}

func BenchmarkRegisterCancelHandles_100(b *testing.B) {
	b.ReportAllocs()
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		for j := 0; j < 100; j++ {
			sig.RegisterCancelHandles(noop)
		}
	}
}

func BenchmarkRegisterCancelHandles_Batch10(b *testing.B) {
	b.ReportAllocs()
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
	b.ReportAllocs()
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		sig.RegisterCancelHandles(noop)
		sig.Close(nil)
	}
}

func BenchmarkClose_Async_10Handlers(b *testing.B) {
	b.ReportAllocs()
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
	b.ReportAllocs()
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
	b.ReportAllocs()
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		sig.RegisterCancelHandles(noop)
		sig.SyncClose(nil)
	}
}

func BenchmarkClose_Sync_10Handlers(b *testing.B) {
	b.ReportAllocs()
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
	b.ReportAllocs()
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
	b.ReportAllocs()
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
	b.ReportAllocs()
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

// BenchmarkConcurrentRegister 测量并发注册时的锁竞争开销。
// 所有并行 goroutine 通过 atomic.Pointer 共享同一个 sig，以保证测量对象是注册锁竞争；
// 每个 goroutine 累计注册 1024 次后重建 sig，使状态有界，避免向同一切片无限 append 带来的扩容噪声
// BenchmarkConcurrentRegister measures the lock contention overhead of concurrent registration.
// All parallel goroutines share one sig via atomic.Pointer so the benchmark measures registration
// lock contention; each goroutine rebuilds the sig every 1024 registrations to bound the state and
// avoid the growth noise of appending to the same slice forever
func BenchmarkConcurrentRegister(b *testing.B) {
	b.ReportAllocs()
	noop := func() {}
	var current atomic.Pointer[TerminateSignal]
	current.Store(NewTerminateSignal())
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		count := 0
		for pb.Next() {
			current.Load().RegisterCancelHandles(noop)
			count++
			// 达到阈值后重建 sig（多个 goroutine 可能同时重建，最后写入者生效，语义无害）
			// Rebuild the sig once the threshold is reached (multiple goroutines may rebuild concurrently; the last writer wins, which is harmless)
			if count == 1024 {
				current.Store(NewTerminateSignal())
				count = 0
			}
		}
	})
}

func BenchmarkFullLifecycle_Async10(b *testing.B) {
	b.ReportAllocs()
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
	b.ReportAllocs()
	noop := func() {}
	for i := 0; i < b.N; i++ {
		sig := NewTerminateSignal()
		for j := 0; j < 10; j++ {
			sig.RegisterCancelHandles(noop)
		}
		sig.SyncClose(nil)
	}
}
