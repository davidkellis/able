package driver

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCombinePackagePreservesPerFileOriginsWhenSkippingKnownSubtrees(t *testing.T) {
	firstName := ast.NewIdentifier("first")
	firstFn := ast.NewFunctionDefinition(firstName, nil, ast.NewBlockExpression(nil), nil, nil, nil, false, false)
	firstModule := ast.NewModule([]ast.Statement{firstFn}, nil, ast.NewPackageStatement([]*ast.Identifier{ast.NewIdentifier("sample")}, false))

	secondName := ast.NewIdentifier("second")
	secondFn := ast.NewFunctionDefinition(secondName, nil, ast.NewBlockExpression(nil), nil, nil, nil, false, false)
	secondModule := ast.NewModule([]ast.Statement{secondFn}, nil, ast.NewPackageStatement([]*ast.Identifier{ast.NewIdentifier("sample")}, false))
	firstOrigins := make(map[ast.Node]string)
	ast.AnnotateOrigins(firstModule, "a.able", firstOrigins)
	secondOrigins := make(map[ast.Node]string)
	ast.AnnotateOrigins(secondModule, "b.able", secondOrigins)

	combined, err := combinePackage("sample", []*fileModule{
		{path: "b.able", packageName: "sample", ast: secondModule, origins: secondOrigins},
		{path: "a.able", packageName: "sample", ast: firstModule, origins: firstOrigins},
	})
	if err != nil {
		t.Fatalf("combinePackage: %v", err)
	}

	for _, node := range []ast.Node{firstModule, firstModule.Package, firstModule.Package.NamePath[0], firstFn, firstName} {
		if got := combined.NodeOrigins[node]; got != "a.able" {
			t.Fatalf("first-file %T origin = %q, want a.able", node, got)
		}
	}
	for _, node := range []ast.Node{secondModule, secondModule.Package, secondModule.Package.NamePath[0], secondFn, secondName} {
		if got := combined.NodeOrigins[node]; got != "b.able" {
			t.Fatalf("second-file %T origin = %q, want b.able", node, got)
		}
	}
	for _, node := range []ast.Node{combined.AST, combined.AST.Package, combined.AST.Package.NamePath[0]} {
		if got := combined.NodeOrigins[node]; got != "a.able" {
			t.Fatalf("combined %T origin = %q, want a.able", node, got)
		}
	}
}
