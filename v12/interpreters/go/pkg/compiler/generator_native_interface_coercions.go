package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) nativeInterfaceAssignable(actual string, expected string) bool {
	if g == nil || actual == "" || expected == "" {
		return false
	}
	actualInfo := g.nativeInterfaceInfoForGoType(actual)
	expectedInfo := g.nativeInterfaceInfoForGoType(expected)
	if actualInfo == nil || expectedInfo == nil {
		return false
	}
	for _, expectedMethod := range expectedInfo.Methods {
		found := false
		for _, actualMethod := range actualInfo.Methods {
			if actualMethod == nil || expectedMethod == nil || actualMethod.Name != expectedMethod.Name {
				continue
			}
			if actualMethod.OptionalLast != expectedMethod.OptionalLast || len(actualMethod.ParamGoTypes) != len(expectedMethod.ParamGoTypes) {
				continue
			}
			if g.nativeInterfaceMethodShapeAssignable(actualMethod, expectedMethod) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (g *generator) nativeInterfaceMethodShapeAssignable(actualMethod, expectedMethod *nativeInterfaceMethod) bool {
	if g == nil || actualMethod == nil || expectedMethod == nil {
		return false
	}
	if actualMethod.ReturnGoType == expectedMethod.ReturnGoType && len(actualMethod.ParamGoTypes) == len(expectedMethod.ParamGoTypes) {
		same := true
		for idx := range actualMethod.ParamGoTypes {
			if actualMethod.ParamGoTypes[idx] != expectedMethod.ParamGoTypes[idx] {
				same = false
				break
			}
		}
		if same {
			return true
		}
	}
	leftVars := make(map[string]string)
	rightVars := make(map[string]string)
	for idx := range actualMethod.ParamTypeExprs {
		if !g.typeExprEquivalentModuloGenerics(actualMethod.ParamTypeExprs[idx], expectedMethod.ParamTypeExprs[idx], leftVars, rightVars) {
			return false
		}
	}
	return g.typeExprEquivalentModuloGenerics(actualMethod.ReturnTypeExpr, expectedMethod.ReturnTypeExpr, leftVars, rightVars)
}

func (g *generator) typeExprEquivalentModuloGenerics(left ast.TypeExpression, right ast.TypeExpression, leftVars map[string]string, rightVars map[string]string) bool {
	switch l := left.(type) {
	case nil:
		return right == nil
	case *ast.SimpleTypeExpression:
		r, ok := right.(*ast.SimpleTypeExpression)
		if !ok || l == nil || r == nil || l.Name == nil || r.Name == nil {
			return false
		}
		return g.simpleTypeEquivalentModuloGenerics(l.Name.Name, r.Name.Name, leftVars, rightVars)
	case *ast.GenericTypeExpression:
		r, ok := right.(*ast.GenericTypeExpression)
		if !ok || l == nil || r == nil || len(l.Arguments) != len(r.Arguments) {
			return false
		}
		if !g.typeExprEquivalentModuloGenerics(l.Base, r.Base, leftVars, rightVars) {
			return false
		}
		for idx := range l.Arguments {
			if !g.typeExprEquivalentModuloGenerics(l.Arguments[idx], r.Arguments[idx], leftVars, rightVars) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		r, ok := right.(*ast.NullableTypeExpression)
		return ok && l != nil && r != nil && g.typeExprEquivalentModuloGenerics(l.InnerType, r.InnerType, leftVars, rightVars)
	case *ast.ResultTypeExpression:
		r, ok := right.(*ast.ResultTypeExpression)
		return ok && l != nil && r != nil && g.typeExprEquivalentModuloGenerics(l.InnerType, r.InnerType, leftVars, rightVars)
	case *ast.UnionTypeExpression:
		r, ok := right.(*ast.UnionTypeExpression)
		if !ok || l == nil || r == nil || len(l.Members) != len(r.Members) {
			return false
		}
		for idx := range l.Members {
			if !g.typeExprEquivalentModuloGenerics(l.Members[idx], r.Members[idx], leftVars, rightVars) {
				return false
			}
		}
		return true
	case *ast.FunctionTypeExpression:
		r, ok := right.(*ast.FunctionTypeExpression)
		if !ok || l == nil || r == nil || len(l.ParamTypes) != len(r.ParamTypes) {
			return false
		}
		for idx := range l.ParamTypes {
			if !g.typeExprEquivalentModuloGenerics(l.ParamTypes[idx], r.ParamTypes[idx], leftVars, rightVars) {
				return false
			}
		}
		return g.typeExprEquivalentModuloGenerics(l.ReturnType, r.ReturnType, leftVars, rightVars)
	default:
		return typeExpressionToString(left) == typeExpressionToString(right)
	}
}

func (g *generator) simpleTypeEquivalentModuloGenerics(leftName string, rightName string, leftVars map[string]string, rightVars map[string]string) bool {
	leftConcrete := g.isConcreteTypeName(leftName)
	rightConcrete := g.isConcreteTypeName(rightName)
	if leftConcrete || rightConcrete {
		return leftConcrete && rightConcrete && leftName == rightName
	}
	if mapped, ok := leftVars[leftName]; ok {
		return mapped == rightName
	}
	if mapped, ok := rightVars[rightName]; ok {
		return mapped == leftName
	}
	leftVars[leftName] = rightName
	rightVars[rightName] = leftName
	return true
}

func (g *generator) isConcreteTypeName(name string) bool {
	switch strings.TrimSpace(name) {
	case "", "bool", "Bool", "String", "string", "char", "Char", "i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64", "isize", "usize", "f32", "f64", "void", "Void", "Error", "Value", "nil":
		return true
	}
	if g == nil {
		return false
	}
	if g.isInterfaceName(name) {
		return true
	}
	if _, ok := g.structInfoByNameUnique(name); ok {
		return true
	}
	if _, ok := g.unionPackages[name]; ok {
		return true
	}
	for _, perPkg := range g.typeAliases {
		if perPkg == nil {
			continue
		}
		if _, ok := perPkg[name]; ok {
			return true
		}
	}
	return false
}

func (g *generator) nativeInterfaceAcceptsActual(info *nativeInterfaceInfo, actual string) bool {
	return g.nativeInterfaceAcceptsActualSeen(info, actual, make(map[string]struct{}))
}

func (g *generator) nativeInterfaceAcceptsActualSeen(info *nativeInterfaceInfo, actual string, seen map[string]struct{}) bool {
	if g == nil || info == nil || actual == "" {
		return false
	}
	if seen == nil {
		seen = make(map[string]struct{})
	}
	seenKey := info.Key + "|" + actual
	if _, ok := seen[seenKey]; ok {
		return false
	}
	seen[seenKey] = struct{}{}
	defer delete(seen, seenKey)
	if actual == info.GoType || actual == "runtime.Value" {
		return true
	}
	if actual == "any" {
		return true
	}
	if g.nativeInterfaceAssignable(actual, info.GoType) {
		return true
	}
	if _, ok := g.nativeInterfaceAdapterForActualSeen(info, actual, seen); ok {
		return true
	}
	actualUnion := g.nativeUnionInfoForGoType(actual)
	for _, adapter := range g.nativeInterfaceKnownAdapters(info) {
		if adapter == nil || adapter.GoType == "" {
			continue
		}
		if g.sameNominalStructFamily(adapter.GoType, actual) {
			return true
		}
		union := g.nativeUnionInfoForGoType(adapter.GoType)
		if union == nil {
			continue
		}
		if actualUnion == nil {
			continue
		}
		if _, ok := g.nativeUnionMember(union, actual); ok {
			return true
		}
		for _, member := range union.Members {
			if member == nil {
				continue
			}
			if iface := g.nativeInterfaceInfoForGoType(member.GoType); iface != nil && g.nativeInterfaceAcceptsActualSeen(iface, actual, seen) {
				return true
			}
			if member.GoType == "runtime.ErrorValue" && g.isNativeErrorCarrierType(actual) {
				return true
			}
			if g.nativeUnionRuntimeMemberAcceptsActual(member, actual) {
				return true
			}
		}
	}
	return false
}

func (g *generator) nativeInterfaceAcceptsActualShallow(info *nativeInterfaceInfo, actual string) bool {
	if g == nil || info == nil || actual == "" {
		return false
	}
	if actual == info.GoType || actual == "runtime.Value" || actual == "any" {
		return true
	}
	if g.nativeInterfaceAssignable(actual, info.GoType) {
		return true
	}
	for _, method := range info.Methods {
		if method == nil {
			continue
		}
		if g.nativeInterfaceMethodImplExactOnly(actual, method) != nil {
			continue
		}
		if method.DefaultDefinition != nil {
			continue
		}
		return false
	}
	for _, method := range info.GenericMethods {
		if method == nil {
			continue
		}
		if g.nativeInterfaceGenericMethodImplExistsExact(actual, method) {
			continue
		}
		if method.DefaultDefinition != nil {
			continue
		}
		return false
	}
	for _, adapter := range g.nativeInterfaceKnownAdapters(info) {
		if adapter == nil || adapter.GoType == "" {
			continue
		}
		if adapter.GoType == actual || g.sameNominalStructFamily(adapter.GoType, actual) {
			return true
		}
		if union := g.nativeUnionInfoForGoType(adapter.GoType); union != nil {
			if g.nativeUnionInfoForGoType(actual) == nil {
				continue
			}
			if _, ok := g.nativeUnionMember(union, actual); ok {
				return true
			}
		}
	}
	return len(info.Methods)+len(info.GenericMethods) > 0
}

func (g *generator) nativeInterfaceWrapLines(ctx *compileContext, expected string, actual string, expr string) ([]string, string, bool) {
	info := g.nativeInterfaceInfoForGoType(expected)
	if info == nil || ctx == nil || actual == "" || expr == "" {
		return nil, "", false
	}
	if actual == expected {
		return nil, expr, true
	}
	if info.AdapterVersion != g.nativeInterfaceAdapterVersion && g.nativeInterfaceRefreshAllowed() {
		g.refreshNativeInterfaceAdapters(info)
	}
	if g.nativeInterfaceAssignable(actual, expected) {
		actualInfo := g.nativeInterfaceInfoForGoType(actual)
		if actualInfo == nil {
			return nil, expr, true
		}
		if actualInfo.AdapterVersion != g.nativeInterfaceAdapterVersion && g.nativeInterfaceRefreshAllowed() {
			g.refreshNativeInterfaceAdapters(actualInfo)
		}
		if directLines, directExpr, ok := g.nativeInterfaceDirectWrapLines(ctx, actualInfo, info, expr); ok {
			return directLines, directExpr, true
		}
		runtimeTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		convertedTemp := ctx.newTemp()
		convertErrTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", runtimeTemp, errTemp, actualInfo.ToRuntimeHelper, expr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines,
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", convertedTemp, convertErrTemp, info.FromRuntimeHelper, runtimeTemp),
			fmt.Sprintf("%s = __able_control_from_error(%s)", controlTemp, convertErrTemp),
		)
		controlLines, ok = g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convertedTemp, true
	}
	if adapter, ok := g.nativeInterfaceAdapterForActual(info, actual); ok {
		return nil, fmt.Sprintf("%s(%s)", adapter.WrapHelper, expr), true
	}
	for _, adapter := range g.nativeInterfaceKnownAdapters(info) {
		if adapter == nil || adapter.GoType == "" {
			continue
		}
		if g.sameNominalStructFamily(adapter.GoType, actual) {
			coerceLines, coercedExpr, ok := g.coerceNominalStructFamilyLines(ctx, expr, actual, adapter.GoType)
			if ok {
				return coerceLines, fmt.Sprintf("%s(%s)", adapter.WrapHelper, coercedExpr), true
			}
		}
		if g.nativeUnionInfoForGoType(adapter.GoType) == nil {
			continue
		}
		unionLines, unionExpr, ok := g.lowerWrapUnion(ctx, adapter.GoType, actual, expr)
		if !ok {
			continue
		}
		return unionLines, fmt.Sprintf("%s(%s)", adapter.WrapHelper, unionExpr), true
	}
	if actual == "any" {
		valueTemp := ctx.newTemp()
		lines := []string{fmt.Sprintf("%s := __able_any_to_value(%s)", valueTemp, expr)}
		moreLines, converted, ok := g.nativeInterfaceWrapLines(ctx, expected, "runtime.Value", valueTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, moreLines...)
		return lines, converted, true
	}
	if actual == "runtime.Value" {
		valueTemp := ctx.newTemp()
		convertedTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := %s", valueTemp, expr),
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", convertedTemp, errTemp, info.FromRuntimeHelper, valueTemp),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convertedTemp, true
	}
	if valueLines, valueExpr, ok := g.lowerRuntimeValue(ctx, expr, actual); ok {
		convertedTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := append([]string{}, valueLines...)
		lines = append(lines,
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", convertedTemp, errTemp, info.FromRuntimeHelper, valueExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		)
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convertedTemp, true
	}
	return nil, "", false
}

func (g *generator) nativeInterfaceDirectWrapLines(ctx *compileContext, actualInfo *nativeInterfaceInfo, expectedInfo *nativeInterfaceInfo, expr string) ([]string, string, bool) {
	if g == nil || ctx == nil || actualInfo == nil || expectedInfo == nil || expr == "" {
		return nil, "", false
	}
	target := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("var %s %s", target, expectedInfo.GoType),
		fmt.Sprintf("if %s == nil {", expr),
		fmt.Sprintf("\t%s = %s(nil)", target, expectedInfo.GoType),
		"} else {",
		fmt.Sprintf("\tswitch typed := %s.(type) {", expr),
	}
	var caseCount int
	for _, actualAdapter := range g.nativeInterfaceKnownAdapters(actualInfo) {
		if actualAdapter == nil || actualAdapter.AdapterType == "" || actualAdapter.GoType == "" {
			continue
		}
		expectedAdapter, ok := g.nativeInterfaceAdapterForActual(expectedInfo, actualAdapter.GoType)
		if !ok || expectedAdapter == nil || expectedAdapter.WrapHelper == "" {
			continue
		}
		lines = append(lines,
			fmt.Sprintf("\tcase %s:", actualAdapter.AdapterType),
			fmt.Sprintf("\t\t%s = %s(typed.Value)", target, expectedAdapter.WrapHelper),
		)
		caseCount++
	}
	if actualInfo.RuntimeIteratorAdapter != "" && expectedInfo.RuntimeIteratorAdapter != "" {
		lines = append(lines,
			fmt.Sprintf("\tcase %s:", actualInfo.RuntimeIteratorAdapter),
			fmt.Sprintf("\t\t%s = %s{Value: typed.Value}", target, expectedInfo.RuntimeIteratorAdapter),
		)
		caseCount++
	}
	if actualInfo.RuntimeAdapter != "" && expectedInfo.RuntimeWrapHelper != "" {
		lines = append(lines,
			fmt.Sprintf("\tcase %s:", actualInfo.RuntimeAdapter),
			fmt.Sprintf("\t\t%s = %s(typed.Value)", target, expectedInfo.RuntimeWrapHelper),
		)
		caseCount++
	}
	if caseCount == 0 {
		return nil, "", false
	}
	lines = append(lines,
		"\tdefault:",
		fmt.Sprintf("\t\t%s = %s(nil)", target, expectedInfo.GoType),
		"\t}",
		"}",
	)
	return lines, target, true
}
