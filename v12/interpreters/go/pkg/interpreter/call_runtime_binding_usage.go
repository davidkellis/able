package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type functionRuntimeGenericBindingPlan struct {
	explicitUsed           bool
	callLocalUsed          bool
	hasGenericConstraints  bool
	returnTypeUsesGenerics bool
	paramUsesGenerics      []bool
}

func addRuntimeBindingTypeName(names map[string]struct{}, name string) map[string]struct{} {
	if name == "" {
		return names
	}
	if names == nil {
		names = make(map[string]struct{}, 4)
	}
	names[name] = struct{}{}
	return names
}

func addRuntimeBindingTypeNamesForGenerics(names map[string]struct{}, generics []*ast.GenericParameter) map[string]struct{} {
	for _, gp := range generics {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			continue
		}
		names = addRuntimeBindingTypeName(names, gp.Name.Name)
	}
	return names
}

func runtimeBindingIdentifierNames(typeNames map[string]struct{}) map[string]struct{} {
	if len(typeNames) == 0 {
		return nil
	}
	names := make(map[string]struct{}, len(typeNames)*2)
	for name := range typeNames {
		if name == "" {
			continue
		}
		names[name] = struct{}{}
		names[name+"_type"] = struct{}{}
	}
	return names
}

func callableNeedsExplicitRuntimeTypeBindings(node ast.Node) bool {
	explicitTypeNames := callableExplicitRuntimeBindingTypeNames(node)
	if len(explicitTypeNames) == 0 {
		return false
	}
	return callableUsesRuntimeBindingNames(node, runtimeBindingIdentifierNames(explicitTypeNames), explicitTypeNames)
}

func (i *Interpreter) callableNeedsExplicitRuntimeTypeBindings(node ast.Node) bool {
	if i == nil || node == nil {
		return false
	}
	if i.envSingleThread {
		if used, ok := i.callableExplicitRuntimeBindingUsageCache[node]; ok {
			return used
		}
		used := callableNeedsExplicitRuntimeTypeBindings(node)
		i.callableExplicitRuntimeBindingUsageCache[node] = used
		return used
	}
	i.callableExplicitRuntimeBindingUsageCacheMu.RLock()
	used, ok := i.callableExplicitRuntimeBindingUsageCache[node]
	i.callableExplicitRuntimeBindingUsageCacheMu.RUnlock()
	if ok {
		return used
	}
	used = callableNeedsExplicitRuntimeTypeBindings(node)
	i.callableExplicitRuntimeBindingUsageCacheMu.Lock()
	if existing, ok := i.callableExplicitRuntimeBindingUsageCache[node]; ok {
		i.callableExplicitRuntimeBindingUsageCacheMu.Unlock()
		return existing
	}
	i.callableExplicitRuntimeBindingUsageCache[node] = used
	i.callableExplicitRuntimeBindingUsageCacheMu.Unlock()
	return used
}

func callableExplicitRuntimeBindingTypeNames(node ast.Node) map[string]struct{} {
	switch decl := node.(type) {
	case *ast.FunctionDefinition:
		return addRuntimeBindingTypeNamesForGenerics(nil, decl.GenericParams)
	case *ast.LambdaExpression:
		return addRuntimeBindingTypeNamesForGenerics(nil, decl.GenericParams)
	default:
		return nil
	}
}

func callableUsesRuntimeBindingNames(node ast.Node, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	if len(identifierNames) == 0 && len(typeNames) == 0 {
		return false
	}
	switch decl := node.(type) {
	case *ast.FunctionDefinition:
		return runtimeBlockUsesBindingNames(decl.Body, identifierNames, typeNames)
	case *ast.LambdaExpression:
		return runtimeExprUsesBindingNames(decl.Body, identifierNames, typeNames)
	default:
		return false
	}
}

func runtimeBlockUsesBindingNames(block *ast.BlockExpression, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.Body {
		if runtimeStmtUsesBindingNames(stmt, identifierNames, typeNames) {
			return true
		}
	}
	return false
}

