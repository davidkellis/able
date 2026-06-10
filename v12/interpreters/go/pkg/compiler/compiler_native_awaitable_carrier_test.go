package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRuntimeUsesLazyNativeAwaitableCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  print(1)",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_default_awaitable struct",
		"func (a *__able_channel_awaitable) MaterializeRuntimeValue() runtime.Value",
		"func (a *__able_mutex_awaitable) MaterializeRuntimeValue() runtime.Value",
		"func (a *__able_timer_awaitable) MaterializeRuntimeValue() runtime.Value",
		"if native, ok := awaitable.(runtime.NativeAwaitableValue); ok",
		"return awaitable, nil",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected generated native Awaitable carrier fragment %q", fragment)
		}
	}
	for _, legacy := range []string{
		"return awaitable.toStruct(), nil",
	} {
		if strings.Contains(compiledSrc, legacy) {
			t.Fatalf("unexpected eager Awaitable materialization %q", legacy)
		}
	}
}

func TestCompilerExecutionContextUsesLazyAwaitServiceCarriers(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"struct CancelReg {}",
		"methods CancelReg {",
		"  fn cancel(self: Self) -> void { nil }",
		"}",
		"",
		"struct ReadyArm {}",
		"methods ReadyArm {",
		"  fn is_ready(self: Self) -> bool { true }",
		"  fn register(self: Self, _waker) -> CancelReg { CancelReg {} }",
		"  fn commit(self: Self) -> i64 { 7_i64 }",
		"  fn is_default(self: Self) -> bool { false }",
		"}",
		"",
		"fn choose() -> i64 {",
		"  await [ReadyArm {}]",
		"}",
		"",
	}, "\n")
	experimental := compileNoFallbackSourceWithCompilerOptions(t, source, Options{
		ExperimentalExecutionContext: true,
	})
	compiled := string(experimental.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_native_await_waker struct",
		"type __able_native_await_registration struct",
		"materialized atomic.Pointer[runtime.StructInstanceValue]",
		"materialized.CompareAndSwap(nil, inst)",
		"awaitState",
		"s.payload.awaitState == s",
		"func __able_acquire_await_state(",
		"if state.waker != nil {",
		"__able_await_waker_pending",
		"func (s *__able_await_state) prepareArmScratch(capacity int)",
		"func (s *__able_await_state) appendArm(awaitable runtime.Value, isDefault bool)",
		"func (s *__able_await_state) releaseReusable()",
		"func (s *__able_await_state) setWaker(waker runtime.Value)",
		"signalWaker(waker runtime.Value)",
		"arm.awaitable = nil",
		"arm.registration = nil",
		"func (s *__able_await_state) clearWaker()",
		"func (p *__able_async_payload) getAwaitState(_ *ast.AwaitExpression)",
		"func (s *__able_await_state) signal() bool",
		"|| !s.waiting",
		"if native, ok := waker.(*__able_native_await_waker); ok",
		"if native, ok := reg.(*__able_native_await_registration); ok",
	} {
		if !strings.Contains(compiled, fragment) {
			t.Fatalf("expected lazy Await service carrier fragment %q", fragment)
		}
	}
	if got := strings.Count(compiled, "materialized atomic.Pointer[runtime.StructInstanceValue]"); got != 2 {
		t.Fatalf("generated atomic Await service materialization fields = %d, want 2", got)
	}

	defaultResult := compileNoFallbackSource(t, source)
	defaultCompiled := string(defaultResult.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_native_await_waker struct",
		"type __able_native_await_registration struct",
		"s.payload.awaitState == s",
		"func __able_acquire_await_state(",
		"func (s *__able_await_state) prepareArmScratch(capacity int)",
	} {
		if strings.Contains(defaultCompiled, fragment) {
			t.Fatalf("default generated source unexpectedly contains %q", fragment)
		}
	}
	if !strings.Contains(defaultCompiled, "awaitStates") {
		t.Fatalf("default generated source must retain its compatibility Await state cache")
	}
}

