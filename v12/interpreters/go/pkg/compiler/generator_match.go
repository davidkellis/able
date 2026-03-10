package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileMatchPattern(ctx *compileContext, pattern ast.Pattern, subjectTemp string, subjectType string) ([]string, string, []string, bool) {
	if pattern == nil {
		ctx.setReason("missing match pattern")
		return nil, "", nil, false
	}
	condLines, cond, ok := g.compileMatchPatternCondition(ctx, pattern, subjectTemp, subjectType)
	if !ok {
		return nil, "", nil, false
	}
	bindLines, ok := g.compileMatchPatternBindings(ctx, pattern, subjectTemp, subjectType)
	if !ok {
		return nil, "", nil, false
	}
	return condLines, cond, bindLines, true
}

func (g *generator) isSingletonPattern(ctx *compileContext, name string) bool {
	if name == "" {
		return false
	}
	if ctx != nil {
		if _, ok := ctx.lookup(name); ok {
			return false
		}
	}
	pkgName := ""
	if ctx != nil {
		pkgName = ctx.packageName
	}
	info, ok := g.structInfoForTypeName(pkgName, name)
	if !ok || info == nil {
		return false
	}
	return info.Kind == ast.StructKindSingleton
}

func (g *generator) compileMatchPatternCondition(ctx *compileContext, pattern ast.Pattern, subjectTemp string, subjectType string) ([]string, string, bool) {
	if pattern == nil {
		ctx.setReason("missing match pattern")
		return nil, "", false
	}
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return nil, "true", true
	case *ast.Identifier:
		if p.Name == "_" {
			return nil, "true", true
		}
		if g.isSingletonPattern(ctx, p.Name) {
			if subjectType == "runtime.Value" || subjectType == "any" {
				effectiveSubject := subjectTemp
				if subjectType == "any" {
					effectiveSubject = fmt.Sprintf("__able_any_to_value(%s)", subjectTemp)
				}
				condExpr := fmt.Sprintf("__able_match_singleton(%s, %q)", effectiveSubject, p.Name)
				return nil, condExpr, true
			}
			pkgName := ""
			if ctx != nil {
				pkgName = ctx.packageName
			}
			info, _ := g.structInfoForTypeName(pkgName, p.Name)
			baseType := subjectType
			if name, ok := g.structBaseName(subjectType); ok {
				baseType = name
			}
			if info != nil && info.GoName != "" && baseType != info.GoName {
				ctx.setReason("singleton pattern type mismatch")
				return nil, "", false
			}
		}
		return nil, "true", true
	case *ast.LiteralPattern:
		return g.compileLiteralMatch(ctx, p.Literal, subjectTemp, subjectType)
	case *ast.TypedPattern:
		if p.TypeAnnotation == nil {
			ctx.setReason("missing typed pattern annotation")
			return nil, "", false
		}
		if subjectType != "runtime.Value" && subjectType != "any" {
			mapped, ok := g.mapTypeExpressionInPackage(ctx.packageName, p.TypeAnnotation)
			if !ok || mapped == "" || mapped == "struct{}" {
				ctx.setReason("unsupported typed pattern")
				return nil, "", false
			}
			if mapped != subjectType {
				ctx.setReason("typed pattern type mismatch")
				return nil, "", false
			}
			return g.compileMatchPatternCondition(ctx, p.Pattern, subjectTemp, subjectType)
		}
		typeExpr, ok := g.renderTypeExpression(p.TypeAnnotation)
		if !ok {
			ctx.setReason("unsupported typed pattern")
			return nil, "", false
		}
		g.needsAst = true
		castTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		castSubject := subjectTemp
		var prefixLines []string
		if subjectType == "any" {
			convTemp := ctx.newTemp()
			prefixLines = append(prefixLines, fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, subjectTemp))
			castSubject = convTemp
		}
		innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, p.Pattern, castTemp, "runtime.Value")
		if !ok {
			return nil, "", false
		}
		if innerCond == "true" && len(innerCondLines) == 0 {
			lines := append(prefixLines, fmt.Sprintf("_, %s := __able_try_cast(%s, %s)", okTemp, castSubject, typeExpr))
			return lines, okTemp, true
		}
		condTemp := ctx.newTemp()
		lines := append(prefixLines,
			fmt.Sprintf("%s, %s := __able_try_cast(%s, %s)", castTemp, okTemp, castSubject, typeExpr),
			fmt.Sprintf("var %s bool", condTemp),
		)
		if len(innerCondLines) == 0 {
			lines = append(lines, fmt.Sprintf("if %s { %s = %s }", okTemp, condTemp, innerCond))
		} else {
			lines = append(lines, fmt.Sprintf("if %s {", okTemp))
			lines = append(lines, indentLines(innerCondLines, 1)...)
			lines = append(lines, fmt.Sprintf("\t%s = %s", condTemp, innerCond))
			lines = append(lines, "}")
		}
		return lines, condTemp, true
	case *ast.StructPattern:
		if subjectType == "runtime.Value" || subjectType == "any" {
			effectiveSubject := subjectTemp
			if subjectType == "any" {
				convTemp := ctx.newTemp()
				convertLine := fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, subjectTemp)
				condLines, condExpr, ok := g.compileRuntimeStructPatternCondition(ctx, p, convTemp)
				if !ok {
					return nil, "", false
				}
				return append([]string{convertLine}, condLines...), condExpr, true
			}
			return g.compileRuntimeStructPatternCondition(ctx, p, effectiveSubject)
		}
		info := g.structInfoByGoName(subjectType)
		if info == nil {
			ctx.setReason("unsupported struct pattern")
			return nil, "", false
		}
		if p.StructType != nil && p.StructType.Name != "" && info.Name != p.StructType.Name {
			ctx.setReason("struct pattern type mismatch")
			return nil, "", false
		}
		if p.IsPositional {
			if info.Kind != ast.StructKindPositional {
				ctx.setReason("struct pattern positional mismatch")
				return nil, "", false
			}
			if len(p.Fields) != len(info.Fields) {
				ctx.setReason("struct pattern arity mismatch")
				return nil, "", false
			}
			var allLines []string
			conds := make([]string, 0, len(p.Fields))
			for idx, field := range p.Fields {
				pattern, ok := positionalStructFieldPattern(field)
				if !ok {
					ctx.setReason("invalid struct pattern field")
					return nil, "", false
				}
				fieldInfo := info.Fields[idx]
				fieldExpr := fmt.Sprintf("%s.%s", subjectTemp, fieldInfo.GoName)
				fieldCondLines, fieldCond, ok := g.compileMatchPatternCondition(ctx, pattern, fieldExpr, fieldInfo.GoType)
				if !ok {
					return nil, "", false
				}
				if len(fieldCondLines) > 0 {
					allLines = append(allLines, fieldCondLines...)
					if fieldCond != "true" {
						temp := ctx.newTemp()
						allLines = append(allLines, fmt.Sprintf("%s := %s", temp, fieldCond))
						conds = append(conds, temp)
					}
				} else if fieldCond != "true" {
					conds = append(conds, fieldCond)
				}
			}
			if len(conds) == 0 {
				return allLines, "true", true
			}
			return allLines, strings.Join(conds, " && "), true
		}
		if info.Kind == ast.StructKindPositional {
			ctx.setReason("struct pattern positional mismatch")
			return nil, "", false
		}
		var allLines []string
		conds := make([]string, 0, len(p.Fields))
		for _, field := range p.Fields {
			if field == nil || field.Pattern == nil {
				ctx.setReason("invalid struct pattern field")
				return nil, "", false
			}
			if field.FieldName == nil || field.FieldName.Name == "" {
				ctx.setReason("struct pattern missing field name")
				return nil, "", false
			}
			fieldInfo := g.fieldInfo(info, field.FieldName.Name)
			if fieldInfo == nil {
				ctx.setReason("unknown struct field")
				return nil, "", false
			}
			fieldExpr := fmt.Sprintf("%s.%s", subjectTemp, fieldInfo.GoName)
			fieldCondLines, fieldCond, ok := g.compileMatchPatternCondition(ctx, field.Pattern, fieldExpr, fieldInfo.GoType)
			if !ok {
				return nil, "", false
			}
			if len(fieldCondLines) > 0 {
				allLines = append(allLines, fieldCondLines...)
				if fieldCond != "true" {
					temp := ctx.newTemp()
					allLines = append(allLines, fmt.Sprintf("%s := %s", temp, fieldCond))
					conds = append(conds, temp)
				}
			} else if fieldCond != "true" {
				conds = append(conds, fieldCond)
			}
		}
		if len(conds) == 0 {
			return allLines, "true", true
		}
		return allLines, strings.Join(conds, " && "), true
	case *ast.ArrayPattern:
		if subjectType != "runtime.Value" {
			ctx.setReason("array pattern unsupported")
			return nil, "", false
		}
		return g.compileRuntimeArrayPatternCondition(ctx, p, subjectTemp)
	default:
		ctx.setReason("unsupported match pattern")
		return nil, "", false
	}
}

