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
		ast.Ty("String"),
		nil,
		nil,
		nil,
	)
	displayIface := ast.Iface("Display", []*ast.FunctionSignature{showSig}, nil, nil, nil, nil, false)
	wrapperStruct := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("String"), "value"),
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
		t.Fatalf("expected diagnostic for missing Display implementation, got %v", diags)
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

func TestConstraintSolverMethodSetGenericWhereFails(t *testing.T) {
	checker := New()
	formatSig := ast.FnSig(
		"format",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("T")),
		},
		ast.Ty("String"),
		nil,
		nil,
		nil,
	)
	displaySig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("String"),
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
			ast.FieldDef(ast.Ty("String"), "value"),
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
		ast.Ty("String"),
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

func TestConstraintSolverMethodSetSelfWhereAppliedConstraintFails(t *testing.T) {
	checker := New()
	formatSig := ast.FnSig(
		"format",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("T")),
		},
		ast.Ty("String"),
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
		ast.Ty("String"),
		nil,
		nil,
		nil,
	)
	describableIface := ast.Iface("Describable", []*ast.FunctionSignature{describeSig}, nil, nil, nil, nil, false)
	wrapperStruct := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("String"), "value"),
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
		ast.Ty("String"),
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
			ast.WhereConstraint("Self", ast.InterfaceConstr(ast.Gen(ast.Ty("Formatter"), ast.Ty("String")))),
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
		t.Fatalf("expected diagnostic for missing Formatter<String> implementation")
	}
	referencedConstraint := false
	for _, d := range diags {
		if strings.Contains(d.Message, "constraint on S") && strings.Contains(d.Message, "Describable") {
			if strings.Contains(d.Message, "method 'format' not provided") {
				referencedConstraint = true
			}
		}
	}
	if !referencedConstraint {
		t.Fatalf("expected diagnostic referencing missing method from where-clause constraint, got %v", diags)
	}
}

func TestConstraintSolverMethodSetGenericObligationFails(t *testing.T) {
	checker := New()
	displaySig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("String"),
		nil,
		nil,
		nil,
	)
	displayIface := ast.Iface("Display", []*ast.FunctionSignature{displaySig}, nil, nil, nil, nil, false)
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
		ast.Ty("String"),
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
	wrapperLiteral := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(1), "value"),
		},
		false,
		"Wrapper",
		nil,
		[]ast.TypeExpression{ast.Ty("i32")},
	)
	assign := ast.Assign(
		ast.ID("result"),
		ast.CallExpr(ast.Member(wrapperLiteral, "describe")),
	)
	module := ast.NewModule([]ast.Statement{
		displayIface,
		wrapperStruct,
		methods,
		assign,
	}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for missing Display implementation, got %v", diags)
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "constraint on T") && strings.Contains(d.Message, "Display") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected diagnostic referencing Display constraint, got %v", diags)
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
		ast.Ty("String"),
		nil,
		nil,
		nil,
	)
	formatterIface := ast.Iface("Formatter", []*ast.FunctionSignature{formatSig}, nil, nil, nil, nil, false)
	wrapperStruct := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("String"), "value"),
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
			ast.Param("value", ast.Ty("String")),
		},
		[]ast.Statement{
			ast.Block(ast.Str("ok")),
		},
		ast.Ty("String"),
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
