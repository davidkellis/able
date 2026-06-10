package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeArrayOwnershipObserverTracksCanonicalCreationPaths(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.enableBytecodeArrayOwnershipObserverForTest()

	vm.stack = append(vm.stack, runtime.NewSmallInt(7, runtime.IntegerI32))
	if err := vm.execArrayLiteral(&bytecodeInstruction{argCount: 1}); err != nil {
		t.Fatalf("execArrayLiteral: %v", err)
	}
	if got := vm.bytecodeArrayOwnershipSnapshot().Created; got != 1 {
		t.Fatalf("literal creations = %d, want 1", got)
	}

	vm.stack = append(vm.stack, runtime.NilValue{})
	if _, handled, err := vm.finishStaticArrayNewMemberFast(bytecodeInstruction{}, len(vm.stack)-1, nil); err != nil || !handled {
		t.Fatalf("finishStaticArrayNewMemberFast = (handled=%v, err=%v), want handled/nil", handled, err)
	}
	if got := vm.bytecodeArrayOwnershipSnapshot().Created; got != 2 {
		t.Fatalf("Array.new creations = %d, want 2", got)
	}

	arrayDef := ast.StructDef(
		"Array",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i64"), "storage_handle"),
			ast.FieldDef(ast.Ty("i32"), "length"),
			ast.FieldDef(ast.Ty("i32"), "capacity"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	handle := runtime.ArrayStoreNewReservedCapacity(2)
	instr := bytecodeInstruction{
		op:       bytecodeOpStructLiteralNamedFast,
		argCount: 3,
		node: ast.StructLit([]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(0), "storage_handle"),
			ast.FieldInit(ast.Int(0), "length"),
			ast.FieldInit(ast.Int(0), "capacity"),
		}, false, "Array", nil, nil),
	}
	program := &bytecodeProgram{namedStructLiterals: map[int]bytecodeNamedStructLiteralPlan{
		vm.ip: {definition: &runtime.StructDefinitionValue{Node: arrayDef}, fieldOrder: []int{0, 1, 2}},
	}}
	vm.stack = append(vm.stack,
		runtime.NewSmallInt(handle, runtime.IntegerI64),
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	)
	if err := vm.execStructLiteralNamedFast(&instr, program); err != nil {
		t.Fatalf("execStructLiteralNamedFast(Array): %v", err)
	}
	if got := vm.bytecodeArrayOwnershipSnapshot().Created; got != 3 {
		t.Fatalf("kernel Array literal creations = %d, want 3", got)
	}
}

func TestBytecodeArrayOwnershipProfileCollectsReleasedVMObserver(t *testing.T) {
	interp := NewBytecode()
	profile := interp.enableBytecodeArrayOwnershipProfile()
	defer interp.disableBytecodeArrayOwnershipProfile()
	vm := interp.acquireBytecodeVM(interp.GlobalEnvironment())
	if vm.arrayOwnershipObserver != nil {
		t.Fatal("profile should not attach an observer before eligible bytecode runs")
	}
	program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpArrayLiteral, argCount: 0},
		{op: bytecodeOpReturn},
	}})
	if _, err := vm.run(program); err != nil {
		t.Fatalf("run array literal: %v", err)
	}
	interp.releaseBytecodeVM(vm)

	snapshot := profile.snapshot()
	if snapshot.Created != 1 {
		t.Fatalf("profile creations = %d, want 1", snapshot.Created)
	}
	profile.reset()
	if snapshot := profile.snapshot(); snapshot.Created != 0 || snapshot.Escaped != 0 {
		t.Fatalf("profile reset = %#v, want empty snapshot", snapshot)
	}
}

func TestBytecodeArrayOwnershipProfileAdoptsDetachedKernelReturn(t *testing.T) {
	interp := NewBytecode()
	profile := interp.enableBytecodeArrayOwnershipProfile()
	defer interp.disableBytecodeArrayOwnershipProfile()

	caller := interp.acquireBytecodeVM(interp.GlobalEnvironment())
	defer interp.releaseBytecodeVM(caller)
	callee := interp.acquireBytecodeVM(interp.GlobalEnvironment())
	result, err := callee.runDetached(finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpArrayLiteral, argCount: 0},
		{op: bytecodeOpReturn},
	}}))
	if err != nil {
		t.Fatalf("detached array literal: %v", err)
	}
	arr, ok := result.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("detached result = %T, want *runtime.ArrayValue", result)
	}
	interp.releaseBytecodeVM(callee)

	caller.adoptBytecodeArrayOwnershipReturnedValue(arr)
	caller.finishBytecodeArrayOwnershipPublicReturn(runtime.NilValue{}, false)
	snapshot := profile.snapshot()
	if snapshot.Created != 1 || snapshot.Transferred != 1 || snapshot.FrameLocal != 1 || snapshot.PublicReturned != 0 {
		t.Fatalf("detached adoption snapshot = %#v, want adopted frame-local Array", snapshot)
	}
}

