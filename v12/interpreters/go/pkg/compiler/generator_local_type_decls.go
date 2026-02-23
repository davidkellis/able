package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileLocalStructDefinitionStatement(ctx *compileContext, def *ast.StructDefinition) ([]string, bool) {
	if def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
		ctx.setReason("unsupported local struct definition")
		return nil, false
	}
	defExpr, ok := g.renderLocalStructDefinitionExpr(def)
	if !ok {
		ctx.setReason("unsupported local struct definition")
		return nil, false
	}
	envName, lines := localDefinitionEnvSetup(ctx)
	defName := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := &runtime.StructDefinitionValue{Node: %s}", defName, defExpr))
	lines = append(lines, fmt.Sprintf("%s.Define(%q, %s)", envName, def.ID.Name, defName))
	lines = append(lines, fmt.Sprintf("%s.DefineStruct(%q, %s)", envName, def.ID.Name, defName))
	return lines, true
}

func (g *generator) compileLocalUnionDefinitionStatement(ctx *compileContext, def *ast.UnionDefinition) ([]string, bool) {
	if def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
		ctx.setReason("unsupported local union definition")
		return nil, false
	}
	defExpr, ok := g.renderLocalUnionDefinitionExpr(def)
	if !ok {
		ctx.setReason("unsupported local union definition")
		return nil, false
	}
	envName, lines := localDefinitionEnvSetup(ctx)
	defName := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := runtime.UnionDefinitionValue{Node: %s}", defName, defExpr))
	lines = append(lines, fmt.Sprintf("%s.Define(%q, %s)", envName, def.ID.Name, defName))
	return lines, true
}

func (g *generator) compileLocalInterfaceDefinitionStatement(ctx *compileContext, def *ast.InterfaceDefinition) ([]string, bool) {
	if def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
		ctx.setReason("unsupported local interface definition")
		return nil, false
	}
	defExpr, ok := g.renderLocalInterfaceDefinitionExpr(def)
	if !ok {
		ctx.setReason("unsupported local interface definition")
		return nil, false
	}
	envName, lines := localDefinitionEnvSetup(ctx)
	defName := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := &runtime.InterfaceDefinitionValue{Node: %s, Env: %s}", defName, defExpr, envName))
	lines = append(lines, fmt.Sprintf("%s.Define(%q, %s)", envName, def.ID.Name, defName))
	return lines, true
}

func (g *generator) compileLocalTypeAliasDefinitionStatement(ctx *compileContext, def *ast.TypeAliasDefinition) ([]string, bool) {
	if def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
		ctx.setReason("unsupported local type alias definition")
		return nil, false
	}
	if def.TargetType == nil {
		ctx.setReason("unsupported local type alias definition")
		return nil, false
	}
	if _, ok := g.renderTypeExpression(def.TargetType); !ok {
		ctx.setReason("unsupported local type alias definition")
		return nil, false
	}
	// Local type aliases only affect compile-time type references in compiled mode.
	return nil, true
}

func localDefinitionEnvSetup(ctx *compileContext) (string, []string) {
	envName := ctx.newTemp()
	lines := []string{
		"if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }",
		fmt.Sprintf("%s := __able_runtime.Env()", envName),
		fmt.Sprintf("if %s == nil { panic(fmt.Errorf(\"compiler: missing global environment\")) }", envName),
	}
	return envName, lines
}

