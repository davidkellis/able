package interpreter

import (
	"fmt"
	"math/big"
	"os"
	"time"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

type externNativeFactory func(i *Interpreter, def *ast.ExternFunctionBody, arity int) *runtime.NativeFunctionValue

var externNativeHandlers = map[ast.HostTarget]map[string]externNativeFactory{
	ast.HostTargetGo: {
		"now_nanos": func(_ *Interpreter, _ *ast.ExternFunctionBody, arity int) *runtime.NativeFunctionValue {
			return &runtime.NativeFunctionValue{
				Name:  "now_nanos",
				Arity: arity,
				Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
					return runtime.IntegerValue{Val: big.NewInt(time.Now().UnixNano()), TypeSuffix: runtime.IntegerI64}, nil
				},
			}
		},
		"read_text": func(_ *Interpreter, _ *ast.ExternFunctionBody, arity int) *runtime.NativeFunctionValue {
			return &runtime.NativeFunctionValue{
				Name:  "read_text",
				Arity: arity,
				Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
					if len(args) < 1 {
						return nil, fmt.Errorf("read_text expects a path argument")
					}
					path, ok := args[0].(runtime.StringValue)
					if !ok {
						return nil, fmt.Errorf("read_text expects a string path")
					}
					data, err := os.ReadFile(path.Val)
					if err != nil {
						return nil, raiseSignal{value: runtime.ErrorValue{Message: err.Error()}}
					}
					return runtime.StringValue{Val: string(data)}, nil
				},
			}
		},
	},
}

func registerExternNativeHandler(target ast.HostTarget, name string, handler externNativeFactory) {
	if _, ok := externNativeHandlers[target]; !ok {
		externNativeHandlers[target] = make(map[string]externNativeFactory)
	}
	externNativeHandlers[target][name] = handler
}

