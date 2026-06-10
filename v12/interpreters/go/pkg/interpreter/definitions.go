package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func canonicalTypeName(env *runtime.Environment, name string) string {
	if env == nil || isReservedScalarPrimitiveTypeName(name) {
		return name
	}
	val, ok := env.Lookup(name)
	if !ok {
		return name
	}
	switch v := val.(type) {
	case *runtime.StructDefinitionValue:
		if v.Node != nil && v.Node.ID != nil && v.Node.ID.Name != "" {
			return v.Node.ID.Name
		}
	case *runtime.UnionDefinitionValue:
		if v.Node != nil && v.Node.ID != nil && v.Node.ID.Name != "" {
			return v.Node.ID.Name
		}
	case runtime.UnionDefinitionValue:
		if v.Node != nil && v.Node.ID != nil && v.Node.ID.Name != "" {
			return v.Node.ID.Name
		}
	case *runtime.InterfaceDefinitionValue:
		if v.QualifiedName != "" {
			return v.QualifiedName
		}
		if v.Node != nil && v.Node.ID != nil && v.Node.ID.Name != "" {
			return v.Node.ID.Name
		}
	}
	return name
}

func isReservedScalarPrimitiveTypeName(name string) bool {
	switch name {
	case "bool", "char", "nil", "void",
		"i8", "i16", "i32", "i64", "i128", "isize",
		"u8", "u16", "u32", "u64", "u128", "usize",
		"f32", "f64":
		return true
	default:
		return false
	}
}

func canonicalizeExpandedTypeExpression(expr ast.TypeExpression, env *runtime.Environment) ast.TypeExpression {
	if expr == nil {
		return nil
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return expr
		}
		name := canonicalTypeName(env, t.Name.Name)
		if name == t.Name.Name {
			return expr
		}
		return ast.Ty(name)
	case *ast.GenericTypeExpression:
		base := canonicalizeExpandedTypeExpression(t.Base, env)
		var args []ast.TypeExpression
		for idx, arg := range t.Arguments {
			if arg == nil {
				continue
			}
			canonicalArg := canonicalizeExpandedTypeExpression(arg, env)
			if canonicalArg == arg {
				continue
			}
			if args == nil {
				args = append([]ast.TypeExpression(nil), t.Arguments...)
			}
			args[idx] = canonicalArg
		}
		if base == t.Base && args == nil {
			return expr
		}
		if args == nil {
			return ast.Gen(base, t.Arguments...)
		}
		return ast.Gen(base, args...)
	case *ast.NullableTypeExpression:
		inner := canonicalizeExpandedTypeExpression(t.InnerType, env)
		if inner == t.InnerType {
			return expr
		}
		return ast.Nullable(inner)
	case *ast.ResultTypeExpression:
		inner := canonicalizeExpandedTypeExpression(t.InnerType, env)
		if inner == t.InnerType {
			return expr
		}
		return ast.Result(inner)
	case *ast.UnionTypeExpression:
		var members []ast.TypeExpression
		for idx, member := range t.Members {
			if member == nil {
				continue
			}
			canonicalMember := canonicalizeExpandedTypeExpression(member, env)
			if canonicalMember == member {
				continue
			}
			if members == nil {
				members = append([]ast.TypeExpression(nil), t.Members...)
			}
			members[idx] = canonicalMember
		}
		if members == nil {
			return expr
		}
		return ast.UnionT(members...)
	case *ast.FunctionTypeExpression:
		ret := canonicalizeExpandedTypeExpression(t.ReturnType, env)
		var params []ast.TypeExpression
		for idx, param := range t.ParamTypes {
			canonicalParam := canonicalizeExpandedTypeExpression(param, env)
			if canonicalParam == param {
				continue
			}
			if params == nil {
				params = append([]ast.TypeExpression(nil), t.ParamTypes...)
			}
			params[idx] = canonicalParam
		}
		if params == nil && ret == t.ReturnType {
			return expr
		}
		if params == nil {
			return ast.FnType(t.ParamTypes, ret)
		}
		return ast.FnType(params, ret)
	default:
		return expr
	}
}