func TestBytecodeArrayOwnershipObserverClassifiesExplicitVMReturn(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.enableBytecodeArrayOwnershipObserverForTest()
	arr := interp.newArrayValue(nil, 0)
	vm.trackBytecodeArrayOwnershipCreation(arr)

	_, err := vm.run(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpConst, value: arr},
		{op: bytecodeOpReturn, node: ast.Int(0)},
	}})
	if _, ok := err.(returnSignal); !ok {
		t.Fatalf("explicit return error = %T (%v), want returnSignal", err, err)
	}
	snapshot := vm.bytecodeArrayOwnershipSnapshot()
	if snapshot.PublicReturned != 1 || snapshot.ErrorUnwound != 0 {
		t.Fatalf("explicit public return snapshot = %#v, want one public return", snapshot)
	}
}

func TestBytecodeArrayOwnershipProfileAdoptsExplicitDetachedReturn(t *testing.T) {
	interp := NewBytecode()
	profile := interp.enableBytecodeArrayOwnershipProfile()
	defer interp.disableBytecodeArrayOwnershipProfile()

	caller := interp.acquireBytecodeVM(interp.GlobalEnvironment())
	defer interp.releaseBytecodeVM(caller)
	callee := interp.acquireBytecodeVM(interp.GlobalEnvironment())
	defer interp.releaseBytecodeVM(callee)
	_, err := callee.runDetached(finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpArrayLiteral, argCount: 0},
		{op: bytecodeOpReturn, node: ast.Int(0)},
	}}))
	returned, ok := err.(returnSignal)
	if !ok {
		t.Fatalf("detached explicit return error = %T (%v), want returnSignal", err, err)
	}
	caller.adoptBytecodeArrayOwnershipReturnedValue(returned.value)
	caller.finishBytecodeArrayOwnershipPublicReturn(runtime.NilValue{}, false)

	snapshot := profile.snapshot()
	if snapshot.Created != 1 || snapshot.Transferred != 1 || snapshot.FrameLocal != 1 || snapshot.PublicReturned != 0 || snapshot.ErrorUnwound != 0 {
		t.Fatalf("explicit detached adoption snapshot = %#v, want adopted frame-local Array", snapshot)
	}
}

func TestBytecodeArrayOwnershipObserverTransfersReturnsAndClassifiesBoundaries(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.enableBytecodeArrayOwnershipObserverForTest()

	root := interp.newArrayValue(nil, 0)
	vm.trackBytecodeArrayOwnershipCreation(root)
	vm.pushCallFrame(1, nil, nil, vm.env, nil, 0, 0, false, false)
	local := interp.newArrayValue(nil, 0)
	returned := interp.newArrayValue([]runtime.Value{local}, 1)
	vm.trackBytecodeArrayOwnershipCreation(local)
	vm.trackBytecodeArrayOwnershipCreation(returned)
	vm.finishBytecodeArrayOwnershipReturn(returned, vm.topBytecodeArrayOwnershipParent())
	if _, _, _, _, _, _, _, _, _, ok := vm.popCallFrameFields(); !ok {
		t.Fatal("expected ownership test call frame")
	}

	// The nested local graph returns as one pointer graph, so both wrappers
	// transfer to the parent instead of becoming a frame-local candidate.
	snapshot := vm.bytecodeArrayOwnershipSnapshot()
	if snapshot.Transferred != 2 || snapshot.FrameLocal != 0 {
		t.Fatalf("return transfer snapshot = %#v, want two transfers and no local candidates", snapshot)
	}

	borrowed := interp.newArrayValue(nil, 0)
	aggregate := interp.newArrayValue(nil, 0)
	closure := interp.newArrayValue(nil, 0)
	future := interp.newArrayValue(nil, 0)
	unknown := interp.newArrayValue(nil, 0)
	vm.trackBytecodeArrayOwnershipCreation(aggregate)
	vm.trackBytecodeArrayOwnershipCreation(closure)
	vm.trackBytecodeArrayOwnershipCreation(future)
	vm.trackBytecodeArrayOwnershipCreation(unknown)
	vm.observeBytecodeArrayOwnershipArrayWrite(borrowed, returned)
	vm.markBytecodeArrayOwnershipValueEscaped(root, bytecodeArrayOwnershipEscapeEnvironment)
	vm.markBytecodeArrayOwnershipValueEscaped(aggregate, bytecodeArrayOwnershipEscapeAggregate)
	vm.markBytecodeArrayOwnershipValueEscaped(closure, bytecodeArrayOwnershipEscapeClosure)
	vm.markBytecodeArrayOwnershipValueEscaped(future, bytecodeArrayOwnershipEscapeFuture)
	vm.markBytecodeArrayOwnershipValuesEscaped([]runtime.Value{unknown}, bytecodeArrayOwnershipEscapeUnknownCall)
	vm.finishBytecodeArrayOwnershipPublicReturn(runtime.NilValue{}, false)

	snapshot = vm.bytecodeArrayOwnershipSnapshot()
	if snapshot.Escaped != 7 {
		t.Fatalf("escaped wrappers = %d, want every boundary-owned wrapper", snapshot.Escaped)
	}
	if snapshot.Escapes[bytecodeArrayOwnershipEscapeEnvironment] != 1 {
		t.Fatalf("environment escapes = %#v, want one root escape", snapshot.Escapes)
	}
	if snapshot.Escapes[bytecodeArrayOwnershipEscapeBorrowedArrayWrite] != 2 {
		t.Fatalf("borrowed-write escapes = %#v, want returned graph", snapshot.Escapes)
	}
	for _, reason := range []bytecodeArrayOwnershipEscape{
		bytecodeArrayOwnershipEscapeAggregate,
		bytecodeArrayOwnershipEscapeClosure,
		bytecodeArrayOwnershipEscapeFuture,
		bytecodeArrayOwnershipEscapeUnknownCall,
	} {
		if snapshot.Escapes[reason] != 1 {
			t.Fatalf("escape reason %d count = %d, want 1", reason, snapshot.Escapes[reason])
		}
	}
}

