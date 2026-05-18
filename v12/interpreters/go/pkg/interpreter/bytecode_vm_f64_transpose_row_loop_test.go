package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsF64TransposeRowLoopPlan(t *testing.T) {
	arrayF64 := ast.Gen(ast.Ty("Array"), ast.Ty("f64"))
	arrayArrayF64 := ast.Gen(ast.Ty("Array"), arrayF64)
	arg := ast.Prop(ast.CallExpr(
		ast.Member(
			ast.Prop(ast.CallExpr(ast.Member(ast.ID("b"), "get"), ast.ID("j"))),
			"get",
		),
		ast.ID("i"),
	))
	def := ast.Fn(
		"transpose_row",
		[]*ast.FunctionParameter{
			ast.Param("ci", arrayF64),
			ast.Param("b", arrayArrayF64),
			ast.Param("i", ast.Ty("i32")),
			ast.Param("n", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("j"), ast.Int(0)),
			ast.Loop(
				ast.Iff(ast.Bin(">=", ast.ID("j"), ast.ID("n")), ast.Brk(nil, nil)),
				ast.CallExpr(ast.Member(ast.ID("ci"), "push"), arg),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("j"), ast.Bin("+", ast.ID("j"), ast.Int(1))),
			),
			ast.ID("ci"),
		},
		arrayF64,
		nil,
		nil,
		false,
		false,
	)

	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if len(program.f64TransposeRowLoops) != 1 {
		t.Fatalf("expected one f64 transpose-row loop plan, got %#v", program.f64TransposeRowLoops)
	}
	for ip, plan := range program.f64TransposeRowLoops {
		if ip < 0 || ip >= len(program.instructions) || program.instructions[ip].op != bytecodeOpLoopEnter {
			t.Fatalf("f64 transpose-row plan attached to non-loop-enter ip %d", ip)
		}
		if !plan.validForSlots(program.frameLayout.slotCount) || plan.successTarget <= ip || plan.resultPushIP <= ip {
			t.Fatalf("unexpected f64 transpose-row plan: ip=%d plan=%#v", ip, plan)
		}
		if got := program.instructions[plan.resultPushIP]; got.op != bytecodeOpCallMemberArraySlot || got.name != "push" || got.argCount != 1 {
			t.Fatalf("transpose-row result push ip points at %#v", got)
		}
	}
}

func TestBytecodeVM_F64TransposeRowLoopFastPathAppendsColumnWithAmortizedCapacity(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	dest := interp.newArrayValue([]runtime.Value{}, 0)
	outer := bytecodeF64TransposeOuterForTest(t, interp,
		monoF64ArrayValueForTest(t, 1, 10),
		monoF64ArrayValueForTest(t, 2, 20),
		monoF64ArrayValueForTest(t, 3, 30),
		monoF64ArrayValueForTest(t, 4, 40),
		monoF64ArrayValueForTest(t, 5, 50),
	)
	program := bytecodeF64TransposeRowLoopProgramForTest()
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(5), runtime.IntegerI32),
		dest,
		outer,
		runtime.NewBigIntValue(big.NewInt(1), runtime.IntegerI32),
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, outer)
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], dest, bytecodeMemberMethodFastPathArrayPush)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 transpose-row fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected f64 transpose-row fast path to complete")
	}
	if vm.ip != 2 {
		t.Fatalf("ip after transpose-row loop = %d, want 2", vm.ip)
	}
	assertIntValue(t, vm.slots[0], runtime.IntegerI32, 5)
	if len(vm.stack) != 1 || !isNilRuntimeValue(vm.stack[0]) {
		t.Fatalf("stack after transpose-row loop = %#v, want nil loop result", vm.stack)
	}
	values, mono, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(dest.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable: %v", err)
	}
	want := []float64{10, 20, 30, 40, 50}
	if !mono || len(values) != len(want) {
		t.Fatalf("transpose row values=%#v mono=%v, want %v", values, mono, want)
	}
	for idx, value := range want {
		if values[idx] != value {
			t.Fatalf("transpose row[%d]=%v, want %v (all=%#v)", idx, values[idx], value, values)
		}
	}
	capacity, err := runtime.ArrayStoreCapacity(dest.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreCapacity: %v", err)
	}
	if capacity != 8 {
		t.Fatalf("transpose-row capacity = %d, want amortized Array.new capacity 8", capacity)
	}
}

func TestBytecodeVM_F64TransposeRowLoopFallsThroughWithoutPartialMutation(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	dest := interp.newArrayValue([]runtime.Value{}, 0)
	outer := bytecodeF64TransposeOuterForTest(t, interp,
		monoF64ArrayValueForTest(t, 1, 10),
		monoF64ArrayValueForTest(t, 2),
	)
	program := bytecodeF64TransposeRowLoopProgramForTest()
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(2), runtime.IntegerI32),
		dest,
		outer,
		runtime.NewBigIntValue(big.NewInt(1), runtime.IntegerI32),
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, outer)
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], dest, bytecodeMemberMethodFastPathArrayPush)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 transpose-row guard miss failed: %v", err)
	}
	if handled {
		t.Fatalf("short transpose source row should fall through to ordinary bytecode")
	}
	if vm.ip != 1 {
		t.Fatalf("ip after transpose-row guard miss = %d, want fallback ip 1", vm.ip)
	}
	assertIntValue(t, vm.slots[0], runtime.IntegerI32, 0)
	size, err := runtime.ArrayStoreSize(dest.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreSize: %v", err)
	}
	if size != 0 {
		t.Fatalf("transpose row size after guard miss = %d, want 0", size)
	}
}

func TestBytecodeVM_F64TransposeRowLoopFallsThroughWhenDestinationAliasesInputRow(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	destAndRow := monoF64ArrayValueForTest(t, 1, 10)
	outer := bytecodeF64TransposeOuterForTest(t, interp, destAndRow)
	program := bytecodeF64TransposeRowLoopProgramForTest()
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(1), runtime.IntegerI32),
		destAndRow,
		outer,
		runtime.NewBigIntValue(big.NewInt(1), runtime.IntegerI32),
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, outer)
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], destAndRow, bytecodeMemberMethodFastPathArrayPush)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 transpose-row alias guard failed: %v", err)
	}
	if handled {
		t.Fatalf("aliased destination/source row should fall through to ordinary bytecode")
	}
	values, mono, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(destAndRow.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable: %v", err)
	}
	if !mono || len(values) != 2 || values[0] != 1 || values[1] != 10 {
		t.Fatalf("aliased row values after guard miss=%#v mono=%v, want original [1 10]", values, mono)
	}
}

func bytecodeF64TransposeOuterForTest(t *testing.T, interp *Interpreter, rows ...runtime.Value) *runtime.ArrayValue {
	t.Helper()
	outer := interp.newArrayValue(rows, len(rows))
	if _, err := interp.ensureArrayState(outer, 0); err != nil {
		t.Fatalf("ensure outer array state: %v", err)
	}
	return outer
}

func bytecodeF64TransposeRowLoopProgramForTest() *bytecodeProgram {
	return &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{
				op:           bytecodeOpLoopEnter,
				loopBreak:    2,
				loopContinue: 0,
			},
			{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1},
		},
		f64TransposeRowLoops: map[int]bytecodeF64TransposeRowLoopPlan{0: {
			indexSlot:     0,
			boundSlot:     1,
			receiverSlot:  2,
			outerSlot:     3,
			colIndexSlot:  4,
			successTarget: 2,
			resultPushIP:  1,
		}},
	}
}