func canonicalizeTypeExpression(expr ast.TypeExpression, env *runtime.Environment, aliases map[string]*ast.TypeAliasDefinition) ast.TypeExpression {
	if expr == nil {
		return nil
	}
	expanded := expandTypeAliases(expr, aliases, nil)
	return canonicalizeExpandedTypeExpression(expanded, env)
}

func (i *Interpreter) canonicalizeTypeExpressionCached(expr ast.TypeExpression, env *runtime.Environment, hasAlias bool) ast.TypeExpression {
	if expr == nil {
		return nil
	}
	if !hasAlias || i == nil || len(i.typeAliases) == 0 {
		if env == nil {
			return expr
		}
		return canonicalizeExpandedTypeExpression(expr, env)
	}
	expanded := i.expandTypeAliasesCached(expr)
	if env == nil {
		return expanded
	}
	return canonicalizeExpandedTypeExpression(expanded, env)
}

func (i *Interpreter) lowerFunctionDefinitionBytecode(def *ast.FunctionDefinition) (*bytecodeProgram, error) {
	return i.lowerFunctionDefinitionBytecodeWithEnv(def, nil)
}

func (i *Interpreter) evaluateFunctionDefinition(def *ast.FunctionDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Function definition requires identifier")
	}
	if err := i.validateGenericConstraints(def); err != nil {
		return nil, err
	}
	fnVal := &runtime.FunctionValue{Declaration: def, Closure: env}
	if def.Body != nil {
		program, err := i.lowerFunctionDefinitionBytecodeWithEnv(def, env)
		if err != nil {
			if i.execMode == execModeBytecode {
				return nil, err
			}
		} else {
			setFunctionBytecodeProgram(fnVal, program)
		}
	}
	i.defineInEnv(env, def.ID.Name, fnVal)
	i.registerSymbol(def.ID.Name, fnVal)
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
	if _, ok := env.Lookup(name); ok {
		return runtime.NilValue{}, nil
	}
	if def.Target == ast.HostTargetGo && strings.TrimSpace(def.Body) == "" && !i.isKernelExtern(name) {
		return nil, raiseSignal{value: runtime.ErrorValue{Message: fmt.Sprintf("extern function %s for %s must provide a host body", name, def.Target)}}
	}
	pkgName := i.currentPackage
	if pkgName == "" {
		pkgName = "<root>"
	}
	native := i.makeExternNative(def, pkgName)
	if native == nil {
		return runtime.NilValue{}, nil
	}
	env.Define(name, native)
	i.registerSymbol(name, native)
	return runtime.NilValue{}, nil
}

func (i *Interpreter) makeExternNative(def *ast.ExternFunctionBody, pkgName string) *runtime.NativeFunctionValue {
	if def == nil || def.Signature == nil || def.Signature.ID == nil {
		return nil
	}
	name := def.Signature.ID.Name
	arity := len(def.Signature.Params)
	if def.Target != ast.HostTargetGo {
		return nil
	}
	return &runtime.NativeFunctionValue{
		Name:        name,
		Arity:       arity,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return i.invokeExternHostFunction(pkgName, def, args)
		},
	}
}

func (i *Interpreter) evaluateStructDefinition(def *ast.StructDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Struct definition requires identifier")
	}
	structVal := newStructDefinitionValue(def)
	i.defineInEnv(env, def.ID.Name, structVal)
	env.DefineStruct(def.ID.Name, structVal)
	i.registerSymbol(def.ID.Name, structVal)
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateUnionDefinition(def *ast.UnionDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Union definition requires identifier")
	}
	unionVal := runtime.UnionDefinitionValue{Node: def}
	i.defineInEnv(env, def.ID.Name, unionVal)
	i.unionDefinitions[def.ID.Name] = &unionVal
	i.registerSymbol(def.ID.Name, unionVal)
	return runtime.NilValue{}, nil
}