func (i *Interpreter) evaluateFunctionDefinition(def *ast.FunctionDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Function definition requires identifier")
	}
	if err := i.validateGenericConstraints(def); err != nil {
		return nil, err
	}
	fnVal := &runtime.FunctionValue{Declaration: def, Closure: env}
	env.Define(def.ID.Name, fnVal)
	i.registerSymbol(def.ID.Name, fnVal)
	if qn := i.qualifiedName(def.ID.Name); qn != "" {
		i.global.Define(qn, fnVal)
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateExternFunctionBody(def *ast.ExternFunctionBody, env *runtime.Environment) (runtime.Value, error) {
	if def == nil || def.Signature == nil || def.Signature.ID == nil {
		return runtime.NilValue{}, nil
	}
	name := def.Signature.ID.Name
	if name == "" {
		return runtime.NilValue{}, nil
	}
	if _, err := env.Get(name); err == nil {
		return runtime.NilValue{}, nil
	}
	native := i.makeExternNative(def)
	if native == nil {
		return runtime.NilValue{}, nil
	}
	env.Define(name, native)
	i.registerSymbol(name, native)
	if qn := i.qualifiedName(name); qn != "" {
		i.global.Define(qn, native)
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) makeExternNative(def *ast.ExternFunctionBody) *runtime.NativeFunctionValue {
	if def == nil || def.Signature == nil || def.Signature.ID == nil {
		return nil
	}
	name := def.Signature.ID.Name
	arity := len(def.Signature.Params)
	if def.Target != ast.HostTargetGo {
		return nil
	}
	if factory, ok := externNativeHandlers[def.Target][name]; ok {
		return factory(i, def, arity)
	}
	return i.makeMissingExternNative(name, def.Target, arity)
}

func (i *Interpreter) makeMissingExternNative(name string, target ast.HostTarget, arity int) *runtime.NativeFunctionValue {
	return &runtime.NativeFunctionValue{
		Name:  name,
		Arity: arity,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return nil, fmt.Errorf("extern function %s for target %s is not implemented", name, target)
		},
	}
}

func (i *Interpreter) evaluateStructDefinition(def *ast.StructDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Struct definition requires identifier")
	}
	structVal := &runtime.StructDefinitionValue{Node: def}
	env.Define(def.ID.Name, structVal)
	i.registerSymbol(def.ID.Name, structVal)
	if qn := i.qualifiedName(def.ID.Name); qn != "" {
		i.global.Define(qn, structVal)
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateUnionDefinition(def *ast.UnionDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Union definition requires identifier")
	}
	unionVal := runtime.UnionDefinitionValue{Node: def}
	env.Define(def.ID.Name, unionVal)
	i.unionDefinitions[def.ID.Name] = &unionVal
	i.registerSymbol(def.ID.Name, unionVal)
	if qn := i.qualifiedName(def.ID.Name); qn != "" {
		i.global.Define(qn, unionVal)
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateInterfaceDefinition(def *ast.InterfaceDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Interface definition requires identifier")
	}
	ifaceVal := &runtime.InterfaceDefinitionValue{Node: def, Env: env}
	env.Define(def.ID.Name, ifaceVal)
	i.interfaces[def.ID.Name] = ifaceVal
	i.registerSymbol(def.ID.Name, ifaceVal)
	if qn := i.qualifiedName(def.ID.Name); qn != "" {
		i.global.Define(qn, ifaceVal)
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateImplementationDefinition(def *ast.ImplementationDefinition, env *runtime.Environment, isBuiltin bool) (runtime.Value, error) {
	if def.InterfaceName == nil {
		return nil, fmt.Errorf("Implementation requires interface name")
	}
	ifaceName := def.InterfaceName.Name
	ifaceDef, ok := i.interfaces[ifaceName]
	if !ok {
		return nil, fmt.Errorf("Interface '%s' is not defined", ifaceName)
	}
	variants, unionSignatures, err := expandImplementationTargetVariants(def.TargetType)
	if err != nil {
		return nil, err
	}
	if len(variants) == 0 {
		return nil, fmt.Errorf("Implementation target must reference at least one concrete type")
	}
	mergedGenerics := i.mergeImplementationGenerics(def, env)
	methods := make(map[string]runtime.Value)
	hasExplicit := false
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			return nil, fmt.Errorf("Implementation method requires identifier")
		}
		mergeFunctionLike(methods, fn.ID.Name, &runtime.FunctionValue{Declaration: fn, Closure: env, MethodPriority: -1})
		hasExplicit = true
	}
	if ifaceDef.Node != nil {
		for _, sig := range ifaceDef.Node.Signatures {
			if sig == nil || sig.Name == nil {
				continue
			}
			name := sig.Name.Name
			if sig.DefaultImpl == nil {
				continue
			}
			if _, exists := methods[name]; exists {
				continue
			}
			defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
			mergeFunctionLike(methods, name, &runtime.FunctionValue{Declaration: defaultDef, Closure: ifaceDef.Env, MethodPriority: -1})
		}
	}
	constraintSpecs := collectConstraintSpecs(def.GenericParams, def.WhereClause)
	baseConstraintSig := constraintSignature(constraintSpecs)
	targetDescription := typeExpressionToString(def.TargetType)
	genericNames := genericNameSet(mergedGenerics)
	for _, variant := range variants {
		if def.ImplName == nil {
			isGenericTarget := false
			if len(genericNames) > 0 {
				_, isGenericTarget = genericNames[variant.typeName]
			}
			if !isGenericTarget {
				if err := i.registerUnnamedImpl(ifaceName, def.InterfaceArgs, variant, unionSignatures, baseConstraintSig, targetDescription, isBuiltin); err != nil {
					return nil, err
				}
			}
			entry := implEntry{
				interfaceName: ifaceName,
				methods:       methods,
				definition:    def,
				argTemplates:  variant.argTemplates,
				genericParams: mergedGenerics,
				whereClause:   def.WhereClause,
				defaultOnly:   !hasExplicit,
			}
			if len(unionSignatures) > 0 {
				entry.unionVariants = append([]string(nil), unionSignatures...)
			}
			if isGenericTarget {
				i.genericImpls = append(i.genericImpls, entry)
			} else {
				i.implMethods[variant.typeName] = append(i.implMethods[variant.typeName], entry)
				if ifaceName == "Range" {
					i.registerRangeImplementation(entry, def.InterfaceArgs)
				}
			}
		}
	}
	if def.ImplName != nil {
		name := def.ImplName.Name
		implVal := runtime.ImplementationNamespaceValue{
			Name:          def.ImplName,
			InterfaceName: def.InterfaceName,
			TargetType:    def.TargetType,
			Methods:       methods,
		}
		env.Define(name, implVal)
		i.registerSymbol(name, implVal)
		if qn := i.qualifiedName(name); qn != "" {
			i.global.Define(qn, implVal)
		}
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateMethodsDefinition(def *ast.MethodsDefinition, env *runtime.Environment) (runtime.Value, error) {
	var typeName string
	switch t := def.TargetType.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return nil, fmt.Errorf("MethodsDefinition requires simple target type")
		}
		typeName = t.Name.Name
	case *ast.GenericTypeExpression:
		base, ok := t.Base.(*ast.SimpleTypeExpression)
		if !ok || base.Name == nil {
			return nil, fmt.Errorf("MethodsDefinition requires simple target type")
		}
		typeName = base.Name.Name
	default:
		return nil, fmt.Errorf("MethodsDefinition requires simple target type")
	}
	bucket, ok := i.inherentMethods[typeName]
	if !ok {
		bucket = make(map[string]runtime.Value)
		i.inherentMethods[typeName] = bucket
	}
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			return nil, fmt.Errorf("Method definition requires identifier")
		}
		mergeFunctionLike(bucket, fn.ID.Name, &runtime.FunctionValue{Declaration: fn, Closure: env})
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) mergeImplementationGenerics(def *ast.ImplementationDefinition, env *runtime.Environment) []*ast.GenericParameter {
	seen := make(map[string]struct{})
	result := make([]*ast.GenericParameter, 0, len(def.GenericParams))
	for _, gp := range def.GenericParams {
		if gp == nil || gp.Name == nil {
			continue
		}
		result = append(result, gp)
		seen[gp.Name.Name] = struct{}{}
	}
	for _, inferred := range i.inferGenericsFromTarget(def.TargetType, env) {
		if inferred == nil || inferred.Name == nil {
			continue
		}
		if _, ok := seen[inferred.Name.Name]; ok {
			continue
		}
		result = append(result, inferred)
		seen[inferred.Name.Name] = struct{}{}
	}
	return result
}

func (i *Interpreter) inferGenericsFromTarget(target ast.TypeExpression, env *runtime.Environment) []*ast.GenericParameter {
	switch t := target.(type) {
	case *ast.GenericTypeExpression:
		baseName, ok := simpleTypeName(t.Base)
		if !ok || env == nil {
			return nil
		}
		defVal, err := env.Get(baseName)
		if err != nil {
			return nil
		}
		structDef, ok := defVal.(*runtime.StructDefinitionValue)
		if !ok || structDef.Node == nil {
			return nil
		}
		if len(structDef.Node.GenericParams) != len(t.Arguments) {
			return nil
		}
		var generics []*ast.GenericParameter
		for idx, arg := range t.Arguments {
			argSimple, ok := arg.(*ast.SimpleTypeExpression)
			if !ok || argSimple.Name == nil {
				continue
			}
			param := structDef.Node.GenericParams[idx]
			if param == nil || param.Name == nil {
				continue
			}
			if argSimple.Name.Name == param.Name.Name {
				generics = append(generics, param)
			}
		}
		return generics
	case *ast.UnionTypeExpression:
		var generics []*ast.GenericParameter
		for _, member := range t.Members {
			generics = append(generics, i.inferGenericsFromTarget(member, env)...)
		}
		return generics
	default:
		return nil
	}
}

func (i *Interpreter) validateGenericConstraints(def *ast.FunctionDefinition) error {
	if def == nil || len(def.GenericParams) == 0 {
		return nil
	}
	for _, param := range def.GenericParams {
		if param == nil || param.Name == nil {
			continue
		}
		for _, constraint := range param.Constraints {
			if constraint == nil || constraint.InterfaceType == nil {
				continue
			}
			ifaceName, ok := simpleTypeName(constraint.InterfaceType)
			if !ok || ifaceName == "" {
				return fmt.Errorf("Unknown interface in constraint on '%s'", param.Name.Name)
			}
			if _, exists := i.interfaces[ifaceName]; !exists {
				return fmt.Errorf("Unknown interface '%s' in constraint on '%s'", ifaceName, param.Name.Name)
			}
		}
	}
	return nil
}

func (i *Interpreter) evaluateStructLiteral(lit *ast.StructLiteral, env *runtime.Environment) (runtime.Value, error) {
	if lit.StructType == nil {
		return nil, fmt.Errorf("Struct literal requires explicit struct type in this milestone")
	}
	structName := lit.StructType.Name
	defValue, err := env.Get(structName)
	if err != nil {
		return nil, err
	}
	structDefVal, err := toStructDefinitionValue(defValue, structName)
	if err != nil {
		return nil, err
	}
	structDef := structDefVal.Node
	if structDef == nil {
		return nil, fmt.Errorf("struct definition '%s' unavailable", structName)
	}
	explicitTypeArgs := append([]ast.TypeExpression(nil), lit.TypeArguments...)
	if lit.IsPositional {
		if structDef.Kind != ast.StructKindPositional && structDef.Kind != ast.StructKindSingleton {
			return nil, fmt.Errorf("Positional struct literal not allowed for struct '%s'", structName)
		}
		if len(lit.Fields) != len(structDef.Fields) {
			return nil, fmt.Errorf("Struct '%s' expects %d fields, got %d", structName, len(structDef.Fields), len(lit.Fields))
		}
		values := make([]runtime.Value, len(lit.Fields))
		for idx, field := range lit.Fields {
			val, err := i.evaluateExpression(field.Value, env)
			if err != nil {
				return nil, err
			}
			values[idx] = val
		}
		typeArgs, err := i.resolveStructTypeArguments(structDef, explicitTypeArgs, nil, nil, values)
		if err != nil {
			return nil, err
		}
		if err := i.enforceStructConstraints(structDef, typeArgs, structName); err != nil {
			return nil, err
		}
		if structName == "Array" {
			fieldMap := make(map[string]runtime.Value, len(values))
			for idx, field := range structDef.Fields {
				if field != nil && field.Name != nil && idx < len(values) {
					fieldMap[field.Name.Name] = values[idx]
				}
			}
			return i.arrayValueFromStructFields(fieldMap)
		}
		return &runtime.StructInstanceValue{Definition: structDefVal, Positional: values, TypeArguments: typeArgs}, nil
	}
	updateCount := len(lit.FunctionalUpdateSources)
	if structDef.Kind == ast.StructKindPositional && updateCount == 0 {
		return nil, fmt.Errorf("Named struct literal not allowed for positional struct '%s'", structName)
	}
	if updateCount > 0 && structDef.Kind == ast.StructKindPositional {
		return nil, fmt.Errorf("Functional update only supported for named structs")
	}
	fields := make(map[string]runtime.Value)
	var baseStruct *runtime.StructInstanceValue
	for idx, srcExpr := range lit.FunctionalUpdateSources {
		base, err := i.evaluateExpression(srcExpr, env)
		if err != nil {
			return nil, err
		}
		instance, ok := base.(*runtime.StructInstanceValue)
		if !ok {
			return nil, fmt.Errorf("Functional update source must be a struct instance")
		}
		if instance.Definition == nil || instance.Definition.Node == nil || instance.Definition.Node.ID == nil || instance.Definition.Node.ID.Name != structName {
			return nil, fmt.Errorf("Functional update source must be same struct type")
		}
		if instance.Fields == nil {
			return nil, fmt.Errorf("Functional update only supported for named structs")
		}
		if idx == 0 {
			baseStruct = instance
		} else if baseStruct != nil {
			if len(baseStruct.TypeArguments) != len(instance.TypeArguments) {
				return nil, fmt.Errorf("Functional update sources must share type arguments")
			}
			for argIdx := range baseStruct.TypeArguments {
				if !typeExpressionsEqual(baseStruct.TypeArguments[argIdx], instance.TypeArguments[argIdx]) {
					return nil, fmt.Errorf("Functional update sources must share type arguments")
				}
			}
		}
		for k, v := range instance.Fields {
			fields[k] = v
		}
	}
	for _, f := range lit.Fields {
		name := ""
		if f.Name != nil {
			name = f.Name.Name
		} else if f.IsShorthand {
			if ident, ok := f.Value.(*ast.Identifier); ok {
				name = ident.Name
			}
		}
		if name == "" {
			return nil, fmt.Errorf("Named struct field initializer must have a field name")
		}
		var val runtime.Value
		var err error
		if f.IsShorthand && f.Value == nil {
			val, err = env.Get(name)
		} else {
			val, err = i.evaluateExpression(f.Value, env)
		}
		if err != nil {
			return nil, err
		}
		fields[name] = val
	}
	if structDef.Kind == ast.StructKindNamed {
		required := make(map[string]struct{}, len(structDef.Fields))
		for _, defField := range structDef.Fields {
			if defField.Name != nil {
				required[defField.Name.Name] = struct{}{}
			}
		}
		for k := range fields {
			delete(required, k)
		}
		if len(required) > 0 {
			for missing := range required {
				return nil, fmt.Errorf("Missing field '%s' for struct '%s'", missing, structName)
			}
		}
	}
	typeArgs, err := i.resolveStructTypeArguments(structDef, explicitTypeArgs, baseStruct, fields, nil)
	if err != nil {
		return nil, err
	}
	if baseStruct != nil && len(baseStruct.TypeArguments) > 0 && len(typeArgs) > 0 {
		if len(baseStruct.TypeArguments) != len(typeArgs) {
			return nil, fmt.Errorf("Functional update must use same type arguments as source")
		}
		for idx := range baseStruct.TypeArguments {
			if !typeExpressionsEqual(baseStruct.TypeArguments[idx], typeArgs[idx]) {
				return nil, fmt.Errorf("Functional update must use same type arguments as source")
			}
		}
	}
	if err := i.enforceStructConstraints(structDef, typeArgs, structName); err != nil {
		return nil, err
	}
	if structName == "Array" {
		return i.arrayValueFromStructFields(fields)
	}
	return &runtime.StructInstanceValue{Definition: structDefVal, Fields: fields, TypeArguments: typeArgs}, nil
}

func (i *Interpreter) resolveStructTypeArguments(def *ast.StructDefinition, explicit []ast.TypeExpression, base *runtime.StructInstanceValue, named map[string]runtime.Value, positional []runtime.Value) ([]ast.TypeExpression, error) {
	if def == nil {
		return nil, fmt.Errorf("Struct definition missing")
	}
	structName := "<anonymous>"
	if def.ID != nil && def.ID.Name != "" {
		structName = def.ID.Name
	}
	genericCount := len(def.GenericParams)
	if genericCount == 0 {
		if len(explicit) > 0 {
			return nil, fmt.Errorf("Type '%s' does not accept type arguments", structName)
		}
		if base != nil && len(base.TypeArguments) > 0 {
			return nil, fmt.Errorf("Type '%s' does not accept type arguments", structName)
		}
		return nil, nil
	}
	if len(explicit) > 0 {
		if len(explicit) != genericCount {
			return nil, fmt.Errorf("Type '%s' expects %d type arguments, got %d", structName, genericCount, len(explicit))
		}
		return append([]ast.TypeExpression(nil), explicit...), nil
	}
	if base != nil {
		if len(base.TypeArguments) != genericCount {
			return nil, fmt.Errorf("Type '%s' expects %d type arguments, got %d", structName, genericCount, len(base.TypeArguments))
		}
		return append([]ast.TypeExpression(nil), base.TypeArguments...), nil
	}
	inferred := i.inferStructTypeArguments(def, named, positional)
	return inferred, nil
}

func (i *Interpreter) inferStructTypeArguments(def *ast.StructDefinition, named map[string]runtime.Value, positional []runtime.Value) []ast.TypeExpression {
	if def == nil || len(def.GenericParams) == 0 {
		return nil
	}
	bindings := make(map[string]ast.TypeExpression)
	genericNames := genericNameSet(def.GenericParams)
	switch def.Kind {
	case ast.StructKindPositional, ast.StructKindSingleton:
		for idx, field := range def.Fields {
			if field == nil || field.FieldType == nil || idx >= len(positional) {
				continue
			}
			actual := i.typeExpressionFromValue(positional[idx])
			if actual == nil {
				continue
			}
			matchTypeExpressionTemplate(field.FieldType, actual, genericNames, bindings)
		}
	default:
		for _, field := range def.Fields {
			if field == nil || field.FieldType == nil || field.Name == nil {
				continue
			}
			val, ok := named[field.Name.Name]
			if !ok {
				continue
			}
			actual := i.typeExpressionFromValue(val)
			if actual == nil {
				continue
			}
			matchTypeExpressionTemplate(field.FieldType, actual, genericNames, bindings)
		}
	}
	typeArgs := make([]ast.TypeExpression, len(def.GenericParams))
	for idx, gp := range def.GenericParams {
		if gp != nil && gp.Name != nil {
			if bound, ok := bindings[gp.Name.Name]; ok {
				typeArgs[idx] = bound
				continue
			}
		}
		typeArgs[idx] = ast.NewWildcardTypeExpression()
	}
	return typeArgs
}

func (i *Interpreter) enforceStructConstraints(def *ast.StructDefinition, typeArgs []ast.TypeExpression, structName string) error {
	if def == nil || len(def.GenericParams) == 0 {
		return nil
	}
	constraints := collectConstraintSpecs(def.GenericParams, def.WhereClause)
	if len(constraints) == 0 {
		return nil
	}
	bindings, err := mapTypeArguments(def.GenericParams, typeArgs, fmt.Sprintf("instantiating %s", structName))
	if err != nil {
		return err
	}
	return i.enforceConstraintSpecs(constraints, bindings)
}