func runtimeStmtUsesBindingNames(stmt ast.Statement, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	if stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *ast.FunctionDefinition:
		return runtimeFunctionDefinitionUsesBindingNames(s, identifierNames, typeNames)
	case *ast.MethodsDefinition:
		return runtimeMethodsDefinitionUsesBindingNames(s, identifierNames, typeNames)
	case *ast.ImplementationDefinition:
		return runtimeImplementationDefinitionUsesBindingNames(s, identifierNames, typeNames)
	case *ast.StructDefinition:
		return runtimeStructDefinitionUsesBindingNames(s, typeNames)
	case *ast.UnionDefinition:
		return runtimeUnionDefinitionUsesBindingNames(s, typeNames)
	case *ast.TypeAliasDefinition:
		return runtimeTypeExpressionUsesBindingNames(s.TargetType, typeNames) ||
			runtimeWhereClauseUsesBindingNames(s.WhereClause, typeNames)
	case *ast.InterfaceDefinition:
		return runtimeInterfaceDefinitionUsesBindingNames(s, identifierNames, typeNames)
	case *ast.ExternFunctionBody:
		return runtimeFunctionDefinitionUsesBindingNames(s.Signature, identifierNames, typeNames)
	case *ast.ReturnStatement:
		return runtimeExprUsesBindingNames(s.Argument, identifierNames, typeNames)
	case *ast.YieldStatement:
		return runtimeExprUsesBindingNames(s.Expression, identifierNames, typeNames)
	case *ast.RaiseStatement:
		return runtimeExprUsesBindingNames(s.Expression, identifierNames, typeNames)
	case *ast.BreakStatement:
		return runtimeExprUsesBindingNames(s.Value, identifierNames, typeNames)
	case *ast.ForLoop:
		return runtimePatternUsesBindingNames(s.Pattern, identifierNames, typeNames) ||
			runtimeExprUsesBindingNames(s.Iterable, identifierNames, typeNames) ||
			runtimeBlockUsesBindingNames(s.Body, identifierNames, typeNames)
	case *ast.WhileLoop:
		return runtimeExprUsesBindingNames(s.Condition, identifierNames, typeNames) ||
			runtimeBlockUsesBindingNames(s.Body, identifierNames, typeNames)
	case *ast.PackageStatement, *ast.ImportStatement, *ast.DynImportStatement,
		*ast.PreludeStatement, *ast.ContinueStatement, *ast.RethrowStatement:
		return false
	case ast.Expression:
		return runtimeExprUsesBindingNames(s, identifierNames, typeNames)
	default:
		return true
	}
}

func runtimeFunctionDefinitionUsesBindingNames(def *ast.FunctionDefinition, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	if def == nil {
		return false
	}
	for _, param := range def.Params {
		if param == nil {
			continue
		}
		if runtimePatternUsesBindingNames(param.Name, identifierNames, typeNames) ||
			runtimeTypeExpressionUsesBindingNames(param.ParamType, typeNames) {
			return true
		}
	}
	return runtimeTypeExpressionUsesBindingNames(def.ReturnType, typeNames) ||
		runtimeWhereClauseUsesBindingNames(def.WhereClause, typeNames) ||
		runtimeBlockUsesBindingNames(def.Body, identifierNames, typeNames)
}

func runtimeLambdaUsesBindingNames(expr *ast.LambdaExpression, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	if expr == nil {
		return false
	}
	for _, param := range expr.Params {
		if param == nil {
			continue
		}
		if runtimePatternUsesBindingNames(param.Name, identifierNames, typeNames) ||
			runtimeTypeExpressionUsesBindingNames(param.ParamType, typeNames) {
			return true
		}
	}
	return runtimeTypeExpressionUsesBindingNames(expr.ReturnType, typeNames) ||
		runtimeWhereClauseUsesBindingNames(expr.WhereClause, typeNames) ||
		runtimeExprUsesBindingNames(expr.Body, identifierNames, typeNames)
}

