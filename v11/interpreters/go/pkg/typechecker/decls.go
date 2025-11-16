package typechecker

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter10-go/pkg/ast"
)

var reservedTypeNames = map[string]struct{}{
	"bool":     {},
	"string":   {},
	"char":     {},
	"nil":      {},
	"void":     {},
	"Self":     {},
	"i8":       {},
	"i16":      {},
	"i32":      {},
	"i64":      {},
	"i128":     {},
	"isize":    {},
	"u8":       {},
	"u16":      {},
	"u32":      {},
	"u64":      {},
	"u128":     {},
	"usize":    {},
	"f32":      {},
	"f64":      {},
	"Array":    {},
	"Map":      {},
	"Range":    {},
	"Iterator": {},
	"Result":   {},
	"Option":   {},
	"Proc":     {},
	"Future":   {},
	"Channel":  {},
	"Mutex":    {},
	"Error":    {},
}

type typeIdentifierOccurrence struct {
	name            string
	node            ast.Node
	fromWhereClause bool
}

type inferenceSkipReason int

const (
	skipNone inferenceSkipReason = iota
	skipAlreadyKnown
	skipReserved
	skipKnownType
)

// declarationCollector walks statements to populate the global environment.
type declarationCollector struct {
	env         *Environment
	origins     map[ast.Node]string
	declNodes   map[string]ast.Node
	diags       []Diagnostic
	impls       []ImplementationSpec
	methodSets  []MethodSetSpec
	obligations []ConstraintObligation
	exports     []exportRecord
}

func (c *Checker) collectDeclarations(module *ast.Module) []Diagnostic {
	builtinEnv := NewEnvironment(nil)
	registerBuiltins(builtinEnv)
	rootEnv := NewEnvironment(builtinEnv)
	if c.preludeEnv != nil {
		c.preludeEnv.ForEach(func(name string, typ Type) {
			rootEnv.Define(name, typ)
		})
	}
	collector := &declarationCollector{
		env:       rootEnv,
		origins:   c.nodeOrigins,
		declNodes: make(map[string]ast.Node),
	}
	// Register built-in primitives in the global scope for convenience.
	collector.env.Define("true", PrimitiveType{Kind: PrimitiveBool})
	collector.env.Define("false", PrimitiveType{Kind: PrimitiveBool})

	for _, stmt := range module.Body {
		collector.registerTypeDeclaration(stmt)
	}
	for _, stmt := range module.Body {
		collector.visitStatement(stmt)
	}

	// Store the global environment for later lookups.
	c.global = collector.env
	c.implementations = collector.impls
	c.methodSets = collector.methodSets
	c.obligations = collector.obligations
	c.publicDeclarations = collector.exports
	return collector.diags
}

func (c *declarationCollector) registerTypeDeclaration(stmt ast.Statement) {
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
				Variants:   make([]Type, 0, len(s.Variants)),
			}
			if len(s.Variants) > 0 {
				if paramScope == nil {
					paramScope = make(map[string]Type)
				}
				for _, variant := range s.Variants {
					if variant == nil {
						continue
					}
					unionType.Variants = append(unionType.Variants, c.resolveTypeExpression(variant, paramScope))
				}
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
				InterfaceName:   s.ID.Name,
				TypeParams:      params,
				Where:           where,
				Methods:         methods,
				SelfTypePattern: s.SelfTypePattern,
			}
			c.declare(s.ID.Name, ifaceType, s)
		}
	case *ast.TypeAliasDefinition:
		c.collectTypeAliasDefinition(s)
	}
}

func (c *declarationCollector) visitStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
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
	if name == "" || node == nil {
		return
	}
	if prev, exists := c.declNodes[name]; exists {
		location := formatNodeLocation(prev, c.origins)
		msg := fmt.Sprintf("typechecker: duplicate declaration '%s' (previous declaration at %s)", name, location)
		c.diags = append(c.diags, Diagnostic{Message: msg, Node: node})
		return
	}
	c.env.Define(name, typ)
	c.declNodes[name] = node
	if shouldExportTopLevel(node) {
		c.exports = append(c.exports, exportRecord{name: name, node: node})
	}
}

