package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) recoverTypedPatternCarrier(ctx *compileContext, expr ast.TypeExpression) (string, bool) {
	if g == nil || expr == nil {
		return "", false
	}
	mapped, ok := g.lowerCarrierType(ctx, g.lowerNormalizedTypeExpr(ctx, expr))
	if !ok || mapped == "" || mapped == "struct{}" || mapped == "runtime.Value" || mapped == "any" {
		return "", false
	}
	return mapped, true
}

func (g *generator) runtimeSubjectDirectTypedPatternMatch(ctx *compileContext, expr ast.TypeExpression) bool {
	if g == nil || ctx == nil || expr == nil || ctx.matchSubjectTypeExpr == nil {
		return false
	}
	subjectExpr := g.lowerNormalizedTypeExpr(ctx, ctx.matchSubjectTypeExpr)
	patternExpr := g.lowerNormalizedTypeExpr(ctx, expr)
	if subjectExpr == nil || patternExpr == nil {
		return false
	}
	subjectKey := normalizeTypeExprIdentityKey(g, ctx.packageName, subjectExpr)
	patternKey := normalizeTypeExprIdentityKey(g, ctx.packageName, patternExpr)
	return subjectKey != "" && subjectKey == patternKey
}

func (g *generator) directRuntimeTypeCheckedTypedPatternCarrier(goType string) bool {
	if g == nil || goType == "" {
		return false
	}
	return goType == "bool" || goType == "string" || goType == "rune" || g.isNumericType(goType)
}

