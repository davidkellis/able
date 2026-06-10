package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CastOpcodeFastProducesRawU64FromRawU32(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{bytecodeRawIntegerResultValue(runtime.IntegerU32, 17)}

	if err := vm.execCastOpcode(&bytecodeInstruction{
		op:   bytecodeOpCast,
		node: ast.NewTypeCastExpression(ast.Int(17), ast.Ty("u64")),
	}); err != nil {
		t.Fatalf("execCastOpcode failed: %v", err)
	}

	gotKind, gotRaw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || gotKind != runtime.IntegerU64 || gotRaw != 17 {
		t.Fatalf("cast result = %#v, want raw 17_u64", vm.stack[0])
	}
	if _, boxed := vm.stack[0].(runtime.IntegerValue); boxed {
		t.Fatalf("cast result should stay raw inside the VM, got boxed %#v", vm.stack[0])
	}
}

func TestBytecodeVM_StoreSlotTypedExactRawIntegerPreservesRawU64(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = make([]runtime.Value, 1)
	vm.stack = []runtime.Value{bytecodeRawIntegerResultValue(runtime.IntegerU64, 42)}

	if err := vm.execStoreSlot(&bytecodeInstruction{
		op:         bytecodeOpStoreSlot,
		target:     0,
		storeTyped: true,
		typeExpr:   cachedSimpleTypeExpression("u64"),
	}); err != nil {
		t.Fatalf("execStoreSlot failed: %v", err)
	}

	if _, ok := vm.slots[0].(*bytecodeRawIntegerSlotCell); ok {
		t.Fatalf("small raw u64 slot should reuse shared boxed value, got raw cell %#v", vm.slots[0])
	}
	assertIntValue(t, vm.slots[0], runtime.IntegerU64, 42)

	stackKind, stackRaw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || stackKind != runtime.IntegerU64 || stackRaw != 42 {
		t.Fatalf("stack result = %#v, want raw 42_u64", vm.stack[0])
	}

	vm.stack = nil
	if err := vm.execLoadSlotOpcode(&bytecodeInstruction{op: bytecodeOpLoadSlot, target: 0}); err != nil {
		t.Fatalf("execLoadSlotOpcode failed: %v", err)
	}
	loadKind, loadRaw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || loadKind != runtime.IntegerU64 || loadRaw != 42 {
		t.Fatalf("loaded value = %#v, want raw 42_u64", vm.stack[0])
	}
	assertIntValue(t, vm.stack[0], runtime.IntegerU64, 42)
}

func TestBytecodeVM_StoreSlotTypedExactLargeRawUnsignedIntegerStillUsesRawSlotCell(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = make([]runtime.Value, 1)
	vm.stack = []runtime.Value{bytecodeRawIntegerResultValue(runtime.IntegerU64, 65536)}

	if err := vm.execStoreSlot(&bytecodeInstruction{
		op:         bytecodeOpStoreSlot,
		target:     0,
		storeTyped: true,
		typeExpr:   cachedSimpleTypeExpression("u64"),
	}); err != nil {
		t.Fatalf("execStoreSlot failed: %v", err)
	}

	cell, ok := vm.slots[0].(*bytecodeRawIntegerSlotCell)
	if !ok || cell == nil || cell.TypeSuffix != runtime.IntegerU64 || cell.Raw != 65536 {
		t.Fatalf("stored slot = %#v, want raw u64 cell 65536", vm.slots[0])
	}
}

func TestBytecodeVM_StoreRawIntegerSlotReusesReleasedCell(t *testing.T) {
	pooled := &bytecodeRawIntegerSlotCell{Raw: 1, TypeSuffix: runtime.IntegerU64}
	vm := &bytecodeVM{
		slots:                  make([]runtime.Value, 1),
		rawIntegerSlotCellPool: []*bytecodeRawIntegerSlotCell{pooled},
	}

	got := vm.storeRawIntegerSlot(0, runtime.IntegerU64, 65536)
	if got != pooled || vm.slots[0] != pooled {
		t.Fatalf("stored value = %#v, want pooled raw integer cell %#v", got, pooled)
	}
	if pooled.TypeSuffix != runtime.IntegerU64 || pooled.Raw != 65536 {
		t.Fatalf("pooled cell = %#v, want u64 65536", pooled)
	}
	if len(vm.rawIntegerSlotCellPool) != 0 {
		t.Fatalf("raw integer pool length = %d, want 0 after reuse", len(vm.rawIntegerSlotCellPool))
	}
}

