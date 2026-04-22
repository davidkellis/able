package interpreter

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_BinaryFastPathIntegerParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(9)),
		ast.Assign(ast.ID("y"), ast.Int(4)),
		ast.Bin("<=", ast.Bin("-", ast.ID("x"), ast.ID("y")), ast.Bin("+", ast.Int(2), ast.Int(3))),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode integer fast-path mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_DirectIntegerComparisonFastPath(t *testing.T) {
	lessLeft := runtime.NewSmallInt(4, runtime.IntegerI32)
	lessRight := runtime.NewSmallInt(9, runtime.IntegerI32)
	greaterLeft := runtime.NewSmallInt(9, runtime.IntegerI32)
	greaterRight := runtime.NewSmallInt(4, runtime.IntegerI32)
	equalLeft := runtime.NewSmallInt(7, runtime.IntegerI32)
	equalRight := runtime.NewSmallInt(7, runtime.IntegerI32)
	cases := []struct {
		name  string
		op    string
		left  runtime.Value
		right runtime.Value
		want  bool
	}{
		{name: "less", op: "<", left: lessLeft, right: lessRight, want: true},
		{name: "less_equal", op: "<=", left: lessLeft, right: lessRight, want: true},
		{name: "greater", op: ">", left: greaterLeft, right: greaterRight, want: true},
		{name: "greater_equal", op: ">=", left: greaterLeft, right: greaterRight, want: true},
		{name: "equal", op: "==", left: equalLeft, right: equalRight, want: true},
		{name: "not_equal", op: "!=", left: greaterLeft, right: greaterRight, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, handled := execBinaryDirectIntegerComparisonFast(tc.op, tc.left, tc.right)
			if !handled {
				t.Fatalf("expected fast path to handle %s", tc.op)
			}
			boolVal, ok := got.(runtime.BoolValue)
			if !ok {
				t.Fatalf("expected bool result, got %#v", got)
			}
			if boolVal.Val != tc.want {
				t.Fatalf("unexpected comparison result for %s: got=%v want=%v", tc.op, boolVal.Val, tc.want)
			}
		})
	}
}

func TestBytecodeVM_DirectSmallIntegerComparisonFastPath(t *testing.T) {
	leftVal := runtime.NewSmallInt(4, runtime.IntegerI32)
	rightVal := runtime.NewSmallInt(9, runtime.IntegerI32)
	leftPtr := runtime.NewSmallInt(9, runtime.IntegerI32)
	rightPtr := runtime.NewSmallInt(4, runtime.IntegerI32)

	cases := []struct {
		name  string
		op    string
		left  runtime.Value
		right runtime.Value
		want  bool
	}{
		{name: "value_pair", op: "<", left: leftVal, right: rightVal, want: true},
		{name: "pointer_pair", op: ">", left: &leftPtr, right: &rightPtr, want: true},
		{name: "mixed_pair", op: ">=", left: &leftPtr, right: rightVal, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := bytecodeDirectIntegerCompare(tc.op, tc.left, tc.right)
			if !ok {
				t.Fatalf("expected direct small-int comparison to handle %s", tc.name)
			}
			if got.Val != tc.want {
				t.Fatalf("unexpected result for %s: got=%v want=%v", tc.name, got.Val, tc.want)
			}
		})
	}
}

func TestBytecodeVM_DirectSameTypeSmallIntPair(t *testing.T) {
	leftVal := runtime.NewSmallInt(6, runtime.IntegerI32)
	rightVal := runtime.NewSmallInt(3, runtime.IntegerI32)
	rightPtr := runtime.NewSmallInt(2, runtime.IntegerI32)
	otherType := runtime.NewSmallInt(2, runtime.IntegerI64)

	kind, left, right, ok := bytecodeDirectSameTypeSmallIntPair(leftVal, rightVal)
	if !ok {
		t.Fatalf("expected exact same-type small-int value pair to succeed")
	}
	if kind != runtime.IntegerI32 || left != 6 || right != 3 {
		t.Fatalf("unexpected exact same-type small-int pair: kind=%v left=%d right=%d", kind, left, right)
	}

	kind, left, right, ok = bytecodeDirectSameTypeSmallIntPair(leftVal, &rightPtr)
	if !ok {
		t.Fatalf("expected same-type small-int pair to succeed")
	}
	if kind != runtime.IntegerI32 || left != 6 || right != 2 {
		t.Fatalf("unexpected same-type small-int pair: kind=%v left=%d right=%d", kind, left, right)
	}

	if _, _, _, ok := bytecodeDirectSameTypeSmallIntPair(leftVal, otherType); ok {
		t.Fatalf("expected mismatched integer kinds to miss fast pair path")
	}
}