func (g *generator) compileDirectRuntimeTypeCheckedDynamicTypedPatternCast(ctx *compileContext, runtimeExpr string, narrowedType string) ([]string, string, string, string, bool) {
	if g == nil || ctx == nil || runtimeExpr == "" || narrowedType == "" || !g.directRuntimeTypeCheckedTypedPatternCarrier(narrowedType) {
		return nil, "", "", "", false
	}
	checkLines, checkCond, ok := g.runtimeTypeCheckExpr(ctx, runtimeExpr, narrowedType)
	if !ok {
		return nil, "", "", "", false
	}
	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx.probeChild(), runtimeExpr, narrowedType)
	if !ok {
		return nil, "", "", "", false
	}
	narrowedTemp := ctx.newTemp()
	lines := append([]string{}, checkLines...)
	lines = append(lines, fmt.Sprintf("var %s %s", narrowedTemp, narrowedType))
	lines = append(lines, fmt.Sprintf("if %s {", checkCond))
	lines = append(lines, indentLines(convLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", narrowedTemp, converted))
	lines = append(lines, "}")
	return lines, narrowedTemp, narrowedType, checkCond, true
}

func (g *generator) compileDirectRuntimeTypeCheckedDynamicTypedPatternCondition(ctx *compileContext, runtimeExpr string, narrowedType string) ([]string, string, bool) {
	if g == nil || ctx == nil || runtimeExpr == "" || narrowedType == "" || !g.directRuntimeTypeCheckedTypedPatternCarrier(narrowedType) {
		return nil, "", false
	}
	return g.runtimeTypeCheckExpr(ctx, runtimeExpr, narrowedType)
}

func (g *generator) directDynamicStructTypedPatternInfo(goType string) *structInfo {
	if g == nil || !g.isNativeStructPointerType(goType) {
		return nil
	}
	return g.structInfoByGoName(goType)
}

func (g *generator) compileDirectDynamicTypedPatternCast(ctx *compileContext, runtimeExpr string, expr ast.TypeExpression) ([]string, string, string, string, bool) {
	if g == nil || ctx == nil || runtimeExpr == "" || expr == nil {
		return nil, "", "", "", false
	}
	normalizedExpr := g.lowerNormalizedTypeExpr(ctx, expr)
	if nullableExpr, ok := normalizedExpr.(*ast.NullableTypeExpression); ok && nullableExpr != nil {
		narrowedType, ok := g.recoverTypedPatternCarrier(ctx, nullableExpr)
		if !ok || narrowedType == "" {
			return nil, "", "", "", false
		}
		if helper, ok := g.nativeNullableFromRuntimeHelper(narrowedType); ok {
			narrowedTemp := ctx.newTemp()
			okTemp := ctx.newTemp()
			errTemp := ctx.newTemp()
			controlTemp := ctx.newTemp()
			lines := []string{
				fmt.Sprintf("var %s %s", narrowedTemp, narrowedType),
				fmt.Sprintf("var %s bool", okTemp),
				fmt.Sprintf("var %s error", errTemp),
				fmt.Sprintf("%s, %s = %s(%s)", narrowedTemp, errTemp, helper, runtimeExpr),
				fmt.Sprintf("%s = %s == nil", okTemp, errTemp),
				fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
			}
			controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
			if !ok {
				return nil, "", "", "", false
			}
			lines = append(lines, controlLines...)
			lines = append(lines, fmt.Sprintf("_ = %s", okTemp))
			return lines, narrowedTemp, narrowedType, okTemp, true
		}
		return nil, "", "", "", false
	}
	narrowedType, ok := g.recoverTypedPatternCarrier(ctx, expr)
	if !ok || narrowedType == "" {
		return nil, "", "", "", false
	}
	if union := g.nativeUnionInfoForGoType(narrowedType); union != nil {
		narrowedTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("var %s %s", narrowedTemp, narrowedType),
			fmt.Sprintf("var %s bool", okTemp),
			fmt.Sprintf("var %s error", errTemp),
			fmt.Sprintf("%s, %s, %s = %s(__able_runtime, %s)", narrowedTemp, okTemp, errTemp, union.TryFromRuntimeHelper, runtimeExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", "", "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines, fmt.Sprintf("_ = %s", okTemp))
		return lines, narrowedTemp, narrowedType, okTemp, true
	}
	if info := g.directDynamicStructTypedPatternInfo(narrowedType); info != nil {
		narrowedTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("var %s %s", narrowedTemp, narrowedType),
			fmt.Sprintf("var %s bool", okTemp),
			fmt.Sprintf("var %s error", errTemp),
			fmt.Sprintf("%s, %s, %s = __able_struct_%s_try_from(%s)", narrowedTemp, okTemp, errTemp, info.GoName, runtimeExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", "", "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines, fmt.Sprintf("_ = %s", okTemp))
		return lines, narrowedTemp, narrowedType, okTemp, true
	}
	if iface := g.nativeInterfaceInfoForGoType(narrowedType); iface != nil {
		narrowedTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("var %s %s", narrowedTemp, narrowedType),
			fmt.Sprintf("var %s bool", okTemp),
			fmt.Sprintf("var %s error", errTemp),
			fmt.Sprintf("%s, %s, %s = %s(__able_runtime, %s)", narrowedTemp, okTemp, errTemp, iface.TryFromRuntimeHelper, runtimeExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", "", "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines, fmt.Sprintf("_ = %s", okTemp))
		return lines, narrowedTemp, narrowedType, okTemp, true
	}
	if callable := g.nativeCallableInfoForGoType(narrowedType); callable != nil {
		narrowedTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("var %s %s", narrowedTemp, narrowedType),
			fmt.Sprintf("var %s bool", okTemp),
			fmt.Sprintf("var %s error", errTemp),
			fmt.Sprintf("%s, %s, %s = %s(__able_runtime, %s)", narrowedTemp, okTemp, errTemp, callable.TryFromRuntimeHelper, runtimeExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", "", "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines, fmt.Sprintf("_ = %s", okTemp))
		return lines, narrowedTemp, narrowedType, okTemp, true
	}
	if lines, narrowedTemp, narrowedType, okTemp, ok := g.compileDirectRuntimeTypeCheckedDynamicTypedPatternCast(ctx, runtimeExpr, narrowedType); ok {
		return lines, narrowedTemp, narrowedType, okTemp, true
	}
	if narrowedType == "runtime.ErrorValue" {
		narrowedTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("var %s runtime.ErrorValue", narrowedTemp),
			fmt.Sprintf("%s := __able_is_error(%s)", okTemp, runtimeExpr),
			fmt.Sprintf("if %s {", okTemp),
		}
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx.probeChild(), runtimeExpr, narrowedType)
		if !ok {
			return nil, "", "", "", false
		}
		lines = append(lines, indentLines(convLines, 1)...)
		lines = append(lines, fmt.Sprintf("\t%s = %s", narrowedTemp, converted))
		lines = append(lines, "}")
		return lines, narrowedTemp, narrowedType, okTemp, true
	}
	return nil, "", "", "", false
}

