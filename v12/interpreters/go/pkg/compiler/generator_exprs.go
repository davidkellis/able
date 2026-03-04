package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileExpr(ctx *compileContext, expr ast.Expression, expected string) (string, string, bool) {
	if expected == "runtime.Value" {
		exprValue, exprType, ok := g.compileExpr(ctx, expr, "")
		if !ok {
			return "", "", false
		}
		if exprType == "runtime.Value" {
			return exprValue, "runtime.Value", true
		}
		converted, ok := g.runtimeValueExpr(exprValue, exprType)
		if !ok {
			ctx.setReason("expression type mismatch")
			return "", "", false
		}
		return converted, "runtime.Value", true
	}
	return g.compileExprExpected(ctx, expr, expected)
}

func (g *generator) compileExprExpected(ctx *compileContext, expr ast.Expression, expected string) (string, string, bool) {
	if value, goType, ok := g.compilePlaceholderLambda(ctx, expr); ok {
		if !g.typeMatches(expected, goType) {
			ctx.setReason("placeholder lambda type mismatch")
			return "", "", false
		}
		return value, goType, true
	}
	switch e := expr.(type) {
	case *ast.StringLiteral:
		actual := "string"
		if !g.typeMatches(expected, actual) {
			ctx.setReason("expected string literal")
			return "", "", false
		}
		return strconv.Quote(e.Value), actual, true
	case *ast.StringInterpolation:
		return g.compileStringInterpolation(ctx, e, expected)
	case *ast.BooleanLiteral:
		actual := "bool"
		if !g.typeMatches(expected, actual) {
			ctx.setReason("expected bool literal")
			return "", "", false
		}
		return strconv.FormatBool(e.Value), actual, true
	case *ast.NilLiteral:
		actual := "runtime.Value"
		if expected != "" && expected != actual {
			ctx.setReason("nil literal type mismatch")
			return "", "", false
		}
		return "runtime.NilValue{}", actual, true
	case *ast.IntegerLiteral:
		return g.compileIntegerLiteral(ctx, e, expected)
	case *ast.FloatLiteral:
		return g.compileFloatLiteral(ctx, e, expected)
	case *ast.CharLiteral:
		return g.compileCharLiteral(ctx, e, expected)
	case *ast.StructLiteral:
		return g.compileStructLiteral(ctx, e, expected)
	case *ast.ArrayLiteral:
		return g.compileArrayLiteral(ctx, e, expected)
	case *ast.MapLiteral:
		return g.compileMapLiteral(ctx, e, expected)
	case *ast.IteratorLiteral:
		return g.compileIteratorLiteral(ctx, e, expected)
	case *ast.Identifier:
		return g.compileIdentifier(ctx, e, expected)
	case *ast.PlaceholderExpression:
		return g.compilePlaceholderExpression(ctx, e, expected)
	case *ast.ImplicitMemberExpression:
		return g.compileImplicitMemberExpression(ctx, e, expected)
	case *ast.UnaryExpression:
		return g.compileUnaryExpression(ctx, e, expected)
	case *ast.BinaryExpression:
		return g.compileBinaryExpression(ctx, e, expected)
	case *ast.FunctionCall:
		return g.compileFunctionCall(ctx, e, expected)
	case *ast.MemberAccessExpression:
		return g.compileMemberAccess(ctx, e, expected)
	case *ast.IndexExpression:
		return g.compileIndexExpression(ctx, e, expected)
	case *ast.RangeExpression:
		return g.compileRangeExpression(ctx, e, expected)
	case *ast.TypeCastExpression:
		return g.compileTypeCast(ctx, e, expected)
	case *ast.LambdaExpression:
		return g.compileLambdaExpression(ctx, e, expected)
	case *ast.RescueExpression:
		return g.compileRescueExpression(ctx, e, expected)
	case *ast.EnsureExpression:
		return g.compileEnsureExpression(ctx, e, expected)
	case *ast.SpawnExpression:
		return g.compileSpawnExpression(ctx, e, expected)
	case *ast.AwaitExpression:
		return g.compileAwaitExpression(ctx, e, expected)
	case *ast.PropagationExpression:
		return g.compilePropagationExpression(ctx, e, expected)
	case *ast.OrElseExpression:
		return g.compileOrElseExpression(ctx, e, expected)
	case *ast.MatchExpression:
		return g.compileMatchExpression(ctx, e, expected)
	case *ast.LoopExpression:
		return g.compileLoopExpression(ctx, e, expected)
	case *ast.BreakpointExpression:
		return g.compileBreakpointExpression(ctx, e, expected)
	case *ast.AssignmentExpression, *ast.BlockExpression, *ast.IfExpression:
		lines, exprValue, exprType, ok := g.compileTailExpression(ctx, expected, e)
		if !ok {
			return "", "", false
		}
		return g.wrapLinesAsExpression(ctx, lines, exprValue, exprType)
	default:
		ctx.setReason("unsupported expression")
		return "", "", false
	}
}