func (g *generator) renderLocalStructDefinitionExpr(def *ast.StructDefinition) (string, bool) {
	if g == nil || def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
		return "", false
	}
	genericExpr, ok := g.renderGenericParamsExpr(def.GenericParams)
	if !ok {
		return "", false
	}
	whereExpr, ok := g.renderWhereClauseExpr(def.WhereClause)
	if !ok {
		return "", false
	}
	fieldParts := make([]string, 0, len(def.Fields))
	for _, field := range def.Fields {
		if field == nil || field.FieldType == nil {
			return "", false
		}
		fieldType, ok := g.renderTypeExpression(field.FieldType)
		if !ok {
			return "", false
		}
		nameExpr := "nil"
		if field.Name != nil && strings.TrimSpace(field.Name.Name) != "" {
			nameExpr = fmt.Sprintf("ast.NewIdentifier(%q)", field.Name.Name)
		}
		fieldParts = append(fieldParts, fmt.Sprintf("ast.NewStructFieldDefinition(%s, %s)", fieldType, nameExpr))
	}
	fieldsExpr := "nil"
	if len(fieldParts) > 0 {
		fieldsExpr = "[]*ast.StructFieldDefinition{" + strings.Join(fieldParts, ", ") + "}"
	}
	return fmt.Sprintf("ast.NewStructDefinition(ast.NewIdentifier(%q), %s, ast.StructKind(%q), %s, %s, %t)", def.ID.Name, fieldsExpr, string(def.Kind), genericExpr, whereExpr, def.IsPrivate), true
}

func (g *generator) renderLocalUnionDefinitionExpr(def *ast.UnionDefinition) (string, bool) {
	if g == nil || def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
		return "", false
	}
	genericExpr, ok := g.renderGenericParamsExpr(def.GenericParams)
	if !ok {
		return "", false
	}
	whereExpr, ok := g.renderWhereClauseExpr(def.WhereClause)
	if !ok {
		return "", false
	}
	variantParts := make([]string, 0, len(def.Variants))
	for _, variant := range def.Variants {
		if variant == nil {
			continue
		}
		rendered, ok := g.renderTypeExpression(variant)
		if !ok {
			return "", false
		}
		variantParts = append(variantParts, rendered)
	}
	variantsExpr := "nil"
	if len(variantParts) > 0 {
		variantsExpr = "[]ast.TypeExpression{" + strings.Join(variantParts, ", ") + "}"
	}
	return fmt.Sprintf("ast.NewUnionDefinition(ast.NewIdentifier(%q), %s, %s, %s, %t)", def.ID.Name, variantsExpr, genericExpr, whereExpr, def.IsPrivate), true
}

func (g *generator) renderLocalInterfaceDefinitionExpr(def *ast.InterfaceDefinition) (string, bool) {
	if g == nil || def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
		return "", false
	}
	genericExpr, ok := g.renderGenericParamsExpr(def.GenericParams)
	if !ok {
		return "", false
	}
	whereExpr, ok := g.renderWhereClauseExpr(def.WhereClause)
	if !ok {
		return "", false
	}
	signatureParts := make([]string, 0, len(def.Signatures))
	for _, sig := range def.Signatures {
		if sig == nil || sig.Name == nil || strings.TrimSpace(sig.Name.Name) == "" {
			return "", false
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
				rendered, ok := g.renderTypeExpression(param.ParamType)
				if !ok {
					return "", false
				}
				typeExpr = rendered
			}
			paramParts = append(paramParts, fmt.Sprintf("ast.NewFunctionParameter(%s, %s)", nameExpr, typeExpr))
		}
		paramsExpr := "nil"
		if len(paramParts) > 0 {
			paramsExpr = "[]*ast.FunctionParameter{" + strings.Join(paramParts, ", ") + "}"
		}
		returnExpr := "ast.WildT()"
		if sig.ReturnType != nil {
			rendered, ok := g.renderTypeExpression(sig.ReturnType)
			if !ok {
				return "", false
			}
			returnExpr = rendered
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
	selfExpr := "nil"
	if def.SelfTypePattern != nil {
		rendered, ok := g.renderTypeExpression(def.SelfTypePattern)
		if !ok {
			return "", false
		}
		selfExpr = rendered
	}
	baseExpr := "nil"
	if len(def.BaseInterfaces) > 0 {
		rendered, ok := g.renderTypeExpressionList(def.BaseInterfaces)
		if !ok {
			return "", false
		}
		baseExpr = rendered
	}
	return fmt.Sprintf("ast.NewInterfaceDefinition(ast.NewIdentifier(%q), %s, %s, %s, %s, %s, %t)", def.ID.Name, signaturesExpr, genericExpr, selfExpr, whereExpr, baseExpr, def.IsPrivate), true
}
