package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestReusableBytecodeCallEnvForExplicitBindingsCachesEnvironment(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		[]ast.Statement{ast.ID("T_type")},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	call := ast.NewFunctionCall(ast.ID("id"), nil, []ast.TypeExpression{ast.Ty("i32")}, false)
	fn := &runtime.FunctionValue{Declaration: decl, Closure: closure}
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}

	first, ok := interp.reusableBytecodeCallEnvForExplicitBindings(fn, decl, call, program)
	if !ok || first == nil {
		t.Fatalf("expected reusable bytecode env")
	}
	second, ok := interp.reusableBytecodeCallEnvForExplicitBindings(fn, decl, call, program)
	if !ok || second == nil {
		t.Fatalf("expected cached reusable bytecode env")
	}
	if first != second {
		t.Fatalf("expected cached env reuse, got %p and %p", first, second)
	}
	if first.Parent() != closure {
		t.Fatalf("cached env parent = %p, want %p", first.Parent(), closure)
	}
	if got, ok := first.Lookup("T"); !ok {
		t.Fatalf("expected cached env to define T")
	} else if ref, ok := got.(runtime.TypeRefValue); !ok || ref.TypeName != "i32" {
		t.Fatalf("cached env T = %#v, want TypeRefValue{i32}", got)
	}
	if got, ok := first.Lookup("T_type"); !ok {
		t.Fatalf("expected cached env to define T_type")
	} else if str, ok := got.(runtime.StringValue); !ok || str.Val != "i32" {
		t.Fatalf("cached env T_type = %#v, want StringValue{i32}", got)
	}
	if got := len(interp.reusableBytecodeCallEnvCache); got != 1 {
		t.Fatalf("reusable env cache size = %d, want 1", got)
	}
}

func TestReusableBytecodeCallEnvForExplicitBindingsReusesAcrossEquivalentCallSites(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		[]ast.Statement{ast.ID("T_type")},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	fn := &runtime.FunctionValue{Declaration: decl, Closure: closure}
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}
	firstCall := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.Int(1)}, []ast.TypeExpression{ast.Ty("i32")}, false)
	secondCall := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.Int(2)}, []ast.TypeExpression{ast.Ty("i32")}, false)

	first, ok := interp.reusableBytecodeCallEnvForExplicitBindings(fn, decl, firstCall, program)
	if !ok || first == nil {
		t.Fatalf("expected reusable bytecode env for first explicit call site")
	}
	second, ok := interp.reusableBytecodeCallEnvForExplicitBindings(fn, decl, secondCall, program)
	if !ok || second == nil {
		t.Fatalf("expected reusable bytecode env for second explicit call site")
	}
	if first != second {
		t.Fatalf("expected equivalent explicit call sites to share env, got %p and %p", first, second)
	}
	if got := len(interp.reusableBytecodeCallEnvCache); got != 1 {
		t.Fatalf("reusable env cache size = %d, want 1", got)
	}
}

func TestReusableBytecodeCallEnvForExplicitBindingsSkipsIneligibleCalls(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		[]ast.Statement{ast.ID("T_type")},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	call := ast.NewFunctionCall(ast.ID("id"), nil, nil, false)
	fn := &runtime.FunctionValue{Declaration: decl, Closure: closure}

	if env, ok := interp.reusableBytecodeCallEnvForExplicitBindings(fn, decl, call, &bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}); ok || env != nil {
		t.Fatalf("expected no reusable env without explicit or inferred type bindings")
	}
	if env, ok := interp.reusableBytecodeCallEnvForExplicitBindings(fn, decl, ast.NewFunctionCall(ast.ID("id"), nil, []ast.TypeExpression{ast.Ty("i32")}, false), &bytecodeProgram{frameLayout: &bytecodeFrameLayout{needsEnvScopes: true}}); ok || env != nil {
		t.Fatalf("expected no reusable env when frame layout needs env scopes")
	}
	if env, ok := interp.reusableBytecodeCallEnvForExplicitBindings(&runtime.FunctionValue{Declaration: decl}, decl, ast.NewFunctionCall(ast.ID("id"), nil, []ast.TypeExpression{ast.Ty("i32")}, false), &bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}); ok || env != nil {
		t.Fatalf("expected no reusable env for nil closure")
	}

	mutatingDecl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ID("T"), ast.Int(1)),
			ast.ID("T_type"),
		},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	if env, ok := interp.reusableBytecodeCallEnvForExplicitBindings(
		&runtime.FunctionValue{Declaration: mutatingDecl, Closure: closure},
		mutatingDecl,
		ast.NewFunctionCall(ast.ID("id"), nil, []ast.TypeExpression{ast.Ty("i32")}, false),
		&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}},
	); ok || env != nil {
		t.Fatalf("expected no reusable env when body mutates explicit type bindings")
	}
}