func (g *generator) compileMatchPatternBindings(ctx *compileContext, pattern ast.Pattern, subjectTemp string, subjectType string) ([]string, bool) {
	if pattern == nil {
		ctx.setReason("missing match pattern")
		return nil, false
	}
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return nil, true
	case *ast.Identifier:
		if p.Name == "_" {
			return nil, true
		}
		if g.isSingletonPattern(ctx, p.Name) {
			return nil, true
		}
		goName := sanitizeIdent(p.Name)
		ctx.locals[p.Name] = paramInfo{Name: p.Name, GoName: goName, GoType: subjectType}
		return []string{
			fmt.Sprintf("var %s %s = %s", goName, subjectType, subjectTemp),
			fmt.Sprintf("_ = %s", goName),
		}, true
	case *ast.LiteralPattern:
		return nil, true
	case *ast.TypedPattern:
		if p.TypeAnnotation == nil {
			ctx.setReason("missing typed pattern annotation")
			return nil, false
		}
		if subjectType != "runtime.Value" && subjectType != "any" {
			mapped, ok := g.mapTypeExpressionInPackage(ctx.packageName, p.TypeAnnotation)
			if !ok || mapped == "" || mapped == "struct{}" {
				ctx.setReason("unsupported typed pattern")
				return nil, false
			}
			if mapped != subjectType {
				ctx.setReason("typed pattern type mismatch")
				return nil, false
			}
			return g.compileMatchPatternBindings(ctx, p.Pattern, subjectTemp, subjectType)
		}
		typeExpr, ok := g.renderTypeExpression(p.TypeAnnotation)
		if !ok {
			ctx.setReason("unsupported typed pattern")
			return nil, false
		}
		g.needsAst = true
		convertedTemp := ctx.newTemp()
		bindLines, ok := g.compileMatchPatternBindings(ctx, p.Pattern, convertedTemp, "runtime.Value")
		if !ok {
			return nil, false
		}
		if len(bindLines) == 0 {
			return nil, true
		}
		var lines []string
		castSubject := subjectTemp
		if subjectType == "any" {
			convTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, subjectTemp))
			castSubject = convTemp
		}
		lines = append(lines, fmt.Sprintf("%s, _ := __able_try_cast(%s, %s)", convertedTemp, castSubject, typeExpr))
		lines = append(lines, bindLines...)
		return lines, true
	case *ast.StructPattern:
		if subjectType == "runtime.Value" || subjectType == "any" {
			if subjectType == "any" {
				convTemp := ctx.newTemp()
				bindLines, ok := g.compileRuntimeStructPatternBindings(ctx, p, convTemp)
				if !ok {
					return nil, false
				}
				if len(bindLines) == 0 {
					return nil, true
				}
				return append([]string{fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, subjectTemp)}, bindLines...), true
			}
			return g.compileRuntimeStructPatternBindings(ctx, p, subjectTemp)
		}
		info := g.structInfoByGoName(subjectType)
		if info == nil {
			ctx.setReason("unsupported struct pattern")
			return nil, false
		}
		if p.StructType != nil && p.StructType.Name != "" && info.Name != p.StructType.Name {
			ctx.setReason("struct pattern type mismatch")
			return nil, false
		}
		if p.IsPositional {
			if info.Kind != ast.StructKindPositional {
				ctx.setReason("struct pattern positional mismatch")
				return nil, false
			}
			if len(p.Fields) != len(info.Fields) {
				ctx.setReason("struct pattern arity mismatch")
				return nil, false
			}
			lines := []string{}
			for idx, field := range p.Fields {
				pattern, ok := positionalStructFieldPattern(field)
				if !ok {
					ctx.setReason("invalid struct pattern field")
					return nil, false
				}
				fieldInfo := info.Fields[idx]
				fieldExpr := fmt.Sprintf("%s.%s", subjectTemp, fieldInfo.GoName)
				fieldLines, ok := g.compileMatchPatternBindings(ctx, pattern, fieldExpr, fieldInfo.GoType)
				if !ok {
					return nil, false
				}
				lines = append(lines, fieldLines...)
				if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
					bindName := sanitizeIdent(field.Binding.Name)
					ctx.locals[field.Binding.Name] = paramInfo{Name: field.Binding.Name, GoName: bindName, GoType: fieldInfo.GoType}
					lines = append(lines,
						fmt.Sprintf("var %s %s = %s", bindName, fieldInfo.GoType, fieldExpr),
						fmt.Sprintf("_ = %s", bindName),
					)
				}
			}
			return lines, true
		}
		if info.Kind == ast.StructKindPositional {
			ctx.setReason("struct pattern positional mismatch")
			return nil, false
		}
		lines := []string{}
		for _, field := range p.Fields {
			if field == nil || field.Pattern == nil {
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
			fieldLines, ok := g.compileMatchPatternBindings(ctx, field.Pattern, fieldExpr, fieldInfo.GoType)
			if !ok {
				return nil, false
			}
			lines = append(lines, fieldLines...)
			if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
				bindName := sanitizeIdent(field.Binding.Name)
				ctx.locals[field.Binding.Name] = paramInfo{Name: field.Binding.Name, GoName: bindName, GoType: fieldInfo.GoType}
				lines = append(lines,
					fmt.Sprintf("var %s %s = %s", bindName, fieldInfo.GoType, fieldExpr),
					fmt.Sprintf("_ = %s", bindName),
				)
			}
		}
		return lines, true
	case *ast.ArrayPattern:
		if subjectType != "runtime.Value" {
			ctx.setReason("array pattern unsupported")
			return nil, false
		}
		return g.compileRuntimeArrayPatternBindings(ctx, p, subjectTemp)
	default:
		ctx.setReason("unsupported match pattern")
		return nil, false
	}
}

