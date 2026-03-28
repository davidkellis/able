package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) inferMemberAccessTypeExpr(ctx *compileContext, expr *ast.MemberAccessExpression) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || expr == nil || expr.Object == nil {
		return nil, false
	}
	objectTypeExpr, ok := g.inferExpressionTypeExpr(ctx, expr.Object, "")
	if !ok || objectTypeExpr == nil {
		return nil, false
	}
	info, ok := g.structInfoForTypeExpr(ctx.packageName, objectTypeExpr)
	if !ok || info == nil {
		return nil, false
	}
	field, ok := g.structFieldForMember(info, expr.Member)
	if !ok || field == nil || field.TypeExpr == nil {
		return nil, false
	}
	fieldTypeExpr := g.inferStructFieldTypeExpr(ctx, objectTypeExpr, info, field)
	if fieldTypeExpr == nil {
		return nil, false
	}
	fieldTypeExpr = g.wrapSafeNavigationTypeExpr(ctx, objectTypeExpr, expr.Safe, fieldTypeExpr)
	return fieldTypeExpr, true
}

func (g *generator) wrapSafeNavigationTypeExpr(ctx *compileContext, objectTypeExpr ast.TypeExpression, safe bool, valueTypeExpr ast.TypeExpression) ast.TypeExpression {
	if g == nil || ctx == nil || !safe || valueTypeExpr == nil || objectTypeExpr == nil {
		return valueTypeExpr
	}
	if _, ok := valueTypeExpr.(*ast.NullableTypeExpression); ok {
		return valueTypeExpr
	}
	objectGoType, ok := g.lowerCarrierType(ctx, objectTypeExpr)
	if !ok || !g.goTypeHasNilZeroValue(objectGoType) {
		return valueTypeExpr
	}
	return ast.NewNullableTypeExpression(valueTypeExpr)
}

func (g *generator) inferStructFieldTypeExpr(ctx *compileContext, objectTypeExpr ast.TypeExpression, info *structInfo, field *fieldInfo) ast.TypeExpression {
	if g == nil || ctx == nil || info == nil || field == nil || field.TypeExpr == nil {
		return nil
	}
	fieldTypeExpr := g.lowerNormalizedTypeExpr(ctx, field.TypeExpr)
	if info.Specialized || info.Node == nil || len(info.Node.GenericParams) == 0 || objectTypeExpr == nil {
		return fieldTypeExpr
	}
	objectTypeExpr = g.lowerNormalizedTypeExpr(ctx, objectTypeExpr)
	generic, ok := objectTypeExpr.(*ast.GenericTypeExpression)
	if !ok || generic == nil || len(generic.Arguments) != len(info.Node.GenericParams) {
		return fieldTypeExpr
	}
	bindings := make(map[string]ast.TypeExpression, len(info.Node.GenericParams))
	for idx, gp := range info.Node.GenericParams {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" || generic.Arguments[idx] == nil {
			return fieldTypeExpr
		}
		bindings[gp.Name.Name] = normalizeTypeExprForPackage(g, info.Package, generic.Arguments[idx])
	}
	return normalizeTypeExprForPackage(g, ctx.packageName, substituteTypeParams(field.TypeExpr, bindings))
}

func (g *generator) inferStructLiteralGenericTypeExpr(ctx *compileContext, lit *ast.StructLiteral) ast.TypeExpression {
	if g == nil || ctx == nil || lit == nil || lit.StructType == nil || lit.StructType.Name == "" {
		return nil
	}
	info, ok := g.structInfoForTypeName(ctx.packageName, lit.StructType.Name)
	if !ok || info == nil || info.Node == nil || len(info.Node.GenericParams) == 0 {
		return nil
	}
	genericNames := genericParamNameSet(info.Node.GenericParams)
	bindings := make(map[string]ast.TypeExpression)
	if lit.IsPositional {
		for idx, field := range lit.Fields {
			if field == nil || field.Value == nil || idx >= len(info.Fields) {
				continue
			}
			g.inferStructLiteralGenericFieldBinding(ctx, info.Package, info.Fields[idx].TypeExpr, field.Value, genericNames, bindings)
		}
	} else {
		for _, field := range lit.Fields {
			if field == nil {
				continue
			}
			fieldName := ""
			if field.Name != nil {
				fieldName = field.Name.Name
			}
			valueExpr := field.Value
			if fieldName == "" && field.IsShorthand {
				if ident, ok := field.Value.(*ast.Identifier); ok && ident != nil {
					fieldName = ident.Name
					valueExpr = ast.NewIdentifier(fieldName)
				}
			}
			if fieldName == "" || valueExpr == nil {
				continue
			}
			fieldInfo := g.fieldInfo(info, fieldName)
			if fieldInfo == nil {
				continue
			}
			g.inferStructLiteralGenericFieldBinding(ctx, info.Package, fieldInfo.TypeExpr, valueExpr, genericNames, bindings)
		}
	}
	bindings = g.normalizeConcreteTypeBindings(info.Package, bindings, genericNames)
	if len(bindings) == 0 {
		return nil
	}
	args := make([]ast.TypeExpression, 0, len(info.Node.GenericParams))
	for _, gp := range info.Node.GenericParams {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			return nil
		}
		arg := bindings[gp.Name.Name]
		if arg == nil {
			return nil
		}
		args = append(args, normalizeTypeExprForPackage(g, info.Package, arg))
	}
	return normalizeTypeExprForPackage(g, ctx.packageName, ast.NewGenericTypeExpression(ast.Ty(lit.StructType.Name), args))
}

func (g *generator) inferStructLiteralGenericFieldBinding(ctx *compileContext, pkgName string, template ast.TypeExpression, value ast.Expression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) {
	if g == nil || ctx == nil || template == nil || value == nil {
		return
	}
	actual, ok := g.inferExpressionTypeExpr(ctx, value, "")
	if !ok || actual == nil {
		return
	}
	g.specializedTypeTemplateMatches(pkgName, template, actual, genericNames, bindings, make(map[string]struct{}))
}
