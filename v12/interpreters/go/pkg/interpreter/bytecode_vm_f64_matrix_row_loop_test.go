package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsF64MatrixRowLoopPlan(t *testing.T) {
	arrayF64 := ast.Gen(ast.Ty("Array"), ast.Ty("f64"))
	arrayArrayF64 := ast.Gen(ast.Ty("Array"), arrayF64)
	def := ast.Fn(
		"row",
		[]*ast.FunctionParameter{
			ast.Param("di", arrayF64),
			ast.Param("ai", arrayF64),
			ast.Param("c", arrayArrayF64),
			ast.Param("n", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("j"), ast.Int(0)),
			ast.Loop(
				ast.Iff(ast.Bin(">=", ast.ID("j"), ast.ID("n")), ast.Brk(nil, nil)),
				ast.Assign(ast.ID("s"), ast.Flt(0)),
				ast.Assign(ast.ID("cj"), ast.Prop(ast.CallExpr(ast.Member(ast.ID("c"), "get"), ast.ID("j")))),
				ast.Assign(ast.ID("k"), ast.Int(0)),
				ast.Loop(
					ast.Iff(ast.Bin(">=", ast.ID("k"), ast.ID("n")), ast.Brk(nil, nil)),
					ast.AssignOp(ast.AssignmentAssign, ast.ID("s"), ast.Bin("+", ast.ID("s"), ast.Bin("*",
						ast.Prop(ast.CallExpr(ast.Member(ast.ID("ai"), "get"), ast.ID("k"))),
						ast.Prop(ast.CallExpr(ast.Member(ast.ID("cj"), "get"), ast.ID("k"))),
					))),
					ast.AssignOp(ast.AssignmentAssign, ast.ID("k"), ast.Bin("+", ast.ID("k"), ast.Int(1))),
				),
				ast.CallExpr(ast.Member(ast.ID("di"), "push"), ast.ID("s")),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("j"), ast.Bin("+", ast.ID("j"), ast.Int(1))),
			),
			ast.ID("di"),
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
	if len(program.f64MatrixRowLoops) != 1 {
		t.Fatalf("expected one f64 matrix-row loop plan, got %#v", program.f64MatrixRowLoops)
	}
	for ip, plan := range program.f64MatrixRowLoops {
		if ip < 0 || ip >= len(program.instructions) || program.instructions[ip].op != bytecodeOpLoopEnter {
			t.Fatalf("f64 matrix-row plan attached to non-loop-enter ip %d", ip)
		}
		if !plan.validForSlots(program.frameLayout.slotCount) || plan.successTarget <= ip || plan.resultPushIP <= ip {
			t.Fatalf("unexpected f64 matrix-row plan: ip=%d plan=%#v", ip, plan)
		}
		if got := program.instructions[plan.resultPushIP]; got.op != bytecodeOpCallMemberArraySlot || got.name != "push" || got.argCount != 1 {
			t.Fatalf("matrix-row result push ip points at %#v", got)
		}
	}
}

func TestBytecodeVM_F64MatrixRowLoopFastPathAppendsRow(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	dest := interp.newArrayValue([]runtime.Value{}, 2)
	left := monoF64ArrayValueForTest(t, 1, 2)
	row0 := monoF64ArrayValueForTest(t, 3, 4)
	row1 := monoF64ArrayValueForTest(t, 5, 6)
	outer := interp.newArrayValue([]runtime.Value{row0, row1}, 2)
	if _, err := interp.ensureArrayState(outer, 0); err != nil {
		t.Fatalf("ensure outer array state: %v", err)
	}
	program := bytecodeF64MatrixRowLoopProgramForTest()
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(2), runtime.IntegerI32),
		dest,
		left,
		outer,
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, outer)
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], dest, bytecodeMemberMethodFastPathArrayPush)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 matrix-row fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected f64 matrix-row fast path to complete")
	}
	if vm.ip != 2 {
		t.Fatalf("ip after matrix-row loop = %d, want 2", vm.ip)
	}
	assertIntValue(t, vm.slots[0], runtime.IntegerI32, 2)
	if len(vm.stack) != 1 || !isNilRuntimeValue(vm.stack[0]) {
		t.Fatalf("stack after matrix-row loop = %#v, want nil loop result", vm.stack)
	}
	values, mono, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(dest.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable: %v", err)
	}
	if !mono || len(values) != 2 || values[0] != 11 || values[1] != 17 {
		t.Fatalf("result row values=%#v mono=%v, want [11 17]", values, mono)
	}
}

