package compiler

import (
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type nativeCallableInfo struct {
	Key                  string
	PackageName          string
	GoType               string
	TypeExpr             *ast.FunctionTypeExpression
	TypeString           string
	ParamGoTypes         []string
	ParamTypeExprs       []ast.TypeExpression
	ReturnGoType         string
	ReturnTypeExpr       ast.TypeExpression
	FromRuntimeHelper    string
	TryFromRuntimeHelper string
	FromRuntimePanic     string
	ToRuntimeHelper      string
	ToRuntimePanic       string
}

func (g *generator) nativeCallableInfoForGoType(goType string) *nativeCallableInfo {
	if g == nil || goType == "" || g.nativeCallables == nil {
		return nil
	}
	for _, info := range g.nativeCallables {
		if info != nil && info.GoType == goType {
			return info
		}
	}
	return nil
}

func (g *generator) sortedNativeCallableKeys() []string {
	if g == nil || len(g.nativeCallables) == 0 {
		return nil
	}
	keys := make([]string, 0, len(g.nativeCallables))
	for key := range g.nativeCallables {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func nativeCallableKey(paramGoTypes []string, returnGoType string) string {
	parts := make([]string, 0, len(paramGoTypes)+2)
	parts = append(parts, "fn")
	parts = append(parts, paramGoTypes...)
	parts = append(parts, "->", returnGoType)
	return strings.Join(parts, "|")
}

func nativeCallableToken(paramGoTypes []string, returnGoType string) string {
	parts := make([]string, 0, len(paramGoTypes)+2)
	parts = append(parts, "__able_fn")
	if len(paramGoTypes) == 0 {
		parts = append(parts, "void")
	} else {
		parts = append(parts, paramGoTypes...)
	}
	parts = append(parts, "to", returnGoType)
	return sanitizeIdent(strings.Join(parts, "_"))
}

func (g *generator) ensureNativeCallableInfo(pkgName string, expr *ast.FunctionTypeExpression) (*nativeCallableInfo, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	paramExprs := make([]ast.TypeExpression, 0, len(expr.ParamTypes))
	paramGoTypes := make([]string, 0, len(expr.ParamTypes))
	for _, paramExpr := range expr.ParamTypes {
		if paramExpr == nil {
			return nil, false
		}
		normalized := normalizeTypeExprForPackage(g, pkgName, paramExpr)
		paramPkg := g.resolvedTypeExprPackage(pkgName, normalized)
		goType, ok := NewTypeMapper(g, paramPkg).Map(normalized)
		goType, ok = g.recoverRepresentableCarrierType(paramPkg, normalized, goType)
		if !ok || goType == "" {
			return nil, false
		}
		paramExprs = append(paramExprs, normalized)
		paramGoTypes = append(paramGoTypes, goType)
	}
	if expr.ReturnType == nil {
		return nil, false
	}
	returnExpr := normalizeTypeExprForPackage(g, pkgName, expr.ReturnType)
	returnPkg := g.resolvedTypeExprPackage(pkgName, returnExpr)
	returnGoType, ok := NewTypeMapper(g, returnPkg).Map(returnExpr)
	returnGoType, ok = g.recoverRepresentableCarrierType(returnPkg, returnExpr, returnGoType)
	if !ok || returnGoType == "" {
		return nil, false
	}
	return g.ensureNativeCallableInfoFromSignatureInPackage(pkgName, paramExprs, paramGoTypes, returnExpr, returnGoType)
}

func (g *generator) ensureNativeCallableInfoFromSignature(paramExprs []ast.TypeExpression, paramGoTypes []string, returnExpr ast.TypeExpression, returnGoType string) (*nativeCallableInfo, bool) {
	return g.ensureNativeCallableInfoFromSignatureInPackage("", paramExprs, paramGoTypes, returnExpr, returnGoType)
}

func (g *generator) ensureNativeCallableInfoFromSignatureInPackage(pkgName string, paramExprs []ast.TypeExpression, paramGoTypes []string, returnExpr ast.TypeExpression, returnGoType string) (*nativeCallableInfo, bool) {
	if g == nil || returnGoType == "" {
		return nil, false
	}
	if len(paramExprs) != len(paramGoTypes) {
		expanded := make([]ast.TypeExpression, 0, len(paramGoTypes))
		for idx, goType := range paramGoTypes {
			if idx < len(paramExprs) && paramExprs[idx] != nil {
				expanded = append(expanded, paramExprs[idx])
				continue
			}
			expr, ok := g.typeExprForGoType(goType)
			if !ok {
				return nil, false
			}
			expanded = append(expanded, expr)
		}
		paramExprs = expanded
	}
	if returnExpr == nil {
		expr, ok := g.typeExprForGoType(returnGoType)
		if !ok {
			return nil, false
		}
		returnExpr = expr
	}
	key := nativeCallableKey(paramGoTypes, returnGoType)
	if info, ok := g.nativeCallables[key]; ok && info != nil {
		if info.PackageName == "" && strings.TrimSpace(pkgName) != "" {
			info.PackageName = strings.TrimSpace(pkgName)
			if info.TypeExpr != nil {
				info.TypeExpr = g.recordResolvedTypeExprPackage(info.TypeExpr, info.PackageName).(*ast.FunctionTypeExpression)
			}
		}
		return info, true
	}
	paramExprCopy := append([]ast.TypeExpression{}, paramExprs...)
	paramGoCopy := append([]string{}, paramGoTypes...)
	typeExpr := ast.NewFunctionTypeExpression(paramExprCopy, returnExpr)
	baseToken := nativeCallableToken(paramGoCopy, returnGoType)
	info := &nativeCallableInfo{
		Key:                  key,
		PackageName:          strings.TrimSpace(pkgName),
		GoType:               baseToken,
		TypeExpr:             g.recordResolvedTypeExprPackage(typeExpr, pkgName).(*ast.FunctionTypeExpression),
		TypeString:           typeExpressionToString(typeExpr),
		ParamGoTypes:         paramGoCopy,
		ParamTypeExprs:       paramExprCopy,
		ReturnGoType:         returnGoType,
		ReturnTypeExpr:       returnExpr,
		FromRuntimeHelper:    baseToken + "_from_runtime_value",
		TryFromRuntimeHelper: baseToken + "_try_from_runtime_value",
		FromRuntimePanic:     baseToken + "_from_runtime_value_or_panic",
		ToRuntimeHelper:      baseToken + "_to_runtime_value",
		ToRuntimePanic:       baseToken + "_to_runtime_value_or_panic",
	}
	g.nativeCallables[key] = info
	return info, true
}
