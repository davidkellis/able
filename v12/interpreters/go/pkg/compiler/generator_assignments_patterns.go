package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

type patternBindingMode struct {
	declare  bool
	newNames map[string]struct{}
}

func (g *generator) patternExpectedType(pattern ast.Pattern) string {
	if pattern == nil {
		return ""
	}
	if typed, ok := pattern.(*ast.TypedPattern); ok && typed.TypeAnnotation != nil {
		if mapped, ok := g.mapTypeExpression(typed.TypeAnnotation); ok && mapped != "" && mapped != "struct{}" {
			return mapped
		}
	}
	return ""
}

func collectPatternBindingNames(pattern ast.Pattern, into map[string]struct{}) {
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p != nil && p.Name != "" && p.Name != "_" {
			into[p.Name] = struct{}{}
		}
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
				into[field.Binding.Name] = struct{}{}
			}
			fieldPattern, ok := positionalStructFieldPattern(field)
			if !ok || fieldPattern == nil {
				continue
			}
			collectPatternBindingNames(fieldPattern, into)
		}
	case *ast.ArrayPattern:
		for _, elem := range p.Elements {
			if elem == nil {
				continue
			}
			if inner, ok := elem.(ast.Pattern); ok {
				collectPatternBindingNames(inner, into)
			}
		}
		if rest := p.RestPattern; rest != nil {
			if inner, ok := rest.(ast.Pattern); ok {
				collectPatternBindingNames(inner, into)
			} else if ident, ok := rest.(*ast.Identifier); ok && ident.Name != "" && ident.Name != "_" {
				into[ident.Name] = struct{}{}
			}
		}
	case *ast.TypedPattern:
		if inner, ok := p.Pattern.(ast.Pattern); ok {
			collectPatternBindingNames(inner, into)
		}
	}
}

