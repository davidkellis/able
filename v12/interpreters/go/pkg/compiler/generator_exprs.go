package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"able/interpreter-go/pkg/ast"
)

// compileExprLines compiles an expression, returning any setup lines separately
// from the final expression value. Callers should emit the lines before using
// the expression. This avoids wrapping in IIFEs.
func (g *generator) compileExprLines(ctx *compileContext, expr ast.Expression, expected string) ([]string, string, string, bool) {
	if expected == "runtime.Value" {
		lines, exprValue, exprType, ok := g.compileExprLines(ctx, expr, "")
		if !ok {
			return nil, "", "", false
		}
		if exprType == "runtime.Value" {
			return lines, exprValue, "runtime.Value", true
		}
		convLines, converted, ok := g.runtimeValueLines(ctx, exprValue, exprType)
		if !ok {
			ctx.setReason("expression type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, "runtime.Value", true
	}
	if value, goType, ok := g.compilePlaceholderLambda(ctx, expr); ok {
		if !g.typeMatches(expected, goType) {
			ctx.setReason("placeholder lambda type mismatch")
			return nil, "", "", false
		}
		return nil, value, goType, true
	}
	switch e := expr.(type) {
	case *ast.AssignmentExpression, *ast.BlockExpression, *ast.IfExpression:
		return g.compileTailExpression(ctx, expected, e)
	case *ast.StructLiteral:
		return g.compileStructLiteral(ctx, e, expected)
	case *ast.ArrayLiteral:
		return g.compileArrayLiteral(ctx, e, expected)
	case *ast.StringInterpolation:
		return g.compileStringInterpolation(ctx, e, expected)
	case *ast.MatchExpression:
		return g.compileMatchExpression(ctx, e, expected)
	case *ast.IndexExpression:
		return g.compileIndexExpression(ctx, e, expected)
	case *ast.LoopExpression:
		return g.compileLoopExpression(ctx, e, expected)
	case *ast.PropagationExpression:
		return g.compilePropagationExpression(ctx, e, expected)
	case *ast.BreakpointExpression:
		return g.compileBreakpointExpression(ctx, e, expected)
	case *ast.BinaryExpression:
		return g.compileBinaryExpression(ctx, e, expected)
	case *ast.Identifier:
		return g.compileIdentifier(ctx, e, expected)
	case *ast.UnaryExpression:
		return g.compileUnaryExpression(ctx, e, expected)
	case *ast.MemberAccessExpression:
		return g.compileMemberAccess(ctx, e, expected)
	case *ast.ImplicitMemberExpression:
		return g.compileImplicitMemberExpression(ctx, e, expected)
	case *ast.FunctionCall:
		return g.compileFunctionCall(ctx, e, expected)
	case *ast.RangeExpression:
		return g.compileRangeExpression(ctx, e, expected)
	case *ast.TypeCastExpression:
		return g.compileTypeCast(ctx, e, expected)
	case *ast.SpawnExpression:
		return g.compileSpawnExpression(ctx, e, expected)
	case *ast.AwaitExpression:
		return g.compileAwaitExpression(ctx, e, expected)
	case *ast.EnsureExpression:
		return g.compileEnsureExpression(ctx, e, expected)
	case *ast.RescueExpression:
		return g.compileRescueExpression(ctx, e, expected)
	case *ast.OrElseExpression:
		return g.compileOrElseExpression(ctx, e, expected)
	default:
		v, t, ok := g.compileExprExpected(ctx, expr, expected)
		return nil, v, t, ok
	}
}

// compileExpr compiles an expression to a single expression string.
// If the expression requires setup lines, they are wrapped in an IIFE.
// Prefer compileExprLines when the caller can emit lines separately.
func (g *generator) compileExpr(ctx *compileContext, expr ast.Expression, expected string) (string, string, bool) {
	lines, v, t, ok := g.compileExprLines(ctx, expr, expected)
	if !ok {
		return "", "", false
	}
	if len(lines) == 0 {
		return v, t, true
	}
	// Lines will be wrapped in an IIFE — temps defined there are scoped to
	// the IIFE and not reachable from the enclosing scope. Clear the CSE
	// extraction cache so subsequent accesses don't reference out-of-scope temps.
	ctx.originExtractions = nil
	return g.wrapLinesAsExpression(ctx, lines, v, t)
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
	case *ast.BooleanLiteral:
		actual := "bool"
		if !g.typeMatches(expected, actual) {
			ctx.setReason("expected bool literal")
			return "", "", false
		}
		return strconv.FormatBool(e.Value), actual, true
	case *ast.NilLiteral:
		if expected == "any" || expected == "" {
			return "nil", "any", true
		}
		if strings.HasPrefix(expected, "*") || strings.HasPrefix(expected, "[]") {
			return "nil", expected, true
		}
		if expected == "runtime.Value" {
			return "runtime.NilValue{}", "runtime.Value", true
		}
		ctx.setReason("nil literal type mismatch")
		return "", "", false
	case *ast.IntegerLiteral:
		return g.compileIntegerLiteral(ctx, e, expected)
	case *ast.FloatLiteral:
		return g.compileFloatLiteral(ctx, e, expected)
	case *ast.CharLiteral:
		return g.compileCharLiteral(ctx, e, expected)
	case *ast.MapLiteral:
		return g.compileMapLiteral(ctx, e, expected)
	case *ast.IteratorLiteral:
		return g.compileIteratorLiteral(ctx, e, expected)
	case *ast.PlaceholderExpression:
		return g.compilePlaceholderExpression(ctx, e, expected)
	case *ast.LambdaExpression:
		return g.compileLambdaExpression(ctx, e, expected)
	default:
		ctx.setReason("unsupported expression")
		return "", "", false
	}
}

func (g *generator) compileStringInterpolation(ctx *compileContext, expr *ast.StringInterpolation, expected string) ([]string, string, string, bool) {
	if expr == nil {
		ctx.setReason("missing string interpolation")
		return nil, "", "", false
	}
	actual := "string"
	if !g.typeMatches(expected, actual) {
		ctx.setReason("string interpolation type mismatch")
		return nil, "", "", false
	}
	if len(expr.Parts) == 0 {
		return nil, "\"\"", actual, true
	}
	lines := make([]string, 0, len(expr.Parts))
	parts := make([]string, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if part == nil {
			ctx.setReason("string interpolation missing part")
			return nil, "", "", false
		}
		if lit, ok := part.(*ast.StringLiteral); ok {
			temp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", temp, strconv.Quote(lit.Value)))
			parts = append(parts, temp)
			continue
		}
		partLines, exprValue, exprType, ok := g.compileExprLines(ctx, part, "")
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, partLines...)
		if exprType == "string" {
			parts = append(parts, exprValue)
			continue
		}
		if stringifyExpr, ok := g.typedStringifyExpr(exprValue, exprType); ok {
			temp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", temp, stringifyExpr))
			parts = append(parts, temp)
			continue
		}
		interpConvLines, runtimeValue, ok := g.runtimeValueLines(ctx, exprValue, exprType)
		if !ok {
			ctx.setReason("string interpolation part unsupported")
			return nil, "", "", false
		}
		lines = append(lines, interpConvLines...)
		temp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_stringify(%s)", temp, runtimeValue))
		parts = append(parts, temp)
	}
	concat := strings.Join(parts, " + ")
	if len(lines) == 0 {
		return nil, concat, actual, true
	}
	return lines, concat, actual, true
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

func (g *generator) compileStructLiteral(ctx *compileContext, lit *ast.StructLiteral, expected string) ([]string, string, string, bool) {
	if lit == nil || lit.StructType == nil {
		ctx.setReason("unsupported struct literal")
		return nil, "", "", false
	}
	info, ok := g.structInfoForTypeName(ctx.packageName, lit.StructType.Name)
	if expected == "runtime.Value" || !ok || info == nil || !info.Supported || len(lit.TypeArguments) > 0 {
		return g.compileStructLiteralRuntime(ctx, lit)
	}
	structType := "*" + info.GoName
	if expected != "" && expected != "any" {
		if expected == "runtime.Value" {
			return g.compileStructLiteralRuntime(ctx, lit)
		}
		baseExpected := expected
		if baseName, ok := g.structBaseName(expected); ok {
			baseExpected = baseName
		}
		if baseExpected != info.GoName {
			ctx.setReason("struct literal type mismatch")
			return nil, "", "", false
		}
	}
	if !info.Supported {
		ctx.setReason("unsupported struct type")
		return nil, "", "", false
	}
	if lit.IsPositional {
		if len(lit.FunctionalUpdateSources) > 0 {
			ctx.setReason("functional update unsupported")
			return nil, "", "", false
		}
		if info.Kind != ast.StructKindPositional && info.Kind != ast.StructKindSingleton {
			ctx.setReason("struct literal positional mismatch")
			return nil, "", "", false
		}
		if len(lit.Fields) != len(info.Fields) {
			ctx.setReason("struct literal missing fields")
			return nil, "", "", false
		}
		var fieldLines []string
		parts := make([]string, 0, len(info.Fields))
		for idx, field := range lit.Fields {
			if field == nil || field.Value == nil || field.Name != nil {
				ctx.setReason("unsupported struct field")
				return nil, "", "", false
			}
			fieldInfo := info.Fields[idx]
			fLines, expr, _, ok := g.compileExprLines(ctx, field.Value, fieldInfo.GoType)
			if !ok {
				return nil, "", "", false
			}
			fieldLines = append(fieldLines, fLines...)
			parts = append(parts, expr)
		}
		return fieldLines, fmt.Sprintf("&%s{%s}", info.GoName, strings.Join(parts, ", ")), structType, true
	}
	updateCount := len(lit.FunctionalUpdateSources)
	if info.Kind == ast.StructKindPositional {
		if updateCount > 0 {
			ctx.setReason("functional update unsupported")
			return nil, "", "", false
		}
		ctx.setReason("struct literal positional mismatch")
		return nil, "", "", false
	}
	if updateCount > 0 {
		if lines, expr, exprType, ok, handled := g.compileStructUpdateFallback(ctx, lit, structType, expected); handled {
			return lines, expr, exprType, ok
		}
	}
	var fieldLines []string
	fieldValues := make(map[string]string, len(lit.Fields))
	for _, field := range lit.Fields {
		if field == nil {
			ctx.setReason("unsupported struct field")
			return nil, "", "", false
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
			return nil, "", "", false
		}
		valueExpr := field.Value
		if valueExpr == nil && field.IsShorthand {
			valueExpr = ast.NewIdentifier(fieldName)
		}
		if valueExpr == nil {
			ctx.setReason("unsupported struct field")
			return nil, "", "", false
		}
		fieldInfo := g.fieldInfo(info, fieldName)
		if fieldInfo == nil {
			ctx.setReason("unknown struct field")
			return nil, "", "", false
		}
		fLines, expr, _, ok := g.compileExprLines(ctx, valueExpr, fieldInfo.GoType)
		if !ok {
			return nil, "", "", false
		}
		fieldLines = append(fieldLines, fLines...)
		fieldValues[fieldInfo.GoName] = expr
	}
	if updateCount == 0 && len(fieldValues) != len(info.Fields) {
		ctx.setReason("struct literal missing fields")
		return nil, "", "", false
	}
	if updateCount > 0 {
		lines := append([]string{}, fieldLines...)
		sourceTemps := make([]string, 0, updateCount)
		for _, source := range lit.FunctionalUpdateSources {
			if source == nil {
				ctx.setReason("functional update source missing")
				return nil, "", "", false
			}
			sLines, expr, _, ok := g.compileExprLines(ctx, source, structType)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, sLines...)
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
		return lines, resultTemp, structType, true
	}
	parts := make([]string, 0, len(info.Fields))
	for _, field := range info.Fields {
		value, ok := fieldValues[field.GoName]
		if !ok {
			ctx.setReason("struct literal missing field values")
			return nil, "", "", false
		}
		parts = append(parts, fmt.Sprintf("%s: %s", field.GoName, value))
	}
	return fieldLines, fmt.Sprintf("&%s{%s}", info.GoName, strings.Join(parts, ", ")), structType, true
}

func (g *generator) compileStructLiteralRuntime(ctx *compileContext, lit *ast.StructLiteral) ([]string, string, string, bool) {
	if lit == nil || lit.StructType == nil || lit.StructType.Name == "" {
		ctx.setReason("unsupported struct literal")
		return nil, "", "", false
	}
	structName := lit.StructType.Name
	typeArgsExpr := "[]ast.TypeExpression(nil)"
	if len(lit.TypeArguments) > 0 {
		args := make([]string, 0, len(lit.TypeArguments))
		for _, arg := range lit.TypeArguments {
			rendered, ok := g.renderTypeExpression(arg)
			if !ok {
				ctx.setReason("unsupported struct literal type arguments")
				return nil, "", "", false
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
				return nil, "", "", false
			}
			fLines, expr, _, ok := g.compileExprLines(ctx, field.Value, "runtime.Value")
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, fLines...)
			values = append(values, expr)
		}
		valuesTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := []runtime.Value{%s}", valuesTemp, strings.Join(values, ", ")))
		lines = append(lines, fmt.Sprintf("if len(%s) != len(%s.Fields) { panic(fmt.Errorf(\"Struct '%s' expects %%d fields, got %%d\", len(%s.Fields), len(%s))) }", valuesTemp, structDefTemp, structName, structDefTemp, valuesTemp))
		resultTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := &runtime.StructInstanceValue{Definition: %s, Positional: %s, TypeArguments: %s}", resultTemp, defTemp, valuesTemp, typeArgsExpr))
		return lines, resultTemp, "runtime.Value", true
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
				return nil, "", "", false
			}
			sLines, sourceExpr, _, ok := g.compileExprLines(ctx, source, "runtime.Value")
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, sLines...)
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
			return nil, "", "", false
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
			return nil, "", "", false
		}
		valueExpr := field.Value
		if valueExpr == nil && field.IsShorthand {
			valueExpr = ast.NewIdentifier(fieldName)
		}
		if valueExpr == nil {
			ctx.setReason("unsupported struct field")
			return nil, "", "", false
		}
		fLines, expr, _, ok := g.compileExprLines(ctx, valueExpr, "runtime.Value")
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, fLines...)
		lines = append(lines, fmt.Sprintf("%s[%q] = %s", fieldsTemp, fieldName, expr))
	}

	lines = append(lines, fmt.Sprintf("if %s.Kind == %q { for _, defField := range %s.Fields { if defField == nil || defField.Name == nil { continue }; if _, ok := %s[defField.Name.Name]; !ok { panic(fmt.Errorf(\"Missing field '%%s' for struct '%s'\", defField.Name.Name)) } } }", structDefTemp, "named", structDefTemp, fieldsTemp, structName))

	typeArgsTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", typeArgsTemp, typeArgsExpr))
	if updateCount > 0 {
		lines = append(lines, fmt.Sprintf("if len(%s) == 0 && %s != nil { %s = %s.TypeArguments }", typeArgsTemp, baseTemp, typeArgsTemp, baseTemp))
	}
	resultTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := &runtime.StructInstanceValue{Definition: %s, Fields: %s, TypeArguments: %s}", resultTemp, defTemp, fieldsTemp, typeArgsTemp))
	return lines, resultTemp, "runtime.Value", true
}

