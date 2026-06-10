package compiler

import "able/interpreter-go/pkg/ast"

// staticGenericUnionMemberCall records the checked receiver type and declared
// union target used by generic named-union dispatch. Generic named unions use
// their structural runtime representation, so a value alone cannot distinguish
// two methods with the same name on different unions. The generated call keeps
// this fact without changing value layout.
func (g *generator) staticGenericUnionMemberCall(ctx *compileContext, call *ast.FunctionCall, callee *ast.MemberAccessExpression, memberName string) bool {
	if g == nil || ctx == nil || call == nil || callee == nil || callee.Object == nil || memberName == "" {
		return false
	}
	receiverType := g.dispatchReceiverTypeExpr(ctx, callee.Object)
	if receiverType == nil {
		return false
	}
	targetName, ok := g.genericNamedUnionMethodTargetName(ctx, call, callee, memberName)
	if !ok {
		return false
	}
	if g.staticCallReceiverTypeHints == nil {
		g.staticCallReceiverTypeHints = make(map[*ast.FunctionCall]ast.TypeExpression)
	}
	g.staticCallReceiverTypeHints[call] = receiverType
	if g.staticGenericUnionMethodTargets == nil {
		g.staticGenericUnionMethodTargets = make(map[*ast.FunctionCall]string)
	}
	g.staticGenericUnionMethodTargets[call] = targetName
	return true
}

// genericNamedUnionMethodTargetName resolves the declared generic-union
// target for a checked call. Chained results use their structural union form
// during lowering, so this cannot rely on the runtime variant's type name.
func (g *generator) genericNamedUnionMethodTargetName(ctx *compileContext, call *ast.FunctionCall, callee *ast.MemberAccessExpression, memberName string) (string, bool) {
	if g == nil || ctx == nil || call == nil || callee == nil || callee.Object == nil || memberName == "" {
		return "", false
	}
	receiverType := g.dispatchReceiverTypeExpr(ctx, callee.Object)
	if receiverType == nil {
		return "", false
	}
	for _, method := range g.methodList {
		if method == nil || !method.ExpectsSelf || method.MethodName != memberName || method.Info == nil || method.Info.Definition == nil {
			continue
		}
		target, ok := g.genericNamedUnionMethodTargetTypeExpr(method)
		if !ok || target == nil {
			continue
		}
		genericNames := g.methodGenericNames(method)
		bindings := make(map[string]ast.TypeExpression)
		if !g.genericNamedUnionStructuralReceiverMatches(method.Info.Package, target, receiverType, genericNames, bindings) {
			continue
		}
		targetName, ok := typeExprBaseName(method.TargetType)
		if !ok || targetName == "" {
			continue
		}
		return targetName, true
	}
	return "", false
}

// genericNamedUnionMethodResultTypeExpr recovers a concrete result type for a
// call on a generic named union. The runtime representation of those unions is
// structural, but compiler diagnostics need the same structural receiver fact
// for a subsequent chained call. This applies to any generic union method; it
// does not name or otherwise privilege a particular stdlib union.
func (g *generator) genericNamedUnionMethodResultTypeExpr(ctx *compileContext, call *ast.FunctionCall, callee *ast.MemberAccessExpression) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || callee == nil || callee.Object == nil {
		return nil, false
	}
	member, ok := callee.Member.(*ast.Identifier)
	if !ok || member == nil || member.Name == "" {
		return nil, false
	}
	receiverType := g.dispatchReceiverTypeExpr(ctx, callee.Object)
	if receiverType == nil {
		return nil, false
	}
	for _, method := range g.methodList {
		if method == nil || !method.ExpectsSelf || method.MethodName != member.Name || method.Info == nil || method.Info.Definition == nil {
			continue
		}
		target, ok := g.genericNamedUnionMethodTargetTypeExpr(method)
		if !ok || target == nil {
			continue
		}
		genericNames := g.methodGenericNames(method)
		bindings := make(map[string]ast.TypeExpression)
		if !g.genericNamedUnionStructuralReceiverMatches(method.Info.Package, target, receiverType, genericNames, bindings) {
			continue
		}
		if !g.applyGenericNamedUnionMethodCallBindings(ctx, call, method, genericNames, bindings) {
			continue
		}
		result := g.functionReturnTypeExprWithBindings(method.Info, bindings)
		if result == nil || g.typeExprHasGeneric(result, genericNames) {
			continue
		}
		return g.genericNamedUnionStructuralTypeExpr(method.Info.Package, result), true
	}
	return nil, false
}

