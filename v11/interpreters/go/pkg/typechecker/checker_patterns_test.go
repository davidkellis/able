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
func TestTypedPatternMismatchDoesNotProduceDiagnosticInAssignment(t *testing.T) {
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
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for typed pattern mismatch, got %v", diags)
	}
}

func TestTypedPatternReportsLiteralOverflow(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(ast.ID("value"), ast.Ty("u8")),
		ast.Int(300),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected diagnostic for literal overflow, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "literal 300 does not fit in u8") {
		t.Fatalf("expected literal overflow message, got %q", diags[0].Message)
	}
}
func TestTypedArrayPatternMismatchDoesNotProduceDiagnostic(t *testing.T) {
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
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for typed array pattern mismatch, got %v", diags)
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

func TestTypedArrayPatternAdoptsIntegerLiterals(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(
			ast.ID("values"),
			ast.Gen(ast.Ty("Array"), ast.Ty("u8")),
		),
		ast.Arr(ast.Int(1), ast.Int(2)),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for adopting integer literals, got %v", diags)
	}
}

func TestTypedArrayPatternReportsLiteralOverflow(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(
			ast.ID("values"),
			ast.Gen(ast.Ty("Array"), ast.Ty("u8")),
		),
		ast.Arr(ast.Int(300)),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected diagnostic for literal overflow, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "literal 300 does not fit in u8") {
		t.Fatalf("expected literal overflow message, got %q", diags[0].Message)
	}
}

func TestTypedMapPatternAdoptsNestedLiterals(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(
			ast.ID("headers"),
			ast.Gen(
				ast.Ty("Map"),
				ast.Ty("string"),
				ast.Gen(ast.Ty("Array"), ast.Ty("u8")),
			),
		),
		ast.MapLit([]ast.MapLiteralElement{
			ast.MapEntry(ast.Str("ok"), ast.Arr(ast.Int(1), ast.Int(2))),
		}),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for nested map literals, got %v", diags)
	}
}

func TestTypedMapPatternReportsNestedLiteralOverflow(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(
			ast.ID("headers"),
			ast.Gen(
				ast.Ty("Map"),
				ast.Ty("string"),
				ast.Gen(ast.Ty("Array"), ast.Ty("u8")),
			),
		),
		ast.MapLit([]ast.MapLiteralElement{
			ast.MapEntry(ast.Str("bad"), ast.Arr(ast.Int(512))),
		}),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected diagnostic for nested literal overflow, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "literal 512 does not fit in u8") {
		t.Fatalf("expected literal overflow message, got %q", diags[0].Message)
	}
}

func TestTypedRangePatternReportsLiteralOverflow(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(
			ast.ID("window"),
			ast.Gen(ast.Ty("Range"), ast.Ty("u8")),
		),
		ast.Range(ast.Int(0), ast.Int(512), true),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected diagnostic for range literal overflow, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "literal 512 does not fit in u8") {
		t.Fatalf("expected literal overflow message, got %q", diags[0].Message)
	}
}

func TestTypedRangePatternAdoptsLiterals(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(
			ast.ID("window"),
			ast.Gen(ast.Ty("Range"), ast.Ty("u8")),
		),
		ast.Range(ast.Int(1), ast.Int(10), true),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for matching range literals, got %v", diags)
	}
}

func TestTypedIteratorPatternReportsLiteralOverflow(t *testing.T) {
	checker := New()
	iter := ast.IteratorLit(
		ast.Yield(ast.Int(512)),
	)
	assign := ast.Assign(
		ast.TypedP(
			ast.ID("iter"),
			ast.Gen(ast.Ty("Iterator"), ast.Ty("u8")),
		),
		iter,
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected diagnostic for iterator literal overflow, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "literal 512 does not fit in u8") {
		t.Fatalf("expected literal overflow message, got %q", diags[0].Message)
	}
}

func TestTypedProcPatternReportsLiteralOverflow(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(
			ast.ID("handle"),
			ast.Gen(ast.Ty("Proc"), ast.Ty("u8")),
		),
		ast.Proc(ast.Int(512)),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected diagnostic for proc literal overflow, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "literal 512 does not fit in u8") {
		t.Fatalf("expected literal overflow message, got %q", diags[0].Message)
	}
}

func TestTypedFuturePatternReportsLiteralOverflow(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.TypedP(
			ast.ID("task"),
			ast.Gen(ast.Ty("Future"), ast.Ty("u8")),
		),
		ast.Spawn(ast.Int(512)),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected diagnostic for future literal overflow, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "literal 512 does not fit in u8") {
		t.Fatalf("expected literal overflow message, got %q", diags[0].Message)
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
	if typeName(typ) != "string" {
		t.Fatalf("expected value to infer string type, got %q", typeName(typ))
	}
	matchType, ok := checker.infer[matchExpr]
	if !ok {
		t.Fatalf("expected inference entry for match expression")
	}
	if typeName(matchType) != "string" {
		t.Fatalf("expected match expression to infer string, got %q", typeName(matchType))
	}
}
func TestTypedPatternMismatchDoesNotProduceDiagnosticInMatch(t *testing.T) {
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
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for typed pattern mismatch, got %v", diags)
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
	if typeName(inferred) != "i32" {
		t.Fatalf("expected literal pattern to infer i32, got %q", typeName(inferred))
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
	if typ == nil || typeName(typ) != "i32" {
		t.Fatalf("expected match expression to infer i32, got %#v", typ)
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
	if arrType.Element == nil || typeName(arrType.Element) != "i32" {
		t.Fatalf("expected tail element type i32, got %#v", arrType.Element)
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
	if typeName(typ) != "i32" {
		t.Fatalf("expected head to have type i32, got %q", typeName(typ))
	}
}