func TestBytecodeVM_F64MatrixRowLoopBatchesFourRowsWithRemainder(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	dest := interp.newArrayValue([]runtime.Value{}, 5)
	left := monoF64ArrayValueForTest(t, 1, 2, 3, 4, 5)
	rows := []runtime.Value{
		monoF64ArrayValueForTest(t, 1, 0, 0, 0, 0),
		monoF64ArrayValueForTest(t, 0, 1, 0, 0, 0),
		monoF64ArrayValueForTest(t, 0, 0, 1, 0, 0),
		monoF64ArrayValueForTest(t, 0, 0, 0, 1, 0),
		monoF64ArrayValueForTest(t, 0, 0, 0, 0, 1),
	}
	outer := interp.newArrayValue(rows, len(rows))
	if _, err := interp.ensureArrayState(outer, 0); err != nil {
		t.Fatalf("ensure outer array state: %v", err)
	}
	program := bytecodeF64MatrixRowLoopProgramForTest()
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(5), runtime.IntegerI32),
		dest,
		left,
		outer,
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, outer)
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], dest, bytecodeMemberMethodFastPathArrayPush)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 matrix-row batch fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected f64 matrix-row batch fast path to complete")
	}
	values, mono, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(dest.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable: %v", err)
	}
	want := []float64{1, 2, 3, 4, 5}
	if !mono || len(values) != len(want) {
		t.Fatalf("batched result row values=%#v mono=%v, want %v", values, mono, want)
	}
	for idx, value := range want {
		if values[idx] != value {
			t.Fatalf("batched result row[%d]=%v, want %v (all=%#v)", idx, values[idx], value, values)
		}
	}
}

func TestBytecodeVM_F64MatrixRowLoopFallsThroughWithoutPartialMutation(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	dest := interp.newArrayValue([]runtime.Value{}, 2)
	left := monoF64ArrayValueForTest(t, 1, 2)
	row0 := monoF64ArrayValueForTest(t, 3, 4)
	row1 := monoF64ArrayValueForTest(t, 5)
	outer := interp.newArrayValue([]runtime.Value{row0, row1}, 2)
	if _, err := interp.ensureArrayState(outer, 0); err != nil {
		t.Fatalf("ensure outer array state: %v", err)
	}
	program := bytecodeF64MatrixRowLoopProgramForTest()
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(2), runtime.IntegerI32),
		dest,
		left,
		outer,
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, outer)
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], dest, bytecodeMemberMethodFastPathArrayPush)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 matrix-row guard miss failed: %v", err)
	}
	if handled {
		t.Fatalf("short row should fall through to ordinary bytecode")
	}
	if vm.ip != 1 {
		t.Fatalf("ip after matrix-row guard miss = %d, want fallback ip 1", vm.ip)
	}
	assertIntValue(t, vm.slots[0], runtime.IntegerI32, 0)
	if len(vm.stack) != 0 {
		t.Fatalf("stack after matrix-row guard miss = %#v, want empty", vm.stack)
	}
	size, err := runtime.ArrayStoreSize(dest.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreSize: %v", err)
	}
	if size != 0 {
		t.Fatalf("result row size after guard miss = %d, want 0", size)
	}
}

