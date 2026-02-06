package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// CompiledThunk executes a compiled function/method body using the provided environment.
// It should return a runtime.Value or an error.
type CompiledThunk func(env *runtime.Environment, args []runtime.Value) (runtime.Value, error)

// RegisterCompiledMethod wires a compiled thunk to an existing inherent method entry.
func (i *Interpreter) RegisterCompiledMethod(typeName, methodName string, expectsSelf bool, thunk CompiledThunk) error {
	if i == nil {
		return fmt.Errorf("interpreter: nil interpreter")
	}
	if typeName == "" || methodName == "" {
		return fmt.Errorf("interpreter: missing method registration target")
	}
	if thunk == nil {
		return fmt.Errorf("interpreter: missing compiled method thunk")
	}
	bucket := i.inherentMethods[typeName]
	if bucket == nil {
		return fmt.Errorf("interpreter: missing methods for %s", typeName)
	}
	method := bucket[methodName]
	if method == nil {
		return fmt.Errorf("interpreter: missing method %s on %s", methodName, typeName)
	}
	updated := false
	applyThunk := func(fn *runtime.FunctionValue) {
		if fn == nil {
			return
		}
		def, ok := fn.Declaration.(*ast.FunctionDefinition)
		if !ok || def == nil {
			return
		}
		if functionDefinitionExpectsSelf(def) != expectsSelf {
			return
		}
		fn.Bytecode = thunk
		updated = true
	}
	switch v := method.(type) {
	case *runtime.FunctionValue:
		applyThunk(v)
	case *runtime.FunctionOverloadValue:
		if v != nil {
			for _, entry := range v.Overloads {
				applyThunk(entry)
			}
		}
	}
	if !updated {
		return fmt.Errorf("interpreter: no matching method for %s.%s", typeName, methodName)
	}
	return nil
}

// RegisterCompiledMethodOverload wires a compiled thunk to an inherent method overload that matches its signature.
func (i *Interpreter) RegisterCompiledMethodOverload(typeName, methodName string, expectsSelf bool, targetType ast.TypeExpression, paramTypes []ast.TypeExpression, thunk CompiledThunk) error {
	if i == nil {
		return fmt.Errorf("interpreter: nil interpreter")
	}
	if typeName == "" || methodName == "" {
		return fmt.Errorf("interpreter: missing method registration target")
	}
	if thunk == nil {
		return fmt.Errorf("interpreter: missing compiled method thunk")
	}
	bucket := i.inherentMethods[typeName]
	if bucket == nil {
		return fmt.Errorf("interpreter: missing methods for %s", typeName)
	}
	method := bucket[methodName]
	if method == nil {
		return fmt.Errorf("interpreter: missing method %s on %s", methodName, typeName)
	}
	matches := make([]*runtime.FunctionValue, 0, 1)
	applyThunk := func(fn *runtime.FunctionValue) {
		if fn == nil {
			return
		}
		if set := fn.MethodSet; set != nil && targetType != nil {
			if !methodTargetCompatible(targetType, set) {
				return
			}
		}
		def, ok := fn.Declaration.(*ast.FunctionDefinition)
		if !ok || def == nil {
			return
		}
		if functionDefinitionExpectsSelf(def) != expectsSelf {
			return
		}
		defParams := methodDefinitionParamTypes(def, targetType, expectsSelf)
		if len(defParams) != len(paramTypes) {
			return
		}
		for idx := range defParams {
			left := resolveSelfTypeExpr(defParams[idx], targetType)
			right := resolveSelfTypeExpr(paramTypes[idx], targetType)
			if !typeExpressionsEqual(left, right) {
				return
			}
		}
		matches = append(matches, fn)
	}
	switch v := method.(type) {
	case *runtime.FunctionValue:
		applyThunk(v)
	case *runtime.FunctionOverloadValue:
		if v != nil {
			for _, entry := range v.Overloads {
				applyThunk(entry)
			}
		}
	}
	if len(matches) == 0 {
		return fmt.Errorf("interpreter: no matching method for %s.%s", typeName, methodName)
	}
	for _, match := range matches {
		match.Bytecode = thunk
	}
	return nil
}

func methodDefinitionParamTypes(def *ast.FunctionDefinition, target ast.TypeExpression, expectsSelf bool) []ast.TypeExpression {
	if def == nil {
		return nil
	}
	params := make([]ast.TypeExpression, 0, len(def.Params)+1)
	if expectsSelf && def.IsMethodShorthand {
		params = append(params, resolveSelfTypeExpr(target, target))
	}
	for _, param := range def.Params {
		if param == nil {
			params = append(params, nil)
			continue
		}
		paramType := param.ParamType
		if paramType == nil {
			if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && strings.EqualFold(ident.Name, "self") {
				paramType = target
			}
		}
		params = append(params, resolveSelfTypeExpr(paramType, target))
	}
	return params
}

func resolveSelfTypeExpr(expr ast.TypeExpression, target ast.TypeExpression) ast.TypeExpression {
	if expr == nil {
		return expr
	}
	if simple, ok := expr.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil {
		if simple.Name.Name == "Self" {
			return target
		}
	}
	return expr
}

func methodTargetCompatible(target ast.TypeExpression, set *runtime.MethodSet) bool {
	if target == nil || set == nil || set.TargetType == nil {
		return true
	}
	left := resolveSelfTypeExpr(set.TargetType, target)
	right := resolveSelfTypeExpr(target, target)
	if typeExpressionsEqual(left, right) {
		return true
	}
	targetBase, ok := typeBaseName(target)
	if !ok {
		return false
	}
	setBase, ok := typeBaseName(set.TargetType)
	if !ok || targetBase != setBase {
		return false
	}
	if argsAreGenericParams(target, set.GenericParams) || argsAreGenericParams(set.TargetType, set.GenericParams) {
		return true
	}
	return false
}

func typeBaseName(expr ast.TypeExpression) (string, bool) {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return "", false
		}
		return t.Name.Name, true
	case *ast.GenericTypeExpression:
		return typeBaseName(t.Base)
	default:
		return "", false
	}
}

func argsAreGenericParams(expr ast.TypeExpression, params []*ast.GenericParameter) bool {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		return true
	case *ast.GenericTypeExpression:
		if len(t.Arguments) == 0 {
			return true
		}
		if len(params) == 0 {
			return false
		}
		names := make(map[string]struct{}, len(params))
		for _, gp := range params {
			if gp == nil || gp.Name == nil || gp.Name.Name == "" {
				continue
			}
			names[gp.Name.Name] = struct{}{}
		}
		if len(names) == 0 {
			return false
		}
		for _, arg := range t.Arguments {
			simple, ok := arg.(*ast.SimpleTypeExpression)
			if !ok || simple == nil || simple.Name == nil {
				return false
			}
			if _, ok := names[simple.Name.Name]; !ok {
				return false
			}
		}
		return true
	default:
		return false
	}
}
