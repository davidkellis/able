package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_DirectArrayIndexFastPath(t *testing.T) {
	cases := []struct {
		name  string
		value runtime.Value
		want  int
	}{
		{name: "small_value", value: runtime.NewSmallInt(7, runtime.IntegerI32), want: 7},
		{name: "small_pointer", value: func() runtime.Value {
			v := runtime.NewSmallInt(11, runtime.IntegerI32)
			return &v
		}(), want: 11},
		{name: "boxed_big_value", value: runtime.NewBigIntValue(big.NewInt(19), runtime.IntegerI32), want: 19},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, handled, err := bytecodeDirectArrayIndex(tc.value)
			if err != nil {
				t.Fatalf("direct array index returned error: %v", err)
			}
			if !handled {
				t.Fatalf("expected direct array index to handle %T", tc.value)
			}
			if got != tc.want {
				t.Fatalf("unexpected direct array index: got=%d want=%d", got, tc.want)
			}
		})
	}
}

func TestBytecodeVM_DirectSmallArrayIndexFastPath(t *testing.T) {
	got, handled := bytecodeDirectSmallArrayIndex(runtime.NewSmallInt(7, runtime.IntegerI32))
	if !handled {
		t.Fatalf("expected small array index helper to handle small integer")
	}
	if got != 7 {
		t.Fatalf("small array index = %d, want 7", got)
	}

	big := runtime.NewBigIntValue(big.NewInt(11), runtime.IntegerI32)
	if _, handled := bytecodeDirectSmallArrayIndex(big); handled {
		t.Fatalf("small array index helper should not handle boxed big integer")
	}
}

func TestBytecodeVM_DirectArrayIndexSetSyncsSharedAliases(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.global)
	first := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, 1)
	_, err := interp.ensureArrayState(first, 0)
	if err != nil {
		t.Fatalf("ensure first array state: %v", err)
	}
	second, err := interp.arrayValueFromHandle(first.Handle, 0, 0)
	if err != nil {
		t.Fatalf("arrayValueFromHandle: %v", err)
	}
	if !first.TrackedAliases || !second.TrackedAliases {
		t.Fatalf("expected both aliases to be marked shared before direct set")
	}

	written := runtime.StringValue{Val: "x"}
	got, handled, err := vm.resolveDirectArrayIndexSet(first, runtime.NewSmallInt(0, runtime.IntegerI32), written, ast.AssignmentAssign, "", false)
	if err != nil {
		t.Fatalf("direct array index set returned error: %v", err)
	}
	if !handled {
		t.Fatalf("expected direct array index set to handle tracked array write")
	}
	if !valuesEqual(got, written) {
		t.Fatalf("unexpected direct array index set result: got=%#v want=%#v", got, written)
	}
	if observed, ok := second.Elements[0].(runtime.StringValue); !ok || observed.Val != "x" {
		t.Fatalf("expected shared alias to observe direct bytecode set, got %#v", second.Elements[0])
	}
}