func TestBytecodeVM_F64MatrixRowLoopFallsThroughWhenDestinationAliasesInput(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	destAndLeft := monoF64ArrayValueForTest(t, 1, 2)
	row0 := monoF64ArrayValueForTest(t, 3, 4)
	outer := interp.newArrayValue([]runtime.Value{row0}, 1)
	if _, err := interp.ensureArrayState(outer, 0); err != nil {
		t.Fatalf("ensure outer array state: %v", err)
	}
	program := bytecodeF64MatrixRowLoopProgramForTest()
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(1), runtime.IntegerI32),
		destAndLeft,
		destAndLeft,
		outer,
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, outer)
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], destAndLeft, bytecodeMemberMethodFastPathArrayPush)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 matrix-row alias guard failed: %v", err)
	}
	if handled {
		t.Fatalf("aliased destination/input should fall through to ordinary bytecode")
	}
	values, mono, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(destAndLeft.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable: %v", err)
	}
	if !mono || len(values) != 2 || values[0] != 1 || values[1] != 2 {
		t.Fatalf("aliased input values after guard miss=%#v mono=%v, want original [1 2]", values, mono)
	}
}

func TestBytecodeVM_F64MatrixRowLoopRowCacheInvalidatesOnMonoF64Revision(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	left := monoF64ArrayValueForTest(t, 1, 1)
	row0 := monoF64ArrayValueForTest(t, 3, 4)
	row1 := monoF64ArrayValueForTest(t, 5, 6)
	outer := interp.newArrayValue([]runtime.Value{row0, row1}, 2)
	if _, err := interp.ensureArrayState(outer, 0); err != nil {
		t.Fatalf("ensure outer array state: %v", err)
	}
	program := bytecodeF64MatrixRowLoopProgramForTest()
	vm.currentProgram = program
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, outer)

	dest1 := interp.newArrayValue([]runtime.Value{}, 2)
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(2), runtime.IntegerI32),
		dest1,
		left,
		outer,
	}
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], dest1, bytecodeMemberMethodFastPathArrayPush)
	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 matrix-row first fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected first f64 matrix-row fast path to complete")
	}
	if len(vm.f64MatrixRowsCache) != 1 {
		t.Fatalf("expected one cached transposed row set, got %d", len(vm.f64MatrixRowsCache))
	}
	values, mono, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(dest1.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable dest1: %v", err)
	}
	if !mono || len(values) != 2 || values[0] != 7 || values[1] != 11 {
		t.Fatalf("first matrix row values=%#v mono=%v, want [7 11]", values, mono)
	}

	if err := runtime.ArrayStoreReserve(row0.Handle, 16); err != nil {
		t.Fatalf("ArrayStoreReserve row0: %v", err)
	}
	if err := runtime.ArrayStoreMonoWriteF64(row0.Handle, 0, 30); err != nil {
		t.Fatalf("ArrayStoreMonoWriteF64 row0: %v", err)
	}
	dest2 := interp.newArrayValue([]runtime.Value{}, 2)
	vm.ip = 0
	vm.stack = vm.stack[:0]
	vm.slots = []runtime.Value{
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(2), runtime.IntegerI32),
		dest2,
		left,
		outer,
	}
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], dest2, bytecodeMemberMethodFastPathArrayPush)
	programPtr = program
	instructions = program.instructions
	handled, err = vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 matrix-row second fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected second f64 matrix-row fast path to complete")
	}
	values, mono, err = runtime.ArrayStoreMonoF64ValuesIfAvailable(dest2.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable dest2: %v", err)
	}
	if !mono || len(values) != 2 || values[0] != 34 || values[1] != 11 {
		t.Fatalf("second matrix row values=%#v mono=%v, want [34 11]", values, mono)
	}
}

func bytecodeF64MatrixRowLoopProgramForTest() *bytecodeProgram {
	return &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{
				op:           bytecodeOpLoopEnter,
				loopBreak:    2,
				loopContinue: 0,
			},
			{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1},
		},
		f64MatrixRowLoops: map[int]bytecodeF64MatrixRowLoopPlan{0: {
			indexSlot:          0,
			boundSlot:          1,
			resultReceiverSlot: 2,
			leftReceiverSlot:   3,
			rightOuterSlot:     4,
			successTarget:      2,
			resultPushIP:       1,
		}},
	}
}
