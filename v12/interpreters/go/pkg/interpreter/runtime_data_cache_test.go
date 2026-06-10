package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestInterpreterRuntimeDataFromEnvCacheInvalidatesAcrossParentAndChildUpdates(t *testing.T) {
	interp := New()
	parent := runtime.NewEnvironment(nil)
	parent.SetSingleThread()
	child := runtime.NewEnvironment(parent)

	if got := interp.runtimeDataFromEnv(child); got != nil {
		t.Fatalf("initial runtimeDataFromEnv(child) = %#v, want nil", got)
	}

	parent.SetRuntimeData("root-data")
	if got := interp.runtimeDataFromEnv(child); got != "root-data" {
		t.Fatalf("runtimeDataFromEnv(child) after parent set = %#v, want root-data", got)
	}

	child.SetRuntimeData("child-data")
	if got := interp.runtimeDataFromEnv(child); got != "child-data" {
		t.Fatalf("runtimeDataFromEnv(child) after child override = %#v, want child-data", got)
	}

	child.SetRuntimeData(nil)
	if got := interp.runtimeDataFromEnv(child); got != "root-data" {
		t.Fatalf("runtimeDataFromEnv(child) after child clear = %#v, want root-data", got)
	}

	parent.SetRuntimeData("next-root")
	if got := interp.runtimeDataFromEnv(child); got != "next-root" {
		t.Fatalf("runtimeDataFromEnv(child) after parent update = %#v, want next-root", got)
	}

	parent.SetRuntimeData(nil)
	if got := interp.runtimeDataFromEnv(child); got != nil {
		t.Fatalf("runtimeDataFromEnv(child) after parent clear = %#v, want nil", got)
	}
}

func TestBytecodeVMRuntimeDataCacheInvalidatesAcrossParentAndChildUpdates(t *testing.T) {
	interp := NewBytecode()
	parent := runtime.NewEnvironment(nil)
	parent.SetSingleThread()
	child := runtime.NewEnvironment(parent)
	vm := newBytecodeVM(interp, child)

	if got := vm.runtimeData(); got != nil {
		t.Fatalf("initial vm.runtimeData() = %#v, want nil", got)
	}

	parent.SetRuntimeData("root-data")
	if got := vm.runtimeData(); got != "root-data" {
		t.Fatalf("vm.runtimeData() after parent set = %#v, want root-data", got)
	}

	child.SetRuntimeData("child-data")
	if got := vm.runtimeData(); got != "child-data" {
		t.Fatalf("vm.runtimeData() after child override = %#v, want child-data", got)
	}

	child.SetRuntimeData(nil)
	if got := vm.runtimeData(); got != "root-data" {
		t.Fatalf("vm.runtimeData() after child clear = %#v, want root-data", got)
	}

	parent.SetRuntimeData("next-root")
	if got := vm.runtimeData(); got != "next-root" {
		t.Fatalf("vm.runtimeData() after parent update = %#v, want next-root", got)
	}

	parent.SetRuntimeData(nil)
	if got := vm.runtimeData(); got != nil {
		t.Fatalf("vm.runtimeData() after parent clear = %#v, want nil", got)
	}
}

func TestInterpreterImplMethodContextFromEnvUsesRuntimeDataCache(t *testing.T) {
	interp := New()
	env := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 1)
	ctx := &implMethodContext{interfaceName: "I"}
	env.SetRuntimeData(ctx)

	if got := interp.implMethodContextFromEnv(env); got != ctx {
		t.Fatalf("implMethodContextFromEnv(env) = %#v, want %#v", got, ctx)
	}
	if !interp.runtimeDataCacheKnown {
		t.Fatalf("expected implMethodContextFromEnv to seed interpreter runtime-data cache")
	}
	if interp.runtimeDataCacheEnv != env {
		t.Fatalf("runtime-data cache env = %#v, want %#v", interp.runtimeDataCacheEnv, env)
	}
	if interp.runtimeDataCacheValue != ctx {
		t.Fatalf("runtime-data cache value = %#v, want %#v", interp.runtimeDataCacheValue, ctx)
	}

	next := &implMethodContext{interfaceName: "J"}
	env.SetRuntimeData(next)
	if got := interp.implMethodContextFromEnv(env); got != next {
		t.Fatalf("implMethodContextFromEnv(env) after update = %#v, want %#v", got, next)
	}
	if interp.runtimeDataCacheValue != next {
		t.Fatalf("runtime-data cache value after update = %#v, want %#v", interp.runtimeDataCacheValue, next)
	}
}

