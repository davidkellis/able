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
		ast.Ty("String"),
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
	if method.Return == nil || typeName(method.Return) != "String" {
		t.Fatalf("expected return type String, got %#v", method.Return)
	}
}

func TestInterfaceMemberAccessUsesSignature(t *testing.T) {
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
	if fnType.Return == nil || typeName(fnType.Return) != "String" {
		t.Fatalf("expected method return type String, got %#v", fnType.Return)
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
	if fnType.Return == nil || typeName(fnType.Return) != "String" {
		t.Fatalf("expected method return type String, got %#v", fnType.Return)
	}
}

func TestInterfaceDuplicateMethodDiagnostic(t *testing.T) {
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
	otherShow := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("fmt", ast.Ty("String")),
		},
		ast.Ty("String"),
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
		ast.Ty("String"),
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
		ast.Ty("String"),
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
		[]ast.TypeExpression{ast.Ty("String"), ast.Ty("i32")},
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
		ast.Ty("String"),
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
		if strings.Contains(d.Message, "return type expected String") {
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
			ast.Param("self", ast.Ty("Wrapper")),
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
}

func TestImplementationRejectsBareConstructorWithoutSelfPattern(t *testing.T) {
	checker := New()
	displayIface := ast.Iface("Display", nil, nil, nil, nil, nil, false)
	arrayStruct := buildGenericStructDefinition("Array", []string{"T"})
	impl := ast.Impl("Display", ast.Ty("Array"), nil, nil, nil, nil, nil, false)
	module := ast.NewModule([]ast.Statement{displayIface, arrayStruct, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for bare constructor implementation")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "type constructor") && strings.Contains(d.Message, "Display") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected bare constructor diagnostic, got %v", diags)
	}
}

func TestImplementationAllowsBareConstructorWithSelfPattern(t *testing.T) {
	checker := New()
	selfPattern := ast.Gen(ast.Ty("M"), ast.WildT())
	mapperIface := ast.Iface("Mapper", nil, nil, selfPattern, nil, nil, false)
	arrayStruct := buildGenericStructDefinition("Array", []string{"T"})
	impl := ast.Impl("Mapper", ast.Ty("Array"), nil, nil, nil, nil, nil, false)
	module := ast.NewModule([]ast.Statement{mapperIface, arrayStruct, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestImplementationAllowsBareConstructorWithMethodGenerics(t *testing.T) {
	checker := New()
	selfPattern := ast.Gen(ast.Ty("F"), ast.WildT())
	wrapSig := ast.FnSig(
		"wrap",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Self"), ast.Ty("T"))),
			ast.Param("value", ast.Ty("T")),
		},
		ast.Gen(ast.Ty("Self"), ast.Ty("T")),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		nil,
	)
	wrapperIface := ast.Iface("Wrapper", []*ast.FunctionSignature{wrapSig}, nil, selfPattern, nil, nil, false)
	holderFields := []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Ty("T"), "value"),
	}
	holderStruct := ast.StructDef("Holder", holderFields, ast.StructKindNamed, []*ast.GenericParameter{ast.GenericParam("T")}, nil, false)
	wrapFn := ast.Fn(
		"wrap",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Holder"), ast.Ty("T"))),
			ast.Param("value", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.Ret(
				ast.StructLit(
					[]*ast.StructFieldInitializer{
						ast.FieldInit(ast.ID("value"), "value"),
					},
					false,
					"Holder",
					nil,
					[]ast.TypeExpression{ast.Ty("T")},
				),
			),
		},
		ast.Gen(ast.Ty("Holder"), ast.Ty("T")),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	impl := ast.Impl("Wrapper", ast.Ty("Holder"), []*ast.FunctionDefinition{wrapFn}, nil, nil, nil, nil, false)
	module := ast.NewModule([]ast.Statement{wrapperIface, holderStruct, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestImplementationRejectsExplicitSelfPatternMismatch(t *testing.T) {
	checker := New()
	pointIface := ast.Iface("PointOnly", nil, nil, ast.Ty("Point"), nil, nil, false)
	pointStruct := buildPointStructDefinition()
	lineStruct := ast.StructDef("Line", nil, ast.StructKindNamed, nil, nil, false)
	impl := ast.Impl("PointOnly", ast.Ty("Line"), nil, nil, nil, nil, nil, false)
	module := ast.NewModule([]ast.Statement{pointIface, pointStruct, lineStruct, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for self type mismatch")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "self type 'Point'") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected explicit self pattern diagnostic, got %v", diags)
	}
}

func TestImplementationRejectsGenericSelfPatternMismatch(t *testing.T) {
	checker := New()
	arrayPattern := ast.Gen(ast.Ty("Array"), ast.Ty("T"))
	arrayIface := ast.Iface("ArrayOnly", nil, nil, arrayPattern, nil, nil, false)
	arrayStruct := buildGenericStructDefinition("Array", []string{"T"})
	resultStruct := buildGenericStructDefinition("Result", []string{"T"})
	target := ast.Gen(ast.Ty("Result"), ast.Ty("i32"))
	impl := ast.Impl("ArrayOnly", target, nil, nil, nil, nil, nil, false)
	module := ast.NewModule([]ast.Statement{arrayIface, arrayStruct, resultStruct, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for generic self pattern mismatch")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "self type 'Array T'") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected generic self pattern diagnostic, got %v", diags)
	}
}

func TestImplementationRejectsConcreteTargetForHigherKindedPattern(t *testing.T) {
	checker := New()
	selfPattern := ast.Gen(ast.Ty("F"), ast.WildT())
	applicativeIface := ast.Iface("Applicative", nil, nil, selfPattern, nil, nil, false)
	arrayStruct := buildGenericStructDefinition("Array", []string{"T"})
	implTarget := ast.Gen(ast.Ty("Array"), ast.Ty("i32"))
	impl := ast.Impl("Applicative", implTarget, nil, nil, nil, nil, nil, false)
	module := ast.NewModule([]ast.Statement{applicativeIface, arrayStruct, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for concrete higher-kinded target")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "self type 'F _'") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected higher-kinded self type diagnostic, got %v", diags)
	}
}

func buildGenericStructDefinition(name string, generics []string) *ast.StructDefinition {
	params := make([]*ast.GenericParameter, 0, len(generics))
	for _, g := range generics {
		params = append(params, ast.GenericParam(g))
	}
	return ast.StructDef(name, nil, ast.StructKindNamed, params, nil, false)
}

func buildPointStructDefinition() *ast.StructDefinition {
	fields := []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Ty("i32"), "x"),
		ast.FieldDef(ast.Ty("i32"), "y"),
	}
	return ast.StructDef("Point", fields, ast.StructKindNamed, nil, nil, false)
}
