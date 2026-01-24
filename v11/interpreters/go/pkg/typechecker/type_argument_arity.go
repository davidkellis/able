package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

var builtinTypeArgumentArity = map[string]int{
	"Array":    1,
	"Iterator": 1,
	"Range":    1,
	"Future":   1,
	"Map":      2,
	"HashMap":  2,
	"Channel":  1,
	"Mutex":    0,
}

func expectedTypeArgumentCount(name string, base Type) (int, bool) {
	if expected, ok := expectedTypeArgumentCountFromBase(base); ok {
		return expected, true
	}
	if builtin, ok := builtinTypeArgumentArity[name]; ok {
		return builtin, true
	}
	return 0, false
}

func expectedTypeArgumentCountFromBase(base Type) (int, bool) {
	switch t := base.(type) {
	case StructType:
		return len(t.TypeParams), true
	case InterfaceType:
		return len(t.TypeParams), true
	case UnionType:
		return len(t.TypeParams), true
	case AliasType:
		return len(t.TypeParams), true
	case PrimitiveType, IntegerType, FloatType:
		return 0, true
	default:
		return 0, false
	}
}

func shouldSkipTypeArgumentCheck(name string, localTypeNames map[string]struct{}, declNodes map[string]ast.Node) bool {
	if name == "" || localTypeNames == nil || declNodes == nil {
		return false
	}
	if _, ok := localTypeNames[name]; !ok {
		return false
	}
	if _, declared := declNodes[name]; declared {
		return false
	}
	return true
}

func typeArgumentArityDiagnostic(name string, expected, actual int, node ast.Node) Diagnostic {
	return Diagnostic{
		Message: fmt.Sprintf("typechecker: type '%s' expects %d type argument(s), got %d", name, expected, actual),
		Node:    node,
	}
}
