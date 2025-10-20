package typechecker

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

// declarationCollector walks statements to populate the global environment.
type declarationCollector struct {
	env         *Environment
	diags       []Diagnostic
	impls       []ImplementationSpec
	methodSets  []MethodSetSpec
	obligations []ConstraintObligation
}

func (c *Checker) collectDeclarations(module *ast.Module) []Diagnostic {
	builtinEnv := NewEnvironment(nil)
	registerBuiltins(builtinEnv)
	collector := &declarationCollector{env: NewEnvironment(builtinEnv)}
	// Register built-in primitives in the global scope for convenience.
	collector.env.Define("true", PrimitiveType{Kind: PrimitiveBool})
	collector.env.Define("false", PrimitiveType{Kind: PrimitiveBool})

	for _, stmt := range module.Body {
		collector.visitStatement(stmt)
	}

	// Store the global environment for later lookups.
	c.global = collector.env
	c.implementations = collector.impls
	c.methodSets = collector.methodSets
	c.obligations = collector.obligations
	return collector.diags
}

func (c *declarationCollector) visitStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.StructDefinition:
		if s.ID != nil {
			params, paramScope := c.convertGenericParams(s.GenericParams)
			where := c.convertWhereClause(s.WhereClause, paramScope)
			fields, positional := c.collectStructFields(s, paramScope)
			structType := StructType{
				StructName: s.ID.Name,
				TypeParams: params,
				Fields:     fields,
				Positional: positional,
				Where:      where,
			}
			c.declare(s.ID.Name, structType, s)
		}
	case *ast.UnionDefinition:
		if s.ID != nil {
			params, paramScope := c.convertGenericParams(s.GenericParams)
			where := c.convertWhereClause(s.WhereClause, paramScope)
			unionType := UnionType{
				UnionName:  s.ID.Name,
				TypeParams: params,
				Where:      where,
			}
			c.declare(s.ID.Name, unionType, s)
		}
	case *ast.InterfaceDefinition:
		if s.ID != nil {
			params, paramScope := c.convertGenericParams(s.GenericParams)
			where := c.convertWhereClause(s.WhereClause, paramScope)
			if paramScope == nil {
				paramScope = make(map[string]Type)
			}
			if _, exists := paramScope["Self"]; !exists {
				paramScope["Self"] = TypeParameterType{ParameterName: "Self"}
			}
			methods := c.collectInterfaceMethods(s, paramScope)
			ifaceType := InterfaceType{
				InterfaceName: s.ID.Name,
				TypeParams:    params,
				Where:         where,
				Methods:       methods,
			}
			c.declare(s.ID.Name, ifaceType, s)
		}
	case *ast.FunctionDefinition:
		if s.ID != nil {
			owner := fmt.Sprintf("fn %s", functionName(s))
			sig := c.functionTypeFromDefinition(s, nil, owner, s)
			c.declare(s.ID.Name, sig, s)
		}
	case *ast.ImplementationDefinition:
		spec, diags := c.collectImplementationDefinition(s)
		c.diags = append(c.diags, diags...)
		if spec != nil {
			c.impls = append(c.impls, *spec)
		}
	case *ast.MethodsDefinition:
		spec, diags := c.collectMethodsDefinition(s)
		c.diags = append(c.diags, diags...)
		if spec != nil {
			c.methodSets = append(c.methodSets, *spec)
		}
	}
}

func (c *declarationCollector) declare(name string, typ Type, node ast.Node) {
	if _, exists := c.env.symbols[name]; exists {
		msg := fmt.Sprintf("typechecker: duplicate declaration '%s'", name)
		c.diags = append(c.diags, Diagnostic{Message: msg, Node: node})
		return
	}
	c.env.Define(name, typ)
}

