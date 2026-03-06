package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func bytecodeUnsupported(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{errBytecodeUnsupported}, args...)...)
}

func resolveAssignmentTargetName(target ast.AssignmentTarget) (string, bool) {
	switch t := target.(type) {
	case *ast.Identifier:
		return t.Name, true
	case *ast.TypedPattern:
		return resolvePatternTargetName(t)
	}
	return "", false
}

func resolvePatternTargetName(pattern ast.Pattern) (string, bool) {
	switch p := pattern.(type) {
	case *ast.Identifier:
		return p.Name, true
	case *ast.TypedPattern:
		if p == nil {
			return "", false
		}
		return resolvePatternTargetName(p.Pattern)
	default:
		return "", false
	}
}
