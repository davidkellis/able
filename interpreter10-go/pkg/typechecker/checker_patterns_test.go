package typechecker

import (
	"able/interpreter10-go/pkg/ast"
	"strings"
	"testing"
)

func TestWildcardPatternIgnoresValue(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.Wc(), ast.Int(5))
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}
func TestTypedPatternMismatchProducesDiagnosticInMatch(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(ast.ID("value"), ast.Ty("string")),
		ast.Int(1),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected typed pattern diagnostic")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "pattern expected type") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing typed pattern diagnostic: %v", diags)
	}
}
func TestTypedArrayPatternMismatchProducesDiagnostic(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(ast.ID("values"), ast.Gen(ast.Ty("Array"), ast.Ty("string"))),
		ast.Arr(ast.Int(1)),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "Array[String]") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected array pattern mismatch diagnostic, got %v", diags)
	}
}
func TestTypedStructPatternMatchesInstance(t *testing.T) {
	checker := New()
	structDef := ast.StructDef(
		"Pair",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "head"),
			ast.FieldDef(ast.Ty("i32"), "tail"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	assign := ast.Assign(
		ast.TypedP(
			ast.StructP(
				[]*ast.StructPatternField{
					ast.FieldP(ast.ID("head"), "head", nil),
					ast.FieldP(ast.ID("tail"), "tail", nil),
				},
				false,
				"Pair",
			),
			ast.Ty("Pair"),
		),
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(1), "head"),
				ast.FieldInit(ast.Int(2), "tail"),
			},
			false,
			"Pair",
			nil,
			nil,
		),
	)
	module := ast.NewModule([]ast.Statement{structDef, assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for matching struct typed pattern, got %v", diags)
	}
}
func TestTypedArrayPatternMatchesElementType(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(
			ast.ID("values"),
			ast.Gen(ast.Ty("Array"), ast.Ty("i32")),
		),
		ast.Arr(ast.Int(1), ast.Int(2)),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for matching array typed pattern, got %v", diags)
	}
}
func TestTypedPatternProvidesAnnotationTypeWhenSubjectUnknown(t *testing.T) {
	checker := New()
	valueUse := ast.ID("value")
	clause := ast.Mc(
		ast.TypedP(ast.ID("value"), ast.Ty("string")),
		valueUse,
	)
	matchExpr := ast.Match(ast.ID("subject"), clause)
	fn := ast.Fn(
		"demo",
		[]*ast.FunctionParameter{
			ast.Param("subject", nil),
		},
		[]ast.Statement{
			ast.Ret(matchExpr),
		},
		ast.Ty("string"),
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
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[valueUse]
	if !ok {
		t.Fatalf("expected inference entry for value identifier")
	}
	if typ.Name() != "String" {
		t.Fatalf("expected value to infer String type, got %q", typ.Name())
	}
	matchType, ok := checker.infer[matchExpr]
	if !ok {
		t.Fatalf("expected inference entry for match expression")
	}
	if matchType.Name() != "String" {
		t.Fatalf("expected match expression to infer String, got %q", matchType.Name())
	}
}
func TestTypedPatternMismatchProducesDiagnostic(t *testing.T) {
	checker := New()
	clause := ast.Mc(
		ast.TypedP(ast.ID("value"), ast.Ty("string")),
		ast.ID("value"),
	)
	matchExpr := ast.Match(ast.ID("subject"), clause)
	fn := ast.Fn(
		"demo",
		[]*ast.FunctionParameter{
			ast.Param("subject", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(matchExpr),
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
		t.Fatalf("expected diagnostic for typed pattern mismatch")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "pattern expected type String") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected typed pattern mismatch diagnostic, got %v", diags)
	}
}
func TestLiteralPatternMatchesSubjectType(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("value"), ast.Int(42))
	match := ast.Match(
		ast.ID("value"),
		ast.Mc(ast.LitP(ast.Int(42)), ast.Int(1)),
	)
	module := ast.NewModule([]ast.Statement{assign, match}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for matching literal pattern, got %v", diags)
	}
	clause := match.Clauses[0]
	literalPat, ok := clause.Pattern.(*ast.LiteralPattern)
	if !ok {
		t.Fatalf("expected literal pattern in clause")
	}
	inferred, ok := checker.infer[literalPat]
	if !ok {
		t.Fatalf("expected inference entry for literal pattern")
	}
	if inferred.Name() != "Int:i32" {
		t.Fatalf("expected literal pattern to infer Int:i32, got %q", inferred.Name())
	}
}
func TestLiteralPatternMismatchAllowed(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("value"), ast.Str("hello"))
	match := ast.Match(
		ast.ID("value"),
		ast.Mc(ast.LitP(ast.Int(1)), ast.Str("fallback")),
	)
	module := ast.NewModule([]ast.Statement{assign, match}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for literal pattern mismatch, got %v", diags)
	}
}
func TestMatchStructPatternBindsNestedFields(t *testing.T) {
	checker := New()
	inner := ast.StructDef(
		"Inner",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	outer := ast.StructDef(
		"Outer",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("Inner"), "inner"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	assign := ast.Assign(
		ast.ID("outer"),
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(
					ast.StructLit(
						[]*ast.StructFieldInitializer{
							ast.FieldInit(ast.Int(3), "value"),
						},
						false,
						"Inner",
						nil,
						nil,
					),
					"inner",
				),
			},
			false,
			"Outer",
			nil,
			nil,
		),
	)
	matchClause := ast.Mc(
		ast.StructP(
			[]*ast.StructPatternField{
				ast.FieldP(
					ast.StructP(
						[]*ast.StructPatternField{
							ast.FieldP(ast.ID("val"), "value", nil),
						},
						false,
						"Inner",
					),
					"inner",
					nil,
				),
			},
			false,
			"Outer",
		),
		ast.ID("val"),
	)
	matchExpr := ast.Match(ast.ID("outer"), matchClause)
	module := ast.NewModule([]ast.Statement{inner, outer, assign, matchExpr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for nested struct pattern, got %v", diags)
	}
	typ, ok := checker.infer[matchExpr]
	if !ok {
		t.Fatalf("expected inference entry for match expression")
	}
	if typ == nil || typ.Name() != "Int:i32" {
		t.Fatalf("expected match expression to infer Int:i32, got %#v", typ)
	}
}
func TestMatchArrayPatternInfersRestType(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.ID("items"),
		ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3)),
	)
	tailID := ast.ID("tail")
	matchClause := ast.Mc(
		ast.ArrP([]ast.Pattern{ast.ID("head")}, tailID),
		ast.ID("tail"),
	)
	matchExpr := ast.Match(ast.ID("items"), matchClause)
	module := ast.NewModule([]ast.Statement{assign, matchExpr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for array rest pattern, got %v", diags)
	}
	inferred, ok := checker.infer[tailID]
	if !ok {
		t.Fatalf("expected inference entry for tail identifier")
	}
	arrType, ok := inferred.(ArrayType)
	if !ok {
		t.Fatalf("expected tail to have ArrayType, got %#v", inferred)
	}
	if arrType.Element == nil || arrType.Element.Name() != "Int:i32" {
		t.Fatalf("expected tail element type Int:i32, got %#v", arrType.Element)
	}
}
func TestForLoopPatternBindsIdentifier(t *testing.T) {
	checker := New()
	structDef := ast.StructDef(
		"Pair",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "head"),
			ast.FieldDef(ast.Ty("i32"), "tail"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	items := ast.Assign(
		ast.ID("inbox"),
		ast.Arr(
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "head"),
					ast.FieldInit(ast.Int(2), "tail"),
				},
				false,
				"Pair",
				nil,
				nil,
			),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(3), "head"),
					ast.FieldInit(ast.Int(4), "tail"),
				},
				false,
				"Pair",
				nil,
				nil,
			),
		),
	)
	headExpr := ast.ID("head")
	loop := ast.ForIn(
		ast.StructP(
			[]*ast.StructPatternField{
				ast.FieldP(ast.ID("head"), "head", nil),
				ast.FieldP(ast.ID("tail"), "tail", nil),
			},
			false,
			"Pair",
		),
		ast.ID("inbox"),
		headExpr,
	)
	module := ast.NewModule([]ast.Statement{structDef, items, loop}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[headExpr]
	if !ok {
		t.Fatalf("expected inference for loop body identifier")
	}
	if typ.Name() != "Int:i32" {
		t.Fatalf("expected head to have type Int:i32, got %q", typ.Name())
	}
}
