package typechecker

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestMethodsDefinitionMemberAccessProvidesMethodType(t *testing.T) {
	checker := New()
	structDef := ast.StructDef("Wrapper", nil, ast.StructKindNamed, nil, nil, false)
	describeMethod := ast.Fn(
		"describe",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Ret(ast.Str("value")),
		},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	methods := ast.Methods(ast.Ty("Wrapper"), []*ast.FunctionDefinition{describeMethod}, nil, nil)
	module := ast.NewModule([]ast.Statement{structDef, methods}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}

	env := NewEnvironment(checker.global)
	env.Define("value", StructInstanceType{StructName: "Wrapper"})

	member := ast.Member(ast.ID("value"), "describe")
	mDiags, memberType := checker.checkMemberAccess(env, member)
	if len(mDiags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", mDiags)
	}
	fnType, ok := memberType.(FunctionType)
	if !ok {
		t.Fatalf("expected FunctionType, got %#v", memberType)
	}
	if fnType.Return == nil || typeName(fnType.Return) != "String" {
		t.Fatalf("expected return type String, got %#v", fnType.Return)
	}
}

func TestMethodsDefinitionMemberAccessSubstitutesGenerics(t *testing.T) {
	checker := New()
	structDef := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	valueMethod := ast.Fn(
		"value",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Ret(ast.Member(ast.ID("self"), "value")),
		},
		ast.Ty("T"),
		nil,
		nil,
		false,
		false,
	)
	methods := ast.Methods(
		ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
		[]*ast.FunctionDefinition{valueMethod},
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
	)
	module := ast.NewModule([]ast.Statement{structDef, methods}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}

	env := NewEnvironment(checker.global)
	env.Define("value", AppliedType{
		Base: StructType{StructName: "Wrapper"},
		Arguments: []Type{
			PrimitiveType{Kind: PrimitiveString},
		},
	})

	member := ast.Member(ast.ID("value"), "value")
	mDiags, memberType := checker.checkMemberAccess(env, member)
	if len(mDiags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", mDiags)
	}
	fnType, ok := memberType.(FunctionType)
	if !ok {
		t.Fatalf("expected FunctionType, got %#v", memberType)
	}
	if fnType.Return == nil || typeName(fnType.Return) != "String" {
		t.Fatalf("expected return type String, got %#v", fnType.Return)
	}
}

func TestMethodsDefinitionAllowsImplicitSelfWithoutAnnotation(t *testing.T) {
	checker := New()
	structDef := ast.StructDef(
		"Channel",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i64"), "handle"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	sendMethod := ast.Fn(
		"send",
		[]*ast.FunctionParameter{
			ast.Param("self", nil),
			ast.Param("value", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Member(ast.ID("self"), "handle"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	methods := ast.Methods(ast.Ty("Channel"), []*ast.FunctionDefinition{sendMethod}, nil, nil)
	module := ast.NewModule([]ast.Statement{structDef, methods}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestImplementationAllowsImplicitSelfWithoutAnnotation(t *testing.T) {
	checker := New()
	iface := ast.Iface(
		"Show",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"show",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
				},
				ast.Ty("i32"),
				nil,
				nil,
				nil,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	structDef := ast.StructDef(
		"Meter",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "reading"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	showMethod := ast.Fn(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", nil),
		},
		[]ast.Statement{
			ast.Ret(ast.Member(ast.ID("self"), "reading")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	impl := ast.Impl(
		"Show",
		ast.Ty("Meter"),
		[]*ast.FunctionDefinition{showMethod},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	module := ast.NewModule([]ast.Statement{iface, structDef, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestImplementationMemberAccessProvidesMethodType(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("String"),
		nil,
		nil,
		nil,
	)
	displayIface := ast.Iface("Display", []*ast.FunctionSignature{showSig}, nil, nil, nil, nil, false)
	wrapperStruct := ast.StructDef("Wrapper", nil, ast.StructKindNamed, nil, nil, false)
	showMethod := ast.Fn(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Ret(ast.Str("value")),
		},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	impl := ast.Impl("Display", ast.Ty("Wrapper"), []*ast.FunctionDefinition{showMethod}, nil, nil, nil, nil, false)
	module := ast.NewModule([]ast.Statement{displayIface, wrapperStruct, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}

	env := NewEnvironment(checker.global)
	env.Define("value", StructInstanceType{StructName: "Wrapper"})

	member := ast.Member(ast.ID("value"), "show")
	mDiags, memberType := checker.checkMemberAccess(env, member)
	if len(mDiags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", mDiags)
	}
	fnType, ok := memberType.(FunctionType)
	if !ok {
		t.Fatalf("expected FunctionType, got %#v", memberType)
	}
	if fnType.Return == nil || typeName(fnType.Return) != "String" {
		t.Fatalf("expected return type String, got %#v", fnType.Return)
	}
}

func TestImplementationMemberAccessSubstitutesGenerics(t *testing.T) {
	checker := New()
	itemSig := ast.FnSig(
		"item",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("T"),
		nil,
		nil,
		nil,
	)
	containerIface := ast.Iface(
		"Container",
		[]*ast.FunctionSignature{itemSig},
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		nil,
		nil,
		false,
	)
	structDef := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	itemMethod := ast.Fn(
		"item",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Ret(ast.Member(ast.ID("self"), "value")),
		},
		ast.Ty("T"),
		nil,
		nil,
		false,
		false,
	)
	impl := ast.Impl(
		"Container",
		ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
		[]*ast.FunctionDefinition{itemMethod},
		nil,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		[]ast.TypeExpression{ast.Ty("T")},
		nil,
		false,
	)
	module := ast.NewModule([]ast.Statement{containerIface, structDef, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}

	env := NewEnvironment(checker.global)
	env.Define("value", AppliedType{
		Base: StructType{StructName: "Wrapper"},
		Arguments: []Type{
			PrimitiveType{Kind: PrimitiveString},
		},
	})

	member := ast.Member(ast.ID("value"), "item")
	mDiags, memberType := checker.checkMemberAccess(env, member)
	if len(mDiags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", mDiags)
	}
	fnType, ok := memberType.(FunctionType)
	if !ok {
		t.Fatalf("expected FunctionType, got %#v", memberType)
	}
	if fnType.Return == nil || typeName(fnType.Return) != "String" {
		t.Fatalf("expected return type String, got %#v", fnType.Return)
	}
}
func TestTypeParameterMemberAccessWithConstraint(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("String"),
		nil,
		nil,
		nil,
	)
	displayIface := ast.Iface("Display", []*ast.FunctionSignature{showSig}, nil, nil, nil, nil, false)
	member := ast.Member(ast.ID("value"), "show")
	call := ast.CallExpr(member)
	fn := ast.Fn(
		"print",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.Ret(call),
		},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Display")))},
		nil,
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
	typ, ok := checker.infer.get(member)
	if !ok {
		t.Fatalf("expected inference for member access")
	}
	fnType, ok := typ.(FunctionType)
	if !ok {
		t.Fatalf("expected FunctionType, got %#v", typ)
	}
	if fnType.Return == nil || typeName(fnType.Return) != "String" {
		t.Fatalf("expected return type String, got %#v", fnType.Return)
	}
}

func TestTypeParameterMemberAccessWithoutConstraint(t *testing.T) {
	checker := New()
	member := ast.Member(ast.ID("value"), "show")
	call := ast.CallExpr(member)
	fn := ast.Fn(
		"print",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.Ret(call),
		},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
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
		t.Fatalf("expected diagnostic for missing constraint")
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