func (g *generator) compileStringInterpolation(ctx *compileContext, expr *ast.StringInterpolation, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing string interpolation")
		return "", "", false
	}
	actual := "string"
	if !g.typeMatches(expected, actual) {
		ctx.setReason("string interpolation type mismatch")
		return "", "", false
	}
	if len(expr.Parts) == 0 {
		return "\"\"", actual, true
	}
	lines := make([]string, 0, len(expr.Parts))
	parts := make([]string, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if part == nil {
			ctx.setReason("string interpolation missing part")
			return "", "", false
		}
		if lit, ok := part.(*ast.StringLiteral); ok {
			temp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", temp, strconv.Quote(lit.Value)))
			parts = append(parts, temp)
			continue
		}
		exprValue, exprType, ok := g.compileExpr(ctx, part, "")
		if !ok {
			return "", "", false
		}
		runtimeValue, ok := g.runtimeValueExpr(exprValue, exprType)
		if !ok {
			ctx.setReason("string interpolation part unsupported")
			return "", "", false
		}
		temp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_stringify(%s)", temp, runtimeValue))
		parts = append(parts, temp)
	}
	concat := strings.Join(parts, " + ")
	if len(lines) == 0 {
		return concat, actual, true
	}
	return fmt.Sprintf("func() string { %s; return %s }()", strings.Join(lines, "; "), concat), actual, true
}