func (g *generator) compileAssignmentPatternBindings(ctx *compileContext, pattern ast.Pattern, subjectTemp string, subjectType string, mode patternBindingMode) ([]string, bool) {
	if pattern == nil {
		ctx.setReason("missing assignment pattern")
		return nil, false
	}
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return nil, true
	case *ast.Identifier:
		return g.bindPatternIdentifier(ctx, p.Name, subjectTemp, subjectType, nil, mode)
	case *ast.LiteralPattern:
		return nil, true
	case *ast.TypedPattern:
		if p.TypeAnnotation == nil {
			ctx.setReason("missing typed pattern annotation")
			return nil, false
		}
		if g.nativeUnionInfoForGoType(subjectType) != nil {
			return g.compileNativeUnionTypedAssignmentPatternBindings(ctx, subjectTemp, subjectType, p, mode)
		}
		if subjectType != "runtime.Value" && subjectType != "any" {
			mapped, ok := g.lowerCarrierType(ctx, p.TypeAnnotation)
			if !ok || mapped == "" || mapped == "struct{}" {
				ctx.setReason("unsupported typed pattern")
				return nil, false
			}
			if mapped != subjectType {
				return nil, true
			}
			return g.compileAssignmentPatternBindings(ctx, p.Pattern, subjectTemp, subjectType, mode)
		}
		var lines []string
		castSubject := subjectTemp
		if subjectType == "any" {
			convTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, subjectTemp))
			castSubject = convTemp
		}
		if g.runtimeSubjectDirectTypedPatternMatch(ctx, p.TypeAnnotation) {
			previousExpected := ctx.expectedTypeExpr
			ctx.expectedTypeExpr = p.TypeAnnotation
			bindLines, ok := g.compileAssignmentPatternBindings(ctx, p.Pattern, castSubject, "runtime.Value", mode)
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
		dynamicLines, narrowedTemp, narrowedType, _, ok := g.compileDynamicTypedPatternCast(ctx, castSubject, "runtime.Value", p.TypeAnnotation)
		if !ok {
			ctx.setReason("unsupported typed pattern")
			return nil, false
		}
		bindLines, ok := g.compileAssignmentPatternBindings(ctx, p.Pattern, narrowedTemp, narrowedType, mode)
		if !ok {
			return nil, false
		}
		if len(bindLines) == 0 {
			return nil, true
		}
		lines = append(lines, dynamicLines...)
		lines = append(lines, bindLines...)
		return lines, true
	case *ast.StructPattern:
		if subjectType == "runtime.Value" || subjectType == "any" {
			if subjectType == "any" {
				convTemp := ctx.newTemp()
				bindLines, ok := g.compileRuntimeStructPatternAssignmentBindings(ctx, p, convTemp, mode)
				if !ok {
					return nil, false
				}
				if len(bindLines) == 0 {
					return nil, true
				}
				return append([]string{fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, subjectTemp)}, bindLines...), true
			}
			return g.compileRuntimeStructPatternAssignmentBindings(ctx, p, subjectTemp, mode)
		}
		info := g.structInfoByGoName(subjectType)
		if info == nil {
			ctx.setReason("unsupported struct pattern")
			return nil, false
		}
		if p.StructType != nil && p.StructType.Name != "" && info.Name != p.StructType.Name {
			return nil, true
		}
		if effectiveStructPatternPositional(p, info) {
			if info.Kind != ast.StructKindPositional {
				return nil, true
			}
			if len(p.Fields) != len(info.Fields) {
				return nil, true
			}
			lines := []string{}
			for idx, field := range p.Fields {
				fieldPattern, ok := positionalStructFieldPattern(field)
				if !ok {
					ctx.setReason("invalid struct pattern field")
					return nil, false
				}
				fieldInfo := info.Fields[idx]
				fieldExpr := fmt.Sprintf("%s.%s", subjectTemp, fieldInfo.GoName)
				fieldLines, ok := g.compileAssignmentPatternBindings(ctx, fieldPattern, fieldExpr, fieldInfo.GoType, mode)
				if !ok {
					return nil, false
				}
				lines = append(lines, fieldLines...)
				if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
					bindTypeExpr := g.lowerNormalizedTypeExpr(ctx, fieldInfo.TypeExpr)
					bindLines, ok := g.bindPatternIdentifier(ctx, field.Binding.Name, fieldExpr, fieldInfo.GoType, bindTypeExpr, mode)
					if !ok {
						return nil, false
					}
					lines = append(lines, bindLines...)
				}
			}
			return lines, true
		}
		if info.Kind == ast.StructKindPositional {
			return nil, true
		}
		lines := []string{}
		for _, field := range p.Fields {
			fieldPattern, ok := positionalStructFieldPattern(field)
			if !ok {
				ctx.setReason("invalid struct pattern field")
				return nil, false
			}
			if field.FieldName == nil || field.FieldName.Name == "" {
				ctx.setReason("struct pattern missing field name")
				return nil, false
			}
			fieldInfo := g.fieldInfo(info, field.FieldName.Name)
			if fieldInfo == nil {
				ctx.setReason("unknown struct field")
				return nil, false
			}
			fieldExpr := fmt.Sprintf("%s.%s", subjectTemp, fieldInfo.GoName)
			fieldLines, ok := g.compileAssignmentPatternBindings(ctx, fieldPattern, fieldExpr, fieldInfo.GoType, mode)
			if !ok {
				return nil, false
			}
			lines = append(lines, fieldLines...)
			if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
				bindTypeExpr := g.lowerNormalizedTypeExpr(ctx, fieldInfo.TypeExpr)
				bindLines, ok := g.bindPatternIdentifier(ctx, field.Binding.Name, fieldExpr, fieldInfo.GoType, bindTypeExpr, mode)
				if !ok {
					return nil, false
				}
				lines = append(lines, bindLines...)
			}
		}
		return lines, true
	case *ast.ArrayPattern:
		if subjectType == "runtime.Value" || subjectType == "any" {
			if recoveredLines, recoveredTemp, recoveredType, recovered := g.recoverStaticArrayPatternSubject(ctx, subjectTemp, subjectType); recovered {
				bindLines, ok := g.compileNativeArrayPatternAssignmentBindings(ctx, p, recoveredTemp, recoveredType, mode)
				if !ok {
					return nil, false
				}
				return append(recoveredLines, bindLines...), true
			}
		}
		if subjectType == "runtime.Value" {
			return g.compileRuntimeArrayPatternAssignmentBindings(ctx, p, subjectTemp, mode)
		}
		if subjectType == "any" {
			convertedTemp := ctx.newTemp()
			lines := []string{fmt.Sprintf("%s := __able_any_to_value(%s)", convertedTemp, subjectTemp)}
			bindLines, ok := g.compileRuntimeArrayPatternAssignmentBindings(ctx, p, convertedTemp, mode)
			if !ok {
				return nil, false
			}
			return append(lines, bindLines...), true
		}
		if !g.isStaticArrayType(subjectType) {
			ctx.setReason("array pattern unsupported")
			return nil, false
		}
		return g.compileNativeArrayPatternAssignmentBindings(ctx, p, subjectTemp, subjectType, mode)
	default:
		ctx.setReason("unsupported assignment pattern")
		return nil, false
	}
}

