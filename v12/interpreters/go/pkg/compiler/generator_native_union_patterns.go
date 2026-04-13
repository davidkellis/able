package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

type nativeUnionPatternTarget struct {
	GoType          string
	Member          *nativeUnionMember
	InterfaceBranch bool
	Nested          *nativeUnionPatternTarget
}

func (g *generator) nativeCarrierImplementsInterface(goType string, interfaceName string) bool {
	if g == nil || goType == "" || interfaceName == "" {
		return false
	}
	if goType == "runtime.ErrorValue" && interfaceName == "Error" {
		return true
	}
	for _, impl := range g.implDefinitions {
		if impl == nil || impl.Definition == nil || impl.Definition.InterfaceName == nil {
			continue
		}
		if impl.Definition.InterfaceName.Name != interfaceName || impl.Definition.TargetType == nil {
			continue
		}
		targetType := normalizeTypeExprForPackage(g, impl.Package, impl.Definition.TargetType)
		if targetType == nil {
			targetType = impl.Definition.TargetType
		}
		mapped, ok := g.lowerCarrierTypeInPackage(impl.Package, targetType)
		if ok && mapped == goType {
			return true
		}
	}
	return false
}

func (g *generator) nativeUnionInterfacePatternMember(subjectType string, patternType ast.TypeExpression, pkgName string) (*nativeUnionMember, bool) {
	info := g.nativeUnionInfoForGoType(subjectType)
	if info == nil {
		return nil, false
	}
	normalized := normalizeTypeExprForPackage(g, pkgName, patternType)
	if expectedGoType, ok := g.lowerCarrierTypeInPackage(pkgName, normalized); ok && expectedGoType != "" {
		if expectedGoType == "runtime.ErrorValue" {
			var matched *nativeUnionMember
			for _, member := range info.Members {
				if member == nil || !g.isNativeErrorCarrierType(member.GoType) {
					continue
				}
				if matched != nil {
					return nil, false
				}
				matched = member
			}
			return matched, matched != nil
		}
		if iface := g.nativeInterfaceInfoForGoType(expectedGoType); iface != nil {
			var matched *nativeUnionMember
			for _, member := range info.Members {
				if member == nil || !g.nativeInterfaceAcceptsActual(iface, member.GoType) {
					continue
				}
				if matched != nil {
					return nil, false
				}
				matched = member
			}
			return matched, matched != nil
		}
	}
	simple, ok := normalized.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil || simple.Name.Name == "" {
		return nil, false
	}
	interfaceName := simple.Name.Name
	if !g.isInterfaceName(interfaceName) && interfaceName != "Error" {
		return nil, false
	}
	var matched *nativeUnionMember
	for _, member := range info.Members {
		if member == nil || !g.nativeCarrierImplementsInterface(member.GoType, interfaceName) {
			continue
		}
		if matched != nil {
			return nil, false
		}
		matched = member
	}
	return matched, matched != nil
}

func (g *generator) resolveNativeUnionTypedPattern(subjectType string, patternType ast.TypeExpression, pkgName string) (nativeUnionPatternTarget, bool) {
	return g.resolveNativeUnionTypedPatternSeen(subjectType, patternType, pkgName, make(map[string]struct{}))
}

