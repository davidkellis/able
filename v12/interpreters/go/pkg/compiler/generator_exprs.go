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
			"func() runtime.Value { val, ok := new(big.Int).SetString(%q, 10); if !ok { panic(fmt.Errorf(\"invalid integer literal: %%s\", %q)) }; return runtime.IntegerValue{Val: val, TypeSuffix: %s} }()",
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
	info, ok := g.structs[lit.StructType.Name]
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

func (g *generator) compileFunctionCall(ctx *compileContext, call *ast.FunctionCall, expected string) (string, string, bool) {
	if call == nil {
		ctx.setReason("missing function call")
		return "", "", false
	}
	callNode := g.diagNodeName(call, "*ast.FunctionCall", "call")
	if len(call.TypeArguments) > 0 {
		if callee, ok := call.Callee.(*ast.Identifier); ok && callee != nil {
			// Generic calls on local values (e.g. generic lambdas) must call the
			// bound value, not global name lookup.
			if _, ok := ctx.lookup(callee.Name); ok {
				return g.compileDynamicCall(ctx, call, expected, "", callNode)
			}
			if !g.hasDynamicFeature && !g.mayResolveStaticNamedCall(ctx, callee.Name) && !g.mayResolveStaticUFCSCall(ctx, call, callee.Name) {
				ctx.setReason(fmt.Sprintf("unresolved static call (%s)", callee.Name))
				return "", "", false
			}
			return g.compileDynamicCall(ctx, call, expected, callee.Name, callNode)
		}
		return g.compileDynamicCall(ctx, call, expected, "", callNode)
	}
	if callee, ok := call.Callee.(*ast.Identifier); ok && callee != nil {
		if info, ok := ctx.functions[callee.Name]; ok && info != nil && info.Compileable {
			needsRuntimeValue := expected == "runtime.Value" && info.ReturnType != "runtime.Value"
			needsExpect := expected != "" && expected != "runtime.Value" && info.ReturnType == "runtime.Value"
			if !g.typeMatches(expected, info.ReturnType) && !needsRuntimeValue && !needsExpect {
				ctx.setReason("call return type mismatch")
				return "", "", false
			}
			optionalLast := g.hasOptionalLastParam(info)
			if len(call.Arguments) != len(info.Params) {
				if !(optionalLast && len(call.Arguments) == len(info.Params)-1) {
					ctx.setReason("call arity mismatch")
					return "", "", false
				}
			}
			missingOptional := optionalLast && len(call.Arguments) == len(info.Params)-1
			if missingOptional && len(info.Params) > 0 && info.Params[len(info.Params)-1].GoType != "runtime.Value" {
				ctx.setReason("call arity mismatch")
				return "", "", false
			}
			args := make([]string, 0, len(call.Arguments))
			preLines := make([]string, 0, len(call.Arguments))
			postLines := make([]string, 0, len(call.Arguments))
			for idx, arg := range call.Arguments {
				param := info.Params[idx]
				if g.typeCategory(param.GoType) == "struct" {
					if ident, ok := arg.(*ast.Identifier); ok && ident != nil {
						if binding, ok := ctx.lookup(ident.Name); ok && binding.GoType == "runtime.Value" {
							runtimeTemp := ctx.newTemp()
							preLines = append(preLines, fmt.Sprintf("%s := %s", runtimeTemp, binding.GoName))
							structExpr, ok := g.expectRuntimeValueExpr(runtimeTemp, param.GoType)
							if !ok {
								ctx.setReason("call argument unsupported")
								return "", "", false
							}
							structTemp := ctx.newTemp()
							preLines = append(preLines, fmt.Sprintf("%s := %s", structTemp, structExpr))
							args = append(args, structTemp)
							baseName, ok := g.structBaseName(param.GoType)
							if !ok {
								baseName = strings.TrimPrefix(param.GoType, "*")
							}
							postLines = append(postLines, fmt.Sprintf("if err := __able_struct_%s_apply(__able_runtime, %s, %s); err != nil { panic(err) }", baseName, runtimeTemp, structTemp))
							continue
						}
					}
				}
				expr, exprType, ok := g.compileExpr(ctx, arg, param.GoType)
				if !ok {
					return "", "", false
				}
				argExpr := expr
				argType := exprType
				if ifaceType, ok := g.interfaceTypeExpr(param.TypeExpr); ok {
					if argType != "runtime.Value" {
						valueExpr, ok := g.runtimeValueExpr(argExpr, argType)
						if !ok {
							ctx.setReason("interface argument unsupported")
							return "", "", false
						}
						argExpr = valueExpr
						argType = "runtime.Value"
					}
					coerced, ok := g.interfaceArgExpr(argExpr, ifaceType, callee.Name, ctx.genericNames)
					if !ok {
						ctx.setReason("interface argument unsupported")
						return "", "", false
					}
					argExpr = coerced
				}
				args = append(args, argExpr)
			}
			if missingOptional {
				args = append(args, "runtime.NilValue{}")
			}
			callExpr := fmt.Sprintf("__able_compiled_%s(%s)", info.GoName, strings.Join(args, ", "))
			bodyLines := []string{
				fmt.Sprintf("__able_push_call_frame(%s)", callNode),
				"defer __able_pop_call_frame()",
			}
			bodyLines = append(bodyLines, preLines...)
			if len(postLines) == 0 {
				bodyLines = append(bodyLines, fmt.Sprintf("return %s", callExpr))
			} else {
				resultTemp := ctx.newTemp()
				bodyLines = append(bodyLines, fmt.Sprintf("%s := %s", resultTemp, callExpr))
				bodyLines = append(bodyLines, postLines...)
				bodyLines = append(bodyLines, fmt.Sprintf("return %s", resultTemp))
			}
			wrapped := fmt.Sprintf("func() %s { %s }()", info.ReturnType, strings.Join(bodyLines, "; "))
			if needsRuntimeValue {
				converted, ok := g.runtimeValueExpr(wrapped, info.ReturnType)
				if !ok {
					ctx.setReason("call return type mismatch")
					return "", "", false
				}
				return converted, "runtime.Value", true
			}
			if needsExpect {
				converted, ok := g.expectRuntimeValueExpr(wrapped, expected)
				if !ok {
					ctx.setReason("call return type mismatch")
					return "", "", false
				}
				return converted, expected, true
			}
			return wrapped, info.ReturnType, true
		}
		if _, ok := ctx.overloads[callee.Name]; ok {
			return g.compileOverloadCall(ctx, call, expected, callee.Name, callNode)
		}
		if _, ok := ctx.lookup(callee.Name); !ok {
			if !g.hasDynamicFeature && !g.mayResolveStaticNamedCall(ctx, callee.Name) && !g.mayResolveStaticUFCSCall(ctx, call, callee.Name) {
				ctx.setReason(fmt.Sprintf("unresolved static call (%s)", callee.Name))
				return "", "", false
			}
			return g.compileDynamicCall(ctx, call, expected, callee.Name, callNode)
		}
	}
	return g.compileDynamicCall(ctx, call, expected, "", callNode)
}

