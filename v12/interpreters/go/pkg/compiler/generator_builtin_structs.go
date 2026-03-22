package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) ensureBuiltinArrayStruct() {
	if g == nil {
		return
	}
	if info, ok := g.structInfoByNameUnique("Array"); ok && info != nil {
		return
	}
	arrayNode := ast.NewStructDefinition(
		ast.NewIdentifier("Array"),
		[]*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("i32")), ast.NewIdentifier("length")),
			ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("i32")), ast.NewIdentifier("capacity")),
			ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("ArrayHandle")), ast.NewIdentifier("storage_handle")),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.NewGenericParameter(ast.NewIdentifier("T"), nil)},
		nil,
		false,
	)
	g.structs["Array"] = &structInfo{
		Name:    "Array",
		Package: "",
		GoName:  g.mangler.unique(exportIdent("Array")),
		Kind:    ast.StructKindNamed,
		Node:    arrayNode,
	}
}

func (g *generator) ensureBuiltinDivModStruct() {
	if g == nil {
		return
	}
	if info, ok := g.structInfoByNameUnique("DivMod"); ok && info != nil {
		return
	}
	divModNode := ast.NewStructDefinition(
		ast.NewIdentifier("DivMod"),
		[]*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("T")), ast.NewIdentifier("quotient")),
			ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("T")), ast.NewIdentifier("remainder")),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.NewGenericParameter(ast.NewIdentifier("T"), nil)},
		nil,
		false,
	)
	g.structs["DivMod"] = &structInfo{
		Name:    "DivMod",
		Package: "",
		GoName:  g.mangler.unique(exportIdent("DivMod")),
		Kind:    ast.StructKindNamed,
		Node:    divModNode,
	}
}