func (g *generator) compileRuntimeStructPatternCondition(ctx *compileContext, pattern *ast.StructPattern, subjectTemp string) ([]string, string, bool) {
	if pattern == nil {
		ctx.setReason("missing struct pattern")
		return nil, "", false
	}
	condTemp := ctx.newTemp()
	condLabel := ctx.newTemp()
	instTemp := ctx.newTemp()

	iterEndAction := fmt.Sprintf("break %s", condLabel)
	if pattern.StructType != nil && pattern.StructType.Name == "IteratorEnd" && len(pattern.Fields) == 0 {
		iterEndAction = fmt.Sprintf("%s = true; break %s", condTemp, condLabel)
	}

	inner := []string{
		fmt.Sprintf("switch %s.(type) { case runtime.IteratorEndValue, *runtime.IteratorEndValue: %s }", subjectTemp, iterEndAction),
		fmt.Sprintf("%s := __able_struct_instance(%s)", instTemp, subjectTemp),
		fmt.Sprintf("if %s == nil { break %s }", instTemp, condLabel),
	}
	if pattern.StructType != nil && pattern.StructType.Name != "" {
		inner = append(inner, fmt.Sprintf("if %s.Definition == nil || %s.Definition.Node == nil || %s.Definition.Node.ID == nil || %s.Definition.Node.ID.Name != %q { break %s }", instTemp, instTemp, instTemp, instTemp, pattern.StructType.Name, condLabel))
	}
	positionalTemp := ctx.newTemp()
	inner = append(inner, fmt.Sprintf("%s := %s.Positional", positionalTemp, instTemp))
	inner = append(inner, fmt.Sprintf("if %s != nil {", positionalTemp))
	inner = append(inner, fmt.Sprintf("\tif len(%s) != %d { break %s }", positionalTemp, len(pattern.Fields), condLabel))
	for idx, field := range pattern.Fields {
		fieldPattern, ok := positionalStructFieldPattern(field)
		if !ok {
			ctx.setReason("invalid struct pattern field")
			return nil, "", false
		}
		fieldExpr := fmt.Sprintf("%s[%d]", positionalTemp, idx)
		fieldCondLines, fieldCond, ok := g.compileMatchPatternCondition(ctx, fieldPattern, fieldExpr, "runtime.Value")
		if !ok {
			return nil, "", false
		}
		if len(fieldCondLines) > 0 {
			inner = append(inner, indentLines(fieldCondLines, 1)...)
		}
		if fieldCond != "true" {
			inner = append(inner, fmt.Sprintf("\tif !(%s) { break %s }", fieldCond, condLabel))
		}
	}
	inner = append(inner, fmt.Sprintf("\t%s = true; break %s", condTemp, condLabel), "}")
	if len(pattern.Fields) > 0 {
		inner = append(inner, fmt.Sprintf("if %s.Fields == nil { break %s }", instTemp, condLabel))
	}
	for _, field := range pattern.Fields {
		fieldPattern, ok := positionalStructFieldPattern(field)
		if !ok {
			ctx.setReason("invalid struct pattern field")
			return nil, "", false
		}
		if field.FieldName == nil || field.FieldName.Name == "" {
			inner = append(inner, fmt.Sprintf("break %s", condLabel))
			break
		}
		fieldOk := ctx.newTemp()
		fieldExpr := fmt.Sprintf("%s.Fields[%q]", instTemp, field.FieldName.Name)
		fieldCondLines, fieldCond, ok := g.compileMatchPatternCondition(ctx, fieldPattern, fieldExpr, "runtime.Value")
		if !ok {
			return nil, "", false
		}
		if fieldCond == "true" && len(fieldCondLines) == 0 {
			inner = append(inner, fmt.Sprintf("_, %s := %s", fieldOk, fieldExpr))
			inner = append(inner, fmt.Sprintf("if !%s { break %s }", fieldOk, condLabel))
			continue
		}
		fieldTemp := ctx.newTemp()
		inner = append(inner, fmt.Sprintf("%s, %s := %s", fieldTemp, fieldOk, fieldExpr))
		inner = append(inner, fmt.Sprintf("if !%s { break %s }", fieldOk, condLabel))
		fieldCondLines2, fieldCond2, ok := g.compileMatchPatternCondition(ctx, fieldPattern, fieldTemp, "runtime.Value")
		if !ok {
			return nil, "", false
		}
		if len(fieldCondLines2) > 0 {
			inner = append(inner, fieldCondLines2...)
		}
		inner = append(inner, fmt.Sprintf("if !(%s) { break %s }", fieldCond2, condLabel))
	}
	inner = append(inner, fmt.Sprintf("%s = true", condTemp))

	lines := []string{
		fmt.Sprintf("%s := false", condTemp),
		fmt.Sprintf("%s: switch { default: %s }", condLabel, strings.Join(inner, "; ")),
	}
	return lines, condTemp, true
}

