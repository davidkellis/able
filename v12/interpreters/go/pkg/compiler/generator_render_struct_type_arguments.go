package compiler

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

// renderStructRuntimeTypeArguments preserves a specialized nominal's concrete
// generic identity when its native carrier crosses into runtime representation.
func (g *generator) renderStructRuntimeTypeArguments(info *structInfo) string {
	if g == nil || info == nil || !info.Specialized {
		return "nil"
	}
	generic, ok := info.TypeExpr.(*ast.GenericTypeExpression)
	if !ok || generic == nil || len(generic.Arguments) == 0 {
		return "nil"
	}
	rendered := make([]string, 0, len(generic.Arguments))
	for _, argument := range generic.Arguments {
		typeArgument, ok := g.renderTypeExpression(argument)
		if !ok {
			return "nil"
		}
		rendered = append(rendered, typeArgument)
	}
	return "[]ast.TypeExpression{" + strings.Join(rendered, ", ") + "}"
}
