package bridge

import (
	"math/big"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

func TestAsFloatAcceptsInteger(t *testing.T) {
	val := runtime.IntegerValue{Val: big.NewInt(42), TypeSuffix: runtime.IntegerI64}
	out, err := AsFloat(val)
	if err != nil {
		t.Fatalf("AsFloat error: %v", err)
	}
	if out != 42 {
		t.Fatalf("AsFloat = %v, want 42", out)
	}
}

func TestAsStringAcceptsStringStruct(t *testing.T) {
	byteArr := &runtime.ArrayValue{
		Elements: []runtime.Value{
			runtime.IntegerValue{Val: big.NewInt(72), TypeSuffix: runtime.IntegerU8},
			runtime.IntegerValue{Val: big.NewInt(105), TypeSuffix: runtime.IntegerU8},
		},
	}
	definition := &runtime.StructDefinitionValue{
		Node: ast.StructDef("String", nil, ast.StructKindNamed, nil, nil, false),
	}
	inst := &runtime.StructInstanceValue{
		Definition: definition,
		Fields: map[string]runtime.Value{
			"bytes":     byteArr,
			"len_bytes": runtime.IntegerValue{Val: big.NewInt(2), TypeSuffix: runtime.IntegerI32},
		},
	}
	out, err := AsString(inst)
	if err != nil {
		t.Fatalf("AsString error: %v", err)
	}
	if out != "Hi" {
		t.Fatalf("AsString = %q, want %q", out, "Hi")
	}
}

func TestAsStringAcceptsInterfaceWrappedArrayBytes(t *testing.T) {
	byteArr := &runtime.ArrayValue{
		Elements: []runtime.Value{
			runtime.IntegerValue{Val: big.NewInt(72), TypeSuffix: runtime.IntegerU8},
			runtime.IntegerValue{Val: big.NewInt(105), TypeSuffix: runtime.IntegerU8},
		},
	}
	definition := &runtime.StructDefinitionValue{
		Node: ast.StructDef("String", nil, ast.StructKindNamed, nil, nil, false),
	}
	inst := &runtime.StructInstanceValue{
		Definition: definition,
		Fields: map[string]runtime.Value{
			"bytes": runtime.InterfaceValue{Underlying: byteArr},
		},
	}
	out, err := AsString(inst)
	if err != nil {
		t.Fatalf("AsString error: %v", err)
	}
	if out != "Hi" {
		t.Fatalf("AsString = %q, want %q", out, "Hi")
	}
}

func TestRuntimeEnvFallsBackAfterSwapNil(t *testing.T) {
	rt := New(interpreter.New())
	base := runtime.NewEnvironment(nil)
	rt.SetEnv(base)
	prev := rt.SwapEnv(nil)
	if prev != base {
		t.Fatalf("SwapEnv(nil) previous = %p, want %p", prev, base)
	}
	if got := rt.Env(); got != base {
		t.Fatalf("Env() = %p, want fallback %p", got, base)
	}
}

func TestSwapEnvIfNeededSkipsSameEnvironment(t *testing.T) {
	rt := New(interpreter.New())
	env := runtime.NewEnvironment(nil)
	rt.SetEnv(env)

	prev, swapped := SwapEnvIfNeeded(rt, env)
	if swapped {
		t.Fatalf("SwapEnvIfNeeded swapped identical environment")
	}
	if prev != nil {
		t.Fatalf("SwapEnvIfNeeded prev = %p, want nil", prev)
	}
	if got := rt.Env(); got != env {
		t.Fatalf("Env() = %p, want %p", got, env)
	}
}

func TestSwapEnvIfNeededSwapsDifferentEnvironment(t *testing.T) {
	rt := New(interpreter.New())
	base := runtime.NewEnvironment(nil)
	next := runtime.NewEnvironment(nil)
	rt.SetEnv(base)

	prev, swapped := SwapEnvIfNeeded(rt, next)
	if !swapped {
		t.Fatalf("SwapEnvIfNeeded did not swap distinct environment")
	}
	if prev != base {
		t.Fatalf("SwapEnvIfNeeded prev = %p, want %p", prev, base)
	}
	if got := rt.Env(); got != next {
		t.Fatalf("Env() after swap = %p, want %p", got, next)
	}

	RestoreEnvIfNeeded(rt, prev, swapped)
	if got := rt.Env(); got != base {
		t.Fatalf("Env() after restore = %p, want %p", got, base)
	}
}

func TestExecutorKind(t *testing.T) {
	if got := ExecutorKind(nil); got != "serial" {
		t.Fatalf("ExecutorKind(nil) = %q, want serial", got)
	}
	serial := New(interpreter.New())
	if got := ExecutorKind(serial); got != "serial" {
		t.Fatalf("ExecutorKind(serial) = %q, want serial", got)
	}
	goroutine := New(interpreter.NewWithExecutor(interpreter.NewGoroutineExecutor(nil)))
	if got := ExecutorKind(goroutine); got != "goroutine" {
		t.Fatalf("ExecutorKind(goroutine) = %q, want goroutine", got)
	}
}

func TestStructDefinitionCacheScopesByEnvironment(t *testing.T) {
	interp := interpreter.New()
	rt := New(interp)

	envA := runtime.NewEnvironment(nil)
	envB := runtime.NewEnvironment(nil)

	defA := &runtime.StructDefinitionValue{Node: ast.StructDef("String", nil, ast.StructKindNamed, nil, nil, false)}
	defB := &runtime.StructDefinitionValue{Node: ast.StructDef("String", nil, ast.StructKindNamed, nil, nil, false)}
	envA.DefineStruct("String", defA)
	envB.DefineStruct("String", defB)

	rt.SetEnv(envA)
	gotA, err := rt.StructDefinition("String")
	if err != nil {
		t.Fatalf("StructDefinition envA error: %v", err)
	}
	if gotA != defA {
		t.Fatalf("StructDefinition envA = %p, want %p", gotA, defA)
	}

	rt.SetEnv(envB)
	gotB, err := rt.StructDefinition("String")
	if err != nil {
		t.Fatalf("StructDefinition envB error: %v", err)
	}
	if gotB != defB {
		t.Fatalf("StructDefinition envB = %p, want %p", gotB, defB)
	}
}

func TestStructDefinitionHydratesFromInterpreterLookupWithoutFallbackCounters(t *testing.T) {
	interp := interpreter.New()
	def := &runtime.StructDefinitionValue{Node: ast.StructDef("Thing", nil, ast.StructKindNamed, nil, nil, false)}
	interp.GlobalEnvironment().DefineStruct("Thing", def)
	interp.GlobalEnvironment().Define("Thing", def)

	rt := New(interp)
	rt.SetEnv(runtime.NewEnvironment(nil))
	rt.SetGlobalLookupFallbackEnabled(false)

	ResetGlobalLookupFallbackCounters()
	got, err := rt.StructDefinition("Thing")
	if err != nil {
		t.Fatalf("StructDefinition error: %v", err)
	}
	if got != def {
		t.Fatalf("StructDefinition = %p, want %p", got, def)
	}
	if calls := GlobalLookupFallbackStats(); calls != 0 {
		t.Fatalf("GlobalLookupFallbackStats = %d, want 0", calls)
	}
	if envCalls, registryCalls := GlobalLookupFallbackBucketStats(); envCalls != 0 || registryCalls != 0 {
		t.Fatalf("GlobalLookupFallbackBucketStats = (%d, %d), want (0, 0)", envCalls, registryCalls)
	}
}

func TestStructDefinitionFallsBackFromQualifiedVisibleAliasToLocalStruct(t *testing.T) {
	interp := interpreter.New()
	def := &runtime.StructDefinitionValue{Node: ast.StructDef("Box", nil, ast.StructKindNamed, nil, nil, false)}
	env := runtime.NewEnvironment(nil)
	env.DefineStruct("Box", def)

	rt := New(interp)
	rt.SetEnv(env)
	rt.SetGlobalLookupFallbackEnabled(false)

	got, err := rt.StructDefinition("demo.Box")
	if err != nil {
		t.Fatalf("StructDefinition qualified visible alias error: %v", err)
	}
	if got != def {
		t.Fatalf("StructDefinition qualified visible alias = %p, want %p", got, def)
	}
}

func TestRuntimeCallFallsBackToGlobalEnvironment(t *testing.T) {
	interp := interpreter.New()
	interp.GlobalEnvironment().Define("greet", runtime.NativeFunctionValue{
		Name:  "greet",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return runtime.StringValue{Val: "hello"}, nil
		},
	})
	rt := New(interp)
	rt.SetEnv(runtime.NewEnvironment(nil))

	value, err := rt.Call("greet", nil)
	if err != nil {
		t.Fatalf("Call error: %v", err)
	}
	if got, ok := value.(runtime.StringValue); !ok || got.Val != "hello" {
		t.Fatalf("Call = %#v, want String(\"hello\")", value)
	}
}

