package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsBinaryFloatMulSlotConst(t *testing.T) {
	def := ast.Fn(
		"scale",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("zr"), ast.Flt(3)),
			ast.Assign(ast.ID("scaled"), ast.Bin("*", ast.Flt(2), ast.ID("zr"))),
			ast.ID("scaled"),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpBinaryFloatMulSlotConst) {
		t.Fatalf("expected lowering to emit float slot-const multiply opcode")
	}
}

func TestBytecodeVM_LoweringUsesBinaryFloatMulSlotConstInsideFloatAddMulUpdate(t *testing.T) {
	def := ast.Fn(
		"step",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("zi"), ast.Flt(0.5)),
			ast.Assign(ast.ID("zr"), ast.Flt(2)),
			ast.Assign(ast.ID("ci"), ast.Flt(1.25)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("zi"), ast.Bin("+", ast.Bin("*", ast.Bin("*", ast.Flt(2), ast.ID("zr")), ast.ID("zi")), ast.ID("ci"))),
			ast.ID("zi"),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpBinaryFloatMulSlotConst) {
		t.Fatalf("expected lowering to emit float slot-const multiply opcode inside float add-mul update")
	}
}

func TestBytecodeVM_LoweringCachesBinaryFloatMulSlotConstImmediate(t *testing.T) {
	expr := ast.Bin("*", ast.Flt(2), ast.ID("zr"))
	def := ast.Fn(
		"scale",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("zr"), ast.Flt(3)),
			ast.Assign(ast.ID("scaled"), expr),
			ast.ID("scaled"),
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
	found := false
	for _, instr := range program.instructions {
		if instr.op != bytecodeOpBinaryFloatMulSlotConst {
			continue
		}
		found = true
		if !instr.hasFloatImmediate {
			t.Fatalf("expected float slot-const multiply to cache raw float immediate")
		}
		if instr.floatImmediateKind != runtime.FloatF64 {
			t.Fatalf("cached float immediate kind = %v, want %v", instr.floatImmediateKind, runtime.FloatF64)
		}
		if instr.floatImmediateRaw != 2 {
			t.Fatalf("cached float immediate raw = %v, want 2", instr.floatImmediateRaw)
		}
	}
	if !found {
		t.Fatalf("expected lowering to emit float slot-const multiply opcode")
	}
}

func TestBytecodeVM_BinaryFloatMulSlotConstFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{
		runtime.FloatValue{Val: 3.5, TypeSuffix: runtime.FloatF64},
	}
	instr := &bytecodeInstruction{
		op:     bytecodeOpBinaryFloatMulSlotConst,
		target: 0,
		value:  runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
	}

	got, handled, err := vm.execBinaryFloatMulSlotConst(instr)
	if err != nil {
		t.Fatalf("float slot-const multiply failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected float slot-const multiply opcode to handle slot-backed float")
	}
	assertFloatValue(t, got, runtime.FloatF64, 7)
	if _, ok := got.(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("float slot-const multiply result = %#v, want raw f64 slot value", got)
	}
}

func TestBytecodeVM_BinaryFloatMulSlotConstFastPathUsesEmbeddedImmediate(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{
		runtime.FloatValue{Val: 3.5, TypeSuffix: runtime.FloatF64},
	}
	instr := &bytecodeInstruction{
		op:                 bytecodeOpBinaryFloatMulSlotConst,
		target:             0,
		floatImmediateRaw:  2,
		floatImmediateKind: runtime.FloatF64,
		hasFloatImmediate:  true,
	}

	got, handled, err := vm.execBinaryFloatMulSlotConst(instr)
	if err != nil {
		t.Fatalf("float slot-const multiply with embedded immediate failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected embedded float immediate fast path to handle slot-backed float")
	}
	assertFloatValue(t, got, runtime.FloatF64, 7)
	if _, ok := got.(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("embedded float slot-const multiply result = %#v, want raw f64 slot value", got)
	}
}

func TestBytecodeVM_BinaryFloatMulSlotConstParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"scale",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("zr"), ast.Flt(3.5)),
				ast.Assign(ast.ID("scaled"), ast.Bin("*", ast.Flt(2), ast.ID("zr"))),
				ast.ID("scaled"),
			},
			ast.Ty("f64"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("scale"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode float slot-const multiply mismatch: got=%#v want=%#v", got, want)
	}
	assertFloatValue(t, got, runtime.FloatF64, 7)
}
