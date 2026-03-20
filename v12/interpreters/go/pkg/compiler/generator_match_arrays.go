package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileRuntimeArrayPatternCondition(ctx *compileContext, pattern *ast.ArrayPattern, subjectTemp string) ([]string, string, bool) {
	if pattern == nil {
		ctx.setReason("missing array pattern")
		return nil, "", false
	}
	if pattern.RestPattern != nil {
		switch pattern.RestPattern.(type) {
		case *ast.Identifier, *ast.WildcardPattern:
		default:
			ctx.setReason("unsupported rest pattern")
			return nil, "", false
		}
	}
	condTemp := ctx.newTemp()
	condLabel := ctx.newTemp()
	valuesTemp := ctx.newTemp()
	okTemp := ctx.newTemp()
	inner := []string{
		fmt.Sprintf("%s, %s := __able_array_values(%s)", valuesTemp, okTemp, subjectTemp),
		fmt.Sprintf("if !%s { break %s }", okTemp, condLabel),
	}
	if pattern.RestPattern == nil {
		inner = append(inner, fmt.Sprintf("if len(%s) != %d { break %s }", valuesTemp, len(pattern.Elements), condLabel))
	} else {
		inner = append(inner, fmt.Sprintf("if len(%s) < %d { break %s }", valuesTemp, len(pattern.Elements), condLabel))
	}
	for idx, elem := range pattern.Elements {
		if elem == nil {
			ctx.setReason("invalid array pattern element")
			return nil, "", false
		}
		elemExpr := fmt.Sprintf("%s[%d]", valuesTemp, idx)
		elemCondLines, elemCond, ok := g.compileMatchPatternCondition(ctx, elem, elemExpr, "runtime.Value")
		if !ok {
			return nil, "", false
		}
		if len(elemCondLines) > 0 {
			inner = append(inner, elemCondLines...)
		}
		if elemCond != "true" {
			inner = append(inner, fmt.Sprintf("if !(%s) { break %s }", elemCond, condLabel))
		}
	}
	inner = append(inner, fmt.Sprintf("%s = true", condTemp))

	lines := []string{
		fmt.Sprintf("%s := false", condTemp),
		fmt.Sprintf("%s: switch { default: %s }", condLabel, strings.Join(inner, "; ")),
	}
	return lines, condTemp, true
}

func (g *generator) compileNativeArrayPatternCondition(ctx *compileContext, pattern *ast.ArrayPattern, subjectTemp string, subjectType string) ([]string, string, bool) {
	if pattern == nil {
		ctx.setReason("missing array pattern")
		return nil, "", false
	}
	if pattern.RestPattern != nil {
		switch pattern.RestPattern.(type) {
		case *ast.Identifier, *ast.WildcardPattern:
		default:
			ctx.setReason("unsupported rest pattern")
			return nil, "", false
		}
	}
	valuesExpr, ok := g.nativeArrayValuesExpr(subjectTemp, subjectType)
	if !ok {
		ctx.setReason("array pattern unsupported")
		return nil, "", false
	}
	elemType := g.staticArrayElemGoType(subjectType)
	if elemType == "" {
		ctx.setReason("array pattern unsupported")
		return nil, "", false
	}
	condTemp := ctx.newTemp()
	condLabel := ctx.newTemp()
	valuesTemp := ctx.newTemp()
	inner := make([]string, 0, len(pattern.Elements)+4)
	if strings.HasPrefix(subjectType, "*") {
		inner = append(inner, fmt.Sprintf("if %s == nil { break %s }", subjectTemp, condLabel))
	}
	inner = append(inner, fmt.Sprintf("%s := %s", valuesTemp, valuesExpr))
	if pattern.RestPattern == nil {
		inner = append(inner, fmt.Sprintf("if len(%s) != %d { break %s }", valuesTemp, len(pattern.Elements), condLabel))
	} else {
		inner = append(inner, fmt.Sprintf("if len(%s) < %d { break %s }", valuesTemp, len(pattern.Elements), condLabel))
	}
	for idx, elem := range pattern.Elements {
		if elem == nil {
			ctx.setReason("invalid array pattern element")
			return nil, "", false
		}
		elemExpr := fmt.Sprintf("%s[%d]", valuesTemp, idx)
		elemCondLines, elemCond, ok := g.compileMatchPatternCondition(ctx, elem, elemExpr, elemType)
		if !ok {
			return nil, "", false
		}
		if len(elemCondLines) > 0 {
			inner = append(inner, elemCondLines...)
		}
		if elemCond != "true" {
			inner = append(inner, fmt.Sprintf("if !(%s) { break %s }", elemCond, condLabel))
		}
	}
	inner = append(inner, fmt.Sprintf("%s = true", condTemp))
	lines := []string{
		fmt.Sprintf("%s := false", condTemp),
		fmt.Sprintf("%s: switch { default: %s }", condLabel, strings.Join(inner, "; ")),
	}
	return lines, condTemp, true
}