func TestRuntimeCallCanDisableGlobalEnvironmentFallback(t *testing.T) {
	interp := interpreter.New()
	interp.GlobalEnvironment().Define("greet", runtime.NativeFunctionValue{
		Name:  "greet",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return runtime.StringValue{Val: "hello"}, nil
		},
	})
	rt := New(interp)
	rt.SetEnv(runtime.NewEnvironment(nil))
	rt.SetGlobalLookupFallbackEnabled(false)

	if _, err := rt.Call("greet", nil); err == nil {
		t.Fatalf("expected Call to fail when global fallback is disabled")
	}
}

func TestCallNamedFallsBackToGlobalEnvironment(t *testing.T) {
	interp := interpreter.New()
	interp.GlobalEnvironment().Define("greet", runtime.NativeFunctionValue{
		Name:  "greet",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return runtime.StringValue{Val: "hello"}, nil
		},
	})
	rt := New(interp)
	rt.SetEnv(runtime.NewEnvironment(nil))

	value, err := CallNamed(rt, "greet", nil)
	if err != nil {
		t.Fatalf("CallNamed error: %v", err)
	}
	if got, ok := value.(runtime.StringValue); !ok || got.Val != "hello" {
		t.Fatalf("CallNamed = %#v, want String(\"hello\")", value)
	}
}