func TestBytecodeVM_InlineCallArgSnapshotDoesNotAliasRawIntegerStackCell(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	source := vm.stackRawIntegerValue(0, runtime.IntegerU64, 1048576)
	slots := make([]runtime.Value, 1)

	vm.copyInlineCallArgToSlot(slots, 0, source)
	stackCell, ok := source.(*bytecodeRawIntegerSlotCell)
	if !ok || stackCell == nil {
		t.Fatalf("source = %#v, want raw integer stack cell", source)
	}
	stackCell.Raw = 1

	kind, raw, ok := bytecodeRawIntegerValueInfo(slots[0])
	if !ok || kind != runtime.IntegerU64 || raw != 1048576 {
		t.Fatalf("inline argument snapshot = %#v, want independent raw u64 1048576", slots[0])
	}
}

func TestBytecodeVM_StoreRawI64SlotReusesReleasedCell(t *testing.T) {
	pooled := &bytecodeRawI64SlotCell{Val: 1}
	vm := &bytecodeVM{
		slots:              make([]runtime.Value, 1),
		rawI64SlotCellPool: []*bytecodeRawI64SlotCell{pooled},
	}

	got := vm.storeRawI64Slot(0, 42)
	if got != pooled || vm.slots[0] != pooled {
		t.Fatalf("stored value = %#v, want pooled raw i64 cell %#v", got, pooled)
	}
	if pooled.Val != 42 {
		t.Fatalf("pooled cell = %#v, want i64 42", pooled)
	}
	if len(vm.rawI64SlotCellPool) != 0 {
		t.Fatalf("raw i64 pool length = %d, want 0 after reuse", len(vm.rawI64SlotCellPool))
	}
}

func TestBytecodeVM_StoreSlotUntypedDirectRawIntegerCarriers(t *testing.T) {
	tests := []struct {
		name  string
		value runtime.Value
		kind  runtime.IntegerType
		raw   int64
	}{
		{name: "i32", value: bytecodeRawI32SlotValue(37), kind: runtime.IntegerI32, raw: 37},
		{name: "i32 stack cell", value: &bytecodeRawI32StackCell{Val: 41}, kind: runtime.IntegerI32, raw: 41},
		{name: "i64 result", value: bytecodeRawI64ResultValue(43), kind: runtime.IntegerI64, raw: 43},
		{name: "i64 slot cell", value: &bytecodeRawI64SlotCell{Val: 47}, kind: runtime.IntegerI64, raw: 47},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vm := newBytecodeVM(nil, nil)
			vm.slots = make([]runtime.Value, 1)
			vm.stack = []runtime.Value{tc.value}

			if err := vm.execStoreSlot(&bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: 0}); err != nil {
				t.Fatalf("execStoreSlot failed: %v", err)
			}
			kind, raw, ok := bytecodeRawIntegerValueInfo(vm.slots[0])
			if !ok || kind != tc.kind || raw != tc.raw {
				t.Fatalf("stored slot = %#v, want %d_%s", vm.slots[0], tc.raw, tc.kind)
			}
			stackKind, stackRaw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
			if !ok || stackKind != tc.kind || stackRaw != tc.raw {
				t.Fatalf("store result = %#v, want unchanged %d_%s", vm.stack[0], tc.raw, tc.kind)
			}
		})
	}
}

func TestBytecodeVM_StoreSlotUntypedRawI64DoesNotAliasSourceSlot(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	source := &bytecodeRawI64SlotCell{Val: 53}
	vm.slots = []runtime.Value{source, nil}
	vm.stack = []runtime.Value{source}

	if err := vm.execStoreSlot(&bytecodeInstruction{op: bytecodeOpStoreSlot, target: 1}); err != nil {
		t.Fatalf("execStoreSlot failed: %v", err)
	}
	stored, ok := vm.slots[1].(*bytecodeRawI64SlotCell)
	if !ok || stored == nil || stored == source || stored.Val != 53 {
		t.Fatalf("stored slot = %#v, want independent raw i64 cell 53", vm.slots[1])
	}
	source.Val = 59
	if stored.Val != 53 {
		t.Fatalf("stored slot changed with source: got %d want 53", stored.Val)
	}
}

func TestBytecodeVM_StoreSlotUntypedPrimitiveNonIntegersKeepOrdinaryStorage(t *testing.T) {
	values := []runtime.Value{
		runtime.BoolValue{Val: true},
		runtime.CharValue{Val: 'x'},
		runtime.StringValue{Val: "text"},
		runtime.FloatValue{Val: 1.25, TypeSuffix: runtime.FloatF64},
		bytecodeRawF32SlotValue(2.5),
	}
	for _, value := range values {
		vm := newBytecodeVM(nil, nil)
		vm.slots = make([]runtime.Value, 1)
		vm.stack = []runtime.Value{value}
		if err := vm.execStoreSlot(&bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: 0}); err != nil {
			t.Fatalf("execStoreSlot(%T) failed: %v", value, err)
		}
		if vm.slots[0] == nil || vm.slots[0].Kind() != value.Kind() {
			t.Fatalf("stored slot for %T = %#v, want ordinary %v storage", value, vm.slots[0], value.Kind())
		}
	}
}

