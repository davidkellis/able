package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_PrepareResolvedFunctionCallArgsWithOptionalReceiverInlineScratchHotAllocFree(t *testing.T) {
	vm := newBytecodeVM(NewBytecode(), runtime.NewEnvironment(nil))
	args := []runtime.Value{
		runtime.NewSmallInt(7, runtime.IntegerI32),
	}
	receiver := runtime.StringValue{Val: "self"}

	got := vm.prepareResolvedFunctionCallArgsWithOptionalReceiver(args, false, receiver, true)
	if len(got) != 2 {
		t.Fatalf("prepared arg count = %d, want 2", len(got))
	}
	if !valuesEqual(got[0], receiver) {
		t.Fatalf("prepared receiver = %#v, want %#v", got[0], receiver)
	}
	if &got[0] == &args[0] {
		t.Fatalf("expected prepared receiver path to use VM scratch instead of aliasing explicit args")
	}
	got2 := vm.prepareResolvedFunctionCallArgsWithOptionalReceiver(args, false, receiver, true)
	if &got[0] != &got2[0] {
		t.Fatalf("expected repeated inline scratch calls to reuse the same VM buffer")
	}
}

func TestBytecodeVM_PrepareResolvedFunctionCallArgsWithOptionalReceiverSpillScratchHotAllocFree(t *testing.T) {
	vm := newBytecodeVM(NewBytecode(), runtime.NewEnvironment(nil))
	args := make([]runtime.Value, bytecodeInlinePreparedCallArgStorage+1)
	for idx := range args {
		args[idx] = runtime.NewSmallInt(int64(idx+1), runtime.IntegerI32)
	}

	prepared := vm.prepareResolvedFunctionCallArgsWithOptionalReceiver(args, true, nil, false)
	if len(prepared) != len(args) {
		t.Fatalf("prepared arg count = %d, want %d", len(prepared), len(args))
	}
	if cap(vm.resolvedCallArgsSpill) < len(args) {
		t.Fatalf("expected spill scratch capacity >= %d, got %d", len(args), cap(vm.resolvedCallArgsSpill))
	}
	prepared2 := vm.prepareResolvedFunctionCallArgsWithOptionalReceiver(args, true, nil, false)
	if &prepared[0] != &prepared2[0] {
		t.Fatalf("expected repeated spill scratch calls to reuse the same VM buffer")
	}
}

func TestBytecodeVM_ResetForRunDropsLargeResolvedCallArgScratch(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironment(nil)
	vm := newBytecodeVM(interp, env)
	vm.resolvedCallArgsSpill = make([]runtime.Value, bytecodeResolvedCallArgRetainLimit+1)
	vm.resolvedCallArgsInline[0] = runtime.StringValue{Val: "keepalive"}

	vm.resetForRun(interp, env)

	if vm.resolvedCallArgsSpill != nil {
		t.Fatalf("expected large resolved-call scratch spill to be released, got cap %d", cap(vm.resolvedCallArgsSpill))
	}
	if vm.resolvedCallArgsInline[0] != nil {
		t.Fatalf("expected inline resolved-call scratch to be cleared, got %#v", vm.resolvedCallArgsInline[0])
	}
}

func TestBytecodeVM_PrepareResolvedFunctionCallArgsPreservesRawI32StackCell(t *testing.T) {
	vm := newBytecodeVM(NewBytecode(), runtime.NewEnvironment(nil))
	raw := vm.stackRawI32Value(0, int32(bytecodeRawI32SlotCacheMax+17))
	args := []runtime.Value{raw}

	got := vm.prepareResolvedFunctionCallArgsWithOptionalReceiver(args, false, nil, false)
	if len(got) != 1 {
		t.Fatalf("prepared arg count = %d, want 1", len(got))
	}
	cell, ok := got[0].(*bytecodeRawI32StackCell)
	if !ok || cell == nil || cell.Val != int32(bytecodeRawI32SlotCacheMax+17) {
		t.Fatalf("prepared arg = %#v, want reusable raw i32 stack cell", got[0])
	}
	if _, boxed := got[0].(runtime.IntegerValue); boxed {
		t.Fatalf("prepared arg should stay raw, got boxed %#v", got[0])
	}
}