func TestCallNamedCanDisableGlobalEnvironmentFallback(t *testing.T) {
	interp := interpreter.New()
	interp.GlobalEnvironment().Define("greet", runtime.NativeFunctionValue{
		Name:  "greet",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return runtime.StringValue{Val: "hello"}, nil
		},
	})
	rt := New(interp)
	rt.SetEnv(runtime.NewEnvironment(nil))
	rt.SetGlobalLookupFallbackEnabled(false)

	if _, err := CallNamed(rt, "greet", nil); err == nil {
		t.Fatalf("expected CallNamed to fail when global fallback is disabled")
	}
}

func TestGetCanDisableGlobalEnvironmentFallback(t *testing.T) {
	interp := interpreter.New()
	interp.GlobalEnvironment().Define("greet", runtime.NativeFunctionValue{
		Name:  "greet",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return runtime.StringValue{Val: "hello"}, nil
		},
	})
	rt := New(interp)
	rt.SetEnv(runtime.NewEnvironment(nil))
	rt.SetGlobalLookupFallbackEnabled(false)

	if _, err := Get(rt, "greet"); err == nil {
		t.Fatalf("expected Get to fail when global fallback is disabled")
	}
}

func TestGlobalLookupFallbackCounters(t *testing.T) {
	interp := interpreter.New()
	interp.GlobalEnvironment().Define("greet", runtime.NativeFunctionValue{
		Name:  "greet",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return runtime.StringValue{Val: "hello"}, nil
		},
	})
	rt := New(interp)
	rt.SetEnv(runtime.NewEnvironment(nil))

	ResetGlobalLookupFallbackCounters()
	if _, err := rt.Call("greet", nil); err != nil {
		t.Fatalf("Call error: %v", err)
	}
	if got := GlobalLookupFallbackStats(); got != 1 {
		t.Fatalf("GlobalLookupFallbackStats = %d, want 1", got)
	}
	envCalls, registryCalls := GlobalLookupFallbackBucketStats()
	if envCalls != 1 || registryCalls != 0 {
		t.Fatalf("GlobalLookupFallbackBucketStats = (%d, %d), want (1, 0)", envCalls, registryCalls)
	}
	snapshot := GlobalLookupFallbackSnapshot()
	if snapshot == "" || !strings.Contains(snapshot, "call:greet=1") {
		t.Fatalf("GlobalLookupFallbackSnapshot = %q, want call:greet entry", snapshot)
	}

	ResetGlobalLookupFallbackCounters()
	if got := GlobalLookupFallbackStats(); got != 0 {
		t.Fatalf("GlobalLookupFallbackStats after reset = %d, want 0", got)
	}
	envCalls, registryCalls = GlobalLookupFallbackBucketStats()
	if envCalls != 0 || registryCalls != 0 {
		t.Fatalf("GlobalLookupFallbackBucketStats after reset = (%d, %d), want (0, 0)", envCalls, registryCalls)
	}
	if snapshot := GlobalLookupFallbackSnapshot(); snapshot != "" {
		t.Fatalf("GlobalLookupFallbackSnapshot after reset = %q, want empty", snapshot)
	}
}