func (g *generator) resolveNativeUnionTypedPatternSeen(subjectType string, patternType ast.TypeExpression, pkgName string, seen map[string]struct{}) (nativeUnionPatternTarget, bool) {
	if mapped, ok := g.lowerNativeUnionPatternMemberTypeInPackage(pkgName, subjectType, patternType); ok {
		if mapped == subjectType {
			return nativeUnionPatternTarget{GoType: mapped}, true
		}
		union := g.nativeUnionInfoForGoType(subjectType)
		if union == nil {
			return nativeUnionPatternTarget{}, false
		}
		member, ok := g.nativeUnionMember(union, mapped)
		if !ok {
			for _, nullableMember := range union.Members {
				if nullableMember == nil || !g.nativeNullableWraps(nullableMember.GoType, mapped) {
					continue
				}
				return nativeUnionPatternTarget{GoType: mapped, Member: nullableMember}, true
			}
			return nativeUnionPatternTarget{}, false
		}
		return nativeUnionPatternTarget{GoType: mapped, Member: member}, true
	}
	member, ok := g.nativeUnionInterfacePatternMember(subjectType, patternType, pkgName)
	if ok && member != nil {
		return nativeUnionPatternTarget{
			GoType:          member.GoType,
			Member:          member,
			InterfaceBranch: true,
		}, true
	}
	union := g.nativeUnionInfoForGoType(subjectType)
	if union == nil {
		return nativeUnionPatternTarget{}, false
	}
	if _, ok := seen[subjectType]; ok {
		return nativeUnionPatternTarget{}, false
	}
	seen[subjectType] = struct{}{}
	defer delete(seen, subjectType)
	var matched *nativeUnionPatternTarget
	for _, outerMember := range union.Members {
		if outerMember == nil || outerMember.GoType == "" {
			continue
		}
		if g.nativeUnionInfoForGoType(outerMember.GoType) == nil {
			continue
		}
		nested, ok := g.resolveNativeUnionTypedPatternSeen(outerMember.GoType, patternType, pkgName, seen)
		if !ok {
			continue
		}
		candidate := &nativeUnionPatternTarget{
			GoType: outerMember.GoType,
			Member: outerMember,
			Nested: &nested,
		}
		if matched != nil {
			return nativeUnionPatternTarget{}, false
		}
		matched = candidate
	}
	if matched == nil {
		return nativeUnionPatternTarget{}, false
	}
	return *matched, true
}

func (g *generator) resolveNativeUnionTypedPatternInContext(ctx *compileContext, subjectType string, patternType ast.TypeExpression) (nativeUnionPatternTarget, bool) {
	if ctx == nil {
		return g.resolveNativeUnionTypedPattern(subjectType, patternType, "")
	}
	return g.resolveNativeUnionTypedPattern(subjectType, g.lowerNormalizedTypeExpr(ctx, patternType), ctx.packageName)
}

func (g *generator) nativeUnionDynamicTypedPatternMember(ctx *compileContext, subjectType string, patternType ast.TypeExpression) (*nativeUnionMember, string, bool) {
	if g == nil || ctx == nil || patternType == nil {
		return nil, "", false
	}
	union := g.nativeUnionInfoForGoType(subjectType)
	if union == nil {
		return nil, "", false
	}
	narrowedType, ok := g.recoverTypedPatternCarrier(ctx, patternType)
	if !ok {
		return nil, "", false
	}
	if g.isNativeErrorCarrierType(narrowedType) {
		if member, ok := g.nativeUnionMember(union, "runtime.ErrorValue"); ok && member != nil {
			return member, narrowedType, true
		}
	}
	if member, ok := g.nativeUnionMember(union, "runtime.Value"); ok && member != nil {
		return member, narrowedType, true
	}
	return nil, "", false
}

