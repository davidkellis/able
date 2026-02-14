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

// RegisterCompiledImplMethodOverload wires a compiled thunk to an impl method overload that matches its signature.
func (i *Interpreter) RegisterCompiledImplMethodOverload(interfaceName string, targetType ast.TypeExpression, interfaceArgs []ast.TypeExpression, constraintSig string, implName string, methodName string, paramTypes []ast.TypeExpression, thunk CompiledThunk) error {
	if i == nil {
		return fmt.Errorf("interpreter: nil interpreter")
	}
	if interfaceName == "" || methodName == "" {
		return fmt.Errorf("interpreter: missing impl method registration target")
	}
	if thunk == nil {
		return fmt.Errorf("interpreter: missing compiled impl method thunk")
	}
	if targetType == nil {
		return fmt.Errorf("interpreter: missing impl method target type")
	}
	if constraintSig == "" {
		constraintSig = "<none>"
	}
	normalizedTarget := expandTypeAliases(targetType, i.typeAliases, nil)
	normalizedArgs := interfaceArgs
	entries := make([]*implEntry, 0, len(i.implMethods)+len(i.genericImpls))
	for _, bucket := range i.implMethods {
		for idx := range bucket {
			entries = append(entries, &bucket[idx])
		}
	}
	for idx := range i.genericImpls {
		entries = append(entries, &i.genericImpls[idx])
	}
	matches := 0
	ifaceBindings := make(map[string]ast.TypeExpression)
	if ifaceDef := i.interfaces[interfaceName]; ifaceDef != nil && ifaceDef.Node != nil {
		for idx, gp := range ifaceDef.Node.GenericParams {
			if gp == nil || gp.Name == nil || gp.Name.Name == "" {
				continue
			}
			if idx >= len(normalizedArgs) || normalizedArgs[idx] == nil {
				continue
			}
			ifaceBindings[gp.Name.Name] = normalizedArgs[idx]
		}
	}
	for _, entry := range entries {
		if entry == nil || entry.definition == nil {
			continue
		}
		if entry.interfaceName != interfaceName {
			continue
		}
		if implName != "" {
			if entry.definition.ImplName == nil || entry.definition.ImplName.Name != implName {
				continue
			}
		} else if entry.definition.ImplName != nil {
			continue
		}
		matchedTarget, ok := i.matchCompiledImplTarget(entry, normalizedTarget)
		if !ok {
			continue
		}
		if !interfaceArgsEqual(i, entry.definition.InterfaceArgs, normalizedArgs) {
			continue
		}
		constraints := collectConstraintSpecs(entry.genericParams, entry.whereClause)
		// Compiler-side registration can still reference source aliases while
		// runtime impl entries are canonicalized. Accept either signature form.
		entryConstraintSig := constraintSignature(constraints, typeExpressionToString)
		entryConstraintExpandedSig := constraintSignature(constraints, func(expr ast.TypeExpression) string {
			return typeExpressionToString(expandTypeAliases(expr, i.typeAliases, nil))
		})
		if entryConstraintSig != constraintSig && entryConstraintExpandedSig != constraintSig {
			continue
		}
		method := entry.methods[methodName]
		if method == nil {
			continue
		}
		applyThunk := func(fn *runtime.FunctionValue) {
			if fn == nil || fn.Declaration == nil {
				return
			}
			def, ok := fn.Declaration.(*ast.FunctionDefinition)
			if !ok || def == nil {
				return
			}
			expectsSelf := functionDefinitionExpectsSelf(def)
			defParams := methodDefinitionParamTypes(def, matchedTarget, expectsSelf)
			if len(defParams) != len(paramTypes) {
				return
			}
			for idx := range defParams {
				left := expandTypeAliases(defParams[idx], i.typeAliases, nil)
				left = substituteCompiledThunkTypeParams(left, ifaceBindings)
				right := expandTypeAliases(paramTypes[idx], i.typeAliases, nil)
				right = substituteCompiledThunkTypeParams(right, ifaceBindings)
				if !typeExpressionsEqual(left, right) {
					return
				}
			}
			fn.Bytecode = thunk
			matches++
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
	}
	if matches == 0 {
		return fmt.Errorf(
			"interpreter: no matching impl method for %s.%s target=%s ifaceArgs=%s constraints=%s params=%s",
			interfaceName,
			methodName,
			typeExpressionToString(targetType),
			typeExpressionListString(interfaceArgs),
			constraintSig,
			typeExpressionListString(paramTypes),
		)
	}
	return nil
}

func (i *Interpreter) matchCompiledImplTarget(entry *implEntry, target ast.TypeExpression) (ast.TypeExpression, bool) {
	if i == nil || entry == nil || target == nil {
		return nil, false
	}
	candidates := make([]ast.TypeExpression, 0, 2)
	if entry.registrationTarget != nil {
		candidates = append(candidates, entry.registrationTarget)
	}
	if entry.definition != nil && entry.definition.TargetType != nil {
		candidates = append(candidates, entry.definition.TargetType)
	}
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		normalizedCandidate := expandTypeAliases(candidate, i.typeAliases, nil)
		if typeExpressionsEqual(normalizedCandidate, target) {
			return candidate, true
		}
	}
	return nil, false
}

