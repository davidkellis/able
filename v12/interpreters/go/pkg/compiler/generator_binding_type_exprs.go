package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) reconcileConcreteBindingTypeExpr(ctx *compileContext, goType string, typeExpr ast.TypeExpression) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || typeExpr == nil || goType == "" || goType == "runtime.Value" || goType == "any" || g.isVoidType(goType) {
		return nil, false
	}
	carrierExpr, ok := g.typeExprForGoType(goType)
	if !ok || carrierExpr == nil {
		return nil, false
	}
	boundExpr := g.lowerNormalizedTypeExpr(ctx, typeExpr)
	carrierExpr = g.lowerNormalizedTypeExpr(ctx, carrierExpr)
	if boundExpr == nil || carrierExpr == nil {
		return nil, false
	}
	boundKey := normalizeTypeExprIdentityKey(g, ctx.packageName, boundExpr)
	carrierKey := normalizeTypeExprIdentityKey(g, ctx.packageName, carrierExpr)
	if boundKey != "" && carrierKey != "" && boundKey == carrierKey {
		return nil, false
	}
	if carrierKey == "" && typeExpressionToString(boundExpr) == typeExpressionToString(carrierExpr) {
		return nil, false
	}
	if _, ok := g.interfaceTypeExpr(boundExpr); ok {
		return nil, false
	}
	if boundBase, boundOK := typeExprBaseName(boundExpr); boundOK {
		if carrierBase, carrierOK := typeExprBaseName(carrierExpr); carrierOK && boundBase == carrierBase {
			if boundKey != carrierKey && g.typeExprFullyBound(ctx.packageName, carrierExpr) {
				return carrierExpr, true
			}
			return nil, false
		}
	}
	if !g.typeExprFullyBound(ctx.packageName, carrierExpr) {
		return nil, false
	}
	boundType, ok := g.joinCarrierTypeFromTypeExpr(ctx, boundExpr)
	if !ok || boundType == "" || boundType == goType || boundType == "runtime.Value" || boundType == "any" || g.isVoidType(boundType) {
		return nil, false
	}
	if g.nativeUnionInfoForGoType(boundType) != nil {
		return carrierExpr, true
	}
	if innerType, nullable := g.nativeNullableValueInnerType(boundType); nullable {
		if innerType == goType || g.canCoerceStaticExpr(innerType, goType) {
			return carrierExpr, true
		}
		if innerType == "runtime.ErrorValue" && g.isNativeErrorCarrierType(goType) {
			return carrierExpr, true
		}
	}
	if boundType == "runtime.ErrorValue" && g.isNativeErrorCarrierType(goType) {
		return carrierExpr, true
	}
	return nil, false
}
