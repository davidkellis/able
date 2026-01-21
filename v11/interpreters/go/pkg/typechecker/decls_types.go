package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (c *declarationCollector) collectTypeAliasDefinition(def *ast.TypeAliasDefinition) {
	if def == nil || def.ID == nil || def.ID.Name == "" {
		return
	}
	if def.ID.Name == "_" {
		c.diags = append(c.diags, Diagnostic{Message: "typechecker: type alias name '_' is reserved", Node: def})
		return
	}
	params, paramScope := c.convertGenericParams(def.GenericParams)
	target := c.resolveTypeExpression(def.TargetType, paramScope)
	if target == nil {
		target = UnknownType{}
	}
	where := c.convertWhereClause(def.WhereClause, paramScope)
	alias := AliasType{
		AliasName:   def.ID.Name,
		TypeParams:  params,
		Target:      target,
		Where:       where,
		Definition:  def,
		Obligations: obligationsFromSpecs(fmt.Sprintf("type alias %s", def.ID.Name), params, where, def),
	}
	c.declare(def.ID.Name, alias, def)
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
			case "IoHandle":
				return PrimitiveType{Kind: PrimitiveIoHandle}
			case "ProcHandle":
				return PrimitiveType{Kind: PrimitiveProcHandle}
			case "string":
				return PrimitiveType{Kind: PrimitiveString}
			case "String":
				return PrimitiveType{Kind: PrimitiveString}
			case "char":
				return PrimitiveType{Kind: PrimitiveChar}
			case "nil":
				return PrimitiveType{Kind: PrimitiveNil}
			case "_":
				return UnknownType{}
			case "i8", "i16", "i32", "i64", "i128", "isize", "u8", "u16", "u32", "u64", "u128", "usize":
				return IntegerType{Suffix: name}
			case "f32", "f64":
				return FloatType{Suffix: name}
			default:
				if decl, ok := c.env.Lookup(name); ok {
					if alias, ok := decl.(AliasType); ok {
						inst, _ := instantiateAlias(alias, nil)
						return inst
					}
					return decl
				}
				return StructType{StructName: name}
			}
		}
	case *ast.GenericTypeExpression:
		var baseName string
		if simple, ok := t.Base.(*ast.SimpleTypeExpression); ok && simple.Name != nil {
			baseName = simple.Name.Name
			if _, exists := typeParams[baseName]; !exists {
				if decl, ok := c.env.Lookup(baseName); ok {
					if alias, ok := decl.(AliasType); ok {
						args := make([]Type, len(t.Arguments))
						for i, arg := range t.Arguments {
							args[i] = c.resolveTypeExpression(arg, typeParams)
						}
						if !shouldSkipTypeArgumentCheck(baseName, c.localTypeNames, c.declNodes) {
							if expected, ok := expectedTypeArgumentCount(baseName, alias); ok && len(args) > expected {
								c.diags = append(c.diags, typeArgumentArityDiagnostic(baseName, expected, len(args), t))
							}
						}
						inst, _ := instantiateAlias(alias, args)
						return inst
					}
				}
			}
		}
		base := c.resolveTypeExpression(t.Base, typeParams)
		args := make([]Type, len(t.Arguments))
		for i, arg := range t.Arguments {
			args[i] = c.resolveTypeExpression(arg, typeParams)
		}
		if baseName != "" && !shouldSkipTypeArgumentCheck(baseName, c.localTypeNames, c.declNodes) {
			if _, known := c.env.Lookup(baseName); known {
				if expected, ok := expectedTypeArgumentCount(baseName, base); ok && len(args) > expected {
					c.diags = append(c.diags, typeArgumentArityDiagnostic(baseName, expected, len(args), t))
				}
			} else if expected, ok := builtinTypeArgumentArity[baseName]; ok && len(args) > expected {
				c.diags = append(c.diags, typeArgumentArityDiagnostic(baseName, expected, len(args), t))
			}
		}
		if unionBase, ok := base.(UnionType); ok {
			return c.instantiateUnionType(unionBase, args)
		}
		if base == nil {
			return UnknownType{}
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
		if decl, ok := c.env.Lookup("Result"); ok {
			if union, ok := decl.(UnionType); ok {
				return c.instantiateUnionType(union, []Type{inner})
			}
			if alias, ok := decl.(AliasType); ok {
				inst, _ := instantiateAlias(alias, []Type{inner})
				return inst
			}
		}
		return AppliedType{
			Base:      StructType{StructName: "Result"},
			Arguments: []Type{inner},
		}
	case *ast.UnionTypeExpression:
		entries := make([]unionMember, 0, len(t.Members))
		for _, member := range t.Members {
			entries = append(entries, unionMember{
				typ:  c.resolveTypeExpression(member, typeParams),
				node: member,
			})
		}
		return normalizeUnionMembers(entries, unionNormalizationOptions{
			warnRedundant: c.warnRedundantUnionMember,
		})
	}
	return UnknownType{}
}