func (g *generator) compileUnaryExpression(ctx *compileContext, expr *ast.UnaryExpression, expected string) ([]string, string, string, bool) {
	if expr == nil {
		ctx.setReason("missing unary expression")
		return nil, "", "", false
	}
	switch expr.Operator {
	case ast.UnaryOperatorNegate:
		operandLines, operand, operandType, ok := g.compileExprLines(ctx, expr.Operand, expected)
		if !ok {
			return nil, "", "", false
		}
		if g.isIntegerType(operandType) {
			if !g.typeMatches(expected, operandType) {
				ctx.setReason("unary expression type mismatch")
				return nil, "", "", false
			}
			nodeName := g.diagNodeName(expr, "*ast.UnaryExpression", "unary")
			temp := ctx.newTemp()
			bitsExpr := g.bitSizeExpr(operandType)
			lines := append([]string{}, operandLines...)
			lines = append(lines, fmt.Sprintf("%s := %s", temp, operand))
			if g.isUnsignedIntegerType(operandType) {
				return lines, fmt.Sprintf("%s(__able_checked_sub_unsigned(uint64(0), uint64(%s), %s, %s))", operandType, temp, bitsExpr, nodeName), operandType, true
			}
			return lines, fmt.Sprintf("%s(__able_checked_sub_signed(int64(0), int64(%s), %s, %s))", operandType, temp, bitsExpr, nodeName), operandType, true
		}
		if !g.isNumericType(operandType) {
			opConvLines, operandRuntime, ok := g.runtimeValueLines(ctx, operand, operandType)
			if !ok {
				ctx.setReason("unsupported unary operand type")
				return nil, "", "", false
			}
			operandLines = append(operandLines, opConvLines...)
			unaryExpr := fmt.Sprintf("__able_unary_op(%q, %s)", string(expr.Operator), operandRuntime)
			if expected == "" || expected == "runtime.Value" {
				return operandLines, unaryExpr, "runtime.Value", true
			}
			convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, unaryExpr, expected)
			if !ok {
				ctx.setReason("unary expression type mismatch")
				return nil, "", "", false
			}
			lines := append([]string{}, operandLines...)
			lines = append(lines, convLines...)
			return lines, converted, expected, true
		}
		if !g.typeMatches(expected, operandType) {
			ctx.setReason("unary expression type mismatch")
			return nil, "", "", false
		}
		return operandLines, fmt.Sprintf("(-%s)", operand), operandType, true
	case ast.UnaryOperatorNot:
		if expected != "" && expected != "bool" {
			ctx.setReason("unary expression type mismatch")
			return nil, "", "", false
		}
		operandLines, operand, operandType, ok := g.compileExprLines(ctx, expr.Operand, "")
		if !ok {
			return nil, "", "", false
		}
		if operandType == "bool" {
			return operandLines, fmt.Sprintf("(!%s)", operand), "bool", true
		}
		operandRuntime := operand
		if operandType != "runtime.Value" {
			convLines, converted, ok := g.runtimeValueLines(ctx, operand, operandType)
			if !ok {
				ctx.setReason("unsupported unary operand type")
				return nil, "", "", false
			}
			operandLines = append(operandLines, convLines...)
			operandRuntime = converted
		}
		return operandLines, fmt.Sprintf("!__able_truthy(%s)", operandRuntime), "bool", true
	case ast.UnaryOperatorBitNot:
		operandLines, operand, operandType, ok := g.compileExprLines(ctx, expr.Operand, expected)
		if !ok {
			return nil, "", "", false
		}
		if !g.isIntegerType(operandType) {
			opConvLines, operandRuntime, ok := g.runtimeValueLines(ctx, operand, operandType)
			if !ok {
				ctx.setReason("unsupported bitwise operand type")
				return nil, "", "", false
			}
			operandLines = append(operandLines, opConvLines...)
			unaryExpr := fmt.Sprintf("__able_unary_op(%q, %s)", string(expr.Operator), operandRuntime)
			if expected == "" || expected == "runtime.Value" {
				return operandLines, unaryExpr, "runtime.Value", true
			}
			convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, unaryExpr, expected)
			if !ok {
				ctx.setReason("unary expression type mismatch")
				return nil, "", "", false
			}
			lines := append([]string{}, operandLines...)
			lines = append(lines, convLines...)
			return lines, converted, expected, true
		}
		if !g.typeMatches(expected, operandType) {
			ctx.setReason("unary expression type mismatch")
			return nil, "", "", false
		}
		return operandLines, fmt.Sprintf("(^%s)", operand), operandType, true
	default:
		ctx.setReason("unsupported unary operator")
		return nil, "", "", false
	}
}
