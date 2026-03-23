package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func genericStructTypeExprForDefinition(def *ast.StructDefinition) ast.TypeExpression {
	if def == nil || def.ID == nil || def.ID.Name == "" {
		return nil
	}
	if len(def.GenericParams) == 0 {
		return ast.Ty(def.ID.Name)
	}
	args := make([]ast.TypeExpression, 0, len(def.GenericParams))
	for _, gp := range def.GenericParams {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			return ast.Ty(def.ID.Name)
		}
		args = append(args, ast.Ty(gp.Name.Name))
	}
	return ast.NewGenericTypeExpression(ast.Ty(def.ID.Name), args)
}

func (g *generator) allStructInfos() []*structInfo {
	if g == nil {
		return nil
	}
	infos := make([]*structInfo, 0, len(g.structs)+len(g.specializedStructs))
	for _, info := range g.structs {
		if info != nil {
			infos = append(infos, info)
		}
	}
	for _, info := range g.specializedStructs {
		if info != nil {
			infos = append(infos, info)
		}
	}
	return infos
}

func (g *generator) sortedAllStructInfos() []*structInfo {
	infos := g.allStructInfos()
	sort.Slice(infos, func(i, j int) bool {
		left := infos[i]
		right := infos[j]
		if left == nil || right == nil {
			return left != nil
		}
		if left.GoName != right.GoName {
			return left.GoName < right.GoName
		}
		if left.Package != right.Package {
			return left.Package < right.Package
		}
		return left.Name < right.Name
	})
	return infos
}

func (g *generator) structInfoForTypeExpr(pkgName string, expr ast.TypeExpression) (*structInfo, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return nil, false
		}
		return g.structInfoForTypeName(pkgName, t.Name.Name)
	case *ast.GenericTypeExpression:
		if t == nil {
			return nil, false
		}
		baseName, ok := typeExprBaseName(t)
		if !ok || baseName == "" {
			return nil, false
		}
		if info, ok := g.ensureSpecializedStructInfo(pkgName, expr); ok && info != nil {
			return info, true
		}
		return g.structInfoForTypeName(pkgName, baseName)
	default:
		return nil, false
	}
}

func (g *generator) nativeStructCarrierTypeForExpr(pkgName string, expr ast.TypeExpression) (string, bool) {
	info, ok := g.structInfoForTypeExpr(pkgName, expr)
	if !ok || info == nil {
		return "", false
	}
	return "*" + info.GoName, true
}

func (g *generator) ensureSpecializedStructInfo(pkgName string, expr ast.TypeExpression) (*structInfo, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
	generic, ok := expr.(*ast.GenericTypeExpression)
	if !ok || generic == nil {
		return nil, false
	}
	baseName, ok := typeExprBaseName(generic.Base)
	if !ok || baseName == "" || baseName == "Array" {
		return nil, false
	}
	baseInfo, ok := g.structInfoForTypeName(pkgName, baseName)
	if !ok || baseInfo == nil || baseInfo.Node == nil || len(baseInfo.Node.GenericParams) == 0 {
		return nil, false
	}
	if len(baseInfo.Node.GenericParams) != len(generic.Arguments) {
		return nil, false
	}
	if !g.typeExprIsConcreteInPackage(baseInfo.Package, expr) {
		return nil, false
	}
	key := specializedStructKey(g, baseInfo.Package, expr)
	if key == "" {
		return nil, false
	}
	if existing := g.specializedStructs[key]; existing != nil {
		return existing, true
	}
	suffix := specializedStructSuffix(g, baseInfo.Package, generic.Arguments)
	if suffix == "" {
		return nil, false
	}
	info := &structInfo{
		Name:        baseInfo.Name,
		Package:     baseInfo.Package,
		GoName:      g.mangler.unique(exportIdent(baseInfo.Name) + "_" + suffix),
		TypeExpr:    expr,
		Kind:        baseInfo.Kind,
		Node:        baseInfo.Node,
		Specialized: true,
	}
	g.specializedStructs[key] = info
	bindings := make(map[string]ast.TypeExpression, len(baseInfo.Node.GenericParams))
	for idx, gp := range baseInfo.Node.GenericParams {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" || generic.Arguments[idx] == nil {
			info.Supported = false
			return info, false
		}
		bindings[gp.Name.Name] = normalizeTypeExprForPackage(g, baseInfo.Package, generic.Arguments[idx])
	}
	mapper := NewTypeMapper(g, baseInfo.Package)
	fields := make([]fieldInfo, 0, len(baseInfo.Node.Fields))
	supported := true
	for idx, field := range baseInfo.Node.Fields {
		fieldName := ""
		if field.Name != nil {
			fieldName = field.Name.Name
		}
		if fieldName == "" {
			fieldName = fmt.Sprintf("field_%d", idx+1)
		}
		substituted := normalizeTypeExprForPackage(g, baseInfo.Package, substituteTypeParams(field.FieldType, bindings))
		goType, ok := mapper.Map(substituted)
		if !ok || goType == "" {
			supported = false
		}
		fields = append(fields, fieldInfo{
			Name:      fieldName,
			GoName:    exportIdent(fieldName),
			GoType:    goType,
			TypeExpr:  substituted,
			Supported: ok,
		})
	}
	info.Fields = fields
	info.Supported = supported
	return info, supported
}