func TestInterpreterRuntimeDataFromEnvCacheDistinguishesReusedEnvAcrossSharedStates(t *testing.T) {
	interp := New()

	firstParent := runtime.NewEnvironment(nil)
	firstParent.SetSingleThread()
	firstParent.SetRuntimeData("first")

	reused := runtime.NewEnvironment(firstParent)
	if got := interp.runtimeDataFromEnv(reused); got != "first" {
		t.Fatalf("runtimeDataFromEnv(reused) first parent = %#v, want first", got)
	}

	secondParent := runtime.NewEnvironment(nil)
	secondParent.SetSingleThread()
	secondParent.SetRuntimeData("second")

	reused.ResetForSingleBindingReuse(secondParent, 0, "", nil)
	if got := interp.runtimeDataFromEnv(reused); got != "second" {
		t.Fatalf("runtimeDataFromEnv(reused) second parent = %#v, want second", got)
	}
}

func TestBytecodeVMRuntimeDataCacheDistinguishesReusedEnvAcrossSharedStates(t *testing.T) {
	interp := NewBytecode()

	firstParent := runtime.NewEnvironment(nil)
	firstParent.SetSingleThread()
	firstParent.SetRuntimeData("first")

	reused := runtime.NewEnvironment(firstParent)
	vm := newBytecodeVM(interp, reused)
	if got := vm.runtimeData(); got != "first" {
		t.Fatalf("vm.runtimeData() first parent = %#v, want first", got)
	}

	secondParent := runtime.NewEnvironment(nil)
	secondParent.SetSingleThread()
	secondParent.SetRuntimeData("second")

	reused.ResetForSingleBindingReuse(secondParent, 0, "", nil)
	if got := vm.runtimeData(); got != "second" {
		t.Fatalf("vm.runtimeData() second parent = %#v, want second", got)
	}
}

func TestInterpreterRuntimeDataCacheDistinguishesReusedEnvAcrossSameSharedState(t *testing.T) {
	interp := New()
	root := runtime.NewEnvironment(nil)
	root.SetSingleThread()

	firstParent := runtime.NewEnvironment(root)
	firstParent.SetRuntimeData("first")
	secondParent := runtime.NewEnvironment(root)
	secondParent.SetRuntimeData("second")

	reused := runtime.NewEnvironment(firstParent)
	if got := interp.runtimeDataFromEnv(reused); got != "first" {
		t.Fatalf("runtimeDataFromEnv(reused) first parent = %#v, want first", got)
	}

	reused.ResetForSingleBindingReuse(secondParent, 0, "", nil)
	if got := interp.runtimeDataFromEnv(reused); got != "second" {
		t.Fatalf("runtimeDataFromEnv(reused) second parent = %#v, want second", got)
	}
}

func TestBytecodeVMRuntimeDataCacheDistinguishesReusedEnvAcrossSameSharedState(t *testing.T) {
	interp := NewBytecode()
	root := runtime.NewEnvironment(nil)
	root.SetSingleThread()

	firstParent := runtime.NewEnvironment(root)
	firstParent.SetRuntimeData("first")
	secondParent := runtime.NewEnvironment(root)
	secondParent.SetRuntimeData("second")

	reused := runtime.NewEnvironment(firstParent)
	vm := newBytecodeVM(interp, reused)
	if got := vm.runtimeData(); got != "first" {
		t.Fatalf("vm.runtimeData() first parent = %#v, want first", got)
	}

	reused.ResetForSingleBindingReuse(secondParent, 0, "", nil)
	if got := vm.runtimeData(); got != "second" {
		t.Fatalf("vm.runtimeData() second parent = %#v, want second", got)
	}
}
