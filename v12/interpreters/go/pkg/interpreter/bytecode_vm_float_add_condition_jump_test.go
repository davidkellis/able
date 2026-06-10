package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsFloatAddCompareConstJump(t *testing.T) {
	def := ast.Fn(
		"escaped",
		[]*ast.FunctionParameter{
			ast.Param("zr", ast.Ty("f64")),
			ast.Param("zi", ast.Ty("f64")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("zr2"), ast.Bin("*", ast.ID("zr"), ast.ID("zr"))),
			ast.Assign(ast.ID("zi2"), ast.Bin("*", ast.ID("zi"), ast.ID("zi"))),
			ast.IfExpr(
				ast.Bin(">", ast.Bin("+", ast.ID("zr2"), ast.ID("zi2")), ast.Flt(4.0)),
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
		if instr.op != bytecodeOpJumpIfFloatAddCompareConstFalse {
			continue
		}
		sawJump = true
		if instr.operator != ">" {
			t.Fatalf("float add compare operator = %q, want >", instr.operator)
		}
		plan, ok := program.floatAddCompareConstJumps[ip]
		if !ok {
			t.Fatalf("missing float add compare plan for ip %d", ip)
		}
		if plan.leftSlot < 0 || plan.rightSlot < 0 || plan.leftSlot == plan.rightSlot {
			t.Fatalf("float add compare slots = %#v, want two distinct resolved local slots", plan)
		}
		if plan.rightImmediate.TypeSuffix != runtime.FloatF64 || plan.rightImmediate.Val != 4.0 {
			t.Fatalf("float add compare immediate = %#v, want f64 4.0", plan.rightImmediate)
		}
	}
	if !sawJump {
		t.Fatalf("expected lowering to emit float add compare const jump")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfFalse) {
		t.Fatalf("expected fused float add compare condition to skip generic jump-if-false bool materialization")
	}
}

func TestBytecodeVM_FloatAddCompareConstJumpParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"escaped",
			[]*ast.FunctionParameter{
				ast.Param("zr", ast.Ty("f64")),
				ast.Param("zi", ast.Ty("f64")),
			},
			[]ast.Statement{
				ast.Assign(ast.ID("zr2"), ast.Bin("*", ast.ID("zr"), ast.ID("zr"))),
				ast.Assign(ast.ID("zi2"), ast.Bin("*", ast.ID("zi"), ast.ID("zi"))),
				ast.IfExpr(
					ast.Bin(">", ast.Bin("+", ast.ID("zr2"), ast.ID("zi2")), ast.Flt(4.0)),
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
		ast.CallExpr(ast.ID("escaped"), ast.Flt(2.0), ast.Flt(1.0)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode float add compare mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_FloatAddCompareConstJumpFastPathWithRawFloatSlots(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		floatAddCompareConstJumps: map[int]bytecodeFloatAddCompareConstJumpPlan{
			0: {
				leftSlot:  0,
				rightSlot: 1,
				rightImmediate: runtime.FloatValue{
					Val:        4.0,
					TypeSuffix: runtime.FloatF64,
				},
			},
		},
	}
	vm.slots = []runtime.Value{
		bytecodeRawFloatSlotValue(1.5, runtime.FloatF64),
		bytecodeRawFloatSlotValue(1.0, runtime.FloatF64),
	}
	instr := &bytecodeInstruction{
		op:       bytecodeOpJumpIfFloatAddCompareConstFalse,
		target:   9,
		operator: ">",
	}

	if err := vm.execJumpIfFloatAddCompareConstFalse(instr, program); err != nil {
		t.Fatalf("float add compare fallback plan failed: %v", err)
	}
	if vm.ip != 9 {
		t.Fatalf("false float add compare should jump to 9, got %d", vm.ip)
	}

	vm.ip = 0
	vm.slots[0] = bytecodeRawFloatSlotValue(2.5, runtime.FloatF64)
	vm.slots[1] = bytecodeRawFloatSlotValue(2.0, runtime.FloatF64)
	if err := vm.execJumpIfFloatAddCompareConstFalse(instr, program); err != nil {
		t.Fatalf("float add compare true path failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("truthy float add compare should advance ip to 1, got %d", vm.ip)
	}
}
