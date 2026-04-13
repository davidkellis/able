package compiler

import "fmt"

func (g *generator) nativeUnionAcceptsActual(info *nativeUnionInfo, actual string) bool {
	return g.nativeUnionAcceptsActualSeen(info, actual, make(map[string]struct{}))
}

func (g *generator) nativeUnionAcceptsActualSeen(info *nativeUnionInfo, actual string, seen map[string]struct{}) bool {
	if g == nil || info == nil || actual == "" {
		return false
	}
	if _, ok := seen[info.GoType]; ok {
		return false
	}
	seen[info.GoType] = struct{}{}
	defer delete(seen, info.GoType)
	if _, ok := g.nativeUnionMember(info, actual); ok {
		return true
	}
	for _, member := range info.Members {
		if member == nil {
			continue
		}
		if iface := g.nativeInterfaceInfoForGoType(member.GoType); iface != nil && g.nativeInterfaceAcceptsActual(iface, actual) {
			return true
		}
		if member.GoType == "runtime.ErrorValue" && g.isNativeErrorCarrierType(actual) {
			return true
		}
		if g.nativeNullableWraps(member.GoType, actual) {
			return true
		}
		if g.nativeUnionRuntimeMemberAcceptsActual(member, actual) {
			return true
		}
		if inner := g.nativeUnionInfoForGoType(member.GoType); inner != nil && g.nativeUnionAcceptsActualSeen(inner, actual, seen) {
			return true
		}
	}
	return false
}

func (g *generator) nativeUnionLiteralTargetType(expected string, actual string, explicit bool, predicate func(string) bool) (string, bool) {
	if g == nil || predicate == nil {
		return "", false
	}
	info := g.nativeUnionInfoForGoType(expected)
	if info == nil {
		return "", false
	}
	return g.nativeUnionLiteralTargetTypeSeen(info, actual, explicit, predicate, make(map[string]struct{}))
}

func (g *generator) nativeUnionLiteralTargetTypeSeen(info *nativeUnionInfo, actual string, explicit bool, predicate func(string) bool, seen map[string]struct{}) (string, bool) {
	if g == nil || info == nil || predicate == nil {
		return "", false
	}
	if actual != "" && predicate(actual) && g.nativeUnionAcceptsActualSeen(info, actual, seen) {
		return actual, true
	}
	if explicit {
		return "", false
	}
	if _, ok := seen[info.GoType]; ok {
		return "", false
	}
	seen[info.GoType] = struct{}{}
	defer delete(seen, info.GoType)
	targetType := ""
	for _, member := range info.Members {
		if member == nil {
			continue
		}
		candidate := ""
		switch {
		case predicate(member.GoType):
			candidate = member.GoType
		case func() bool {
			innerType, ok := g.nativeNullableValueInnerType(member.GoType)
			if !ok || !predicate(innerType) {
				return false
			}
			candidate = innerType
			return true
		}():
		case g.nativeUnionInfoForGoType(member.GoType) != nil:
			inner := g.nativeUnionInfoForGoType(member.GoType)
			var ok bool
			candidate, ok = g.nativeUnionLiteralTargetTypeSeen(inner, actual, false, predicate, seen)
			if !ok {
				continue
			}
		default:
			continue
		}
		if candidate == "" {
			continue
		}
		if targetType != "" && targetType != candidate {
			return "", false
		}
		targetType = candidate
	}
	if targetType == "" {
		return "", false
	}
	return targetType, true
}

func (g *generator) nativeUnionWrapExprSeen(info *nativeUnionInfo, actual, expr string, seen map[string]struct{}) (string, bool) {
	if g == nil || info == nil || actual == "" || expr == "" {
		return "", false
	}
	if _, ok := seen[info.GoType]; ok {
		return "", false
	}
	seen[info.GoType] = struct{}{}
	defer delete(seen, info.GoType)
	if member, ok := g.nativeUnionMember(info, actual); ok && member != nil {
		return fmt.Sprintf("%s(%s)", member.WrapHelper, expr), true
	}
	for _, member := range info.Members {
		if member == nil {
			continue
		}
		if g.nativeNullableWraps(member.GoType, actual) {
			return fmt.Sprintf("%s(__able_ptr(%s))", member.WrapHelper, expr), true
		}
		if inner := g.nativeUnionInfoForGoType(member.GoType); inner != nil {
			if wrapped, ok := g.nativeUnionWrapExprSeen(inner, actual, expr, seen); ok {
				return fmt.Sprintf("%s(%s)", member.WrapHelper, wrapped), true
			}
		}
	}
	return "", false
}

func (g *generator) nativeUnionWrapLinesSeen(ctx *compileContext, info *nativeUnionInfo, actual, expr string, seen map[string]struct{}) ([]string, string, bool) {
	if wrapped, ok := g.nativeUnionWrapExprSeen(info, actual, expr, seen); ok {
		return nil, wrapped, true
	}
	if g == nil || ctx == nil || info == nil || actual == "" || expr == "" {
		return nil, "", false
	}
	if _, ok := seen[info.GoType]; ok {
		return nil, "", false
	}
	seen[info.GoType] = struct{}{}
	defer delete(seen, info.GoType)
	for _, member := range info.Members {
		if member == nil {
			continue
		}
		if iface := g.nativeInterfaceInfoForGoType(member.GoType); iface != nil && g.nativeInterfaceAcceptsActual(iface, actual) {
			ifaceLines, ifaceExpr, ok := g.lowerWrapInterface(ctx, member.GoType, actual, expr)
			if !ok {
				return nil, "", false
			}
			return ifaceLines, fmt.Sprintf("%s(%s)", member.WrapHelper, ifaceExpr), true
		}
		if member.GoType == "runtime.ErrorValue" && g.isNativeErrorCarrierType(actual) {
			lines, errorExpr, ok := g.nativeErrorValueLines(ctx, actual, expr)
			if !ok {
				return nil, "", false
			}
			return lines, fmt.Sprintf("%s(%s)", member.WrapHelper, errorExpr), true
		}
		if inner := g.nativeUnionInfoForGoType(member.GoType); inner != nil {
			lines, wrapped, ok := g.nativeUnionWrapLinesSeen(ctx, inner, actual, expr, seen)
			if ok {
				return lines, fmt.Sprintf("%s(%s)", member.WrapHelper, wrapped), true
			}
		}
		if !g.nativeUnionRuntimeMemberAcceptsActual(member, actual) {
			continue
		}
		if actual == "runtime.Value" {
			return nil, fmt.Sprintf("%s(%s)", member.WrapHelper, expr), true
		}
		runtimeLines, runtimeExpr, ok := g.lowerRuntimeValue(ctx, expr, actual)
		if !ok {
			return nil, "", false
		}
		return runtimeLines, fmt.Sprintf("%s(%s)", member.WrapHelper, runtimeExpr), true
	}
	return nil, "", false
}