func TestBytecodeVM_LoweringEmitsArrayIndexGetSlotOpcode(t *testing.T) {
	def := ast.Fn(
		"load",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("i", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Index(ast.ID("arr"), ast.ID("i")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	var sawSlotIndex bool
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpArrayIndexGetSlot {
			sawSlotIndex = true
			if instr.argCount != 0 || instr.loopBreak != 1 {
				t.Fatalf("array index slots = receiver %d index %d, want 0/1", instr.argCount, instr.loopBreak)
			}
		}
		if instr.op == bytecodeOpIndexGet {
			t.Fatalf("slot-shaped array index should avoid stack IndexGet opcode")
		}
	}
	if !sawSlotIndex {
		t.Fatalf("expected lowering to emit array index slot opcode")
	}
}

func TestBytecodeVM_LoweringEmitsArrayIndexSetSlotOpcode(t *testing.T) {
	def := ast.Fn(
		"store",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("i", ast.Ty("i32")),
			ast.Param("v", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.Index(ast.ID("arr"), ast.ID("i")), ast.ID("v")),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	sawSetSlot := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpArrayIndexSetSlot {
			sawSetSlot = true
		}
		if instr.op == bytecodeOpIndexSet {
			t.Fatalf("did not expect generic index set for slot-shaped array assignment")
		}
	}
	if !sawSetSlot {
		t.Fatalf("expected lowering to emit array index set slot opcode")
	}
}

func TestBytecodeVM_LoweringKeepsI32RegisterFrameForArrayIndexReadWrite(t *testing.T) {
	def := ast.Fn(
		"bump",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("i", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("idx"), ast.Ty("i32")), ast.ID("i")),
			ast.Assign(
				ast.TypedP(ast.ID("current"), ast.Ty("i32")),
				ast.NewTypeCastExpression(ast.Index(ast.ID("arr"), ast.ID("idx")), ast.Ty("i32")),
			),
			ast.AssignIndex(ast.ID("arr"), ast.ID("idx"), ast.ID("current")),
			ast.NewTypeCastExpression(ast.Index(ast.ID("arr"), ast.ID("idx")), ast.Ty("i32")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.frameLayout == nil || !program.frameLayout.i32RegisterFrame {
		t.Fatalf("expected array index read/write helper to keep i32 register frame")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpArrayIndexGetSlot) {
		t.Fatalf("expected index read helper to lower through array index slot opcode")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpArrayIndexSetSlot) {
		t.Fatalf("expected index write helper to lower through array index set slot opcode")
	}
}

func TestBytecodeVM_I32RegisterFrameArrayIndexReadWriteParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"bump",
			[]*ast.FunctionParameter{
				ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
				ast.Param("i", ast.Ty("i32")),
			},
			[]ast.Statement{
				ast.Assign(ast.TypedP(ast.ID("idx"), ast.Ty("i32")), ast.ID("i")),
				ast.Assign(
					ast.TypedP(ast.ID("current"), ast.Ty("i32")),
					ast.NewTypeCastExpression(ast.Index(ast.ID("arr"), ast.ID("idx")), ast.Ty("i32")),
				),
				ast.AssignIndex(ast.ID("arr"), ast.ID("idx"), ast.ID("current")),
				ast.NewTypeCastExpression(ast.Index(ast.ID("arr"), ast.ID("idx")), ast.Ty("i32")),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("bump", ast.Arr(ast.Int(4), ast.Int(5)), ast.Int(1)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode array index register-frame parity mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ArrayIndexGetSlotFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
		runtime.StringValue{Val: "one"},
	}, 0)
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  0,
		loopBreak: 1,
	}
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}

	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("array index slot opcode failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("array index slot opcode ip = %d, want 1", vm.ip)
	}
	if want := (runtime.StringValue{Val: "one"}); !valuesEqual(vm.stack[0], want) {
		t.Fatalf("array index slot opcode result = %#v, want %#v", vm.stack[0], want)
	}

	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(-1, runtime.IntegerI32),
	}
	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("negative array index slot opcode failed: %v", err)
	}
	if _, ok := vm.stack[0].(runtime.ErrorValue); !ok {
		t.Fatalf("negative array index slot result = %#v, want error value", vm.stack[0])
	}
}

func TestBytecodeVM_ArrayIndexGetSlotUsesI32RegisterIndex(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
		runtime.StringValue{Val: "one"},
	}, 0)
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{
		slotCount:        2,
		slotKinds:        []bytecodeCellKind{bytecodeCellKindValue, bytecodeCellKindI32},
		hasTypedSlots:    true,
		i32RegisterFrame: true,
	}}
	vm.slots = []runtime.Value{arr, nil}
	vm.activateI32RegisterFrame(program)
	if !vm.setI32RegisterRaw(1, 1) {
		t.Fatalf("expected register frame to accept i32 index")
	}

	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  0,
		loopBreak: 1,
	}
	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("array index slot register-index opcode failed: %v", err)
	}
	if want := (runtime.StringValue{Val: "one"}); !valuesEqual(vm.stack[0], want) {
		t.Fatalf("array index slot register-index result = %#v, want %#v", vm.stack[0], want)
	}
}