func (g *generator) compileDynamicCall(ctx *compileContext, call *ast.FunctionCall, expected string, calleeName string, callNode string) (string, string, bool) {
	if call == nil {
		ctx.setReason("missing function call")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" && !g.isVoidType(expected) && g.typeCategory(expected) == "unknown" {
		ctx.setReason("call return type mismatch")
		return "", "", false
	}
	lines := make([]string, 0, len(call.Arguments)+2)
	args := make([]string, 0, len(call.Arguments))
	calleeTemp := ""
	writebackNeeded := false
	writebackObjExpr := ""
	writebackObjType := ""
	writebackObjTemp := ""
	if calleeName == "" {
		switch callee := call.Callee.(type) {
		case *ast.MemberAccessExpression:
			if callee.Member != nil && !callee.Safe {
				if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
					if method, ok := g.resolveStaticMethodCall(ctx, callee.Object, ident.Name); ok {
						return g.compileResolvedMethodCall(ctx, call, expected, method, "", callNode)
					}
				}
			}
			objExpr, objType, ok := g.compileExpr(ctx, callee.Object, "")
			if !ok {
				return "", "", false
			}
			if callee.Member != nil && !callee.Safe {
				if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
					if method := g.methodForReceiver(objType, ident.Name); method != nil {
						return g.compileResolvedMethodCall(ctx, call, expected, method, objExpr, callNode)
					}
				}
			}
			if callee.Safe && g.typeCategory(objType) == "runtime" {
				return g.compileSafeMemberCall(ctx, call, callee, expected, objExpr, objType, callNode)
			}
			// Check for impl sibling methods: default interface methods calling
			// sibling methods on self (e.g., describe() calling self.name())
			siblingHandled := false
			if len(ctx.implSiblings) > 0 && ctx.hasImplicitReceiver && !callee.Safe {
				if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil {
					if sibling, hasSibling := ctx.implSiblings[ident.Name]; hasSibling {
						if objIdent, ok := callee.Object.(*ast.Identifier); ok && objIdent != nil && objIdent.Name == ctx.implicitReceiver.Name {
							objValue, ok := g.runtimeValueExpr(objExpr, objType)
							if ok {
								objTemp := ctx.newTemp()
								calleeTemp = ctx.newTemp()
								lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objValue))
								lines = append(lines, fmt.Sprintf("%s := __able_impl_self_method(%s, %q, %d, __able_wrap_%s)", calleeTemp, objTemp, ident.Name, sibling.Arity, sibling.GoName))
								if g.typeCategory(objType) == "struct" && g.isAddressableMemberObject(callee.Object) {
									writebackNeeded = true
									writebackObjExpr = objExpr
									writebackObjType = objType
									writebackObjTemp = objTemp
								}
								siblingHandled = true
							}
						}
					}
				}
			}
			if !siblingHandled {
				objValue, ok := g.runtimeValueExpr(objExpr, objType)
				if !ok {
					ctx.setReason("method call receiver unsupported")
					return "", "", false
				}
				memberValue, ok := g.memberAssignmentRuntimeValue(ctx, callee.Member)
				if !ok {
					ctx.setReason("method call target unsupported")
					return "", "", false
				}
				objTemp := ctx.newTemp()
				memberTemp := ctx.newTemp()
				calleeTemp = ctx.newTemp()
				lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objValue))
				lines = append(lines, fmt.Sprintf("%s := %s", memberTemp, memberValue))
				lines = append(lines, fmt.Sprintf("%s := __able_member_get_method(%s, %s)", calleeTemp, objTemp, memberTemp))
				if g.typeCategory(objType) == "struct" && g.isAddressableMemberObject(callee.Object) {
					writebackNeeded = true
					writebackObjExpr = objExpr
					writebackObjType = objType
					writebackObjTemp = objTemp
				}
			}
		default:
			calleeExpr, calleeType, ok := g.compileExpr(ctx, call.Callee, "")
			if !ok {
				return "", "", false
			}
			calleeValue, ok := g.runtimeValueExpr(calleeExpr, calleeType)
			if !ok {
				ctx.setReason("call target unsupported")
				return "", "", false
			}
			calleeTemp = ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", calleeTemp, calleeValue))
		}
	}

	for _, arg := range call.Arguments {
		expr, goType, ok := g.compileExpr(ctx, arg, "")
		if !ok {
			return "", "", false
		}
		valueExpr, ok := g.runtimeValueExpr(expr, goType)
		if !ok {
			ctx.setReason("call argument unsupported")
			return "", "", false
		}
		temp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", temp, valueExpr))
		args = append(args, temp)
	}

	argList := strings.Join(args, ", ")
	if argList != "" {
		argList = "[]runtime.Value{" + argList + "}"
	} else {
		argList = "nil"
	}

	callExpr := ""
	if calleeName != "" {
		if wrapper, ok := g.externCallWrapper(calleeName); ok {
			callExpr = fmt.Sprintf("%s(%s, %s)", wrapper, argList, callNode)
		} else {
			callExpr = fmt.Sprintf("__able_call_named(%q, %s, %s)", calleeName, argList, callNode)
		}
	} else {
		callExpr = fmt.Sprintf("__able_call_value(%s, %s, %s)", calleeTemp, argList, callNode)
	}

	if writebackNeeded {
		baseName, ok := g.structBaseName(writebackObjType)
		if !ok {
			baseName = strings.TrimPrefix(writebackObjType, "*")
		}
		convertedTemp := ctx.newTemp()
		writebackLines := []string{
			fmt.Sprintf("%s, err := __able_struct_%s_from(%s)", convertedTemp, baseName, writebackObjTemp),
			"if err != nil { panic(err) }",
		}
		if strings.HasPrefix(writebackObjType, "*") {
			writebackLines = append(writebackLines, fmt.Sprintf("*%s = *%s", writebackObjExpr, convertedTemp))
		} else {
			writebackLines = append(writebackLines, fmt.Sprintf("%s = *%s", writebackObjExpr, convertedTemp))
		}
		if g.isVoidType(expected) {
			lines = append(lines, fmt.Sprintf("_ = %s", callExpr))
			lines = append(lines, writebackLines...)
			return fmt.Sprintf("func() struct{} { %s; return struct{}{} }()", strings.Join(lines, "; ")), "struct{}", true
		}
		resultTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", resultTemp, callExpr))
		lines = append(lines, writebackLines...)
		resultExpr := resultTemp
		resultType := "runtime.Value"
		if expected != "" && expected != "runtime.Value" {
			converted, ok := g.expectRuntimeValueExpr(resultTemp, expected)
			if !ok {
				ctx.setReason("call return type mismatch")
				return "", "", false
			}
			resultExpr = converted
			resultType = expected
		}
		return fmt.Sprintf("func() %s { %s; return %s }()", resultType, strings.Join(lines, "; "), resultExpr), resultType, true
	}

	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", callExpr))
		return fmt.Sprintf("func() struct{} { %s; return struct{}{} }()", strings.Join(lines, "; ")), "struct{}", true
	}

	resultExpr := callExpr
	resultType := "runtime.Value"
	if expected != "" && expected != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(callExpr, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return "", "", false
		}
		resultExpr = converted
		resultType = expected
	}
	if len(lines) == 0 {
		return resultExpr, resultType, true
	}
	return fmt.Sprintf("func() %s { %s; return %s }()", resultType, strings.Join(lines, "; "), resultExpr), resultType, true
}

