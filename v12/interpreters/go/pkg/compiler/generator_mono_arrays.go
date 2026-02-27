package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type monoArrayElemKind uint8

const (
	monoArrayElemKindUnknown monoArrayElemKind = iota
	monoArrayElemKindI32
	monoArrayElemKindI64
	monoArrayElemKindBool
	monoArrayElemKindU8
)

func (g *generator) monoArraysEnabled() bool {
	return g != nil && g.opts.ExperimentalMonoArrays
}

func (g *generator) monoArrayElemGoType(kind monoArrayElemKind) string {
	switch kind {
	case monoArrayElemKindI32:
		return "int32"
	case monoArrayElemKindI64:
		return "int64"
	case monoArrayElemKindBool:
		return "bool"
	case monoArrayElemKindU8:
		return "uint8"
	default:
		return ""
	}
}

func (g *generator) monoArrayNewWithCapacityExpr(kind monoArrayElemKind, capacityExpr string) (string, bool) {
	switch kind {
	case monoArrayElemKindI32:
		return fmt.Sprintf("runtime.ArrayStoreMonoNewWithCapacityI32(%s)", capacityExpr), true
	case monoArrayElemKindI64:
		return fmt.Sprintf("runtime.ArrayStoreMonoNewWithCapacityI64(%s)", capacityExpr), true
	case monoArrayElemKindBool:
		return fmt.Sprintf("runtime.ArrayStoreMonoNewWithCapacityBool(%s)", capacityExpr), true
	case monoArrayElemKindU8:
		return fmt.Sprintf("runtime.ArrayStoreMonoNewWithCapacityU8(%s)", capacityExpr), true
	default:
		return "", false
	}
}

func (g *generator) monoArrayReadExpr(kind monoArrayElemKind, handleExpr string, indexExpr string) (string, string, bool) {
	switch kind {
	case monoArrayElemKindI32:
		return fmt.Sprintf("runtime.ArrayStoreMonoReadI32(%s, %s)", handleExpr, indexExpr), "int32", true
	case monoArrayElemKindI64:
		return fmt.Sprintf("runtime.ArrayStoreMonoReadI64(%s, %s)", handleExpr, indexExpr), "int64", true
	case monoArrayElemKindBool:
		return fmt.Sprintf("runtime.ArrayStoreMonoReadBool(%s, %s)", handleExpr, indexExpr), "bool", true
	case monoArrayElemKindU8:
		return fmt.Sprintf("runtime.ArrayStoreMonoReadU8(%s, %s)", handleExpr, indexExpr), "uint8", true
	default:
		return "", "", false
	}
}

func (g *generator) monoArrayWriteExpr(kind monoArrayElemKind, handleExpr string, indexExpr string, valueExpr string) (string, bool) {
	switch kind {
	case monoArrayElemKindI32:
		return fmt.Sprintf("runtime.ArrayStoreMonoWriteI32(%s, %s, %s)", handleExpr, indexExpr, valueExpr), true
	case monoArrayElemKindI64:
		return fmt.Sprintf("runtime.ArrayStoreMonoWriteI64(%s, %s, %s)", handleExpr, indexExpr, valueExpr), true
	case monoArrayElemKindBool:
		return fmt.Sprintf("runtime.ArrayStoreMonoWriteBool(%s, %s, %s)", handleExpr, indexExpr, valueExpr), true
	case monoArrayElemKindU8:
		return fmt.Sprintf("runtime.ArrayStoreMonoWriteU8(%s, %s, %s)", handleExpr, indexExpr, valueExpr), true
	default:
		return "", false
	}
}

func (g *generator) monoArrayKindFromGoType(goType string) (monoArrayElemKind, bool) {
	switch goType {
	case "int32":
		return monoArrayElemKindI32, true
	case "int64":
		return monoArrayElemKindI64, true
	case "bool":
		return monoArrayElemKindBool, true
	case "uint8":
		return monoArrayElemKindU8, true
	default:
		return monoArrayElemKindUnknown, false
	}
}