func TestBytecodeVM_ArrayIndexGetSlotMonoCharFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'l', 'e')
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  0,
		loopBreak: 1,
	}

	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("mono char array index slot opcode failed: %v", err)
	}
	if got, ok := vm.stack[0].(runtime.CharValue); !ok || got.Val != 'l' {
		t.Fatalf("mono char array index slot result = %#v, want char 'l'", vm.stack[0])
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono char array index slot should not materialize boxed state")
	}
}

func TestBytecodeVM_ArrayIndexGetSlotMonoU32UsesReusableRawStackCell(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	value := uint32(bytecodeSmallIntBoxMax + 70000)
	arr := monoU32ArrayValueForTest(t, value)
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(0, runtime.IntegerI32),
	}
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  0,
		loopBreak: 1,
	}

	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("mono u32 array index slot opcode failed: %v", err)
	}
	cell, ok := vm.stack[0].(*bytecodeRawIntegerSlotCell)
	if !ok || cell == nil || cell.TypeSuffix != runtime.IntegerU32 || cell.Raw != int64(value) {
		t.Fatalf("mono u32 array index slot result = %#v, want reusable raw u32 stack cell %d", vm.stack[0], value)
	}
	first := cell
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono u32 array index slot should not materialize boxed state")
	}

	vm.stack = vm.stack[:0]
	vm.ip = 0
	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("mono u32 array index slot second opcode failed: %v", err)
	}
	if vm.stack[0] != first {
		t.Fatalf("mono u32 array index slot result pointer = %#v, want reuse of %#v", vm.stack[0], first)
	}
}

func TestBytecodeVM_ArrayIndexGetSlotSkipsFollowingPropagationForMonoChar(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'l', 'e')
	vm.currentProgram = &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpArrayIndexGetSlot},
		{op: bytecodeOpPropagation},
	}}
	vm.ip = 0
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  0,
		loopBreak: 1,
	}

	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("mono char array index slot propagation skip failed: %v", err)
	}
	if vm.ip != 2 {
		t.Fatalf("ip after mono char array index slot propagation skip = %d, want 2", vm.ip)
	}
	if got, ok := vm.stack[0].(runtime.CharValue); !ok || got.Val != 'l' {
		t.Fatalf("mono char array index slot propagated result = %#v, want char 'l'", vm.stack[0])
	}
}

func TestBytecodeVM_ArrayIndexGetSlotTrackedNilFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
	}, 1)
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	state.Values[0] = nil
	arr.Elements = state.Values
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(0, runtime.IntegerI32),
	}
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  0,
		loopBreak: 1,
	}

	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("nil array index slot opcode failed: %v", err)
	}
	if _, ok := vm.stack[0].(runtime.ErrorValue); !ok {
		t.Fatalf("nil array index slot result = %#v, want error value", vm.stack[0])
	}
}

func TestBytecodeVM_ArrayIndexSetSlotFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}, 2)
	if _, err := interp.ensureArrayState(arr, 0); err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	written := runtime.NewSmallInt(9, runtime.IntegerI32)
	vm.slots = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}
	vm.stack = []runtime.Value{written}
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexSetSlot,
		argCount:  0,
		loopBreak: 1,
	}
	if err := vm.execArrayIndexSetSlot(instr); err != nil {
		t.Fatalf("array index set slot fast path returned error: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("expected ip to advance, got %d", vm.ip)
	}
	if len(vm.stack) != 1 || !valuesEqual(vm.stack[0], written) {
		t.Fatalf("expected assignment result on stack, got %#v", vm.stack)
	}
	if !valuesEqual(arr.Elements[1], written) {
		t.Fatalf("expected array element write, got %#v", arr.Elements[1])
	}
}

func TestBytecodeVM_ArrayIndexSetSlotUsesI32RegisterIndex(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}, 2)
	if _, err := interp.ensureArrayState(arr, 0); err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{
		slotCount:        2,
		slotKinds:        []bytecodeCellKind{bytecodeCellKindValue, bytecodeCellKindI32},
		hasTypedSlots:    true,
		i32RegisterFrame: true,
	}}
	written := runtime.NewSmallInt(9, runtime.IntegerI32)
	vm.slots = []runtime.Value{arr, nil}
	vm.activateI32RegisterFrame(program)
	if !vm.setI32RegisterRaw(1, 1) {
		t.Fatalf("expected register frame to accept i32 index")
	}
	vm.stack = []runtime.Value{written}

	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexSetSlot,
		argCount:  0,
		loopBreak: 1,
	}
	if err := vm.execArrayIndexSetSlot(instr); err != nil {
		t.Fatalf("array index set slot register-index opcode failed: %v", err)
	}
	if !valuesEqual(arr.Elements[1], written) {
		t.Fatalf("expected array element write via register index, got %#v", arr.Elements[1])
	}
}

