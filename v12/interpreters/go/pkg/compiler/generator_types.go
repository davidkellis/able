package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

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

func (g *generator) structBaseName(goType string) (string, bool) {
	if strings.HasPrefix(goType, "*") {
		goType = strings.TrimPrefix(goType, "*")
	}
	for _, info := range g.structs {
		if info != nil && info.GoName == goType {
			return info.GoName, true
		}
	}
	return "", false
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

func (g *generator) nativeIntegerWidenExpr(expr string, srcType string, targetType string) (string, bool) {
	if srcType == targetType {
		return expr, true
	}
	if !g.isIntegerType(srcType) || !g.isIntegerType(targetType) {
		return "", false
	}
	if g.isSignedIntegerType(srcType) != g.isSignedIntegerType(targetType) {
		return "", false
	}
	if g.intBits(srcType) > g.intBits(targetType) {
		return "", false
	}
	return fmt.Sprintf("%s(%s)", targetType, expr), true
}

func (g *generator) integerTypeSuffix(goType string) (string, bool) {
	switch goType {
	case "int8":
		return "i8", true
	case "int16":
		return "i16", true
	case "int32":
		return "i32", true
	case "int64":
		return "i64", true
	case "uint8":
		return "u8", true
	case "uint16":
		return "u16", true
	case "uint32":
		return "u32", true
	case "uint64":
		return "u64", true
	case "int":
		return "isize", true
	case "uint":
		return "usize", true
	default:
		return "", false
	}
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
	case ast.IntegerTypeI128:
		return "runtime.Value"
	case ast.IntegerTypeU8:
		return "uint8"
	case ast.IntegerTypeU16:
		return "uint16"
	case ast.IntegerTypeU32:
		return "uint32"
	case ast.IntegerTypeU64:
		return "uint64"
	case ast.IntegerTypeU128:
		return "runtime.Value"
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
	return g.mapTypeExpressionInPackage("", expr)
}

func (g *generator) mapTypeExpressionInPackage(pkgName string, expr ast.TypeExpression) (string, bool) {
	mapper := NewTypeMapper(g, pkgName)
	return mapper.Map(expr)
}

func (g *generator) interfaceTypeExpr(expr ast.TypeExpression) (ast.TypeExpression, bool) {
	if expr == nil {
		return nil, false
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return nil, false
		}
		if g.isInterfaceName(t.Name.Name) {
			return expr, true
		}
	case *ast.GenericTypeExpression:
		if t == nil {
			return nil, false
		}
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil {
			if g.isInterfaceName(base.Name.Name) {
				return expr, true
			}
		}
	}
	return nil, false
}

func (g *generator) isResultVoidTypeExpr(expr ast.TypeExpression) bool {
	res, ok := expr.(*ast.ResultTypeExpression)
	if !ok || res == nil || res.InnerType == nil {
		return false
	}
	inner, ok := res.InnerType.(*ast.SimpleTypeExpression)
	if !ok || inner == nil || inner.Name == nil {
		return false
	}
	return inner.Name.Name == "void" || inner.Name.Name == "Void"
}

func (g *generator) isInterfaceName(name string) bool {
	if name == "" || g == nil || g.interfaces == nil {
		return false
	}
	_, ok := g.interfaces[name]
	return ok
}

func (g *generator) renderTypeExpression(expr ast.TypeExpression) (string, bool) {
	if expr == nil {
		return "", false
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return "", false
		}
		return fmt.Sprintf("ast.Ty(%q)", t.Name.Name), true
	case *ast.GenericTypeExpression:
		if t == nil {
			return "", false
		}
		baseExpr, ok := g.renderTypeExpression(t.Base)
		if !ok {
			return "", false
		}
		args := make([]string, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			rendered, ok := g.renderTypeExpression(arg)
			if !ok {
				return "", false
			}
			args = append(args, rendered)
		}
		if len(args) == 0 {
			return fmt.Sprintf("ast.Gen(%s)", baseExpr), true
		}
		return fmt.Sprintf("ast.Gen(%s, %s)", baseExpr, strings.Join(args, ", ")), true
	case *ast.FunctionTypeExpression:
		if t == nil {
			return "", false
		}
		params := make([]string, 0, len(t.ParamTypes))
		for _, param := range t.ParamTypes {
			rendered, ok := g.renderTypeExpression(param)
			if !ok {
				return "", false
			}
			params = append(params, rendered)
		}
		ret, ok := g.renderTypeExpression(t.ReturnType)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("ast.FnType([]ast.TypeExpression{%s}, %s)", strings.Join(params, ", "), ret), true
	case *ast.NullableTypeExpression:
		if t == nil {
			return "", false
		}
		inner, ok := g.renderTypeExpression(t.InnerType)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("ast.Nullable(%s)", inner), true
	case *ast.ResultTypeExpression:
		if t == nil {
			return "", false
		}
		inner, ok := g.renderTypeExpression(t.InnerType)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("ast.Result(%s)", inner), true
	case *ast.UnionTypeExpression:
		if t == nil {
			return "", false
		}
		members := make([]string, 0, len(t.Members))
		for _, member := range t.Members {
			rendered, ok := g.renderTypeExpression(member)
			if !ok {
				return "", false
			}
			members = append(members, rendered)
		}
		return fmt.Sprintf("ast.UnionT(%s)", strings.Join(members, ", ")), true
	case *ast.WildcardTypeExpression:
		return "ast.WildT()", true
	default:
		return "", false
	}
}

func typeExpressionToString(expr ast.TypeExpression) string {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return "<?>"
		}
		return t.Name.Name
	case *ast.GenericTypeExpression:
		base := typeExpressionToString(t.Base)
		args := make([]string, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			args = append(args, typeExpressionToString(arg))
		}
		return fmt.Sprintf("%s<%s>", base, strings.Join(args, ", "))
	case *ast.NullableTypeExpression:
		return typeExpressionToString(t.InnerType) + "?"
	case *ast.FunctionTypeExpression:
		parts := make([]string, 0, len(t.ParamTypes))
		for _, p := range t.ParamTypes {
			parts = append(parts, typeExpressionToString(p))
		}
		return fmt.Sprintf("fn(%s) -> %s", strings.Join(parts, ", "), typeExpressionToString(t.ReturnType))
	case *ast.UnionTypeExpression:
		parts := make([]string, 0, len(t.Members))
		for _, member := range t.Members {
			parts = append(parts, typeExpressionToString(member))
		}
		return strings.Join(parts, " | ")
	default:
		return "<?>"
	}
}

func typeNameFromGoType(goType string) string {
	switch goType {
	case "bool":
		return "bool"
	case "string":
		return "String"
	case "rune":
		return "char"
	case "int8":
		return "i8"
	case "int16":
		return "i16"
	case "int32":
		return "i32"
	case "int64":
		return "i64"
	case "uint8":
		return "u8"
	case "uint16":
		return "u16"
	case "uint32":
		return "u32"
	case "uint64":
		return "u64"
	case "int":
		return "isize"
	case "uint":
		return "usize"
	case "float32":
		return "f32"
	case "float64":
		return "f64"
	case "struct{}":
		return "void"
	}
	if strings.HasPrefix(goType, "*") {
		return strings.TrimPrefix(goType, "*")
	}
	return goType
}

func (g *generator) hasOptionalLastParam(info *functionInfo) bool {
	if info == nil || info.Definition == nil {
		return false
	}
	params := info.Definition.Params
	if len(params) == 0 {
		return false
	}
	last := params[len(params)-1]
	if last == nil || last.ParamType == nil {
		return false
	}
	_, ok := last.ParamType.(*ast.NullableTypeExpression)
	return ok
}
