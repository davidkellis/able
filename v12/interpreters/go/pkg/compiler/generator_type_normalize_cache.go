package compiler

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

func normalizeTypeExprCacheKey(pkgName string, expr ast.TypeExpression) string {
	if expr == nil {
		return ""
	}
	return strings.TrimSpace(pkgName) + "::" + typeExpressionToString(expr)
}