func TestCompilerReusableAwaitStateFallsBackForNestedCommitAwait(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiled nested Await execution test in short mode")
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{AwaitRegistration, AwaitWaker}",
		"",
		"struct CallbackArm { callback: fn() -> i64 }",
		"",
		"methods CallbackArm {",
		"  fn is_ready(self: Self) -> bool { true }",
		"  fn register(self: Self, _waker: AwaitWaker) -> AwaitRegistration {",
		"    AwaitRegistration { cancel: fn() -> void { nil } }",
		"  }",
		"  fn commit(self: Self) -> i64 { self.callback() }",
		"  fn is_default(self: Self) -> bool { false }",
		"}",
		"",
		"fn main() -> void {",
		"  joined := spawn {",
		"    await [CallbackArm { callback: fn() -> i64 {",
		"      await [CallbackArm { callback: fn() -> i64 { 42_i64 } }]",
		"    } }]",
		"  }",
		"  future_flush()",
		"  print(joined.value())",
		"}",
		"",
	}, "\n")

	got := compileAndRunExecSourceWithOptions(t, "ablec-reusable-await-state-nested", source, Options{
		PackageName:                  "main",
		EmitMain:                     true,
		ExperimentalExecutionContext: true,
	})
	if want := "42\n"; got != want {
		t.Fatalf("compiled nested Await output = %q, want %q", got, want)
	}
}

func TestCompilerReusableAwaitStateFallsBackForNestedArmCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiled nested Await arm-collection test in short mode")
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{AwaitRegistration, AwaitWaker}",
		"",
		"struct ReadyArm { value: i64 }",
		"methods ReadyArm {",
		"  fn is_ready(self: Self) -> bool { true }",
		"  fn register(self: Self, _waker: AwaitWaker) -> AwaitRegistration {",
		"    AwaitRegistration { cancel: fn() -> void { nil } }",
		"  }",
		"  fn commit(self: Self) -> i64 { self.value }",
		"  fn is_default(self: Self) -> bool { false }",
		"}",
		"",
		"struct CollectArm { inspect: fn() -> i64 }",
		"methods CollectArm {",
		"  fn is_ready(self: Self) -> bool { true }",
		"  fn register(self: Self, _waker: AwaitWaker) -> AwaitRegistration {",
		"    AwaitRegistration { cancel: fn() -> void { nil } }",
		"  }",
		"  fn commit(self: Self) -> i64 { 7_i64 }",
		"  fn is_default(self: Self) -> bool {",
		"    self.inspect()",
		"    false",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  joined := spawn {",
		"    await [CollectArm { inspect: fn() -> i64 {",
		"      await [ReadyArm { value: 42_i64 }]",
		"    } }]",
		"  }",
		"  future_flush()",
		"  print(joined.value())",
		"}",
		"",
	}, "\n")

	got := compileAndRunExecSourceWithOptions(t, "ablec-reusable-await-state-collect-nested", source, Options{
		PackageName:                  "main",
		EmitMain:                     true,
		ExperimentalExecutionContext: true,
	})
	if want := "7\n"; got != want {
		t.Fatalf("compiled nested arm-collection Await output = %q, want %q", got, want)
	}
}

func TestCompilerNativeAwaitableCarrierPreservesExplicitProtocolCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiled native Awaitable protocol execution test in short mode")
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.concurrency.{Await}",
		"import able.kernel.{AwaitWaker, Mutex}",
		"",
		"fn main() -> void {",
		"  default_arm := Await.default(fn() -> i64 { 7_i64 })",
		"  print(default_arm.is_default())",
		"  print(default_arm.is_ready())",
		"  print(default_arm.commit())",
		"",
		"  mutex := Mutex.new()",
		"  mutex_arm := mutex.await_lock(fn() -> i64 {",
		"    do { 9_i64 } ensure { mutex.unlock() }",
		"  })",
		"  print(mutex_arm.is_default())",
		"  print(mutex_arm.is_ready())",
		"  print(mutex_arm.commit())",
		"",
		"  registration := mutex.await_lock(fn() -> i64 { 11_i64 }).register(",
		"    AwaitWaker { wake: fn() -> void { nil } }",
		"  )",
		"  registration.cancel()",
		"}",
		"",
	}, "\n")

	for _, tc := range []struct {
		name         string
		experimental bool
	}{
		{name: "default"},
		{name: "execution_context", experimental: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := compileAndRunExecSourceWithOptions(t, "ablec-native-awaitable-protocol", source, Options{
				PackageName:                  "main",
				EmitMain:                     true,
				ExperimentalExecutionContext: tc.experimental,
			})
			want := "true\ntrue\n7\nfalse\ntrue\n9\n"
			if got != want {
				t.Fatalf("compiled explicit Awaitable protocol output = %q, want %q", got, want)
			}
		})
	}
}

