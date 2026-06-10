package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestEvaluateLambdaExpression_LowersSlotBytecodeProgram(t *testing.T) {
	interp := NewBytecode()
	lambda := ast.Lam(
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("i32"))},
		ast.Bin("+", ast.ID("x"), ast.Int(1)),
	)

	value, err := interp.evaluateLambdaExpression(lambda, interp.GlobalEnvironment())
	if err != nil {
		t.Fatalf("evaluateLambdaExpression: %v", err)
	}
	fn, ok := value.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("expected function value, got %#v", value)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil {
		t.Fatalf("expected bytecode program on lambda, got %#v", fn.Bytecode)
	}
	if program.frameLayout == nil {
		t.Fatalf("expected lambda bytecode to use slot frame layout")
	}
	if program.frameLayout.paramSlots != 1 {
		t.Fatalf("paramSlots = %d, want 1", program.frameLayout.paramSlots)
	}
}

func TestEvaluateLambdaExpression_ReusesProgramAcrossUnrelatedBindingShapeChanges(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironment(interp.GlobalEnvironment())
	lambda := ast.Lam(
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("i32"))},
		ast.Bin("+", ast.ID("x"), ast.Int(1)),
	)

	firstValue, err := interp.evaluateLambdaExpression(lambda, env)
	if err != nil {
		t.Fatalf("first evaluateLambdaExpression: %v", err)
	}
	secondValue, err := interp.evaluateLambdaExpression(lambda, env)
	if err != nil {
		t.Fatalf("second evaluateLambdaExpression: %v", err)
	}
	first := firstValue.(*runtime.FunctionValue).Bytecode.(*bytecodeProgram)
	second := secondValue.(*runtime.FunctionValue).Bytecode.(*bytecodeProgram)
	if first != second {
		t.Fatalf("expected matching lambda and environment shape to reuse bytecode program")
	}

	env.Define("new_binding", runtime.NilValue{})
	thirdValue, err := interp.evaluateLambdaExpression(lambda, env)
	if err != nil {
		t.Fatalf("third evaluateLambdaExpression: %v", err)
	}
	third := thirdValue.(*runtime.FunctionValue).Bytecode.(*bytecodeProgram)
	if third != first {
		t.Fatalf("unreferenced binding-shape change invalidated cached lambda bytecode")
	}
}

func TestEvaluateLambdaExpression_InvalidatesProgramWhenReferencedStructDefinitionChanges(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironment(interp.GlobalEnvironment())
	lambda := ast.Lam(nil, ast.StructLit(nil, false, "Box", nil, nil))

	firstValue, err := interp.evaluateLambdaExpression(lambda, env)
	if err != nil {
		t.Fatalf("first evaluateLambdaExpression: %v", err)
	}
	first := firstValue.(*runtime.FunctionValue).Bytecode.(*bytecodeProgram)

	boxDef := ast.StructDef("Box", nil, ast.StructKindNamed, nil, nil, false)
	env.DefineStruct("Box", &runtime.StructDefinitionValue{Node: boxDef})
	secondValue, err := interp.evaluateLambdaExpression(lambda, env)
	if err != nil {
		t.Fatalf("second evaluateLambdaExpression: %v", err)
	}
	second := secondValue.(*runtime.FunctionValue).Bytecode.(*bytecodeProgram)
	if second == first {
		t.Fatalf("referenced struct-definition change did not invalidate cached lambda bytecode")
	}
}

func TestEvaluateLambdaExpression_TracksReferencedBindingShapeOnly(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironment(interp.GlobalEnvironment())
	lambda := ast.Lam(nil, ast.ID("captured"))

	firstValue, err := interp.evaluateLambdaExpression(lambda, env)
	if err != nil {
		t.Fatalf("first evaluateLambdaExpression: %v", err)
	}
	first := firstValue.(*runtime.FunctionValue).Bytecode.(*bytecodeProgram)

	env.Define("captured", runtime.NewSmallInt(1, runtime.IntegerI32))
	secondValue, err := interp.evaluateLambdaExpression(lambda, env)
	if err != nil {
		t.Fatalf("second evaluateLambdaExpression: %v", err)
	}
	second := secondValue.(*runtime.FunctionValue).Bytecode.(*bytecodeProgram)
	if second == first {
		t.Fatalf("referenced binding addition did not invalidate cached lambda bytecode")
	}

	if err := env.Assign("captured", runtime.NewSmallInt(2, runtime.IntegerI32)); err != nil {
		t.Fatalf("assign captured: %v", err)
	}
	thirdValue, err := interp.evaluateLambdaExpression(lambda, env)
	if err != nil {
		t.Fatalf("third evaluateLambdaExpression: %v", err)
	}
	third := thirdValue.(*runtime.FunctionValue).Bytecode.(*bytecodeProgram)
	if third != second {
		t.Fatalf("ordinary value update invalidated immutable cached lambda bytecode")
	}
}

func TestEvaluateLambdaExpression_GenericLambdaCachesReturnNames(t *testing.T) {
	interp := NewBytecode()
	lambda := ast.NewLambdaExpression(
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		ast.ID("x"),
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)

	value, err := interp.evaluateLambdaExpression(lambda, interp.GlobalEnvironment())
	if err != nil {
		t.Fatalf("evaluateLambdaExpression: %v", err)
	}
	fn, ok := value.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("expected function value, got %#v", value)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil {
		t.Fatalf("expected bytecode program on lambda, got %#v", fn.Bytecode)
	}
	if program.frameLayout == nil {
		t.Fatalf("expected generic lambda bytecode to use slot frame layout")
	}
	if !program.returnGenericNamesCached {
		t.Fatalf("expected generic return-name cache to be populated")
	}
	if _, ok := program.returnGenericNames["T"]; !ok {
		t.Fatalf("expected generic lambda return names to include T, got %#v", program.returnGenericNames)
	}
}

func TestBytecodeVM_GenericLambdaCallWithTypeArgumentsPreservesRuntimeTypeBindings(t *testing.T) {
	lambda := ast.NewLambdaExpression(
		[]*ast.FunctionParameter{ast.Param("x", nil)},
		ast.NewTypeCastExpression(ast.ID("x"), ast.Ty("T")),
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	call := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.Int(4)}, []ast.TypeExpression{ast.Ty("i32")}, false)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("id"), lambda),
		call,
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("generic lambda bytecode mismatch: got=%#v want=%#v", got, want)
	}
}