// RegisterUnionDefinition registers a union definition in the interpreter's
// union definition table for use by matchesType pattern matching.
func (i *Interpreter) RegisterUnionDefinition(name string, def *runtime.UnionDefinitionValue) {
	if i == nil || name == "" || def == nil {
		return
	}
	i.unionDefinitions[name] = def
}

// RegisterInterfaceDefinition registers an interface definition in the
// interpreter's interface table for use by matchesType and impl resolution.
func (i *Interpreter) RegisterInterfaceDefinition(name string, def *runtime.InterfaceDefinitionValue) {
	if i == nil || name == "" || def == nil {
		return
	}
	identity := interfaceDefinitionIdentity(def)
	if identity == "" {
		identity = name
	}
	if _, exists := i.interfaces[identity]; !exists {
		i.interfaces[identity] = def
	}
	if _, exists := i.interfaces[name]; !exists {
		i.interfaces[name] = def
	}
}

func interfaceDefinitionIdentity(def *runtime.InterfaceDefinitionValue) string {
	if def == nil {
		return ""
	}
	if def.QualifiedName != "" {
		return def.QualifiedName
	}
	if def.Node != nil && def.Node.ID != nil {
		return def.Node.ID.Name
	}
	return ""
}

func (i *Interpreter) canonicalInterfaceName(name string) string {
	if i == nil || name == "" {
		return name
	}
	def, ok := i.interfaces[name]
	if !ok {
		return name
	}
	if identity := interfaceDefinitionIdentity(def); identity != "" {
		return identity
	}
	return name
}

// RegisterPackageSymbol registers a symbol in the interpreter's package
// registry so isKnownTypeName can recognize it.
func (i *Interpreter) RegisterPackageSymbol(pkgName string, name string, val runtime.Value) {
	if i == nil || name == "" {
		return
	}
	bucket, ok := i.packageRegistry[pkgName]
	if !ok {
		bucket = make(map[string]runtime.Value)
		i.packageRegistry[pkgName] = bucket
	}
	bucket[name] = val
	i.updateKnownTypeNameCacheForPackageSymbol(name, val)
}