func (c *declarationCollector) resolveTypeExpression(expr ast.TypeExpression, typeParams map[string]Type) Type {
	if expr == nil {
		return UnknownType{}
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			name := t.Name.Name
			if local, ok := typeParams[name]; ok {
				return local
			}
			switch name {
			case "bool":
				return PrimitiveType{Kind: PrimitiveBool}
			case "string":
				return PrimitiveType{Kind: PrimitiveString}
			case "char":
				return PrimitiveType{Kind: PrimitiveChar}
			case "nil":
				return PrimitiveType{Kind: PrimitiveNil}
			case "i32", "i64", "isize", "u32", "u64", "usize":
				return IntegerType{Suffix: name}
			case "f32", "f64":
				return FloatType{Suffix: name}
			default:
				if decl, ok := c.env.Lookup(name); ok {
					return decl
				}
				return StructType{StructName: name}
			}
		}
	case *ast.GenericTypeExpression:
		base := c.resolveTypeExpression(t.Base, typeParams)
		args := make([]Type, len(t.Arguments))
		for i, arg := range t.Arguments {
			args[i] = c.resolveTypeExpression(arg, typeParams)
		}
		if st, ok := base.(StructType); ok && st.StructName == "Array" {
			var elem Type = UnknownType{}
			if len(args) > 0 && args[0] != nil {
				elem = args[0]
			}
			return ArrayType{Element: elem}
		}
		return AppliedType{Base: base, Arguments: args}
	case *ast.FunctionTypeExpression:
		params := make([]Type, len(t.ParamTypes))
		for i, param := range t.ParamTypes {
			params[i] = c.resolveTypeExpression(param, typeParams)
		}
		return FunctionType{Params: params, Return: c.resolveTypeExpression(t.ReturnType, typeParams)}
	case *ast.NullableTypeExpression:
		return NullableType{Inner: c.resolveTypeExpression(t.InnerType, typeParams)}
	case *ast.ResultTypeExpression:
		inner := c.resolveTypeExpression(t.InnerType, typeParams)
		return AppliedType{
			Base:      StructType{StructName: "Result"},
			Arguments: []Type{inner},
		}
	case *ast.UnionTypeExpression:
		members := make([]Type, len(t.Members))
		for i, member := range t.Members {
			members[i] = c.resolveTypeExpression(member, typeParams)
		}
		return UnionLiteralType{Members: members}
	}
	return UnknownType{}
}

func (c *declarationCollector) convertGenericParams(params []*ast.GenericParameter) ([]GenericParamSpec, map[string]Type) {
	if len(params) == 0 {
		return nil, map[string]Type{}
	}
	specs := make([]GenericParamSpec, 0, len(params))
	typeScope := make(map[string]Type, len(params))
	for _, param := range params {
		if param == nil || param.Name == nil {
			continue
		}
		name := param.Name.Name
		typeParam := TypeParameterType{ParameterName: name}
		typeScope[name] = typeParam

		constraints := make([]Type, 0, len(param.Constraints))
		for _, constraint := range param.Constraints {
			if constraint == nil {
				continue
			}
			constraints = append(constraints, c.resolveTypeExpression(constraint.InterfaceType, typeScope))
		}
		specs = append(specs, GenericParamSpec{
			Name:        name,
			Constraints: constraints,
		})
	}
	return specs, typeScope
}

func (c *declarationCollector) convertWhereClause(where []*ast.WhereClauseConstraint, typeParams map[string]Type) []WhereConstraintSpec {
	if len(where) == 0 {
		return nil
	}
	specs := make([]WhereConstraintSpec, 0, len(where))
	for _, clause := range where {
		if clause == nil || clause.TypeParam == nil {
			continue
		}
		name := clause.TypeParam.Name
		constraints := make([]Type, 0, len(clause.Constraints))
		for _, constraint := range clause.Constraints {
			if constraint == nil {
				continue
			}
			constraints = append(constraints, c.resolveTypeExpression(constraint.InterfaceType, typeParams))
		}
		specs = append(specs, WhereConstraintSpec{
			TypeParam:   name,
			Constraints: constraints,
		})
	}
	return specs
}

func (c *declarationCollector) collectStructFields(def *ast.StructDefinition, scope map[string]Type) (map[string]Type, []Type) {
	if def == nil || len(def.Fields) == 0 {
		return nil, nil
	}
	fields := make(map[string]Type, len(def.Fields))
	positional := make([]Type, len(def.Fields))
	for idx, field := range def.Fields {
		if field == nil || field.FieldType == nil {
			continue
		}
		typ := c.resolveTypeExpression(field.FieldType, scope)
		positional[idx] = typ
		if field.Name != nil {
			fields[field.Name.Name] = typ
		}
	}
	return fields, positional
}

func (c *declarationCollector) collectInterfaceMethods(def *ast.InterfaceDefinition, baseScope map[string]Type) map[string]FunctionType {
	if def == nil || len(def.Signatures) == 0 {
		return nil
	}
	methods := make(map[string]FunctionType, len(def.Signatures))
	for _, sig := range def.Signatures {
		if sig == nil || sig.Name == nil {
			continue
		}
		name := sig.Name.Name
		if _, exists := methods[name]; exists {
			c.diags = append(c.diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: duplicate interface method '%s'", name),
				Node:    sig,
			})
			continue
		}
		fnType, diags := c.convertFunctionSignature(sig, baseScope)
		c.diags = append(c.diags, diags...)
		methods[name] = fnType
	}
	return methods
}