func runtimeMethodsDefinitionUsesBindingNames(def *ast.MethodsDefinition, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	if def == nil {
		return false
	}
	if runtimeTypeExpressionUsesBindingNames(def.TargetType, typeNames) ||
		runtimeWhereClauseUsesBindingNames(def.WhereClause, typeNames) {
		return true
	}
	for _, method := range def.Definitions {
		if runtimeFunctionDefinitionUsesBindingNames(method, identifierNames, typeNames) {
			return true
		}
	}
	return false
}

func runtimeImplementationDefinitionUsesBindingNames(def *ast.ImplementationDefinition, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	if def == nil {
		return false
	}
	if runtimeTypeExpressionUsesBindingNames(def.TargetType, typeNames) ||
		runtimeTypeExpressionsUseBindingNames(def.InterfaceArgs, typeNames) ||
		runtimeWhereClauseUsesBindingNames(def.WhereClause, typeNames) {
		return true
	}
	for _, method := range def.Definitions {
		if runtimeFunctionDefinitionUsesBindingNames(method, identifierNames, typeNames) {
			return true
		}
	}
	return false
}

func runtimeStructDefinitionUsesBindingNames(def *ast.StructDefinition, typeNames map[string]struct{}) bool {
	if def == nil {
		return false
	}
	for _, field := range def.Fields {
		if field != nil && runtimeTypeExpressionUsesBindingNames(field.FieldType, typeNames) {
			return true
		}
	}
	return runtimeWhereClauseUsesBindingNames(def.WhereClause, typeNames)
}

func runtimeUnionDefinitionUsesBindingNames(def *ast.UnionDefinition, typeNames map[string]struct{}) bool {
	if def == nil {
		return false
	}
	return runtimeTypeExpressionsUseBindingNames(def.Variants, typeNames) ||
		runtimeWhereClauseUsesBindingNames(def.WhereClause, typeNames)
}

func runtimeInterfaceDefinitionUsesBindingNames(def *ast.InterfaceDefinition, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	if def == nil {
		return false
	}
	if runtimeTypeExpressionUsesBindingNames(def.SelfTypePattern, typeNames) ||
		runtimeTypeExpressionsUseBindingNames(def.BaseInterfaces, typeNames) ||
		runtimeWhereClauseUsesBindingNames(def.WhereClause, typeNames) {
		return true
	}
	for _, sig := range def.Signatures {
		if sig == nil {
			continue
		}
		for _, param := range sig.Params {
			if param == nil {
				continue
			}
			if runtimePatternUsesBindingNames(param.Name, identifierNames, typeNames) ||
				runtimeTypeExpressionUsesBindingNames(param.ParamType, typeNames) {
				return true
			}
		}
		if runtimeTypeExpressionUsesBindingNames(sig.ReturnType, typeNames) ||
			runtimeWhereClauseUsesBindingNames(sig.WhereClause, typeNames) ||
			runtimeBlockUsesBindingNames(sig.DefaultImpl, identifierNames, typeNames) {
			return true
		}
	}
	return false
}

func runtimeWhereClauseUsesBindingNames(where []*ast.WhereClauseConstraint, typeNames map[string]struct{}) bool {
	for _, clause := range where {
		if clause == nil {
			continue
		}
		if runtimeTypeExpressionUsesBindingNames(clause.TypeParam, typeNames) {
			return true
		}
		for _, constraint := range clause.Constraints {
			if constraint != nil && runtimeTypeExpressionUsesBindingNames(constraint.InterfaceType, typeNames) {
				return true
			}
		}
	}
	return false
}