func (g *generator) bindPatternIdentifier(ctx *compileContext, name string, expr string, goType string, typeExpr ast.TypeExpression, mode patternBindingMode) ([]string, bool) {
	if name == "" || name == "_" {
		return nil, true
	}
	lines := []string{}
	if (goType == "runtime.Value" || goType == "any") && typeExpr == nil && g != nil && ctx != nil && ctx.expectedTypeExpr != nil {
		typeExpr = g.lowerNormalizedTypeExpr(ctx, ctx.expectedTypeExpr)
	}
	if typeExpr == nil && g != nil {
		typeExpr, _ = g.typeExprForGoType(goType)
		typeExpr = g.lowerNormalizedTypeExpr(ctx, typeExpr)
	}
	if (goType == "runtime.Value" || goType == "any") && typeExpr != nil && g != nil && ctx != nil {
		if recoveredType, ok := g.joinCarrierTypeFromTypeExpr(ctx, typeExpr); ok && recoveredType != "" && recoveredType != "runtime.Value" && recoveredType != "any" {
			convLines, converted, convertedType, ok := g.lowerCoerceExpectedStaticExpr(ctx, nil, expr, goType, recoveredType)
			if ok {
				lines = append(lines, convLines...)
				expr = converted
				goType = convertedType
			}
		}
	}
	if mode.declare {
		if _, ok := mode.newNames[name]; ok {
			goName := sanitizeIdent(name)
			ctx.setLocalBinding(name, paramInfo{Name: name, GoName: goName, GoType: goType, TypeExpr: typeExpr})
			lines = append(lines, []string{
				fmt.Sprintf("var %s %s = %s", goName, goType, expr),
				fmt.Sprintf("_ = %s", goName),
			}...)
			return lines, true
		}
		existing, exists := ctx.lookup(name)
		if !exists {
			goName := sanitizeIdent(name)
			ctx.setLocalBinding(name, paramInfo{Name: name, GoName: goName, GoType: goType, TypeExpr: typeExpr})
			lines = append(lines, []string{
				fmt.Sprintf("var %s %s = %s", goName, goType, expr),
				fmt.Sprintf("_ = %s", goName),
			}...)
			return lines, true
		}
		if !g.typeMatches(existing.GoType, goType) {
			if existing.GoType == "runtime.Value" {
				convLines, converted, ok := g.lowerRuntimeValue(ctx, expr, goType)
				if !ok {
					ctx.setReason("pattern assignment type mismatch")
					return nil, false
				}
				convLines = append(convLines, fmt.Sprintf("%s = %s", existing.GoName, converted))
				return append(lines, convLines...), true
			}
			if goType == "runtime.Value" {
				convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, expr, existing.GoType)
				if !ok {
					ctx.setReason("pattern assignment type mismatch")
					return nil, false
				}
				convLines = append(convLines, fmt.Sprintf("%s = %s", existing.GoName, converted))
				return append(lines, convLines...), true
			}
			ctx.setReason("pattern assignment type mismatch")
			return nil, false
		}
		lines = append(lines, fmt.Sprintf("%s = %s", existing.GoName, expr))
		return lines, true
	}
	existing, exists := ctx.lookup(name)
	if exists {
		if !g.typeMatches(existing.GoType, goType) {
			if existing.GoType == "runtime.Value" {
				convLines, converted, ok := g.lowerRuntimeValue(ctx, expr, goType)
				if !ok {
					ctx.setReason("pattern assignment type mismatch")
					return nil, false
				}
				convLines = append(convLines, fmt.Sprintf("%s = %s", existing.GoName, converted))
				return append(lines, convLines...), true
			}
			if goType == "runtime.Value" {
				convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, expr, existing.GoType)
				if !ok {
					ctx.setReason("pattern assignment type mismatch")
					return nil, false
				}
				convLines = append(convLines, fmt.Sprintf("%s = %s", existing.GoName, converted))
				return append(lines, convLines...), true
			}
			ctx.setReason("pattern assignment type mismatch")
			return nil, false
		}
		lines = append(lines, fmt.Sprintf("%s = %s", existing.GoName, expr))
		return lines, true
	}
	goName := sanitizeIdent(name)
	ctx.setLocalBinding(name, paramInfo{Name: name, GoName: goName, GoType: goType, TypeExpr: typeExpr})
	lines = append(lines, []string{
		fmt.Sprintf("var %s %s = %s", goName, goType, expr),
		fmt.Sprintf("_ = %s", goName),
	}...)
	return lines, true
}

