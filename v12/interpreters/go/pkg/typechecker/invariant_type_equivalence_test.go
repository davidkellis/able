package typechecker

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestNestedTypeArgumentsAreInvariant(t *testing.T) {
	i8 := IntegerType{Suffix: "i8"}
	i32 := IntegerType{Suffix: "i32"}
	box := StructType{StructName: "Box"}

	tests := []struct {
		name     string
		actual   Type
		expected Type
	}{
		{"array", ArrayType{Element: i8}, ArrayType{Element: i32}},
		{"map key", MapType{Key: i8, Value: i32}, MapType{Key: i32, Value: i32}},
		{"range", RangeType{Element: i8}, RangeType{Element: i32}},
		{"iterator", IteratorType{Element: i8}, IteratorType{Element: i32}},
		{"future", FutureType{Result: i8}, FutureType{Result: i32}},
		{"nullable", NullableType{Inner: i8}, NullableType{Inner: i32}},
		{
			"user generic",
			AppliedType{Base: box, Arguments: []Type{i8}},
			AppliedType{Base: box, Arguments: []Type{i32}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if typeAssignable(tc.actual, tc.expected) {
				t.Fatalf("%s unexpectedly assignable to %s", typeName(tc.actual), typeName(tc.expected))
			}
			if got := assignabilityDiagnosticCode(tc.actual, tc.expected); got != DiagnosticCodeInvariantTypeArgument {
				t.Fatalf("diagnostic code = %q, want %q", got, DiagnosticCodeInvariantTypeArgument)
			}
		})
	}
}

func TestTopLevelIntegerWideningRemainsAssignable(t *testing.T) {
	if !typeAssignable(IntegerType{Suffix: "i8"}, IntegerType{Suffix: "i32"}) {
		t.Fatal("top-level i8 to i32 widening must remain assignable")
	}
}

func TestCallableTypesRequireCompleteEquivalentSignatures(t *testing.T) {
	i8 := IntegerType{Suffix: "i8"}
	i32 := IntegerType{Suffix: "i32"}
	display := InterfaceType{InterfaceName: "Display"}
	exact := FunctionType{Params: []Type{i8}, Return: i8}

	tests := []FunctionType{
		{Params: []Type{i32}, Return: i8},
		{Params: []Type{i8}, Return: i32},
		{Params: []Type{i8, i8}, Return: i8},
		{
			Params:     []Type{TypeParameterType{ParameterName: "T"}},
			Return:     TypeParameterType{ParameterName: "T"},
			TypeParams: []GenericParamSpec{{Name: "T", Constraints: []Type{display}}},
		},
	}
	for _, mismatch := range tests {
		if typeAssignable(exact, mismatch) {
			t.Fatalf("callable %s unexpectedly assignable to %s", formatType(exact), formatType(mismatch))
		}
		if got := assignabilityDiagnosticCode(exact, mismatch); got != DiagnosticCodeCallableSignatureMismatch {
			t.Fatalf("diagnostic code = %q, want %q", got, DiagnosticCodeCallableSignatureMismatch)
		}
	}
	if !typeAssignable(exact, FunctionType{Params: []Type{i8}, Return: i8}) {
		t.Fatal("equivalent callable signatures must remain assignable")
	}
	if !typeAssignable(exact, NullableType{Inner: exact}) {
		t.Fatal("equivalent callable must lift into a nullable callable")
	}
	if typeAssignable(exact, NullableType{Inner: FunctionType{Params: []Type{i32}, Return: i8}}) {
		t.Fatal("nullable wrapping must not weaken callable signature invariance")
	}
}

func TestGenericCallableEquivalenceIsAlphaRenamingSafe(t *testing.T) {
	display := InterfaceType{InterfaceName: "Display"}
	actual := FunctionType{
		Params:     []Type{TypeParameterType{ParameterName: "T"}},
		Return:     TypeParameterType{ParameterName: "T"},
		TypeParams: []GenericParamSpec{{Name: "T", Constraints: []Type{display}}},
	}
	expected := FunctionType{
		Params:     []Type{TypeParameterType{ParameterName: "U"}},
		Return:     TypeParameterType{ParameterName: "U"},
		TypeParams: []GenericParamSpec{{Name: "U", Constraints: []Type{display}}},
	}
	if !typeAssignable(actual, expected) {
		t.Fatal("alpha-renamed equivalent generic callable signatures must be assignable")
	}
}

func TestInferredTypeArgumentsCanonicalizeAliases(t *testing.T) {
	concrete := StructType{StructName: "Less"}
	alias := AliasType{AliasName: "OrderingLess", Target: concrete}
	subst := map[string]Type{"T": concrete}

	if diags := bindTypeParameter("T", alias, subst, nil, 0); len(diags) != 0 {
		t.Fatalf("alias for inferred concrete type must remain equivalent: %v", diags)
	}
}

