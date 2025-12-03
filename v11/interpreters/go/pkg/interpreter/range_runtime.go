package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

type rangeImplementation struct {
	entry         implEntry
	interfaceArgs []ast.TypeExpression
}

func (i *Interpreter) registerRangeImplementation(entry implEntry, interfaceArgs []ast.TypeExpression) {
	args := make([]ast.TypeExpression, 0, len(interfaceArgs))
	for _, arg := range interfaceArgs {
		if arg == nil {
			continue
		}
		args = append(args, arg)
	}
	i.rangeImplementations = append(i.rangeImplementations, rangeImplementation{entry: entry, interfaceArgs: args})
}

func (i *Interpreter) typeExpressionForValue(val runtime.Value) ast.TypeExpression {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case runtime.IntegerValue:
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(string(v.TypeSuffix)))
	case runtime.FloatValue:
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(string(v.TypeSuffix)))
	case runtime.StringValue:
		return ast.NewSimpleTypeExpression(ast.NewIdentifier("string"))
	case runtime.BoolValue:
		return ast.NewSimpleTypeExpression(ast.NewIdentifier("bool"))
	case runtime.CharValue:
		return ast.NewSimpleTypeExpression(ast.NewIdentifier("char"))
	case runtime.NilValue:
		return ast.NewSimpleTypeExpression(ast.NewIdentifier("nil"))
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return nil
		}
		base := ast.NewSimpleTypeExpression(ast.NewIdentifier(v.Definition.Node.ID.Name))
		if len(v.TypeArguments) == 0 {
			return base
		}
		args := make([]ast.TypeExpression, len(v.TypeArguments))
		copy(args, v.TypeArguments)
		return ast.NewGenericTypeExpression(base, args)
	case runtime.InterfaceValue:
		if v.Interface == nil || v.Interface.Node == nil || v.Interface.Node.ID == nil {
			return nil
		}
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(v.Interface.Node.ID.Name))
	default:
		return nil
	}
}

func (i *Interpreter) describeRuntimeType(val runtime.Value) string {
	if val == nil {
		return "<nil>"
	}
	switch v := val.(type) {
	case runtime.IntegerValue:
		return string(v.TypeSuffix)
	case runtime.FloatValue:
		return string(v.TypeSuffix)
	case runtime.StringValue:
		return "string"
	case runtime.BoolValue:
		return "bool"
	case runtime.CharValue:
		return "char"
	case runtime.NilValue:
		return "nil"
	case *runtime.StructInstanceValue:
		if v != nil && v.Definition != nil && v.Definition.Node != nil && v.Definition.Node.ID != nil {
			return v.Definition.Node.ID.Name
		}
	case runtime.InterfaceValue:
		if v.Interface != nil && v.Interface.Node != nil && v.Interface.Node.ID != nil {
			return v.Interface.Node.ID.Name
		}
	case *runtime.ArrayValue:
		return "Array"
	case *runtime.IteratorValue:
		return "Iterator"
	}
	return fmt.Sprintf("%T", val)
}

func (i *Interpreter) tryInvokeRangeImplementation(start, end runtime.Value, inclusive bool, env *runtime.Environment) (runtime.Value, error) {
	if len(i.rangeImplementations) == 0 {
		return nil, nil
	}
	startType := i.typeExpressionForValue(start)
	endType := i.typeExpressionForValue(end)
	if startType == nil || endType == nil {
		return nil, nil
	}
	for _, record := range i.rangeImplementations {
		if len(record.interfaceArgs) < 2 {
			continue
		}
		bindings := make(map[string]ast.TypeExpression)
		genericNames := genericNameSet(record.entry.genericParams)
		if !matchTypeExpressionTemplate(record.interfaceArgs[0], startType, genericNames, bindings) {
			continue
		}
		if !matchTypeExpressionTemplate(record.interfaceArgs[1], endType, genericNames, bindings) {
			continue
		}
		missing := false
		for _, gp := range record.entry.genericParams {
			if gp == nil || gp.Name == nil {
				continue
			}
			if _, ok := bindings[gp.Name.Name]; !ok {
				missing = true
				break
			}
		}
		if missing {
			continue
		}
		constraints := collectConstraintSpecs(record.entry.genericParams, record.entry.whereClause)
		if err := i.enforceConstraintSpecs(constraints, bindings); err != nil {
			return nil, err
		}
		methodName := "exclusive_range"
		if inclusive {
			methodName = "inclusive_range"
		}
		method, ok := record.entry.methods[methodName]
		if !ok {
			return nil, fmt.Errorf("Range implementation missing method '%s'", methodName)
		}
		result, err := i.callCallableValue(method, []runtime.Value{start, end}, env, nil)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, nil
}