func (g *generator) compileDirectDynamicTypedPatternCondition(ctx *compileContext, runtimeExpr string, expr ast.TypeExpression) ([]string, string, bool) {
	if g == nil || ctx == nil || runtimeExpr == "" || expr == nil {
		return nil, "", false
	}
	normalizedExpr := g.lowerNormalizedTypeExpr(ctx, expr)
	if nullableExpr, ok := normalizedExpr.(*ast.NullableTypeExpression); ok && nullableExpr != nil {
		narrowedType, ok := g.recoverTypedPatternCarrier(ctx, nullableExpr)
		if !ok || narrowedType == "" {
			return nil, "", false
		}
		if helper, ok := g.nativeNullableFromRuntimeHelper(narrowedType); ok {
			okTemp := ctx.newTemp()
			errTemp := ctx.newTemp()
			controlTemp := ctx.newTemp()
			lines := []string{
				fmt.Sprintf("var %s bool", okTemp),
				fmt.Sprintf("var %s error", errTemp),
				fmt.Sprintf("_, %s = %s(%s)", errTemp, helper, runtimeExpr),
				fmt.Sprintf("%s = %s == nil", okTemp, errTemp),
				fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
			}
			controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
			if !ok {
				return nil, "", false
			}
			lines = append(lines, controlLines...)
			return lines, okTemp, true
		}
		return nil, "", false
	}
	narrowedType, ok := g.recoverTypedPatternCarrier(ctx, expr)
	if !ok || narrowedType == "" {
		return nil, "", false
	}
	if union := g.nativeUnionInfoForGoType(narrowedType); union != nil {
		okTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("var %s bool", okTemp),
			fmt.Sprintf("var %s error", errTemp),
			fmt.Sprintf("_, %s, %s = %s(__able_runtime, %s)", okTemp, errTemp, union.TryFromRuntimeHelper, runtimeExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, okTemp, true
	}
	if info := g.directDynamicStructTypedPatternInfo(narrowedType); info != nil {
		okTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("var %s bool", okTemp),
			fmt.Sprintf("var %s error", errTemp),
			fmt.Sprintf("_, %s, %s = __able_struct_%s_try_from(%s)", okTemp, errTemp, info.GoName, runtimeExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, okTemp, true
	}
	if iface := g.nativeInterfaceInfoForGoType(narrowedType); iface != nil {
		okTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("var %s bool", okTemp),
			fmt.Sprintf("var %s error", errTemp),
			fmt.Sprintf("_, %s, %s = %s(__able_runtime, %s)", okTemp, errTemp, iface.TryFromRuntimeHelper, runtimeExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, okTemp, true
	}
	if callable := g.nativeCallableInfoForGoType(narrowedType); callable != nil {
		okTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("var %s bool", okTemp),
			fmt.Sprintf("var %s error", errTemp),
			fmt.Sprintf("_, %s, %s = %s(__able_runtime, %s)", okTemp, errTemp, callable.TryFromRuntimeHelper, runtimeExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, okTemp, true
	}
	if lines, okTemp, ok := g.compileDirectRuntimeTypeCheckedDynamicTypedPatternCondition(ctx, runtimeExpr, narrowedType); ok {
		return lines, okTemp, true
	}
	if narrowedType == "runtime.ErrorValue" {
		okTemp := ctx.newTemp()
		lines := []string{fmt.Sprintf("%s := __able_is_error(%s)", okTemp, runtimeExpr)}
		return lines, okTemp, true
	}
	return nil, "", false
}

func (g *generator) compileDynamicTypedPatternCast(ctx *compileContext, subjectTemp string, subjectType string, expr ast.TypeExpression) ([]string, string, string, string, bool) {
	if g == nil || ctx == nil || subjectTemp == "" || expr == nil {
		return nil, "", "", "", false
	}
	lines := []string{}
	castSubject := subjectTemp
	if subjectType == "any" {
		convTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, subjectTemp))
		castSubject = convTemp
	}
	if directLines, narrowedTemp, narrowedType, okTemp, ok := g.compileDirectDynamicTypedPatternCast(ctx, castSubject, expr); ok {
		lines = append(lines, directLines...)
		return lines, narrowedTemp, narrowedType, okTemp, true
	}
	typeExpr, ok := g.renderTypeExpression(g.lowerNormalizedTypeExpr(ctx, expr))
	if !ok {
		return nil, "", "", "", false
	}
	g.needsAst = true

	runtimeTemp := ctx.newTemp()
	okTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s, %s := __able_try_cast(%s, %s)", runtimeTemp, okTemp, controlTemp, castSubject, typeExpr))
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", "", "", false
	}
	lines = append(lines, controlLines...)

	narrowedType := "runtime.Value"
	if mapped, ok := g.recoverTypedPatternCarrier(ctx, expr); ok {
		narrowedType = mapped
	}
	narrowedTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("var %s %s", narrowedTemp, narrowedType))
	if narrowedType == "runtime.Value" {
		lines = append(lines, fmt.Sprintf("if %s { %s = %s }", okTemp, narrowedTemp, runtimeTemp))
		return lines, narrowedTemp, narrowedType, okTemp, true
	}

	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, runtimeTemp, narrowedType)
	if !ok {
		return nil, "", "", "", false
	}
	lines = append(lines, fmt.Sprintf("if %s {", okTemp))
	lines = append(lines, indentLines(convLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", narrowedTemp, converted))
	lines = append(lines, "}")
	return lines, narrowedTemp, narrowedType, okTemp, true
}
