package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsFloatMulAddMulCompareConstJump(t *testing.T) {
	def := ast.Fn(
		"inside_unit_circle",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("f64")),
			ast.Param("y", ast.Ty("f64")),
		},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.Bin("+", ast.Bin("*", ast.ID("x"), ast.ID("x")), ast.Bin("*", ast.ID("y"), ast.ID("y"))), ast.Flt(1.0)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Int(0),
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
	var sawJump bool
	for ip, instr := range program.instructions {
		if instr.op != bytecodeOpJumpIfFloatMulAddMulCompareConstFalse {
			continue
		}
		sawJump = true
		if instr.operator != "<=" {
			t.Fatalf("float mul-add compare operator = %q, want <=", instr.operator)
		}
		plan, ok := program.floatMulAddMulJumps[ip]
		if !ok {
			t.Fatalf("missing float mul-add compare plan for ip %d", ip)
		}
		if plan.leftMulLeftSlot != 0 || plan.leftMulRightSlot != 0 || plan.rightMulLeftSlot != 1 || plan.rightMulRightSlot != 1 {
			t.Fatalf("float mul-add compare slots = %#v, want x*x + y*y", plan)
		}
		if plan.rightImmediate.TypeSuffix != runtime.FloatF64 || plan.rightImmediate.Val != 1.0 {
			t.Fatalf("float mul-add compare immediate = %#v, want f64 1.0", plan.rightImmediate)
		}
	}
	if !sawJump {
		t.Fatalf("expected lowering to emit float mul-add compare const jump")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfFalse) {
		t.Fatalf("expected fused float condition to skip generic jump-if-false bool materialization")
	}
}

func TestBytecodeVM_LoweringEmitsFloatMulAddMulCompareConstJumpWithTempStores(t *testing.T) {
	def := ast.Fn(
		"escape_or_diff",
		[]*ast.FunctionParameter{
			ast.Param("zr", ast.Ty("f64")),
			ast.Param("zi", ast.Ty("f64")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("zr2"), ast.Bin("*", ast.ID("zr"), ast.ID("zr"))),
			ast.Assign(ast.ID("zi2"), ast.Bin("*", ast.ID("zi"), ast.ID("zi"))),
			ast.IfExpr(
				ast.Bin(">", ast.Bin("+", ast.ID("zr2"), ast.ID("zi2")), ast.Flt(4.0)),
				ast.Block(ast.Ret(ast.Flt(999.0))),
			),
			ast.Bin("-", ast.ID("zr2"), ast.ID("zi2")),
		},
		ast.Ty("f64"),
		nil,
		nil,
		false,
		false,
	)

	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}

	var (
		sawJump          bool
		floatBinaryCount int
	)
	for ip, instr := range program.instructions {
		if instr.op == bytecodeOpStoreSlotFloatBinary {
			floatBinaryCount++
		}
		if instr.op != bytecodeOpJumpIfFloatMulAddMulCompareConstFalse {
			continue
		}
		sawJump = true
		plan, ok := program.floatMulAddMulJumps[ip]
		if !ok {
			t.Fatalf("missing float mul-add compare plan for ip %d", ip)
		}
		if !plan.storeProducts {
			t.Fatalf("expected temp-square fused jump to store products on false path")
		}
		if plan.leftTargetSlot < 0 || plan.rightTargetSlot < 0 || plan.leftTargetSlot == plan.rightTargetSlot {
			t.Fatalf("temp-square fused jump target slots = (%d, %d), want distinct non-negative slots", plan.leftTargetSlot, plan.rightTargetSlot)
		}
		if plan.leftTargetSlot == plan.leftMulLeftSlot || plan.rightTargetSlot == plan.rightMulLeftSlot {
			t.Fatalf("temp-square fused jump should write into temp slots, got %#v", plan)
		}
	}
	if !sawJump {
		t.Fatalf("expected lowering to emit float mul-add compare const jump for temp-square pattern")
	}
	if floatBinaryCount != 0 {
		t.Fatalf("expected temp-square fused jump to absorb square stores, saw %d float binary stores", floatBinaryCount)
	}
}