func runtimeExprUsesBindingNames(expr ast.Expression, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	if expr == nil {
		return false
	}
	switch n := expr.(type) {
	case *ast.Identifier:
		if n == nil {
			return false
		}
		_, ok := identifierNames[n.Name]
		return ok
	case *ast.StringLiteral, *ast.BooleanLiteral, *ast.CharLiteral,
		*ast.NilLiteral, *ast.IntegerLiteral, *ast.FloatLiteral,
		*ast.PlaceholderExpression, *ast.ImplicitMemberExpression:
		return false
	case *ast.UnaryExpression:
		return runtimeExprUsesBindingNames(n.Operand, identifierNames, typeNames)
	case *ast.BinaryExpression:
		return runtimeExprUsesBindingNames(n.Left, identifierNames, typeNames) ||
			runtimeExprUsesBindingNames(n.Right, identifierNames, typeNames)
	case *ast.AssignmentExpression:
		return runtimeAssignmentTargetUsesBindingNames(n.Left, identifierNames, typeNames) ||
			runtimeExprUsesBindingNames(n.Right, identifierNames, typeNames)
	case *ast.FunctionCall:
		return runtimeExprUsesBindingNames(n.Callee, identifierNames, typeNames) ||
			runtimeExpressionsUseBindingNames(n.Arguments, identifierNames, typeNames) ||
			runtimeTypeExpressionsUseBindingNames(n.TypeArguments, typeNames)
	case *ast.MemberAccessExpression:
		return runtimeExprUsesBindingNames(n.Object, identifierNames, typeNames)
	case *ast.IndexExpression:
		return runtimeExprUsesBindingNames(n.Object, identifierNames, typeNames) ||
			runtimeExprUsesBindingNames(n.Index, identifierNames, typeNames)
	case *ast.BlockExpression:
		return runtimeBlockUsesBindingNames(n, identifierNames, typeNames)
	case *ast.IfExpression:
		if runtimeExprUsesBindingNames(n.IfCondition, identifierNames, typeNames) ||
			runtimeBlockUsesBindingNames(n.IfBody, identifierNames, typeNames) ||
			runtimeBlockUsesBindingNames(n.ElseBody, identifierNames, typeNames) {
			return true
		}
		for _, clause := range n.ElseIfClauses {
			if clause == nil {
				continue
			}
			if runtimeExprUsesBindingNames(clause.Condition, identifierNames, typeNames) ||
				runtimeBlockUsesBindingNames(clause.Body, identifierNames, typeNames) {
				return true
			}
		}
		return false
	case *ast.MatchExpression:
		if runtimeExprUsesBindingNames(n.Subject, identifierNames, typeNames) {
			return true
		}
		for _, clause := range n.Clauses {
			if clause == nil {
				continue
			}
			if runtimePatternUsesBindingNames(clause.Pattern, identifierNames, typeNames) ||
				runtimeExprUsesBindingNames(clause.Guard, identifierNames, typeNames) ||
				runtimeExprUsesBindingNames(clause.Body, identifierNames, typeNames) {
				return true
			}
		}
		return false
	case *ast.ArrayLiteral:
		return runtimeExpressionsUseBindingNames(n.Elements, identifierNames, typeNames)
	case *ast.MapLiteral:
		for _, el := range n.Elements {
			switch item := el.(type) {
			case *ast.MapLiteralEntry:
				if runtimeExprUsesBindingNames(item.Key, identifierNames, typeNames) ||
					runtimeExprUsesBindingNames(item.Value, identifierNames, typeNames) {
					return true
				}
			case *ast.MapLiteralSpread:
				if runtimeExprUsesBindingNames(item.Expression, identifierNames, typeNames) {
					return true
				}
			default:
				return true
			}
		}
		return false
	case *ast.StructLiteral:
		if runtimeTypeExpressionsUseBindingNames(n.TypeArguments, typeNames) ||
			runtimeExpressionsUseBindingNames(n.FunctionalUpdateSources, identifierNames, typeNames) {
			return true
		}
		for _, field := range n.Fields {
			if field != nil && runtimeExprUsesBindingNames(field.Value, identifierNames, typeNames) {
				return true
			}
		}
		return false
	case *ast.StringInterpolation:
		return runtimeExpressionsUseBindingNames(n.Parts, identifierNames, typeNames)
	case *ast.TypeCastExpression:
		return runtimeExprUsesBindingNames(n.Expression, identifierNames, typeNames) ||
			runtimeTypeExpressionUsesBindingNames(n.TargetType, typeNames)
	case *ast.RangeExpression:
		return runtimeExprUsesBindingNames(n.Start, identifierNames, typeNames) ||
			runtimeExprUsesBindingNames(n.End, identifierNames, typeNames)
	case *ast.PropagationExpression:
		return runtimeExprUsesBindingNames(n.Expression, identifierNames, typeNames)
	case *ast.AwaitExpression:
		return runtimeExprUsesBindingNames(n.Expression, identifierNames, typeNames)
	case *ast.LoopExpression:
		return runtimeBlockUsesBindingNames(n.Body, identifierNames, typeNames)
	case *ast.LambdaExpression:
		return runtimeLambdaUsesBindingNames(n, identifierNames, typeNames)
	case *ast.SpawnExpression:
		return runtimeExprUsesBindingNames(n.Expression, identifierNames, typeNames)
	case *ast.IteratorLiteral:
		return runtimeTypeExpressionUsesBindingNames(n.ElementType, typeNames) ||
			runtimeBlockStatementsUseBindingNames(n.Body, identifierNames, typeNames)
	case *ast.RescueExpression:
		if runtimeExprUsesBindingNames(n.MonitoredExpression, identifierNames, typeNames) {
			return true
		}
		for _, clause := range n.Clauses {
			if clause == nil {
				continue
			}
			if runtimePatternUsesBindingNames(clause.Pattern, identifierNames, typeNames) ||
				runtimeExprUsesBindingNames(clause.Guard, identifierNames, typeNames) ||
				runtimeExprUsesBindingNames(clause.Body, identifierNames, typeNames) {
				return true
			}
		}
		return false
	case *ast.EnsureExpression:
		return runtimeExprUsesBindingNames(n.TryExpression, identifierNames, typeNames) ||
			runtimeBlockUsesBindingNames(n.EnsureBlock, identifierNames, typeNames)
	case *ast.BreakpointExpression:
		return runtimeBlockUsesBindingNames(n.Body, identifierNames, typeNames)
	case *ast.OrElseExpression:
		return runtimeExprUsesBindingNames(n.Expression, identifierNames, typeNames) ||
			runtimeBlockUsesBindingNames(n.Handler, identifierNames, typeNames)
	default:
		return true
	}
}