func (g *generator) compileRuntimeStructPatternAssignmentBindings(ctx *compileContext, pattern *ast.StructPattern, subjectTemp string, mode patternBindingMode) ([]string, bool) {
	if pattern == nil {
		ctx.setReason("missing struct pattern")
		return nil, false
	}
	instTemp := ctx.newTemp()
	positionalTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := __able_struct_instance(%s)", instTemp, subjectTemp),
		fmt.Sprintf("%s := %s.Positional", positionalTemp, instTemp),
	}
	for idx, field := range pattern.Fields {
		fieldPattern, ok := positionalStructFieldPattern(field)
		if !ok {
			ctx.setReason("invalid struct pattern field")
			return nil, false
		}
		fieldTemp := ctx.newTemp()
		if field.FieldName != nil && field.FieldName.Name != "" {
			lines = append(lines, fmt.Sprintf("var %s runtime.Value", fieldTemp))
			lines = append(lines, fmt.Sprintf("if %s != nil { %s = %s[%d] } else { %s = %s.Fields[%q] }", positionalTemp, fieldTemp, positionalTemp, idx, fieldTemp, instTemp, field.FieldName.Name))
		} else {
			lines = append(lines, fmt.Sprintf("var %s runtime.Value", fieldTemp))
			lines = append(lines, fmt.Sprintf("if %s != nil { %s = %s[%d] } else { %s = runtime.NilValue{} }", positionalTemp, fieldTemp, positionalTemp, idx, fieldTemp))
		}
		lines = append(lines, fmt.Sprintf("_ = %s", fieldTemp))
		fieldExpr := fieldTemp
		fieldLines, ok := g.compileAssignmentPatternBindings(ctx, fieldPattern, fieldExpr, "runtime.Value", mode)
		if !ok {
			return nil, false
		}
		lines = append(lines, fieldLines...)
		if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
			bindLines, ok := g.bindPatternIdentifier(ctx, field.Binding.Name, fieldExpr, "runtime.Value", nil, mode)
			if !ok {
				return nil, false
			}
			lines = append(lines, bindLines...)
		}
	}
	return lines, true
}

func (g *generator) compileRuntimeArrayPatternAssignmentBindings(ctx *compileContext, pattern *ast.ArrayPattern, subjectTemp string, mode patternBindingMode) ([]string, bool) {
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
	}
	for idx, elem := range pattern.Elements {
		if elem == nil {
			ctx.setReason("invalid array pattern element")
			return nil, false
		}
		elemExpr := fmt.Sprintf("%s[%d]", valuesTemp, idx)
		elemLines, ok := g.compileAssignmentPatternBindings(ctx, elem, elemExpr, "runtime.Value", mode)
		if !ok {
			return nil, false
		}
		lines = append(lines, elemLines...)
	}
	if pattern.RestPattern != nil {
		switch rest := pattern.RestPattern.(type) {
		case *ast.Identifier:
			restLines, restExpr, restTypeExpr, ok := g.runtimeArrayRestCarrierLines(ctx, fmt.Sprintf("%s[%d:]", valuesTemp, len(pattern.Elements)))
			if !ok {
				ctx.setReason("array pattern unsupported")
				return nil, false
			}
			lines = append(lines, restLines...)
			bindLines, ok := g.bindPatternIdentifier(ctx, rest.Name, restExpr, "*Array", restTypeExpr, mode)
			if !ok {
				return nil, false
			}
			lines = append(lines, bindLines...)
		case *ast.WildcardPattern:
		}
	}
	return lines, true
}

func (g *generator) compileNativeArrayPatternAssignmentBindings(ctx *compileContext, pattern *ast.ArrayPattern, subjectTemp string, subjectType string, mode patternBindingMode) ([]string, bool) {
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
	}
	for idx, elem := range pattern.Elements {
		if elem == nil {
			ctx.setReason("invalid array pattern element")
			return nil, false
		}
		elemExpr := fmt.Sprintf("%s[%d]", valuesTemp, idx)
		elemLines, ok := g.compileAssignmentPatternBindings(ctx, elem, elemExpr, elemType, mode)
		if !ok {
			return nil, false
		}
		lines = append(lines, elemLines...)
	}
	if pattern.RestPattern != nil {
		switch rest := pattern.RestPattern.(type) {
		case *ast.Identifier:
			restLines, restExpr, ok := g.nativeArrayFromElementsLines(ctx, subjectType, fmt.Sprintf("%s[%d:]", valuesTemp, len(pattern.Elements)))
			if !ok {
				ctx.setReason("array pattern unsupported")
				return nil, false
			}
			lines = append(lines, restLines...)
			bindLines, ok := g.bindPatternIdentifier(ctx, rest.Name, restExpr, subjectType, nil, mode)
			if !ok {
				return nil, false
			}
			lines = append(lines, bindLines...)
		case *ast.WildcardPattern:
		}
	}
	return lines, true
}
