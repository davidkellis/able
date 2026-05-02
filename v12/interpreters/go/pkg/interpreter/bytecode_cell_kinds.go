package interpreter

import "able/interpreter-go/pkg/ast"

type bytecodeCellKind uint8

const (
	bytecodeCellKindValue bytecodeCellKind = iota
	bytecodeCellKindI32
	bytecodeCellKindBool
)

func bytecodeCellKindForSimpleTypeName(typeName string) bytecodeCellKind {
	switch typeName {
	case "i32":
		return bytecodeCellKindI32
	case "bool":
		return bytecodeCellKindBool
	default:
		return bytecodeCellKindValue
	}
}

func bytecodeCellKindForTypeExpr(typeExpr ast.TypeExpression) bytecodeCellKind {
	return bytecodeCellKindForSimpleTypeName(cachedSimpleTypeName(typeExpr))
}