func TestBytecodeVM_LoweringKeepsFloatBinaryStoresWhenIfBodyUsesTempSquares(t *testing.T) {
	def := ast.Fn(
		"escape_or_square",
		[]*ast.FunctionParameter{
			ast.Param("zr", ast.Ty("f64")),
			ast.Param("zi", ast.Ty("f64")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("zr2"), ast.Bin("*", ast.ID("zr"), ast.ID("zr"))),
			ast.Assign(ast.ID("zi2"), ast.Bin("*", ast.ID("zi"), ast.ID("zi"))),
			ast.IfExpr(
				ast.Bin(">", ast.Bin("+", ast.ID("zr2"), ast.ID("zi2")), ast.Flt(4.0)),
				ast.Block(ast.Ret(ast.ID("zr2"))),
			),
			ast.Bin("-", ast.ID("zr2"), ast.ID("zi2")),
		},
		ast.Ty("f64"),
		nil,
		nil,
		false,
		false,
	)

	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatBinary) {
		t.Fatalf("expected lowering to keep explicit float binary stores when if body reads temp squares")
	}
	for _, plan := range program.floatMulAddMulJumps {
		if plan.storeProducts {
			t.Fatalf("if body reference should block temp-store fused jump plan: %#v", plan)
		}
	}
}

func TestBytecodeVM_FloatMulAddMulCompareConstJumpParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"inside_unit_circle",
			[]*ast.FunctionParameter{
				ast.Param("x", ast.Ty("f64")),
				ast.Param("y", ast.Ty("f64")),
			},
			[]ast.Statement{
				ast.IfExpr(
					ast.Bin("<=", ast.Bin("+", ast.Bin("*", ast.ID("x"), ast.ID("x")), ast.Bin("*", ast.ID("y"), ast.ID("y"))), ast.Flt(1.0)),
					ast.Block(ast.Ret(ast.Int(1))),
				),
				ast.Int(0),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.CallExpr(ast.ID("inside_unit_circle"), ast.Flt(0.5), ast.Flt(0.25)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode float mul-add compare mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_FloatMulAddMulCompareConstJumpTempStoreParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"escape_or_diff",
			[]*ast.FunctionParameter{
				ast.Param("zr", ast.Ty("f64")),
				ast.Param("zi", ast.Ty("f64")),
			},
			[]ast.Statement{
				ast.Assign(ast.ID("zr2"), ast.Bin("*", ast.ID("zr"), ast.ID("zr"))),
				ast.Assign(ast.ID("zi2"), ast.Bin("*", ast.ID("zi"), ast.ID("zi"))),
				ast.IfExpr(
					ast.Bin(">", ast.Bin("+", ast.ID("zr2"), ast.ID("zi2")), ast.Flt(4.0)),
					ast.Block(ast.Ret(ast.Flt(999.0))),
				),
				ast.Bin("-", ast.ID("zr2"), ast.ID("zi2")),
			},
			ast.Ty("f64"),
			nil,
			nil,
			false,
			false,
		),
		ast.Bin("+",
			ast.CallExpr(ast.ID("escape_or_diff"), ast.Flt(1.0), ast.Flt(1.0)),
			ast.CallExpr(ast.ID("escape_or_diff"), ast.Flt(3.0), ast.Flt(4.0)),
		),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode temp-store float mul-add compare mismatch: got=%#v want=%#v", got, want)
	}
	assertFloatValue(t, got, runtime.FloatF64, 999)
}