func TestBytecodeVM_SubtractIntegerImmediateFast(t *testing.T) {
	leftVal := runtime.NewSmallInt(6, runtime.IntegerI32)
	leftPtr := runtime.NewSmallInt(9, runtime.IntegerI32)
	rightVal := runtime.NewSmallInt(2, runtime.IntegerI32)
	otherType := runtime.NewSmallInt(2, runtime.IntegerI64)

	got, handled, err := bytecodeSubtractIntegerImmediateFast(leftVal, rightVal)
	if err != nil {
		t.Fatalf("unexpected value-path error: %v", err)
	}
	if !handled {
		t.Fatalf("expected value-path immediate subtract fast path to handle operands")
	}
	if !valuesEqual(got, runtime.NewSmallInt(4, runtime.IntegerI32)) {
		t.Fatalf("unexpected value-path immediate subtract result: got=%#v", got)
	}

	got, handled, err = bytecodeSubtractIntegerImmediateFast(&leftPtr, rightVal)
	if err != nil {
		t.Fatalf("unexpected pointer-path error: %v", err)
	}
	if !handled {
		t.Fatalf("expected pointer-path immediate subtract fast path to handle operands")
	}
	if !valuesEqual(got, runtime.NewSmallInt(7, runtime.IntegerI32)) {
		t.Fatalf("unexpected pointer-path immediate subtract result: got=%#v", got)
	}

	if _, handled, err := bytecodeSubtractIntegerImmediateFast(leftVal, otherType); err != nil || handled {
		t.Fatalf("expected mismatched integer kinds to miss immediate subtract fast path, handled=%v err=%v", handled, err)
	}
}

func TestBytecodeVM_AddSmallI32PairFast(t *testing.T) {
	rightPtr := runtime.NewSmallInt(4, runtime.IntegerI32)
	cases := []struct {
		name  string
		left  runtime.Value
		right runtime.Value
		want  runtime.Value
	}{
		{
			name:  "value_pair",
			left:  runtime.NewSmallInt(6, runtime.IntegerI32),
			right: runtime.NewSmallInt(3, runtime.IntegerI32),
			want:  runtime.NewSmallInt(9, runtime.IntegerI32),
		},
		{
			name:  "mixed_pair",
			left:  runtime.NewSmallInt(5, runtime.IntegerI32),
			right: &rightPtr,
			want:  runtime.NewSmallInt(9, runtime.IntegerI32),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, handled, err := bytecodeAddSmallI32PairFast(tc.left, tc.right)
			if err != nil {
				t.Fatalf("unexpected add fast-path error: %v", err)
			}
			if !handled {
				t.Fatalf("expected i32 pair add fast path to handle operands")
			}
			if !valuesEqual(got, tc.want) {
				t.Fatalf("unexpected i32 pair add result: got=%#v want=%#v", got, tc.want)
			}
		})
	}
}

func TestBytecodeVM_SubtractSmallI32PairFast(t *testing.T) {
	left := runtime.NewSmallInt(9, runtime.IntegerI32)
	right := runtime.NewSmallInt(2, runtime.IntegerI32)

	got, handled, err := bytecodeSubtractSmallI32PairFast(left, right)
	if err != nil {
		t.Fatalf("unexpected subtract fast-path error: %v", err)
	}
	if !handled {
		t.Fatalf("expected i32 pair subtract fast path to handle operands")
	}
	if !valuesEqual(got, runtime.NewSmallInt(7, runtime.IntegerI32)) {
		t.Fatalf("unexpected i32 pair subtract result: got=%#v", got)
	}
}

func TestBytecodeVM_BinaryFastPathGeneralIntegerComparisonParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Bin("<", ast.Int(4), ast.Int(9)),
		ast.Bin(">=", ast.Int(9), ast.Int(4)),
		ast.Bin(">", ast.Int(9), ast.Int(4)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode general integer comparison fast-path mismatch: got=%#v want=%#v", got, want)
	}
}

