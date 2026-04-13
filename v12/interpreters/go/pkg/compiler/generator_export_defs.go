package compiler

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) renderInterfaceConstraintsExpr(constraints []*ast.InterfaceConstraint) (string, bool) {
	if len(constraints) == 0 {
		return "nil", true
	}
	parts := make([]string, 0, len(constraints))
	for _, constraint := range constraints {
		if constraint == nil || constraint.InterfaceType == nil {
			continue
		}
		renderedType, ok := g.renderTypeExpression(constraint.InterfaceType)
		if !ok {
			return "", false
		}
		parts = append(parts, fmt.Sprintf("ast.NewInterfaceConstraint(%s)", renderedType))
	}
	if len(parts) == 0 {
		return "nil", true
	}
	return "[]*ast.InterfaceConstraint{" + strings.Join(parts, ", ") + "}", true
}

func (g *generator) renderGenericParamsExpr(params []*ast.GenericParameter) (string, bool) {
	if len(params) == 0 {
		return "nil", true
	}
	parts := make([]string, 0, len(params))
	for _, param := range params {
		if param == nil || param.Name == nil || strings.TrimSpace(param.Name.Name) == "" {
			continue
		}
		constraintsExpr, ok := g.renderInterfaceConstraintsExpr(param.Constraints)
		if !ok {
			return "", false
		}
		parts = append(parts, fmt.Sprintf("ast.NewGenericParameter(ast.NewIdentifier(%q), %s)", param.Name.Name, constraintsExpr))
	}
	if len(parts) == 0 {
		return "nil", true
	}
	return "[]*ast.GenericParameter{" + strings.Join(parts, ", ") + "}", true
}

func (g *generator) renderWhereClauseExpr(whereClause []*ast.WhereClauseConstraint) (string, bool) {
	if len(whereClause) == 0 {
		return "nil", true
	}
	parts := make([]string, 0, len(whereClause))
	for _, clause := range whereClause {
		if clause == nil || clause.TypeParam == nil {
			continue
		}
		typeParamExpr, ok := g.renderTypeExpression(clause.TypeParam)
		if !ok {
			return "", false
		}
		constraintsExpr, ok := g.renderInterfaceConstraintsExpr(clause.Constraints)
		if !ok {
			return "", false
		}
		parts = append(parts, fmt.Sprintf("ast.NewWhereClauseConstraint(%s, %s)", typeParamExpr, constraintsExpr))
	}
	if len(parts) == 0 {
		return "nil", true
	}
	return "[]*ast.WhereClauseConstraint{" + strings.Join(parts, ", ") + "}", true
}

func (g *generator) renderBlockExpressionExpr(block *ast.BlockExpression) (string, bool) {
	if block == nil {
		return "nil", true
	}
	encoded, err := json.Marshal(block)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("func() *ast.BlockExpression { node, err := interpreter.DecodeNodeJSON([]byte(%q)); if err != nil { return nil }; typed, ok := node.(*ast.BlockExpression); if !ok { return nil }; return typed }()", string(encoded)), true
}

func (g *generator) renderMethodsDefinitionExpr(def *ast.MethodsDefinition) (string, bool) {
	if def == nil {
		return "", false
	}
	encoded, err := json.Marshal(def)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("func() *ast.MethodsDefinition { node, err := interpreter.DecodeNodeJSON([]byte(%q)); if err != nil { return nil }; typed, ok := node.(*ast.MethodsDefinition); if !ok { return nil }; return typed }()", string(encoded)), true
}

func (g *generator) renderImplementationDefinitionExpr(def *ast.ImplementationDefinition) (string, bool) {
	if def == nil {
		return "", false
	}
	encoded, err := json.Marshal(def)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("func() *ast.ImplementationDefinition { node, err := interpreter.DecodeNodeJSON([]byte(%q)); if err != nil { return nil }; typed, ok := node.(*ast.ImplementationDefinition); if !ok { return nil }; return typed }()", string(encoded)), true
}

func (g *generator) sortedInterfaceDefsForPackage(pkgName string) []*ast.InterfaceDefinition {
	if g == nil || strings.TrimSpace(pkgName) == "" {
		return nil
	}
	defs := g.interfacesByPackage[pkgName]
	names := make([]string, 0, len(defs))
	for name, def := range defs {
		if def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
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
		if def := defs[name]; def != nil {
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
		signatureGenericExpr, ok := g.renderGenericParamsExpr(sig.GenericParams)
		if !ok {
			return "", false
		}
		signatureWhereExpr, ok := g.renderWhereClauseExpr(sig.WhereClause)
		if !ok {
			return "", false
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
		defaultExpr, ok := g.renderBlockExpressionExpr(sig.DefaultImpl)
		if !ok {
			return "", false
		}
		signatureParts = append(signatureParts, fmt.Sprintf("ast.NewFunctionSignature(ast.NewIdentifier(%q), %s, %s, %s, %s, %s)", sig.Name.Name, paramsExpr, returnExpr, signatureGenericExpr, signatureWhereExpr, defaultExpr))
	}
	signaturesExpr := "nil"
	if len(signatureParts) > 0 {
		signaturesExpr = "[]*ast.FunctionSignature{" + strings.Join(signatureParts, ", ") + "}"
	}
	genericExpr, ok := g.renderGenericParamsExpr(def.GenericParams)
	if !ok {
		return "", false
	}
	whereExpr, ok := g.renderWhereClauseExpr(def.WhereClause)
	if !ok {
		return "", false
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
	return fmt.Sprintf("&runtime.InterfaceDefinitionValue{Node: ast.NewInterfaceDefinition(ast.NewIdentifier(%q), %s, %s, %s, %s, %s, %t), Env: %s}", def.ID.Name, signaturesExpr, genericExpr, selfExpr, whereExpr, baseExpr, def.IsPrivate, envExpr), true
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
	genericExpr, ok := g.renderGenericParamsExpr(def.GenericParams)
	if !ok {
		return "", false
	}
	whereExpr, ok := g.renderWhereClauseExpr(def.WhereClause)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("runtime.UnionDefinitionValue{Node: ast.NewUnionDefinition(ast.NewIdentifier(%q), %s, %s, %s, %t)}", def.ID.Name, variantsExpr, genericExpr, whereExpr, def.IsPrivate), true
}