func (g *generator) compileNativeUnionDynamicTypedPatternCondition(ctx *compileContext, subjectTemp string, member *nativeUnionMember, pattern *ast.TypedPattern) ([]string, string, bool) {
	if g == nil || ctx == nil || member == nil || pattern == nil || pattern.TypeAnnotation == nil {
		return nil, "", false
	}
	memberTemp := ctx.newTemp()
	okTemp := ctx.newTemp()
	guardLines := []string{fmt.Sprintf("%s, %s := %s(%s)", memberTemp, okTemp, member.UnwrapHelper, subjectTemp)}
	if g.nativeUnionDynamicTypedPatternDirectRuntimeMatch(ctx, member, pattern.TypeAnnotation) {
		if isTypedPatternConditionOnly(pattern.Pattern) {
			return guardLines, okTemp, true
		}
		innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, pattern.Pattern, memberTemp, member.GoType)
		if !ok {
			return nil, "", false
		}
		guardedLines, cond, ok := g.guardMatchConditionWithPredicate(ctx, okTemp, innerCondLines, innerCond)
		if !ok {
			return nil, "", false
		}
		return append(guardLines, guardedLines...), cond, true
	}
	if isTypedPatternConditionOnly(pattern.Pattern) {
		if directLines, directOK, ok := g.compileDirectDynamicTypedPatternCondition(ctx, memberTemp, pattern.TypeAnnotation); ok {
			lines := append([]string{}, guardLines...)
			lines = append(lines, directLines...)
			return lines, fmt.Sprintf("(%s && %s)", okTemp, directOK), true
		}
		typeExpr, ok := g.renderTypeExpression(g.lowerNormalizedTypeExpr(ctx, pattern.TypeAnnotation))
		if !ok {
			return nil, "", false
		}
		g.needsAst = true
		castOK := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := append([]string{}, guardLines...)
		lines = append(lines, fmt.Sprintf("_, %s, %s := __able_try_cast(%s, %s)", castOK, controlTemp, memberTemp, typeExpr))
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, fmt.Sprintf("(%s && %s)", okTemp, castOK), true
	}
	castLines, narrowedTemp, narrowedType, narrowedOK, ok := g.compileDynamicTypedPatternCast(ctx, memberTemp, member.GoType, pattern.TypeAnnotation)
	if !ok {
		return nil, "", false
	}
	innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, pattern.Pattern, narrowedTemp, narrowedType)
	if !ok {
		return nil, "", false
	}
	innerLines := append([]string{}, castLines...)
	innerLines = append(innerLines, innerCondLines...)
	guardCond := narrowedOK
	if innerCond != "true" {
		guardCond = fmt.Sprintf("(%s && %s)", narrowedOK, innerCond)
	}
	guardedLines, cond, ok := g.guardMatchConditionWithPredicate(ctx, okTemp, innerLines, guardCond)
	if !ok {
		return nil, "", false
	}
	return append(guardLines, guardedLines...), cond, true
}

func (g *generator) compileNativeUnionDynamicTypedPatternBindings(ctx *compileContext, subjectTemp string, member *nativeUnionMember, pattern *ast.TypedPattern) ([]string, bool) {
	if g == nil || ctx == nil || member == nil || pattern == nil || pattern.TypeAnnotation == nil {
		return nil, false
	}
	memberTemp := ctx.newTemp()
	castLines := []string{}
	narrowedTemp := memberTemp
	narrowedType := member.GoType
	if !g.nativeUnionDynamicTypedPatternDirectRuntimeMatch(ctx, member, pattern.TypeAnnotation) {
		var ok bool
		castLines, narrowedTemp, narrowedType, _, ok = g.compileDynamicTypedPatternCast(ctx, memberTemp, member.GoType, pattern.TypeAnnotation)
		if !ok {
			return nil, false
		}
	}
	previousExpected := ctx.expectedTypeExpr
	ctx.expectedTypeExpr = pattern.TypeAnnotation
	bindLines, ok := g.compileMatchPatternBindings(ctx, pattern.Pattern, narrowedTemp, narrowedType)
	ctx.expectedTypeExpr = previousExpected
	if !ok {
		return nil, false
	}
	if len(bindLines) == 0 {
		return nil, true
	}
	lines := []string{fmt.Sprintf("%s, _ := %s(%s)", memberTemp, member.UnwrapHelper, subjectTemp)}
	lines = append(lines, castLines...)
	lines = append(lines, bindLines...)
	return lines, true
}

func isTypedPatternConditionOnly(pattern ast.Pattern) bool {
	switch pattern.(type) {
	case nil:
		return true
	case *ast.WildcardPattern:
		return true
	case *ast.Identifier:
		return true
	default:
		return false
	}
}

