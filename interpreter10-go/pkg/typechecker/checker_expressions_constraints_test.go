package typechecker

import (
	"able/interpreter10-go/pkg/ast"
	"strings"
	"testing"
)

func TestConstraintSolverReportsMissingImpl(t *testing.T) {
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
	module := ast.NewModule([]ast.Statement{displayIface, wrapperStruct, fn, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for missing Display implementation")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "constraint on T") && strings.Contains(d.Message, "Wrapper") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected constraint diagnostic mentioning Wrapper, got %v", diags)
	}
}

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

func TestConstraintSolverMethodSetGenericWhereFails(t *testing.T) {
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
	genericMethod := ast.Fn(
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
	methods := ast.Methods(
		ast.Ty("Wrapper"),
		[]*ast.FunctionDefinition{genericMethod},
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
	module := ast.NewModule([]ast.Statement{displayIface, formatterIface, wrapperStruct, methods, useFormatter, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for missing Formatter implementation")
	}
	found := false
	contextIncluded := false
	for _, d := range diags {
		if strings.Contains(d.Message, "constraint on T") && strings.Contains(d.Message, "Formatter") {
			found = true
			if strings.Contains(d.Message, "via method 'format'") {
				contextIncluded = true
			}
		}
	}
	if !found {
		t.Fatalf("expected formatter diagnostic, got %v", diags)
	}
	if !contextIncluded {
		t.Fatalf("expected diagnostic to mention method context, got %v", diags)
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

func TestConstraintSolverMethodSetSelfWhereAppliedConstraintFails(t *testing.T) {
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
		methods,
		useDescribable,
		call,
	}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for missing Formatter<string> implementation")
	}
	referencedConstraint := false
	contextIncluded := false
	for _, d := range diags {
		if strings.Contains(d.Message, "constraint on S") && strings.Contains(d.Message, "Describable") {
			if strings.Contains(d.Message, "method 'format' not provided") {
				referencedConstraint = true
			}
			if strings.Contains(d.Message, "via method set") {
				contextIncluded = true
			}
		}
	}
	if !referencedConstraint {
		t.Fatalf("expected diagnostic referencing missing method from where-clause constraint, got %v", diags)
	}
	if !contextIncluded {
		t.Fatalf("expected diagnostic to include method-set context, got %v", diags)
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

func TestConstraintSolverMethodSetAmbiguousFailureIncludesCandidate(t *testing.T) {
	checker := New()
	formatSig := ast.FnSig(
		"format",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("bool")),
		},
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)
	formatterIface := ast.Iface("Formatter", []*ast.FunctionSignature{formatSig}, nil, nil, nil, nil, false)
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
	// Method set provides a mismatched signature so the solver should report the candidate.
	mismatchMethod := ast.Fn(
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
		[]*ast.FunctionDefinition{mismatchMethod},
		nil,
		nil,
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
	module := ast.NewModule([]ast.Statement{formatterIface, wrapperStruct, methods, useFormatter, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for method-set failure, got none")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "constraint on T") &&
			strings.Contains(d.Message, "Formatter") &&
			strings.Contains(d.Message, "methods for Wrapper") &&
			strings.Contains(d.Message, "method 'format'") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected formatter diagnostic citing method-set candidate, got %v", diags)
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
		if strings.Contains(d.Message, "Iterable<Array<i32>>") {
			foundDiag = true
			break
		}
	}
	if !foundDiag {
		t.Fatalf("expected diagnostic referencing Iterable<Array<i32>>, got %v", diags)
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
		t.Fatalf("expected substituted obligation for Iterable<Array<Int>>, got %#v", checker.obligations)
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
		if ob.Owner == "fn print" && ob.TypeParam == "T" && ob.Constraint != nil && ob.Constraint.Name() == "Interface:Display" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected obligation for T:Display, got %#v", checker.obligations)
	}
}
