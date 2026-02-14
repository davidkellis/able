package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func renderGenericParamsExpr(params []*ast.GenericParameter) string {
	if len(params) == 0 {
		return "nil"
	}
	parts := make([]string, 0, len(params))
	for _, param := range params {
		if param == nil || param.Name == nil || strings.TrimSpace(param.Name.Name) == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("ast.NewGenericParameter(ast.NewIdentifier(%q), nil)", param.Name.Name))
	}
	if len(parts) == 0 {
		return "nil"
	}
	return "[]*ast.GenericParameter{" + strings.Join(parts, ", ") + "}"
}

func (g *generator) sortedInterfaceDefsForPackage(pkgName string) []*ast.InterfaceDefinition {
	if g == nil || strings.TrimSpace(pkgName) == "" {
		return nil
	}
	names := make([]string, 0, len(g.interfaces))
	for name, def := range g.interfaces {
		if def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
			continue
		}
		if g.interfacePackages[name] != pkgName {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)
	out := make([]*ast.InterfaceDefinition, 0, len(names))
	for _, name := range names {
		if def := g.interfaces[name]; def != nil {
			out = append(out, def)
		}
	}
	return out
}

func (g *generator) sortedUnionDefsForPackage(pkgName string) []*ast.UnionDefinition {
	if g == nil || strings.TrimSpace(pkgName) == "" {
		return nil
	}
	names := make([]string, 0, len(g.unions))
	for name, def := range g.unions {
		if def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
			continue
		}
		if g.unionPackages[name] != pkgName {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)
	out := make([]*ast.UnionDefinition, 0, len(names))
	for _, name := range names {
		if def := g.unions[name]; def != nil {
			out = append(out, def)
		}
	}
	return out
}

func (g *generator) renderInterfaceDefinitionExpr(def *ast.InterfaceDefinition, envExpr string) (string, bool) {
	if g == nil || def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" || strings.TrimSpace(envExpr) == "" {
		return "", false
	}
	signatureParts := make([]string, 0, len(def.Signatures))
	for _, sig := range def.Signatures {
		if sig == nil || sig.Name == nil || strings.TrimSpace(sig.Name.Name) == "" {
			continue
		}
		paramParts := make([]string, 0, len(sig.Params))
		for _, param := range sig.Params {
			if param == nil {
				continue
			}
			nameExpr := "ast.NewIdentifier(\"_\")"
			if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && strings.TrimSpace(ident.Name) != "" {
				nameExpr = fmt.Sprintf("ast.NewIdentifier(%q)", ident.Name)
			}
			typeExpr := "ast.WildT()"
			if param.ParamType != nil {
				if rendered, ok := g.renderTypeExpression(param.ParamType); ok {
					typeExpr = rendered
				}
			}
			paramParts = append(paramParts, fmt.Sprintf("ast.NewFunctionParameter(%s, %s)", nameExpr, typeExpr))
		}
		paramsExpr := "nil"
		if len(paramParts) > 0 {
			paramsExpr = "[]*ast.FunctionParameter{" + strings.Join(paramParts, ", ") + "}"
		}
		returnExpr := "ast.WildT()"
		if sig.ReturnType != nil {
			if rendered, ok := g.renderTypeExpression(sig.ReturnType); ok {
				returnExpr = rendered
			}
		}
		signatureParts = append(signatureParts, fmt.Sprintf("ast.NewFunctionSignature(ast.NewIdentifier(%q), %s, %s, %s, nil, nil)", sig.Name.Name, paramsExpr, returnExpr, renderGenericParamsExpr(sig.GenericParams)))
	}
	signaturesExpr := "nil"
	if len(signatureParts) > 0 {
		signaturesExpr = "[]*ast.FunctionSignature{" + strings.Join(signatureParts, ", ") + "}"
	}
	selfExpr := "nil"
	if def.SelfTypePattern != nil {
		if rendered, ok := g.renderTypeExpression(def.SelfTypePattern); ok {
			selfExpr = rendered
		}
	}
	baseExpr := "nil"
	if len(def.BaseInterfaces) > 0 {
		if rendered, ok := g.renderTypeExpressionList(def.BaseInterfaces); ok {
			baseExpr = rendered
		}
	}
	return fmt.Sprintf("&runtime.InterfaceDefinitionValue{Node: ast.NewInterfaceDefinition(ast.NewIdentifier(%q), %s, %s, %s, nil, %s, %t), Env: %s}", def.ID.Name, signaturesExpr, renderGenericParamsExpr(def.GenericParams), selfExpr, baseExpr, def.IsPrivate, envExpr), true
}

func (g *generator) renderUnionDefinitionExpr(def *ast.UnionDefinition) (string, bool) {
	if g == nil || def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
		return "", false
	}
	variantParts := make([]string, 0, len(def.Variants))
	for _, variant := range def.Variants {
		if variant == nil {
			continue
		}
		rendered, ok := g.renderTypeExpression(variant)
		if !ok {
			continue
		}
		variantParts = append(variantParts, rendered)
	}
	variantsExpr := "nil"
	if len(variantParts) > 0 {
		variantsExpr = "[]ast.TypeExpression{" + strings.Join(variantParts, ", ") + "}"
	}
	return fmt.Sprintf("runtime.UnionDefinitionValue{Node: ast.NewUnionDefinition(ast.NewIdentifier(%q), %s, %s, nil, %t)}", def.ID.Name, variantsExpr, renderGenericParamsExpr(def.GenericParams), def.IsPrivate), true
}