// resolveGenericNamedUnionInstanceMethod keeps a checked call on a generic
// named union on the ordinary specialized-method path. These unions use a
// structural Go carrier, so methodForReceiver cannot recover their nominal
// method from the carrier name alone.
func (g *generator) resolveGenericNamedUnionInstanceMethod(ctx *compileContext, call *ast.FunctionCall, receiver ast.Expression, memberName string, expected string) (*methodInfo, bool) {
	if g == nil || ctx == nil || call == nil || receiver == nil || memberName == "" {
		return nil, false
	}
	receiverType := g.dispatchReceiverTypeExpr(ctx, receiver)
	if receiverType == nil {
		return nil, false
	}
	var resolved *methodInfo
	for _, method := range g.methodList {
		if method == nil || !method.ExpectsSelf || method.MethodName != memberName || method.Info == nil || method.Info.Definition == nil {
			continue
		}
		target, ok := g.genericNamedUnionMethodTargetTypeExpr(method)
		if !ok || target == nil {
			continue
		}
		genericNames := g.nominalMethodSpecializationGenericNames(method)
		bindings := g.concreteCompileContextBindings(method.Info, genericNames)
		bindings = g.mergeConcreteTypeBindings(method.Info.Package, genericNames, bindings, ctx.typeBindings)
		if bindings == nil {
			bindings = make(map[string]ast.TypeExpression)
		}
		if !g.genericNamedUnionStructuralReceiverMatches(method.Info.Package, target, receiverType, genericNames, bindings) {
			continue
		}
		bindings, ok = g.finishSpecializedNominalMethodBindings(ctx, call, method, genericNames, bindings, expected)
		if !ok || len(bindings) == 0 {
			continue
		}
		specialized, ok := g.ensureSpecializedNominalMethod(method, bindings)
		if !ok || specialized == nil || specialized.Info == nil || !specialized.Info.Compileable {
			continue
		}
		specialized.ConcreteResolved = true
		if resolved != nil && resolved.Info != specialized.Info {
			return nil, false
		}
		resolved = specialized
	}
	return resolved, resolved != nil
}

// genericNamedUnionStructuralReceiverMatches requires the complete structural
// union shape of a generic named-union receiver. The general type-template
// matcher intentionally lets a union template match one of its members; that
// is correct for pattern-style matching but would make a method on `nil | T`
// appear applicable to every T-shaped receiver (for example Iterator<String>).
// Named-union dispatch needs the whole receiver identity instead.
func (g *generator) genericNamedUnionStructuralReceiverMatches(pkgName string, target ast.TypeExpression, receiver ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || target == nil || receiver == nil {
		return false
	}
	target = g.genericNamedUnionStructuralTypeExpr(pkgName, target)
	receiver = g.genericNamedUnionStructuralTypeExpr(pkgName, receiver)
	targetUnion, targetOK := target.(*ast.UnionTypeExpression)
	receiverUnion, receiverOK := receiver.(*ast.UnionTypeExpression)
	if !targetOK || targetUnion == nil || !receiverOK || receiverUnion == nil || len(targetUnion.Members) != len(receiverUnion.Members) {
		return false
	}
	for index, targetMember := range targetUnion.Members {
		if !g.specializedTypeTemplateMatches(pkgName, targetMember, receiverUnion.Members[index], genericNames, bindings, make(map[string]struct{})) {
			return false
		}
	}
	return true
}

func (g *generator) genericNamedUnionMethodTargetTypeExpr(method *methodInfo) (ast.TypeExpression, bool) {
	if g == nil || method == nil || method.Info == nil || method.TargetType == nil {
		return nil, false
	}
	return g.genericNamedUnionStructuralTypeExprForGenericUnion(method.Info.Package, method.TargetType)
}

