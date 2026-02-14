package compiler

import (
	"fmt"
	"sort"
	"strings"
)

func (g *generator) sortedStructInfosForPackage(pkgName string) []*structInfo {
	if g == nil || pkgName == "" {
		return nil
	}
	names := make([]string, 0, len(g.structs))
	for name, info := range g.structs {
		if info == nil || info.Package != pkgName {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)
	infos := make([]*structInfo, 0, len(names))
	for _, name := range names {
		if info := g.structs[name]; info != nil {
			infos = append(infos, info)
		}
	}
	return infos
}

func (g *generator) renderStructDefinitionExpr(info *structInfo) (string, bool) {
	if g == nil || info == nil || info.Node == nil || info.Name == "" {
		return "", false
	}
	fieldExprs := make([]string, 0, len(info.Node.Fields))
	for _, field := range info.Node.Fields {
		fieldTypeExpr := "ast.WildT()"
		if field != nil && field.FieldType != nil {
			if rendered, ok := g.renderTypeExpression(field.FieldType); ok {
				fieldTypeExpr = rendered
			}
		}
		fieldNameExpr := "nil"
		if field != nil && field.Name != nil && field.Name.Name != "" {
			fieldNameExpr = fmt.Sprintf("ast.NewIdentifier(%q)", field.Name.Name)
		}
		fieldExprs = append(fieldExprs, fmt.Sprintf("ast.NewStructFieldDefinition(%s, %s)", fieldTypeExpr, fieldNameExpr))
	}
	fieldsExpr := "nil"
	if len(fieldExprs) > 0 {
		fieldsExpr = "[]*ast.StructFieldDefinition{" + strings.Join(fieldExprs, ", ") + "}"
	}
	genericExpr := "nil"
	if len(info.Node.GenericParams) > 0 {
		parts := make([]string, 0, len(info.Node.GenericParams))
		for _, param := range info.Node.GenericParams {
			if param == nil || param.Name == nil || strings.TrimSpace(param.Name.Name) == "" {
				continue
			}
			parts = append(parts, fmt.Sprintf("ast.NewGenericParameter(ast.NewIdentifier(%q), nil)", param.Name.Name))
		}
		if len(parts) > 0 {
			genericExpr = "[]*ast.GenericParameter{" + strings.Join(parts, ", ") + "}"
		}
	}
	return fmt.Sprintf("&runtime.StructDefinitionValue{Node: ast.NewStructDefinition(ast.NewIdentifier(%q), %s, ast.StructKind(%q), %s, nil, %t)}", info.Name, fieldsExpr, string(info.Kind), genericExpr, info.Node.IsPrivate), true
}
