package typechecker

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
)

func TestInterfaceDeclarationRegistersMethods(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	iface := ast.Iface("Display", []*ast.FunctionSignature{showSig}, nil, nil, nil, nil, false)
	module := ast.NewModule([]ast.Statement{iface}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}

	decl, ok := checker.global.Lookup("Display")
	if !ok {
		t.Fatalf("expected Display interface in global environment")
	}
	ifaceType, ok := decl.(InterfaceType)
	if !ok {
		t.Fatalf("expected InterfaceType, got %#v", decl)
	}
	if ifaceType.Methods == nil {
		t.Fatalf("expected methods map to be initialised")
	}
	method, ok := ifaceType.Methods["show"]
	if !ok {
		t.Fatalf("expected show method entry in interface methods")
	}
	if len(method.Params) != 1 {
		t.Fatalf("expected one parameter, got %d", len(method.Params))
	}
	if typeName(method.Params[0]) != "Self" {
		t.Fatalf("expected parameter type Self, got %q", typeName(method.Params[0]))
	}
	if method.Return == nil || typeName(method.Return) != "string" {
		t.Fatalf("expected return type string, got %#v", method.Return)
	}
}

func TestInterfaceMemberAccessUsesSignature(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	iface := ast.Iface("Display", []*ast.FunctionSignature{showSig}, nil, nil, nil, nil, false)
	module := ast.NewModule([]ast.Statement{iface}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}

	ifaceDecl, ok := checker.global.Lookup("Display")
	if !ok {
		t.Fatalf("expected Display interface in global environment")
	}
	ifaceType, ok := ifaceDecl.(InterfaceType)
	if !ok {
		t.Fatalf("expected InterfaceType, got %#v", ifaceDecl)
	}

	env := NewEnvironment(checker.global)
	env.Define("value", ifaceType)

	member := ast.Member(ast.ID("value"), "show")
	mDiags, memberType := checker.checkMemberAccess(env, member)
	if len(mDiags) != 0 {
		t.Fatalf("expected no diagnostics for valid interface method, got %v", mDiags)
	}
	fnType, ok := memberType.(FunctionType)
	if !ok {
		t.Fatalf("expected FunctionType, got %#v", memberType)
	}
	if fnType.Return == nil || typeName(fnType.Return) != "string" {
		t.Fatalf("expected method return type string, got %#v", fnType.Return)
	}

	missing := ast.Member(ast.ID("value"), "missing")
	missingDiags, _ := checker.checkMemberAccess(env, missing)
	if len(missingDiags) == 0 {
		t.Fatalf("expected diagnostic for missing interface method")
	}
	found := false
	for _, d := range missingDiags {
		if strings.Contains(d.Message, "interface 'Display' has no method 'missing'") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing interface method diagnostic not found: %v", missingDiags)
	}
}

func TestInterfaceGenericMethodSubstitutesArguments(t *testing.T) {
	checker := New()
	genericParam := ast.GenericParam("T")
	unwrapSig := ast.FnSig(
		"unwrap",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("T"),
		nil,
		nil,
		nil,
	)
	iface := ast.Iface("Wrapper", []*ast.FunctionSignature{unwrapSig}, []*ast.GenericParameter{genericParam}, nil, nil, nil, false)
	module := ast.NewModule([]ast.Statement{iface}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}

	decl, ok := checker.global.Lookup("Wrapper")
	if !ok {
		t.Fatalf("expected Wrapper interface in global environment")
	}
	ifaceType, ok := decl.(InterfaceType)
	if !ok {
		t.Fatalf("expected InterfaceType, got %#v", decl)
	}

	applied := AppliedType{
		Base: ifaceType,
		Arguments: []Type{
			PrimitiveType{Kind: PrimitiveString},
		},
	}
	env := NewEnvironment(checker.global)
	env.Define("wrapper", applied)

	member := ast.Member(ast.ID("wrapper"), "unwrap")
	mDiags, memberType := checker.checkMemberAccess(env, member)
	if len(mDiags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", mDiags)
	}
	fnType, ok := memberType.(FunctionType)
	if !ok {
		t.Fatalf("expected FunctionType, got %#v", memberType)
	}
	if fnType.Return == nil || typeName(fnType.Return) != "string" {
		t.Fatalf("expected method return type string, got %#v", fnType.Return)
	}
}

func TestInterfaceDuplicateMethodDiagnostic(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	otherShow := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("fmt", ast.Ty("string")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	iface := ast.Iface("Display", []*ast.FunctionSignature{showSig, otherShow}, nil, nil, nil, nil, false)
	module := ast.NewModule([]ast.Statement{iface}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected one diagnostic, got %v", diags)
	}
	if want := "duplicate interface method 'show'"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic containing %q, got %q", want, diags[0].Message)
	}
}

func TestFunctionConstraintMissingInterfaceProducesDiagnostic(t *testing.T) {
	checker := New()
	genericParam := ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Missing")))
	fn := ast.Fn(
		"printer",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		[]ast.Statement{ast.Ret(ast.ID("value"))},
		ast.Ty("T"),
		[]*ast.GenericParameter{genericParam},
		nil,
		false,
		false,
	)
	module := ast.NewModule([]ast.Statement{fn}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %v", diags)
	}
	if want := "references unknown interface 'Missing'"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic containing %q, got %q", want, diags[0].Message)
	}
}

