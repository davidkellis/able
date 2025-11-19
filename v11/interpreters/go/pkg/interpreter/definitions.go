package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

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

func (i *Interpreter) evaluateImplementationDefinition(def *ast.ImplementationDefinition, env *runtime.Environment) (runtime.Value, error) {
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
	methods := make(map[string]*runtime.FunctionValue)
	hasExplicit := false
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			return nil, fmt.Errorf("Implementation method requires identifier")
		}
		methods[fn.ID.Name] = &runtime.FunctionValue{Declaration: fn, Closure: env}
		hasExplicit = true
	}
	if ifaceDef.Node != nil {
		for _, sig := range ifaceDef.Node.Signatures {
			if sig == nil || sig.Name == nil {
				continue
			}
			name := sig.Name.Name
			if _, ok := methods[name]; ok {
				continue
			}
			if sig.DefaultImpl == nil {
				continue
			}
			defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
			methods[name] = &runtime.FunctionValue{Declaration: defaultDef, Closure: ifaceDef.Env}
		}
	}
	constraintSpecs := collectConstraintSpecs(def.GenericParams, def.WhereClause)
	baseConstraintSig := constraintSignature(constraintSpecs)
	targetDescription := typeExpressionToString(def.TargetType)
	for _, variant := range variants {
		if def.ImplName == nil {
			if err := i.registerUnnamedImpl(ifaceName, variant, unionSignatures, baseConstraintSig, targetDescription); err != nil {
				return nil, err
			}
			entry := implEntry{
				interfaceName: ifaceName,
				methods:       methods,
				definition:    def,
				argTemplates:  variant.argTemplates,
				genericParams: def.GenericParams,
				whereClause:   def.WhereClause,
				defaultOnly:   !hasExplicit,
			}
			if len(unionSignatures) > 0 {
				entry.unionVariants = append([]string(nil), unionSignatures...)
			}
			i.implMethods[variant.typeName] = append(i.implMethods[variant.typeName], entry)
			if ifaceName == "Range" {
				i.registerRangeImplementation(entry, def.InterfaceArgs)
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
	simpleType, ok := def.TargetType.(*ast.SimpleTypeExpression)
	if !ok || simpleType.Name == nil {
		return nil, fmt.Errorf("MethodsDefinition requires simple target type")
	}
	typeName := simpleType.Name.Name
	bucket, ok := i.inherentMethods[typeName]
	if !ok {
		bucket = make(map[string]*runtime.FunctionValue)
		i.inherentMethods[typeName] = bucket
	}
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			return nil, fmt.Errorf("Method definition requires identifier")
		}
		bucket[fn.ID.Name] = &runtime.FunctionValue{Declaration: fn, Closure: env}
	}
	return runtime.NilValue{}, nil
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
		typeArgs, err := i.resolveStructTypeArguments(structDef, explicitTypeArgs, nil)
		if err != nil {
			return nil, err
		}
		if err := i.enforceStructConstraints(structDef, typeArgs, structName); err != nil {
			return nil, err
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
		val, err := i.evaluateExpression(f.Value, env)
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
	typeArgs, err := i.resolveStructTypeArguments(structDef, explicitTypeArgs, baseStruct)
	if err != nil {
		return nil, err
	}
	if err := i.enforceStructConstraints(structDef, typeArgs, structName); err != nil {
		return nil, err
	}
	return &runtime.StructInstanceValue{Definition: structDefVal, Fields: fields, TypeArguments: typeArgs}, nil
}

func (i *Interpreter) resolveStructTypeArguments(def *ast.StructDefinition, explicit []ast.TypeExpression, base *runtime.StructInstanceValue) ([]ast.TypeExpression, error) {
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
	return nil, fmt.Errorf("Type '%s' requires type arguments", structName)
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
