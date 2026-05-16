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