func (c *declarationCollector) instantiateUnionType(union UnionType, args []Type) UnionType {
	return instantiateUnionTypeArgs(union, args)
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
		constraintNodes := make([]ast.TypeExpression, 0, len(param.Constraints))
		for _, constraint := range param.Constraints {
			if constraint == nil {
				continue
			}
			if constraint.InterfaceType == nil {
				continue
			}
			constraints = append(constraints, c.resolveTypeExpression(constraint.InterfaceType, typeScope))
			constraintNodes = append(constraintNodes, constraint.InterfaceType)
		}
		specs = append(specs, GenericParamSpec{
			Name:            name,
			Constraints:     constraints,
			ConstraintNodes: constraintNodes,
			IsInferred:      param.IsInferred,
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
		label := formatTypeExpressionNode(clause.TypeParam)
		subject := c.resolveTypeExpression(clause.TypeParam, typeParams)
		constraints := make([]Type, 0, len(clause.Constraints))
		constraintNodes := make([]ast.TypeExpression, 0, len(clause.Constraints))
		for _, constraint := range clause.Constraints {
			if constraint == nil {
				continue
			}
			if constraint.InterfaceType == nil {
				continue
			}
			constraints = append(constraints, c.resolveTypeExpression(constraint.InterfaceType, typeParams))
			constraintNodes = append(constraintNodes, constraint.InterfaceType)
		}
		specs = append(specs, WhereConstraintSpec{
			TypeParam:       label,
			Subject:         subject,
			Constraints:     constraints,
			ConstraintNodes: constraintNodes,
		})
	}
	return specs
}

func (c *declarationCollector) collectStructFields(def *ast.StructDefinition, scope map[string]Type) (map[string]Type, []Type) {
	if def == nil || len(def.Fields) == 0 {
		return nil, nil
	}
	fields := make(map[string]Type, len(def.Fields))
	var positional []Type
	if def.Kind == ast.StructKindPositional {
		positional = make([]Type, len(def.Fields))
	}
	for idx, field := range def.Fields {
		if field == nil || field.FieldType == nil {
			continue
		}
		typ := c.resolveTypeExpression(field.FieldType, scope)
		if positional != nil {
			positional[idx] = typ
		}
		if field.Name != nil {
			fields[field.Name.Name] = typ
		}
	}
	return fields, positional
}

func (c *declarationCollector) collectInterfaceMethods(def *ast.InterfaceDefinition, baseScope map[string]Type) (map[string]FunctionType, map[string]bool) {
	if def == nil {
		return nil, nil
	}
	methods := make(map[string]FunctionType)
	defaults := make(map[string]bool)
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
		if sig.DefaultImpl != nil {
			defaults[name] = true
		}
	}
	if len(def.BaseInterfaces) > 0 {
		for _, baseExpr := range def.BaseInterfaces {
			if baseExpr == nil {
				continue
			}
			baseType := c.resolveTypeExpression(baseExpr, baseScope)
			iface, args, ok := resolveInterfaceDecl(baseType, nil)
			if !ok {
				if baseName := interfaceBaseName(baseExpr); baseName != "" && c.shouldSkipUnknownInterfaceBase(baseName, baseScope) {
					continue
				}
				c.diags = append(c.diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: interface base must be an interface (got %s)", typeName(baseType)),
					Node:    def,
				})
				continue
			}
			substitution := make(map[string]Type)
			for idx, param := range iface.TypeParams {
				if param.Name == "" {
					continue
				}
				var arg Type = UnknownType{}
				if idx < len(args) && args[idx] != nil {
					arg = args[idx]
				}
				substitution[param.Name] = arg
			}
			for name, method := range iface.Methods {
				if _, exists := methods[name]; exists {
					continue
				}
				if len(substitution) > 0 {
					method = substituteFunctionType(method, substitution)
				}
				methods[name] = method
				if iface.DefaultMethods != nil && iface.DefaultMethods[name] {
					defaults[name] = true
				}
			}
		}
	}
	if len(methods) == 0 {
		return nil, nil
	}
	if len(defaults) == 0 {
		defaults = nil
	}
	return methods, defaults
}

func interfaceBaseName(expr ast.TypeExpression) string {
	if expr == nil {
		return ""
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			return t.Name.Name
		}
	case *ast.GenericTypeExpression:
		if simple, ok := t.Base.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil {
			return simple.Name.Name
		}
	}
	return ""
}

func (c *declarationCollector) shouldSkipUnknownInterfaceBase(name string, scope map[string]Type) bool {
	if name == "" {
		return false
	}
	if scope != nil {
		if _, ok := scope[name]; ok {
			return true
		}
	}
	if c == nil || c.env == nil {
		return false
	}
	if _, ok := c.env.Lookup(name); ok {
		return false
	}
	return true
}

func (c *declarationCollector) convertFunctionSignature(sig *ast.FunctionSignature, baseScope map[string]Type) (FunctionType, []Diagnostic) {
	scope := copyTypeScope(baseScope)
	c.ensureSignatureGenericInference(sig, scope)
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