func TestBytecodeVM_StoreSlotUntypedRareRawIntegerUsesFallback(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	vm.slots = make([]runtime.Value, 1)
	vm.stack = []runtime.Value{bytecodeRawU64ResultValue(61)}
	if err := vm.execStoreSlot(&bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: 0}); err != nil {
		t.Fatalf("execStoreSlot failed: %v", err)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.slots[0])
	if !ok || kind != runtime.IntegerU64 || raw != 61 {
		t.Fatalf("stored slot = %#v, want fallback raw 61_u64", vm.slots[0])
	}
}

func TestBytecodeVM_TryStoreRawIntegerSlotValuePreservesAllSmallCarriers(t *testing.T) {
	boxedI16 := runtime.NewSmallInt(67, runtime.IntegerI16)
	tests := []struct {
		name  string
		value runtime.Value
		kind  runtime.IntegerType
		raw   int64
	}{
		{name: "u8 result", value: bytecodeRawU8ResultValue(2), kind: runtime.IntegerU8, raw: 2},
		{name: "u16 result", value: bytecodeRawU16ResultValue(3), kind: runtime.IntegerU16, raw: 3},
		{name: "u32 result", value: bytecodeRawU32ResultValue(5), kind: runtime.IntegerU32, raw: 5},
		{name: "u64 result", value: bytecodeRawU64ResultValue(7), kind: runtime.IntegerU64, raw: 7},
		{name: "usize result", value: bytecodeRawUsizeResultValue(11), kind: runtime.IntegerUsize, raw: 11},
		{name: "generic value", value: bytecodeRawIntegerValue{Raw: 13, TypeSuffix: runtime.IntegerI8}, kind: runtime.IntegerI8, raw: 13},
		{name: "generic slot", value: &bytecodeRawIntegerSlotCell{Raw: 17, TypeSuffix: runtime.IntegerU32}, kind: runtime.IntegerU32, raw: 17},
		{name: "return scratch", value: &bytecodeRawIntegerReturnScratch{Raw: 19, TypeSuffix: runtime.IntegerIsize}, kind: runtime.IntegerIsize, raw: 19},
		{name: "boxed value", value: boxedI16, kind: runtime.IntegerI16, raw: 67},
		{name: "boxed pointer", value: &boxedI16, kind: runtime.IntegerI16, raw: 67},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vm := newBytecodeVM(nil, nil)
			vm.slots = make([]runtime.Value, 1)
			stored, ok := vm.tryStoreRawIntegerSlotValue(0, tc.value)
			if !ok || stored == nil {
				t.Fatalf("tryStoreRawIntegerSlotValue(%T) missed", tc.value)
			}
			kind, raw, ok := bytecodeRawIntegerValueInfo(vm.slots[0])
			if !ok || kind != tc.kind || raw != tc.raw {
				t.Fatalf("stored slot = %#v, want %d_%s", vm.slots[0], tc.raw, tc.kind)
			}
		})
	}

	vm := newBytecodeVM(nil, nil)
	vm.slots = make([]runtime.Value, 1)
	if stored, ok := vm.tryStoreRawIntegerSlotValue(0, runtime.StringValue{Val: "not integer"}); ok || stored != nil {
		t.Fatalf("non-integer store unexpectedly handled: %#v", stored)
	}
}

