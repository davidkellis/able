package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func resolveMethodSetReceiver(def *ast.FunctionDefinition, args []runtime.Value) (runtime.Value, bool) {
	if def == nil || !functionDefinitionExpectsSelf(def) {
		return nil, false
	}
	if len(args) == 0 {
		return nil, false
	}
	return args[0], true
}

func (i *Interpreter) enforceMethodSetConstraints(fn *runtime.FunctionValue, receiver runtime.Value) error {
	if fn == nil || fn.MethodSet == nil {
		return nil
	}
	constraints := collectConstraintSpecs(fn.MethodSet.GenericParams, fn.MethodSet.WhereClause)
	if len(constraints) == 0 {
		return nil
	}
	actual := i.typeExpressionFromValue(receiver)
	if actual == nil {
		return nil
	}
	bindings := make(map[string]ast.TypeExpression)
	genericNames := genericNameSet(fn.MethodSet.GenericParams)
	actualExpanded := expandTypeAliases(actual, i.typeAliases, nil)
	if actualExpanded == nil {
		actualExpanded = actual
	}
	if fn.MethodSet.TargetType != nil && len(genericNames) > 0 {
		target := expandTypeAliases(fn.MethodSet.TargetType, i.typeAliases, nil)
		if target == nil {
			target = fn.MethodSet.TargetType
		}
		matchTypeExpressionTemplate(target, actualExpanded, genericNames, bindings)
	}
	bindings["Self"] = actualExpanded
	i.addStructTypeArgBindings(bindings, receiver)
	return i.enforceConstraintSpecs(constraints, bindings)
}

func (i *Interpreter) addStructTypeArgBindings(bindings map[string]ast.TypeExpression, receiver runtime.Value) {
	inst := structInstanceFromValue(receiver)
	if inst == nil || inst.Definition == nil || inst.Definition.Node == nil {
		return
	}
	generics := inst.Definition.Node.GenericParams
	if len(generics) == 0 {
		return
	}
	typeArgs := inst.TypeArguments
	if len(typeArgs) != len(generics) {
		inferred := i.inferStructTypeArguments(inst.Definition.Node, inst.Fields, inst.Positional)
		if len(inferred) == len(generics) {
			typeArgs = inferred
		}
	}
	if len(typeArgs) != len(generics) {
		return
	}
	mapped, err := mapTypeArguments(generics, typeArgs, "method set")
	if err != nil {
		return
	}
	for name, expr := range mapped {
		if _, ok := bindings[name]; ok {
			continue
		}
		bindings[name] = expr
	}
}

func structInstanceFromValue(value runtime.Value) *runtime.StructInstanceValue {
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		return v
	case *runtime.InterfaceValue:
		if v == nil {
			return nil
		}
		return structInstanceFromValue(v.Underlying)
	case runtime.InterfaceValue:
		return structInstanceFromValue(v.Underlying)
	default:
		return nil
	}
}
