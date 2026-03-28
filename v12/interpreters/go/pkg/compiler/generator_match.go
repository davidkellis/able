package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) staticTypedPatternCompatible(subjectType string, patternType string) bool {
	if g == nil || subjectType == "" || patternType == "" {
		return false
	}
	if subjectType == patternType {
		return true
	}
	return g.receiverGoTypeCompatible(patternType, subjectType) || g.receiverGoTypeCompatible(subjectType, patternType)
}

func (g *generator) compileMatchPattern(ctx *compileContext, pattern ast.Pattern, subjectTemp string, subjectType string) ([]string, string, []string, bool) {
	if pattern == nil {
		ctx.setReason("missing match pattern")
		return nil, "", nil, false
	}
	condLines, cond, ok := g.compileMatchPatternCondition(ctx, pattern, subjectTemp, subjectType)
	if !ok {
		return nil, "", nil, false
	}
	if cond == "false" && len(condLines) == 0 {
		return nil, "false", nil, true
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

func (g *generator) guardMatchConditionWithPredicate(ctx *compileContext, guardExpr string, innerLines []string, innerCond string) ([]string, string, bool) {
	if innerCond == "true" && len(innerLines) == 0 {
		return nil, fmt.Sprintf("(%s)", guardExpr), true
	}
	condTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := false", condTemp),
		fmt.Sprintf("if %s {", guardExpr),
	}
	lines = append(lines, indentLines(innerLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", condTemp, innerCond))
	lines = append(lines, "}")
	return lines, condTemp, true
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
			if union := g.nativeUnionInfoForGoType(subjectType); union != nil {
				info, _ := g.structInfoForTypeName(ctx.packageName, p.Name)
				memberType := ""
				if info != nil {
					memberType = "*" + info.GoName
				}
				if member, ok := g.nativeUnionMember(union, memberType); ok {
					condTemp := ctx.newTemp()
					lines := []string{
						fmt.Sprintf("_, %s := %s(%s)", condTemp, member.UnwrapHelper, subjectTemp),
					}
					return lines, condTemp, true
				}
				return nil, "false", true
			}
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
				return nil, "false", true
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
		if g.nativeUnionInfoForGoType(subjectType) != nil {
			return g.compileNativeUnionTypedPatternCondition(ctx, subjectTemp, subjectType, p)
		}
		if subjectType != "runtime.Value" && subjectType != "any" {
			mapped, ok := g.lowerCarrierType(ctx, p.TypeAnnotation)
			if !ok || mapped == "" || mapped == "struct{}" {
				ctx.setReason("unsupported typed pattern")
				return nil, "", false
			}
			if innerType, nullable := g.nativeNullableValueInnerType(subjectType); nullable {
				if mapped != innerType {
					return nil, "false", true
				}
				innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, p.Pattern, fmt.Sprintf("(*%s)", subjectTemp), innerType)
				if !ok {
					return nil, "", false
				}
				return g.guardMatchConditionWithPredicate(ctx, fmt.Sprintf("%s != nil", subjectTemp), innerCondLines, innerCond)
			}
			if !g.staticTypedPatternCompatible(subjectType, mapped) {
				return nil, "false", true
			}
			innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, p.Pattern, subjectTemp, subjectType)
			if !ok {
				return nil, "", false
			}
			if g.isNativeStructPointerType(subjectType) && !g.typeExprAllowsNilInPackage(ctx.packageName, g.lowerNormalizedTypeExpr(ctx, p.TypeAnnotation)) {
				return g.guardMatchConditionWithPredicate(ctx, fmt.Sprintf("%s != nil", subjectTemp), innerCondLines, innerCond)
			}
			return innerCondLines, innerCond, true
		}
		typeExpr, ok := g.renderTypeExpression(g.lowerNormalizedTypeExpr(ctx, p.TypeAnnotation))
		castSubject := subjectTemp
		var prefixLines []string
		if subjectType == "any" {
			convTemp := ctx.newTemp()
			prefixLines = append(prefixLines, fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, subjectTemp))
			castSubject = convTemp
		}
		if mapped, mappedOK := g.lowerCarrierType(ctx, p.TypeAnnotation); mappedOK && g.isNativeStructPointerType(mapped) {
			if checkLines, checkCond, checkOK := g.lowerPatternRuntimeTypeCheck(ctx, p.TypeAnnotation, castSubject); checkOK {
				baseName, _ := g.structBaseName(mapped)
				convertedTemp := ctx.newTemp()
				innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, p.Pattern, convertedTemp, mapped)
				if !ok {
					return nil, "", false
				}
				lines := append(prefixLines, checkLines...)
				if innerCond == "true" && len(innerCondLines) == 0 {
					return lines, checkCond, true
				}
				errTemp := ctx.newTemp()
				controlTemp := ctx.newTemp()
				innerLines := []string{
					fmt.Sprintf("%s, %s := __able_struct_%s_from(%s)", convertedTemp, errTemp, baseName, castSubject),
					fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
				}
				controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
				if !ok {
					return nil, "", false
				}
				innerLines = append(innerLines, controlLines...)
				innerLines = append(innerLines, innerCondLines...)
				guardedLines, cond, ok := g.guardMatchConditionWithPredicate(ctx, checkCond, innerLines, innerCond)
				if !ok {
					return nil, "", false
				}
				return append(lines, guardedLines...), cond, true
			}
		}
		if mapped, mappedOK := g.lowerCarrierType(ctx, p.TypeAnnotation); mappedOK && g.isNativeStructPointerType(mapped) {
			info := g.structInfoByGoName(mapped)
			if info != nil && info.Kind != ast.StructKindSingleton {
				baseName, _ := g.structBaseName(mapped)
				instTemp := ctx.newTemp()
				convertedTemp := ctx.newTemp()
				innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, p.Pattern, convertedTemp, mapped)
				if !ok {
					return nil, "", false
				}
				lines := append(prefixLines, fmt.Sprintf("%s := __able_struct_instance(%s)", instTemp, castSubject))
				guardExpr := fmt.Sprintf("%s != nil && %s.Definition != nil && %s.Definition.Node != nil && %s.Definition.Node.ID != nil && %s.Definition.Node.ID.Name == %q", instTemp, instTemp, instTemp, instTemp, instTemp, info.Name)
				if innerCond == "true" && len(innerCondLines) == 0 {
					return lines, guardExpr, true
				}
				errTemp := ctx.newTemp()
				innerLines := []string{fmt.Sprintf("%s, %s := __able_struct_%s_from(%s)", convertedTemp, errTemp, baseName, castSubject)}
				controlTemp := ctx.newTemp()
				innerLines = append(innerLines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
				controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
				if !ok {
					return nil, "", false
				}
				innerLines = append(innerLines, controlLines...)
				innerLines = append(innerLines, innerCondLines...)
				guardedLines, cond, ok := g.guardMatchConditionWithPredicate(ctx, guardExpr, innerLines, innerCond)
				if !ok {
					return nil, "", false
				}
				return append(lines, guardedLines...), cond, true
			}
		}
		if !ok {
			ctx.setReason("unsupported typed pattern")
			return nil, "", false
		}
		g.needsAst = true
		castTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, p.Pattern, castTemp, "runtime.Value")
		if !ok {
			return nil, "", false
		}
		if innerCond == "true" && len(innerCondLines) == 0 {
			controlTemp := ctx.newTemp()
			lines := append(prefixLines, fmt.Sprintf("_, %s, %s := __able_try_cast(%s, %s)", okTemp, controlTemp, castSubject, typeExpr))
			controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
			if !ok {
				return nil, "", false
			}
			lines = append(lines, controlLines...)
			return lines, okTemp, true
		}
		condTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := append(prefixLines,
			fmt.Sprintf("%s, %s, %s := __able_try_cast(%s, %s)", castTemp, okTemp, controlTemp, castSubject, typeExpr),
			fmt.Sprintf("var %s bool", condTemp),
		)
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
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
		if union := g.nativeUnionInfoForGoType(subjectType); union != nil && p.StructType != nil && p.StructType.Name != "" {
			mapped, ok := g.lowerNativeUnionPatternMemberType(ctx, subjectType, ast.Ty(p.StructType.Name))
			if !ok {
				ctx.setReason("unsupported struct pattern")
				return nil, "", false
			}
			member, ok := g.nativeUnionMember(union, mapped)
			if !ok {
				return nil, "false", true
			}
			okTemp := ctx.newTemp()
			castTemp := ctx.newTemp()
			innerLines, innerCond, ok := g.compileMatchPatternCondition(ctx, p, castTemp, mapped)
			if !ok {
				return nil, "", false
			}
			if innerCond == "true" && len(innerLines) == 0 {
				return []string{fmt.Sprintf("_, %s := %s(%s)", okTemp, member.UnwrapHelper, subjectTemp)}, okTemp, true
			}
			condTemp := ctx.newTemp()
			lines := []string{
				fmt.Sprintf("%s, %s := %s(%s)", castTemp, okTemp, member.UnwrapHelper, subjectTemp),
				fmt.Sprintf("var %s bool", condTemp),
				fmt.Sprintf("if %s {", okTemp),
			}
			lines = append(lines, indentLines(innerLines, 1)...)
			lines = append(lines, fmt.Sprintf("\t%s = %s", condTemp, innerCond))
			lines = append(lines, "}")
			return lines, condTemp, true
		}
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
			return nil, "false", true
		}
		if p.IsPositional {
			if info.Kind != ast.StructKindPositional {
				return nil, "false", true
			}
			if len(p.Fields) != len(info.Fields) {
				return nil, "false", true
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
				if g.isNativeStructPointerType(subjectType) {
					return g.guardMatchConditionWithPredicate(ctx, fmt.Sprintf("%s != nil", subjectTemp), allLines, "true")
				}
				return allLines, "true", true
			}
			condExpr := strings.Join(conds, " && ")
			if g.isNativeStructPointerType(subjectType) {
				return g.guardMatchConditionWithPredicate(ctx, fmt.Sprintf("%s != nil", subjectTemp), allLines, condExpr)
			}
			return allLines, condExpr, true
		}
		if info.Kind == ast.StructKindPositional {
			return nil, "false", true
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
			if g.isNativeStructPointerType(subjectType) {
				return g.guardMatchConditionWithPredicate(ctx, fmt.Sprintf("%s != nil", subjectTemp), allLines, "true")
			}
			return allLines, "true", true
		}
		condExpr := strings.Join(conds, " && ")
		if g.isNativeStructPointerType(subjectType) {
			return g.guardMatchConditionWithPredicate(ctx, fmt.Sprintf("%s != nil", subjectTemp), allLines, condExpr)
		}
		return allLines, condExpr, true
	case *ast.ArrayPattern:
		if subjectType == "runtime.Value" {
			return g.compileRuntimeArrayPatternCondition(ctx, p, subjectTemp)
		}
		if subjectType == "any" {
			convertedTemp := ctx.newTemp()
			lines := []string{fmt.Sprintf("%s := __able_any_to_value(%s)", convertedTemp, subjectTemp)}
			innerLines, cond, ok := g.compileRuntimeArrayPatternCondition(ctx, p, convertedTemp)
			if !ok {
				return nil, "", false
			}
			return append(lines, innerLines...), cond, true
		}
		if !g.isStaticArrayType(subjectType) {
			ctx.setReason("array pattern unsupported")
			return nil, "", false
		}
		return g.compileNativeArrayPatternCondition(ctx, p, subjectTemp, subjectType)
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
		typeExpr := ast.TypeExpression(nil)
		switch {
		case (subjectType == "runtime.Value" || subjectType == "any") && ctx != nil && ctx.expectedTypeExpr != nil:
			typeExpr = g.lowerNormalizedTypeExpr(ctx, ctx.expectedTypeExpr)
		case subjectType != "" && subjectType != "runtime.Value" && subjectType != "any":
			typeExpr, _ = g.typeExprForGoType(subjectType)
			typeExpr = g.lowerNormalizedTypeExpr(ctx, typeExpr)
		}
		ctx.setLocalBinding(p.Name, paramInfo{Name: p.Name, GoName: goName, GoType: subjectType, TypeExpr: typeExpr})
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
		if g.nativeUnionInfoForGoType(subjectType) != nil {
			return g.compileNativeUnionTypedPatternBindings(ctx, subjectTemp, subjectType, p)
		}
		if subjectType != "runtime.Value" && subjectType != "any" {
			mapped, ok := g.lowerCarrierType(ctx, p.TypeAnnotation)
			if !ok || mapped == "" || mapped == "struct{}" {
				ctx.setReason("unsupported typed pattern")
				return nil, false
			}
			if innerType, nullable := g.nativeNullableValueInnerType(subjectType); nullable {
				if mapped != innerType {
					ctx.setReason("typed pattern type mismatch")
					return nil, false
				}
				return g.compileMatchPatternBindings(ctx, p.Pattern, fmt.Sprintf("(*%s)", subjectTemp), innerType)
			}
			if !g.staticTypedPatternCompatible(subjectType, mapped) {
				ctx.setReason("typed pattern type mismatch")
				return nil, false
			}
			return g.compileMatchPatternBindings(ctx, p.Pattern, subjectTemp, subjectType)
		}
		convertedTemp := ctx.newTemp()
		var lines []string
		castSubject := subjectTemp
		if subjectType == "any" {
			convTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, subjectTemp))
			castSubject = convTemp
		}
		if mapped, mappedOK := g.lowerCarrierType(ctx, p.TypeAnnotation); mappedOK && g.isNativeStructPointerType(mapped) {
			baseName, _ := g.structBaseName(mapped)
			bindLines, ok := g.compileMatchPatternBindings(ctx, p.Pattern, convertedTemp, mapped)
			if !ok {
				return nil, false
			}
			if len(bindLines) == 0 {
				return nil, true
			}
			errTemp := ctx.newTemp()
			controlTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s, %s := __able_struct_%s_from(%s)", convertedTemp, errTemp, baseName, castSubject))
			lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
			controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
			if !ok {
				return nil, false
			}
			lines = append(lines, controlLines...)
			lines = append(lines, bindLines...)
			return lines, true
		}
		dynamicLines, narrowedTemp, narrowedType, _, ok := g.compileDynamicTypedPatternCast(ctx, castSubject, "runtime.Value", p.TypeAnnotation)
		if !ok {
			ctx.setReason("unsupported typed pattern")
			return nil, false
		}
		previousExpected := ctx.expectedTypeExpr
		ctx.expectedTypeExpr = p.TypeAnnotation
		bindLines, ok := g.compileMatchPatternBindings(ctx, p.Pattern, narrowedTemp, narrowedType)
		ctx.expectedTypeExpr = previousExpected
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
		if union := g.nativeUnionInfoForGoType(subjectType); union != nil && p.StructType != nil && p.StructType.Name != "" {
			mapped, ok := g.lowerNativeUnionPatternMemberType(ctx, subjectType, ast.Ty(p.StructType.Name))
			if !ok {
				ctx.setReason("unsupported struct pattern")
				return nil, false
			}
			member, ok := g.nativeUnionMember(union, mapped)
			if !ok {
				ctx.setReason("struct pattern type mismatch")
				return nil, false
			}
			convertedTemp := ctx.newTemp()
			bindLines, ok := g.compileMatchPatternBindings(ctx, p, convertedTemp, mapped)
			if !ok {
				return nil, false
			}
			if len(bindLines) == 0 {
				return nil, true
			}
			lines := []string{fmt.Sprintf("%s, _ := %s(%s)", convertedTemp, member.UnwrapHelper, subjectTemp)}
			lines = append(lines, bindLines...)
			return lines, true
		}
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
					bindingTypeExpr := g.lowerNormalizedTypeExpr(ctx, fieldInfo.TypeExpr)
					if bindingTypeExpr == nil {
						bindingTypeExpr, _ = g.typeExprForGoType(fieldInfo.GoType)
						bindingTypeExpr = g.lowerNormalizedTypeExpr(ctx, bindingTypeExpr)
					}
					ctx.setLocalBinding(field.Binding.Name, paramInfo{
						Name:     field.Binding.Name,
						GoName:   bindName,
						GoType:   fieldInfo.GoType,
						TypeExpr: bindingTypeExpr,
					})
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
				bindingTypeExpr := g.lowerNormalizedTypeExpr(ctx, fieldInfo.TypeExpr)
				if bindingTypeExpr == nil {
					bindingTypeExpr, _ = g.typeExprForGoType(fieldInfo.GoType)
					bindingTypeExpr = g.lowerNormalizedTypeExpr(ctx, bindingTypeExpr)
				}
				ctx.setLocalBinding(field.Binding.Name, paramInfo{
					Name:     field.Binding.Name,
					GoName:   bindName,
					GoType:   fieldInfo.GoType,
					TypeExpr: bindingTypeExpr,
				})
				lines = append(lines,
					fmt.Sprintf("var %s %s = %s", bindName, fieldInfo.GoType, fieldExpr),
					fmt.Sprintf("_ = %s", bindName),
				)
			}
		}
		return lines, true
	case *ast.ArrayPattern:
		if subjectType == "runtime.Value" {
			return g.compileRuntimeArrayPatternBindings(ctx, p, subjectTemp)
		}
		if subjectType == "any" {
			convertedTemp := ctx.newTemp()
			lines := []string{fmt.Sprintf("%s := __able_any_to_value(%s)", convertedTemp, subjectTemp)}
			bindLines, ok := g.compileRuntimeArrayPatternBindings(ctx, p, convertedTemp)
			if !ok {
				return nil, false
			}
			return append(lines, bindLines...), true
		}
		if !g.isStaticArrayType(subjectType) {
			ctx.setReason("array pattern unsupported")
			return nil, false
		}
		return g.compileNativeArrayPatternBindings(ctx, p, subjectTemp, subjectType)
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
	inner = append(inner, fmt.Sprintf("\tif %s != %d { break %s }", g.staticSliceLenExpr(positionalTemp), len(pattern.Fields), condLabel))
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
			ctx.setLocalBinding(field.Binding.Name, paramInfo{Name: field.Binding.Name, GoName: bindName, GoType: "runtime.Value"})
			lines = append(lines,
				fmt.Sprintf("var %s runtime.Value = %s", bindName, fieldExpr),
				fmt.Sprintf("_ = %s", bindName),
			)
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
	if field.Binding != nil {
		return field.Binding, true
	}
	if field.FieldName != nil {
		return field.FieldName, true
	}
	return nil, false
}
