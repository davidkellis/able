package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ReturnBinaryIntAddI32ReportsKnownSimpleTypeForHandledFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	instr := &bytecodeInstruction{op: bytecodeOpReturnBinaryIntAddI32, operator: "+"}

	vm.stack = []runtime.Value{
		runtime.NewSmallInt(19, runtime.IntegerI32),
		runtime.NewSmallInt(23, runtime.IntegerI32),
	}
	got, known, err := vm.execReturnBinaryIntAdd(instr)
	if err != nil {
		t.Fatalf("unexpected return-add i32 error: %v", err)
	}
	if known != bytecodeSimpleTypeCheckI32 {
		t.Fatalf("expected handled i32 return-add to report known i32, got %v", known)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(got)
	if !ok || kind != runtime.IntegerI32 || raw != 42 {
		t.Fatalf("unexpected return-add i32 raw result: got=%#v kind=%v raw=%d ok=%v", got, kind, raw, ok)
	}

	vm.stack = []runtime.Value{
		runtime.NewSmallInt(19, runtime.IntegerI64),
		runtime.NewSmallInt(23, runtime.IntegerI64),
	}
	got, known, err = vm.execReturnBinaryIntAdd(instr)
	if err != nil {
		t.Fatalf("unexpected generic return-add error: %v", err)
	}
	if known != bytecodeSimpleTypeCheckUnknown {
		t.Fatalf("expected generic fallback return-add to keep unknown simple type, got %v", known)
	}
	if !valuesEqual(bytecodeMaterializeRawValue(got), runtime.NewSmallInt(42, runtime.IntegerI64)) {
		t.Fatalf("unexpected generic return-add result: got=%#v", got)
	}
}

func TestBytecodeVM_ReturnBinaryIntAddRawI64UsesScratchCarrier(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = make([]runtime.Value, 0, 2)
	instr := &bytecodeInstruction{op: bytecodeOpReturnBinaryIntAdd, operator: "+"}
	left := &bytecodeRawI64SlotCell{Val: 19}
	right := &bytecodeRawI64SlotCell{Val: 23}
	runReturnBinaryRawI64Add(t, vm, instr, left, right)

	allocs := testing.AllocsPerRun(1000, func() {
		runReturnBinaryRawI64Add(t, vm, instr, left, right)
	})
	if allocs != 0 {
		t.Fatalf("expected raw i64 return-add hot path allocations to be zero, got %.2f", allocs)
	}
}

func runReturnBinaryRawI64Add(t *testing.T, vm *bytecodeVM, instr *bytecodeInstruction, left *bytecodeRawI64SlotCell, right *bytecodeRawI64SlotCell) {
	t.Helper()
	left.Val = 19
	right.Val = 23
	vm.stack = vm.stack[:0]
	vm.stack = append(vm.stack, left, right)
	got, known, err := vm.execReturnBinaryIntAdd(instr)
	if err != nil {
		t.Fatalf("unexpected return-add error: %v", err)
	}
	if known != bytecodeSimpleTypeCheckUnknown {
		t.Fatalf("expected generic raw return-add to report unknown simple type, got %v", known)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("expected return-add to pop operands, stack=%#v", vm.stack)
	}
	if _, ok := got.(*bytecodeRawIntegerReturnScratch); !ok {
		t.Fatalf("return-add result type = %T, want raw integer return scratch", got)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(got)
	if !ok || kind != runtime.IntegerI64 || raw != 42 {
		t.Fatalf("return-add result = %#v, want raw i64 42", got)
	}
}

func TestBytecodeVM_RawI64ProgramReturnNoCoercionKeepsScratchCarrier(t *testing.T) {
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:        ast.Ty("i64"),
			returnSimpleType:  "i64",
			returnSimpleCheck: bytecodeSimpleTypeCheckI64,
		},
	}
	value := &bytecodeRawIntegerReturnScratch{Raw: 42, TypeSuffix: runtime.IntegerI64}
	got, ok := bytecodeTryMaterializedProgramReturnNoCoercion(nil, nil, program, nil, value, bytecodeSimpleTypeCheckUnknown)
	if !ok {
		t.Fatalf("expected raw i64 return to skip materialization")
	}
	if got != value {
		t.Fatalf("raw return no-coercion result = %#v, want original scratch carrier", got)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		got, ok := bytecodeTryMaterializedProgramReturnNoCoercion(nil, nil, program, nil, value, bytecodeSimpleTypeCheckUnknown)
		if !ok || got != value {
			t.Fatalf("expected raw i64 return to skip materialization")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected raw i64 return no-coercion to allocate zero, got %.2f", allocs)
	}
}

