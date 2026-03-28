package compiler

import (
	"fmt"
	"reflect"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func normalizeTypeExprCacheKey(pkgName string, expr ast.TypeExpression) string {
	if expr == nil {
		return ""
	}
	value := reflect.ValueOf(expr)
	if value.IsValid() && value.Kind() == reflect.Ptr && !value.IsNil() {
		return fmt.Sprintf("%s::%T::%x", strings.TrimSpace(pkgName), expr, value.Pointer())
	}
	return strings.TrimSpace(pkgName) + "::" + typeExpressionToString(expr)
}
