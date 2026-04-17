package compiler

func (g *generator) coerceExpectedNativeUnionExpr(ctx *compileContext, lines []string, expr string, actual string, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || expected == "" || actual == "" || expected == actual {
		return nil, "", "", false
	}
	targetType, ok := g.nativeUnionCoercibleTargetType(expected, actual)
	if !ok || targetType == "" || targetType == actual {
		return nil, "", "", false
	}
	coerceLines, coercedExpr, coercedType, ok := g.lowerCoerceExpectedStaticExpr(ctx, nil, expr, actual, targetType)
	if !ok || coercedType != targetType {
		return nil, "", "", false
	}
	wrapLines, wrapped, ok := g.lowerWrapUnion(ctx, expected, coercedType, coercedExpr)
	if !ok {
		return nil, "", "", false
	}
	out := append([]string{}, lines...)
	out = append(out, coerceLines...)
	out = append(out, wrapLines...)
	return out, wrapped, expected, true
}

func (g *generator) nativeUnionCoercibleTargetType(expected string, actual string) (string, bool) {
	if g == nil || expected == "" || actual == "" {
		return "", false
	}
	info := g.nativeUnionInfoForGoType(expected)
	if info == nil {
		return "", false
	}
	return g.nativeUnionCoercibleTargetTypeSeen(info, actual, make(map[string]struct{}))
}

func (g *generator) nativeUnionCoercibleTargetTypeSeen(info *nativeUnionInfo, actual string, seen map[string]struct{}) (string, bool) {
	if g == nil || info == nil || actual == "" {
		return "", false
	}
	if _, ok := seen[info.GoType]; ok {
		return "", false
	}
	seen[info.GoType] = struct{}{}
	defer delete(seen, info.GoType)
	targetType := ""
	ambiguous := false
	consider := func(candidate string) {
		if ambiguous || candidate == "" || candidate == actual || !g.canCoerceStaticExpr(candidate, actual) {
			return
		}
		if targetType == "" {
			targetType = candidate
			return
		}
		if targetType != candidate {
			ambiguous = true
		}
	}
	for _, member := range info.Members {
		if ambiguous || member == nil || member.GoType == "" {
			continue
		}
		if inner := g.nativeUnionInfoForGoType(member.GoType); inner != nil {
			candidate, ok := g.nativeUnionCoercibleTargetTypeSeen(inner, actual, seen)
			if ok {
				consider(candidate)
			}
			continue
		}
		if innerType, ok := g.nativeNullableValueInnerType(member.GoType); ok {
			consider(innerType)
		}
		consider(member.GoType)
	}
	if ambiguous || targetType == "" {
		return "", false
	}
	return targetType, true
}