func TestBytecodeVM_GenericInlineReturnKeepsRawIntegerCarrier(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:             ast.Ty("T"),
			returnSimpleType:       "T",
			returnTypeUsesGenerics: true,
		},
	}
	genericNames := map[string]struct{}{"T": {}}
	value := vm.rawIntegerReturnValue(runtime.IntegerI64, 42)
	got, err := vm.coerceInlineProgramReturnValue(program, nil, value, bytecodeSimpleTypeCheckUnknown, genericNames)
	if err != nil {
		t.Fatalf("generic inline return coercion failed: %v", err)
	}
	if got != value {
		t.Fatalf("generic inline return = %#v, want original raw carrier", got)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		value := vm.rawIntegerReturnValue(runtime.IntegerI64, 42)
		got, err := vm.coerceInlineProgramReturnValue(program, nil, value, bytecodeSimpleTypeCheckUnknown, genericNames)
		if err != nil || got != value {
			t.Fatalf("generic inline return = %#v err=%v, want original raw carrier", got, err)
		}
	})
	if allocs != 0 {
		t.Fatalf("expected generic inline raw return to allocate zero, got %.2f", allocs)
	}
}

func TestBytecodeVM_GenericProgramReturnStillMaterializesRawIntegerCarrier(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:             ast.Ty("T"),
			returnSimpleType:       "T",
			returnTypeUsesGenerics: true,
		},
	}
	genericNames := map[string]struct{}{"T": {}}
	value := vm.rawIntegerReturnValue(runtime.IntegerI64, 42)
	got, err := vm.coerceProgramReturnValue(program, nil, value, bytecodeSimpleTypeCheckUnknown, genericNames)
	if err != nil {
		t.Fatalf("generic program return coercion failed: %v", err)
	}
	if got == value {
		t.Fatalf("generic program return kept raw carrier on host-visible path")
	}
	assertIntValue(t, got, runtime.IntegerI64, 42)
}

func TestBytecodeVM_InferredInlineReturnKeepsRawIntegerCarrier(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{},
	}
	value := vm.rawIntegerReturnValue(runtime.IntegerI64, 42)
	got, err := vm.coerceInlineProgramReturnValue(program, nil, value, bytecodeSimpleTypeCheckUnknown, nil)
	if err != nil {
		t.Fatalf("inferred inline return coercion failed: %v", err)
	}
	if got != value {
		t.Fatalf("inferred inline return = %#v, want original raw carrier", got)
	}

	materialized, err := vm.coerceProgramReturnValue(program, nil, value, bytecodeSimpleTypeCheckUnknown, nil)
	if err != nil {
		t.Fatalf("inferred program return coercion failed: %v", err)
	}
	if materialized == value {
		t.Fatalf("inferred program return kept raw carrier on host-visible path")
	}
	assertIntValue(t, materialized, runtime.IntegerI64, 42)

	allocs := testing.AllocsPerRun(1000, func() {
		value := vm.rawIntegerReturnValue(runtime.IntegerI64, 42)
		got, err := vm.coerceInlineProgramReturnValue(program, nil, value, bytecodeSimpleTypeCheckUnknown, nil)
		if err != nil || got != value {
			t.Fatalf("inferred inline return = %#v err=%v, want original raw carrier", got, err)
		}
	})
	if allocs != 0 {
		t.Fatalf("expected inferred inline raw return to allocate zero, got %.2f", allocs)
	}
}