func (g *generator) monoArrayKindFromElementTypeExpr(pkgName string, expr ast.TypeExpression) (monoArrayElemKind, bool) {
	if expr == nil {
		return monoArrayElemKindUnknown, false
	}
	expanded := g.expandTypeAliasForPackage(pkgName, expr)
	simple, ok := expanded.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return monoArrayElemKindUnknown, false
	}
	switch strings.TrimSpace(simple.Name.Name) {
	case "i32":
		return monoArrayElemKindI32, true
	case "i64":
		return monoArrayElemKindI64, true
	case "bool", "Bool":
		return monoArrayElemKindBool, true
	case "u8":
		return monoArrayElemKindU8, true
	default:
		return monoArrayElemKindUnknown, false
	}
}

func (g *generator) monoArrayKindFromArrayTypeExpr(pkgName string, expr ast.TypeExpression) (monoArrayElemKind, bool) {
	if expr == nil {
		return monoArrayElemKindUnknown, false
	}
	expanded := g.expandTypeAliasForPackage(pkgName, expr)
	generic, ok := expanded.(*ast.GenericTypeExpression)
	if !ok || generic == nil || len(generic.Arguments) != 1 {
		return monoArrayElemKindUnknown, false
	}
	baseSimple, ok := generic.Base.(*ast.SimpleTypeExpression)
	if !ok || baseSimple == nil || baseSimple.Name == nil {
		return monoArrayElemKindUnknown, false
	}
	if strings.TrimSpace(baseSimple.Name.Name) != "Array" {
		return monoArrayElemKindUnknown, false
	}
	return g.monoArrayKindFromElementTypeExpr(pkgName, generic.Arguments[0])
}

func (g *generator) monoArrayKindForObject(ctx *compileContext, object ast.Expression, objectType string) (monoArrayElemKind, bool) {
	if !g.monoArraysEnabled() || ctx == nil || !g.isArrayStructType(objectType) || object == nil {
		return monoArrayElemKindUnknown, false
	}
	ident, ok := object.(*ast.Identifier)
	if !ok || ident == nil || ident.Name == "" {
		return monoArrayElemKindUnknown, false
	}
	param, ok := ctx.lookup(ident.Name)
	if !ok || param.TypeExpr == nil {
		return monoArrayElemKindUnknown, false
	}
	return g.monoArrayKindFromArrayTypeExpr(ctx.packageName, param.TypeExpr)
}

func (g *generator) monoArrayKindForLiteral(ctx *compileContext, elementTypes []string) (monoArrayElemKind, bool) {
	if !g.monoArraysEnabled() {
		return monoArrayElemKindUnknown, false
	}
	if ctx != nil && ctx.expectedTypeExpr != nil {
		if kind, ok := g.monoArrayKindFromArrayTypeExpr(ctx.packageName, ctx.expectedTypeExpr); ok {
			return kind, true
		}
	}
	if len(elementTypes) == 0 {
		return monoArrayElemKindUnknown, false
	}
	var inferred monoArrayElemKind
	for idx, goType := range elementTypes {
		kind, ok := g.monoArrayKindFromGoType(goType)
		if !ok {
			return monoArrayElemKindUnknown, false
		}
		if idx == 0 {
			inferred = kind
			continue
		}
		if inferred != kind {
			return monoArrayElemKindUnknown, false
		}
	}
	return inferred, true
}

func (g *generator) coerceExprToGoType(expr string, exprType string, targetType string) (string, bool) {
	if exprType == targetType {
		return expr, true
	}
	if exprType == "runtime.Value" {
		return g.expectRuntimeValueExpr(expr, targetType)
	}
	valueExpr, ok := g.runtimeValueExpr(expr, exprType)
	if !ok {
		return "", false
	}
	return g.expectRuntimeValueExpr(valueExpr, targetType)
}