func TestApplyBinaryOperatorFast_StringComparison(t *testing.T) {
	got, handled, err := ApplyBinaryOperatorFast("<", runtime.StringValue{Val: "alpha"}, runtime.StringValue{Val: "beta"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handled {
		t.Fatalf("expected string comparison fast path to handle operands")
	}
	boolVal, ok := got.(runtime.BoolValue)
	if !ok {
		t.Fatalf("expected bool result, got %#v", got)
	}
	if !boolVal.Val {
		t.Fatalf("expected alpha < beta to be true")
	}
}

func TestBytecodeVM_BinaryFastPathOverflowParity(t *testing.T) {
	i8 := ast.IntegerTypeI8
	module := ast.Mod([]ast.Statement{
		ast.Bin("+", ast.IntTyped(127, &i8), ast.IntTyped(1, &i8)),
	}, nil, nil)

	treeErr := evalModuleError(t, New(), module)
	if treeErr == nil || !strings.Contains(treeErr.Error(), "integer overflow") {
		t.Fatalf("expected tree overflow error, got: %v", treeErr)
	}
	byteErr := runBytecodeModuleError(t, NewBytecode(), module)
	if byteErr == nil || !strings.Contains(byteErr.Error(), "integer overflow") {
		t.Fatalf("expected bytecode overflow error, got: %v", byteErr)
	}
}

func TestBytecodeVM_BinaryFastPathTypeErrorParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Bin("+", ast.Int(1), ast.Bool(true)),
	}, nil, nil)

	treeErr := evalModuleError(t, New(), module)
	if treeErr == nil || !strings.Contains(treeErr.Error(), "Arithmetic requires numeric operands") {
		t.Fatalf("expected tree arithmetic type error, got: %v", treeErr)
	}
	byteErr := runBytecodeModuleError(t, NewBytecode(), module)
	if byteErr == nil || !strings.Contains(byteErr.Error(), "Arithmetic requires numeric operands") {
		t.Fatalf("expected bytecode arithmetic type error, got: %v", byteErr)
	}
}

func TestBytecodeVM_BinaryFastPathFloatFallbackParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Bin("<=", ast.Bin("-", ast.Flt(7.5), ast.Flt(2.0)), ast.Bin("+", ast.Flt(2.25), ast.Flt(3.25))),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode float fallback mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_BinaryIntDivCastFastPathParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.NewTypeCastExpression(ast.Bin("/", ast.Int(9), ast.Int(2)), ast.Ty("i32")),
		ast.NewTypeCastExpression(ast.Bin("/", ast.Int(-9), ast.Int(2)), ast.Ty("i32")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode int-div-cast fast path mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_BinaryIntDivCastFloatFallbackParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.NewTypeCastExpression(ast.Bin("/", ast.Flt(9.0), ast.Flt(2.0)), ast.Ty("i32")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode int-div-cast float fallback mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_BinaryIntDivCastDivisionByZeroParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.NewTypeCastExpression(ast.Bin("/", ast.Int(9), ast.Int(0)), ast.Ty("i32")),
	}, nil, nil)

	treeErr := evalModuleError(t, New(), module)
	if treeErr == nil || !strings.Contains(treeErr.Error(), "division by zero") {
		t.Fatalf("expected tree division by zero error, got: %v", treeErr)
	}
	byteErr := runBytecodeModuleError(t, NewBytecode(), module)
	if byteErr == nil || !strings.Contains(byteErr.Error(), "division by zero") {
		t.Fatalf("expected bytecode division by zero error, got: %v", byteErr)
	}
}

func TestBytecodeVM_LoweringEmitsIntegerBinaryHotOpcodes(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Bin("+", ast.Int(1), ast.Int(2)),
		ast.Bin("-", ast.Int(4), ast.Int(3)),
		ast.Bin("<=", ast.Int(1), ast.Int(2)),
	}, nil, nil)

	interp := NewBytecode()
	program, err := interp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	sawAdd := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntAdd)
	sawSub := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntSub)
	sawLE := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntLessEqual)
	if !sawAdd || !sawSub || !sawLE {
		t.Fatalf("expected lowering to emit all specialized binary opcodes: add=%v sub=%v le=%v", sawAdd, sawSub, sawLE)
	}
}

func TestBytecodeVM_LoweringEmitsIntegerDivCastOpcode(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.NewTypeCastExpression(ast.Bin("/", ast.Int(9), ast.Int(2)), ast.Ty("i32")),
	}, nil, nil)

	interp := NewBytecode()
	program, err := interp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntDivCast) {
		t.Fatalf("expected lowering to emit integer div-cast opcode")
	}
}

