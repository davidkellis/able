package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

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

func (g *generator) compileNativeUnionLiteralMatch(ctx *compileContext, lit ast.Literal, subjectTemp string, subjectType string) ([]string, string, bool) {
	if g == nil || ctx == nil || lit == nil {
		return nil, "", false
	}
	union := g.nativeUnionInfoForGoType(subjectType)
	if union == nil {
		return nil, "", false
	}
	if _, isNil := lit.(*ast.NilLiteral); isNil {
		var (
			lines []string
			conds []string
		)
		for _, member := range union.Members {
			if member == nil || member.TypeExpr == nil || !g.typeExprIncludesNilInPackage(union.PackageName, member.TypeExpr) {
				continue
			}
			memberTemp := ctx.newTemp()
			okTemp := ctx.newTemp()
			memberLines := []string{fmt.Sprintf("%s, %s := %s(%s)", memberTemp, okTemp, member.UnwrapHelper, subjectTemp)}
			if g.nativeUnionInfoForGoType(member.GoType) != nil {
				innerLines, innerCond, ok := g.compileNativeUnionLiteralMatch(ctx, lit, memberTemp, member.GoType)
				if !ok {
					continue
				}
				guardedLines, cond, ok := g.guardMatchConditionWithPredicate(ctx, okTemp, innerLines, innerCond)
				if !ok {
					continue
				}
				lines = append(lines, memberLines...)
				lines = append(lines, guardedLines...)
				conds = append(conds, cond)
				continue
			}
			var nilCond string
			switch {
			case member.GoType == "runtime.Value":
				nilCond = fmt.Sprintf("__able_is_nil(%s)", memberTemp)
			case member.GoType == "any":
				nilCond = fmt.Sprintf("(%s == nil)", memberTemp)
			case g.goTypeHasNilZeroValue(member.GoType):
				nilCond = fmt.Sprintf("(%s == nil)", memberTemp)
			default:
				continue
			}
			guardedLines, cond, ok := g.guardMatchConditionWithPredicate(ctx, okTemp, nil, nilCond)
			if !ok {
				continue
			}
			lines = append(lines, memberLines...)
			lines = append(lines, guardedLines...)
			conds = append(conds, cond)
		}
		if len(conds) != 0 {
			return lines, fmt.Sprintf("(%s)", strings.Join(conds, " || ")), true
		}
		if g.goTypeHasNilZeroValue(subjectType) {
			return nil, fmt.Sprintf("(%s == nil)", subjectTemp), true
		}
		return nil, "false", true
	}
	litExpr, ok := lit.(ast.Expression)
	if !ok || litExpr == nil {
		return nil, "", false
	}
	tryMember := func(member *nativeUnionMember, expr string, exprType string, lines []string) ([]string, string, bool) {
		if member == nil || member.GoType == "" || member.GoType != exprType || !g.isEqualityComparable(member.GoType) {
			return nil, "", false
		}
		memberTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		out := append([]string{}, lines...)
		out = append(out, fmt.Sprintf("%s, %s := %s(%s)", memberTemp, okTemp, member.UnwrapHelper, subjectTemp))
		return out, fmt.Sprintf("(%s && %s == %s)", okTemp, memberTemp, expr), true
	}
	litLines, litValue, litType, ok := g.compileExprLines(ctx.probeChild(), litExpr, "")
	if ok {
		if member, memberOK := g.nativeUnionMember(union, litType); memberOK {
			if lines, cond, ok := tryMember(member, litValue, litType, litLines); ok {
				return lines, cond, true
			}
		}
	}
	for _, member := range union.Members {
		if member == nil || member.GoType == "" || member.GoType == "runtime.Value" || !g.isEqualityComparable(member.GoType) {
			continue
		}
		memberCtx := ctx.probeChild()
		memberLines, memberValue, memberType, ok := g.compileExprLines(memberCtx, litExpr, member.GoType)
		if !ok {
			continue
		}
		if lines, cond, ok := tryMember(member, memberValue, memberType, memberLines); ok {
			return lines, cond, true
		}
	}
	return nil, "false", true
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
		litConvLines, litRuntime, ok := g.lowerRuntimeValue(ctx, expr, goType)
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
		condTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s, %s := __able_binary_op(%q, %s, %s)", condTemp, controlTemp, "==", effectiveSubject, litRuntime))
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			ctx.setReason("unsupported literal comparison")
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		expectLines, converted, ok := g.lowerExpectRuntimeValue(ctx, condTemp, "bool")
		if !ok {
			ctx.setReason("unsupported literal comparison")
			return nil, "", false
		}
		lines = append(lines, expectLines...)
		return lines, converted, true
	}
	if innerType, ok := g.nativeNullableValueInnerType(subjectType); ok {
		if _, isNil := lit.(*ast.NilLiteral); isNil {
			return nil, fmt.Sprintf("(%s == nil)", subjectTemp), true
		}
		litLines, expr, _, ok := g.compileExprLines(ctx, lit.(ast.Expression), innerType)
		if !ok {
			return nil, "", false
		}
		return litLines, fmt.Sprintf("(%s != nil && (*%s == %s))", subjectTemp, subjectTemp, expr), true
	}
	if lines, cond, ok := g.compileNativeUnionLiteralMatch(ctx, lit, subjectTemp, subjectType); ok {
		return lines, cond, true
	}
	if _, ok := lit.(*ast.NilLiteral); ok && g.goTypeHasNilZeroValue(subjectType) {
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