func TestInvokeFunctionReusesCachedBytecodeCallEnvForGenericSlotFunction(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		[]ast.Statement{ast.ID("T_type")},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	program, err := interp.lowerFunctionDefinitionBytecode(decl)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	fn := &runtime.FunctionValue{Declaration: decl, Closure: closure}
	setFunctionBytecodeProgram(fn, program)
	call := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.Int(41)}, []ast.TypeExpression{ast.Ty("i32")}, false)
	args := []runtime.Value{runtime.NewSmallInt(41, runtime.IntegerI32)}

	first, err := interp.callResolvedFunctionValue(fn, fn, args, closure, call, false)
	if err != nil {
		t.Fatalf("first generic call failed: %v", err)
	}
	second, err := interp.callResolvedFunctionValue(fn, fn, args, closure, call, false)
	if err != nil {
		t.Fatalf("second generic call failed: %v", err)
	}
	if !valuesEqual(first, second) {
		t.Fatalf("generic call results differ: first=%#v second=%#v", first, second)
	}
	if got := len(interp.reusableBytecodeCallEnvCache); got != 1 {
		t.Fatalf("reusable env cache size = %d, want 1", got)
	}
}

func TestReusableBytecodeCallEnvForCallLocalBindingsCachesEnvironment(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	boxValue := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: boxDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
	}
	decl := ast.Fn(
		"box_value",
		[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Box"), ast.Ty("T")))},
		[]ast.Statement{ast.ID("self")},
		ast.Gen(ast.Ty("Box"), ast.Ty("T")),
		nil,
		nil,
		false,
		false,
	)
	fn := &runtime.FunctionValue{
		Declaration: decl,
		MethodSet: &runtime.MethodSet{
			TargetType:    ast.Gen(ast.Ty("Box"), ast.Ty("T")),
			GenericParams: []*ast.GenericParameter{ast.GenericParam("T")},
		},
		Closure: closure,
	}
	call := ast.NewFunctionCall(ast.ID("box_value"), []ast.Expression{ast.ID("box")}, nil, false)
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}

	first, ok := interp.reusableBytecodeCallEnvForBindings(fn, decl, call, boxValue, true, program)
	if !ok || first == nil {
		t.Fatalf("expected reusable bytecode env")
	}
	second, ok := interp.reusableBytecodeCallEnvForBindings(fn, decl, call, boxValue, true, program)
	if !ok || second == nil {
		t.Fatalf("expected cached reusable bytecode env")
	}
	if first != second {
		t.Fatalf("expected cached env reuse, got %p and %p", first, second)
	}
	if got, ok := first.Lookup("T"); !ok {
		t.Fatalf("expected cached env to define T")
	} else if ref, ok := got.(runtime.TypeRefValue); !ok || ref.TypeName != "i32" {
		t.Fatalf("cached env T = %#v, want TypeRefValue{i32}", got)
	}
	if got, ok := first.Lookup("Self_type"); !ok {
		t.Fatalf("expected cached env to define Self_type")
	} else if str, ok := got.(runtime.StringValue); !ok || str.Val != "Box<i32>" {
		t.Fatalf("cached env Self_type = %#v, want StringValue{Box<i32>}", got)
	}
	if got := len(interp.reusableBytecodeCallEnvCache); got != 1 {
		t.Fatalf("reusable env cache size = %d, want 1", got)
	}
}

