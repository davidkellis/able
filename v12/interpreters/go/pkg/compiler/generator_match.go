package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileMatchPattern(ctx *compileContext, pattern ast.Pattern, subjectTemp string, subjectType string) (string, []string, bool) {
	if pattern == nil {
		ctx.setReason("missing match pattern")
		return "", nil, false
	}
	cond, ok := g.compileMatchPatternCondition(ctx, pattern, subjectTemp, subjectType)
	if !ok {
		return "", nil, false
	}
	bindLines, ok := g.compileMatchPatternBindings(ctx, pattern, subjectTemp, subjectType)
	if !ok {
		return "", nil, false
	}
	return cond, bindLines, true
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
	info, ok := g.structs[name]
	if !ok || info == nil {
		return false
	}
	return info.Kind == ast.StructKindSingleton
}

func (g *generator) compileMatchPatternCondition(ctx *compileContext, pattern ast.Pattern, subjectTemp string, subjectType string) (string, bool) {
	if pattern == nil {
		ctx.setReason("missing match pattern")
		return "", false
	}
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return "true", true
	case *ast.Identifier:
		if p.Name == "_" {
			return "true", true
		}
		if g.isSingletonPattern(ctx, p.Name) {
			if subjectType == "runtime.Value" {
				condExpr := fmt.Sprintf("__able_match_singleton(%s, %q)", subjectTemp, p.Name)
				return condExpr, true
			}
			info := g.structs[p.Name]
			baseType := subjectType
			if name, ok := g.structBaseName(subjectType); ok {
				baseType = name
			}
			if info != nil && info.GoName != "" && baseType != info.GoName {
				ctx.setReason("singleton pattern type mismatch")
				return "", false
			}
		}
		return "true", true
	case *ast.LiteralPattern:
		cond, ok := g.compileLiteralMatch(ctx, p.Literal, subjectTemp, subjectType)
		if !ok {
			return "", false
		}
		return cond, true
	case *ast.TypedPattern:
		if p.TypeAnnotation == nil {
			ctx.setReason("missing typed pattern annotation")
			return "", false
		}
		if subjectType != "runtime.Value" {
			mapped, ok := g.mapTypeExpressionInPackage(ctx.packageName, p.TypeAnnotation)
			if !ok || mapped == "" || mapped == "struct{}" {
				ctx.setReason("unsupported typed pattern")
				return "", false
			}
			if mapped != subjectType {
				ctx.setReason("typed pattern type mismatch")
				return "", false
			}
			return g.compileMatchPatternCondition(ctx, p.Pattern, subjectTemp, subjectType)
		}
		typeExpr, ok := g.renderTypeExpression(p.TypeAnnotation)
		if !ok {
			ctx.setReason("unsupported typed pattern")
			return "", false
		}
		g.needsAst = true
		castTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		innerCond, ok := g.compileMatchPatternCondition(ctx, p.Pattern, castTemp, "runtime.Value")
		if !ok {
			return "", false
		}
		if innerCond == "true" {
			lines := []string{
				fmt.Sprintf("_, %s := __able_try_cast(%s, %s)", okTemp, subjectTemp, typeExpr),
				fmt.Sprintf("return %s", okTemp),
			}
			return fmt.Sprintf("func() bool { %s }()", strings.Join(lines, "; ")), true
		}
		lines := []string{
			fmt.Sprintf("%s, %s := __able_try_cast(%s, %s)", castTemp, okTemp, subjectTemp, typeExpr),
			fmt.Sprintf("if !%s { return false }", okTemp),
			fmt.Sprintf("return %s", innerCond),
		}
		return fmt.Sprintf("func() bool { %s }()", strings.Join(lines, "; ")), true
	case *ast.StructPattern:
		if subjectType == "runtime.Value" {
			return g.compileRuntimeStructPatternCondition(ctx, p, subjectTemp)
		}
		info := g.structInfoByGoName(subjectType)
		if info == nil {
			ctx.setReason("unsupported struct pattern")
			return "", false
		}
		if p.StructType != nil && p.StructType.Name != "" && info.Name != p.StructType.Name {
			ctx.setReason("struct pattern type mismatch")
			return "", false
		}
		if p.IsPositional {
			if info.Kind != ast.StructKindPositional {
				ctx.setReason("struct pattern positional mismatch")
				return "", false
			}
			if len(p.Fields) != len(info.Fields) {
				ctx.setReason("struct pattern arity mismatch")
				return "", false
			}
			conds := make([]string, 0, len(p.Fields))
			for idx, field := range p.Fields {
				pattern, ok := positionalStructFieldPattern(field)
				if !ok {
					ctx.setReason("invalid struct pattern field")
					return "", false
				}
				fieldInfo := info.Fields[idx]
				fieldExpr := fmt.Sprintf("%s.%s", subjectTemp, fieldInfo.GoName)
				fieldCond, ok := g.compileMatchPatternCondition(ctx, pattern, fieldExpr, fieldInfo.GoType)
				if !ok {
					return "", false
				}
				conds = append(conds, fieldCond)
			}
			if len(conds) == 0 {
				return "true", true
			}
			return strings.Join(conds, " && "), true
		}
		if info.Kind == ast.StructKindPositional {
			ctx.setReason("struct pattern positional mismatch")
			return "", false
		}
		conds := make([]string, 0, len(p.Fields))
		for _, field := range p.Fields {
			if field == nil || field.Pattern == nil {
				ctx.setReason("invalid struct pattern field")
				return "", false
			}
			if field.FieldName == nil || field.FieldName.Name == "" {
				ctx.setReason("struct pattern missing field name")
				return "", false
			}
			fieldInfo := g.fieldInfo(info, field.FieldName.Name)
			if fieldInfo == nil {
				ctx.setReason("unknown struct field")
				return "", false
			}
			fieldExpr := fmt.Sprintf("%s.%s", subjectTemp, fieldInfo.GoName)
			fieldCond, ok := g.compileMatchPatternCondition(ctx, field.Pattern, fieldExpr, fieldInfo.GoType)
			if !ok {
				return "", false
			}
			conds = append(conds, fieldCond)
		}
		if len(conds) == 0 {
			return "true", true
		}
		return strings.Join(conds, " && "), true
	case *ast.ArrayPattern:
		if subjectType != "runtime.Value" {
			ctx.setReason("array pattern unsupported")
			return "", false
		}
		return g.compileRuntimeArrayPatternCondition(ctx, p, subjectTemp)
	default:
		ctx.setReason("unsupported match pattern")
		return "", false
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
		if subjectType != "runtime.Value" {
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
		lines := []string{fmt.Sprintf("%s, _ := __able_try_cast(%s, %s)", convertedTemp, subjectTemp, typeExpr)}
		lines = append(lines, bindLines...)
		return lines, true
	case *ast.StructPattern:
		if subjectType == "runtime.Value" {
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

func (g *generator) compileRuntimeStructPatternCondition(ctx *compileContext, pattern *ast.StructPattern, subjectTemp string) (string, bool) {
	if pattern == nil {
		ctx.setReason("missing struct pattern")
		return "", false
	}
	instTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("switch %s.(type) { case runtime.IteratorEndValue, *runtime.IteratorEndValue: %s }", subjectTemp, func() string {
			if pattern.StructType != nil && pattern.StructType.Name == "IteratorEnd" && len(pattern.Fields) == 0 {
				return "return true"
			}
			return "return false"
		}()),
		fmt.Sprintf("%s := __able_struct_instance(%s)", instTemp, subjectTemp),
		fmt.Sprintf("if %s == nil { return false }", instTemp),
	}
	if pattern.StructType != nil && pattern.StructType.Name != "" {
		lines = append(lines, fmt.Sprintf("if %s.Definition == nil || %s.Definition.Node == nil || %s.Definition.Node.ID == nil || %s.Definition.Node.ID.Name != %q { return false }", instTemp, instTemp, instTemp, instTemp, pattern.StructType.Name))
	}
	positionalTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s.Positional", positionalTemp, instTemp))
	lines = append(lines, fmt.Sprintf("if %s != nil {", positionalTemp))
	lines = append(lines, fmt.Sprintf("if len(%s) != %d { return false }", positionalTemp, len(pattern.Fields)))
	for idx, field := range pattern.Fields {
		fieldPattern, ok := positionalStructFieldPattern(field)
		if !ok {
			ctx.setReason("invalid struct pattern field")
			return "", false
		}
		fieldExpr := fmt.Sprintf("%s[%d]", positionalTemp, idx)
		fieldCond, ok := g.compileMatchPatternCondition(ctx, fieldPattern, fieldExpr, "runtime.Value")
		if !ok {
			return "", false
		}
		if fieldCond != "true" {
			lines = append(lines, fmt.Sprintf("if !(%s) { return false }", fieldCond))
		}
	}
	lines = append(lines, "return true", "}")
	if len(pattern.Fields) > 0 {
		lines = append(lines, fmt.Sprintf("if %s.Fields == nil { return false }", instTemp))
	}
	for _, field := range pattern.Fields {
		fieldPattern, ok := positionalStructFieldPattern(field)
		if !ok {
			ctx.setReason("invalid struct pattern field")
			return "", false
		}
		if field.FieldName == nil || field.FieldName.Name == "" {
			lines = append(lines, "return false")
			break
		}
		fieldOk := ctx.newTemp()
		fieldExpr := fmt.Sprintf("%s.Fields[%q]", instTemp, field.FieldName.Name)
		fieldCond, ok := g.compileMatchPatternCondition(ctx, fieldPattern, fieldExpr, "runtime.Value")
		if !ok {
			return "", false
		}
		if fieldCond == "true" {
			lines = append(lines, fmt.Sprintf("_, %s := %s", fieldOk, fieldExpr))
			lines = append(lines, fmt.Sprintf("if !%s { return false }", fieldOk))
			continue
		}
		fieldTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s, %s := %s", fieldTemp, fieldOk, fieldExpr))
		lines = append(lines, fmt.Sprintf("if !%s { return false }", fieldOk))
		fieldCond, ok = g.compileMatchPatternCondition(ctx, fieldPattern, fieldTemp, "runtime.Value")
		if !ok {
			return "", false
		}
		lines = append(lines, fmt.Sprintf("if !(%s) { return false }", fieldCond))
	}
	lines = append(lines, "return true")
	return fmt.Sprintf("func() bool { %s }()", strings.Join(lines, "; ")), true
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
		fieldExpr := ""
		if field.FieldName != nil && field.FieldName.Name != "" {
			fieldExpr = fmt.Sprintf("func() runtime.Value { if %s != nil { return %s[%d] }; return %s.Fields[%q] }()", positionalTemp, positionalTemp, idx, instTemp, field.FieldName.Name)
		} else {
			fieldExpr = fmt.Sprintf("func() runtime.Value { if %s != nil { return %s[%d] }; return runtime.NilValue{} }()", positionalTemp, positionalTemp, idx)
		}
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

func (g *generator) compileRuntimeArrayPatternCondition(ctx *compileContext, pattern *ast.ArrayPattern, subjectTemp string) (string, bool) {
	if pattern == nil {
		ctx.setReason("missing array pattern")
		return "", false
	}
	if pattern.RestPattern != nil {
		switch pattern.RestPattern.(type) {
		case *ast.Identifier, *ast.WildcardPattern:
		default:
			ctx.setReason("unsupported rest pattern")
			return "", false
		}
	}
	valuesTemp := ctx.newTemp()
	okTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s, %s := __able_array_values(%s)", valuesTemp, okTemp, subjectTemp),
		fmt.Sprintf("if !%s { return false }", okTemp),
	}
	if pattern.RestPattern == nil {
		lines = append(lines, fmt.Sprintf("if len(%s) != %d { return false }", valuesTemp, len(pattern.Elements)))
	} else {
		lines = append(lines, fmt.Sprintf("if len(%s) < %d { return false }", valuesTemp, len(pattern.Elements)))
	}
	for idx, elem := range pattern.Elements {
		if elem == nil {
			ctx.setReason("invalid array pattern element")
			return "", false
		}
		elemExpr := fmt.Sprintf("%s[%d]", valuesTemp, idx)
		elemCond, ok := g.compileMatchPatternCondition(ctx, elem, elemExpr, "runtime.Value")
		if !ok {
			return "", false
		}
		if elemCond != "true" {
			lines = append(lines, fmt.Sprintf("if !(%s) { return false }", elemCond))
		}
	}
	lines = append(lines, "return true")
	return fmt.Sprintf("func() bool { %s }()", strings.Join(lines, "; ")), true
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

func (g *generator) runtimeTypeCheckForTypeExpression(ctx *compileContext, expr ast.TypeExpression, subjectTemp string) (string, bool) {
	if expr == nil {
		return "", false
	}
	switch t := expr.(type) {
	case *ast.GenericTypeExpression:
		base, ok := t.Base.(*ast.SimpleTypeExpression)
		if !ok || base.Name == nil {
			return "", false
		}
		switch base.Name.Name {
		case "Array":
			return fmt.Sprintf("func() bool { _, ok := __able_array_values(%s); return ok }()", subjectTemp), true
		case "Map", "HashMap":
			return fmt.Sprintf("func() bool { switch v := %s.(type) { case *runtime.HashMapValue: return v != nil; case *runtime.StructInstanceValue: if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil { return false }; return v.Definition.Node.ID.Name == \"HashMap\"; default: return false } }()", subjectTemp), true
		case "DivMod":
			return fmt.Sprintf("func() bool { v, ok := %s.(*runtime.StructInstanceValue); if !ok || v == nil { return false }; if v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil { return false }; return v.Definition.Node.ID.Name == %q }()", subjectTemp, base.Name.Name), true
		}
	}
	return "", false
}

func (g *generator) runtimeTypeCheckExpr(ctx *compileContext, subjectTemp string, goType string) (string, bool) {
	switch g.typeCategory(goType) {
	case "bool":
		return fmt.Sprintf("func() bool { switch v := %s.(type) { case runtime.BoolValue: return true; case *runtime.BoolValue: return v != nil; default: return false } }()", subjectTemp), true
	case "string":
		return fmt.Sprintf("func() bool { switch v := %s.(type) { case runtime.StringValue: return true; case *runtime.StringValue: return v != nil; default: return false } }()", subjectTemp), true
	case "rune":
		return fmt.Sprintf("func() bool { switch v := %s.(type) { case runtime.CharValue: return true; case *runtime.CharValue: return v != nil; default: return false } }()", subjectTemp), true
	case "int", "uint", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64":
		suffix, ok := g.runtimeIntegerSuffix(goType)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("func() bool { switch v := %s.(type) { case runtime.IntegerValue: return v.TypeSuffix == runtime.IntegerType(%q); case *runtime.IntegerValue: return v != nil && v.TypeSuffix == runtime.IntegerType(%q); default: return false } }()", subjectTemp, suffix, suffix), true
	case "float32":
		return fmt.Sprintf("func() bool { switch v := %s.(type) { case runtime.FloatValue: return v.TypeSuffix == runtime.FloatF32; case *runtime.FloatValue: return v != nil && v.TypeSuffix == runtime.FloatF32; default: return false } }()", subjectTemp), true
	case "float64":
		return fmt.Sprintf("func() bool { switch v := %s.(type) { case runtime.FloatValue: return v.TypeSuffix == runtime.FloatF64; case *runtime.FloatValue: return v != nil && v.TypeSuffix == runtime.FloatF64; default: return false } }()", subjectTemp), true
	case "struct":
		info := g.structInfoByGoName(goType)
		if info == nil || info.Name == "" {
			return "", false
		}
		return fmt.Sprintf("func() bool { v, ok := %s.(*runtime.StructInstanceValue); if !ok || v == nil { return false }; if v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil { return false }; return v.Definition.Node.ID.Name == %q }()", subjectTemp, info.Name), true
	default:
		return "", false
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

func (g *generator) compileLiteralMatch(ctx *compileContext, lit ast.Literal, subjectTemp string, subjectType string) (string, bool) {
	if lit == nil {
		ctx.setReason("missing literal pattern")
		return "", false
	}
	if subjectType == "runtime.Value" {
		if _, ok := lit.(*ast.NilLiteral); ok {
			return fmt.Sprintf("__able_is_nil(%s)", subjectTemp), true
		}
		expr, goType, ok := g.compileExpr(ctx, lit.(ast.Expression), "")
		if !ok {
			return "", false
		}
		litRuntime, ok := g.runtimeValueExpr(expr, goType)
		if !ok {
			ctx.setReason("unsupported literal pattern")
			return "", false
		}
		condExpr := fmt.Sprintf("__able_binary_op(%q, %s, %s)", "==", subjectTemp, litRuntime)
		converted, ok := g.expectRuntimeValueExpr(condExpr, "bool")
		if !ok {
			ctx.setReason("unsupported literal comparison")
			return "", false
		}
		return converted, true
	}
	expr, _, ok := g.compileExpr(ctx, lit.(ast.Expression), subjectType)
	if !ok {
		return "", false
	}
	if !g.isEqualityComparable(subjectType) {
		ctx.setReason("unsupported literal comparison")
		return "", false
	}
	return fmt.Sprintf("(%s == %s)", subjectTemp, expr), true
}