func TestRepeatedInferredCallsRetainCanonicalTypes(t *testing.T) {
	checker := New()
	param := TypeParameterType{ParameterName: "T"}
	identity := FunctionType{
		Params:     []Type{param},
		Return:     param,
		TypeParams: []GenericParamSpec{{Name: "T"}},
	}
	arrayBase := StructType{
		StructName: "Array",
		TypeParams: []GenericParamSpec{{Name: "Element"}},
	}
	tests := []struct {
		name   string
		actual Type
	}{
		{"interface", InterfaceType{InterfaceName: "Framework"}},
		{"union", UnionType{UnionName: "Result", Variants: []Type{IntegerType{Suffix: "i32"}}}},
		{
			"generic",
			AppliedType{
				Base:      arrayBase,
				Arguments: []Type{PrimitiveType{Kind: PrimitiveString}},
			},
		},
		{
			"nullable",
			NullableType{Inner: InterfaceType{InterfaceName: "Framework"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			call := ast.Call("identity", ast.ID("value"))
			if _, diags := checker.instantiateFunctionCall(identity, call, []Type{tc.actual}, nil); len(diags) != 0 {
				t.Fatalf("first inference failed: %v", diags)
			}
			if len(call.TypeArguments) != 1 {
				t.Fatalf("inferred type argument count = %d, want 1", len(call.TypeArguments))
			}
			instantiated, diags := checker.instantiateFunctionCall(identity, call, []Type{tc.actual}, nil)
			if len(diags) != 0 {
				t.Fatalf("repeated inference lost canonical type: %v", diags)
			}
			if !invariantTypeEquivalent(instantiated.Params[0], tc.actual) {
				t.Fatalf("repeated inference resolved %s as %s", formatType(tc.actual), formatType(instantiated.Params[0]))
			}
		})
	}
}

func TestFunctionCallReportsInvariantTypeArgumentCode(t *testing.T) {
	takes := ast.Fn(
		"takes",
		[]*ast.FunctionParameter{ast.Param("values", ast.Gen(ast.Ty("Array"), ast.Ty("i32")))},
		[]ast.Statement{ast.Nil()},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	values := ast.Assign(
		ast.TypedP(ast.ID("values"), ast.Gen(ast.Ty("Array"), ast.Ty("i8"))),
		ast.Arr(ast.Int(7)),
	)
	call := ast.Call("takes", ast.ID("values"))
	diags, err := New().CheckModule(ast.NewModule([]ast.Statement{takes, values, call}, nil, nil))
	if err != nil {
		t.Fatalf("unexpected checker error: %v", err)
	}
	assertDiagnosticCode(t, diags, DiagnosticCodeInvariantTypeArgument)
}

func TestFunctionCallReportsCallableSignatureMismatchCode(t *testing.T) {
	takes := ast.Fn(
		"takes",
		[]*ast.FunctionParameter{
			ast.Param("callable", ast.FnType([]ast.TypeExpression{ast.Ty("i32")}, ast.Ty("i32"))),
		},
		[]ast.Statement{ast.Nil()},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	small := ast.Fn(
		"small",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("i8"))},
		[]ast.Statement{ast.ID("value")},
		ast.Ty("i8"),
		nil,
		nil,
		false,
		false,
	)
	diags, err := New().CheckModule(ast.NewModule(
		[]ast.Statement{takes, small, ast.Call("takes", ast.ID("small"))},
		nil,
		nil,
	))
	if err != nil {
		t.Fatalf("unexpected checker error: %v", err)
	}
	assertDiagnosticCode(t, diags, DiagnosticCodeCallableSignatureMismatch)
}

func TestContextualLambdaAllowsTopLevelInterfaceUpcast(t *testing.T) {
	checker := New()
	probe := StructType{StructName: "ProbeError", Fields: map[string]Type{}}
	errorInterface := InterfaceType{InterfaceName: "Error", Methods: map[string]FunctionType{}}
	checker.global.Define("ProbeError", probe)
	checker.implementations = []ImplementationSpec{{
		InterfaceName: "Error",
		Interface:     errorInterface,
		Target:        probe,
	}}

	lambda := ast.Lam(
		nil,
		ast.StructLit(nil, false, "ProbeError", nil, nil),
	)
	expected := FunctionType{Return: errorInterface}
	diagnostics, inferred := checker.checkLambdaExpressionWithExpectedType(
		checker.global,
		lambda,
		&expected,
	)
	if len(diagnostics) != 0 {
		t.Fatalf("top-level lambda result upcast produced diagnostics: %v", diagnostics)
	}
	if !invariantTypeEquivalent(inferred, expected) {
		t.Fatalf("inferred lambda = %s, want %s", formatType(inferred), formatType(expected))
	}
}

func assertDiagnosticCode(t *testing.T, diagnostics []Diagnostic, want DiagnosticCode) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == want {
			return
		}
	}
	t.Fatalf("expected diagnostic code %q, got %v", want, diagnostics)
}