func (i *Interpreter) evaluateInterfaceDefinition(def *ast.InterfaceDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Interface definition requires identifier")
	}
	identity := i.qualifiedName(def.ID.Name)
	if identity == "" {
		identity = def.ID.Name
	}
	ifaceVal := &runtime.InterfaceDefinitionValue{Node: def, Env: env, QualifiedName: identity}
	i.defineInEnv(env, def.ID.Name, ifaceVal)
	i.RegisterInterfaceDefinition(def.ID.Name, ifaceVal)
	// A dynamic redefinition replaces the package-qualified definition, as the
	// previous short-name registry did, while the compatibility short alias
	// remains bound to the first visible nominal definition.
	i.interfaces[identity] = ifaceVal
	i.registerSymbol(def.ID.Name, ifaceVal)
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateImplementationDefinition(def *ast.ImplementationDefinition, env *runtime.Environment, isBuiltin bool) (runtime.Value, error) {
	if def.InterfaceName == nil {
		return nil, fmt.Errorf("Implementation requires interface name")
	}
	ifaceName := canonicalTypeName(env, def.InterfaceName.Name)
	ifaceDef, ok := i.interfaces[ifaceName]
	if !ok {
		return nil, fmt.Errorf("Interface '%s' is not defined", ifaceName)
	}
	canonicalTarget := canonicalizeTypeExpression(def.TargetType, env, i.typeAliases)
	canonicalDef := *def
	canonicalDef.InterfaceName = ast.ID(ifaceName)
	canonicalDef.TargetType = canonicalTarget
	variants, unionSignatures, err := expandImplementationTargetVariants(canonicalDef.TargetType, i.typeAliases)
	if err != nil {
		return nil, err
	}
	if len(variants) == 0 {
		return nil, fmt.Errorf("Implementation target must reference at least one concrete type")
	}
	mergedGenerics := i.mergeImplementationGenerics(&canonicalDef, env)
	methodSet := &runtime.MethodSet{
		TargetType:    canonicalDef.TargetType,
		GenericParams: mergedGenerics,
		WhereClause:   canonicalDef.WhereClause,
	}
	methods := make(map[string]runtime.Value)
	implCtx := &implMethodContext{
		implName:      "",
		interfaceName: canonicalDef.InterfaceName.Name,
		target:        canonicalDef.TargetType,
		methods:       methods,
	}
	if canonicalDef.ImplName != nil {
		implCtx.implName = canonicalDef.ImplName.Name
	}
	implTarget := expandTypeAliases(canonicalDef.TargetType, i.typeAliases, nil)
	if implTarget == nil {
		implTarget = canonicalDef.TargetType
	}
	hasExplicit := false
	for _, fn := range canonicalDef.Definitions {
		if fn == nil || fn.ID == nil {
			return nil, fmt.Errorf("Implementation method requires identifier")
		}
		fnEnv := runtime.NewEnvironment(env)
		fnEnv.SetRuntimeData(implCtx)
		fnVal := &runtime.FunctionValue{Declaration: fn, Closure: fnEnv, MethodPriority: -1, MethodSet: methodSet}
		if program, err := i.lowerFunctionDefinitionBytecodeWithMethodSetEnv(fn, fnEnv, methodSet); err != nil {
			if i.execMode == execModeBytecode {
				return nil, err
			}
		} else {
			setFunctionBytecodeProgram(fnVal, program)
		}
		mergeFunctionLike(methods, fn.ID.Name, fnVal)
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
			defaultEnv := runtime.NewEnvironment(ifaceDef.Env)
			defaultEnv.SetRuntimeData(implCtx)
			selfBindings := make(map[string]ast.TypeExpression)
			i.bindInterfaceSelfPatternBindings(selfBindings, ifaceName, implTarget)
			i.defineTypeBindingValues(defaultEnv, selfBindings)
			defaultVal := &runtime.FunctionValue{Declaration: defaultDef, Closure: defaultEnv, MethodPriority: -1, MethodSet: methodSet}
			if program, err := i.lowerFunctionDefinitionBytecodeWithMethodSetEnv(defaultDef, defaultEnv, methodSet); err != nil {
				if i.execMode == execModeBytecode {
					return nil, err
				}
			} else {
				setFunctionBytecodeProgram(defaultVal, program)
			}
			mergeFunctionLike(methods, name, defaultVal)
		}
	}
	attachImplMethodContext(methods, implCtx)
	constraintSpecs := collectConstraintSpecs(canonicalDef.GenericParams, canonicalDef.WhereClause)
	baseConstraintSig := constraintSignature(constraintSpecs, func(expr ast.TypeExpression) string {
		return typeExpressionToString(expandTypeAliases(expr, i.typeAliases, nil))
	})
	targetDescription := typeExpressionToString(expandTypeAliases(canonicalDef.TargetType, i.typeAliases, nil))
	genericNames := genericNameSet(mergedGenerics)
	for _, variant := range variants {
		if canonicalDef.ImplName == nil {
			isGenericTarget := false
			if len(genericNames) > 0 {
				_, isGenericTarget = genericNames[variant.typeName]
			}
			if !isGenericTarget {
				if err := i.registerUnnamedImpl(ifaceName, canonicalDef.InterfaceArgs, variant, unionSignatures, baseConstraintSig, targetDescription, isBuiltin); err != nil {
					return nil, err
				}
			}
			entry := implEntry{
				interfaceName:      ifaceName,
				methods:            methods,
				definition:         &canonicalDef,
				registrationTarget: def.TargetType,
				argTemplates:       variant.argTemplates,
				genericParams:      mergedGenerics,
				whereClause:        canonicalDef.WhereClause,
				defaultOnly:        !hasExplicit,
				isBuiltin:          isBuiltin,
			}
			if len(unionSignatures) > 0 {
				entry.unionVariants = append([]string(nil), unionSignatures...)
			}
			if isGenericTarget {
				i.genericImpls = append(i.genericImpls, entry)
				i.invalidateMethodCache()
				i.noteIndexImplementation(ifaceName, variant.typeName, true)
			} else {
				i.implMethods[variant.typeName] = append(i.implMethods[variant.typeName], entry)
				i.invalidateMethodCache()
				i.noteIndexImplementation(ifaceName, variant.typeName, false)
				if ifaceName == "Range" {
					i.registerRangeImplementation(entry, canonicalDef.InterfaceArgs)
				}
			}
		}
	}
	if canonicalDef.ImplName != nil {
		name := canonicalDef.ImplName.Name
		implVal := runtime.ImplementationNamespaceValue{
			Name:          canonicalDef.ImplName,
			InterfaceName: canonicalDef.InterfaceName,
			TargetType:    canonicalDef.TargetType,
			Methods:       methods,
			IsPrivate:     canonicalDef.IsPrivate,
		}
		i.defineInEnv(env, name, implVal)
		i.registerSymbol(name, implVal)
	}
	return runtime.NilValue{}, nil
}