func runtimeAssignmentTargetUsesBindingNames(target ast.AssignmentTarget, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	if target == nil {
		return false
	}
	switch n := target.(type) {
	case ast.Expression:
		return runtimeExprUsesBindingNames(n, identifierNames, typeNames)
	case ast.Pattern:
		return runtimePatternUsesBindingNames(n, identifierNames, typeNames)
	default:
		return true
	}
}

func runtimePatternUsesBindingNames(pattern ast.Pattern, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	if pattern == nil {
		return false
	}
	switch p := pattern.(type) {
	case *ast.Identifier, *ast.WildcardPattern, *ast.LiteralPattern:
		return false
	case *ast.StructPattern:
		if p.StructType != nil {
			if _, ok := typeNames[p.StructType.Name]; ok {
				return true
			}
		}
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if runtimePatternUsesBindingNames(field.Pattern, identifierNames, typeNames) ||
				runtimeTypeExpressionUsesBindingNames(field.TypeAnnotation, typeNames) {
				return true
			}
		}
		return false
	case *ast.ArrayPattern:
		for _, el := range p.Elements {
			if runtimePatternUsesBindingNames(el, identifierNames, typeNames) {
				return true
			}
		}
		return runtimePatternUsesBindingNames(p.RestPattern, identifierNames, typeNames)
	case *ast.TypedPattern:
		return runtimePatternUsesBindingNames(p.Pattern, identifierNames, typeNames) ||
			runtimeTypeExpressionUsesBindingNames(p.TypeAnnotation, typeNames)
	default:
		return true
	}
}

