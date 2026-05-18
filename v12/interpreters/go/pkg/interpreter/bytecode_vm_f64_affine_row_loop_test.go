package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsF64AffineRowLoopPlan(t *testing.T) {
	arrayF64 := ast.Gen(ast.Ty("Array"), ast.Ty("f64"))
	expr := ast.Bin("*",
		ast.ID("t"),
		ast.Bin("*",
			ast.NewTypeCastExpression(ast.Bin("-", ast.ID("i"), ast.ID("j")), ast.Ty("f64")),
			ast.NewTypeCastExpression(ast.Bin("+", ast.ID("i"), ast.ID("j")), ast.Ty("f64")),
		),
	)
	def := ast.Fn(
		"fill_row",
		[]*ast.FunctionParameter{
			ast.Param("row", arrayF64),
			ast.Param("t", ast.Ty("f64")),
			ast.Param("i", ast.Ty("i32")),
			ast.Param("n", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("j"), ast.Int(0)),
			ast.Loop(
				ast.Iff(ast.Bin(">=", ast.ID("j"), ast.ID("n")), ast.Brk(nil, nil)),
				ast.CallExpr(ast.Member(ast.ID("row"), "push"), expr),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("j"), ast.Bin("+", ast.ID("j"), ast.Int(1))),
			),
			ast.ID("row"),
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
	if len(program.f64AffineRowLoops) != 1 {
		t.Fatalf("expected one f64 affine-row loop plan, got %#v", program.f64AffineRowLoops)
	}
	for ip, plan := range program.f64AffineRowLoops {
		if ip < 0 || ip >= len(program.instructions) || program.instructions[ip].op != bytecodeOpLoopEnter {
			t.Fatalf("f64 affine-row plan attached to non-loop-enter ip %d", ip)
		}
		if !plan.validForSlots(program.frameLayout.slotCount) || plan.successTarget <= ip || plan.resultPushIP <= ip {
			t.Fatalf("unexpected f64 affine-row plan: ip=%d plan=%#v", ip, plan)
		}
		if got := program.instructions[plan.resultPushIP]; got.op != bytecodeOpCallMemberArraySlot || got.name != "push" || got.argCount != 1 {
			t.Fatalf("affine-row result push ip points at %#v", got)
		}
	}
}

func TestBytecodeVM_F64AffineRowLoopFastPathAppendsRowWithAmortizedCapacity(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	dest := interp.newArrayValue([]runtime.Value{}, 0)
	program := bytecodeF64AffineRowLoopProgramForTest()
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(5), runtime.IntegerI32),
		dest,
		runtime.FloatValue{Val: 0.5, TypeSuffix: runtime.FloatF64},
		runtime.NewBigIntValue(big.NewInt(2), runtime.IntegerI32),
	}
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], dest, bytecodeMemberMethodFastPathArrayPush)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 affine-row fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected f64 affine-row fast path to complete")
	}
	if vm.ip != 2 {
		t.Fatalf("ip after affine-row loop = %d, want 2", vm.ip)
	}
	assertIntValue(t, vm.slots[0], runtime.IntegerI32, 5)
	if len(vm.stack) != 1 || !isNilRuntimeValue(vm.stack[0]) {
		t.Fatalf("stack after affine-row loop = %#v, want nil loop result", vm.stack)
	}
	values, mono, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(dest.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable: %v", err)
	}
	want := []float64{2, 1.5, 0, -2.5, -6}
	if !mono || len(values) != len(want) {
		t.Fatalf("result row values=%#v mono=%v, want %v", values, mono, want)
	}
	for idx, value := range want {
		if values[idx] != value {
			t.Fatalf("result row[%d]=%v, want %v (all=%#v)", idx, values[idx], value, values)
		}
	}
	capacity, err := runtime.ArrayStoreCapacity(dest.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreCapacity: %v", err)
	}
	if capacity != 8 {
		t.Fatalf("affine-row capacity = %d, want amortized Array.new capacity 8", capacity)
	}
}

func TestBytecodeVM_F64AffineRowLoopFallsThroughWithoutPartialMutation(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	dest := interp.newArrayValue([]runtime.Value{}, 0)
	program := bytecodeF64AffineRowLoopProgramForTest()
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(5), runtime.IntegerI32),
		dest,
		runtime.NilValue{},
		runtime.NewBigIntValue(big.NewInt(2), runtime.IntegerI32),
	}
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], dest, bytecodeMemberMethodFastPathArrayPush)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 affine-row guard miss failed: %v", err)
	}
	if handled {
		t.Fatalf("non-f64 scale should fall through to ordinary bytecode")
	}
	if vm.ip != 1 {
		t.Fatalf("ip after affine-row guard miss = %d, want fallback ip 1", vm.ip)
	}
	assertIntValue(t, vm.slots[0], runtime.IntegerI32, 0)
	size, err := runtime.ArrayStoreSize(dest.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreSize: %v", err)
	}
	if size != 0 {
		t.Fatalf("result row size after guard miss = %d, want 0", size)
	}
}

func bytecodeF64AffineRowLoopProgramForTest() *bytecodeProgram {
	return &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{
				op:           bytecodeOpLoopEnter,
				loopBreak:    2,
				loopContinue: 0,
			},
			{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1},
		},
		f64AffineRowLoops: map[int]bytecodeF64AffineRowLoopPlan{0: {
			indexSlot:     0,
			boundSlot:     1,
			receiverSlot:  2,
			scaleSlot:     3,
			leftSlot:      4,
			successTarget: 2,
			resultPushIP:  1,
		}},
	}
}