func (g *generator) compileRuntimeStructPatternBindings(ctx *compileContext, pattern *ast.StructPattern, subjectTemp string) ([]string, bool) {
	if pattern == nil {
		ctx.setReason("missing struct pattern")
		return nil, false
	}
	if len(pattern.Fields) == 0 {
		return nil, true
	}
	instTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := __able_struct_instance(%s)", instTemp, subjectTemp),
	}
	positionalTemp := ""
	positionalTemp = ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s := %s.Positional", positionalTemp, instTemp),
		fmt.Sprintf("_ = %s", positionalTemp),
	)
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
		fieldLines, ok := g.compileMatchPatternBindings(ctx, fieldPattern, fieldExpr, "runtime.Value")
		if !ok {
			return nil, false
		}
		lines = append(lines, fieldLines...)
		if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
			bindName := sanitizeIdent(field.Binding.Name)
			ctx.locals[field.Binding.Name] = paramInfo{Name: field.Binding.Name, GoName: bindName, GoType: "runtime.Value"}
			lines = append(lines,
				fmt.Sprintf("var %s runtime.Value = %s", bindName, fieldExpr),
				fmt.Sprintf("_ = %s", bindName),
			)
		}
	}
	return lines, true
}

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

func positionalStructFieldPattern(field *ast.StructPatternField) (ast.Pattern, bool) {
	if field == nil {
		return nil, false
	}
	if field.Pattern != nil {
		return field.Pattern, true
	}
	if field.FieldName != nil {
		return field.FieldName, true
	}
	return nil, false
}