func (g *generator) externCallWrapper(name string) (string, bool) {
	switch name {
	case "__able_array_new":
		return "__able_extern_array_new", true
	case "__able_array_with_capacity":
		return "__able_extern_array_with_capacity", true
	case "__able_array_size":
		return "__able_extern_array_size", true
	case "__able_array_capacity":
		return "__able_extern_array_capacity", true
	case "__able_array_set_len":
		return "__able_extern_array_set_len", true
	case "__able_array_read":
		return "__able_extern_array_read", true
	case "__able_array_write":
		return "__able_extern_array_write", true
	case "__able_array_reserve":
		return "__able_extern_array_reserve", true
	case "__able_array_clone":
		return "__able_extern_array_clone", true
	case "__able_hash_map_new":
		return "__able_extern_hash_map_new", true
	case "__able_hash_map_with_capacity":
		return "__able_extern_hash_map_with_capacity", true
	case "__able_hash_map_get":
		return "__able_extern_hash_map_get", true
	case "__able_hash_map_set":
		return "__able_extern_hash_map_set", true
	case "__able_hash_map_remove":
		return "__able_extern_hash_map_remove", true
	case "__able_hash_map_contains":
		return "__able_extern_hash_map_contains", true
	case "__able_hash_map_size":
		return "__able_extern_hash_map_size", true
	case "__able_hash_map_clear":
		return "__able_extern_hash_map_clear", true
	case "__able_hash_map_for_each":
		return "__able_extern_hash_map_for_each", true
	case "__able_hash_map_clone":
		return "__able_extern_hash_map_clone", true
	case "__able_String_from_builtin":
		return "__able_extern_string_from_builtin", true
	case "__able_String_to_builtin":
		return "__able_extern_string_to_builtin", true
	case "__able_char_from_codepoint":
		return "__able_extern_char_from_codepoint", true
	case "__able_char_to_codepoint":
		return "__able_extern_char_to_codepoint", true
	case "__able_ratio_from_float":
		return "__able_extern_ratio_from_float", true
	case "__able_f32_bits":
		return "__able_extern_f32_bits", true
	case "__able_f64_bits":
		return "__able_extern_f64_bits", true
	case "__able_u64_mul":
		return "__able_extern_u64_mul", true
	case "__able_channel_new":
		return "__able_extern_channel_new", true
	case "__able_channel_send":
		return "__able_extern_channel_send", true
	case "__able_channel_receive":
		return "__able_extern_channel_receive", true
	case "__able_channel_try_send":
		return "__able_extern_channel_try_send", true
	case "__able_channel_try_receive":
		return "__able_extern_channel_try_receive", true
	case "__able_channel_await_try_recv":
		return "__able_extern_channel_await_try_recv", true
	case "__able_channel_await_try_send":
		return "__able_extern_channel_await_try_send", true
	case "__able_channel_close":
		return "__able_extern_channel_close", true
	case "__able_channel_is_closed":
		return "__able_extern_channel_is_closed", true
	case "__able_mutex_new":
		return "__able_extern_mutex_new", true
	case "__able_mutex_lock":
		return "__able_extern_mutex_lock", true
	case "__able_mutex_unlock":
		return "__able_extern_mutex_unlock", true
	case "__able_mutex_await_lock":
		return "__able_extern_mutex_await_lock", true
	default:
		return "", false
	}
}