func shouldExportTopLevel(node ast.Node) bool {
	switch def := node.(type) {
	case *ast.StructDefinition:
		return def != nil && def.ID != nil && !def.IsPrivate
	case *ast.UnionDefinition:
		return def != nil && def.ID != nil && !def.IsPrivate
	case *ast.InterfaceDefinition:
		return def != nil && def.ID != nil && !def.IsPrivate
	case *ast.FunctionDefinition:
		if def == nil || def.ID == nil {
			return false
		}
		if def.IsPrivate || def.IsMethodShorthand {
			return false
		}
		return true
	case *ast.TypeAliasDefinition:
		return def != nil && def.ID != nil && !def.IsPrivate
	default:
		return false
	}
}

func formatNodeLocation(node ast.Node, origins map[ast.Node]string) string {
	if node == nil {
		return "<unknown location>"
	}
	path := "<unknown file>"
	if origins != nil {
		if origin, ok := origins[node]; ok && origin != "" {
			path = origin
		}
	}
	span := node.Span()
	line := span.Start.Line
	column := span.Start.Column
	if line <= 0 {
		line = 0
	}
	if column <= 0 {
		column = 0
	}
	return fmt.Sprintf("%s:%d:%d", path, line, column)
}

func (c *declarationCollector) collectTypeAliasDefinition(def *ast.TypeAliasDefinition) {
	if def == nil || def.ID == nil || def.ID.Name == "" {
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
			case "string":
				return PrimitiveType{Kind: PrimitiveString}
			case "char":
				return PrimitiveType{Kind: PrimitiveChar}
			case "nil":
				return PrimitiveType{Kind: PrimitiveNil}
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
		if simple, ok := t.Base.(*ast.SimpleTypeExpression); ok && simple.Name != nil {
			if _, exists := typeParams[simple.Name.Name]; !exists {
				if decl, ok := c.env.Lookup(simple.Name.Name); ok {
					if alias, ok := decl.(AliasType); ok {
						args := make([]Type, len(t.Arguments))
						for i, arg := range t.Arguments {
							args[i] = c.resolveTypeExpression(arg, typeParams)
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
			TypeParam:       name,
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

func (c *declarationCollector) ensureFunctionGenericInference(def *ast.FunctionDefinition, scope map[string]Type) {
	if def == nil {
		return
	}
	occs := collectFunctionTypeOccurrences(def)
	inferred := c.selectInferredGenericParameters(occs, def.GenericParams, scope)
	if len(inferred) == 0 {
		return
	}
	paramMap := make(map[string]*ast.GenericParameter, len(inferred))
	for _, param := range inferred {
		if param == nil || param.Name == nil {
			continue
		}
		param.IsInferred = true
		paramMap[param.Name.Name] = param
	}
	def.WhereClause = hoistWhereConstraints(def.WhereClause, paramMap)
	def.GenericParams = append(def.GenericParams, inferred...)
}

func (c *declarationCollector) ensureSignatureGenericInference(sig *ast.FunctionSignature, scope map[string]Type) {
	if sig == nil {
		return
	}
	occs := collectSignatureTypeOccurrences(sig)
	inferred := c.selectInferredGenericParameters(occs, sig.GenericParams, scope)
	if len(inferred) == 0 {
		return
	}
	paramMap := make(map[string]*ast.GenericParameter, len(inferred))
	for _, param := range inferred {
		if param == nil || param.Name == nil {
			continue
		}
		param.IsInferred = true
		paramMap[param.Name.Name] = param
	}
	sig.WhereClause = hoistWhereConstraints(sig.WhereClause, paramMap)
	sig.GenericParams = append(sig.GenericParams, inferred...)
}

func (c *declarationCollector) selectInferredGenericParameters(
	occs []typeIdentifierOccurrence,
	existing []*ast.GenericParameter,
	scope map[string]Type,
) []*ast.GenericParameter {
	if len(occs) == 0 {
		return nil
	}
	known := make(map[string]struct{})
	for name := range scope {
		if name == "" {
			continue
		}
		known[name] = struct{}{}
	}
	for _, param := range existing {
		if param == nil || param.Name == nil {
			continue
		}
		known[param.Name.Name] = struct{}{}
	}
	var inferred []*ast.GenericParameter
	reportedKnownType := make(map[string]bool)
	for _, occ := range occs {
		infer, reason := c.shouldInferGenericParameter(occ.name, known)
		if !infer {
			if reason == skipKnownType && occ.fromWhereClause && occ.name != "" && !reportedKnownType[occ.name] {
				msg := fmt.Sprintf("typechecker: cannot infer type parameter '%s' because a type with the same name exists; declare it explicitly or qualify the type", occ.name)
				c.diags = append(c.diags, Diagnostic{Message: msg, Node: occ.node})
				reportedKnownType[occ.name] = true
			}
			continue
		}
		param := newInferredGenericParameter(occ.name, occ.node)
		inferred = append(inferred, param)
		known[occ.name] = struct{}{}
	}
	return inferred
}

func (c *declarationCollector) shouldInferGenericParameter(name string, known map[string]struct{}) (bool, inferenceSkipReason) {
	if name == "" {
		return false, skipReserved
	}
	if _, exists := known[name]; exists {
		return false, skipAlreadyKnown
	}
	if strings.ContainsRune(name, '.') {
		return false, skipReserved
	}
	if _, reserved := reservedTypeNames[name]; reserved {
		return false, skipReserved
	}
	if c.env != nil {
		if decl, exists := c.env.Lookup(name); exists {
			if _, isFn := decl.(FunctionType); !isFn {
				return false, skipKnownType
			}
		}
	}
	return true, skipNone
}

func newInferredGenericParameter(name string, node ast.Node) *ast.GenericParameter {
	id := ast.NewIdentifier(name)
	if node != nil {
		ast.SetSpan(id, node.Span())
	}
	param := ast.NewGenericParameter(id, nil)
	param.IsInferred = true
	if node != nil {
		ast.SetSpan(param, node.Span())
	}
	return param
}

func collectFunctionTypeOccurrences(def *ast.FunctionDefinition) []typeIdentifierOccurrence {
	var occs []typeIdentifierOccurrence
	if def == nil {
		return occs
	}
	for _, param := range def.Params {
		if param == nil {
			continue
		}
		collectTypeExpressionOccurrences(param.ParamType, &occs)
	}
	collectTypeExpressionOccurrences(def.ReturnType, &occs)
	for _, clause := range def.WhereClause {
		if clause == nil || clause.TypeParam == nil {
			continue
		}
		occs = append(occs, typeIdentifierOccurrence{name: clause.TypeParam.Name, node: clause.TypeParam, fromWhereClause: true})
	}
	return occs
}

func collectSignatureTypeOccurrences(sig *ast.FunctionSignature) []typeIdentifierOccurrence {
	var occs []typeIdentifierOccurrence
	if sig == nil {
		return occs
	}
	for _, param := range sig.Params {
		if param == nil {
			continue
		}
		collectTypeExpressionOccurrences(param.ParamType, &occs)
	}
	collectTypeExpressionOccurrences(sig.ReturnType, &occs)
	for _, clause := range sig.WhereClause {
		if clause == nil || clause.TypeParam == nil {
			continue
		}
		occs = append(occs, typeIdentifierOccurrence{name: clause.TypeParam.Name, node: clause.TypeParam, fromWhereClause: true})
	}
	return occs
}

func collectTypeExpressionOccurrences(expr ast.TypeExpression, occs *[]typeIdentifierOccurrence) {
	if expr == nil {
		return
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			*occs = append(*occs, typeIdentifierOccurrence{name: t.Name.Name, node: t.Name})
		}
	case *ast.GenericTypeExpression:
		collectTypeExpressionOccurrences(t.Base, occs)
		for _, arg := range t.Arguments {
			collectTypeExpressionOccurrences(arg, occs)
		}
	case *ast.FunctionTypeExpression:
		for _, param := range t.ParamTypes {
			collectTypeExpressionOccurrences(param, occs)
		}
		collectTypeExpressionOccurrences(t.ReturnType, occs)
	case *ast.NullableTypeExpression:
		collectTypeExpressionOccurrences(t.InnerType, occs)
	case *ast.ResultTypeExpression:
		collectTypeExpressionOccurrences(t.InnerType, occs)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			collectTypeExpressionOccurrences(member, occs)
		}
	}
}

func hoistWhereConstraints(where []*ast.WhereClauseConstraint, inferred map[string]*ast.GenericParameter) []*ast.WhereClauseConstraint {
	if len(where) == 0 || len(inferred) == 0 {
		return where
	}
	kept := make([]*ast.WhereClauseConstraint, 0, len(where))
	for _, clause := range where {
		if clause == nil || clause.TypeParam == nil {
			kept = append(kept, clause)
			continue
		}
		name := clause.TypeParam.Name
		param := inferred[name]
		if param == nil {
			kept = append(kept, clause)
			continue
		}
		if len(clause.Constraints) > 0 {
			param.Constraints = append(param.Constraints, clause.Constraints...)
		}
	}
	return kept
}

func registerBuiltins(env *Environment) {
	if env == nil {
		return
	}

	nilType := PrimitiveType{Kind: PrimitiveNil}
	boolType := PrimitiveType{Kind: PrimitiveBool}
	i32Type := IntegerType{Suffix: "i32"}
	i64Type := IntegerType{Suffix: "i64"}
	anyType := UnknownType{}
	stringType := PrimitiveType{Kind: PrimitiveString}
	charType := PrimitiveType{Kind: PrimitiveChar}
	byteArrayType := ArrayType{Element: i32Type}

	procYield := FunctionType{
		Params: nil,
		Return: nilType,
	}
	procCancelled := FunctionType{
		Params: nil,
		Return: boolType,
	}
	procFlush := FunctionType{
		Params: nil,
		Return: nilType,
	}
	procPendingTasks := FunctionType{
		Params: nil,
		Return: i32Type,
	}
	printFn := FunctionType{
		Params: []Type{UnknownType{}},
		Return: nilType,
	}

	env.Define("proc_yield", procYield)
	env.Define("proc_cancelled", procCancelled)
	env.Define("proc_flush", procFlush)
	env.Define("proc_pending_tasks", procPendingTasks)
	env.Define("print", printFn)

	env.Define("__able_channel_new", FunctionType{
		Params: []Type{i32Type},
		Return: i64Type,
	})
	env.Define("__able_channel_send", FunctionType{
		Params: []Type{anyType, anyType},
		Return: nilType,
	})
	env.Define("__able_channel_receive", FunctionType{
		Params: []Type{anyType},
		Return: anyType,
	})
	env.Define("__able_channel_try_send", FunctionType{
		Params: []Type{anyType, anyType},
		Return: boolType,
	})
	env.Define("__able_channel_try_receive", FunctionType{
		Params: []Type{anyType},
		Return: anyType,
	})
	env.Define("__able_channel_close", FunctionType{
		Params: []Type{anyType},
		Return: nilType,
	})
	env.Define("__able_channel_is_closed", FunctionType{
		Params: []Type{anyType},
		Return: boolType,
	})

	env.Define("__able_mutex_new", FunctionType{
		Params: nil,
		Return: i64Type,
	})
	env.Define("__able_mutex_lock", FunctionType{
		Params: []Type{i64Type},
		Return: nilType,
	})
	env.Define("__able_mutex_unlock", FunctionType{
		Params: []Type{i64Type},
		Return: nilType,
	})

	env.Define("__able_string_from_builtin", FunctionType{
		Params: []Type{stringType},
		Return: byteArrayType,
	})
	env.Define("__able_string_to_builtin", FunctionType{
		Params: []Type{byteArrayType},
		Return: stringType,
	})
	env.Define("__able_char_from_codepoint", FunctionType{
		Params: []Type{i32Type},
		Return: charType,
	})

	env.Define("__able_hasher_create", FunctionType{
		Params: nil,
		Return: i64Type,
	})
	env.Define("__able_hasher_write", FunctionType{
		Params: []Type{i64Type, stringType},
		Return: nilType,
	})
	env.Define("__able_hasher_finish", FunctionType{
		Params: []Type{i64Type},
		Return: i64Type,
	})

	env.Define("Display", InterfaceType{
		InterfaceName: "Display",
	})
	env.Define("Clone", InterfaceType{
		InterfaceName: "Clone",
	})
	procErrorFields := map[string]Type{
		"details": stringType,
	}
	env.Define("ProcError", StructType{
		StructName: "ProcError",
		Fields:     procErrorFields,
	})
}

func (c *declarationCollector) functionTypeFromDefinition(def *ast.FunctionDefinition, parentScope map[string]Type, owner string, node ast.Node) FunctionType {
	scope := copyTypeScope(parentScope)
	c.ensureFunctionGenericInference(def, scope)
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
	interfaceName := identifierName(def.InterfaceName)

	params, paramScope := c.convertGenericParams(def.GenericParams)
	scope := copyTypeScope(paramScope)

	targetType := c.resolveTypeExpression(def.TargetType, scope)
	if targetType == nil {
		targetType = UnknownType{}
	}
	scope["Self"] = targetType
	targetLabel := nonEmpty(typeName(targetType))

	interfaceArgs := make([]Type, len(def.InterfaceArgs))
	for i, arg := range def.InterfaceArgs {
		interfaceArgs[i] = c.resolveTypeExpression(arg, scope)
	}

	var ifaceType InterfaceType
	if interfaceName != "" {
		if decl, ok := c.env.Lookup(interfaceName); ok {
			if typed, ok := decl.(InterfaceType); ok {
				ifaceType = typed
			} else {
				c.diags = append(c.diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: impl references '%s' which is not an interface", interfaceName),
					Node:    def,
				})
			}
		} else {
			c.diags = append(c.diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: impl references unknown interface '%s'", interfaceName),
				Node:    def,
			})
		}
	}

	if interfaceName != "" {
		expectedParams := len(ifaceType.TypeParams)
		providedArgs := len(def.InterfaceArgs)
		if expectedParams == 0 && providedArgs > 0 {
			c.diags = append(c.diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: impl %s does not accept type arguments", interfaceName),
				Node:    def,
			})
		}
		if expectedParams > 0 {
			if providedArgs == 0 {
				c.diags = append(c.diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: impl %s for %s requires %d interface type argument(s)", interfaceName, typeName(targetType), expectedParams),
					Node:    def,
				})
			} else if providedArgs != expectedParams {
				c.diags = append(c.diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: impl %s for %s expected %d interface type argument(s), got %d", interfaceName, typeName(targetType), expectedParams, providedArgs),
					Node:    def,
				})
			}
		}
	}

	implGenericNames := collectGenericParamNameSet(params)
	if ifaceType.InterfaceName != "" {
		targetValid := c.validateImplementationSelfTypePattern(def, ifaceType, interfaceName, targetLabel, implGenericNames)
		if !targetValid {
			return nil, diags
		}
	}

	where := c.convertWhereClause(def.WhereClause, scope)
	implLabel := fmt.Sprintf("impl %s for %s", nonEmpty(interfaceName), targetLabel)

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
		fnType = applyImplicitSelfParam(fn, fnType, targetType)
		methods[fn.ID.Name] = fnType
	}

	spec := &ImplementationSpec{
		ImplName:      identifierName(def.ImplName),
		InterfaceName: interfaceName,
		TypeParams:    params,
		Target:        targetType,
		InterfaceArgs: interfaceArgs,
		Methods:       methods,
		Where:         where,
		UnionVariants: collectUnionVariantLabelsFromType(targetType),
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
		fnType = applyImplicitSelfParam(fn, fnType, targetType)
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
	return spec, diags
}