func (g *generator) runtimeTypeCheckForTypeExpression(ctx *compileContext, expr ast.TypeExpression, subjectTemp string) ([]string, string, bool) {
	if expr == nil {
		return nil, "", false
	}
	switch t := expr.(type) {
	case *ast.GenericTypeExpression:
		base, ok := t.Base.(*ast.SimpleTypeExpression)
		if !ok || base.Name == nil {
			return nil, "", false
		}
		switch base.Name.Name {
		case "Array":
			okTemp := ctx.newTemp()
			lines := []string{fmt.Sprintf("_, %s := __able_array_values(%s)", okTemp, subjectTemp)}
			return lines, okTemp, true
		case "Map", "HashMap":
			condTemp := ctx.newTemp()
			lines := []string{
				fmt.Sprintf("var %s bool", condTemp),
				fmt.Sprintf("switch v := %s.(type) {", subjectTemp),
				fmt.Sprintf("case *runtime.HashMapValue: %s = v != nil", condTemp),
				fmt.Sprintf("case *runtime.StructInstanceValue: %s = v != nil && v.Definition != nil && v.Definition.Node != nil && v.Definition.Node.ID != nil && v.Definition.Node.ID.Name == \"HashMap\"", condTemp),
				"}",
			}
			return lines, condTemp, true
		case "DivMod":
			condTemp := ctx.newTemp()
			castTemp := ctx.newTemp()
			castOkTemp := ctx.newTemp()
			lines := []string{
				fmt.Sprintf("%s, %s := %s.(*runtime.StructInstanceValue)", castTemp, castOkTemp, subjectTemp),
				fmt.Sprintf("%s := %s && %s != nil && %s.Definition != nil && %s.Definition.Node != nil && %s.Definition.Node.ID != nil && %s.Definition.Node.ID.Name == %q",
					condTemp, castOkTemp, castTemp, castTemp, castTemp, castTemp, castTemp, base.Name.Name),
			}
			return lines, condTemp, true
		}
	}
	return nil, "", false
}