func (g *generator) resolveStaticMethodCall(ctx *compileContext, object ast.Expression, memberName string) (*methodInfo, bool) {
	if g == nil || object == nil || memberName == "" {
		return nil, false
	}
	ident, ok := object.(*ast.Identifier)
	if !ok || ident == nil || ident.Name == "" {
		return nil, false
	}
	if ctx != nil {
		if _, ok := ctx.lookup(ident.Name); ok {
			return nil, false
		}
	}
	if _, ok := g.structs[ident.Name]; !ok {
		return nil, false
	}
	method := g.methodForTypeName(ident.Name, memberName, false)
	if method == nil {
		return nil, false
	}
	return method, true
}

func (g *generator) compileResolvedMethodCall(ctx *compileContext, call *ast.FunctionCall, expected string, method *methodInfo, receiverExpr string, callNode string) (string, string, bool) {
	if call == nil || method == nil || method.Info == nil {
		ctx.setReason("missing method call")
		return "", "", false
	}
	info := method.Info
	if !info.Compileable {
		ctx.setReason("unsupported method call")
		return "", "", false
	}
	if !g.typeMatches(expected, info.ReturnType) {
		ctx.setReason("call return type mismatch")
		return "", "", false
	}
	paramOffset := 0
	args := make([]string, 0, len(call.Arguments)+1)
	if method.ExpectsSelf {
		if receiverExpr == "" {
			ctx.setReason("method receiver missing")
			return "", "", false
		}
		args = append(args, receiverExpr)
		paramOffset = 1
	}
	params := info.Params
	if paramOffset > len(params) {
		ctx.setReason("method params missing")
		return "", "", false
	}
	callArgCount := len(call.Arguments)
	paramCount := len(params) - paramOffset
	optionalLast := g.hasOptionalLastParam(info)
	if callArgCount != paramCount {
		if !(optionalLast && callArgCount == paramCount-1) {
			ctx.setReason("call arity mismatch")
			return "", "", false
		}
	}
	missingOptional := optionalLast && callArgCount == paramCount-1
	if missingOptional && paramCount > 0 && params[len(params)-1].GoType != "runtime.Value" {
		ctx.setReason("call arity mismatch")
		return "", "", false
	}
	for idx, arg := range call.Arguments {
		param := params[paramOffset+idx]
		expr, exprType, ok := g.compileExpr(ctx, arg, param.GoType)
		if !ok {
			return "", "", false
		}
		argExpr := expr
		argType := exprType
		if ifaceType, ok := g.interfaceTypeExpr(param.TypeExpr); ok {
			if argType != "runtime.Value" {
				valueExpr, ok := g.runtimeValueExpr(argExpr, argType)
				if !ok {
					ctx.setReason("interface argument unsupported")
					return "", "", false
				}
				argExpr = valueExpr
				argType = "runtime.Value"
			}
			coerced, ok := g.interfaceArgExpr(argExpr, ifaceType, info.Name, ctx.genericNames)
			if !ok {
				ctx.setReason("interface argument unsupported")
				return "", "", false
			}
			argExpr = coerced
		}
		args = append(args, argExpr)
	}
	if missingOptional {
		args = append(args, "runtime.NilValue{}")
	}
	callExpr := fmt.Sprintf("__able_compiled_%s(%s)", info.GoName, strings.Join(args, ", "))
	return fmt.Sprintf("func() %s { __able_push_call_frame(%s); defer __able_pop_call_frame(); return %s }()", info.ReturnType, callNode, callExpr), info.ReturnType, true
}

