package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsFloatAddSubSlotUpdate(t *testing.T) {
	def := ast.Fn(
		"step",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("zr"), ast.Flt(3)),
			ast.Assign(ast.ID("zi"), ast.Flt(4)),
			ast.Assign(ast.ID("cr"), ast.Flt(1.5)),
			ast.Assign(ast.ID("zr2"), ast.Bin("*", ast.ID("zr"), ast.ID("zr"))),
			ast.Assign(ast.ID("zi2"), ast.Bin("*", ast.ID("zi"), ast.ID("zi"))),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("zr"), ast.Bin("+", ast.Bin("-", ast.ID("zr2"), ast.ID("zi2")), ast.ID("cr"))),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatAddSub) {
		t.Fatalf("expected lowering to emit fused float add-sub slot update")
	}
}

func TestBytecodeVM_StoreSlotFloatAddSubFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NilValue{}}
	vm.stack = []runtime.Value{
		runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 9, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 4, TypeSuffix: runtime.FloatF64},
	}
	instr := &bytecodeInstruction{op: bytecodeOpStoreSlotFloatAddSub, target: 0}

	if err := vm.execStoreSlotFloatAddSub(instr); err != nil {
		t.Fatalf("fused float add-sub store failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("ip after fused float add-sub store = %d, want 1", vm.ip)
	}
	assertFloatValue(t, vm.slots[0], runtime.FloatF64, 6.5)
	if _, ok := vm.slots[0].(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("stored fused float add-sub slot = %#v, want raw f64 slot value", vm.slots[0])
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack length after fused float add-sub store = %d, want 1", len(vm.stack))
	}
	assertFloatValue(t, vm.stack[0], runtime.FloatF64, 6.5)
}

func TestBytecodeVM_StoreSlotFloatAddSubFastPathNormalizesF32Result(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NilValue{}}
	vm.stack = []runtime.Value{
		runtime.FloatValue{Val: 1.1, TypeSuffix: runtime.FloatF32},
		runtime.FloatValue{Val: 0.6, TypeSuffix: runtime.FloatF32},
		runtime.FloatValue{Val: 0.2, TypeSuffix: runtime.FloatF32},
	}
	instr := &bytecodeInstruction{op: bytecodeOpStoreSlotFloatAddSub, target: 0}

	if err := vm.execStoreSlotFloatAddSub(instr); err != nil {
		t.Fatalf("fused f32 float add-sub store failed: %v", err)
	}
	want := normalizeFloat(runtime.FloatF32, (0.6-0.2)+1.1)
	assertFloatValue(t, vm.slots[0], runtime.FloatF32, want)
	if _, ok := vm.slots[0].(bytecodeRawF32SlotValue); !ok {
		t.Fatalf("stored fused f32 float add-sub slot = %#v, want raw f32 slot value", vm.slots[0])
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack length after fused f32 float add-sub store = %d, want 1", len(vm.stack))
	}
	assertFloatValue(t, vm.stack[0], runtime.FloatF32, want)
	if _, ok := vm.stack[0].(bytecodeRawF32SlotValue); !ok {
		t.Fatalf("stack result = %#v, want raw f32 slot value", vm.stack[0])
	}
}

func TestBytecodeVM_FloatAddSubSlotUpdateParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"step",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("zr"), ast.Flt(3)),
				ast.Assign(ast.ID("zi"), ast.Flt(4)),
				ast.Assign(ast.ID("cr"), ast.Flt(1.5)),
				ast.Assign(ast.ID("zr2"), ast.Bin("*", ast.ID("zr"), ast.ID("zr"))),
				ast.Assign(ast.ID("zi2"), ast.Bin("*", ast.ID("zi"), ast.ID("zi"))),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("zr"), ast.Bin("+", ast.Bin("-", ast.ID("zr2"), ast.ID("zi2")), ast.ID("cr"))),
				ast.ID("zr"),
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
		t.Fatalf("bytecode float add-sub slot update mismatch: got=%#v want=%#v", got, want)
	}
	assertFloatValue(t, got, runtime.FloatF64, -5.5)
}
