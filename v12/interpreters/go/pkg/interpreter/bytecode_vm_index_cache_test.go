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

	arr.TrackedAliases = true
	if bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, 0, written) {
		t.Fatalf("expected aliased tracked array write to use the shared sync path")
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