func (g *generator) compileSafeMemberCall(ctx *compileContext, call *ast.FunctionCall, callee *ast.MemberAccessExpression, expected string, objExpr string, objType string, callNode string) (string, string, bool) {
	if call == nil || callee == nil {
		ctx.setReason("missing safe member call")
		return "", "", false
	}
	objValue, ok := g.runtimeValueExpr(objExpr, objType)
	if !ok {
		ctx.setReason("method call receiver unsupported")
		return "", "", false
	}
	memberValue, ok := g.memberAssignmentRuntimeValue(ctx, callee.Member)
	if !ok {
		ctx.setReason("method call target unsupported")
		return "", "", false
	}
	lines := make([]string, 0, len(call.Arguments)+4)
	args := make([]string, 0, len(call.Arguments))
	objTemp := ctx.newTemp()
	memberTemp := ctx.newTemp()
	calleeTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objValue))
	lines = append(lines, fmt.Sprintf("if __able_is_nil(%s) { return %s }", objTemp, safeNilReturnExpr(expected)))
	lines = append(lines, fmt.Sprintf("%s := %s", memberTemp, memberValue))
	lines = append(lines, fmt.Sprintf("%s := __able_member_get_method(%s, %s)", calleeTemp, objTemp, memberTemp))
	for _, arg := range call.Arguments {
		expr, goType, ok := g.compileExpr(ctx, arg, "")
		if !ok {
			return "", "", false
		}
		valueExpr, ok := g.runtimeValueExpr(expr, goType)
		if !ok {
			ctx.setReason("call argument unsupported")
			return "", "", false
		}
		temp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", temp, valueExpr))
		args = append(args, temp)
	}
	argList := strings.Join(args, ", ")
	if argList != "" {
		argList = "[]runtime.Value{" + argList + "}"
	} else {
		argList = "nil"
	}
	callExpr := fmt.Sprintf("__able_call_value(%s, %s, %s)", calleeTemp, argList, callNode)
	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", callExpr))
		lines = append(lines, "return struct{}{}")
		return fmt.Sprintf("func() struct{} { %s }()", strings.Join(lines, "; ")), "struct{}", true
	}
	lines = append(lines, fmt.Sprintf("return %s", callExpr))
	baseExpr := fmt.Sprintf("func() runtime.Value { %s }()", strings.Join(lines, "; "))
	if expected == "" || expected == "runtime.Value" {
		return baseExpr, "runtime.Value", true
	}
	converted, ok := g.expectRuntimeValueExpr(baseExpr, expected)
	if !ok {
		ctx.setReason("call return type mismatch")
		return "", "", false
	}
	return converted, expected, true
}

func safeNilReturnExpr(expected string) string {
	if expected == "struct{}" {
		return "struct{}{}"
	}
	return "runtime.NilValue{}"
}

func (g *generator) compileTypeCast(ctx *compileContext, expr *ast.TypeCastExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Expression == nil || expr.TargetType == nil {
		ctx.setReason("missing type cast")
		return "", "", false
	}
	valueExpr, valueType, ok := g.compileExpr(ctx, expr.Expression, "")
	if !ok {
		return "", "", false
	}
	valueRuntime, ok := g.runtimeValueExpr(valueExpr, valueType)
	if !ok {
		ctx.setReason("cast operand unsupported")
		return "", "", false
	}
	targetExpr, ok := g.renderTypeExpression(expr.TargetType)
	if !ok {
		ctx.setReason("unsupported cast type")
		return "", "", false
	}
	castExpr := fmt.Sprintf("__able_cast(%s, %s)", valueRuntime, targetExpr)
	if expected == "runtime.Value" {
		return castExpr, "runtime.Value", true
	}
	desiredType := "runtime.Value"
	if expected != "" && expected != "runtime.Value" {
		desiredType = expected
	} else if mapped, ok := g.mapTypeExpressionInPackage(ctx.packageName, expr.TargetType); ok && mapped != "" {
		desiredType = mapped
	}
	if desiredType == "struct{}" {
		ctx.setReason("cast to void unsupported")
		return "", "", false
	}
	if desiredType == "runtime.Value" {
		return castExpr, "runtime.Value", true
	}
	converted, ok := g.expectRuntimeValueExpr(castExpr, desiredType)
	if !ok {
		ctx.setReason("cast type mismatch")
		return "", "", false
	}
	return converted, desiredType, true
}