// RegisterCompiledImplNamespaceMethod wires a compiled thunk to a named impl namespace method.
func (i *Interpreter) RegisterCompiledImplNamespaceMethod(env *runtime.Environment, implName string, methodName string, paramTypes []ast.TypeExpression, thunk CompiledThunk) error {
	if i == nil {
		return fmt.Errorf("interpreter: nil interpreter")
	}
	if env == nil {
		return fmt.Errorf("interpreter: missing environment")
	}
	if implName == "" || methodName == "" {
		return fmt.Errorf("interpreter: missing impl namespace registration target")
	}
	if thunk == nil {
		return fmt.Errorf("interpreter: missing compiled impl namespace thunk")
	}
	val, err := env.Get(implName)
	if err != nil {
		return fmt.Errorf("interpreter: missing impl namespace %s", implName)
	}
	var ns runtime.ImplementationNamespaceValue
	switch v := val.(type) {
	case runtime.ImplementationNamespaceValue:
		ns = v
	case *runtime.ImplementationNamespaceValue:
		if v == nil {
			return fmt.Errorf("interpreter: missing impl namespace %s", implName)
		}
		ns = *v
	default:
		return fmt.Errorf("interpreter: %s is not an impl namespace", implName)
	}
	if ns.Methods == nil {
		return fmt.Errorf("interpreter: impl namespace %s missing methods", implName)
	}
	method := ns.Methods[methodName]
	if method == nil {
		return fmt.Errorf("interpreter: impl namespace %s missing method %s", implName, methodName)
	}
	matches := 0
	applyThunk := func(fn *runtime.FunctionValue) {
		if fn == nil || fn.Declaration == nil {
			return
		}
		def, ok := fn.Declaration.(*ast.FunctionDefinition)
		if !ok || def == nil {
			return
		}
		expectsSelf := functionDefinitionExpectsSelf(def)
		defParams := methodDefinitionParamTypes(def, ns.TargetType, expectsSelf)
		if len(defParams) != len(paramTypes) {
			return
		}
		for idx := range defParams {
			left := expandTypeAliases(defParams[idx], i.typeAliases, nil)
			right := expandTypeAliases(paramTypes[idx], i.typeAliases, nil)
			if !typeExpressionsEqual(left, right) {
				return
			}
		}
		fn.Bytecode = thunk
		matches++
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
	if matches == 0 {
		return fmt.Errorf("interpreter: no matching impl namespace method for %s.%s", implName, methodName)
	}
	return nil
}

// RegisterCompiledFunctionOverload wires a compiled thunk to a function overload that matches its signature.
func (i *Interpreter) RegisterCompiledFunctionOverload(env *runtime.Environment, name string, paramTypes []ast.TypeExpression, thunk CompiledThunk) error {
	if i == nil {
		return fmt.Errorf("interpreter: nil interpreter")
	}
	if env == nil {
		return fmt.Errorf("interpreter: missing environment")
	}
	if name == "" {
		return fmt.Errorf("interpreter: missing function registration target")
	}
	if thunk == nil {
		return fmt.Errorf("interpreter: missing compiled function thunk")
	}
	value, err := env.Get(name)
	if err != nil {
		return fmt.Errorf("interpreter: missing function %s", name)
	}
	matches := make([]*runtime.FunctionValue, 0, 1)
	for _, fn := range runtime.FlattenFunctionOverloads(value) {
		if fn == nil {
			continue
		}
		def, ok := fn.Declaration.(*ast.FunctionDefinition)
		if !ok || def == nil {
			continue
		}
		if len(def.Params) != len(paramTypes) {
			continue
		}
		match := true
		for idx, param := range def.Params {
			var defType ast.TypeExpression
			if param != nil {
				defType = param.ParamType
			}
			if !typeExpressionsEqual(defType, paramTypes[idx]) {
				match = false
				break
			}
		}
		if match {
			matches = append(matches, fn)
		}
	}
	if len(matches) == 0 {
		return fmt.Errorf("interpreter: no matching function for %s", name)
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

func typeExpressionListString(exprs []ast.TypeExpression) string {
	if len(exprs) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		parts = append(parts, typeExpressionToString(expr))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func substituteCompiledThunkTypeParams(expr ast.TypeExpression, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if expr == nil || len(bindings) == 0 {
		return expr
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return expr
		}
		if bound, ok := bindings[t.Name.Name]; ok && bound != nil {
			return bound
		}
		return expr
	case *ast.GenericTypeExpression:
		if t == nil {
			return expr
		}
		base := substituteCompiledThunkTypeParams(t.Base, bindings)
		args := make([]ast.TypeExpression, len(t.Arguments))
		for idx, arg := range t.Arguments {
			args[idx] = substituteCompiledThunkTypeParams(arg, bindings)
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.FunctionTypeExpression:
		if t == nil {
			return expr
		}
		params := make([]ast.TypeExpression, len(t.ParamTypes))
		for idx, param := range t.ParamTypes {
			params[idx] = substituteCompiledThunkTypeParams(param, bindings)
		}
		return ast.NewFunctionTypeExpression(params, substituteCompiledThunkTypeParams(t.ReturnType, bindings))
	case *ast.NullableTypeExpression:
		if t == nil {
			return expr
		}
		return ast.NewNullableTypeExpression(substituteCompiledThunkTypeParams(t.InnerType, bindings))
	case *ast.ResultTypeExpression:
		if t == nil {
			return expr
		}
		return ast.NewResultTypeExpression(substituteCompiledThunkTypeParams(t.InnerType, bindings))
	case *ast.UnionTypeExpression:
		if t == nil {
			return expr
		}
		members := make([]ast.TypeExpression, len(t.Members))
		for idx, member := range t.Members {
			members[idx] = substituteCompiledThunkTypeParams(member, bindings)
		}
		return ast.NewUnionTypeExpression(members)
	default:
		return expr
	}
}