func TestBytecodeVM_ArrayIndexSetSlotMonoCharFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'l', 'e')
	written := runtime.CharValue{Val: 'z'}
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}
	vm.stack = []runtime.Value{written}
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexSetSlot,
		argCount:  0,
		loopBreak: 1,
	}

	if err := vm.execArrayIndexSetSlot(instr); err != nil {
		t.Fatalf("mono char array index set slot opcode failed: %v", err)
	}
	if len(vm.stack) != 1 || !valuesEqual(vm.stack[0], written) {
		t.Fatalf("mono char array index set stack = %#v, want written value", vm.stack)
	}
	raw, ok, err := runtime.ArrayStoreMonoReadCharIfAvailable(arr.Handle, 1)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadCharIfAvailable after direct set: %v", err)
	}
	if !ok || raw != 'z' {
		t.Fatalf("mono char array index set stored = (%q, %v), want ('z', true)", raw, ok)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono char array index set should not materialize boxed state")
	}
}

func TestBytecodeVM_IndexGetSkipsFollowingPropagationForMonoChar(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'l', 'e')
	vm.currentProgram = &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpIndexGet},
		{op: bytecodeOpPropagation},
	}}
	vm.ip = 0
	vm.stack = []runtime.Value{
		arr,
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}

	if err := vm.execIndexGet(bytecodeInstruction{op: bytecodeOpIndexGet}); err != nil {
		t.Fatalf("generic index get propagation skip failed: %v", err)
	}
	if vm.ip != 2 {
		t.Fatalf("ip after generic index get propagation skip = %d, want 2", vm.ip)
	}
	if got, ok := vm.stack[0].(runtime.CharValue); !ok || got.Val != 'b' {
		t.Fatalf("generic index get propagated result = %#v, want char 'b'", vm.stack[0])
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("generic index get should not materialize boxed state")
	}
}

func TestBytecodeVM_CanonicalIndexGetMethodFastPathReadsMonoCharWithoutMaterializingState(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	if interp.canUseDirectArrayIndexGetFastPath() {
		t.Fatalf("expected stdlib bootstrap to install canonical Array Index impl")
	}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'l', 'e')

	got, err := vm.resolveIndexGet(arr, runtime.NewSmallInt(2, runtime.IntegerI32))
	if err != nil {
		t.Fatalf("resolveIndexGet canonical fast path failed: %v", err)
	}
	if charVal, ok := got.(runtime.CharValue); !ok || charVal.Val != 'l' {
		t.Fatalf("resolveIndexGet canonical fast path result = %#v, want char 'l'", got)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("resolveIndexGet canonical fast path should not materialize boxed state")
	}
}

func TestBytecodeVM_CanonicalIndexSetMethodFastPathWritesMonoCharWithoutMaterializingState(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	if interp.canUseDirectArrayIndexSetFastPath() {
		t.Fatalf("expected stdlib bootstrap to install canonical Array IndexMut impl")
	}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'l', 'e')
	written := runtime.CharValue{Val: 'z'}

	got, err := vm.resolveIndexSet(arr, runtime.NewSmallInt(1, runtime.IntegerI32), written, ast.AssignmentAssign, "", false)
	if err != nil {
		t.Fatalf("resolveIndexSet canonical fast path failed: %v", err)
	}
	if !valuesEqual(got, written) {
		t.Fatalf("resolveIndexSet canonical fast path result = %#v, want %#v", got, written)
	}
	raw, ok, err := runtime.ArrayStoreMonoReadCharIfAvailable(arr.Handle, 1)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadCharIfAvailable after canonical set: %v", err)
	}
	if !ok || raw != 'z' {
		t.Fatalf("canonical index set stored = (%q, %v), want ('z', true)", raw, ok)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("resolveIndexSet canonical fast path should not materialize boxed state")
	}
}