func (g *generator) compileRangeExpression(ctx *compileContext, expr *ast.RangeExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Start == nil || expr.End == nil {
		ctx.setReason("missing range expression")
		return "", "", false
	}
	startExpr, startType, ok := g.compileExpr(ctx, expr.Start, "")
	if !ok {
		return "", "", false
	}
	endExpr, endType, ok := g.compileExpr(ctx, expr.End, "")
	if !ok {
		return "", "", false
	}
	startRuntime, ok := g.runtimeValueExpr(startExpr, startType)
	if !ok {
		ctx.setReason("range start unsupported")
		return "", "", false
	}
	endRuntime, ok := g.runtimeValueExpr(endExpr, endType)
	if !ok {
		ctx.setReason("range end unsupported")
		return "", "", false
	}
	startTemp := ctx.newTemp()
	endTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", startTemp, startRuntime),
		fmt.Sprintf("%s := %s", endTemp, endRuntime),
	}
	rangeExpr := fmt.Sprintf("__able_range(%s, %s, %t)", startTemp, endTemp, expr.Inclusive)
	resultType := "runtime.Value"
	resultExpr := rangeExpr
	if expected != "" && expected != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(rangeExpr, expected)
		if !ok {
			ctx.setReason("range expression type mismatch")
			return "", "", false
		}
		resultExpr = converted
		resultType = expected
	}
	return fmt.Sprintf("func() %s { %s; return %s }()", resultType, strings.Join(lines, "; "), resultExpr), resultType, true
}

func (g *generator) compileLambdaExpression(ctx *compileContext, expr *ast.LambdaExpression, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing lambda expression")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" {
		ctx.setReason("lambda expression type mismatch")
		return "", "", false
	}
	if expr.Body == nil {
		ctx.setReason("missing lambda body")
		return "", "", false
	}

	lambdaCtx := ctx.child()
	if lambdaCtx == nil {
		ctx.setReason("missing lambda context")
		return "", "", false
	}
	lambdaCtx.loopDepth = 0

	params := make([]paramInfo, 0, len(expr.Params))
	for idx, param := range expr.Params {
		if param == nil {
			ctx.setReason("missing lambda parameter")
			return "", "", false
		}
		ident, ok := param.Name.(*ast.Identifier)
		if !ok || ident == nil || ident.Name == "" {
			ctx.setReason("unsupported lambda parameter")
			return "", "", false
		}
		goName := safeParamName(ident.Name, idx)
		goType := "runtime.Value"
		if param.ParamType != nil {
			mapped, ok := g.mapTypeExpressionInPackage(ctx.packageName, param.ParamType)
			if !ok {
				ctx.setReason("unsupported lambda parameter type")
				return "", "", false
			}
			goType = mapped
		}
		info := paramInfo{Name: ident.Name, GoName: goName, GoType: goType}
		lambdaCtx.locals[ident.Name] = info
		params = append(params, info)
	}
	if len(params) > 0 {
		lambdaCtx.implicitReceiver = params[0]
		lambdaCtx.hasImplicitReceiver = true
	}
	genericValueVars := make(map[string]string)
	if len(expr.GenericParams) > 0 || len(expr.WhereClause) > 0 {
		generics := genericNameSet(expr.GenericParams)
		for idx, param := range expr.Params {
			if param == nil || param.ParamType == nil {
				continue
			}
			if simple, ok := param.ParamType.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil {
				if _, ok := generics[simple.Name.Name]; ok {
					genericValueVars[simple.Name.Name] = fmt.Sprintf("__able_lambda_arg_%d_value", idx)
				}
			}
		}
	}

	desiredReturn := ""
	if expr.ReturnType != nil {
		mapped, ok := g.mapTypeExpressionInPackage(ctx.packageName, expr.ReturnType)
		if !ok {
			ctx.setReason("unsupported lambda return type")
			return "", "", false
		}
		desiredReturn = mapped
	}

	var bodyLines []string
	var bodyExpr string
	var bodyType string
	var ok bool
	if expr.IsVerboseSyntax {
		block, isBlock := expr.Body.(*ast.BlockExpression)
		if !isBlock || block == nil {
			ctx.setReason("verbose lambda requires block body")
			return "", "", false
		}
		bodyLines, bodyExpr, bodyType, ok = g.compileLambdaBlockBody(lambdaCtx, desiredReturn, block)
	} else {
		bodyLines, bodyExpr, bodyType, ok = g.compileTailExpression(lambdaCtx, desiredReturn, expr.Body)
	}
	if !ok {
		if ctx.reason == "" && lambdaCtx.reason != "" {
			ctx.setReason(lambdaCtx.reason)
		}
		return "", "", false
	}
	if desiredReturn != "" && !g.typeMatches(desiredReturn, bodyType) {
		ctx.setReason("lambda return type mismatch")
		return "", "", false
	}

	lambdaResultName := lambdaCtx.newTemp()
	lambdaErrName := lambdaCtx.newTemp()
	implLines := make([]string, 0, len(bodyLines)+len(params)*4+7)
	implLines = append(implLines, fmt.Sprintf("defer func() { if recovered := recover(); recovered != nil { switch recovered.(type) { case __able_break, __able_break_label_signal, __able_continue_signal, __able_continue_label_signal: panic(recovered) }; %s = nil; %s = bridge.Recover(__able_runtime, callCtx, recovered) } }()", lambdaResultName, lambdaErrName))
	implLines = append(implLines, "if __able_runtime != nil && callCtx != nil && callCtx.Env != nil { prevEnv := __able_runtime.SwapEnv(callCtx.Env); defer __able_runtime.SwapEnv(prevEnv) }")
	implLines = append(implLines, fmt.Sprintf("if len(args) != %d { return nil, fmt.Errorf(\"lambda expects %d arguments, got %%d\", len(args)) }", len(params), len(params)))
	for idx, param := range params {
		argVar := fmt.Sprintf("__able_lambda_arg_%d", idx)
		valueVar := argVar + "_value"
		implLines = append(implLines, fmt.Sprintf("%s := args[%d]", valueVar, idx))
		convLines, ok := g.lambdaArgConversionLines(argVar, valueVar, param.GoType, param.GoName)
		if !ok {
			ctx.setReason("unsupported lambda parameter type")
			return "", "", false
		}
		implLines = append(implLines, convLines...)
		if param.GoName != "_" {
			implLines = append(implLines, fmt.Sprintf("_ = %s", param.GoName))
		}
	}
	if len(genericValueVars) > 0 {
		constraintLines, ok := g.lambdaConstraintLines(expr, genericValueVars)
		if !ok {
			ctx.setReason("unsupported lambda constraints")
			return "", "", false
		}
		implLines = append(implLines, constraintLines...)
	}

	implLines = append(implLines, bodyLines...)
	if g.isVoidType(bodyType) {
		if bodyExpr != "" {
			implLines = append(implLines, fmt.Sprintf("_ = %s", bodyExpr))
		}
		implLines = append(implLines, "return runtime.VoidValue{}, nil")
	} else {
		resultTemp := lambdaCtx.newTemp()
		implLines = append(implLines, fmt.Sprintf("%s := %s", resultTemp, bodyExpr))
		returnLines, ok := g.lambdaReturnLines(resultTemp, bodyType)
		if !ok {
			ctx.setReason("unsupported lambda return type")
			return "", "", false
		}
		implLines = append(implLines, returnLines...)
	}

	implBody := strings.Join(implLines, "; ")
	lambdaExpr := fmt.Sprintf("runtime.NativeFunctionValue{Name: %q, Arity: %d, Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (%s runtime.Value, %s error) { %s }}", "<lambda>", len(params), lambdaResultName, lambdaErrName, implBody)
	return lambdaExpr, "runtime.Value", true
}

