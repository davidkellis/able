package compiler

func (g *generator) nativeFutureAwaitWakerRegister() string {
	if g == nil || !g.callableExecutionContextsEnabled() {
		return ""
	}
	return `	if nativeWaker, ok := waker.(*__able_native_await_waker); ok && nativeWaker != nil {
		var cancelled atomic.Bool
		handle.AddAwaiter(func() {
			if !cancelled.Load() {
				nativeWaker.wake()
			}
		})
		return __able_make_await_registration_value_ctx(
			func() { cancelled.Store(true) },
			nativeWaker.context,
		), nil
	}
`
}