func TestReusableBytecodeCallEnvForCallLocalBindingsReusesAcrossEquivalentCallSites(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	boxValue := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: boxDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
	}
	decl := ast.Fn(
		"box_value",
		[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Box"), ast.Ty("T")))},
		[]ast.Statement{ast.ID("self")},
		ast.Gen(ast.Ty("Box"), ast.Ty("T")),
		nil,
		nil,
		false,
		false,
	)
	fn := &runtime.FunctionValue{
		Declaration: decl,
		MethodSet: &runtime.MethodSet{
			TargetType:    ast.Gen(ast.Ty("Box"), ast.Ty("T")),
			GenericParams: []*ast.GenericParameter{ast.GenericParam("T")},
		},
		Closure: closure,
	}
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}
	firstCall := ast.NewFunctionCall(ast.ID("box_value"), []ast.Expression{ast.ID("left")}, nil, false)
	secondCall := ast.NewFunctionCall(ast.ID("box_value"), []ast.Expression{ast.ID("right")}, nil, false)

	first, ok := interp.reusableBytecodeCallEnvForBindings(fn, decl, firstCall, boxValue, true, program)
	if !ok || first == nil {
		t.Fatalf("expected reusable bytecode env for first call-local call site")
	}
	second, ok := interp.reusableBytecodeCallEnvForBindings(fn, decl, secondCall, boxValue, true, program)
	if !ok || second == nil {
		t.Fatalf("expected reusable bytecode env for second call-local call site")
	}
	if first != second {
		t.Fatalf("expected equivalent call-local call sites to share env, got %p and %p", first, second)
	}
	if got := len(interp.reusableBytecodeCallEnvCache); got != 1 {
		t.Fatalf("reusable env cache size = %d, want 1", got)
	}
}

func TestReusableBytecodeCallEnvForBindingsSeparatesDifferentExplicitBindingSets(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	boxValue := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: boxDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
	}
	decl := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Box"), ast.Ty("T"))),
			ast.Param("value", ast.Ty("U")),
		},
		[]ast.Statement{ast.ID("Self_type"), ast.ID("U_type")},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("U")},
		nil,
		false,
		false,
	)
	fn := &runtime.FunctionValue{
		Declaration: decl,
		MethodSet: &runtime.MethodSet{
			TargetType:    ast.Gen(ast.Ty("Box"), ast.Ty("T")),
			GenericParams: []*ast.GenericParameter{ast.GenericParam("T")},
		},
		Closure: closure,
	}
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}
	i32Call := ast.NewFunctionCall(ast.ID("apply"), []ast.Expression{ast.ID("left"), ast.Int(1)}, []ast.TypeExpression{ast.Ty("i32")}, false)
	i32CallAgain := ast.NewFunctionCall(ast.ID("apply"), []ast.Expression{ast.ID("right"), ast.Int(2)}, []ast.TypeExpression{ast.Ty("i32")}, false)
	stringCall := ast.NewFunctionCall(ast.ID("apply"), []ast.Expression{ast.ID("left"), ast.Int(3)}, []ast.TypeExpression{ast.Ty("String")}, false)

	first, ok := interp.reusableBytecodeCallEnvForBindings(fn, decl, i32Call, boxValue, true, program)
	if !ok || first == nil {
		t.Fatalf("expected reusable bytecode env for first i32 binding set")
	}
	second, ok := interp.reusableBytecodeCallEnvForBindings(fn, decl, i32CallAgain, boxValue, true, program)
	if !ok || second == nil {
		t.Fatalf("expected reusable bytecode env for second i32 binding set")
	}
	if first != second {
		t.Fatalf("expected repeated i32 binding set to share env, got %p and %p", first, second)
	}
	third, ok := interp.reusableBytecodeCallEnvForBindings(fn, decl, stringCall, boxValue, true, program)
	if !ok || third == nil {
		t.Fatalf("expected reusable bytecode env for string binding set")
	}
	if third == first {
		t.Fatalf("expected different explicit binding sets to keep separate envs")
	}
	if got := len(interp.reusableBytecodeCallEnvCache); got != 2 {
		t.Fatalf("reusable env cache size = %d, want 2", got)
	}
}