func runtimeExpressionsUseBindingNames(values []ast.Expression, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	for _, value := range values {
		if runtimeExprUsesBindingNames(value, identifierNames, typeNames) {
			return true
		}
	}
	return false
}

func runtimeBlockStatementsUseBindingNames(values []ast.Statement, identifierNames map[string]struct{}, typeNames map[string]struct{}) bool {
	for _, value := range values {
		if runtimeStmtUsesBindingNames(value, identifierNames, typeNames) {
			return true
		}
	}
	return false
}

func runtimeTypeExpressionsUseBindingNames(values []ast.TypeExpression, typeNames map[string]struct{}) bool {
	for _, value := range values {
		if runtimeTypeExpressionUsesBindingNames(value, typeNames) {
			return true
		}
	}
	return false
}

func runtimeTypeExpressionUsesBindingNames(expr ast.TypeExpression, typeNames map[string]struct{}) bool {
	if expr == nil || len(typeNames) == 0 {
		return false
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return false
		}
		_, ok := typeNames[t.Name.Name]
		return ok
	case *ast.GenericTypeExpression:
		return runtimeTypeExpressionUsesBindingNames(t.Base, typeNames) ||
			runtimeTypeExpressionsUseBindingNames(t.Arguments, typeNames)
	case *ast.NullableTypeExpression:
		return runtimeTypeExpressionUsesBindingNames(t.InnerType, typeNames)
	case *ast.ResultTypeExpression:
		return runtimeTypeExpressionUsesBindingNames(t.InnerType, typeNames)
	case *ast.UnionTypeExpression:
		return runtimeTypeExpressionsUseBindingNames(t.Members, typeNames)
	case *ast.FunctionTypeExpression:
		return runtimeTypeExpressionsUseBindingNames(t.ParamTypes, typeNames) ||
			runtimeTypeExpressionUsesBindingNames(t.ReturnType, typeNames)
	default:
		return true
	}
}

func (i *Interpreter) functionRuntimeGenericBindingPlan(fn *runtime.FunctionValue) *functionRuntimeGenericBindingPlan {
	if i == nil || fn == nil {
		return nil
	}
	if i.envSingleThread {
		if plan, ok := i.functionRuntimeGenericBindingPlanCache[fn]; ok {
			return plan
		}
		plan := i.buildFunctionRuntimeGenericBindingPlan(fn)
		i.functionRuntimeGenericBindingPlanCache[fn] = plan
		return plan
	}
	i.functionRuntimeGenericBindingPlanCacheMu.RLock()
	plan, ok := i.functionRuntimeGenericBindingPlanCache[fn]
	i.functionRuntimeGenericBindingPlanCacheMu.RUnlock()
	if ok {
		return plan
	}
	plan = i.buildFunctionRuntimeGenericBindingPlan(fn)
	i.functionRuntimeGenericBindingPlanCacheMu.Lock()
	if existing, ok := i.functionRuntimeGenericBindingPlanCache[fn]; ok {
		i.functionRuntimeGenericBindingPlanCacheMu.Unlock()
		return existing
	}
	i.functionRuntimeGenericBindingPlanCache[fn] = plan
	i.functionRuntimeGenericBindingPlanCacheMu.Unlock()
	return plan
}

func buildCallableParamGenericUsage(params []*ast.FunctionParameter, genericNames map[string]struct{}) []bool {
	if len(params) == 0 || len(genericNames) == 0 {
		return nil
	}
	uses := make([]bool, len(params))
	any := false
	for idx, param := range params {
		if param == nil || param.ParamType == nil {
			continue
		}
		if !typeExprUsesGeneric(param.ParamType, genericNames) {
			continue
		}
		uses[idx] = true
		any = true
	}
	if !any {
		return nil
	}
	return uses
}

