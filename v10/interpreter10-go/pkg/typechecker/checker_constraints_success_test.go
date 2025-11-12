package typechecker

import (
	"able/interpreter10-go/pkg/ast"
	"strings"
	"testing"
)

func TestConstraintSolverAcceptsSatisfiedImpl(t *testing.T) {
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
	wrapperStruct := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("string"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	showMethod := ast.Fn(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Block(ast.Str("ok")),
		},
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)
	displayImpl := ast.Impl(
		"Display",
		ast.Ty("Wrapper"),
		[]*ast.FunctionDefinition{showMethod},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	genericParam := ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Display")))
	fn := ast.Fn(
		"useDisplay",
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
	call := ast.Call(
		"useDisplay",
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Str("hi"), "value"),
			},
			false,
			"Wrapper",
			nil,
			nil,
		),
	)
	module := ast.NewModule([]ast.Statement{displayIface, wrapperStruct, displayImpl, fn, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestConstraintSolverAcceptsMethodSet(t *testing.T) {
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
	wrapperStruct := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("string"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	showMethod := ast.Fn(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Block(ast.Str("ok")),
		},
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)
	methods := ast.Methods(
		ast.Ty("Wrapper"),
		[]*ast.FunctionDefinition{showMethod},
		nil,
		nil,
	)
	genericParam := ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Display")))
	fn := ast.Fn(
		"useDisplay",
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
	call := ast.Call(
		"useDisplay",
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Str("hi"), "value"),
			},
			false,
			"Wrapper",
			nil,
			nil,
		),
	)
	module := ast.NewModule([]ast.Statement{displayIface, wrapperStruct, methods, fn, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestConstraintSolverRecognisesUnionImplementation(t *testing.T) {
	checker := New()
	optionUnion := UnionType{
		UnionName: "Option",
		TypeParams: []GenericParamSpec{
			{Name: "T"},
		},
		Variants: []Type{
			PrimitiveType{Kind: PrimitiveNil},
			TypeParameterType{ParameterName: "T"},
		},
	}
	displayIface := InterfaceType{InterfaceName: "Display"}
	subject := AppliedType{
		Base: optionUnion,
		Arguments: []Type{
			StructType{StructName: "Point"},
		},
	}

	if ok, _ := checker.typeImplementsInterface(subject, displayIface, nil); ok {
		t.Fatalf("expected Option Point to not satisfy Display without implementation")
	}

	optionSpec := ImplementationSpec{
		InterfaceName: "Display",
		TypeParams: []GenericParamSpec{
			{Name: "T"},
		},
		Target: AppliedType{
			Base: optionUnion,
			Arguments: []Type{
				TypeParameterType{ParameterName: "T"},
			},
		},
		Methods: make(map[string]FunctionType),
	}

	checker.implementations = []ImplementationSpec{optionSpec}

	if ok, detail := checker.typeImplementsInterface(subject, displayIface, nil); !ok {
		t.Fatalf("expected Option Point to satisfy Display via implementation (detail: %s)", detail)
	}
}

func TestUnionImplementationMatchesIndividualVariants(t *testing.T) {
	checker := New()
	alpha := StructType{StructName: "Alpha"}
	beta := StructType{StructName: "Beta"}
	showIface := InterfaceType{InterfaceName: "Show"}
	unionSpec := ImplementationSpec{
		InterfaceName: "Show",
		Target:        UnionLiteralType{Members: []Type{alpha, beta}},
		Methods:       make(map[string]FunctionType),
		UnionVariants: []string{"Alpha", "Beta"},
	}
	checker.implementations = []ImplementationSpec{unionSpec}
	if ok, detail := checker.implementationProvidesInterface(alpha, showIface, nil); !ok {
		t.Fatalf("expected Alpha to satisfy union-target implementation (detail: %s)", detail)
	}
}

func TestUnionImplementationAmbiguityReported(t *testing.T) {
	checker := New()
	alpha := StructType{StructName: "Alpha"}
	beta := StructType{StructName: "Beta"}
	gamma := StructType{StructName: "Gamma"}
	showIface := InterfaceType{InterfaceName: "Show"}
	implAB := ImplementationSpec{
		InterfaceName: "Show",
		Target:        UnionLiteralType{Members: []Type{alpha, beta}},
		Methods:       make(map[string]FunctionType),
		UnionVariants: []string{"Alpha", "Beta"},
	}
	implAG := ImplementationSpec{
		InterfaceName: "Show",
		Target:        UnionLiteralType{Members: []Type{alpha, gamma}},
		Methods:       make(map[string]FunctionType),
		UnionVariants: []string{"Alpha", "Gamma"},
	}
	checker.implementations = []ImplementationSpec{implAB, implAG}
	if ok, detail := checker.typeImplementsInterface(alpha, showIface, nil); ok {
		t.Fatalf("expected Alpha to fail due to ambiguous implementations")
	} else if !strings.Contains(detail, "ambiguous implementations") {
		t.Fatalf("expected ambiguity detail, got %q", detail)
	}
}

func TestConstraintSolverRecognisesNullableImplementation(t *testing.T) {
	checker := New()
	displayIface := InterfaceType{InterfaceName: "Display"}
	subject := NullableType{Inner: StructType{StructName: "Point"}}

	nullableSpec := ImplementationSpec{
		InterfaceName: "Display",
		TypeParams: []GenericParamSpec{
			{Name: "T", Constraints: []Type{displayIface}},
		},
		Target:  NullableType{Inner: TypeParameterType{ParameterName: "T"}},
		Methods: make(map[string]FunctionType),
	}
	nullableSpec.Obligations = obligationsFromSpecs("impl Display for Nullable T", nullableSpec.TypeParams, nil, nil)
	checker.implementations = []ImplementationSpec{nullableSpec}

	if ok, detail := checker.implementationProvidesInterface(subject, displayIface, nil); ok {
		t.Fatalf("expected Nullable Point to fail constraint without inner implementation")
	} else if detail == "" {
		t.Fatalf("expected nullable implementation failure to report detail")
	}

	if ok, detail := checker.typeImplementsInterface(subject, displayIface, nil); ok {
		t.Fatalf("expected Nullable Point to not satisfy Display without inner implementation")
	} else if detail == "" || !strings.Contains(detail, "impl Display") {
		t.Fatalf("expected nullable constraint failure to mention impl context, got %q", detail)
	}

	pointSpec := ImplementationSpec{
		InterfaceName: "Display",
		Target:        StructType{StructName: "Point"},
		Methods:       make(map[string]FunctionType),
	}
	checker.implementations = append(checker.implementations, pointSpec)

	if ok, detail := checker.implementationProvidesInterface(subject, displayIface, nil); !ok {
		t.Fatalf("expected Nullable Point to match nullable implementation once constraints satisfied (detail: %s)", detail)
	}

	if ok, detail := checker.typeImplementsInterface(subject, displayIface, nil); !ok {
		t.Fatalf("expected Nullable Point to satisfy Display via implementation (detail: %s)", detail)
	}
}

func TestMatchMethodTargetHandlesNullableAndNestedWrappers(t *testing.T) {
	param := GenericParamSpec{Name: "T"}
	subject := NullableType{Inner: StructType{StructName: "Point"}}
	target := NullableType{Inner: TypeParameterType{ParameterName: "T"}}

	subst, score, ok := matchMethodTarget(subject, target, []GenericParamSpec{param})
	if !ok {
		t.Fatalf("expected nullable target to match")
	}
	if score == 0 {
		t.Fatalf("expected nullable match to contribute score")
	}
	actual, exists := subst["T"]
	if !exists || typeName(actual) != "Point" {
		t.Fatalf("expected substitution for T to be Point, got %#v", actual)
	}

	optionUnion := UnionType{
		UnionName:  "Option",
		TypeParams: []GenericParamSpec{{Name: "T"}},
		Variants: []Type{
			PrimitiveType{Kind: PrimitiveNil},
			TypeParameterType{ParameterName: "T"},
		},
	}
	subjectNested := AppliedType{
		Base: StructType{StructName: "Box"},
		Arguments: []Type{
			NullableType{Inner: AppliedType{Base: optionUnion, Arguments: []Type{StructType{StructName: "Point"}}}},
		},
	}
	targetNested := AppliedType{
		Base: StructType{StructName: "Box"},
		Arguments: []Type{
			NullableType{Inner: AppliedType{Base: optionUnion, Arguments: []Type{TypeParameterType{ParameterName: "T"}}}},
		},
	}

	substNested, scoreNested, ok := matchMethodTarget(subjectNested, targetNested, []GenericParamSpec{param})
	if !ok {
		t.Fatalf("expected nested nullable/union target to match")
	}
	if scoreNested == 0 {
		t.Fatalf("expected nested match to contribute score")
	}
	inner, exists := substNested["T"]
	if !exists || typeName(inner) != "Point" {
		t.Fatalf("expected substitution for nested T to resolve to Point, got %#v", inner)
	}
}

func TestConstraintSolverMethodSetGenericWhereSatisfied(t *testing.T) {
	checker := New()
	formatSig := ast.FnSig(
		"format",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("T")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	formatterIface := ast.Iface(
		"Formatter",
		[]*ast.FunctionSignature{formatSig},
		[]*ast.GenericParameter{
			ast.GenericParam("T"),
		},
		nil,
		nil,
		nil,
		false,
	)
	wrapperStruct := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("string"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	formatMethod := ast.Fn(
		"format",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.Block(ast.Str("ok")),
		},
		ast.Ty("string"),
		[]*ast.GenericParameter{
			ast.GenericParam("T"),
		},
		[]*ast.WhereClauseConstraint{
			ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Display"))),
		},
		false,
		false,
	)
	displaySig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	displayIface := ast.Iface("Display", []*ast.FunctionSignature{displaySig}, nil, nil, nil, nil, false)
	displayImpl := ast.Impl(
		"Display",
		ast.Ty("Wrapper"),
		[]*ast.FunctionDefinition{ast.Fn(
			"show",
			[]*ast.FunctionParameter{
				ast.Param("self", ast.Ty("Self")),
			},
			[]ast.Statement{
				ast.Block(ast.Str("ok")),
			},
			ast.Ty("string"),
			nil,
			nil,
			false,
			false,
		)},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	methods := ast.Methods(
		ast.Ty("Wrapper"),
		[]*ast.FunctionDefinition{formatMethod},
		nil,
		nil,
	)
	genericParam := ast.GenericParam("T", ast.InterfaceConstr(ast.Gen(ast.Ty("Formatter"), ast.Ty("T"))))
	useFormatter := ast.Fn(
		"useFormatter",
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
	call := ast.Call(
		"useFormatter",
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Str("hi"), "value"),
			},
			false,
			"Wrapper",
			nil,
			nil,
		),
	)
	module := ast.NewModule([]ast.Statement{displayIface, formatterIface, wrapperStruct, methods, displayImpl, useFormatter, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %+v", diags)
	}
}

func TestConstraintSolverMethodSetSelfWhereAppliedConstraintSatisfied(t *testing.T) {
	checker := New()
	formatSig := ast.FnSig(
		"format",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("T")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	formatterIface := ast.Iface(
		"Formatter",
		[]*ast.FunctionSignature{formatSig},
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		nil,
		nil,
		false,
	)
	describeSig := ast.FnSig(
		"describe",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	describableIface := ast.Iface("Describable", []*ast.FunctionSignature{describeSig}, nil, nil, nil, nil, false)
	wrapperStruct := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("string"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	describeMethod := ast.Fn(
		"describe",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Block(ast.Str("ok")),
		},
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)
	methods := ast.Methods(
		ast.Ty("Wrapper"),
		[]*ast.FunctionDefinition{describeMethod},
		nil,
		[]*ast.WhereClauseConstraint{
			ast.WhereConstraint("Self", ast.InterfaceConstr(ast.Gen(ast.Ty("Formatter"), ast.Ty("string")))),
		},
	)
	formatMethod := ast.Fn(
		"format",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("string")),
		},
		[]ast.Statement{
			ast.Block(ast.Str("ok")),
		},
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)
	formatterImpl := ast.Impl(
		"Formatter",
		ast.Ty("Wrapper"),
		[]*ast.FunctionDefinition{formatMethod},
		nil,
		nil,
		[]ast.TypeExpression{ast.Ty("string")},
		nil,
		false,
	)
	genericParam := ast.GenericParam("S", ast.InterfaceConstr(ast.Ty("Describable")))
	useDescribable := ast.Fn(
		"useDescribable",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("S")),
		},
		[]ast.Statement{
			ast.Ret(ast.ID("value")),
		},
		ast.Ty("S"),
		[]*ast.GenericParameter{genericParam},
		nil,
		false,
		false,
	)
	wrapperLiteral := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Str("hi"), "value"),
		},
		false,
		"Wrapper",
		nil,
		nil,
	)
	call := ast.Call("useDescribable", wrapperLiteral)
	module := ast.NewModule([]ast.Statement{
		formatterIface,
		describableIface,
		wrapperStruct,
		formatterImpl,
		methods,
		useDescribable,
		call,
	}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestConstraintSolverMethodSetGenericObligationSatisfied(t *testing.T) {
	checker := New()
	displaySig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	displayIface := ast.Iface("Display", []*ast.FunctionSignature{displaySig}, nil, nil, nil, nil, false)
	pointStruct := ast.StructDef(
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
	showMethod := ast.Fn(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Point")),
		},
		[]ast.Statement{
			ast.Block(ast.Str("<point>")),
		},
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)
	displayImpl := ast.Impl(
		"Display",
		ast.Ty("Point"),
		[]*ast.FunctionDefinition{showMethod},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	wrapperStruct := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{
			ast.GenericParam("T"),
		},
		nil,
		false,
	)
	describeMethod := ast.Fn(
		"describe",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T"))),
		},
		[]ast.Statement{
			ast.Block(ast.Str("ok")),
		},
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)
	methods := ast.Methods(
		ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
		[]*ast.FunctionDefinition{describeMethod},
		[]*ast.GenericParameter{
			ast.GenericParam("T"),
		},
		[]*ast.WhereClauseConstraint{
			ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Display"))),
		},
	)
	pointLiteral := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(1), "x"),
			ast.FieldInit(ast.Int(2), "y"),
		},
		false,
		"Point",
		nil,
		nil,
	)
	wrapperLiteral := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(pointLiteral, "value"),
		},
		false,
		"Wrapper",
		nil,
		[]ast.TypeExpression{ast.Ty("Point")},
	)
	assign := ast.Assign(
		ast.ID("result"),
		ast.CallExpr(ast.Member(wrapperLiteral, "describe")),
	)
	module := ast.NewModule([]ast.Statement{
		displayIface,
		pointStruct,
		displayImpl,
		wrapperStruct,
		methods,
		assign,
	}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestConstraintSolverMethodSetWhereClauseContextDiagnostic(t *testing.T) {
	checker := New()
	formatSig := ast.FnSig(
		"format",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("string")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	formatterIface := ast.Iface("Formatter", []*ast.FunctionSignature{formatSig}, nil, nil, nil, nil, false)
	displaySig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	displayIface := ast.Iface("Display", []*ast.FunctionSignature{displaySig}, nil, nil, nil, nil, false)
	wrapperStruct := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("string"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	formatMethod := ast.Fn(
		"format",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("string")),
		},
		[]ast.Statement{
			ast.Block(ast.Str("ok")),
		},
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)
	methods := ast.Methods(
		ast.Ty("Wrapper"),
		[]*ast.FunctionDefinition{formatMethod},
		nil,
		[]*ast.WhereClauseConstraint{
			ast.WhereConstraint("Self", ast.InterfaceConstr(ast.Ty("Display"))),
		},
	)
	genericParam := ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Formatter")))
	useFormatter := ast.Fn(
		"useFormatter",
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
	call := ast.Call(
		"useFormatter",
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Str("hi"), "value"),
			},
			false,
			"Wrapper",
			nil,
			nil,
		),
	)
	module := ast.NewModule([]ast.Statement{formatterIface, displayIface, wrapperStruct, methods, useFormatter, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for missing Display implementation")
	}
	found := false
	contextIncluded := false
	for _, d := range diags {
		if strings.Contains(d.Message, "constraint on T") && strings.Contains(d.Message, "Formatter") {
			found = true
			if strings.Contains(d.Message, "via method set") {
				contextIncluded = true
			}
		}
	}
	if !found {
		t.Fatalf("expected formatter diagnostic, got %v", diags)
	}
	if !contextIncluded {
		t.Fatalf("expected diagnostic to mention method-set context, got %v", diags)
	}
}