func TestCompilerNativeAwaitWakerMaterializesForUserAwaitable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiled user Awaitable execution test in short mode")
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{AwaitRegistration, AwaitWaker}",
		"",
		"struct ManualArm { ready: bool }",
		"",
		"methods ManualArm {",
		"  fn is_ready(self: Self) -> bool { self.ready }",
		"",
		"  fn register(self: Self, waker: AwaitWaker) -> AwaitRegistration {",
		"    self.ready = true",
		"    waker.wake()",
		"    AwaitRegistration { cancel: fn() -> void { nil } }",
		"  }",
		"",
		"  fn commit(self: Self) -> i64 { 42_i64 }",
		"  fn is_default(self: Self) -> bool { false }",
		"}",
		"",
		"fn main() -> void {",
		"  task := spawn { await [ManualArm { ready: false }] }",
		"  future_flush()",
		"  print(task.value())",
		"}",
		"",
	}, "\n")

	got := compileAndRunExecSourceWithOptions(t, "ablec-native-await-waker-user-awaitable", source, Options{
		PackageName:                  "main",
		EmitMain:                     true,
		ExperimentalExecutionContext: true,
	})
	if want := "42\n"; got != want {
		t.Fatalf("compiled user Awaitable output = %q, want %q", got, want)
	}
}

func TestCompilerNativeAwaitWakerRegistersWithFutureAwaitable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiled Future Awaitable execution test in short mode")
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  child := spawn {",
		"    step := 0_i64",
		"    loop {",
		"      if step >= 64_i64 { break }",
		"      future_yield()",
		"      step = step + 1_i64",
		"    }",
		"    42_i64",
		"  }",
		"  joined := spawn { await [child] }",
		"  future_flush()",
		"  print(joined.value())",
		"}",
		"",
	}, "\n")

	got := compileAndRunExecSourceWithOptions(t, "ablec-native-await-waker-future", source, Options{
		PackageName:                  "main",
		EmitMain:                     true,
		ExperimentalExecutionContext: true,
	})
	if want := "42\n"; got != want {
		t.Fatalf("compiled Future Awaitable output = %q, want %q", got, want)
	}
}

func TestCompilerTaskOwnedAwaitSignalIgnoresStaleWaker(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiled stale Await waker execution test in short mode")
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.concurrency.{Await}",
		"import able.kernel.{AwaitRegistration, AwaitWaker}",
		"",
		"struct ReadyArm { ready: bool }",
		"",
		"methods ReadyArm {",
		"  fn is_ready(self: Self) -> bool { self.ready }",
		"",
		"  fn register(self: Self, waker: AwaitWaker) -> AwaitRegistration {",
		"    self.ready = true",
		"    waker.wake()",
		"    AwaitRegistration { cancel: fn() -> void { waker.wake() } }",
		"  }",
		"",
		"  fn commit(self: Self) -> i64 { 7_i64 }",
		"  fn is_default(self: Self) -> bool { false }",
		"}",
		"",
		"struct ObserverArm { on_register: fn() -> void }",
		"",
		"methods ObserverArm {",
		"  fn is_ready(self: Self) -> bool { false }",
		"  fn register(self: Self, _waker: AwaitWaker) -> AwaitRegistration {",
		"    self.on_register()",
		"    AwaitRegistration { cancel: fn() -> void { nil } }",
		"  }",
		"  fn commit(self: Self) -> i64 { -1_i64 }",
		"  fn is_default(self: Self) -> bool { false }",
		"}",
		"",
		"fn main() -> void {",
		"  joined := spawn {",
		"    await [ReadyArm { ready: false }]",
		"    registrations := 0_i64",
		"    observer := ObserverArm { on_register: fn() -> void { registrations = registrations + 1_i64 } }",
		"    await [",
		"      observer,",
		"      Await.sleep_ms(10_i64, fn() -> i64 { registrations })",
		"    ]",
		"  }",
		"  future_flush()",
		"  print(joined.value())",
		"}",
		"",
	}, "\n")

	t.Setenv("ABLE_EXECUTOR", "goroutine")
	got := compileAndRunExecSourceWithOptions(t, "ablec-task-owned-await-stale-waker", source, Options{
		PackageName:                  "main",
		EmitMain:                     true,
		ExperimentalExecutionContext: true,
	})
	if want := "1\n"; got != want {
		t.Fatalf("compiled stale Await waker output = %q, want %q", got, want)
	}
}