func TestInvokeFunctionReusesCachedBytecodeCallEnvForMethodSetSlotFunction(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	boxValue := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: boxDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(11, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
	}
	decl := ast.Fn(
		"box_value",
		[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Box"), ast.Ty("T")))},
		[]ast.Statement{ast.ID("Self_type")},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	program, err := interp.lowerFunctionDefinitionBytecode(decl)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	fn := &runtime.FunctionValue{
		Declaration: decl,
		MethodSet: &runtime.MethodSet{
			TargetType:    ast.Gen(ast.Ty("Box"), ast.Ty("T")),
			GenericParams: []*ast.GenericParameter{ast.GenericParam("T")},
		},
		Closure: closure,
	}
	setFunctionBytecodeProgram(fn, program)
	call := ast.NewFunctionCall(ast.ID("box_value"), []ast.Expression{ast.ID("box")}, nil, false)
	args := []runtime.Value{boxValue}

	first, err := interp.callResolvedFunctionValue(fn, fn, args, closure, call, false)
	if err != nil {
		t.Fatalf("first method-set call failed: %v", err)
	}
	second, err := interp.callResolvedFunctionValue(fn, fn, args, closure, call, false)
	if err != nil {
		t.Fatalf("second method-set call failed: %v", err)
	}
	if !valuesEqual(first, second) {
		t.Fatalf("method-set call results differ: first=%#v second=%#v", first, second)
	}
	if got := len(interp.callLocalTypeBindingCache); got != 1 {
		t.Fatalf("call-local type binding cache size = %d, want 1", got)
	}
	if got := len(interp.reusableBytecodeCallEnvCache); got != 1 {
		t.Fatalf("reusable env cache size = %d, want 1", got)
	}
}

func TestReusableBytecodeCallEnvForBindingsHotHitAvoidsAllocations(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("U"), "value")},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("U")},
		nil,
		false,
	)
	boxValue := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: boxDef},
		Fields: map[string]runtime.Value{
			"value": runtime.StringValue{Val: "hi"},
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("String")},
	}
	decl := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Box"), ast.Ty("U"))),
			ast.Param("value", ast.Ty("T")),
		},
		[]ast.Statement{ast.ID("value")},
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	fn := &runtime.FunctionValue{
		Declaration: decl,
		MethodSet: &runtime.MethodSet{
			TargetType:    ast.Gen(ast.Ty("Box"), ast.Ty("U")),
			GenericParams: []*ast.GenericParameter{ast.GenericParam("U")},
		},
		Closure: closure,
	}
	call := ast.NewFunctionCall(ast.ID("apply"), []ast.Expression{ast.ID("box"), ast.Int(1)}, []ast.TypeExpression{ast.Ty("i32")}, false)
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}

	first, ok := interp.reusableBytecodeCallEnvForBindings(fn, decl, call, boxValue, true, program)
	if !ok || first == nil {
		t.Fatalf("expected reusable bytecode env")
	}

	allocs := testing.AllocsPerRun(1000, func() {
		env, ok := interp.reusableBytecodeCallEnvForBindings(fn, decl, call, boxValue, true, program)
		if !ok || env != first {
			panic("expected cached reusable bytecode env")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected hot reusable bytecode env hit to avoid allocations, got %.2f", allocs)
	}
}

func TestReusableBytecodeCallEnvForBindingsHotHitMemoizesInferredReceiverTypeArgs(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("U"), "value")},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("U")},
		nil,
		false,
	)
	boxValue := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: boxDef},
		Fields: map[string]runtime.Value{
			"value": runtime.StringValue{Val: "hi"},
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("U")},
	}
	decl := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Box"), ast.Ty("U"))),
			ast.Param("value", ast.Ty("T")),
		},
		[]ast.Statement{ast.ID("value")},
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	fn := &runtime.FunctionValue{
		Declaration: decl,
		MethodSet: &runtime.MethodSet{
			TargetType:    ast.Gen(ast.Ty("Box"), ast.Ty("U")),
			GenericParams: []*ast.GenericParameter{ast.GenericParam("U")},
		},
		Closure: closure,
	}
	call := ast.NewFunctionCall(ast.ID("apply"), []ast.Expression{ast.ID("box"), ast.Int(1)}, []ast.TypeExpression{ast.Ty("i32")}, false)
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}

	first, ok := interp.reusableBytecodeCallEnvForBindings(fn, decl, call, boxValue, true, program)
	if !ok || first == nil {
		t.Fatalf("expected reusable bytecode env")
	}
	if len(boxValue.TypeArguments) != 1 || typeExpressionToString(boxValue.TypeArguments[0]) != "String" {
		t.Fatalf("expected receiver type args to memoize inferred result, got %#v", boxValue.TypeArguments)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		env, ok := interp.reusableBytecodeCallEnvForBindings(fn, decl, call, boxValue, true, program)
		if !ok || env != first {
			panic("expected cached reusable bytecode env")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected hot inferred-receiver reusable env hit to avoid allocations, got %.2f", allocs)
	}
}