func TestBytecodeArrayOwnershipObserverTransfersToEmptyCallerFrame(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.enableBytecodeArrayOwnershipObserverForTest()

	// The caller has no direct Array allocation before entering its child. It
	// still needs an ownership frame so the child's returned wrapper is local to
	// the caller rather than misclassified as a public VM result.
	vm.pushCallFrame(1, nil, nil, vm.env, nil, 0, 0, false, false)
	vm.pushCallFrame(2, nil, nil, vm.env, nil, 0, 0, false, false)
	arr := interp.newArrayValue(nil, 0)
	vm.trackBytecodeArrayOwnershipCreation(arr)
	vm.finishBytecodeArrayOwnershipReturn(arr, vm.topBytecodeArrayOwnershipParent())
	if _, _, _, _, _, _, _, _, _, ok := vm.popCallFrameFields(); !ok {
		t.Fatal("expected nested ownership test call frame")
	}

	snapshot := vm.bytecodeArrayOwnershipSnapshot()
	if snapshot.Created != 1 || snapshot.Transferred != 1 || snapshot.PublicReturned != 0 {
		t.Fatalf("empty-parent transfer snapshot = %#v, want one non-public transfer", snapshot)
	}

	vm.finishBytecodeArrayOwnershipPublicReturn(runtime.NilValue{}, false)
	snapshot = vm.bytecodeArrayOwnershipSnapshot()
	if snapshot.FrameLocal != 1 || snapshot.PublicReturned != 0 {
		t.Fatalf("empty-parent completion snapshot = %#v, want one frame-local Array", snapshot)
	}
}

func TestBytecodeArrayOwnershipObserverClassifiesErrorUnwindWithoutRelease(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.enableBytecodeArrayOwnershipObserverForTest()

	vm.pushCallFrame(1, nil, nil, vm.env, nil, 0, 0, false, false)
	arr := interp.newArrayValue(nil, 0)
	vm.trackBytecodeArrayOwnershipCreation(arr)
	vm.finishBytecodeArrayOwnershipError(vm.topBytecodeArrayOwnershipParent())
	if _, _, _, _, _, _, _, _, _, ok := vm.popCallFrameFields(); !ok {
		t.Fatal("expected ownership test error frame")
	}

	snapshot := vm.bytecodeArrayOwnershipSnapshot()
	if snapshot.ErrorUnwound != 1 || snapshot.FrameLocal != 0 || snapshot.Escaped != 0 {
		t.Fatalf("error unwind snapshot = %#v, want one unreleased error candidate", snapshot)
	}
	if _, err := runtime.ArrayStoreSize(arr.Handle); err != nil {
		t.Fatalf("observer must not release ArrayStore lease during error unwind: %v", err)
	}
}

func TestBytecodeArrayOwnershipObserverDoesNotTransferEscapedReturn(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.enableBytecodeArrayOwnershipObserverForTest()

	parent := interp.newArrayValue(nil, 0)
	vm.trackBytecodeArrayOwnershipCreation(parent)
	vm.pushCallFrame(1, nil, nil, vm.env, nil, 0, 0, false, false)
	escaped := interp.newArrayValue(nil, 0)
	vm.trackBytecodeArrayOwnershipCreation(escaped)
	vm.markBytecodeArrayOwnershipValueEscaped(escaped, bytecodeArrayOwnershipEscapeUnknownCall)
	vm.finishBytecodeArrayOwnershipReturn(escaped, vm.topBytecodeArrayOwnershipParent())
	if _, _, _, _, _, _, _, _, _, ok := vm.popCallFrameFields(); !ok {
		t.Fatal("expected ownership test call frame")
	}

	snapshot := vm.bytecodeArrayOwnershipSnapshot()
	if snapshot.Transferred != 0 || snapshot.Escaped != 1 || snapshot.Escapes[bytecodeArrayOwnershipEscapeUnknownCall] != 1 {
		t.Fatalf("escaped return snapshot = %#v, want unknown-call escape without transfer", snapshot)
	}
}