func TestBytecodeVM_AppendReturnValueCopiesRawIntegerScratch(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = make([]runtime.Value, 0, 1)
	value := vm.rawIntegerReturnValue(runtime.IntegerI64, 42)
	vm.appendReturnValue(value)
	if len(vm.stack) != 1 {
		t.Fatalf("stack size = %d, want 1", len(vm.stack))
	}
	if vm.stack[0] == value {
		t.Fatalf("appendReturnValue retained mutable return scratch")
	}
	vm.rawIntegerReturnScratch.Raw = 99
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerI64 || raw != 42 {
		t.Fatalf("stack result = %#v, want copied raw i64 42", vm.stack[0])
	}

	allocs := testing.AllocsPerRun(1000, func() {
		vm.stack = vm.stack[:0]
		value := vm.rawIntegerReturnValue(runtime.IntegerI64, 42)
		vm.appendReturnValue(value)
		if kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0]); !ok || kind != runtime.IntegerI64 || raw != 42 {
			t.Fatalf("stack result = %#v, want copied raw i64 42", vm.stack[0])
		}
	})
	if allocs != 0 {
		t.Fatalf("expected raw integer return append to allocate zero, got %.2f", allocs)
	}
}

func TestBytecodeVM_ReturnBinaryReportsKnownBoolForComparison(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	instr := &bytecodeInstruction{op: bytecodeOpReturnBinary, operator: "=="}

	vm.stack = []runtime.Value{
		runtime.NewSmallInt(17, runtime.IntegerI32),
		runtime.NewSmallInt(17, runtime.IntegerI32),
	}
	got, known, err := vm.execReturnBinary(instr)
	if err != nil {
		t.Fatalf("unexpected return-binary comparison error: %v", err)
	}
	if known != bytecodeSimpleTypeCheckBool {
		t.Fatalf("expected return-binary comparison to report known bool, got %v", known)
	}
	if !valuesEqual(got, runtime.BoolValue{Val: true}) {
		t.Fatalf("unexpected return-binary comparison result: got=%#v", got)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("expected return-binary comparison to pop operands, stack=%#v", vm.stack)
	}
}

func TestBytecodeVM_LoweringEmitsReturnBinaryIntAddForImplicitFinalExpression(t *testing.T) {
	def := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpReturnBinaryIntAddI32) {
		t.Fatalf("expected i32 lowering to emit fused return-add-i32 opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntAdd) {
		t.Fatalf("expected fused return-add shape to replace standalone add opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpReturnBinaryIntAdd) {
		t.Fatalf("expected i32 return-add shape to avoid generic return-add opcode")
	}
}

func TestBytecodeVM_ReturnBinaryIntAddImplicitFinalExpressionParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"sum",
			nil,
			[]ast.Statement{
				ast.Bin("+", ast.Int(19), ast.Int(23)),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("sum"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode return-add implicit final expression mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_LoweringEmitsReturnBinaryForImplicitFinalComparison(t *testing.T) {
	def := ast.Fn(
		"same",
		[]*ast.FunctionParameter{
			ast.Param("left", ast.Ty("i32")),
			ast.Param("right", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Bin("==", ast.ID("left"), ast.ID("right")),
		},
		ast.Ty("bool"),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpReturnBinary) {
		t.Fatalf("expected comparison final expression to emit generic return-binary opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpBinary) {
		t.Fatalf("expected return-binary comparison shape to replace standalone binary opcode")
	}
}

func TestBytecodeVM_ReturnBinaryImplicitFinalExpressionParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"same",
			[]*ast.FunctionParameter{
				ast.Param("left", ast.Ty("i32")),
				ast.Param("right", ast.Ty("i32")),
			},
			[]ast.Statement{
				ast.Bin("==", ast.ID("left"), ast.ID("right")),
			},
			ast.Ty("bool"),
			nil,
			nil,
			false,
			false,
		),
		ast.Fn(
			"diff",
			[]*ast.FunctionParameter{
				ast.Param("left", ast.Ty("i32")),
				ast.Param("right", ast.Ty("i32")),
			},
			[]ast.Statement{
				ast.Bin("-", ast.ID("left"), ast.ID("right")),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.NewIfExpression(
			ast.Call("same", ast.Call("diff", ast.Int(12), ast.Int(5)), ast.Int(7)),
			ast.Block(ast.Int(1)),
			[]*ast.ElseIfClause{},
			ast.Block(ast.Int(0)),
		),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode return-binary implicit final expression mismatch: got=%#v want=%#v", got, want)
	}
}