func attachImplMethodContext(methods map[string]runtime.Value, ctx *implMethodContext) {
	if ctx == nil {
		return
	}
	for _, method := range methods {
		switch fn := method.(type) {
		case *runtime.FunctionValue:
			if fn == nil {
				continue
			}
			wrapped := runtime.NewEnvironment(fn.Closure)
			wrapped.SetRuntimeData(ctx)
			fn.Closure = wrapped
		case *runtime.FunctionOverloadValue:
			if fn == nil {
				continue
			}
			for _, entry := range fn.Overloads {
				if entry == nil {
					continue
				}
				wrapped := runtime.NewEnvironment(entry.Closure)
				wrapped.SetRuntimeData(ctx)
				entry.Closure = wrapped
			}
		}
	}
}

func (i *Interpreter) evaluateMethodsDefinition(def *ast.MethodsDefinition, env *runtime.Environment) (runtime.Value, error) {
	var typeName string
	target := canonicalizeTypeExpression(def.TargetType, env, i.typeAliases)
	switch t := target.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return nil, fmt.Errorf("MethodsDefinition requires simple target type")
		}
		typeName = canonicalTypeName(env, t.Name.Name)
	case *ast.GenericTypeExpression:
		base, ok := t.Base.(*ast.SimpleTypeExpression)
		if !ok || base.Name == nil {
			return nil, fmt.Errorf("MethodsDefinition requires simple target type")
		}
		typeName = canonicalTypeName(env, base.Name.Name)
	default:
		return nil, fmt.Errorf("MethodsDefinition requires simple target type")
	}
	bucket, ok := i.inherentMethods[typeName]
	if !ok {
		bucket = make(map[string]runtime.Value)
		i.inherentMethods[typeName] = bucket
	}
	mergedGenerics := i.mergeMethodsGenerics(def, target, env)
	methodSet := &runtime.MethodSet{
		TargetType:    target,
		GenericParams: mergedGenerics,
		WhereClause:   def.WhereClause,
	}
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			return nil, fmt.Errorf("Method definition requires identifier")
		}
		expectsSelf := functionDefinitionExpectsSelf(fn)
		exportedName := fn.ID.Name
		if !expectsSelf {
			exportedName = fmt.Sprintf("%s.%s", typeName, fn.ID.Name)
		}
		fnVal := &runtime.FunctionValue{Declaration: fn, Closure: env, TypeQualified: !expectsSelf, MethodSet: methodSet}
		if program, err := i.lowerFunctionDefinitionBytecodeWithMethodSetEnv(fn, env, methodSet); err != nil {
			if i.execMode == execModeBytecode {
				return nil, err
			}
		} else {
			setFunctionBytecodeProgram(fnVal, program)
		}
		mergeFunctionLike(bucket, fn.ID.Name, fnVal)
		i.defineInEnv(env, exportedName, fnVal)
		i.registerSymbol(exportedName, fnVal)
	}
	return runtime.NilValue{}, nil
}