func (g *generator) genericNamedUnionStructuralTypeExpr(pkgName string, expr ast.TypeExpression) ast.TypeExpression {
	if g == nil || expr == nil {
		return nil
	}
	switch value := expr.(type) {
	case *ast.NullableTypeExpression:
		if value == nil {
			return expr
		}
		return ast.NewUnionTypeExpression([]ast.TypeExpression{ast.Ty("nil"), g.genericNamedUnionStructuralTypeExpr(pkgName, value.InnerType)})
	case *ast.ResultTypeExpression:
		if value == nil {
			return expr
		}
		return ast.NewUnionTypeExpression([]ast.TypeExpression{ast.Ty("Error"), g.genericNamedUnionStructuralTypeExpr(pkgName, value.InnerType)})
	case *ast.UnionTypeExpression:
		if value == nil {
			return expr
		}
		members := make([]ast.TypeExpression, len(value.Members))
		for index, member := range value.Members {
			members[index] = g.genericNamedUnionStructuralTypeExpr(pkgName, member)
		}
		return ast.NewUnionTypeExpression(members)
	}
	if expanded, ok := g.genericNamedUnionStructuralTypeExprForGenericUnion(pkgName, expr); ok {
		return expanded
	}
	return expr
}

func (g *generator) genericNamedUnionStructuralTypeExprForGenericUnion(pkgName string, expr ast.TypeExpression) (ast.TypeExpression, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	baseName, ok := typeExprBaseName(expr)
	if !ok || baseName == "" {
		return nil, false
	}
	union := g.unions[baseName]
	if union == nil || len(union.GenericParams) == 0 {
		return nil, false
	}
	_, members, ok := g.expandedUnionMembersInPackage(pkgName, expr)
	if !ok || len(members) == 0 {
		return nil, false
	}
	structural := make([]ast.TypeExpression, 0, len(members))
	for _, member := range members {
		if member == nil {
			return nil, false
		}
		structural = append(structural, g.genericNamedUnionStructuralTypeExpr(pkgName, member))
	}
	return ast.NewUnionTypeExpression(structural), true
}

func (g *generator) applyGenericNamedUnionMethodCallBindings(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || method.Info.Definition == nil || bindings == nil {
		return false
	}
	if len(call.TypeArguments) > 0 {
		if len(call.TypeArguments) != len(method.Info.Definition.GenericParams) {
			return false
		}
		if !g.bindGenericTypeArguments(method.Info.Package, bindings, method.Info.Definition.GenericParams, call.TypeArguments) {
			return false
		}
	}
	paramOffset := 0
	if method.ExpectsSelf {
		paramOffset = 1
	}
	for index, arg := range call.Arguments {
		paramIndex := paramOffset + index
		if paramIndex >= len(method.Info.Params) {
			break
		}
		paramType := method.Info.Params[paramIndex].TypeExpr
		if paramType == nil {
			continue
		}
		actual, actualGoType, ok := g.specializedCallActualTypeExpr(ctx, method.Info.Package, arg, paramType, bindings)
		actual, ok = g.specializationConcreteArgTypeExprForParam(method.Info.Package, paramType, actual, actualGoType)
		if !ok || actual == nil {
			continue
		}
		_ = g.bindSpecializedCallArgument(method.Info.Package, paramType, actual, genericNames, bindings)
	}
	return true
}

func (g *generator) hasGenericNamedUnionInstanceMethod(memberName string) bool {
	if g == nil || memberName == "" {
		return false
	}
	for _, method := range g.methodList {
		if method == nil || !method.ExpectsSelf || method.MethodName != memberName || method.TargetType == nil {
			continue
		}
		targetName, ok := typeExprBaseName(method.TargetType)
		if !ok || targetName == "" {
			continue
		}
		union := g.unions[targetName]
		if union != nil && len(union.GenericParams) > 0 {
			return true
		}
	}
	return false
}