func (g *generator) compileLambdaBlockBody(ctx *compileContext, returnType string, block *ast.BlockExpression) ([]string, string, string, bool) {
	if block == nil {
		ctx.setReason("missing lambda body")
		return nil, "", "", false
	}
	statements := block.Body
	if len(statements) == 0 {
		if returnType == "" || g.isVoidType(returnType) {
			return nil, "struct{}{}", "struct{}", true
		}
		ctx.setReason("empty lambda body requires void return")
		return nil, "", "", false
	}
	lines := make([]string, 0, len(statements))
	for idx, stmt := range statements {
		isLast := idx == len(statements)-1
		if ret, ok := stmt.(*ast.ReturnStatement); ok {
			if ret.Argument == nil {
				if returnType != "" && !g.isVoidType(returnType) {
					ctx.setReason("missing return value")
					return nil, "", "", false
				}
				return lines, "struct{}{}", "struct{}", true
			}
			compileExpected := returnType
			if g.isVoidType(returnType) {
				compileExpected = ""
			}
			stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, compileExpected, ret.Argument)
			if !ok {
				return nil, "", "", false
			}
			if returnType != "" && !g.isVoidType(returnType) && !g.typeMatches(returnType, valueType) {
				ctx.setReason("lambda return type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, stmtLines...)
			if g.isVoidType(returnType) {
				if valueExpr != "" {
					lines = append(lines, fmt.Sprintf("_ = %s", valueExpr))
				}
				return lines, "struct{}{}", "struct{}", true
			}
			finalType := valueType
			if returnType != "" {
				finalType = returnType
			}
			return lines, valueExpr, finalType, true
		}
		if isLast {
			if expr, ok := stmt.(ast.Expression); ok && expr != nil {
				compileExpected := returnType
				if g.isVoidType(returnType) {
					compileExpected = ""
				}
				stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, compileExpected, expr)
				if !ok {
					return nil, "", "", false
				}
				if returnType != "" && !g.isVoidType(returnType) && !g.typeMatches(returnType, valueType) {
					ctx.setReason("lambda return type mismatch")
					return nil, "", "", false
				}
				lines = append(lines, stmtLines...)
				if g.isVoidType(returnType) {
					if valueExpr != "" {
						lines = append(lines, fmt.Sprintf("_ = %s", valueExpr))
					}
					return lines, "struct{}{}", "struct{}", true
				}
				finalType := valueType
				if returnType != "" {
					finalType = returnType
				}
				return lines, valueExpr, finalType, true
			}
			if returnType == "" || g.isVoidType(returnType) {
				stmtLines, ok := g.compileStatement(ctx, stmt)
				if !ok {
					return nil, "", "", false
				}
				lines = append(lines, stmtLines...)
				return lines, "struct{}{}", "struct{}", true
			}
			ctx.setReason("missing return statement")
			return nil, "", "", false
		}
		stmtLines, ok := g.compileStatement(ctx, stmt)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, stmtLines...)
	}
	ctx.setReason("missing return statement")
	return nil, "", "", false
}