func (g *generator) compileNativeUnionDynamicTypedAssignmentPatternBindings(ctx *compileContext, subjectTemp string, member *nativeUnionMember, pattern *ast.TypedPattern, mode patternBindingMode) ([]string, bool) {
	if g == nil || ctx == nil || member == nil || pattern == nil || pattern.TypeAnnotation == nil {
		return nil, false
	}
	memberTemp := ctx.newTemp()
	castLines := []string{}
	narrowedTemp := memberTemp
	narrowedType := member.GoType
	if !g.nativeUnionDynamicTypedPatternDirectRuntimeMatch(ctx, member, pattern.TypeAnnotation) {
		var ok bool
		castLines, narrowedTemp, narrowedType, _, ok = g.compileDynamicTypedPatternCast(ctx, memberTemp, member.GoType, pattern.TypeAnnotation)
		if !ok {
			return nil, false
		}
	}
	previousExpected := ctx.expectedTypeExpr
	ctx.expectedTypeExpr = pattern.TypeAnnotation
	bindLines, ok := g.compileAssignmentPatternBindings(ctx, pattern.Pattern, narrowedTemp, narrowedType, mode)
	ctx.expectedTypeExpr = previousExpected
	if !ok {
		return nil, false
	}
	if len(bindLines) == 0 {
		return nil, true
	}
	lines := []string{fmt.Sprintf("%s, _ := %s(%s)", memberTemp, member.UnwrapHelper, subjectTemp)}
	lines = append(lines, castLines...)
	lines = append(lines, bindLines...)
	return lines, true
}

func (g *generator) nativeUnionDynamicTypedPatternDirectRuntimeMatch(ctx *compileContext, member *nativeUnionMember, patternType ast.TypeExpression) bool {
	if g == nil || ctx == nil || member == nil || patternType == nil || member.GoType != "runtime.Value" || member.TypeExpr == nil {
		return false
	}
	memberExpr := g.lowerNormalizedTypeExpr(ctx, member.TypeExpr)
	patternExpr := g.lowerNormalizedTypeExpr(ctx, patternType)
	if memberExpr == nil || patternExpr == nil {
		return false
	}
	memberKey := normalizeTypeExprIdentityKey(g, ctx.packageName, memberExpr)
	patternKey := normalizeTypeExprIdentityKey(g, ctx.packageName, patternExpr)
	return memberKey != "" && memberKey == patternKey
}

func nativeUnionWholeValueBinding(pattern ast.Pattern) bool {
	ident, ok := pattern.(*ast.Identifier)
	return ok && ident != nil && ident.Name != "" && ident.Name != "_"
}