func functionName(def *ast.FunctionDefinition) string {
	if def != nil && def.ID != nil && def.ID.Name != "" {
		return def.ID.Name
	}
	return "<anonymous>"
}

func collectUnionVariantLabelsFromType(t Type) []string {
	literal, ok := t.(UnionLiteralType)
	if !ok {
		return nil
	}
	seen := make(map[string]struct{}, len(literal.Members))
	labels := make([]string, 0, len(literal.Members))
	for _, member := range literal.Members {
		label := formatType(member)
		if label == "" || label == "<unknown>" {
			label = typeName(member)
		}
		if label == "" {
			label = "<unknown>"
		}
		if _, exists := seen[label]; exists {
			continue
		}
		seen[label] = struct{}{}
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return labels
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
		for idx, constraint := range param.Constraints {
			if constraint == nil || isUnknownType(constraint) {
				continue
			}
			var constraintNode ast.Node = node
			if idx >= 0 && idx < len(param.ConstraintNodes) && param.ConstraintNodes[idx] != nil {
				if n, ok := param.ConstraintNodes[idx].(ast.Node); ok {
					constraintNode = n
				}
			}
			obligations = append(obligations, ConstraintObligation{
				Owner:      owner,
				TypeParam:  param.Name,
				Constraint: constraint,
				Subject:    TypeParameterType{ParameterName: param.Name},
				Node:       constraintNode,
			})
		}
	}
	for _, clause := range where {
		if clause.TypeParam == "" {
			continue
		}
		for idx, constraint := range clause.Constraints {
			if constraint == nil || isUnknownType(constraint) {
				continue
			}
			var constraintNode ast.Node = node
			if idx >= 0 && idx < len(clause.ConstraintNodes) && clause.ConstraintNodes[idx] != nil {
				if n, ok := clause.ConstraintNodes[idx].(ast.Node); ok {
					constraintNode = n
				}
			}
			obligations = append(obligations, ConstraintObligation{
				Owner:      owner,
				TypeParam:  clause.TypeParam,
				Constraint: constraint,
				Subject:    TypeParameterType{ParameterName: clause.TypeParam},
				Node:       constraintNode,
			})
		}
	}
	return obligations
}

func applyImplicitSelfParam(def *ast.FunctionDefinition, fnType FunctionType, target Type) FunctionType {
	if def == nil || len(def.Params) == 0 || len(fnType.Params) == 0 {
		return fnType
	}
	firstParam := def.Params[0]
	if firstParam == nil {
		return fnType
	}
	if firstParam.ParamType != nil && !isUnknownType(fnType.Params[0]) {
		return fnType
	}
	name := functionParameterName(firstParam)
	if name == "" || !strings.EqualFold(name, "self") {
		return fnType
	}
	if !isUnknownType(fnType.Params[0]) {
		return fnType
	}
	if target == nil || isUnknownType(target) {
		fnType.Params[0] = TypeParameterType{ParameterName: "Self"}
	} else {
		fnType.Params[0] = target
	}
	return fnType
}

func functionParameterName(param *ast.FunctionParameter) string {
	if param == nil || param.Name == nil {
		return ""
	}
	switch name := param.Name.(type) {
	case *ast.Identifier:
		return name.Name
	default:
		return ""
	}
}
