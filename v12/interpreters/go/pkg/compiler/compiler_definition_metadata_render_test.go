package compiler

import (
	"sort"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

func testInterfaceConstraint(typeExpr ast.TypeExpression) []*ast.InterfaceConstraint {
	return []*ast.InterfaceConstraint{
		ast.NewInterfaceConstraint(typeExpr),
	}
}

func testWhereConstraint(typeParam ast.TypeExpression, constraints []*ast.InterfaceConstraint) []*ast.WhereClauseConstraint {
	return []*ast.WhereClauseConstraint{
		ast.NewWhereClauseConstraint(typeParam, constraints),
	}
}

func testProgramFromModule(pkg string, module *ast.Module) *driver.Program {
	entry := annotatedModule(pkg, module, "main.able", nil)
	return &driver.Program{Entry: entry, Modules: []*driver.Module{entry}}
}

func combinedGeneratedSource(result *Result) string {
	if result == nil || len(result.Files) == 0 {
		return ""
	}
	keys := make([]string, 0, len(result.Files))
	for name := range result.Files {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	var out strings.Builder
	for _, name := range keys {
		out.Write(result.Files[name])
		out.WriteString("\n")
	}
	return out.String()
}

func TestCompilerPreservesDefinitionGenericConstraintsAndWhereClauses(t *testing.T) {
	tagIface := ast.NewInterfaceDefinition(
		ast.NewIdentifier("Tag"),
		[]*ast.FunctionSignature{
			ast.NewFunctionSignature(
				ast.NewIdentifier("tag"),
				[]*ast.FunctionParameter{
					ast.NewFunctionParameter(ast.NewIdentifier("self"), ast.Ty("Self")),
				},
				ast.Ty("String"),
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

	tagConstraint := testInterfaceConstraint(ast.Ty("Tag"))
	genericT := []*ast.GenericParameter{
		ast.NewGenericParameter(ast.NewIdentifier("T"), tagConstraint),
	}
	whereT := testWhereConstraint(ast.Ty("T"), tagConstraint)

	boxStruct := ast.NewStructDefinition(
		ast.NewIdentifier("Box"),
		[]*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.Ty("T"), ast.NewIdentifier("value")),
		},
		ast.StructKindNamed,
		genericT,
		whereT,
		false,
	)
	maybeUnion := ast.NewUnionDefinition(
		ast.NewIdentifier("Maybe"),
		[]ast.TypeExpression{ast.Ty("nil"), ast.Ty("T")},
		genericT,
		whereT,
		false,
	)
	describeSig := ast.NewFunctionSignature(
		ast.NewIdentifier("describe"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), ast.Ty("Self")),
		},
		ast.Ty("String"),
		genericT,
		whereT,
		nil,
	)
	withDefaultSig := ast.NewFunctionSignature(
		ast.NewIdentifier("with_default"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), ast.Ty("Self")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		ast.NewBlockExpression([]ast.Statement{ast.Int(7)}),
	)
	describerIface := ast.NewInterfaceDefinition(
		ast.NewIdentifier("Describer"),
		[]*ast.FunctionSignature{describeSig, withDefaultSig},
		genericT,
		ast.Ty("T"),
		whereT,
		nil,
		false,
	)
	mainFn := ast.NewFunctionDefinition(ast.NewIdentifier("main"), nil, ast.NewBlockExpression(nil), ast.Ty("void"), nil, nil, false, false)
	module := ast.NewModule(
		[]ast.Statement{tagIface, boxStruct, maybeUnion, describerIface, mainFn},
		nil,
		ast.NewPackageStatement([]*ast.Identifier{ast.NewIdentifier("demo")}, false),
	)

	result, err := New(Options{PackageName: "compiled"}).Compile(testProgramFromModule("demo", module))
	if err != nil {
		t.Fatalf("compile definitions with constraints/where: %v", err)
	}
	compiledSrc := combinedGeneratedSource(result)

	if !strings.Contains(compiledSrc, "ast.NewGenericParameter(ast.NewIdentifier(\"T\"),") || !strings.Contains(compiledSrc, "ast.NewInterfaceConstraint(ast.Ty(\"Tag\"))") {
		t.Fatalf("expected definition rendering to preserve generic parameter constraints")
	}
	if !strings.Contains(compiledSrc, "ast.NewWhereClauseConstraint(ast.Ty(\"T\"), []*ast.InterfaceConstraint{ast.NewInterfaceConstraint(ast.Ty(\"Tag\"))})") {
		t.Fatalf("expected definition rendering to preserve where-clause constraints")
	}
	if !strings.Contains(compiledSrc, "ast.NewFunctionSignature(ast.NewIdentifier(\"with_default\")") || !strings.Contains(compiledSrc, "interpreter.DecodeNodeJSON([]byte(") {
		t.Fatalf("expected interface signature default impl body to be preserved in rendered metadata")
	}
}

func TestCompilerNoFallbacksForLocalDefinitionConstraintsAndWhereClauses(t *testing.T) {
	tagIface := ast.NewInterfaceDefinition(
		ast.NewIdentifier("Tag"),
		[]*ast.FunctionSignature{
			ast.NewFunctionSignature(
				ast.NewIdentifier("tag"),
				[]*ast.FunctionParameter{
					ast.NewFunctionParameter(ast.NewIdentifier("self"), ast.Ty("Self")),
				},
				ast.Ty("String"),
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

	tagConstraint := testInterfaceConstraint(ast.Ty("Tag"))
	genericT := []*ast.GenericParameter{
		ast.NewGenericParameter(ast.NewIdentifier("T"), tagConstraint),
	}
	whereT := testWhereConstraint(ast.Ty("T"), tagConstraint)

	localStruct := ast.NewStructDefinition(
		ast.NewIdentifier("LocalStruct"),
		[]*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.Ty("T"), ast.NewIdentifier("value")),
		},
		ast.StructKindNamed,
		genericT,
		whereT,
		false,
	)
	localUnion := ast.NewUnionDefinition(
		ast.NewIdentifier("LocalUnion"),
		[]ast.TypeExpression{ast.Ty("nil"), ast.Ty("T")},
		genericT,
		whereT,
		false,
	)
	localSig := ast.NewFunctionSignature(
		ast.NewIdentifier("describe"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), ast.Ty("Self")),
		},
		ast.Ty("String"),
		genericT,
		whereT,
		nil,
	)
	localDefaultSig := ast.NewFunctionSignature(
		ast.NewIdentifier("with_default"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), ast.Ty("Self")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		ast.NewBlockExpression([]ast.Statement{ast.Int(9)}),
	)
	localIface := ast.NewInterfaceDefinition(
		ast.NewIdentifier("LocalIface"),
		[]*ast.FunctionSignature{localSig, localDefaultSig},
		genericT,
		ast.Ty("T"),
		whereT,
		nil,
		false,
	)
	mainFn := ast.NewFunctionDefinition(
		ast.NewIdentifier("main"),
		nil,
		ast.NewBlockExpression([]ast.Statement{localStruct, localUnion, localIface, ast.Int(1)}),
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.NewModule(
		[]ast.Statement{tagIface, mainFn},
		nil,
		ast.NewPackageStatement([]*ast.Identifier{ast.NewIdentifier("demo")}, false),
	)

	result, err := New(Options{PackageName: "compiled", RequireNoFallbacks: true}).Compile(testProgramFromModule("demo", module))
	if err != nil {
		t.Fatalf("compile local definitions with constraints/where under no-fallbacks: %v", err)
	}
	if len(result.Fallbacks) > 0 {
		t.Fatalf("expected no fallbacks, got %v", result.Fallbacks)
	}
	compiledSrc := combinedGeneratedSource(result)
	if !strings.Contains(compiledSrc, "ast.NewWhereClauseConstraint(ast.Ty(\"T\"), []*ast.InterfaceConstraint{ast.NewInterfaceConstraint(ast.Ty(\"Tag\"))})") {
		t.Fatalf("expected local definition rendering to preserve where-clause constraints")
	}
	if !strings.Contains(compiledSrc, "ast.NewFunctionSignature(ast.NewIdentifier(\"with_default\")") || !strings.Contains(compiledSrc, "interpreter.DecodeNodeJSON([]byte(") {
		t.Fatalf("expected local interface default impl body to be preserved in rendered metadata")
	}
	if strings.Contains(compiledSrc, "CallOriginal(\"demo.main\"") {
		t.Fatalf("expected local definition constraints/where path to stay compiled without call_original fallback")
	}
}
