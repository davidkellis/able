package typechecker

import (
	"able/interpreter10-go/pkg/ast"
	"strings"
	"testing"
)

func TestFunctionCallInferredReturnType(t *testing.T) {
	checker := New()
	addFn := ast.Fn(
		"add",
		[]*ast.FunctionParameter{
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Block(ast.ID("a")), // body ignored by checker today
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	call := ast.Call("add", ast.Int(1), ast.Int(2))
	module := ast.NewModule([]ast.Statement{addFn, call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[call]
	if !ok {
		t.Fatalf("expected inferred type for call")
	}
	if typName := typeName(typ); typName != "i32" {
		t.Fatalf("expected call to have type i32, got %q", typName)
	}
}
func TestFunctionCallArgumentCountMismatch(t *testing.T) {
	checker := New()
	addFn := ast.Fn(
		"add",
		[]*ast.FunctionParameter{
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Block(ast.ID("a")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	call := ast.Call("add", ast.Int(1))
	module := ast.NewModule([]ast.Statement{addFn, call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected one diagnostic, got %v", diags)
	}
	if want := "expects 2 arguments"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic to mention %q, got %q", want, diags[0].Message)
	}
}

func TestGenericFunctionCallInfersTypeArguments(t *testing.T) {
	checker := New()
	genericParam := ast.GenericParam("T")
	fn := ast.Fn(
		"identity",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.Ret(ast.ID("value")),
		},
		ast.Ty("T"),
		[]*ast.GenericParameter{genericParam},
		nil,
		false,
		false,
	)
	call := ast.Call("identity", ast.Int(42))
	module := ast.NewModule([]ast.Statement{fn, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[call]
	if !ok {
		t.Fatalf("expected inferred type for call expression")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected identity call to infer i32, got %q", typeName(typ))
	}
}

func TestGenericFunctionCallConflictingInferenceDiagnostic(t *testing.T) {
	checker := New()
	genericParam := ast.GenericParam("T")
	fn := ast.Fn(
		"choose",
		[]*ast.FunctionParameter{
			ast.Param("first", ast.Ty("T")),
			ast.Param("second", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.Block(ast.ID("first")),
		},
		ast.Ty("T"),
		[]*ast.GenericParameter{genericParam},
		nil,
		false,
		false,
	)
	call := ast.Call("choose", ast.Int(1), ast.Str("oops"))
	module := ast.NewModule([]ast.Statement{fn, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for conflicting generic inference")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "type parameter T") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected type parameter diagnostic, got %v", diags)
	}
}
func TestGenericFunctionCallInfersStructLiteralArgument(t *testing.T) {
	checker := New()
	wrapperDef := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	genericParam := ast.GenericParam("T")
	fn := ast.Fn(
		"unwrap",
		[]*ast.FunctionParameter{
			ast.Param("input", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T"))),
		},
		[]ast.Statement{
			ast.Block(ast.Member(ast.ID("input"), "value")),
		},
		ast.Ty("T"),
		[]*ast.GenericParameter{genericParam},
		nil,
		false,
		false,
	)
	call := ast.Call(
		"unwrap",
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(42), "value"),
			},
			false,
			"Wrapper",
			nil,
			nil,
		),
	)
	module := ast.NewModule([]ast.Statement{wrapperDef, fn, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[call]
	if !ok {
		t.Fatalf("expected inferred type for call expression")
	}
	intType, ok := typ.(IntegerType)
	if !ok {
		t.Fatalf("expected call to infer IntegerType, got %T", typ)
	}
	if intType.Suffix != "i32" {
		t.Fatalf("expected inferred integer type i32, got %s", intType.Suffix)
	}
}
func TestStructLiteralTypeArgumentsResolveInGenericFunction(t *testing.T) {
	checker := New()
	wrapperDef := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	genericParam := ast.GenericParam("T")
	fn := ast.Fn(
		"wrap",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.Block(ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.ID("value"), "value"),
				},
				false,
				"Wrapper",
				nil,
				[]ast.TypeExpression{ast.Ty("T")},
			)),
		},
		ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
		[]*ast.GenericParameter{genericParam},
		nil,
		false,
		false,
	)
	call := ast.Call("wrap", ast.Int(10))
	module := ast.NewModule([]ast.Statement{wrapperDef, fn, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[call]
	if !ok {
		t.Fatalf("expected inferred type for call expression")
	}
	applied, ok := typ.(AppliedType)
	if !ok {
		t.Fatalf("expected call to infer applied struct type, got %T", typ)
	}
	name, ok := structName(applied.Base)
	if !ok || name != "Wrapper" {
		t.Fatalf("expected struct name Wrapper, got %q", name)
	}
	if len(applied.Arguments) != 1 {
		t.Fatalf("expected one type argument on Wrapper, got %d", len(applied.Arguments))
	}
	arg, ok := applied.Arguments[0].(IntegerType)
	if !ok || arg.Suffix != "i32" {
		t.Fatalf("expected Wrapper argument i32, got %#v", applied.Arguments[0])
	}
}
func TestNestedArrayGenericInference(t *testing.T) {
	checker := New()
	genericParam := ast.GenericParam("T")
	fn := ast.Fn(
		"flatten",
		[]*ast.FunctionParameter{
			ast.Param("values", ast.Gen(ast.Ty("Array"), ast.Gen(ast.Ty("Array"), ast.Ty("T")))),
		},
		[]ast.Statement{
			ast.Ret(ast.Index(ast.ID("values"), ast.Int(0))),
		},
		ast.Gen(ast.Ty("Array"), ast.Ty("T")),
		[]*ast.GenericParameter{genericParam},
		nil,
		false,
		false,
	)
	call := ast.Call(
		"flatten",
		ast.Arr(
			ast.Arr(ast.Int(1), ast.Int(2)),
			ast.Arr(ast.Int(3), ast.Int(4)),
		),
	)
	module := ast.NewModule([]ast.Statement{fn, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[call]
	if !ok {
		t.Fatalf("expected inferred type for call expression")
	}
	arr, ok := typ.(ArrayType)
	if !ok {
		t.Fatalf("expected flatten call to infer ArrayType, got %T", typ)
	}
	inner, ok := arr.Element.(IntegerType)
	if !ok || inner.Suffix != "i32" {
		t.Fatalf("expected inner element type i32, got %#v", arr.Element)
	}
}
func TestUnaryNegationInferredType(t *testing.T) {
	checker := New()
	expr := ast.Un(ast.UnaryOperatorNegate, ast.Int(1))
	module := ast.NewModule([]ast.Statement{expr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[expr]
	if !ok {
		t.Fatalf("expected unary expression inference entry")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected unary negation to infer i32, got %q", typeName(typ))
	}
}
func TestUnaryNotRequiresBoolDiagnostic(t *testing.T) {
	checker := New()
	expr := ast.Un(ast.UnaryOperatorNot, ast.Int(1))
	module := ast.NewModule([]ast.Statement{expr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for unary ! operand")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "unary '!'") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing unary ! diagnostic: %v", diags)
	}
}
func TestBinaryAdditionNumericInference(t *testing.T) {
	checker := New()
	expr := ast.Bin("+", ast.Int(1), ast.Flt(2))
	module := ast.NewModule([]ast.Statement{expr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[expr]
	if !ok {
		t.Fatalf("expected binary expression inference entry")
	}
	if typeName(typ) != "f64" {
		t.Fatalf("expected numeric addition to widen to f64, got %q", typeName(typ))
	}
}
func TestBinaryAdditionStringConcatenation(t *testing.T) {
	checker := New()
	expr := ast.Bin("+", ast.Str("a"), ast.Str("b"))
	module := ast.NewModule([]ast.Statement{expr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[expr]
	if !ok {
		t.Fatalf("expected binary expression inference entry")
	}
	if typeName(typ) != "string" {
		t.Fatalf("expected string concatenation to infer string, got %q", typeName(typ))
	}
}
func TestBinaryAdditionMismatchedOperandsDiagnostic(t *testing.T) {
	checker := New()
	expr := ast.Bin("+", ast.Int(1), ast.Str("oops"))
	module := ast.NewModule([]ast.Statement{expr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for mismatched operands")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "requires numeric operands") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing mismatched operands diagnostic: %v", diags)
	}
}
func TestBinaryLogicalOperandsMustBeBool(t *testing.T) {
	checker := New()
	expr := ast.Bin("&&", ast.Bool(true), ast.Int(1))
	module := ast.NewModule([]ast.Statement{expr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for logical operands")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "right operand must be bool") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing logical operand diagnostic: %v", diags)
	}
}
func TestFunctionDefinitionReturnTypeMismatch(t *testing.T) {
	checker := New()
	fn := ast.Fn(
		"answer",
		nil,
		[]ast.Statement{
			ast.Ret(ast.Str("nope")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.NewModule([]ast.Statement{fn}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for mismatched return type")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "return expects i32") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing return type diagnostic: %v", diags)
	}
}
func TestLambdaInferenceUsesBodyType(t *testing.T) {
	checker := New()
	lambda := ast.Lam(
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("i32")),
		},
		ast.ID("x"),
	)
	assign := ast.Assign(ast.ID("f"), lambda)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[lambda]
	if !ok {
		t.Fatalf("expected inference entry for lambda expression")
	}
	fnType, ok := typ.(FunctionType)
	if !ok {
		t.Fatalf("expected FunctionType, got %#v", typ)
	}
	if len(fnType.Params) != 1 {
		t.Fatalf("expected 1 lambda parameter, got %d", len(fnType.Params))
	}
	if typeName(fnType.Params[0]) != "i32" {
		t.Fatalf("expected parameter type i32, got %q", typeName(fnType.Params[0]))
	}
	if fnType.Return == nil || typeName(fnType.Return) != "i32" {
		t.Fatalf("expected lambda return type i32, got %#v", fnType.Return)
	}
}

func TestPipelineAllowsTopicAndPlaceholderCalls(t *testing.T) {
	checker := New()
	addFn := ast.Fn(
		"add",
		[]*ast.FunctionParameter{
			ast.Param("left", ast.Ty("i32")),
			ast.Param("right", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Block(ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	doubleFn := ast.Fn(
		"double",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Block(ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	valueDecl := ast.Assign(ast.ID("value"), ast.Int(21))
	topicPipe := ast.Assign(
		ast.ID("topicResult"),
		ast.Bin(
			"|>",
			ast.ID("value"),
			ast.CallExpr(ast.ID("double"), ast.TopicRef()),
		),
	)
	placeholderPipe := ast.Assign(
		ast.ID("placeholderResult"),
		ast.Bin(
			"|>",
			ast.ID("value"),
			ast.CallExpr(ast.ID("add"), ast.Placeholder(), ast.Int(5)),
		),
	)
	module := ast.NewModule([]ast.Statement{
		addFn,
		doubleFn,
		valueDecl,
		topicPipe,
		placeholderPipe,
	}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for pipeline placeholders, got %v", diags)
	}
}

func TestLambdaReturnAnnotationMismatch(t *testing.T) {
	checker := New()
	param := ast.Param("x", ast.Ty("i32"))
	lambda := ast.NewLambdaExpression(
		[]*ast.FunctionParameter{param},
		ast.ID("x"),
		ast.Ty("string"),
		nil,
		nil,
		false,
	)
	module := ast.NewModule([]ast.Statement{lambda}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for lambda return mismatch")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "lambda body returns") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing lambda return diagnostic: %v", diags)
	}
}