func TestFunctionConstraintRequiresTypeArguments(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	wrapperIface := ast.Iface(
		"Wrapper",
		[]*ast.FunctionSignature{showSig},
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		nil,
		nil,
		false,
	)
	genericParam := ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Wrapper")))
	fn := ast.Fn(
		"printer",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		[]ast.Statement{ast.Ret(ast.ID("value"))},
		ast.Ty("T"),
		[]*ast.GenericParameter{genericParam},
		nil,
		false,
		false,
	)
	module := ast.NewModule([]ast.Statement{wrapperIface, fn}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %v", diags)
	}
	if want := "requires 1 type argument"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic containing %q, got %q", want, diags[0].Message)
	}
}

func TestImplementationConstraintMissingInterfaceDiagnostic(t *testing.T) {
	checker := New()
	wrapperStruct := ast.StructDef("Wrapper", nil, ast.StructKindNamed, []*ast.GenericParameter{ast.GenericParam("T")}, nil, false)
	implGeneric := ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Missing")))
	showMethod := ast.Fn(
		"show",
		[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
		[]ast.Statement{ast.Ret(ast.Str("wrapper"))},
		ast.Ty("string"),
		[]*ast.GenericParameter{implGeneric},
		nil,
		false,
		false,
	)
	impl := ast.Impl(
		"Display",
		ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
		[]*ast.FunctionDefinition{showMethod},
		nil,
		[]*ast.GenericParameter{implGeneric},
		nil,
		nil,
		false,
	)
	module := ast.NewModule([]ast.Statement{wrapperStruct, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for missing interface constraint")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "references unknown interface 'Missing'") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected missing interface diagnostic, got %v", diags)
	}
}

func TestImplementationInterfaceTypeArgumentMissing(t *testing.T) {
	checker := New()
	displayIface := ast.Iface(
		"Display",
		[]*ast.FunctionSignature{},
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		nil,
		nil,
		false,
	)
	wrapperStruct := ast.StructDef(
		"Wrapper",
		nil,
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	impl := ast.Impl(
		"Display",
		ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
		nil,
		nil,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		nil,
		false,
	)
	module := ast.NewModule([]ast.Statement{displayIface, wrapperStruct, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for missing interface type arguments")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "requires 1 interface type argument") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected interface type argument diagnostic, got %v", diags)
	}
}

func TestImplementationInterfaceTypeArgumentMismatch(t *testing.T) {
	checker := New()
	displayIface := ast.Iface(
		"Display",
		[]*ast.FunctionSignature{},
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		nil,
		nil,
		false,
	)
	wrapperStruct := ast.StructDef(
		"Wrapper",
		nil,
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	impl := ast.Impl(
		"Display",
		ast.Ty("Wrapper"),
		nil,
		nil,
		nil,
		[]ast.TypeExpression{ast.Ty("string"), ast.Ty("i32")},
		nil,
		false,
	)
	module := ast.NewModule([]ast.Statement{displayIface, wrapperStruct, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for interface type argument mismatch")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "expected 1 interface type argument") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected interface argument mismatch diagnostic, got %v", diags)
	}
}

func TestImplementationMissingMethodDiagnostic(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	displayIface := ast.Iface("Display", []*ast.FunctionSignature{showSig}, nil, nil, nil, nil, false)
	wrapperStruct := ast.StructDef("Wrapper", nil, ast.StructKindNamed, nil, nil, false)
	impl := ast.Impl("Display", ast.Ty("Wrapper"), nil, nil, nil, nil, nil, false)

	module := ast.NewModule([]ast.Statement{displayIface, wrapperStruct, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for missing implementation method")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "missing method 'show'") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected missing method diagnostic, got %v", diags)
	}
}

func TestImplementationMethodSignatureMismatch(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	displayIface := ast.Iface("Display", []*ast.FunctionSignature{showSig}, nil, nil, nil, nil, false)
	wrapperStruct := ast.StructDef("Wrapper", nil, ast.StructKindNamed, nil, nil, false)

	showMethod := ast.Fn(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Wrapper")),
		},
		[]ast.Statement{
			ast.Ret(ast.Str("value")),
		},
		ast.Ty("i32"),
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
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for signature mismatch")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "return type expected string") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected return type mismatch diagnostic, got %v", diags)
	}
}

func TestImplementationMethodMatchesInterface(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	displayIface := ast.Iface("Display", []*ast.FunctionSignature{showSig}, nil, nil, nil, nil, false)
	wrapperStruct := ast.StructDef("Wrapper", nil, ast.StructKindNamed, nil, nil, false)

	showMethod := ast.Fn(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Wrapper")),
		},
		[]ast.Statement{
			ast.Ret(ast.Str("value")),
		},
		ast.Ty("string"),
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
}

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
		ast.Ty("string"),
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
	if fnType.Return == nil || typeName(fnType.Return) != "string" {
		t.Fatalf("expected return type string, got %#v", fnType.Return)
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
	if fnType.Return == nil || typeName(fnType.Return) != "string" {
		t.Fatalf("expected return type string, got %#v", fnType.Return)
	}
}

func TestImplementationMemberAccessProvidesMethodType(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
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
		ast.Ty("string"),
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
	if fnType.Return == nil || typeName(fnType.Return) != "string" {
		t.Fatalf("expected return type string, got %#v", fnType.Return)
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
	if fnType.Return == nil || typeName(fnType.Return) != "string" {
		t.Fatalf("expected return type string, got %#v", fnType.Return)
	}
}
func TestTypeParameterMemberAccessWithConstraint(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
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
		ast.Ty("string"),
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
	if fnType.Return == nil || typeName(fnType.Return) != "string" {
		t.Fatalf("expected return type string, got %#v", fnType.Return)
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
		ast.Ty("string"),
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