func (g *generator) runtimeTypeCheckExpr(ctx *compileContext, subjectTemp string, goType string) ([]string, string, bool) {
	condTemp := ctx.newTemp()
	typeSwitch := func(valueType, ptrType, condition string) ([]string, string, bool) {
		lines := []string{
			fmt.Sprintf("var %s bool", condTemp),
			fmt.Sprintf("switch v := %s.(type) {", subjectTemp),
			fmt.Sprintf("case %s: %s = %s", valueType, condTemp, condition),
			fmt.Sprintf("case %s: %s = v != nil", ptrType, condTemp),
			"}",
		}
		if condition != "true" {
			lines[3] = fmt.Sprintf("case %s: %s = v != nil && %s", ptrType, condTemp, condition)
		}
		return lines, condTemp, true
	}
	switch g.typeCategory(goType) {
	case "bool":
		return typeSwitch("runtime.BoolValue", "*runtime.BoolValue", "true")
	case "string":
		return typeSwitch("runtime.StringValue", "*runtime.StringValue", "true")
	case "rune":
		return typeSwitch("runtime.CharValue", "*runtime.CharValue", "true")
	case "int", "uint", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64":
		suffix, ok := g.runtimeIntegerSuffix(goType)
		if !ok {
			return nil, "", false
		}
		cond := fmt.Sprintf("v.TypeSuffix == runtime.IntegerType(%q)", suffix)
		return typeSwitch("runtime.IntegerValue", "*runtime.IntegerValue", cond)
	case "float32":
		return typeSwitch("runtime.FloatValue", "*runtime.FloatValue", "v.TypeSuffix == runtime.FloatF32")
	case "float64":
		return typeSwitch("runtime.FloatValue", "*runtime.FloatValue", "v.TypeSuffix == runtime.FloatF64")
	case "struct":
		info := g.structInfoByGoName(goType)
		if info == nil || info.Name == "" {
			return nil, "", false
		}
		castTemp := ctx.newTemp()
		castOkTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s, %s := %s.(*runtime.StructInstanceValue)", castTemp, castOkTemp, subjectTemp),
			fmt.Sprintf("%s := %s && %s != nil && %s.Definition != nil && %s.Definition.Node != nil && %s.Definition.Node.ID != nil && %s.Definition.Node.ID.Name == %q",
				condTemp, castOkTemp, castTemp, castTemp, castTemp, castTemp, castTemp, info.Name),
		}
		return lines, condTemp, true
	default:
		return nil, "", false
	}
}