func (g *generator) compileResolvedNativeUnionTypedPatternCondition(ctx *compileContext, subjectTemp string, subjectType string, target nativeUnionPatternTarget, pattern *ast.TypedPattern) ([]string, string, bool) {
	if target.Member == nil {
		innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, pattern.Pattern, subjectTemp, subjectType)
		if !ok {
			return nil, "", false
		}
		return g.guardStaticTypedPatternNonNil(ctx, subjectTemp, subjectType, pattern.TypeAnnotation, innerCondLines, innerCond)
	}
	if target.Nested != nil {
		okTemp := ctx.newTemp()
		memberTemp := ctx.newTemp()
		nestedLines, nestedCond, ok := g.compileResolvedNativeUnionTypedPatternCondition(ctx, memberTemp, target.GoType, *target.Nested, pattern)
		if !ok {
			return nil, "", false
		}
		lines := []string{fmt.Sprintf("%s, %s := %s(%s)", memberTemp, okTemp, target.Member.UnwrapHelper, subjectTemp)}
		guardedLines, cond, ok := g.guardMatchConditionWithPredicate(ctx, okTemp, nestedLines, nestedCond)
		if !ok {
			return nil, "", false
		}
		return append(lines, guardedLines...), cond, true
	}
	if target.Member.GoType != target.GoType && g.nativeNullableWraps(target.Member.GoType, target.GoType) {
		okTemp := ctx.newTemp()
		memberTemp := ctx.newTemp()
		innerLines, innerCond, ok := g.compileMatchPatternCondition(ctx, pattern, memberTemp, target.Member.GoType)
		if !ok {
			return nil, "", false
		}
		lines := []string{fmt.Sprintf("%s, %s := %s(%s)", memberTemp, okTemp, target.Member.UnwrapHelper, subjectTemp)}
		guardedLines, cond, ok := g.guardMatchConditionWithPredicate(ctx, okTemp, innerLines, innerCond)
		if !ok {
			return nil, "", false
		}
		return append(lines, guardedLines...), cond, true
	}
	if g.nativeUnionDynamicTypedPatternDirectRuntimeMatch(ctx, target.Member, pattern.TypeAnnotation) {
		okTemp := ctx.newTemp()
		castTemp := ctx.newTemp()
		innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, pattern.Pattern, castTemp, target.GoType)
		if !ok {
			return nil, "", false
		}
		if innerCond == "true" && len(innerCondLines) == 0 {
			return []string{fmt.Sprintf("%s, %s := %s(%s)", castTemp, okTemp, target.Member.UnwrapHelper, subjectTemp)}, okTemp, true
		}
		condTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s, %s := %s(%s)", castTemp, okTemp, target.Member.UnwrapHelper, subjectTemp),
			fmt.Sprintf("var %s bool", condTemp),
			fmt.Sprintf("if %s {", okTemp),
		}
		lines = append(lines, indentLines(innerCondLines, 1)...)
		lines = append(lines, fmt.Sprintf("\t%s = %s", condTemp, innerCond))
		lines = append(lines, "}")
		return lines, condTemp, true
	}
	okTemp := ctx.newTemp()
	castTemp := ctx.newTemp()
	innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, pattern.Pattern, castTemp, target.GoType)
	if !ok {
		return nil, "", false
	}
	if innerCond == "true" && len(innerCondLines) == 0 {
		return []string{fmt.Sprintf("_, %s := %s(%s)", okTemp, target.Member.UnwrapHelper, subjectTemp)}, okTemp, true
	}
	condTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s, %s := %s(%s)", castTemp, okTemp, target.Member.UnwrapHelper, subjectTemp),
		fmt.Sprintf("var %s bool", condTemp),
		fmt.Sprintf("if %s {", okTemp),
	}
	lines = append(lines, indentLines(innerCondLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", condTemp, innerCond))
	lines = append(lines, "}")
	return lines, condTemp, true
}

func (g *generator) compileResolvedNativeUnionTypedPatternBindings(ctx *compileContext, subjectTemp string, subjectType string, target nativeUnionPatternTarget, pattern *ast.TypedPattern) ([]string, bool) {
	if target.Member == nil {
		return g.compileMatchPatternBindings(ctx, pattern.Pattern, subjectTemp, subjectType)
	}
	convertedTemp := ctx.newTemp()
	lines := []string{fmt.Sprintf("%s, _ := %s(%s)", convertedTemp, target.Member.UnwrapHelper, subjectTemp)}
	if target.Nested != nil {
		bindLines, ok := g.compileResolvedNativeUnionTypedPatternBindings(ctx, convertedTemp, target.GoType, *target.Nested, pattern)
		if !ok {
			return nil, false
		}
		if len(bindLines) == 0 {
			return nil, true
		}
		lines = append(lines, bindLines...)
		return lines, true
	}
	if target.Member.GoType != target.GoType && g.nativeNullableWraps(target.Member.GoType, target.GoType) {
		bindLines, ok := g.compileMatchPatternBindings(ctx, pattern, convertedTemp, target.Member.GoType)
		if !ok {
			return nil, false
		}
		if len(bindLines) == 0 {
			return nil, true
		}
		lines = append(lines, bindLines...)
		return lines, true
	}
	bindSubject := convertedTemp
	bindType := target.GoType
	expectedTypeExpr := ast.TypeExpression(nil)
	if g.nativeUnionDynamicTypedPatternDirectRuntimeMatch(ctx, target.Member, pattern.TypeAnnotation) {
		previousExpected := ctx.expectedTypeExpr
		ctx.expectedTypeExpr = pattern.TypeAnnotation
		bindLines, ok := g.compileMatchPatternBindings(ctx, pattern.Pattern, bindSubject, bindType)
		ctx.expectedTypeExpr = previousExpected
		if !ok {
			return nil, false
		}
		if len(bindLines) == 0 {
			return nil, true
		}
		lines = append(lines, bindLines...)
		return lines, true
	}
	if target.InterfaceBranch && nativeUnionWholeValueBinding(pattern.Pattern) {
		expectedGoType, ok := g.recoverTypedPatternCarrier(ctx, pattern.TypeAnnotation)
		if !ok {
			ctx.setReason("typed pattern type mismatch")
			return nil, false
		}
		coerceLines, coercedExpr, coercedType, ok := g.lowerCoerceExpectedStaticExpr(ctx, nil, convertedTemp, target.GoType, expectedGoType)
		if !ok {
			ctx.setReason("typed pattern type mismatch")
			return nil, false
		}
		lines = append(lines, coerceLines...)
		bindSubject = coercedExpr
		bindType = coercedType
		expectedTypeExpr = pattern.TypeAnnotation
	}
	previousExpected := ctx.expectedTypeExpr
	if expectedTypeExpr != nil {
		ctx.expectedTypeExpr = expectedTypeExpr
	}
	bindLines, ok := g.compileMatchPatternBindings(ctx, pattern.Pattern, bindSubject, bindType)
	ctx.expectedTypeExpr = previousExpected
	if !ok {
		return nil, false
	}
	if len(bindLines) == 0 {
		return nil, true
	}
	lines = append(lines, bindLines...)
	return lines, true
}

func (g *generator) compileResolvedNativeUnionTypedAssignmentPatternBindings(ctx *compileContext, subjectTemp string, subjectType string, target nativeUnionPatternTarget, pattern *ast.TypedPattern, mode patternBindingMode) ([]string, bool) {
	if target.Member == nil {
		return g.compileAssignmentPatternBindings(ctx, pattern.Pattern, subjectTemp, subjectType, mode)
	}
	convertedTemp := ctx.newTemp()
	lines := []string{fmt.Sprintf("%s, _ := %s(%s)", convertedTemp, target.Member.UnwrapHelper, subjectTemp)}
	if target.Nested != nil {
		bindLines, ok := g.compileResolvedNativeUnionTypedAssignmentPatternBindings(ctx, convertedTemp, target.GoType, *target.Nested, pattern, mode)
		if !ok {
			return nil, false
		}
		if len(bindLines) == 0 {
			return nil, true
		}
		lines = append(lines, bindLines...)
		return lines, true
	}
	if target.Member.GoType != target.GoType && g.nativeNullableWraps(target.Member.GoType, target.GoType) {
		bindLines, ok := g.compileAssignmentPatternBindings(ctx, pattern, convertedTemp, target.Member.GoType, mode)
		if !ok {
			return nil, false
		}
		if len(bindLines) == 0 {
			return nil, true
		}
		lines = append(lines, bindLines...)
		return lines, true
	}
	bindSubject := convertedTemp
	bindType := target.GoType
	expectedTypeExpr := ast.TypeExpression(nil)
	if g.nativeUnionDynamicTypedPatternDirectRuntimeMatch(ctx, target.Member, pattern.TypeAnnotation) {
		previousExpected := ctx.expectedTypeExpr
		ctx.expectedTypeExpr = pattern.TypeAnnotation
		bindLines, ok := g.compileAssignmentPatternBindings(ctx, pattern.Pattern, bindSubject, bindType, mode)
		ctx.expectedTypeExpr = previousExpected
		if !ok {
			return nil, false
		}
		if len(bindLines) == 0 {
			return nil, true
		}
		lines = append(lines, bindLines...)
		return lines, true
	}
	if target.InterfaceBranch && nativeUnionWholeValueBinding(pattern.Pattern) {
		expectedGoType, ok := g.recoverTypedPatternCarrier(ctx, pattern.TypeAnnotation)
		if !ok {
			ctx.setReason("typed pattern type mismatch")
			return nil, false
		}
		coerceLines, coercedExpr, coercedType, ok := g.lowerCoerceExpectedStaticExpr(ctx, nil, convertedTemp, target.GoType, expectedGoType)
		if !ok {
			ctx.setReason("typed pattern type mismatch")
			return nil, false
		}
		lines = append(lines, coerceLines...)
		bindSubject = coercedExpr
		bindType = coercedType
		expectedTypeExpr = pattern.TypeAnnotation
	}
	previousExpected := ctx.expectedTypeExpr
	if expectedTypeExpr != nil {
		ctx.expectedTypeExpr = expectedTypeExpr
	}
	bindLines, ok := g.compileAssignmentPatternBindings(ctx, pattern.Pattern, bindSubject, bindType, mode)
	ctx.expectedTypeExpr = previousExpected
	if !ok {
		return nil, false
	}
	if len(bindLines) == 0 {
		return nil, true
	}
	lines = append(lines, bindLines...)
	return lines, true
}

func (g *generator) compileNativeUnionTypedPatternCondition(ctx *compileContext, subjectTemp string, subjectType string, pattern *ast.TypedPattern) ([]string, string, bool) {
	if pattern == nil || pattern.TypeAnnotation == nil {
		ctx.setReason("missing typed pattern annotation")
		return nil, "", false
	}
	target, ok := g.resolveNativeUnionTypedPatternInContext(ctx, subjectType, pattern.TypeAnnotation)
	if !ok {
		member, _, dynamicOK := g.nativeUnionDynamicTypedPatternMember(ctx, subjectType, pattern.TypeAnnotation)
		if !dynamicOK {
			ctx.setReason("typed pattern type mismatch")
			return nil, "", false
		}
		return g.compileNativeUnionDynamicTypedPatternCondition(ctx, subjectTemp, member, pattern)
	}
	return g.compileResolvedNativeUnionTypedPatternCondition(ctx, subjectTemp, subjectType, target, pattern)
}

func (g *generator) compileNativeUnionTypedPatternBindings(ctx *compileContext, subjectTemp string, subjectType string, pattern *ast.TypedPattern) ([]string, bool) {
	if pattern == nil || pattern.TypeAnnotation == nil {
		ctx.setReason("missing typed pattern annotation")
		return nil, false
	}
	target, ok := g.resolveNativeUnionTypedPatternInContext(ctx, subjectType, pattern.TypeAnnotation)
	if !ok {
		member, _, dynamicOK := g.nativeUnionDynamicTypedPatternMember(ctx, subjectType, pattern.TypeAnnotation)
		if !dynamicOK {
			ctx.setReason("typed pattern type mismatch")
			return nil, false
		}
		return g.compileNativeUnionDynamicTypedPatternBindings(ctx, subjectTemp, member, pattern)
	}
	return g.compileResolvedNativeUnionTypedPatternBindings(ctx, subjectTemp, subjectType, target, pattern)
}

func (g *generator) compileNativeUnionTypedAssignmentPatternBindings(ctx *compileContext, subjectTemp string, subjectType string, pattern *ast.TypedPattern, mode patternBindingMode) ([]string, bool) {
	if pattern == nil || pattern.TypeAnnotation == nil {
		ctx.setReason("missing typed pattern annotation")
		return nil, false
	}
	target, ok := g.resolveNativeUnionTypedPatternInContext(ctx, subjectType, pattern.TypeAnnotation)
	if !ok {
		member, _, dynamicOK := g.nativeUnionDynamicTypedPatternMember(ctx, subjectType, pattern.TypeAnnotation)
		if !dynamicOK {
			ctx.setReason("typed pattern type mismatch")
			return nil, false
		}
		return g.compileNativeUnionDynamicTypedAssignmentPatternBindings(ctx, subjectTemp, member, pattern, mode)
	}
	return g.compileResolvedNativeUnionTypedAssignmentPatternBindings(ctx, subjectTemp, subjectType, target, pattern, mode)
}
