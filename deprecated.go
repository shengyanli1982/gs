package gs

// WaitingForGracefulShutdown 函数等待所有的关闭信号
// The WaitingForGracefulShutdown function waits for all shutdown signals
//
// Deprecated: As of GS v0.1.3 this function simply calls WaitForAsync.
func WaitingForGracefulShutdown(sigs ...*TerminateSignal) {
	WaitForAsync(sigs...)
}
