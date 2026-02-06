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