func (g *generator) compileRuntimeArrayPatternBindings(ctx *compileContext, pattern *ast.ArrayPattern, subjectTemp string) ([]string, bool) {
	if pattern == nil {
		ctx.setReason("missing array pattern")
		return nil, false
	}
	if pattern.RestPattern != nil {
		switch pattern.RestPattern.(type) {
		case *ast.Identifier, *ast.WildcardPattern:
		default:
			ctx.setReason("unsupported rest pattern")
			return nil, false
		}
	}
	valuesTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s, _ := __able_array_values(%s)", valuesTemp, subjectTemp),
		fmt.Sprintf("_ = %s", valuesTemp),
	}
	for idx, elem := range pattern.Elements {
		if elem == nil {
			ctx.setReason("invalid array pattern element")
			return nil, false
		}
		elemExpr := fmt.Sprintf("%s[%d]", valuesTemp, idx)
		elemLines, ok := g.compileMatchPatternBindings(ctx, elem, elemExpr, "runtime.Value")
		if !ok {
			return nil, false
		}
		lines = append(lines, elemLines...)
	}
	if pattern.RestPattern != nil {
		switch rest := pattern.RestPattern.(type) {
		case *ast.Identifier:
			if rest.Name != "" && rest.Name != "_" {
				goName := sanitizeIdent(rest.Name)
				ctx.locals[rest.Name] = paramInfo{Name: rest.Name, GoName: goName, GoType: "runtime.Value"}
				lines = append(lines,
					fmt.Sprintf("var %s runtime.Value = &runtime.ArrayValue{Elements: append([]runtime.Value(nil), %s[%d:]...)}", goName, valuesTemp, len(pattern.Elements)),
					fmt.Sprintf("_ = %s", goName),
				)
			}
		case *ast.WildcardPattern:
		}
	}
	return lines, true
}

func (g *generator) compileNativeArrayPatternBindings(ctx *compileContext, pattern *ast.ArrayPattern, subjectTemp string, subjectType string) ([]string, bool) {
	if pattern == nil {
		ctx.setReason("missing array pattern")
		return nil, false
	}
	if pattern.RestPattern != nil {
		switch pattern.RestPattern.(type) {
		case *ast.Identifier, *ast.WildcardPattern:
		default:
			ctx.setReason("unsupported rest pattern")
			return nil, false
		}
	}
	valuesExpr, ok := g.nativeArrayValuesExpr(subjectTemp, subjectType)
	if !ok {
		ctx.setReason("array pattern unsupported")
		return nil, false
	}
	elemType := g.staticArrayElemGoType(subjectType)
	if elemType == "" {
		ctx.setReason("array pattern unsupported")
		return nil, false
	}
	valuesTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", valuesTemp, valuesExpr),
		fmt.Sprintf("_ = %s", valuesTemp),
	}
	for idx, elem := range pattern.Elements {
		if elem == nil {
			ctx.setReason("invalid array pattern element")
			return nil, false
		}
		elemExpr := fmt.Sprintf("%s[%d]", valuesTemp, idx)
		elemLines, ok := g.compileMatchPatternBindings(ctx, elem, elemExpr, elemType)
		if !ok {
			return nil, false
		}
		lines = append(lines, elemLines...)
	}
	if pattern.RestPattern != nil {
		switch rest := pattern.RestPattern.(type) {
		case *ast.Identifier:
			if rest.Name != "" && rest.Name != "_" {
				restLines, restExpr, ok := g.nativeArrayFromElementsLines(ctx, subjectType, fmt.Sprintf("%s[%d:]", valuesTemp, len(pattern.Elements)))
				if !ok {
					ctx.setReason("array pattern unsupported")
					return nil, false
				}
				lines = append(lines, restLines...)
				goName := sanitizeIdent(rest.Name)
				ctx.locals[rest.Name] = paramInfo{Name: rest.Name, GoName: goName, GoType: subjectType}
				lines = append(lines,
					fmt.Sprintf("var %s %s = %s", goName, subjectType, restExpr),
					fmt.Sprintf("_ = %s", goName),
				)
			}
		case *ast.WildcardPattern:
		}
	}
	return lines, true
}