func specializedStructKey(g *generator, pkgName string, expr ast.TypeExpression) string {
	if g == nil || expr == nil {
		return ""
	}
	return strings.TrimSpace(pkgName) + "::" + normalizeTypeExprString(g, pkgName, expr)
}

func specializedStructSuffix(g *generator, pkgName string, args []ast.TypeExpression) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == nil {
			return ""
		}
		part := sanitizeIdent(normalizeTypeExprString(g, pkgName, arg))
		part = strings.Trim(part, "_")
		if part == "" {
			return ""
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, "_")
}

func (g *generator) typeExprIsConcreteInPackage(pkgName string, expr ast.TypeExpression) bool {
	return g.typeExprIsConcreteInPackageSeen(pkgName, expr, nil, nil)
}

func (g *generator) typeExprIsConcreteInPackageSeen(pkgName string, expr ast.TypeExpression, visiting map[string]struct{}, cache map[string]bool) bool {
	if g == nil || expr == nil {
		return false
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
	if expanded := g.expandTypeAliasForPackage(pkgName, expr); expanded != nil {
		expr = expanded
	}
	key := strings.TrimSpace(pkgName) + "::" + typeExpressionToString(expr)
	if cache != nil {
		if result, ok := cache[key]; ok {
			return result
		}
	}
	if visiting != nil {
		if _, ok := visiting[key]; ok {
			return true
		}
	} else {
		visiting = make(map[string]struct{})
	}
	if cache == nil {
		cache = make(map[string]bool)
	}
	visiting[key] = struct{}{}
	defer delete(visiting, key)
	result := false
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || t.Name.Name == "" {
			break
		}
		name := t.Name.Name
		if name == "any" || name == "Error" || isBuiltinMappedType(name) {
			result = true
			break
		}
		if _, ok := g.structInfoForTypeName(pkgName, name); ok {
			result = true
			break
		}
		if _, _, _, _, ok := interfaceExprInfo(g, pkgName, expr); ok {
			result = true
			break
		}
		if _, _, ok := g.expandedUnionMembersInPackage(pkgName, expr); ok {
			result = true
			break
		}
	case *ast.GenericTypeExpression:
		if t == nil || !g.typeExprIsConcreteInPackageSeen(pkgName, t.Base, visiting, cache) {
			break
		}
		result = true
		for _, arg := range t.Arguments {
			if !g.typeExprIsConcreteInPackageSeen(pkgName, arg, visiting, cache) {
				result = false
				break
			}
		}
	case *ast.NullableTypeExpression:
		result = g.typeExprIsConcreteInPackageSeen(pkgName, t.InnerType, visiting, cache)
	case *ast.ResultTypeExpression:
		result = g.typeExprIsConcreteInPackageSeen(pkgName, t.InnerType, visiting, cache)
	case *ast.UnionTypeExpression:
		result = true
		for _, member := range t.Members {
			if !g.typeExprIsConcreteInPackageSeen(pkgName, member, visiting, cache) {
				result = false
				break
			}
		}
	case *ast.FunctionTypeExpression:
		if !g.typeExprIsConcreteInPackageSeen(pkgName, t.ReturnType, visiting, cache) {
			break
		}
		result = true
		for _, param := range t.ParamTypes {
			if !g.typeExprIsConcreteInPackageSeen(pkgName, param, visiting, cache) {
				result = false
				break
			}
		}
	case *ast.WildcardTypeExpression:
	default:
	}
	cache[key] = result
	return result
}
