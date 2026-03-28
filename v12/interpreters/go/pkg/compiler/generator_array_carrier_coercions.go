package compiler

import "fmt"

func (g *generator) staticArrayCarrierCoercible(expected, actual string) bool {
	if g == nil || expected == "" || actual == "" {
		return false
	}
	if expected == actual {
		return true
	}
	if !g.isStaticArrayType(expected) || !g.isStaticArrayType(actual) {
		return false
	}
	return expected == "*Array" || actual == "*Array"
}

func (g *generator) staticArrayToRuntimeLines(ctx *compileContext, expr string, actual string) ([]string, string, bool) {
	if g == nil || ctx == nil || expr == "" || actual == "" {
		return nil, "", false
	}
	valueTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	var lines []string
	switch {
	case actual == "*Array":
		lines = []string{fmt.Sprintf("%s, %s := __able_struct_Array_to(__able_runtime, %s)", valueTemp, errTemp, expr)}
	case g.isMonoArrayType(actual):
		spec, ok := g.monoArraySpecForGoType(actual)
		if !ok || spec == nil {
			return nil, "", false
		}
		lines = []string{fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", valueTemp, errTemp, spec.ToRuntimeHelper, expr)}
	default:
		return nil, "", false
	}
	lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", false
	}
	lines = append(lines, controlLines...)
	return lines, valueTemp, true
}

func (g *generator) staticArrayFromRuntimeLines(ctx *compileContext, valueExpr string, expected string) ([]string, string, bool) {
	if g == nil || ctx == nil || valueExpr == "" || expected == "" {
		return nil, "", false
	}
	resultTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	var lines []string
	switch {
	case expected == "*Array":
		lines = []string{fmt.Sprintf("%s, %s := __able_struct_Array_from(%s)", resultTemp, errTemp, valueExpr)}
	case g.isMonoArrayType(expected):
		spec, ok := g.monoArraySpecForGoType(expected)
		if !ok || spec == nil {
			return nil, "", false
		}
		lines = []string{fmt.Sprintf("%s, %s := %s(%s)", resultTemp, errTemp, spec.FromRuntimeHelper, valueExpr)}
	default:
		return nil, "", false
	}
	lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", false
	}
	lines = append(lines, controlLines...)
	return lines, resultTemp, true
}

func (g *generator) coerceStaticArrayCarrierLines(ctx *compileContext, expr string, actual string, expected string) ([]string, string, bool) {
	if g == nil || ctx == nil || expr == "" || !g.staticArrayCarrierCoercible(expected, actual) {
		return nil, "", false
	}
	if actual == expected {
		return nil, expr, true
	}
	toRuntimeLines, runtimeExpr, ok := g.staticArrayToRuntimeLines(ctx, expr, actual)
	if !ok {
		return nil, "", false
	}
	fromRuntimeLines, converted, ok := g.staticArrayFromRuntimeLines(ctx, runtimeExpr, expected)
	if !ok {
		return nil, "", false
	}
	lines := append(toRuntimeLines, fromRuntimeLines...)
	return lines, converted, true
}
