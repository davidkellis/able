package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsTryFloatUpdatePair(t *testing.T) {
	def := ast.Fn(
		"step",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("zr"), ast.Flt(1.0)),
			ast.Assign(ast.ID("zi"), ast.Flt(0.5)),
			ast.Assign(ast.ID("ci"), ast.Flt(0.25)),
			ast.Assign(ast.ID("cr"), ast.Flt(-0.5)),
			ast.Assign(ast.ID("zr2"), ast.Flt(1.0)),
			ast.Assign(ast.ID("zi2"), ast.Flt(0.25)),
			ast.Assign(ast.ID("iter"), ast.Int(0)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("zi"), ast.Bin("+", ast.Bin("*", ast.Bin("*", ast.Flt(2.0), ast.ID("zr")), ast.ID("zi")), ast.ID("ci"))),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("zr"), ast.Bin("+", ast.Bin("-", ast.ID("zr2"), ast.ID("zi2")), ast.ID("cr"))),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("iter"), ast.Bin("+", ast.ID("iter"), ast.Int(1))),
			ast.ID("zr"),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpTryFloatUpdatePair) {
		t.Fatalf("expected lowering to emit speculative float-update pair opcode")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatAddMulSlot) || !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatAddSub) {
		t.Fatalf("expected generic float update fallback instructions to remain present")
	}
}

func TestBytecodeVM_LoweringSkipsTryFloatUpdatePairWhenSecondUpdateReadsFirstTarget(t *testing.T) {
	def := ast.Fn(
		"step",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("zr"), ast.Flt(1.0)),
			ast.Assign(ast.ID("zi"), ast.Flt(0.5)),
			ast.Assign(ast.ID("ci"), ast.Flt(0.25)),
			ast.Assign(ast.ID("cr"), ast.Flt(-0.5)),
			ast.Assign(ast.ID("zi2"), ast.Flt(0.25)),
			ast.Assign(ast.ID("iter"), ast.Int(0)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("zi"), ast.Bin("+", ast.Bin("*", ast.Bin("*", ast.Flt(2.0), ast.ID("zr")), ast.ID("zi")), ast.ID("ci"))),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("zr"), ast.Bin("+", ast.Bin("-", ast.ID("zi"), ast.ID("zi2")), ast.ID("cr"))),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("iter"), ast.Bin("+", ast.ID("iter"), ast.Int(1))),
			ast.ID("zr"),
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
	if bytecodeProgramContainsOpcode(program, bytecodeOpTryFloatUpdatePair) {
		t.Fatalf("expected lowering to skip speculative pair when second update reads first target")
	}
}

func TestBytecodeVM_TryFloatUpdatePairFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		floatUpdatePairs: map[int]bytecodeFloatUpdatePairPlan{
			0: {
				firstTargetSlot:      1,
				firstBaseSlot:        2,
				firstMulSlot:         1,
				firstScaleSourceSlot: 0,
				firstScaleImmediate:  runtime.FloatValue{Val: 2.0, TypeSuffix: runtime.FloatF64},
				secondTargetSlot:     0,
				secondBaseSlot:       5,
				secondSubLeftSlot:    3,
				secondSubRightSlot:   4,
			},
		},
	}
	vm.slots = []runtime.Value{
		runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 0.25, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: -0.5, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 2.25, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 0.0625, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 0.75, TypeSuffix: runtime.FloatF64},
	}
	instr := &bytecodeInstruction{op: bytecodeOpTryFloatUpdatePair, target: 4}

	if err := vm.execTryFloatUpdatePair(program, instr); err != nil {
		t.Fatalf("speculative float update pair failed: %v", err)
	}
	if vm.ip != 4 {
		t.Fatalf("speculative float update pair ip = %d, want 4", vm.ip)
	}
	assertFloatValue(t, vm.slots[1], runtime.FloatF64, 0.25)
	assertFloatValue(t, vm.slots[0], runtime.FloatF64, 2.9375)
}

func TestBytecodeVM_TryFloatUpdatePairParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"step",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("zr"), ast.Flt(1.5)),
				ast.Assign(ast.ID("zi"), ast.Flt(0.25)),
				ast.Assign(ast.ID("ci"), ast.Flt(-0.5)),
				ast.Assign(ast.ID("cr"), ast.Flt(0.75)),
				ast.Assign(ast.ID("zr2"), ast.Bin("*", ast.ID("zr"), ast.ID("zr"))),
				ast.Assign(ast.ID("zi2"), ast.Bin("*", ast.ID("zi"), ast.ID("zi"))),
				ast.Assign(ast.ID("iter"), ast.Int(0)),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("zi"), ast.Bin("+", ast.Bin("*", ast.Bin("*", ast.Flt(2.0), ast.ID("zr")), ast.ID("zi")), ast.ID("ci"))),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("zr"), ast.Bin("+", ast.Bin("-", ast.ID("zr2"), ast.ID("zi2")), ast.ID("cr"))),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("iter"), ast.Bin("+", ast.ID("iter"), ast.Int(1))),
				ast.Bin("+", ast.ID("zr"), ast.ID("zi")),
			},
			ast.Ty("f64"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("step"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("speculative float update pair mismatch: got=%#v want=%#v", got, want)
	}
}
