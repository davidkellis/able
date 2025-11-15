package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) evaluateMapLiteral(lit *ast.MapLiteral, env *runtime.Environment) (runtime.Value, error) {
	hm := &runtime.HashMapValue{Entries: make([]runtime.HashMapEntry, 0)}
	if lit == nil || len(lit.Elements) == 0 {
		return hm, nil
	}
	for _, element := range lit.Elements {
		switch entry := element.(type) {
		case *ast.MapLiteralEntry:
			keyVal, err := i.evaluateExpression(entry.Key, env)
			if err != nil {
				return nil, err
			}
			valueVal, err := i.evaluateExpression(entry.Value, env)
			if err != nil {
				return nil, err
			}
			if err := i.hashMapInsertEntry(hm, keyVal, valueVal); err != nil {
				return nil, err
			}
		case *ast.MapLiteralSpread:
			spreadVal, err := i.evaluateExpression(entry.Expression, env)
			if err != nil {
				return nil, err
			}
			source, ok := spreadVal.(*runtime.HashMapValue)
			if !ok {
				return nil, fmt.Errorf("map literal spread expects hash map value")
			}
			for _, existing := range source.Entries {
				if err := i.hashMapInsertEntry(hm, existing.Key, existing.Value); err != nil {
					return nil, err
				}
			}
		default:
			return nil, fmt.Errorf("unsupported map literal element %T", element)
		}
	}
	return hm, nil
}