func (g *generator) compileIntegerLiteral(ctx *compileContext, lit *ast.IntegerLiteral, expected string) (string, string, bool) {
	if lit == nil || lit.Value == nil {
		ctx.setReason("missing integer literal")
		return "", "", false
	}
	actual := g.inferIntegerLiteralType(lit)
	explicit := lit.IntegerType != nil
	if expected == "" {
		expected = actual
	}
	if expected == "runtime.Value" {
		literalText := lit.Value.String()
		return fmt.Sprintf(
			"func() runtime.Value { val, ok := new(big.Int).SetString(%q, 10); if !ok { panic(fmt.Errorf(\"invalid integer literal: %%s\", %q)) }; return runtime.NewBigIntValue(val, %s) }()",
			literalText,
			literalText,
			integerSuffix(lit),
		), "runtime.Value", true
	}
	if explicit && expected != actual {
		ctx.setReason("integer literal type mismatch")
		return "", "", false
	}
	if g.isFloatType(expected) {
		if explicit {
			ctx.setReason("integer literal type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("%s(%s)", expected, lit.Value.String()), expected, true
	}
	if !g.typeMatches(expected, actual) && !g.isIntegerType(expected) {
		ctx.setReason(fmt.Sprintf("unsupported integer literal type (%s)", expected))
		return "", "", false
	}
	return fmt.Sprintf("%s(%s)", expected, lit.Value.String()), expected, true
}

func (g *generator) compileFloatLiteral(ctx *compileContext, lit *ast.FloatLiteral, expected string) (string, string, bool) {
	if lit == nil {
		ctx.setReason("missing float literal")
		return "", "", false
	}
	actual := g.inferFloatLiteralType(lit)
	explicit := lit.FloatType != nil
	if expected == "" {
		expected = actual
	}
	if explicit && expected != actual {
		ctx.setReason("float literal type mismatch")
		return "", "", false
	}
	if !g.typeMatches(expected, actual) && !g.isFloatType(expected) {
		ctx.setReason("unsupported float literal type")
		return "", "", false
	}
	return fmt.Sprintf("%s(%s)", expected, strconv.FormatFloat(lit.Value, 'g', -1, 64)), expected, true
}

func (g *generator) compileStructLiteral(ctx *compileContext, lit *ast.StructLiteral, expected string) (string, string, bool) {
	if lit == nil || lit.StructType == nil {
		ctx.setReason("unsupported struct literal")
		return "", "", false
	}
	info, ok := g.structInfoForTypeName(ctx.packageName, lit.StructType.Name)
	if expected == "runtime.Value" || !ok || info == nil || !info.Supported || len(lit.TypeArguments) > 0 {
		return g.compileStructLiteralRuntime(ctx, lit)
	}
	structType := "*" + info.GoName
	if expected != "" {
		if expected == "runtime.Value" {
			return g.compileStructLiteralRuntime(ctx, lit)
		}
		baseExpected := expected
		if baseName, ok := g.structBaseName(expected); ok {
			baseExpected = baseName
		}
		if baseExpected != info.GoName {
			ctx.setReason("struct literal type mismatch")
			return "", "", false
		}
	}
	if !info.Supported {
		ctx.setReason("unsupported struct type")
		return "", "", false
	}
	if lit.IsPositional {
		if len(lit.FunctionalUpdateSources) > 0 {
			ctx.setReason("functional update unsupported")
			return "", "", false
		}
		if info.Kind != ast.StructKindPositional && info.Kind != ast.StructKindSingleton {
			ctx.setReason("struct literal positional mismatch")
			return "", "", false
		}
		if len(lit.Fields) != len(info.Fields) {
			ctx.setReason("struct literal missing fields")
			return "", "", false
		}
		parts := make([]string, 0, len(info.Fields))
		for idx, field := range lit.Fields {
			if field == nil || field.Value == nil || field.Name != nil {
				ctx.setReason("unsupported struct field")
				return "", "", false
			}
			fieldInfo := info.Fields[idx]
			expr, _, ok := g.compileExpr(ctx, field.Value, fieldInfo.GoType)
			if !ok {
				return "", "", false
			}
			parts = append(parts, expr)
		}
		return fmt.Sprintf("&%s{%s}", info.GoName, strings.Join(parts, ", ")), structType, true
	}
	updateCount := len(lit.FunctionalUpdateSources)
	if info.Kind == ast.StructKindPositional {
		if updateCount > 0 {
			ctx.setReason("functional update unsupported")
			return "", "", false
		}
		ctx.setReason("struct literal positional mismatch")
		return "", "", false
	}
	if updateCount > 0 {
		if expr, exprType, ok, handled := g.compileStructUpdateFallback(ctx, lit, structType, expected); handled {
			return expr, exprType, ok
		}
	}
	fieldValues := make(map[string]string, len(lit.Fields))
	for _, field := range lit.Fields {
		if field == nil {
			ctx.setReason("unsupported struct field")
			return "", "", false
		}
		fieldName := ""
		if field.Name != nil {
			fieldName = field.Name.Name
		}
		if fieldName == "" && field.IsShorthand {
			if ident, ok := field.Value.(*ast.Identifier); ok && ident != nil {
				fieldName = ident.Name
			}
		}
		if fieldName == "" {
			ctx.setReason("unsupported struct field")
			return "", "", false
		}
		valueExpr := field.Value
		if valueExpr == nil && field.IsShorthand {
			valueExpr = ast.NewIdentifier(fieldName)
		}
		if valueExpr == nil {
			ctx.setReason("unsupported struct field")
			return "", "", false
		}
		fieldInfo := g.fieldInfo(info, fieldName)
		if fieldInfo == nil {
			ctx.setReason("unknown struct field")
			return "", "", false
		}
		expr, _, ok := g.compileExpr(ctx, valueExpr, fieldInfo.GoType)
		if !ok {
			return "", "", false
		}
		fieldValues[fieldInfo.GoName] = expr
	}
	if updateCount == 0 && len(fieldValues) != len(info.Fields) {
		ctx.setReason("struct literal missing fields")
		return "", "", false
	}
	if updateCount > 0 {
		lines := []string{}
		sourceTemps := make([]string, 0, updateCount)
		for _, source := range lit.FunctionalUpdateSources {
			if source == nil {
				ctx.setReason("functional update source missing")
				return "", "", false
			}
			expr, _, ok := g.compileExpr(ctx, source, structType)
			if !ok {
				return "", "", false
			}
			temp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", temp, expr))
			sourceTemps = append(sourceTemps, temp)
		}
		resultTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := &%s{}", resultTemp, info.GoName))
		lines = append(lines, fmt.Sprintf("*%s = *%s", resultTemp, sourceTemps[len(sourceTemps)-1]))
		for _, field := range info.Fields {
			value, ok := fieldValues[field.GoName]
			if !ok {
				continue
			}
			lines = append(lines, fmt.Sprintf("%s.%s = %s", resultTemp, field.GoName, value))
		}
		exprValue := fmt.Sprintf("func() %s { %s; return %s }()", structType, strings.Join(lines, "; "), resultTemp)
		return exprValue, structType, true
	}
	parts := make([]string, 0, len(info.Fields))
	for _, field := range info.Fields {
		value, ok := fieldValues[field.GoName]
		if !ok {
			ctx.setReason("struct literal missing field values")
			return "", "", false
		}
		parts = append(parts, fmt.Sprintf("%s: %s", field.GoName, value))
	}
	return fmt.Sprintf("&%s{%s}", info.GoName, strings.Join(parts, ", ")), structType, true
}

func (g *generator) compileStructLiteralRuntime(ctx *compileContext, lit *ast.StructLiteral) (string, string, bool) {
	if lit == nil || lit.StructType == nil || lit.StructType.Name == "" {
		ctx.setReason("unsupported struct literal")
		return "", "", false
	}
	structName := lit.StructType.Name
	typeArgsExpr := "[]ast.TypeExpression(nil)"
	if len(lit.TypeArguments) > 0 {
		args := make([]string, 0, len(lit.TypeArguments))
		for _, arg := range lit.TypeArguments {
			rendered, ok := g.renderTypeExpression(arg)
			if !ok {
				ctx.setReason("unsupported struct literal type arguments")
				return "", "", false
			}
			args = append(args, rendered)
		}
		typeArgsExpr = fmt.Sprintf("[]ast.TypeExpression{%s}", strings.Join(args, ", "))
	}

	lines := []string{
		"if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }",
	}
	defTemp := ctx.newTemp()
	structDefTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, err := __able_runtime.StructDefinition(%q)", defTemp, structName))
	lines = append(lines, "__able_panic_on_error(err)")
	lines = append(lines, fmt.Sprintf("if %s == nil || %s.Node == nil || %s.Node.ID == nil { panic(fmt.Errorf(\"struct definition '%s' unavailable\")) }", defTemp, defTemp, defTemp, structName))
	lines = append(lines, fmt.Sprintf("%s := %s.Node", structDefTemp, defTemp))

	updateCount := len(lit.FunctionalUpdateSources)
	if lit.IsPositional {
		if updateCount > 0 {
			lines = append(lines, "panic(fmt.Errorf(\"Functional update only supported for named structs\"))")
		}
		lines = append(lines, fmt.Sprintf("if %s.Kind != %q && %s.Kind != %q { panic(fmt.Errorf(\"Positional struct literal not allowed for struct '%s'\")) }", structDefTemp, "positional", structDefTemp, "singleton", structName))
		values := make([]string, 0, len(lit.Fields))
		for _, field := range lit.Fields {
			if field == nil || field.Value == nil {
				ctx.setReason("unsupported struct field")
				return "", "", false
			}
			expr, _, ok := g.compileExpr(ctx, field.Value, "runtime.Value")
			if !ok {
				return "", "", false
			}
			values = append(values, expr)
		}
		valuesTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := []runtime.Value{%s}", valuesTemp, strings.Join(values, ", ")))
		lines = append(lines, fmt.Sprintf("if len(%s) != len(%s.Fields) { panic(fmt.Errorf(\"Struct '%s' expects %%d fields, got %%d\", len(%s.Fields), len(%s))) }", valuesTemp, structDefTemp, structName, structDefTemp, valuesTemp))
		expr := fmt.Sprintf("&runtime.StructInstanceValue{Definition: %s, Positional: %s, TypeArguments: %s}", defTemp, valuesTemp, typeArgsExpr)
		return g.wrapLinesAsExpression(ctx, lines, expr, "runtime.Value")
	}

	if updateCount == 0 {
		lines = append(lines, fmt.Sprintf("if %s.Kind == %q { panic(fmt.Errorf(\"Named struct literal not allowed for positional struct '%s'\")) }", structDefTemp, "positional", structName))
	} else {
		lines = append(lines, fmt.Sprintf("if %s.Kind == %q { panic(fmt.Errorf(\"Functional update only supported for named structs\")) }", structDefTemp, "positional"))
	}

	fieldsTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := make(map[string]runtime.Value, %d)", fieldsTemp, len(lit.Fields)))
	var baseTemp string
	if updateCount > 0 {
		baseTemp = ctx.newTemp()
		lines = append(lines, fmt.Sprintf("var %s *runtime.StructInstanceValue", baseTemp))
		for _, source := range lit.FunctionalUpdateSources {
			if source == nil {
				ctx.setReason("functional update source missing")
				return "", "", false
			}
			sourceExpr, _, ok := g.compileExpr(ctx, source, "runtime.Value")
			if !ok {
				return "", "", false
			}
			sourceTemp := ctx.newTemp()
			instanceTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", sourceTemp, sourceExpr))
			lines = append(lines, fmt.Sprintf("%s := __able_struct_instance(%s)", instanceTemp, sourceTemp))
			lines = append(lines, fmt.Sprintf("if %s == nil { panic(fmt.Errorf(\"Functional update source must be a struct instance\")) }", instanceTemp))
			lines = append(lines, fmt.Sprintf("if %s.Definition == nil || %s.Definition.Node == nil || %s.Definition.Node.ID == nil || %s.Definition.Node.ID.Name != %q { panic(fmt.Errorf(\"Functional update source must be same struct type\")) }", instanceTemp, instanceTemp, instanceTemp, instanceTemp, structName))
			lines = append(lines, fmt.Sprintf("if %s.Fields == nil { panic(fmt.Errorf(\"Functional update only supported for named structs\")) }", instanceTemp))
			lines = append(lines, fmt.Sprintf("if %s == nil { %s = %s }", baseTemp, baseTemp, instanceTemp))
			lines = append(lines, fmt.Sprintf("for k, v := range %s.Fields { %s[k] = v }", instanceTemp, fieldsTemp))
		}
	}

	for _, field := range lit.Fields {
		if field == nil {
			ctx.setReason("unsupported struct field")
			return "", "", false
		}
		fieldName := ""
		if field.Name != nil {
			fieldName = field.Name.Name
		}
		if fieldName == "" && field.IsShorthand {
			if ident, ok := field.Value.(*ast.Identifier); ok && ident != nil {
				fieldName = ident.Name
			}
		}
		if fieldName == "" {
			ctx.setReason("unsupported struct field")
			return "", "", false
		}
		valueExpr := field.Value
		if valueExpr == nil && field.IsShorthand {
			valueExpr = ast.NewIdentifier(fieldName)
		}
		if valueExpr == nil {
			ctx.setReason("unsupported struct field")
			return "", "", false
		}
		expr, _, ok := g.compileExpr(ctx, valueExpr, "runtime.Value")
		if !ok {
			return "", "", false
		}
		lines = append(lines, fmt.Sprintf("%s[%q] = %s", fieldsTemp, fieldName, expr))
	}

	lines = append(lines, fmt.Sprintf("if %s.Kind == %q { for _, defField := range %s.Fields { if defField == nil || defField.Name == nil { continue }; if _, ok := %s[defField.Name.Name]; !ok { panic(fmt.Errorf(\"Missing field '%%s' for struct '%s'\", defField.Name.Name)) } } }", structDefTemp, "named", structDefTemp, fieldsTemp, structName))

	typeArgsTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", typeArgsTemp, typeArgsExpr))
	if updateCount > 0 {
		lines = append(lines, fmt.Sprintf("if len(%s) == 0 && %s != nil { %s = %s.TypeArguments }", typeArgsTemp, baseTemp, typeArgsTemp, baseTemp))
	}
	expr := fmt.Sprintf("&runtime.StructInstanceValue{Definition: %s, Fields: %s, TypeArguments: %s}", defTemp, fieldsTemp, typeArgsTemp)
	return g.wrapLinesAsExpression(ctx, lines, expr, "runtime.Value")
}

func (g *generator) compileUnaryExpression(ctx *compileContext, expr *ast.UnaryExpression, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing unary expression")
		return "", "", false
	}
	switch expr.Operator {
	case ast.UnaryOperatorNegate:
		operand, operandType, ok := g.compileExpr(ctx, expr.Operand, expected)
		if !ok {
			return "", "", false
		}
		if g.isIntegerType(operandType) {
			if !g.typeMatches(expected, operandType) {
				ctx.setReason("unary expression type mismatch")
				return "", "", false
			}
			nodeName := g.diagNodeName(expr, "*ast.UnaryExpression", "unary")
			temp := ctx.newTemp()
			bitsExpr := g.bitSizeExpr(operandType)
			if g.isUnsignedIntegerType(operandType) {
				expr := fmt.Sprintf("func() %s { %s := %s; return %s(__able_checked_sub_unsigned(uint64(0), uint64(%s), %s, %s)) }()", operandType, temp, operand, operandType, temp, bitsExpr, nodeName)
				return expr, operandType, true
			}
			expr := fmt.Sprintf("func() %s { %s := %s; return %s(__able_checked_sub_signed(int64(0), int64(%s), %s, %s)) }()", operandType, temp, operand, operandType, temp, bitsExpr, nodeName)
			return expr, operandType, true
		}
		if !g.isNumericType(operandType) {
			operandRuntime, ok := g.runtimeValueExpr(operand, operandType)
			if !ok {
				ctx.setReason("unsupported unary operand type")
				return "", "", false
			}
			unaryExpr := fmt.Sprintf("__able_unary_op(%q, %s)", string(expr.Operator), operandRuntime)
			if expected == "" || expected == "runtime.Value" {
				return unaryExpr, "runtime.Value", true
			}
			converted, ok := g.expectRuntimeValueExpr(unaryExpr, expected)
			if !ok {
				ctx.setReason("unary expression type mismatch")
				return "", "", false
			}
			return converted, expected, true
		}
		if !g.typeMatches(expected, operandType) {
			ctx.setReason("unary expression type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("(-%s)", operand), operandType, true
	case ast.UnaryOperatorNot:
		if expected != "" && expected != "bool" {
			ctx.setReason("unary expression type mismatch")
			return "", "", false
		}
		operand, operandType, ok := g.compileExpr(ctx, expr.Operand, "")
		if !ok {
			return "", "", false
		}
		if operandType == "bool" {
			return fmt.Sprintf("(!%s)", operand), "bool", true
		}
		operandRuntime := operand
		if operandType != "runtime.Value" {
			converted, ok := g.runtimeValueExpr(operand, operandType)
			if !ok {
				ctx.setReason("unsupported unary operand type")
				return "", "", false
			}
			operandRuntime = converted
		}
		return fmt.Sprintf("!__able_truthy(%s)", operandRuntime), "bool", true
	case ast.UnaryOperatorBitNot:
		operand, operandType, ok := g.compileExpr(ctx, expr.Operand, expected)
		if !ok {
			return "", "", false
		}
		if !g.isIntegerType(operandType) {
			operandRuntime, ok := g.runtimeValueExpr(operand, operandType)
			if !ok {
				ctx.setReason("unsupported bitwise operand type")
				return "", "", false
			}
			unaryExpr := fmt.Sprintf("__able_unary_op(%q, %s)", string(expr.Operator), operandRuntime)
			if expected == "" || expected == "runtime.Value" {
				return unaryExpr, "runtime.Value", true
			}
			converted, ok := g.expectRuntimeValueExpr(unaryExpr, expected)
			if !ok {
				ctx.setReason("unary expression type mismatch")
				return "", "", false
			}
			return converted, expected, true
		}
		if !g.typeMatches(expected, operandType) {
			ctx.setReason("unary expression type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("(^%s)", operand), operandType, true
	default:
		ctx.setReason("unsupported unary operator")
		return "", "", false
	}
}