func (c *declarationCollector) convertFunctionSignature(sig *ast.FunctionSignature, baseScope map[string]Type) (FunctionType, []Diagnostic) {
	scope := copyTypeScope(baseScope)
	typeParams, localScope := c.convertGenericParams(sig.GenericParams)
	for name, typ := range localScope {
		scope[name] = typ
	}
	paramTypes := make([]Type, len(sig.Params))
	for idx, param := range sig.Params {
		if param == nil {
			paramTypes[idx] = UnknownType{}
			continue
		}
		paramTypes[idx] = c.resolveTypeExpression(param.ParamType, scope)
	}
	var returnType Type = UnknownType{}
	if sig.ReturnType != nil {
		returnType = c.resolveTypeExpression(sig.ReturnType, scope)
	}
	where := c.convertWhereClause(sig.WhereClause, scope)
	return FunctionType{
		Params:     paramTypes,
		Return:     returnType,
		TypeParams: typeParams,
		Where:      where,
	}, nil
}

func copyTypeScope(scope map[string]Type) map[string]Type {
	if scope == nil {
		return make(map[string]Type)
	}
	clone := make(map[string]Type, len(scope))
	for name, typ := range scope {
		clone[name] = typ
	}
	return clone
}

func registerBuiltins(env *Environment) {
	if env == nil {
		return
	}

	procYield := FunctionType{
		Params: nil,
		Return: PrimitiveType{Kind: PrimitiveNil},
	}
	procCancelled := FunctionType{
		Params: nil,
		Return: PrimitiveType{Kind: PrimitiveBool},
	}
	procFlush := FunctionType{
		Params: nil,
		Return: PrimitiveType{Kind: PrimitiveNil},
	}
	printFn := FunctionType{
		Params: []Type{UnknownType{}},
		Return: PrimitiveType{Kind: PrimitiveNil},
	}

	env.Define("proc_yield", procYield)
	env.Define("proc_cancelled", procCancelled)
	env.Define("proc_flush", procFlush)
	env.Define("print", printFn)
}

func (c *declarationCollector) functionTypeFromDefinition(def *ast.FunctionDefinition, parentScope map[string]Type, owner string, node ast.Node) FunctionType {
	scope := copyTypeScope(parentScope)
	typeParams, localScope := c.convertGenericParams(def.GenericParams)
	for name, typ := range localScope {
		scope[name] = typ
	}

	paramTypes := make([]Type, len(def.Params))
	for idx, param := range def.Params {
		if param == nil {
			paramTypes[idx] = UnknownType{}
			continue
		}
		paramTypes[idx] = c.resolveTypeExpression(param.ParamType, scope)
	}

	var returnType Type = UnknownType{}
	if def.ReturnType != nil {
		returnType = c.resolveTypeExpression(def.ReturnType, scope)
	}

	where := c.convertWhereClause(def.WhereClause, scope)
	fnType := FunctionType{
		Params:     paramTypes,
		Return:     returnType,
		TypeParams: typeParams,
		Where:      where,
	}
	fnType.Obligations = obligationsFromSpecs(owner, typeParams, where, node)
	c.obligations = append(c.obligations, fnType.Obligations...)
	return fnType
}

func (c *declarationCollector) collectImplementationDefinition(def *ast.ImplementationDefinition) (*ImplementationSpec, []Diagnostic) {
	if def == nil {
		return nil, nil
	}

	var diags []Diagnostic
	if def.InterfaceName == nil {
		diags = append(diags, Diagnostic{
			Message: "typechecker: implementation requires an interface name",
			Node:    def,
		})
		return nil, diags
	}

	params, paramScope := c.convertGenericParams(def.GenericParams)
	scope := copyTypeScope(paramScope)

	targetType := c.resolveTypeExpression(def.TargetType, scope)
	if targetType == nil {
		targetType = UnknownType{}
	}
	scope["Self"] = targetType

	interfaceArgs := make([]Type, len(def.InterfaceArgs))
	for i, arg := range def.InterfaceArgs {
		interfaceArgs[i] = c.resolveTypeExpression(arg, scope)
	}

	var ifaceType InterfaceType
	if def.InterfaceName != nil && def.InterfaceName.Name != "" {
		if decl, ok := c.env.Lookup(def.InterfaceName.Name); ok {
			if typed, ok := decl.(InterfaceType); ok {
				ifaceType = typed
			} else {
				c.diags = append(c.diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: impl references '%s' which is not an interface", def.InterfaceName.Name),
					Node:    def,
				})
			}
		} else {
			c.diags = append(c.diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: impl references unknown interface '%s'", def.InterfaceName.Name),
				Node:    def,
			})
		}
	}

	if def.InterfaceName != nil && def.InterfaceName.Name != "" {
		expectedParams := len(ifaceType.TypeParams)
		providedArgs := len(def.InterfaceArgs)
		if expectedParams == 0 && providedArgs > 0 {
			c.diags = append(c.diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: impl %s does not accept type arguments", def.InterfaceName.Name),
				Node:    def,
			})
		}
		if expectedParams > 0 {
			if providedArgs == 0 {
				c.diags = append(c.diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: impl %s for %s requires %d interface type argument(s)", def.InterfaceName.Name, typeName(targetType), expectedParams),
					Node:    def,
				})
			} else if providedArgs != expectedParams {
				c.diags = append(c.diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: impl %s for %s expected %d interface type argument(s), got %d", def.InterfaceName.Name, typeName(targetType), expectedParams, providedArgs),
					Node:    def,
				})
			}
		}
	}

	where := c.convertWhereClause(def.WhereClause, scope)
	implLabel := fmt.Sprintf("impl %s for %s", def.InterfaceName.Name, nonEmpty(typeName(targetType)))

	methods := make(map[string]FunctionType, len(def.Definitions))
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			diags = append(diags, Diagnostic{
				Message: "typechecker: implementation method requires a name",
				Node:    fn,
			})
			continue
		}
		if _, exists := methods[fn.ID.Name]; exists {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: duplicate method '%s' in implementation", fn.ID.Name),
				Node:    fn,
			})
			continue
		}
		methodOwner := fmt.Sprintf("%s::%s", implLabel, functionName(fn))
		fnType := c.functionTypeFromDefinition(fn, scope, methodOwner, fn)
		methods[fn.ID.Name] = fnType
	}

	spec := &ImplementationSpec{
		ImplName:      identifierName(def.ImplName),
		InterfaceName: def.InterfaceName.Name,
		TypeParams:    params,
		Target:        targetType,
		InterfaceArgs: interfaceArgs,
		Methods:       methods,
		Where:         where,
		Definition:    def,
	}
	spec.Obligations = obligationsFromSpecs(implLabel, params, where, def)
	c.obligations = append(c.obligations, spec.Obligations...)

	return spec, diags
}

