package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestBytecodeVM_LoweringEmitsIntegerSlotConstHotOpcodes(t *testing.T) {
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.Bin("+", ast.ID("n"), ast.Int(1)),
			ast.Bin("<=", ast.ID("n"), ast.Int(2)),
			ast.Bin(">=", ast.ID("n"), ast.Int(3)),
			ast.Bin("-", ast.ID("n"), ast.Int(1)),
			ast.ID("n"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	sawAddSlotConst := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntAddSlotConst)
	sawSubSlotConst := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntSubSlotConst)
	sawLESlotConst := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntLessEqualSlotConst)
	sawCompareSlotConst := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntCompareSlotConst)
	if !sawAddSlotConst || !sawSubSlotConst || !sawLESlotConst || !sawCompareSlotConst {
		t.Fatalf("expected lowering to emit slot-const opcodes: add=%v sub=%v le=%v compare=%v", sawAddSlotConst, sawSubSlotConst, sawLESlotConst, sawCompareSlotConst)
	}
	for _, instr := range program.instructions {
		switch instr.op {
		case bytecodeOpBinaryIntAddSlotConst, bytecodeOpBinaryIntSubSlotConst, bytecodeOpBinaryIntLessEqualSlotConst, bytecodeOpBinaryIntCompareSlotConst:
			if !instr.hasIntImmediate {
				t.Fatalf("expected slot-const opcode %v to carry typed integer-immediate metadata", instr.op)
			}
			if got, ok := instr.intImmediate.ToInt64(); !ok || got <= 0 {
				t.Fatalf("expected slot-const opcode %v to keep positive integer immediate, got=%v ok=%v", instr.op, got, ok)
			}
		}
	}
}

func TestBytecodeVM_LoweringFusesSlotConstSelfAssignment(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("i"), ast.Int(1)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(2))),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("-", ast.ID("i"), ast.Int(1))),
			ast.ID("i"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	fused := 0
	for _, instr := range program.instructions {
		if instr.op != bytecodeOpStoreSlotBinaryIntSlotConst {
			continue
		}
		fused++
		if instr.name != "i" || (instr.operator != "+" && instr.operator != "-") {
			t.Fatalf("unexpected fused slot-const assignment: name=%q operator=%q", instr.name, instr.operator)
		}
		if !instr.hasIntImmediate {
			t.Fatalf("expected fused slot-const assignment to carry typed integer-immediate metadata")
		}
	}
	if fused != 2 {
		t.Fatalf("expected two fused slot-const self-assignments, got %d", fused)
	}
}

func TestBytecodeVM_LoweringKeepsTypedI32SelfAssignmentOnRawStore(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("i"), ast.Ty("i32")), ast.Int(1)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(2))),
			ast.ID("i"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotBinaryIntSlotConst) {
		t.Fatalf("typed i32 self-assignment should stay on raw i32 store path")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotI32) {
		t.Fatalf("expected typed i32 self-assignment to emit StoreSlotI32")
	}
}

func TestBytecodeVM_SlotConstSelfAssignmentParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"f",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("i"), ast.Int(3)),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(2))),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("-", ast.ID("i"), ast.Int(1))),
				ast.ID("i"),
			},
			nil,
			nil,
			nil,
			false,
			false,
		),
		ast.Call("f"),
	}, nil, nil)
	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode slot-const self-assignment mismatch: got=%#v want=%#v", got, want)
	}
}
