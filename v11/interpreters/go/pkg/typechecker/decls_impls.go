package typechecker

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

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
	hasSelfParam := len(def.Params) > 0 && strings.EqualFold(functionParameterName(def.Params[0]), "self")
	if def.IsMethodShorthand && !hasSelfParam {
		paramTypes = append([]Type{UnknownType{}}, paramTypes...)
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

	c.ensureImplementationGenericInference(def)

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

	expectedParams := len(ifaceType.TypeParams)
	interfaceArgs := make([]Type, len(def.InterfaceArgs))
	for i, arg := range def.InterfaceArgs {
		interfaceArgs[i] = c.resolveTypeExpression(arg, scope)
	}

	if interfaceName != "" {
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
			// Allow overload-like duplicates by keeping the first declaration.
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

	if spec.ImplName != "" {
		ns := ImplementationNamespaceType{Impl: spec}
		c.declare(spec.ImplName, ns, def)
	}

	return spec, diags
}

func (c *declarationCollector) collectMethodsDefinition(def *ast.MethodsDefinition) (*MethodSetSpec, []Diagnostic) {
	if def == nil {
		return nil, nil
	}

	c.ensureMethodsGenericInference(def)

	params, paramScope := c.convertGenericParams(def.GenericParams)
	scope := copyTypeScope(paramScope)

	targetType := c.resolveTypeExpression(def.TargetType, scope)
	if targetType == nil {
		targetType = UnknownType{}
	}
	scope["Self"] = targetType

	where := c.convertWhereClause(def.WhereClause, scope)
	methodsLabel := fmt.Sprintf("methods for %s", nonEmpty(typeName(targetType)))
	obligations := obligationsFromSpecs(methodsLabel, params, where, def)
	functionObligations := obligations
	if targetType != nil && !isUnknownType(targetType) {
		functionObligations = substituteObligations(obligations, map[string]Type{"Self": targetType})
	}
	if len(functionObligations) > 0 {
		if len(functionObligations) == len(obligations) && &functionObligations[0] == &obligations[0] {
			functionObligations = append([]ConstraintObligation{}, obligations...)
		}
		for idx := range functionObligations {
			if functionObligations[idx].Context == "" {
				functionObligations[idx].Context = "via method set"
			}
		}
	}

	var diags []Diagnostic
	methods := make(map[string]FunctionType, len(def.Definitions))
	typeQualified := make(map[string]bool)
	baseName := typeName(targetType)
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			diags = append(diags, Diagnostic{
				Message: "typechecker: method definition requires a name",
				Node:    fn,
			})
			continue
		}
		name := fn.ID.Name
		if existing, exists := methods[name]; exists {
			// Flag overloaded method sets as unknown to avoid mismatched arity diagnostics.
			if !isUnknownFunctionSignature(existing) {
				methods[name] = FunctionType{Return: UnknownType{}}
			}
			continue
		}
		methodOwner := fmt.Sprintf("%s::%s", methodsLabel, functionName(fn))
		fnType := c.functionTypeFromDefinition(fn, scope, methodOwner, fn)
		fnType = applyImplicitSelfParam(fn, fnType, targetType)
		if len(functionObligations) > 0 {
			fnType.Obligations = append(fnType.Obligations, functionObligations...)
		}
		isSelfMethod := fn.IsMethodShorthand
		if len(fn.Params) > 0 {
			first := fn.Params[0]
			if first != nil {
				if first.Name != nil {
					if ident, ok := first.Name.(*ast.Identifier); ok && ident != nil && strings.EqualFold(ident.Name, "self") {
						isSelfMethod = true
					}
				}
				if simple, ok := first.ParamType.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == "Self" {
					isSelfMethod = true
				}
			}
		}
		if !isSelfMethod {
			typeQualified[name] = true
		}
		methods[name] = fnType
	}

	exported := make(map[string]struct{})
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			continue
		}
		name := fn.ID.Name
		fnType, ok := methods[name]
		if !ok {
			continue
		}
		exportName := name
		if typeQualified[name] && baseName != "" {
			exportName = fmt.Sprintf("%s.%s", baseName, name)
		}
		if _, exists := c.env.Lookup(exportName); !exists {
			c.env.Define(exportName, fnType)
		}
		if !fn.IsPrivate {
			if _, exists := exported[exportName]; !exists {
				c.exports = append(c.exports, exportRecord{name: exportName, node: fn})
				exported[exportName] = struct{}{}
			}
		}
	}

	spec := &MethodSetSpec{
		TypeParams:    params,
		Target:        targetType,
		Methods:       methods,
		TypeQualified: typeQualified,
		Where:         where,
		Definition:    def,
	}
	spec.Obligations = obligations
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
	if def == nil || len(fnType.Params) == 0 {
		return fnType
	}
	if def.IsMethodShorthand {
		if target == nil || isUnknownType(target) {
			fnType.Params[0] = TypeParameterType{ParameterName: "Self"}
		} else {
			fnType.Params[0] = target
		}
		return fnType
	}
	if len(def.Params) == 0 {
		return fnType
	}
	firstParam := def.Params[0]
	if firstParam == nil {
		return fnType
	}
	if firstParam.ParamType != nil && !isUnknownType(fnType.Params[0]) {
		if simple, ok := firstParam.ParamType.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == "Self" {
			if target == nil || isUnknownType(target) {
				fnType.Params[0] = TypeParameterType{ParameterName: "Self"}
			} else {
				fnType.Params[0] = target
			}
		}
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
