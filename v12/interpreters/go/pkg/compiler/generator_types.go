package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) isVoidType(goType string) bool {
	return goType == "struct{}"
}

func (g *generator) isStringType(goType string) bool {
	return goType == "string"
}

func (g *generator) isIntegerType(goType string) bool {
	switch goType {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	}
	return false
}

func (g *generator) isSignedIntegerType(goType string) bool {
	switch goType {
	case "int", "int8", "int16", "int32", "int64":
		return true
	}
	return false
}

func (g *generator) isUnsignedIntegerType(goType string) bool {
	switch goType {
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	}
	return false
}

func (g *generator) isFloatType(goType string) bool {
	return goType == "float32" || goType == "float64"
}

func (g *generator) isNumericType(goType string) bool {
	return g.isIntegerType(goType) || g.isFloatType(goType)
}

func (g *generator) isEqualityComparable(goType string) bool {
	return g.isNumericType(goType) || g.isStringType(goType) || goType == "bool" || goType == "rune"
}

func (g *generator) isOrderedComparable(goType string) bool {
	return g.isNumericType(goType) || g.isStringType(goType) || goType == "rune"
}

func (g *generator) intBits(goType string) int {
	switch goType {
	case "int8", "uint8":
		return 8
	case "int16", "uint16":
		return 16
	case "int32", "uint32":
		return 32
	case "int64", "uint64":
		return 64
	}
	return 64
}

func (g *generator) isUntypedNumericLiteral(expr ast.Expression) bool {
	switch lit := expr.(type) {
	case *ast.IntegerLiteral:
		return lit != nil && lit.IntegerType == nil
	case *ast.FloatLiteral:
		return lit != nil && lit.FloatType == nil
	default:
		return false
	}
}

func (g *generator) inferIntegerLiteralType(lit *ast.IntegerLiteral) string {
	if lit == nil || lit.IntegerType == nil {
		return "int32"
	}
	switch *lit.IntegerType {
	case ast.IntegerTypeI8:
		return "int8"
	case ast.IntegerTypeI16:
		return "int16"
	case ast.IntegerTypeI32:
		return "int32"
	case ast.IntegerTypeI64:
		return "int64"
	case ast.IntegerTypeU8:
		return "uint8"
	case ast.IntegerTypeU16:
		return "uint16"
	case ast.IntegerTypeU32:
		return "uint32"
	case ast.IntegerTypeU64:
		return "uint64"
	default:
		return "int32"
	}
}

func (g *generator) inferFloatLiteralType(lit *ast.FloatLiteral) string {
	if lit == nil || lit.FloatType == nil {
		return "float64"
	}
	switch *lit.FloatType {
	case ast.FloatTypeF32:
		return "float32"
	case ast.FloatTypeF64:
		return "float64"
	default:
		return "float64"
	}
}

func (g *generator) mapTypeExpression(expr ast.TypeExpression) (string, bool) {
	mapper := NewTypeMapper(g.structs)
	return mapper.Map(expr)
}