func (g *generator) lambdaArgConversionLines(argVar string, valueVar string, goType string, target string) ([]string, bool) {
	switch g.typeCategory(goType) {
	case "runtime":
		return []string{fmt.Sprintf("%s := %s", target, valueVar)}, true
	case "bool":
		return []string{
			fmt.Sprintf("%s, err := bridge.AsBool(%s)", target, valueVar),
			"if err != nil { return nil, err }",
		}, true
	case "string":
		return []string{
			fmt.Sprintf("%s, err := bridge.AsString(%s)", target, valueVar),
			"if err != nil { return nil, err }",
		}, true
	case "rune":
		return []string{
			fmt.Sprintf("%s, err := bridge.AsRune(%s)", target, valueVar),
			"if err != nil { return nil, err }",
		}, true
	case "float32":
		raw := argVar + "_raw"
		return []string{
			fmt.Sprintf("%s, err := bridge.AsFloat(%s)", raw, valueVar),
			"if err != nil { return nil, err }",
			fmt.Sprintf("%s := float32(%s)", target, raw),
		}, true
	case "float64":
		return []string{
			fmt.Sprintf("%s, err := bridge.AsFloat(%s)", target, valueVar),
			"if err != nil { return nil, err }",
		}, true
	case "int":
		raw := argVar + "_raw"
		return []string{
			fmt.Sprintf("%s, err := bridge.AsInt(%s, bridge.NativeIntBits)", raw, valueVar),
			"if err != nil { return nil, err }",
			fmt.Sprintf("%s := int(%s)", target, raw),
		}, true
	case "uint":
		raw := argVar + "_raw"
		return []string{
			fmt.Sprintf("%s, err := bridge.AsUint(%s, bridge.NativeIntBits)", raw, valueVar),
			"if err != nil { return nil, err }",
			fmt.Sprintf("%s := uint(%s)", target, raw),
		}, true
	case "int8", "int16", "int32", "int64":
		bits := g.intBits(goType)
		raw := argVar + "_raw"
		return []string{
			fmt.Sprintf("%s, err := bridge.AsInt(%s, %d)", raw, valueVar, bits),
			"if err != nil { return nil, err }",
			fmt.Sprintf("%s := %s(%s)", target, goType, raw),
		}, true
	case "uint8", "uint16", "uint32", "uint64":
		bits := g.intBits(goType)
		raw := argVar + "_raw"
		return []string{
			fmt.Sprintf("%s, err := bridge.AsUint(%s, %d)", raw, valueVar, bits),
			"if err != nil { return nil, err }",
			fmt.Sprintf("%s := %s(%s)", target, goType, raw),
		}, true
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		return []string{
			fmt.Sprintf("%s, err := __able_struct_%s_from(%s)", target, baseName, valueVar),
			"if err != nil { return nil, err }",
		}, true
	default:
		return []string{fmt.Sprintf("%s := %s", target, valueVar)}, true
	}
}

func (g *generator) lambdaReturnLines(resultName string, goType string) ([]string, bool) {
	switch g.typeCategory(goType) {
	case "runtime":
		return []string{fmt.Sprintf("return %s, nil", resultName)}, true
	case "void":
		return []string{
			fmt.Sprintf("_ = %s", resultName),
			"return runtime.VoidValue{}, nil",
		}, true
	case "bool":
		return []string{fmt.Sprintf("return bridge.ToBool(%s), nil", resultName)}, true
	case "string":
		return []string{fmt.Sprintf("return bridge.ToString(%s), nil", resultName)}, true
	case "rune":
		return []string{fmt.Sprintf("return bridge.ToRune(%s), nil", resultName)}, true
	case "float32":
		return []string{fmt.Sprintf("return bridge.ToFloat32(%s), nil", resultName)}, true
	case "float64":
		return []string{fmt.Sprintf("return bridge.ToFloat64(%s), nil", resultName)}, true
	case "int":
		return []string{fmt.Sprintf("return bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\")), nil", resultName)}, true
	case "uint":
		return []string{fmt.Sprintf("return bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\")), nil", resultName)}, true
	case "int8":
		return []string{fmt.Sprintf("return bridge.ToInt(int64(%s), runtime.IntegerType(\"i8\")), nil", resultName)}, true
	case "int16":
		return []string{fmt.Sprintf("return bridge.ToInt(int64(%s), runtime.IntegerType(\"i16\")), nil", resultName)}, true
	case "int32":
		return []string{fmt.Sprintf("return bridge.ToInt(int64(%s), runtime.IntegerType(\"i32\")), nil", resultName)}, true
	case "int64":
		return []string{fmt.Sprintf("return bridge.ToInt(int64(%s), runtime.IntegerType(\"i64\")), nil", resultName)}, true
	case "uint8":
		return []string{fmt.Sprintf("return bridge.ToUint(uint64(%s), runtime.IntegerType(\"u8\")), nil", resultName)}, true
	case "uint16":
		return []string{fmt.Sprintf("return bridge.ToUint(uint64(%s), runtime.IntegerType(\"u16\")), nil", resultName)}, true
	case "uint32":
		return []string{fmt.Sprintf("return bridge.ToUint(uint64(%s), runtime.IntegerType(\"u32\")), nil", resultName)}, true
	case "uint64":
		return []string{fmt.Sprintf("return bridge.ToUint(uint64(%s), runtime.IntegerType(\"u64\")), nil", resultName)}, true
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		return []string{fmt.Sprintf("return __able_struct_%s_to(__able_runtime, %s)", baseName, resultName)}, true
	default:
		return []string{fmt.Sprintf("return %s, nil", resultName)}, true
	}
}