func functionDefinitionExpectsSelf(def *ast.FunctionDefinition) bool {
	if def == nil {
		return false
	}
	if def.IsMethodShorthand {
		return true
	}
	if len(def.Params) == 0 {
		return false
	}
	first := def.Params[0]
	if first == nil {
		return false
	}
	if ident, ok := first.Name.(*ast.Identifier); ok && strings.EqualFold(ident.Name, "self") {
		return true
	}
	if simple, ok := first.ParamType.(*ast.SimpleTypeExpression); ok && simple.Name != nil && simple.Name.Name == "Self" {
		return true
	}
	return false
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

func (i *Interpreter) mergeMethodsGenerics(def *ast.MethodsDefinition, target ast.TypeExpression, env *runtime.Environment) []*ast.GenericParameter {
	seen := make(map[string]struct{})
	result := make([]*ast.GenericParameter, 0, len(def.GenericParams))
	for _, gp := range def.GenericParams {
		if gp == nil || gp.Name == nil {
			continue
		}
		result = append(result, gp)
		seen[gp.Name.Name] = struct{}{}
	}
	for _, inferred := range i.inferGenericsFromTarget(target, env) {
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
	structDefVal, ok := env.StructDefinition(structName)
	if !ok {
		defValue, found := env.Lookup(structName)
		if !found {
			return nil, fmt.Errorf("Undefined variable '%s'", structName)
		}
		var err error
		structDefVal, err = toStructDefinitionValue(defValue, structName)
		if err != nil {
			return nil, err
		}
	}
	structDef := structDefVal.Node
	if structDef == nil {
		return nil, fmt.Errorf("struct definition '%s' unavailable", structName)
	}
	var explicitTypeArgs []ast.TypeExpression
	if len(lit.TypeArguments) > 0 {
		explicitTypeArgs = append([]ast.TypeExpression(nil), lit.TypeArguments...)
	}
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
		if isSingletonStructDef(structDef) {
			return structDefVal, nil
		}
		return &runtime.StructInstanceValue{Definition: structDefVal, Positional: values, TypeArguments: typeArgs}, nil
	}
	updateCount := len(lit.FunctionalUpdateSources)
	if updateCount == 0 {
		if simpleValue, ok, err := i.evaluateSimpleNamedStructLiteralIfPossible(lit, env, structDefVal, explicitTypeArgs); ok {
			return simpleValue, err
		} else if err != nil {
			return nil, err
		}
	}
	if structDef.Kind == ast.StructKindPositional && updateCount == 0 {
		return nil, fmt.Errorf("Named struct literal not allowed for positional struct '%s'", structName)
	}
	if updateCount > 0 && structDef.Kind == ast.StructKindPositional {
		return nil, fmt.Errorf("Functional update only supported for named structs")
	}
	fieldCapacity := len(lit.Fields)
	if fieldCapacity < len(structDef.Fields) {
		fieldCapacity = len(structDef.Fields)
	}
	fields := make(map[string]runtime.Value, fieldCapacity)
	var baseStruct *runtime.StructInstanceValue
	for idx, srcExpr := range lit.FunctionalUpdateSources {
		base, err := i.evaluateExpression(srcExpr, env)
		if err != nil {
			return nil, err
		}
		instance, ok := base.(*runtime.StructInstanceValue)
		if !ok {
			switch defVal := base.(type) {
			case *runtime.StructDefinitionValue:
				if isSingletonStructDef(structDef) && defVal != nil && defVal.Node != nil && defVal.Node.ID != nil && defVal.Node.ID.Name == structName {
					continue
				}
			case runtime.StructDefinitionValue:
				if isSingletonStructDef(structDef) && defVal.Node != nil && defVal.Node.ID != nil && defVal.Node.ID.Name == structName {
					continue
				}
			}
			return nil, fmt.Errorf("Functional update source must be a struct instance")
		}
		if instance.Definition == nil || instance.Definition.Node == nil || instance.Definition.Node.ID == nil || instance.Definition.Node.ID.Name != structName {
			return nil, fmt.Errorf("Functional update source must be same struct type")
		}
		if !structUsesNamedFieldStorage(instance) {
			return nil, fmt.Errorf("Functional update only supported for named structs")
		}
		if idx == 0 {
			baseStruct = instance
		} else if baseStruct != nil {
			baseTypeArgs := i.resolvedStructInstanceTypeArguments(baseStruct)
			instanceTypeArgs := i.resolvedStructInstanceTypeArguments(instance)
			if len(baseTypeArgs) != len(instanceTypeArgs) {
				return nil, fmt.Errorf("Functional update sources must share type arguments")
			}
			for argIdx := range baseTypeArgs {
				if !typeExpressionsEqual(baseTypeArgs[argIdx], instanceTypeArgs[argIdx]) {
					return nil, fmt.Errorf("Functional update sources must share type arguments")
				}
			}
		}
		sourceFields, ok := structCopyNamedFields(instance)
		if !ok {
			return nil, fmt.Errorf("Functional update only supported for named structs")
		}
		for k, v := range sourceFields {
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
			var ok bool
			val, ok = env.Lookup(name)
			if !ok {
				err = fmt.Errorf("Undefined variable '%s'", name)
			}
		} else {
			val, err = i.evaluateExpression(f.Value, env)
		}
		if err != nil {
			return nil, err
		}
		fields[name] = val
	}
	if structDef.Kind == ast.StructKindNamed {
		for _, defField := range structDef.Fields {
			if defField == nil || defField.Name == nil {
				continue
			}
			if _, ok := fields[defField.Name.Name]; !ok {
				missing := defField.Name.Name
				return nil, fmt.Errorf("Missing field '%s' for struct '%s'", missing, structName)
			}
		}
	}
	typeArgs, err := i.resolveStructTypeArguments(structDef, explicitTypeArgs, baseStruct, fields, nil)
	if err != nil {
		return nil, err
	}
	if baseStruct != nil {
		baseTypeArgs := i.resolvedStructInstanceTypeArguments(baseStruct)
		if len(baseTypeArgs) > 0 && len(typeArgs) > 0 {
			if len(baseTypeArgs) != len(typeArgs) {
				return nil, fmt.Errorf("Functional update must use same type arguments as source")
			}
			for idx := range baseTypeArgs {
				if !typeExpressionsEqual(baseTypeArgs[idx], typeArgs[idx]) {
					return nil, fmt.Errorf("Functional update must use same type arguments as source")
				}
			}
		}
	}
	if err := i.enforceStructConstraints(structDef, typeArgs, structName); err != nil {
		return nil, err
	}
	if structName == "Array" {
		return i.arrayValueFromStructFields(fields)
	}
	if isSingletonStructDef(structDef) {
		return structDefVal, nil
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
		return i.cachedTypeExpressionTuple(explicit), nil
	}
	if base != nil {
		baseTypeArgs := i.resolvedStructInstanceTypeArguments(base)
		if len(baseTypeArgs) != genericCount {
			return nil, fmt.Errorf("Type '%s' expects %d type arguments, got %d", structName, genericCount, len(baseTypeArgs))
		}
		return i.cachedTypeExpressionTuple(baseTypeArgs), nil
	}
	if !structHasGenericConstraints(def) {
		return nil, nil
	}
	inferred := i.inferStructTypeArguments(def, named, positional)
	return inferred, nil
}

func (i *Interpreter) inferStructTypeArguments(def *ast.StructDefinition, named map[string]runtime.Value, positional []runtime.Value) []ast.TypeExpression {
	return i.inferStructTypeArgumentsWithSeen(def, named, positional, nil)
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