func TestBytecodeVM_CanonicalIndexMethodFastPathKindsResolveFromStdlibArrayImpl(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'l', 'e')

	getMethod, getFastPath, hasGetMethod, _, err := vm.resolveCachedIndexMethod(nil, 0, arr, "get", "Index")
	if err != nil {
		t.Fatalf("resolveCachedIndexMethod get failed: %v", err)
	}
	if !hasGetMethod || getMethod == nil {
		t.Fatalf("expected canonical Array Index.get method")
	}
	if getFastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("get fast path = %v, want canonical array get", getFastPath)
	}

	setMethod, setFastPath, hasSetMethod, _, err := vm.resolveCachedIndexMethod(nil, 0, arr, "set", "IndexMut")
	if err != nil {
		t.Fatalf("resolveCachedIndexMethod set failed: %v", err)
	}
	if !hasSetMethod || setMethod == nil {
		t.Fatalf("expected canonical Array IndexMut.set method")
	}
	if setFastPath != bytecodeIndexMethodFastPathCanonicalArraySet {
		t.Fatalf("set fast path = %v, want canonical array set", setFastPath)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("resolving canonical index methods should not materialize boxed state")
	}
}

func TestBytecodeVM_IndexMethodCacheIdentityTracksNominalArrayElementType(t *testing.T) {
	interp := NewBytecode()
	inner := monoCharArrayValueForTest(t, 'a', 'b')
	outer := interp.newArrayValue([]runtime.Value{inner}, 1)
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	receiverKind, elemType, typeKey, _, _, ok := vm.indexMethodCacheIdentityKey(outer)
	if !ok {
		t.Fatalf("expected nested array receiver to be cacheable")
	}
	if receiverKind != bytecodeMemberReceiverArray {
		t.Fatalf("receiver kind = %v, want array", receiverKind)
	}
	if elemType != bytecodeIndexTypeUnknown {
		t.Fatalf("element token = %v, want unknown for nominal nested arrays", elemType)
	}
	if typeKey != "Array<char>" {
		t.Fatalf("receiver type key = %q, want Array<char>", typeKey)
	}
}

func TestBytecodeVM_UnaliasedTrackedArrayWriteSyncFastPath(t *testing.T) {
	interp := NewBytecode()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, 1)
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	written := runtime.NewSmallInt(7, runtime.IntegerI32)
	state.Values[0] = written
	state.ElementTypeToken = bytecodeIndexTypeUnknown
	state.ElementTypeTokenKnown = false
	if !bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, 0, written) {
		t.Fatalf("expected unaliased tracked array write to use fast sync")
	}
	if !valuesEqual(arr.Elements[0], written) || arr.State != state {
		t.Fatalf("expected fast sync to refresh array view, elements=%#v state=%p want=%p", arr.Elements, arr.State, state)
	}
	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeI32 {
		t.Fatalf("expected fast sync to refresh element type token, known=%v token=%v", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}
	if !state.ValuesMaterialized {
		t.Fatalf("materialized tracked write should keep array state materialized")
	}

	arr.TrackedAliases = true
	if bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, 0, written) {
		t.Fatalf("expected aliased tracked array write to use the shared sync path")
	}
}

func TestBytecodeVM_UnaliasedTrackedArrayWriteMarksRawStateUnmaterialized(t *testing.T) {
	interp := NewBytecode()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, 1)
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	if !state.ValuesMaterialized {
		t.Fatalf("expected initial tracked state to be materialized")
	}

	written := bytecodeRawI32SlotCachedValue(11)
	state.Values[0] = written
	if !bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, 0, written) {
		t.Fatalf("expected unaliased tracked raw write to use fast sync")
	}
	if state.ValuesMaterialized {
		t.Fatalf("tracked raw write should mark array state as needing materialization")
	}
}