func TestBytecodeVM_FloatMulAddMulCompareConstJumpFastPathWithOwnedFloatSlots(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		floatMulAddMulJumps: map[int]bytecodeFloatMulAddMulCompareConstJumpPlan{
			0: {
				leftMulLeftSlot:   0,
				leftMulRightSlot:  0,
				rightMulLeftSlot:  1,
				rightMulRightSlot: 1,
				rightImmediate:    runtime.FloatValue{Val: 1.0, TypeSuffix: runtime.FloatF64},
			},
		},
	}
	vm.slots = make([]runtime.Value, 2)
	vm.storeOwnedFloatSlot(0, runtime.FloatValue{Val: 0.5, TypeSuffix: runtime.FloatF64})
	vm.storeOwnedFloatSlot(1, runtime.FloatValue{Val: 0.25, TypeSuffix: runtime.FloatF64})
	instr := &bytecodeInstruction{
		op:       bytecodeOpJumpIfFloatMulAddMulCompareConstFalse,
		target:   9,
		operator: "<=",
	}

	if err := vm.execJumpIfFloatMulAddMulCompareConstFalse(instr, program); err != nil {
		t.Fatalf("float mul-add compare fast path failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("truthy float mul-add compare should advance ip to 1, got %d", vm.ip)
	}

	vm.ip = 0
	vm.storeOwnedFloatSlot(0, runtime.FloatValue{Val: 0.9, TypeSuffix: runtime.FloatF64})
	vm.storeOwnedFloatSlot(1, runtime.FloatValue{Val: 0.9, TypeSuffix: runtime.FloatF64})
	if err := vm.execJumpIfFloatMulAddMulCompareConstFalse(instr, program); err != nil {
		t.Fatalf("float mul-add compare false jump failed: %v", err)
	}
	if vm.ip != 9 {
		t.Fatalf("false float mul-add compare should jump to 9, got %d", vm.ip)
	}
}

func TestBytecodeVM_FloatMulAddMulCompareConstJumpStoresTempSquaresOnFalsePath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		floatMulAddMulJumps: map[int]bytecodeFloatMulAddMulCompareConstJumpPlan{
			0: {
				leftMulLeftSlot:   0,
				leftMulRightSlot:  0,
				rightMulLeftSlot:  1,
				rightMulRightSlot: 1,
				rightImmediate:    runtime.FloatValue{Val: 4.0, TypeSuffix: runtime.FloatF64},
				storeProducts:     true,
				leftTargetSlot:    2,
				rightTargetSlot:   3,
			},
		},
	}
	vm.slots = make([]runtime.Value, 4)
	vm.storeOwnedFloatSlot(0, runtime.FloatValue{Val: 1.0, TypeSuffix: runtime.FloatF64})
	vm.storeOwnedFloatSlot(1, runtime.FloatValue{Val: 1.0, TypeSuffix: runtime.FloatF64})
	vm.slots[2] = runtime.NilValue{}
	vm.slots[3] = runtime.NilValue{}
	instr := &bytecodeInstruction{
		op:       bytecodeOpJumpIfFloatMulAddMulCompareConstFalse,
		target:   9,
		operator: ">",
	}

	if err := vm.execJumpIfFloatMulAddMulCompareConstFalse(instr, program); err != nil {
		t.Fatalf("temp-store false jump failed: %v", err)
	}
	if vm.ip != 9 {
		t.Fatalf("false float mul-add compare should jump to 9, got %d", vm.ip)
	}
	assertFloatValue(t, vm.slots[2], runtime.FloatF64, 1.0)
	assertFloatValue(t, vm.slots[3], runtime.FloatF64, 1.0)

	vm.ip = 0
	vm.storeOwnedFloatSlot(0, runtime.FloatValue{Val: 3.0, TypeSuffix: runtime.FloatF64})
	vm.storeOwnedFloatSlot(1, runtime.FloatValue{Val: 4.0, TypeSuffix: runtime.FloatF64})
	vm.storeOwnedFloatSlot(2, runtime.FloatValue{Val: -1.0, TypeSuffix: runtime.FloatF64})
	vm.storeOwnedFloatSlot(3, runtime.FloatValue{Val: -2.0, TypeSuffix: runtime.FloatF64})
	if err := vm.execJumpIfFloatMulAddMulCompareConstFalse(instr, program); err != nil {
		t.Fatalf("temp-store truthy compare failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("truthy float mul-add compare should advance ip to 1, got %d", vm.ip)
	}
	assertFloatValue(t, vm.slots[2], runtime.FloatF64, -1.0)
	assertFloatValue(t, vm.slots[3], runtime.FloatF64, -2.0)
}