func (plan *functionRuntimeGenericBindingPlan) paramUsesGeneric(idx int) bool {
	if plan == nil || idx < 0 || idx >= len(plan.paramUsesGenerics) {
		return false
	}
	return plan.paramUsesGenerics[idx]
}

func (i *Interpreter) functionParamUsesGenerics(fn *runtime.FunctionValue, idx int, fallback ast.TypeExpression) bool {
	plan := i.functionRuntimeGenericBindingPlan(fn)
	if plan != nil {
		return plan.paramUsesGeneric(idx)
	}
	if fn == nil {
		return false
	}
	return typeExprUsesGeneric(fallback, fn.GenericNameSet(nil))
}

func (i *Interpreter) functionReturnTypeUsesGenerics(fn *runtime.FunctionValue, fallback ast.TypeExpression) bool {
	plan := i.functionRuntimeGenericBindingPlan(fn)
	if plan != nil {
		return plan.returnTypeUsesGenerics
	}
	if fn == nil {
		return false
	}
	return typeExpressionUsesGenerics(fallback, fn.GenericNameSet(nil))
}

func (i *Interpreter) buildFunctionRuntimeGenericBindingPlan(fn *runtime.FunctionValue) *functionRuntimeGenericBindingPlan {
	plan := &functionRuntimeGenericBindingPlan{}
	if fn == nil {
		return plan
	}
	plan.explicitUsed = i.callableNeedsExplicitRuntimeTypeBindings(fn.Declaration)
	genericNames := fn.GenericNameSet(nil)
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		plan.hasGenericConstraints = hasAnyGenericConstraints(decl.GenericParams, decl.WhereClause)
		plan.returnTypeUsesGenerics = typeExpressionUsesGenerics(decl.ReturnType, genericNames)
		plan.paramUsesGenerics = buildCallableParamGenericUsage(decl.Params, genericNames)
	case *ast.LambdaExpression:
		plan.hasGenericConstraints = hasAnyGenericConstraints(decl.GenericParams, decl.WhereClause)
		plan.returnTypeUsesGenerics = typeExpressionUsesGenerics(decl.ReturnType, genericNames)
		plan.paramUsesGenerics = buildCallableParamGenericUsage(decl.Params, genericNames)
	}
	callLocalTypeNames := i.callLocalRuntimeBindingTypeNames(fn)
	if len(callLocalTypeNames) == 0 {
		return plan
	}
	plan.callLocalUsed = callableUsesRuntimeBindingNames(
		fn.Declaration,
		runtimeBindingIdentifierNames(callLocalTypeNames),
		callLocalTypeNames,
	)
	return plan
}

func (i *Interpreter) callLocalRuntimeBindingTypeNames(fn *runtime.FunctionValue) map[string]struct{} {
	if i == nil || fn == nil {
		return nil
	}
	var typeNames map[string]struct{}
	if fn.MethodSet != nil {
		typeNames = addRuntimeBindingTypeNamesForGenerics(typeNames, fn.MethodSet.GenericParams)
		typeNames = addRuntimeBindingTypeName(typeNames, "Self")
		typeNames = addRuntimeBindingTypeName(typeNames, "SelfType")
	}
	if ctx := i.implMethodContextFromEnv(fn.Closure); ctx != nil {
		typeNames = addRuntimeBindingTypeName(typeNames, "Self")
		typeNames = addRuntimeBindingTypeName(typeNames, "SelfType")
		iface := i.interfaces[ctx.interfaceName]
		if iface != nil && iface.Node != nil && iface.Node.SelfTypePattern != nil {
			pattern := i.expandTypeAliasesCached(iface.Node.SelfTypePattern)
			if pattern == nil {
				pattern = iface.Node.SelfTypePattern
			}
			for name := range i.typePatternBindingNames(pattern, iface.Env, nil) {
				typeNames = addRuntimeBindingTypeName(typeNames, name)
			}
		}
	}
	return typeNames
}
