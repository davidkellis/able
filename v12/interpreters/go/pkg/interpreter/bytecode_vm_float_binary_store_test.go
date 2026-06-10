package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsFloatBinaryStore(t *testing.T) {
	def := ast.Fn(
		"step",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("zr"), ast.Flt(3)),
			ast.Assign(ast.ID("zr2"), ast.Bin("*", ast.ID("zr"), ast.ID("zr"))),
			ast.ID("zr2"),
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
		t.Fatalf("expected lowering to emit fused float binary store")
	}
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpStoreSlotFloatBinary {
			if instr.target < 0 || instr.argCount < 0 || instr.loopBreak < 0 || instr.operator != "*" {
				t.Fatalf("unexpected fused float binary store instruction: %#v", instr)
			}
		}
	}
}

func TestBytecodeVM_LoweringEmitsFloatAddMulSlotUpdateWithNonTargetBase(t *testing.T) {
	def := ast.Fn(
		"step",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("zi"), ast.Flt(0.5)),
			ast.Assign(ast.ID("zr"), ast.Flt(2)),
			ast.Assign(ast.ID("ci"), ast.Flt(1.25)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("zi"), ast.Bin("+", ast.Bin("*", ast.ID("zr"), ast.ID("zi")), ast.ID("ci"))),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatAddMulSlot) {
		t.Fatalf("expected lowering to emit slot-sourced fused float add-mul update")
	}
}

func TestBytecodeVM_FloatBinaryStoreParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"step",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("zr"), ast.Flt(3)),
				ast.Assign(ast.ID("zi"), ast.Flt(4)),
				ast.Assign(ast.ID("zr2"), ast.Bin("*", ast.ID("zr"), ast.ID("zr"))),
				ast.Assign(ast.ID("zi2"), ast.Bin("*", ast.ID("zi"), ast.ID("zi"))),
				ast.Assign(ast.ID("diff"), ast.Bin("-", ast.ID("zr2"), ast.ID("zi2"))),
				ast.ID("diff"),
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
		t.Fatalf("bytecode float binary store mismatch: got=%#v want=%#v", got, want)
	}
	assertFloatValue(t, got, runtime.FloatF64, -7)
}

func TestBytecodeVM_FloatBinaryStoreDiscardResultKeepsSnapshotSemantics(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{
		runtime.FloatValue{Val: 3, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 4, TypeSuffix: runtime.FloatF64},
		runtime.NilValue{},
	}
	instr := &bytecodeInstruction{
		op:            bytecodeOpStoreSlotFloatBinary,
		target:        2,
		argCount:      0,
		loopBreak:     1,
		operator:      "*",
		discardResult: true,
	}

	if err := vm.execStoreSlotFloatBinary(instr); err != nil {
		t.Fatalf("first fused float binary store failed: %v", err)
	}
	assertFloatValue(t, vm.slots[2], runtime.FloatF64, 12)
	if _, ok := vm.slots[2].(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("first fused float binary store slot = %#v, want raw f64 slot value", vm.slots[2])
	}
	if len(vm.stack) != 0 {
		t.Fatalf("discarded fused float binary store stack = %#v, want empty", vm.stack)
	}

	vm.stack = nil
	if err := vm.execLoadSlotOpcode(&bytecodeInstruction{op: bytecodeOpLoadSlot, target: 2}); err != nil {
		t.Fatalf("load fused float binary result: %v", err)
	}
	snapshot := vm.stack[0]
	assertFloatValue(t, snapshot, runtime.FloatF64, 12)
	if _, ok := snapshot.(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("loaded fused float binary snapshot = %#v, want raw f64 stack value", snapshot)
	}

	vm.slots[0] = runtime.FloatValue{Val: 5, TypeSuffix: runtime.FloatF64}
	vm.slots[1] = runtime.FloatValue{Val: 6, TypeSuffix: runtime.FloatF64}
	vm.stack = nil
	vm.ip = 0
	if err := vm.execStoreSlotFloatBinary(instr); err != nil {
		t.Fatalf("second fused float binary store failed: %v", err)
	}
	assertFloatValue(t, vm.slots[2], runtime.FloatF64, 30)
	assertFloatValue(t, snapshot, runtime.FloatF64, 12)
}

func TestBytecodeVM_FloatAddMulNonTargetBaseParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"step",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("zi"), ast.Flt(0.5)),
				ast.Assign(ast.ID("zr"), ast.Flt(2)),
				ast.Assign(ast.ID("ci"), ast.Flt(1.25)),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("zi"), ast.Bin("+", ast.Bin("*", ast.ID("zr"), ast.ID("zi")), ast.ID("ci"))),
				ast.ID("zi"),
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
		t.Fatalf("bytecode float add-mul non-target base mismatch: got=%#v want=%#v", got, want)
	}
	assertFloatValue(t, got, runtime.FloatF64, 2.25)
}