func TestBytecodeVM_LoweringEmitsIntegerSlotConstHotOpcodes(t *testing.T) {
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.Bin("+", ast.ID("n"), ast.Int(1)),
			ast.Bin("<=", ast.ID("n"), ast.Int(2)),
			ast.Bin("-", ast.ID("n"), ast.Int(1)),
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
	if !sawAddSlotConst || !sawSubSlotConst || !sawLESlotConst {
		t.Fatalf("expected lowering to emit slot-const opcodes: add=%v sub=%v le=%v", sawAddSlotConst, sawSubSlotConst, sawLESlotConst)
	}
	for _, instr := range program.instructions {
		switch instr.op {
		case bytecodeOpBinaryIntAddSlotConst, bytecodeOpBinaryIntSubSlotConst, bytecodeOpBinaryIntLessEqualSlotConst:
			if !instr.hasIntImmediate {
				t.Fatalf("expected slot-const opcode %v to carry typed integer-immediate metadata", instr.op)
			}
			if got, ok := instr.intImmediate.ToInt64(); !ok || got <= 0 {
				t.Fatalf("expected slot-const opcode %v to keep positive integer immediate, got=%v ok=%v", instr.op, got, ok)
			}
		}
	}
}

func TestBytecodeVM_LoweringEmitsFusedSelfCallSlotConstOpcode(t *testing.T) {
	def := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(2)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
			),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpCallSelfIntSubSlotConst) {
		t.Fatalf("expected lowering to emit fused self-call slot-const opcode")
	}
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpCallSelfIntSubSlotConst {
			if !instr.hasIntImmediate {
				t.Fatalf("expected fused self-call slot-const opcode to carry typed integer-immediate metadata")
			}
			if got, ok := instr.intImmediate.ToInt64(); !ok || got <= 0 {
				t.Fatalf("expected fused self-call slot-const immediate, got=%v ok=%v", got, ok)
			}
		}
	}
}

func TestBytecodeVM_LoweringSkipsDeadNilForStatementIfWithoutElse(t *testing.T) {
	def := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("Int"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(1)),
				ast.Block(ast.Ret(ast.ID("n"))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
			),
		},
		ast.Ty("Int"),
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
	for _, instr := range program.instructions {
		if instr.op != bytecodeOpConst {
			continue
		}
		if _, ok := instr.value.(runtime.NilValue); ok {
			t.Fatalf("expected statement-position if without else to skip dead nil const lowering")
		}
	}
}

func TestBytecodeVM_StatementIfWithoutElseParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"f",
			[]*ast.FunctionParameter{ast.Param("n", ast.Ty("Int"))},
			[]ast.Statement{
				ast.IfExpr(
					ast.Bin("<=", ast.ID("n"), ast.Int(1)),
					ast.Block(ast.Ret(ast.ID("n"))),
				),
				ast.Bin("-", ast.ID("n"), ast.Int(1)),
			},
			ast.Ty("Int"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("f", ast.Int(7)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("statement-position if without else mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_BinarySlotConstTypeErrorParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"f",
			[]*ast.FunctionParameter{ast.Param("x", nil)},
			[]ast.Statement{
				ast.Bin("-", ast.ID("x"), ast.Int(1)),
			},
			nil,
			nil,
			nil,
			false,
			false,
		),
		ast.Call("f", ast.Bool(true)),
	}, nil, nil)

	treeErr := evalModuleError(t, New(), module)
	if treeErr == nil || !strings.Contains(treeErr.Error(), "Arithmetic requires numeric operands") {
		t.Fatalf("expected tree arithmetic type error, got: %v", treeErr)
	}
	byteErr := runBytecodeModuleError(t, NewBytecode(), module)
	if byteErr == nil || !strings.Contains(byteErr.Error(), "Arithmetic requires numeric operands") {
		t.Fatalf("expected bytecode arithmetic type error, got: %v", byteErr)
	}
}

func TestBytecodeVM_BinaryAddSlotConstParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(9)),
		ast.Bin("+", ast.ID("x"), ast.Int(1)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode add slot-const mismatch: got=%#v want=%#v", got, want)
	}
}

func bytecodeProgramContainsOpcode(program *bytecodeProgram, target bytecodeOp) bool {
	if program == nil {
		return false
	}
	for _, instr := range program.instructions {
		if instr.op == target {
			return true
		}
		if instr.program != nil && bytecodeProgramContainsOpcode(instr.program, target) {
			return true
		}
	}
	return false
}