func TestBytecodeVM_IndexMethodCacheTracksArrayElementType(t *testing.T) {
	indexIface := ast.Iface(
		"Index",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"get",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("idx", ast.Ty("i32")),
				},
				ast.Ty("i32"),
				nil,
				nil,
				nil,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)

	getI32 := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("idx", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Int(11),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	getString := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Array"), ast.Ty("String"))),
			ast.Param("idx", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Int(22),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		indexIface,
		ast.Impl("Index", ast.Gen(ast.Ty("Array"), ast.Ty("i32")), []*ast.FunctionDefinition{getI32}, nil, nil, nil, nil, false),
		ast.Impl("Index", ast.Gen(ast.Ty("Array"), ast.Ty("String")), []*ast.FunctionDefinition{getString}, nil, nil, nil, nil, false),
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(1), ast.Int(2))),
		ast.Assign(ast.ID("first"), ast.Index(ast.ID("arr"), ast.Int(1))),
		ast.AssignOp(ast.AssignmentAssign, ast.Index(ast.ID("arr"), ast.Int(0)), ast.Str("x")),
		ast.Assign(ast.ID("second"), ast.Index(ast.ID("arr"), ast.Int(1))),
		ast.ID("second"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode index cache array element-type dispatch mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 22 {
		t.Fatalf("expected second index lookup to use Array String impl and return 22, got %#v", got)
	}
}

func TestBytecodeVM_IndexSetCompoundCacheInvalidatesWhenImplAppears(t *testing.T) {
	indexIface := ast.Iface(
		"Index",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"get",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("idx", ast.Ty("i32")),
				},
				ast.Ty("i32"),
				nil,
				nil,
				nil,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	indexMutIface := ast.Iface(
		"IndexMut",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"set",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("idx", ast.Ty("i32")),
					ast.Param("value", ast.Ty("i32")),
				},
				ast.Ty("i32"),
				nil,
				nil,
				nil,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)

	bump := ast.Fn(
		"bump",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("delta", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAdd, ast.Index(ast.ID("arr"), ast.Int(0)), ast.ID("delta")),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	getI32 := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("idx", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ID("marker"), ast.Bin("+", ast.ID("marker"), ast.Int(10))),
			ast.Int(7),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	setI32 := ast.Fn(
		"set",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("idx", ast.Ty("i32")),
			ast.Param("value", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ID("marker"), ast.Bin("+", ast.ID("marker"), ast.ID("value"))),
			ast.Int(0),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		indexIface,
		indexMutIface,
		bump,
		ast.Assign(ast.ID("marker"), ast.Int(0)),
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(1))),
		ast.Call("bump", ast.ID("arr"), ast.Int(2)),
		ast.Impl("Index", ast.Gen(ast.Ty("Array"), ast.Ty("i32")), []*ast.FunctionDefinition{getI32}, nil, nil, nil, nil, false),
		ast.Impl("IndexMut", ast.Gen(ast.Ty("Array"), ast.Ty("i32")), []*ast.FunctionDefinition{setI32}, nil, nil, nil, nil, false),
		ast.Call("bump", ast.ID("arr"), ast.Int(5)),
		ast.ID("marker"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode compound index cache invalidation mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 22 {
		t.Fatalf("expected marker 22 after impl-backed compound assignment, got %#v", got)
	}
}

func TestBytecodeVM_IndexGetFastPathInvalidatesWhenImplAppears(t *testing.T) {
	indexIface := ast.Iface(
		"Index",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"get",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("idx", ast.Ty("i32")),
				},
				ast.Ty("i32"),
				nil,
				nil,
				nil,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)

	read := ast.Fn(
		"read",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
		},
		[]ast.Statement{
			ast.Index(ast.ID("arr"), ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	getI32 := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("idx", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Int(99),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		indexIface,
		read,
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(7))),
		ast.Assign(ast.ID("before"), ast.Call("read", ast.ID("arr"))),
		ast.Impl("Index", ast.Gen(ast.Ty("Array"), ast.Ty("i32")), []*ast.FunctionDefinition{getI32}, nil, nil, nil, nil, false),
		ast.Assign(ast.ID("after"), ast.Call("read", ast.ID("arr"))),
		ast.Bin("+", ast.ID("before"), ast.ID("after")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode index get fast-path invalidation mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 106 {
		t.Fatalf("expected before+after marker 106 after impl-backed read, got %#v", got)
	}
}