func TestFunctionCallConstraintObligationsSubstituted(t *testing.T) {
	checker := New()
	iterSig := ast.FnSig(
		"next",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("Item"),
		nil,
		nil,
		nil,
	)
	iterableIface := ast.Iface(
		"Iterable",
		[]*ast.FunctionSignature{iterSig},
		[]*ast.GenericParameter{ast.GenericParam("Item")},
		nil,
		nil,
		nil,
		false,
	)
	constraintExpr := ast.Gen(
		ast.Ty("Iterable"),
		ast.Gen(ast.Ty("Array"), ast.Ty("T")),
	)
	genericParam := ast.GenericParam("T", ast.InterfaceConstr(constraintExpr))
	fn := ast.Fn(
		"useIterable",
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
	call := ast.Call("useIterable", ast.Int(5))
	module := ast.NewModule([]ast.Statement{iterableIface, fn, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for missing Iterable implementation")
	}
	foundDiag := false
	for _, d := range diags {
		if strings.Contains(d.Message, "Iterable Array i32") {
			foundDiag = true
			break
		}
	}
	if !foundDiag {
		t.Fatalf("expected diagnostic referencing Iterable Array i32, got %v", diags)
	}
	found := false
	for _, ob := range checker.obligations {
		if ob.Owner != "fn useIterable" {
			continue
		}
		applied, ok := ob.Constraint.(AppliedType)
		if !ok {
			continue
		}
		iface, ok := applied.Base.(InterfaceType)
		if !ok || iface.InterfaceName != "Iterable" {
			continue
		}
		if len(applied.Arguments) != 1 {
			continue
		}
		arrayArg, ok := applied.Arguments[0].(ArrayType)
		if !ok {
			continue
		}
		elem, ok := arrayArg.Element.(IntegerType)
		if !ok || elem.Suffix != "i32" {
			continue
		}
		found = true
		break
	}
	if !found {
		t.Fatalf("expected substituted obligation for Iterable Array Int, got %#v", checker.obligations)
	}
}

func TestFunctionGenericConstraintObligationRecorded(t *testing.T) {
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
	genericParam := ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Display")))
	fn := ast.Fn(
		"print",
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
	module := ast.NewModule([]ast.Statement{displayIface, fn}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	found := false
	for _, ob := range checker.obligations {
		if ob.Owner == "fn print" && ob.TypeParam == "T" && ob.Constraint != nil && typeName(ob.Constraint) == "Display" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected obligation for T:Display, got %#v", checker.obligations)
	}
}
