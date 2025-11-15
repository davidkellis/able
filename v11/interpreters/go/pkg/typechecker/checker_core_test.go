package typechecker

import (
	"able/interpreter10-go/pkg/ast"
	"strings"
	"testing"
)

func TestCheckerHandlesEmptyModule(t *testing.T) {
	checker := New()
	module := ast.NewModule(nil, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(diags))
	}
}
func TestCheckerDeclaresLetBinding(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("x"), ast.Int(42))
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if typ, ok := checker.infer[assign.Right]; !ok || typeName(typ) == "" {
		t.Fatalf("expected inferred type for literal, got %#v", typ)
	}
}
func TestCheckerUsesDefinedIdentifier(t *testing.T) {
	checker := New()
	structDef := ast.NewStructDefinition(ast.ID("Point"), nil, ast.StructKindNamed, nil, nil, false)
	assign := ast.Assign(ast.ID("x"), ast.Int(1))
	use := ast.ID("x")
	module := ast.NewModule([]ast.Statement{structDef, assign, use}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if typ, ok := checker.infer[use]; !ok || typeName(typ) == "" {
		t.Fatalf("expected inferred type for identifier, got %#v", typ)
	}
}
func TestCheckerUndefinedIdentifier(t *testing.T) {
	checker := New()
	use := ast.ID("missing")
	module := ast.NewModule([]ast.Statement{use}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %v", diags)
	}
	if want := "undefined identifier"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic containing %q, got %q", want, diags[0].Message)
	}
}
func TestCollectorReportsDuplicateDeclarations(t *testing.T) {
	checker := New()
	first := ast.StructDef("Thing", nil, ast.StructKindNamed, nil, nil, false)
	second := ast.StructDef("Thing", nil, ast.StructKindNamed, nil, nil, false)
	module := ast.NewModule([]ast.Statement{first, second}, nil, nil)
	origins := map[ast.Node]string{
		first:  "first.able",
		second: "second.able",
	}
	checker.SetNodeOrigins(origins)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if want := "duplicate declaration 'Thing'"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic containing %q, got %q", want, diags[0].Message)
	}
	if want := "first.able"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic to reference %q, got %q", want, diags[0].Message)
	}
}
func TestCollectorCapturesGenericFunctionMetadata(t *testing.T) {
	checker := New()
	displayIface := ast.Iface("Display", nil, nil, nil, nil, nil, false)
	genericParam := ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Display")))
	whereClause := []*ast.WhereClauseConstraint{
		ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Display"))),
	}
	fn := ast.Fn(
		"identity",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.ID("value"),
		},
		ast.Ty("T"),
		[]*ast.GenericParameter{genericParam},
		whereClause,
		false,
		false,
	)
	module := ast.NewModule([]ast.Statement{displayIface, fn}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.global.Lookup("identity")
	if !ok {
		t.Fatalf("expected function type recorded for identity")
	}
	fnType, ok := typ.(FunctionType)
	if !ok {
		t.Fatalf("expected FunctionType, got %#v", typ)
	}
	if len(fnType.TypeParams) != 1 {
		t.Fatalf("expected 1 type parameter, got %d", len(fnType.TypeParams))
	}
	if fnType.TypeParams[0].Name != "T" {
		t.Fatalf("expected type parameter name T, got %q", fnType.TypeParams[0].Name)
	}
	if len(fnType.TypeParams[0].Constraints) != 1 {
		t.Fatalf("expected 1 generic constraint, got %d", len(fnType.TypeParams[0].Constraints))
	}
	if constraintName := typeName(fnType.TypeParams[0].Constraints[0]); constraintName != "Display" {
		t.Fatalf("expected constraint interface Display, got %q", constraintName)
	}
	if len(fnType.Where) != 1 || fnType.Where[0].TypeParam != "T" {
		t.Fatalf("expected where clause recorded for T, got %#v", fnType.Where)
	}
	if len(fnType.Params) != 1 {
		t.Fatalf("expected 1 parameter type, got %d", len(fnType.Params))
	}
	if paramTypeName := typeName(fnType.Params[0]); paramTypeName != "T" {
		t.Fatalf("expected parameter type param reference, got %q", paramTypeName)
	}
	if returnTypeName := typeName(fnType.Return); returnTypeName != "T" {
		t.Fatalf("expected return type param reference, got %q", returnTypeName)
	}
}
func TestStringInterpolationInfersStringType(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("name"), ast.Str("world"))
	interp := ast.Interp(ast.Str("Hello, "), ast.ID("name"))
	module := ast.NewModule([]ast.Statement{assign, interp}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[interp]
	if !ok {
		t.Fatalf("expected interpolation inference entry")
	}
	if typeName(typ) != "string" {
		t.Fatalf("expected interpolation to yield string, got %q", typeName(typ))
	}
}
func TestStringInterpolationAllowsNonStringPartsForNow(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("count"), ast.Int(42))
	interp := ast.Interp(ast.Str("Count: "), ast.ID("count"))
	module := ast.NewModule([]ast.Statement{assign, interp}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for interpolation, got %v", diags)
	}
}
func TestRethrowAllowedInsideRescue(t *testing.T) {
	checker := New()
	errorStruct := ast.StructDef(
		"Error",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("string"), "message"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	raiseBlock := ast.Block(
		ast.Raise(
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Str("boom"), "message"),
				},
				false,
				"Error",
				nil,
				nil,
			),
		),
	)
	rescue := ast.Rescue(
		raiseBlock,
		ast.Mc(ast.Wc(), ast.Block(ast.Rethrow())),
	)
	module := ast.NewModule([]ast.Statement{errorStruct, rescue}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for rethrow inside rescue, got %v", diags)
	}
}
func TestRethrowOutsideRescueProducesDiagnostic(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{ast.Rethrow()}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for rethrow outside rescue")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "rethrow is only valid inside rescue handlers") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing rethrow diagnostic: %v", diags)
	}
}
func TestLabeledBreakRequiresBreakpoint(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{ast.Brk("missing", nil)}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for labeled break without breakpoint")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "unknown break label") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing labeled break diagnostic: %v", diags)
	}
}
func TestLabeledBreakInsideBreakpoint(t *testing.T) {
	checker := New()
	breakpoint := ast.Bp("loop", ast.Brk("loop", ast.Int(1)))
	module := ast.NewModule([]ast.Statement{breakpoint}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for labeled break inside breakpoint, got %v", diags)
	}
}
func TestStructMemberAccessInfersFieldType(t *testing.T) {
	checker := New()
	structDef := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	literal := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(5), "x"),
			ast.FieldInit(ast.Int(6), "y"),
		},
		false,
		"Point",
		nil,
		nil,
	)
	assign := ast.Assign(ast.ID("pt"), literal)
	member := ast.Member(ast.ID("pt"), "x")
	module := ast.NewModule([]ast.Statement{structDef, assign, member}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	memberType, ok := checker.infer[member]
	if !ok {
		t.Fatalf("expected inference for member access")
	}
	if typeName(memberType) != "i32" {
		t.Fatalf("expected field type i32, got %q", typeName(memberType))
	}
}
func TestStructMemberAccessMissingField(t *testing.T) {
	checker := New()
	structDef := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	assign := ast.Assign(
		ast.ID("pt"),
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(1), "x"),
			},
			false,
			"Point",
			nil,
			nil,
		),
	)
	member := ast.Member(ast.ID("pt"), "z")
	module := ast.NewModule([]ast.Statement{structDef, assign, member}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for missing field")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "has no member 'z'") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing field diagnostic not found: %v", diags)
	}
}
func TestArrayLiteralInfersElementType(t *testing.T) {
	checker := New()
	arr := ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3))
	assign := ast.Assign(ast.ID("values"), arr)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	arrType, ok := checker.infer[arr]
	if !ok {
		t.Fatalf("expected inference entry for array literal")
	}
	array, ok := arrType.(ArrayType)
	if !ok {
		t.Fatalf("expected ArrayType, got %#v", arrType)
	}
	if array.Element == nil {
		t.Fatalf("expected element type information")
	}
	if typeName(array.Element) != "i32" {
		t.Fatalf("expected element type i32, got %q", typeName(array.Element))
	}
}
func TestArrayLiteralElementTypeMismatchProducesDiagnostic(t *testing.T) {
	checker := New()
	arr := ast.Arr(ast.Int(1), ast.Str("two"))
	module := ast.NewModule([]ast.Statement{arr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for mismatched array elements")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "array element 2") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected mismatch diagnostic, got %v", diags)
	}
}
func TestEmptyArrayLiteralDefaultsToUnknownElement(t *testing.T) {
	checker := New()
	arr := ast.Arr()
	module := ast.NewModule([]ast.Statement{arr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	arrType, ok := checker.infer[arr]
	if !ok {
		t.Fatalf("expected inference entry for empty array")
	}
	array, ok := arrType.(ArrayType)
	if !ok {
		t.Fatalf("expected ArrayType, got %#v", arrType)
	}
	if array.Element == nil || !isUnknownType(array.Element) {
		t.Fatalf("expected unknown element type, got %#v", array.Element)
	}
}
func TestArrayIndexExpressionInfersElementType(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("items"), ast.Arr(ast.Int(1), ast.Int(2)))
	index := ast.Index(ast.ID("items"), ast.Int(1))
	module := ast.NewModule([]ast.Statement{assign, index}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	idxType, ok := checker.infer[index]
	if !ok {
		t.Fatalf("expected inference entry for index expression")
	}
	if typeName(idxType) != "i32" {
		t.Fatalf("expected index expression to have type i32, got %q", typeName(idxType))
	}
}
func TestArrayIndexExpressionRequiresIntegerIndex(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("items"), ast.Arr(ast.Int(1), ast.Int(2)))
	index := ast.Index(ast.ID("items"), ast.Str("oops"))
	module := ast.NewModule([]ast.Statement{assign, index}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "index must be an integer") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected integer index diagnostic, got %v", diags)
	}
}
func TestIndexingNonArrayProducesDiagnostic(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("value"), ast.Int(5))
	index := ast.Index(ast.ID("value"), ast.Int(0))
	module := ast.NewModule([]ast.Statement{assign, index}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "cannot index into type") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected cannot index diagnostic, got %v", diags)
	}
}
func TestRangeExpressionProducesRangeType(t *testing.T) {
	checker := New()
	rangeExpr := ast.Range(ast.Int(1), ast.Int(3), true)
	module := ast.NewModule([]ast.Statement{rangeExpr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[rangeExpr]
	if !ok {
		t.Fatalf("expected inference entry for range expression")
	}
	rt, ok := typ.(RangeType)
	if !ok {
		t.Fatalf("expected RangeType, got %#v", typ)
	}
	if rt.Element == nil || typeName(rt.Element) != "i32" {
		t.Fatalf("expected range element type i32, got %#v", rt.Element)
	}
}
func TestRangeExpressionRequiresNumericBounds(t *testing.T) {
	checker := New()
	rangeExpr := ast.Range(ast.Str("a"), ast.Int(3), false)
	module := ast.NewModule([]ast.Statement{rangeExpr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "range start must be numeric") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected numeric start diagnostic, got %v", diags)
	}
}
func TestRangeExpressionBoundsMustMatchType(t *testing.T) {
	checker := New()
	rangeExpr := ast.Range(ast.Int(1), ast.Flt(1.5), true)
	module := ast.NewModule([]ast.Statement{rangeExpr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "range bounds must share a numeric type") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected bounds mismatch diagnostic, got %v", diags)
	}
}
func TestSpawnExpressionReturnsFutureType(t *testing.T) {
	checker := New()
	expr := ast.Spawn(ast.Int(7))
	module := ast.NewModule([]ast.Statement{expr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	futureType, ok := checker.infer[expr]
	if !ok {
		t.Fatalf("expected future inference entry")
	}
	ft, ok := futureType.(FutureType)
	if !ok {
		t.Fatalf("expected FutureType, got %#v", futureType)
	}
	if ft.Result == nil || typeName(ft.Result) != "i32" {
		t.Fatalf("expected future result i32, got %#v", ft.Result)
	}
}
func TestPropagationExtractsSuccessBranch(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("handle"), ast.Proc(ast.Int(10)))
	valueCall := ast.CallExpr(ast.Member(ast.ID("handle"), "value"))
	prop := ast.Prop(valueCall)
	module := ast.NewModule([]ast.Statement{assign, prop}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	propType, ok := checker.infer[prop]
	if !ok {
		t.Fatalf("expected propagation inference entry")
	}
	if propType == nil || typeName(propType) != "i32" {
		t.Fatalf("expected propagation to produce i32, got %#v", propType)
	}
}
func TestPropagationRequiresProcErrorUnion(t *testing.T) {
	checker := New()
	prop := ast.Prop(ast.Int(5))
	module := ast.NewModule([]ast.Statement{prop}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for propagation on non-union type, got %v", diags)
	}
}