func (c *declarationCollector) collectMethodsDefinition(def *ast.MethodsDefinition) (*MethodSetSpec, []Diagnostic) {
	if def == nil {
		return nil, nil
	}

	params, paramScope := c.convertGenericParams(def.GenericParams)
	scope := copyTypeScope(paramScope)

	targetType := c.resolveTypeExpression(def.TargetType, scope)
	if targetType == nil {
		targetType = UnknownType{}
	}
	scope["Self"] = targetType

	where := c.convertWhereClause(def.WhereClause, scope)
	methodsLabel := fmt.Sprintf("methods for %s", nonEmpty(typeName(targetType)))

	var diags []Diagnostic
	methods := make(map[string]FunctionType, len(def.Definitions))
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			diags = append(diags, Diagnostic{
				Message: "typechecker: method definition requires a name",
				Node:    fn,
			})
			continue
		}
		if _, exists := methods[fn.ID.Name]; exists {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: duplicate method '%s' for target", fn.ID.Name),
				Node:    fn,
			})
			continue
		}
		methodOwner := fmt.Sprintf("%s::%s", methodsLabel, functionName(fn))
		fnType := c.functionTypeFromDefinition(fn, scope, methodOwner, fn)
		methods[fn.ID.Name] = fnType
	}

	spec := &MethodSetSpec{
		TypeParams: params,
		Target:     targetType,
		Methods:    methods,
		Where:      where,
		Definition: def,
	}
	spec.Obligations = obligationsFromSpecs(methodsLabel, params, where, def)
	c.obligations = append(c.obligations, spec.Obligations...)
	return spec, diags
}

func functionName(def *ast.FunctionDefinition) string {
	if def != nil && def.ID != nil && def.ID.Name != "" {
		return def.ID.Name
	}
	return "<anonymous>"
}

func identifierName(id *ast.Identifier) string {
	if id == nil {
		return ""
	}
	return id.Name
}

func nonEmpty(value string) string {
	if value == "" {
		return "<unknown>"
	}
	return value
}

func obligationsFromSpecs(owner string, params []GenericParamSpec, where []WhereConstraintSpec, node ast.Node) []ConstraintObligation {
	if owner == "" {
		owner = "<unknown>"
	}
	var obligations []ConstraintObligation
	for _, param := range params {
		if param.Name == "" {
			continue
		}
		for _, constraint := range param.Constraints {
			if constraint == nil || isUnknownType(constraint) {
				continue
			}
			obligations = append(obligations, ConstraintObligation{
				Owner:      owner,
				TypeParam:  param.Name,
				Constraint: constraint,
				Subject:    TypeParameterType{ParameterName: param.Name},
				Node:       node,
			})
		}
	}
	for _, clause := range where {
		if clause.TypeParam == "" {
			continue
		}
		for _, constraint := range clause.Constraints {
			if constraint == nil || isUnknownType(constraint) {
				continue
			}
			obligations = append(obligations, ConstraintObligation{
				Owner:      owner,
				TypeParam:  clause.TypeParam,
				Constraint: constraint,
				Subject:    TypeParameterType{ParameterName: clause.TypeParam},
				Node:       node,
			})
		}
	}
	return obligations
}