func TestInvokeFunctionColdPathSeedsExplicitAndCallLocalBindings(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	boxValue := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: boxDef},
		Fields: map[string]runtime.Value{
			"value": runtime.StringValue{Val: "hi"},
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("String")},
	}
	decl := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Box"), ast.Ty("T"))),
			ast.Param("value", ast.Ty("U")),
		},
		[]ast.Statement{ast.ID("T_type"), ast.ID("U_type"), ast.ID("Self_type")},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("U")},
		nil,
		false,
		false,
	)
	fn := &runtime.FunctionValue{
		Declaration: decl,
		MethodSet: &runtime.MethodSet{
			TargetType:    ast.Gen(ast.Ty("Box"), ast.Ty("T")),
			GenericParams: []*ast.GenericParameter{ast.GenericParam("T")},
		},
		Closure: closure,
		Bytecode: CompiledThunk(func(localEnv *runtime.Environment, args []runtime.Value) (runtime.Value, error) {
			if got, ok := localEnv.Lookup("T_type"); !ok {
				t.Fatalf("expected T_type binding in cold-path env")
			} else if str, ok := got.(runtime.StringValue); !ok || str.Val != "String" {
				t.Fatalf("T_type = %#v, want String", got)
			}
			if got, ok := localEnv.Lookup("U_type"); !ok {
				t.Fatalf("expected U_type binding in cold-path env")
			} else if str, ok := got.(runtime.StringValue); !ok || str.Val != "i32" {
				t.Fatalf("U_type = %#v, want i32", got)
			}
			if got, ok := localEnv.Lookup("Self_type"); !ok {
				t.Fatalf("expected Self_type binding in cold-path env")
			} else if str, ok := got.(runtime.StringValue); !ok || str.Val != "Box<String>" {
				t.Fatalf("Self_type = %#v, want Box<String>", got)
			}
			return runtime.StringValue{Val: "ok"}, nil
		}),
	}
	call := ast.NewFunctionCall(ast.ID("apply"), []ast.Expression{ast.ID("box"), ast.Int(1)}, []ast.TypeExpression{ast.Ty("i32")}, false)
	result, err := interp.callResolvedFunctionValue(
		fn,
		fn,
		[]runtime.Value{boxValue, runtime.NewSmallInt(1, runtime.IntegerI32)},
		closure,
		call,
		false,
	)
	if err != nil {
		t.Fatalf("cold-path invoke failed: %v", err)
	}
	if got, ok := result.(runtime.StringValue); !ok || got.Val != "ok" {
		t.Fatalf("cold-path invoke result = %#v, want ok", result)
	}
	if got := len(interp.reusableBytecodeCallEnvCache); got != 0 {
		t.Fatalf("unexpected reusable env cache entries for cold thunk path: %d", got)
	}
	if got := len(interp.explicitCallTypeBindingCache); got != 1 {
		t.Fatalf("explicit type binding cache size = %d, want 1", got)
	}
	if got := len(interp.callLocalTypeBindingCache); got != 1 {
		t.Fatalf("call-local type binding cache size = %d, want 1", got)
	}
}

func TestFunctionCallTypeBindingSetEnvValueCapacityAvoidsOverlappingRuntimeTypeBindings(t *testing.T) {
	bindings := functionCallTypeBindingSet{
		explicit: []runtime.EnvironmentBinding{
			{Name: "T_type", Value: runtime.StringValue{Val: "i32"}},
			{Name: "T", Value: runtime.TypeRefValue{TypeName: "i32"}},
		},
		callLocal: []runtime.EnvironmentBinding{
			{Name: "T_type", Value: runtime.StringValue{Val: "i32"}},
			{Name: "T", Value: runtime.TypeRefValue{TypeName: "i32"}},
			{Name: "Self_type", Value: runtime.StringValue{Val: "Box<i32>"}},
			{Name: "Self", Value: runtime.TypeRefValue{TypeName: "Box"}},
		},
	}

	if got := bindings.distinctLen(); got != 4 {
		t.Fatalf("distinctLen() = %d, want 4", got)
	}
	if got := bindings.envValueCapacity(2); got != 4 {
		t.Fatalf("envValueCapacity(base=2) = %d, want 4", got)
	}
}