func TestBytecodeVM_ExecBinaryKeepsRawSameTypeDottedBitwiseResult(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{
		bytecodeRawIntegerResultValue(runtime.IntegerU32, 0xF0),
		bytecodeRawIntegerResultValue(runtime.IntegerU32, 0x0F),
	}

	if _, err := vm.execBinary(&bytecodeInstruction{operator: ".^", bitwiseRawCandidate: true}, nil); err != nil {
		t.Fatalf("execBinary dotted xor failed: %v", err)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack length = %d, want 1", len(vm.stack))
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerU32 || raw != 0xFF {
		t.Fatalf("xor result = %#v, want raw u32 255", vm.stack[0])
	}
	if _, boxed := vm.stack[0].(runtime.IntegerValue); boxed {
		t.Fatalf("xor result = %#v, want raw carrier", vm.stack[0])
	}
}

func TestBytecodeVM_ExecBinaryKeepsRawSameTypeDottedShiftResult(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{
		bytecodeRawIntegerResultValue(runtime.IntegerU32, 0x80000000),
		bytecodeRawIntegerResultValue(runtime.IntegerU32, 1),
	}

	if _, err := vm.execBinary(&bytecodeInstruction{operator: ".>>", bitwiseRawCandidate: true}, nil); err != nil {
		t.Fatalf("execBinary dotted shift failed: %v", err)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerU32 || raw != 0x40000000 {
		t.Fatalf("shift result = %#v, want raw u32 0x40000000", vm.stack[0])
	}
}

func TestBytecodeDottedBitwiseOperator(t *testing.T) {
	for _, test := range []struct {
		op   string
		want bool
	}{
		{op: ".&", want: true},
		{op: ".|", want: true},
		{op: ".^", want: true},
		{op: ".<<", want: true},
		{op: ".>>", want: true},
		{op: "&", want: false},
		{op: ".+", want: false},
	} {
		if got := bytecodeDottedBitwiseOperator(test.op); got != test.want {
			t.Errorf("bytecodeDottedBitwiseOperator(%q) = %v, want %v", test.op, got, test.want)
		}
	}
}

func TestBytecodeLoweringMarksOnlyDottedBitwiseRawCandidates(t *testing.T) {
	for _, test := range []struct {
		operator string
		want     bool
	}{
		{operator: ".^", want: true},
		{operator: ".>>", want: true},
		{operator: "^", want: false},
		{operator: "+", want: false},
	} {
		program, err := NewBytecode().lowerExpressionToBytecode(ast.Bin(test.operator, ast.Int(8), ast.Int(1)))
		if err != nil {
			t.Fatalf("lower %q: %v", test.operator, err)
		}
		var binary *bytecodeInstruction
		for i := range program.instructions {
			if program.instructions[i].operator == test.operator {
				binary = &program.instructions[i]
				break
			}
		}
		if binary == nil {
			t.Fatalf("lower %q: binary instruction not found", test.operator)
		}
		if binary.bitwiseRawCandidate != test.want {
			t.Errorf("lower %q: bitwiseRawCandidate = %v, want %v", test.operator, binary.bitwiseRawCandidate, test.want)
		}
	}
}

func TestBytecodeVM_BinaryIntAddReturnsRawSmallU64Result(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	got, handled, err := vm.execBinarySpecializedOpcode(
		&bytecodeInstruction{op: bytecodeOpBinaryIntAdd},
		bytecodeRawIntegerResultValue(runtime.IntegerU64, 40),
		bytecodeRawIntegerResultValue(runtime.IntegerU64, 2),
	)
	if err != nil {
		t.Fatalf("execBinarySpecializedOpcode failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected specialized u64 add to be handled")
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(got)
	if !ok || kind != runtime.IntegerU64 || raw != 42 {
		t.Fatalf("add result = %#v, want raw 42_u64", got)
	}
}

func TestBytecodeRawIntegerResultValueUsesDirectUnsignedCarriers(t *testing.T) {
	tests := []struct {
		name string
		kind runtime.IntegerType
		raw  int64
	}{
		{name: "u8", kind: runtime.IntegerU8, raw: 65},
		{name: "u16", kind: runtime.IntegerU16, raw: 1024},
		{name: "u32", kind: runtime.IntegerU32, raw: 1 << 20},
		{name: "u64", kind: runtime.IntegerU64, raw: 1 << 20},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bytecodeRawIntegerResultValue(tc.kind, tc.raw)
			if _, ok := got.(bytecodeRawIntegerValue); ok {
				t.Fatalf("result type = %T, want direct unsigned carrier", got)
			}
			kind, raw, ok := bytecodeRawIntegerValueInfo(got)
			if !ok || kind != tc.kind || raw != tc.raw {
				t.Fatalf("raw info = (%q, %d, %v), want (%s, %d, true)", kind, raw, ok, tc.kind, tc.raw)
			}
		})
	}
}

func TestBytecodeRawIntegerResultValueUsesRawI64CarrierForLargeI64(t *testing.T) {
	value := int64(bytecodeSmallIntBoxMax + 100000)
	got := bytecodeRawIntegerResultValue(runtime.IntegerI64, value)

	rawVal, ok := got.(bytecodeRawI64ResultValue)
	if !ok {
		t.Fatalf("result type = %T, want bytecodeRawI64ResultValue", got)
	}
	if int64(rawVal) != value {
		t.Fatalf("result = %#v, want raw i64 %d", rawVal, value)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(got)
	if !ok || kind != runtime.IntegerI64 || raw != value {
		t.Fatalf("raw info = (%q, %d, %v), want (i64, %d, true)", kind, raw, ok, value)
	}
}

func TestBytecodeVM_CastOpcodeFastUsesRawI64StackCellForSmallIntSource(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{bytecodeRawIntegerResultValue(runtime.IntegerU8, 65)}

	if err := vm.execCastOpcode(&bytecodeInstruction{
		op:   bytecodeOpCast,
		node: ast.NewTypeCastExpression(ast.Int(65), ast.Ty("i64")),
	}); err != nil {
		t.Fatalf("execCastOpcode failed: %v", err)
	}

	cell, ok := vm.stack[0].(*bytecodeRawI64SlotCell)
	if !ok || cell == nil || cell.Val != 65 {
		t.Fatalf("cast result = %#v, want reusable raw i64 stack cell 65", vm.stack[0])
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerI64 || raw != 65 {
		t.Fatalf("raw info = (%q, %d, %v), want (i64, 65, true)", kind, raw, ok)
	}
}

func TestBytecodeVM_CastOpcodeFastUsesReusableRawU64StackCellForLargeRawU32(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	value := int64(bytecodeSmallIntBoxMax + 70000)
	vm.stack = []runtime.Value{bytecodeRawIntegerResultValue(runtime.IntegerU32, value)}

	if err := vm.execCastOpcode(&bytecodeInstruction{
		op:   bytecodeOpCast,
		node: ast.NewTypeCastExpression(ast.Int(value), ast.Ty("u64")),
	}); err != nil {
		t.Fatalf("execCastOpcode failed: %v", err)
	}

	cell, ok := vm.stack[0].(*bytecodeRawIntegerSlotCell)
	if !ok || cell == nil || cell.TypeSuffix != runtime.IntegerU64 || cell.Raw != value {
		t.Fatalf("cast result = %#v, want reusable raw u64 stack cell %d", vm.stack[0], value)
	}
}

func TestBytecodeVM_CastOpcodeFastUsesReusableRawU8StackCellForSmallIntSource(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{bytecodeRawI32SlotCachedValue(321)}

	if err := vm.execCastOpcode(&bytecodeInstruction{
		op:   bytecodeOpCast,
		node: ast.NewTypeCastExpression(ast.Int(321), ast.Ty("u8")),
	}); err != nil {
		t.Fatalf("execCastOpcode failed: %v", err)
	}

	cell, ok := vm.stack[0].(*bytecodeRawIntegerSlotCell)
	if !ok || cell == nil || cell.TypeSuffix != runtime.IntegerU8 || cell.Raw != 65 {
		t.Fatalf("cast result = %#v, want reusable raw u8 stack cell 65", vm.stack[0])
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerU8 || raw != 65 {
		t.Fatalf("raw info = (%q, %d, %v), want (u8, 65, true)", kind, raw, ok)
	}
}

func TestBytecodeVM_BinaryIntAddUsesRawI64StackCellForResult(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{
		&bytecodeRawI64SlotCell{Val: 40},
		&bytecodeRawI64SlotCell{Val: 2},
	}

	handled, err := vm.execBinary(&bytecodeInstruction{op: bytecodeOpBinaryIntAdd}, nil)
	if err != nil {
		t.Fatalf("execBinary failed: %v", err)
	}
	if handled {
		t.Fatalf("expected execBinary to complete inline without yielding handled=true")
	}
	cell, ok := vm.stack[0].(*bytecodeRawI64SlotCell)
	if !ok || cell == nil || cell.Val != 42 {
		t.Fatalf("binary result = %#v, want reusable raw i64 stack cell 42", vm.stack[0])
	}
}

func TestBytecodeVM_BinaryIntAddUsesReusableRawU64StackCellBeyondSmallCache(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	left := int64(bytecodeSmallIntBoxMax + 70000)
	vm.stack = []runtime.Value{
		bytecodeRawIntegerResultValue(runtime.IntegerU64, left),
		bytecodeRawIntegerResultValue(runtime.IntegerU64, 2),
	}

	handled, err := vm.execBinary(&bytecodeInstruction{op: bytecodeOpBinaryIntAdd}, nil)
	if err != nil {
		t.Fatalf("execBinary failed: %v", err)
	}
	if handled {
		t.Fatalf("expected execBinary to complete inline without yielding handled=true")
	}
	cell, ok := vm.stack[0].(*bytecodeRawIntegerSlotCell)
	if !ok || cell == nil || cell.TypeSuffix != runtime.IntegerU64 || cell.Raw != left+2 {
		t.Fatalf("binary result = %#v, want reusable raw u64 stack cell %d", vm.stack[0], left+2)
	}
}

func TestBytecodeVM_BinaryIntSubUsesRawI32CarrierForResult(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{
		runtime.NewSmallInt(40, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}

	handled, err := vm.execBinary(&bytecodeInstruction{op: bytecodeOpBinaryIntSub}, nil)
	if err != nil {
		t.Fatalf("execBinary failed: %v", err)
	}
	if handled {
		t.Fatalf("expected execBinary to complete inline without yielding handled=true")
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerI32 || raw != 38 {
		t.Fatalf("binary result = %#v, want raw 38_i32", vm.stack[0])
	}
	if _, boxed := vm.stack[0].(runtime.IntegerValue); boxed {
		t.Fatalf("binary result should stay raw inside the VM, got boxed %#v", vm.stack[0])
	}
}

func TestBytecodeVM_BinaryIntSubUsesReusableRawI32StackCellBeyondStaticCache(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{
		runtime.NewSmallInt(int64(bytecodeRawI32SlotCacheMax+100), runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}

	handled, err := vm.execBinary(&bytecodeInstruction{op: bytecodeOpBinaryIntSub}, nil)
	if err != nil {
		t.Fatalf("execBinary failed: %v", err)
	}
	if handled {
		t.Fatalf("expected execBinary to complete inline without yielding handled=true")
	}
	cell, ok := vm.stack[0].(*bytecodeRawI32StackCell)
	if !ok || cell == nil || cell.Val != int32(bytecodeRawI32SlotCacheMax+98) {
		t.Fatalf("binary result = %#v, want reusable raw i32 stack cell", vm.stack[0])
	}
}

func TestBytecodeVM_BinaryIntMulSlotConstUsesReusableRawI32StackCellBeyondStaticCache(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	start := int32(bytecodeRawI32SlotCacheMax/2 + 100)
	vm.slots = []runtime.Value{runtime.NewSmallInt(int64(start), runtime.IntegerI32)}

	handled, err := vm.execBinary(&bytecodeInstruction{
		op:              bytecodeOpBinaryIntMulSlotConst,
		target:          0,
		operator:        "*",
		intImmediate:    runtime.NewSmallInt(2, runtime.IntegerI32),
		hasIntImmediate: true,
	}, nil)
	if err != nil {
		t.Fatalf("execBinary failed: %v", err)
	}
	if handled {
		t.Fatalf("expected execBinary to complete inline without yielding handled=true")
	}
	cell, ok := vm.stack[0].(*bytecodeRawI32StackCell)
	if !ok || cell == nil || cell.Val != start*2 {
		t.Fatalf("binary slot-const multiply result = %#v, want reusable raw i32 stack cell", vm.stack[0])
	}
}

func TestBytecodeVM_BinaryIntMulSlotConstUsesI32RegisterSlotFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	start := int32(bytecodeRawI32SlotCacheMax/2 + 100)
	vm.slots = make([]runtime.Value, 1)
	vm.i32Registers = []int32{start}
	vm.i32RegisterValid = []bool{true}

	handled, err := vm.execBinary(&bytecodeInstruction{
		op:              bytecodeOpBinaryIntMulSlotConst,
		target:          0,
		operator:        "*",
		intImmediate:    runtime.NewSmallInt(2, runtime.IntegerI32),
		hasIntImmediate: true,
	}, nil)
	if err != nil {
		t.Fatalf("execBinary failed: %v", err)
	}
	if handled {
		t.Fatalf("expected execBinary to complete inline without yielding handled=true")
	}
	cell, ok := vm.stack[0].(*bytecodeRawI32StackCell)
	if !ok || cell == nil || cell.Val != start*2 {
		t.Fatalf("binary slot-const multiply register result = %#v, want reusable raw i32 stack cell", vm.stack[0])
	}
}

func TestBytecodeVM_ReusableRawI32StackCellStaysRawForDivMod(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{
		runtime.NewSmallInt(int64(bytecodeRawI32SlotCacheMax+100), runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}

	handled, err := vm.execBinary(&bytecodeInstruction{op: bytecodeOpBinaryIntSub}, nil)
	if err != nil {
		t.Fatalf("execBinary sub failed: %v", err)
	}
	if handled {
		t.Fatalf("expected execBinary sub to complete inline without yielding handled=true")
	}
	cell, ok := vm.stack[0].(*bytecodeRawI32StackCell)
	if !ok || cell == nil {
		t.Fatalf("sub result = %#v, want reusable raw i32 stack cell", vm.stack[0])
	}

	vm.stack = append(vm.stack, runtime.NewSmallInt(2, runtime.IntegerI32))
	handled, err = vm.execBinary(&bytecodeInstruction{operator: "//"}, nil)
	if err != nil {
		t.Fatalf("execBinary divmod failed: %v", err)
	}
	if handled {
		t.Fatalf("expected execBinary divmod to complete inline without yielding handled=true")
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerI32 || raw != int64((bytecodeRawI32SlotCacheMax+98)/2) {
		t.Fatalf("divmod result = %#v, want raw %d_i32", vm.stack[0], (bytecodeRawI32SlotCacheMax+98)/2)
	}
	if _, boxed := vm.stack[0].(runtime.IntegerValue); boxed {
		t.Fatalf("divmod result should stay raw inside the VM, got boxed %#v", vm.stack[0])
	}
}

func TestApplyBinaryOperatorFast_UsesRawIntegerCarriers(t *testing.T) {
	add, handled, err := ApplyBinaryOperatorFast(
		"+",
		bytecodeRawI32SlotCachedValue(40),
		bytecodeRawI32SlotCachedValue(2),
	)
	if err != nil {
		t.Fatalf("raw add fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected raw add fast path to handle operands")
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(add)
	if !ok || kind != runtime.IntegerI32 || raw != 42 {
		t.Fatalf("raw add result = %#v, want raw 42_i32", add)
	}

	div, handled, err := ApplyBinaryOperatorFast(
		"//",
		bytecodeRawI32SlotCachedValue(40),
		bytecodeRawI32SlotCachedValue(6),
	)
	if err != nil {
		t.Fatalf("raw div fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected raw div fast path to handle operands")
	}
	kind, raw, ok = bytecodeRawIntegerValueInfo(div)
	if !ok || kind != runtime.IntegerI32 || raw != 6 {
		t.Fatalf("raw div result = %#v, want raw 6_i32", div)
	}

	mod, handled, err := ApplyBinaryOperatorFast(
		"%",
		bytecodeRawI32SlotCachedValue(40),
		bytecodeRawI32SlotCachedValue(6),
	)
	if err != nil {
		t.Fatalf("raw mod fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected raw mod fast path to handle operands")
	}
	kind, raw, ok = bytecodeRawIntegerValueInfo(mod)
	if !ok || kind != runtime.IntegerI32 || raw != 4 {
		t.Fatalf("raw mod result = %#v, want raw 4_i32", mod)
	}

	cmp, handled, err := ApplyBinaryOperatorFast(
		"<=",
		bytecodeRawI32SlotCachedValue(40),
		bytecodeRawI32SlotCachedValue(42),
	)
	if err != nil {
		t.Fatalf("raw compare fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected raw compare fast path to handle operands")
	}
	boolVal, ok := cmp.(runtime.BoolValue)
	if !ok || !boolVal.Val {
		t.Fatalf("raw compare result = %#v, want true", cmp)
	}
}

func TestApplyBinaryOperatorFast_SmallIntegerShiftSemantics(t *testing.T) {
	unsignedShift, handled, err := ApplyBinaryOperatorFast(
		">>",
		bytecodeRawIntegerResultValue(runtime.IntegerU32, 8),
		bytecodeRawIntegerResultValue(runtime.IntegerU32, 1),
	)
	if err != nil {
		t.Fatalf("raw unsigned shift failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected raw unsigned shift fast path to handle operands")
	}
	assertIntValue(t, bytecodeMaterializeRawIntegerValue(unsignedShift), runtime.IntegerU32, 4)

	signedShift, handled, err := ApplyBinaryOperatorFast(
		">>",
		runtime.NewSmallInt(-4, runtime.IntegerI8),
		runtime.NewSmallInt(1, runtime.IntegerI8),
	)
	if err != nil {
		t.Fatalf("raw signed shift failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected raw signed shift fast path to handle operands")
	}
	assertIntValue(t, bytecodeMaterializeRawIntegerValue(signedShift), runtime.IntegerI8, -2)

	signedLeftShift, handled, err := ApplyBinaryOperatorFast(
		"<<",
		runtime.NewSmallInt(-1, runtime.IntegerI8),
		runtime.NewSmallInt(1, runtime.IntegerI8),
	)
	if err != nil {
		t.Fatalf("raw signed left shift failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected raw signed left shift fast path to handle operands")
	}
	assertIntValue(t, bytecodeMaterializeRawIntegerValue(signedLeftShift), runtime.IntegerI8, -2)

	_, handled, err = ApplyBinaryOperatorFast(
		"<<",
		bytecodeRawIntegerResultValue(runtime.IntegerU8, 128),
		bytecodeRawIntegerResultValue(runtime.IntegerU8, 1),
	)
	if err == nil {
		t.Fatalf("expected overflowing unsigned shift to fail")
	}
	if !handled {
		t.Fatalf("expected overflowing unsigned shift to stay in the fast path")
	}
	if err.Error() != "integer overflow" {
		t.Fatalf("unexpected overflow error: %v", err)
	}
}

func TestApplyBinaryOperator_MaterializesNotNeededForRawIntegerArithmetic(t *testing.T) {
	interp := NewBytecode()
	got, err := applyBinaryOperator(
		interp,
		"+",
		bytecodeRawIntegerResultValue(runtime.IntegerU32, 40),
		bytecodeRawIntegerResultValue(runtime.IntegerU32, 2),
	)
	if err != nil {
		t.Fatalf("applyBinaryOperator raw add failed: %v", err)
	}
	assertIntValue(t, got, runtime.IntegerU32, 42)
}

func TestInterpreterMatchPatternFastMaterializesRawUnsignedInteger(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()

	matchEnv, matched, handled := interp.matchPatternFast(
		ast.TypedP(ast.ID("byte"), ast.Ty("u8")),
		bytecodeRawIntegerResultValue(runtime.IntegerU8, 65),
		env,
	)
	if !handled {
		t.Fatalf("expected typed primitive match to use fast path")
	}
	if !matched {
		t.Fatalf("expected raw u8 subject to match typed u8 pattern")
	}

	got, err := matchEnv.Get("byte")
	if err != nil {
		t.Fatalf("lookup typed binding: %v", err)
	}
	assertIntValue(t, got, runtime.IntegerU8, 65)
}

func TestInterpreterMatchesTypeMaterializesRawUnsignedInteger(t *testing.T) {
	interp := NewBytecode()
	value := bytecodeRawIntegerResultValue(runtime.IntegerU8, 65)

	if !interp.matchesType(ast.Ty("u8"), value) {
		t.Fatalf("expected raw u8 to match u8")
	}
	if !interp.matchesType(ast.Ty("i32"), value) {
		t.Fatalf("expected raw u8 to match i32 via range coercion")
	}
	if !interp.matchesType(ast.Ty("f64"), value) {
		t.Fatalf("expected raw u8 to match f64 via numeric coercion")
	}
}

func TestInterpreterCoerceValueToTypeMaterializesRawUnsignedInteger(t *testing.T) {
	interp := NewBytecode()

	got, err := interp.coerceValueToType(ast.Ty("i32"), bytecodeRawIntegerResultValue(runtime.IntegerU8, 65))
	if err != nil {
		t.Fatalf("coerceValueToType failed: %v", err)
	}
	assertIntValue(t, got, runtime.IntegerI32, 65)
}

func TestInterpreterTypeExpressionFromValueRecognizesRawNumericCarriers(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	if got := typeExpressionToString(interp.typeExpressionFromValue(vm.stackRawI32Value(0, int32(bytecodeRawI32SlotCacheMax+9)))); got != "i32" {
		t.Fatalf("typeExpressionFromValue(raw i32) = %q, want i32", got)
	}
	if got := typeExpressionToString(interp.typeExpressionFromValue(bytecodeRawFloatSlotValue(3.5, runtime.FloatF32))); got != "f32" {
		t.Fatalf("typeExpressionFromValue(raw f32) = %q, want f32", got)
	}
}

func TestInterpreterEnsureArrayStateMaterializesRawUnsignedIntegerElements(t *testing.T) {
	interp := NewBytecode()
	arr := &runtime.ArrayValue{
		Elements: []runtime.Value{
			bytecodeRawIntegerResultValue(runtime.IntegerU8, 65),
			bytecodeRawIntegerResultValue(runtime.IntegerU8, 66),
		},
	}

	state, err := interp.ensureArrayState(arr, len(arr.Elements))
	if err != nil {
		t.Fatalf("ensureArrayState failed: %v", err)
	}
	if state == nil {
		t.Fatalf("expected array state")
	}
	assertIntValue(t, state.Values[0], runtime.IntegerU8, 65)
	assertIntValue(t, state.Values[1], runtime.IntegerU8, 66)
}

func TestInterpreterMatchesTypeUsesMonoU32MetadataWithoutMaterializingArrayValues(t *testing.T) {
	interp := NewBytecode()
	handle := runtime.ArrayStoreMonoNewWithCapacityU32(4)
	for idx, value := range []uint32{1, 2, 3, 4} {
		if err := runtime.ArrayStoreMonoWriteU32(handle, idx, value); err != nil {
			t.Fatalf("ArrayStoreMonoWriteU32(%d): %v", idx, err)
		}
	}

	arr, err := interp.arrayValueFromHandle(handle, 0, 4)
	if err != nil {
		t.Fatalf("arrayValueFromHandle: %v", err)
	}
	if arr.State != nil || len(arr.Elements) != 0 {
		t.Fatalf("expected mono array view without materialized state, got state=%#v len=%d", arr.State, len(arr.Elements))
	}
	if !interp.matchesType(ast.Gen(ast.Ty("Array"), ast.Ty("u64")), arr) {
		t.Fatalf("expected mono u32 array to match Array<u64> without value materialization")
	}
	if arr.State != nil || len(arr.Elements) != 0 {
		t.Fatalf("matchesType should not materialize mono array state, got state=%#v len=%d", arr.State, len(arr.Elements))
	}
}

func TestInterpreterMatchesTypeEmptyMonoArrayStillMatchesUnknownElementTypeWithoutMaterializing(t *testing.T) {
	interp := NewBytecode()
	handle := runtime.ArrayStoreMonoNewWithCapacityU32(0)

	arr, err := interp.arrayValueFromHandle(handle, 0, 0)
	if err != nil {
		t.Fatalf("arrayValueFromHandle: %v", err)
	}
	if !interp.matchesType(ast.Gen(ast.Ty("Array"), ast.Ty("DefinitelyUnknownType")), arr) {
		t.Fatalf("expected empty mono array to match unknown Array<T> element type")
	}
	if arr.State != nil || len(arr.Elements) != 0 {
		t.Fatalf("empty mono array match should stay unmaterialized, got state=%#v len=%d", arr.State, len(arr.Elements))
	}
}
