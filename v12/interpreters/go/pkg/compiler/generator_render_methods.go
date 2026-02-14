package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) renderMethodParamTypes(method *methodInfo) (string, bool) {
	if method == nil || method.Info == nil || method.Info.Definition == nil {
		return "nil", false
	}
	def := method.Info.Definition
	target := method.TargetType
	parts := make([]string, 0, len(def.Params)+1)
	if method.ExpectsSelf && def.IsMethodShorthand {
		targetExpr, ok := g.renderTypeExpression(resolveSelfTypeExpr(target, target))
		if !ok {
			return "", false
		}
		parts = append(parts, targetExpr)
	}
	for _, param := range def.Params {
		if param == nil {
			parts = append(parts, "nil")
			continue
		}
		paramType := param.ParamType
		if paramType == nil {
			if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && strings.EqualFold(ident.Name, "self") {
				paramType = target
			}
		}
		paramType = resolveSelfTypeExpr(paramType, target)
		if paramType == nil {
			parts = append(parts, "nil")
			continue
		}
		rendered, ok := g.renderTypeExpression(paramType)
		if !ok {
			return "", false
		}
		parts = append(parts, rendered)
	}
	if len(parts) == 0 {
		return "nil", true
	}
	return fmt.Sprintf("[]ast.TypeExpression{%s}", strings.Join(parts, ", ")), true
}

func (g *generator) renderImplMethodParamTypes(method *implMethodInfo) (string, bool) {
	if method == nil || method.Info == nil || method.Info.Definition == nil {
		return "nil", false
	}
	def := method.Info.Definition
	expectsSelf := methodDefinitionExpectsSelf(def)
	params := methodDefinitionParamTypes(def, method.TargetType, expectsSelf)
	interfaceBindings := make(map[string]ast.TypeExpression)
	for idx, gp := range method.InterfaceGenerics {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			continue
		}
		if idx >= len(method.InterfaceArgs) || method.InterfaceArgs[idx] == nil {
			continue
		}
		interfaceBindings[gp.Name.Name] = method.InterfaceArgs[idx]
	}
	parts := make([]string, 0, len(params))
	for _, paramType := range params {
		if paramType == nil {
			parts = append(parts, "nil")
			continue
		}
		paramType = substituteTypeParams(paramType, interfaceBindings)
		rendered, ok := g.renderTypeExpression(paramType)
		if !ok {
			return "", false
		}
		parts = append(parts, rendered)
	}
	if len(parts) == 0 {
		return "nil", true
	}
	return fmt.Sprintf("[]ast.TypeExpression{%s}", strings.Join(parts, ", ")), true
}

func (g *generator) renderFunctionParamTypes(info *functionInfo) (string, bool) {
	if info == nil || info.Definition == nil {
		return "nil", false
	}
	def := info.Definition
	parts := make([]string, 0, len(def.Params))
	for _, param := range def.Params {
		if param == nil || param.ParamType == nil {
			parts = append(parts, "nil")
			continue
		}
		rendered, ok := g.renderTypeExpression(param.ParamType)
		if !ok {
			return "", false
		}
		parts = append(parts, rendered)
	}
	if len(parts) == 0 {
		return "nil", true
	}
	return fmt.Sprintf("[]ast.TypeExpression{%s}", strings.Join(parts, ", ")), true
}

func (g *generator) renderTypeExpressionList(exprs []ast.TypeExpression) (string, bool) {
	if len(exprs) == 0 {
		return "nil", true
	}
	parts := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		if expr == nil {
			parts = append(parts, "nil")
			continue
		}
		rendered, ok := g.renderTypeExpression(expr)
		if !ok {
			return "", false
		}
		parts = append(parts, rendered)
	}
	if len(parts) == 0 {
		return "nil", true
	}
	return fmt.Sprintf("[]ast.TypeExpression{%s}", strings.Join(parts, ", ")), true
}