func TestMemberGetPreferMethodsCounters(t *testing.T) {
	rt := New(interpreter.New())
	ResetMemberGetPreferMethodsCounters()

	_, _ = MemberGetPreferMethods(rt, runtime.StringValue{Val: "hello"}, runtime.StringValue{Val: "len"})
	calls, interfaceCalls := MemberGetPreferMethodsStats()
	if calls != 1 || interfaceCalls != 0 {
		t.Fatalf("MemberGetPreferMethodsStats after non-interface = (%d, %d), want (1, 0)", calls, interfaceCalls)
	}

	_, _ = MemberGetPreferMethods(rt, runtime.InterfaceValue{Underlying: runtime.StringValue{Val: "hello"}}, runtime.StringValue{Val: "len"})
	calls, interfaceCalls = MemberGetPreferMethodsStats()
	if calls != 2 || interfaceCalls != 1 {
		t.Fatalf("MemberGetPreferMethodsStats after interface = (%d, %d), want (2, 1)", calls, interfaceCalls)
	}

	ResetMemberGetPreferMethodsCounters()
	calls, interfaceCalls = MemberGetPreferMethodsStats()
	if calls != 0 || interfaceCalls != 0 {
		t.Fatalf("MemberGetPreferMethodsStats after reset = (%d, %d), want (0, 0)", calls, interfaceCalls)
	}
}

func TestCallNamedWithQualifiedResolverBypassesMemberLookup(t *testing.T) {
	interp := interpreter.New()
	rt := New(interp)
	env := runtime.NewEnvironment(nil)
	rt.SetEnv(env)
	rt.SetQualifiedCallableResolver(func(name string, env *runtime.Environment) (runtime.Value, bool, error) {
		if name != "Fancy.describe" {
			return nil, false, nil
		}
		fn := runtime.NativeFunctionValue{
			Name:  "describe",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				return runtime.StringValue{Val: "ok"}, nil
			},
		}
		return fn, true, nil
	})

	ResetMemberGetPreferMethodsCounters()
	value, err := CallNamedWithNode(rt, "Fancy.describe", nil, nil)
	if err != nil {
		t.Fatalf("CallNamedWithNode error: %v", err)
	}
	if got, ok := value.(runtime.StringValue); !ok || got.Val != "ok" {
		t.Fatalf("CallNamedWithNode = %#v, want String(\"ok\")", value)
	}
	calls, interfaceCalls := MemberGetPreferMethodsStats()
	if calls != 0 || interfaceCalls != 0 {
		t.Fatalf("MemberGetPreferMethodsStats = (%d, %d), want (0, 0)", calls, interfaceCalls)
	}
}