func (g *generator) runtimeIntegerSuffix(goType string) (string, bool) {
	switch goType {
	case "int8":
		return "i8", true
	case "int16":
		return "i16", true
	case "int32":
		return "i32", true
	case "int64":
		return "i64", true
	case "uint8":
		return "u8", true
	case "uint16":
		return "u16", true
	case "uint32":
		return "u32", true
	case "uint64":
		return "u64", true
	case "int":
		return "isize", true
	case "uint":
		return "usize", true
	default:
		return "", false
	}
}

func (g *generator) compileLiteralMatch(ctx *compileContext, lit ast.Literal, subjectTemp string, subjectType string) ([]string, string, bool) {
	if lit == nil {
		ctx.setReason("missing literal pattern")
		return nil, "", false
	}
	if subjectType == "runtime.Value" || subjectType == "any" {
		if _, ok := lit.(*ast.NilLiteral); ok {
			effectiveSubject := subjectTemp
			if subjectType == "any" {
				effectiveSubject = fmt.Sprintf("__able_any_to_value(%s)", subjectTemp)
			}
			return nil, fmt.Sprintf("__able_is_nil(%s)", effectiveSubject), true
		}
		litLines, expr, goType, ok := g.compileExprLines(ctx, lit.(ast.Expression), "")
		if !ok {
			return nil, "", false
		}
		litConvLines, litRuntime, ok := g.runtimeValueLines(ctx, expr, goType)
		if !ok {
			ctx.setReason("unsupported literal pattern")
			return nil, "", false
		}
		var lines []string
		lines = append(lines, litLines...)
		lines = append(lines, litConvLines...)
		effectiveSubject := subjectTemp
		if subjectType == "any" {
			effectiveSubject = fmt.Sprintf("__able_any_to_value(%s)", subjectTemp)
		}
		condExpr := fmt.Sprintf("__able_binary_op(%q, %s, %s)", "==", effectiveSubject, litRuntime)
		expectLines, converted, ok := g.expectRuntimeValueExprLines(ctx, condExpr, "bool")
		if !ok {
			ctx.setReason("unsupported literal comparison")
			return nil, "", false
		}
		lines = append(lines, expectLines...)
		return lines, converted, true
	}
	if _, ok := lit.(*ast.NilLiteral); ok && strings.HasPrefix(subjectType, "*") {
		return nil, fmt.Sprintf("(%s == nil)", subjectTemp), true
	}
	litLines, expr, _, ok := g.compileExprLines(ctx, lit.(ast.Expression), subjectType)
	if !ok {
		return nil, "", false
	}
	if !g.isEqualityComparable(subjectType) {
		ctx.setReason(fmt.Sprintf("unsupported literal comparison (type=%s)", subjectType))
		return nil, "", false
	}
	return litLines, fmt.Sprintf("(%s == %s)", subjectTemp, expr), true
}
