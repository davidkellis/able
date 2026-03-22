package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compilePropagationExpression(ctx *compileContext, expr *ast.PropagationExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.Expression == nil {
		ctx.setReason("missing propagation expression")
		return nil, "", "", false
	}
	if indexExpr, ok := expr.Expression.(*ast.IndexExpression); ok {
		if lines, fastExpr, fastType, fastOK := g.compilePropagationMonoArrayIndex(ctx, indexExpr, expected); fastOK {
			return lines, fastExpr, fastType, true
		}
	}
	if callExpr, ok := expr.Expression.(*ast.FunctionCall); ok {
		if lines, fastExpr, fastType, fastOK := g.compilePropagationStaticArrayAccessCall(ctx, callExpr, expected); fastOK {
			return lines, fastExpr, fastType, true
		}
	}
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", expr.Expression)
	if !ok {
		return nil, "", "", false
	}
	resultType := expected
	if resultType == "" {
		resultType = valueType
	}
	if resultType == "" {
		resultType = "runtime.Value"
	}
	unionSuccessMember, unionFailureMember, valueErrorUnion := g.nativeUnionOrElseMembers(valueType)
	if valueErrorUnion {
		lines := append([]string{}, valueLines...)
		valueTemp := ctx.newTemp()
		failureTemp := ctx.newTemp()
		failureOkTemp := ctx.newTemp()
		successTemp := ctx.newTemp()
		successOkTemp := ctx.newTemp()
		failureRuntimeTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s := %s", valueTemp, valueExpr),
			fmt.Sprintf("%s, %s := %s(%s)", failureTemp, failureOkTemp, unionFailureMember.UnwrapHelper, valueTemp),
		)
		failureRuntimeLines, failureRuntimeExpr, ok := g.runtimeValueLines(ctx, failureTemp, unionFailureMember.GoType)
		if !ok {
			ctx.setReason("propagation type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, indentLines(failureRuntimeLines, 0)...)
		lines = append(lines, fmt.Sprintf("%s := %s", failureRuntimeTemp, failureRuntimeExpr))
		transferLines, ok := g.controlTransferLines(ctx, g.raiseControlExpr("nil", failureRuntimeTemp))
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, fmt.Sprintf("if %s {", failureOkTemp))
		lines = append(lines, indentLines(transferLines, 1)...)
		lines = append(lines, "}")
		lines = append(lines,
			fmt.Sprintf("%s, %s := %s(%s)", successTemp, successOkTemp, unionSuccessMember.UnwrapHelper, valueTemp),
		)
		invariantLines, ok := g.controlTransferLines(ctx, g.runtimeErrorControlExpr("nil", `fmt.Errorf("compiler: native union propagation success branch missing")`))
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, fmt.Sprintf("if !%s {", successOkTemp))
		lines = append(lines, indentLines(invariantLines, 1)...)
		lines = append(lines, "}")
		resultExpr := successTemp
		resultType = unionSuccessMember.GoType
		if expected != "" && expected != resultType {
			convLines, converted, _, ok := g.coerceExpectedStaticExpr(ctx, nil, successTemp, resultType, expected)
			if !ok {
				ctx.setReason("propagation type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			resultExpr = converted
			resultType = expected
		}
		return lines, resultExpr, resultType, true
	}
	if _, nullable := g.nativeNullableValueInnerType(valueType); nullable {
		innerType, _ := g.nativeNullableValueInnerType(valueType)
		lines := append([]string{}, valueLines...)
		valueTemp := ctx.newTemp()
		transferLines, ok := g.controlTransferLines(ctx, g.raiseControlExpr("nil", "runtime.NilValue{}"))
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines,
			fmt.Sprintf("%s := %s", valueTemp, valueExpr),
			fmt.Sprintf("if %s == nil {", valueTemp),
		)
		lines = append(lines, indentLines(transferLines, 1)...)
		lines = append(lines, "}")
		return lines, fmt.Sprintf("(*%s)", valueTemp), innerType, true
	}
	if valueType != "runtime.Value" {
		if !g.typeMatches(resultType, valueType) {
			ctx.setReason("propagation type mismatch")
			return nil, "", "", false
		}
		return valueLines, valueExpr, valueType, true
	}
	valueTemp := ctx.newTemp()
	lines := append([]string{}, valueLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
	transferLines, ok := g.controlTransferLines(ctx, g.raiseControlExpr("nil", fmt.Sprintf("__able_error_value(%s)", valueTemp)))
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, fmt.Sprintf("if __able_is_error(%s) {", valueTemp))
	lines = append(lines, indentLines(transferLines, 1)...)
	lines = append(lines, "}")
	resultExpr := valueTemp
	if resultType != "runtime.Value" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, valueTemp, resultType)
		if !ok {
			ctx.setReason("propagation type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		resultExpr = converted
	}
	return lines, resultExpr, resultType, true
}

func (g *generator) compilePropagationMonoArrayIndex(ctx *compileContext, expr *ast.IndexExpression, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || expr == nil {
		return nil, "", "", false
	}
	objLines, objExpr, objType, ok := g.compileExprLines(ctx, expr.Object, "")
	if !ok || !g.isStaticArrayType(objType) {
		return nil, "", "", false
	}
	idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, expr.Index, "")
	if !ok {
		return nil, "", "", false
	}
	objTemp := ctx.newTemp()
	idxTemp := ctx.newTemp()
	indexTemp := ctx.newTemp()
	lengthTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	resultType, ok := g.staticArrayPropagationResultType(ctx, expr.Object, objType)
	if !ok {
		resultType = expected
		if resultType == "" {
			resultType = "runtime.Value"
		}
	}
	lines := append([]string{}, objLines...)
	lines = append(lines, idxLines...)
	lines = append(lines,
		fmt.Sprintf("%s := %s", objTemp, objExpr),
	)
	lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
	if !ok {
		return nil, "", "", false
	}
	elemLines, elemExpr, _, ok := g.staticArrayResultExprLines(ctx, objType, fmt.Sprintf("%s.Elements[%s]", objTemp, indexTemp), resultType)
	if !ok {
		return nil, "", "", false
	}
	transferLines, ok := g.controlTransferLines(ctx, g.raiseControlExpr("nil", fmt.Sprintf("__able_error_value(__able_index_error(%s, %s))", indexTemp, lengthTemp)))
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines,
		fmt.Sprintf("%s := %s", lengthTemp, g.staticArrayLengthExpr(objTemp)),
		fmt.Sprintf("if %s < 0 || %s >= %s {", indexTemp, indexTemp, lengthTemp),
	)
	lines = append(lines, indentLines(transferLines, 1)...)
	lines = append(lines,
		"}",
	)
	lines = append(lines, elemLines...)
	if resultType == "runtime.Value" {
		lines = append(lines, fmt.Sprintf("var %s runtime.Value = %s", resultTemp, elemExpr))
	} else {
		lines = append(lines, fmt.Sprintf("var %s %s = %s", resultTemp, resultType, elemExpr))
	}
	return lines, resultTemp, resultType, true
}

func (g *generator) compilePropagationStaticArrayAccessCall(ctx *compileContext, call *ast.FunctionCall, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil {
		return nil, "", "", false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil {
		return nil, "", "", false
	}
	memberIdent, ok := member.Member.(*ast.Identifier)
	if !ok || memberIdent == nil {
		return nil, "", "", false
	}
	methodName := memberIdent.Name
	switch methodName {
	case "get", "read_slot", "first", "last", "pop":
	default:
		return nil, "", "", false
	}
	objLines, objExpr, objType, ok := g.compileExprLines(ctx, member.Object, "")
	if !ok || !g.isStaticArrayType(objType) {
		return nil, "", "", false
	}
	resultType, ok := g.staticArrayPropagationResultType(ctx, member.Object, objType)
	if !ok || resultType == "" || resultType == "runtime.Value" {
		return nil, "", "", false
	}
	objTemp := ctx.newTemp()
	lengthTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	lines := append([]string{}, objLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objExpr))
	nilTransferLines, ok := g.controlTransferLines(ctx, g.raiseControlExpr("nil", "runtime.NilValue{}"))
	if !ok {
		return nil, "", "", false
	}
	switch methodName {
	case "get", "read_slot":
		if len(call.Arguments) != 1 {
			return nil, "", "", false
		}
		idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, call.Arguments[0], "")
		if !ok {
			return nil, "", "", false
		}
		idxTemp := ctx.newTemp()
		indexTemp := ctx.newTemp()
		lines = append(lines, idxLines...)
		lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
		if !ok {
			return nil, "", "", false
		}
		elemLines, elemExpr, _, ok := g.staticArrayResultExprLines(ctx, objType, fmt.Sprintf("%s.Elements[%s]", objTemp, indexTemp), resultType)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines,
			fmt.Sprintf("%s := %s", lengthTemp, g.staticArrayLengthExpr(objTemp)),
			fmt.Sprintf("if %s < 0 || %s >= %s {", indexTemp, indexTemp, lengthTemp),
		)
		lines = append(lines, indentLines(nilTransferLines, 1)...)
		lines = append(lines, "}")
		lines = append(lines, elemLines...)
		lines = append(lines, fmt.Sprintf("var %s %s = %s", resultTemp, resultType, elemExpr))
		return lines, resultTemp, resultType, true
	case "first", "last":
		if len(call.Arguments) != 0 {
			return nil, "", "", false
		}
		indexExpr := "0"
		lines = append(lines, fmt.Sprintf("%s := %s", lengthTemp, g.staticArrayLengthExpr(objTemp)))
		lines = append(lines, fmt.Sprintf("if %s <= 0 {", lengthTemp))
		lines = append(lines, indentLines(nilTransferLines, 1)...)
		lines = append(lines, "}")
		if methodName == "last" {
			indexExpr = fmt.Sprintf("%s-1", lengthTemp)
		}
		elemLines, elemExpr, _, ok := g.staticArrayResultExprLines(ctx, objType, fmt.Sprintf("%s.Elements[%s]", objTemp, indexExpr), resultType)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, elemLines...)
		lines = append(lines, fmt.Sprintf("var %s %s = %s", resultTemp, resultType, elemExpr))
		return lines, resultTemp, resultType, true
	case "pop":
		if len(call.Arguments) != 0 {
			return nil, "", "", false
		}
		indexTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", lengthTemp, g.staticArrayLengthExpr(objTemp)))
		lines = append(lines, fmt.Sprintf("if %s <= 0 {", lengthTemp))
		lines = append(lines, indentLines(nilTransferLines, 1)...)
		lines = append(lines, "}")
		lines = append(lines, fmt.Sprintf("%s := %s - 1", indexTemp, lengthTemp))
		elemLines, elemExpr, _, ok := g.staticArrayResultExprLines(ctx, objType, fmt.Sprintf("%s.Elements[%s]", objTemp, indexTemp), resultType)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, elemLines...)
		lines = append(lines, fmt.Sprintf("var %s %s = %s", resultTemp, resultType, elemExpr))
		lines = append(lines,
			fmt.Sprintf("%s.Elements = %s.Elements[:%s]", objTemp, objTemp, indexTemp),
			g.staticArraySyncCall(objType, objTemp),
		)
		return lines, resultTemp, resultType, true
	}
	return nil, "", "", false
}

func (g *generator) compileOrElseExpression(ctx *compileContext, expr *ast.OrElseExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.Expression == nil || expr.Handler == nil {
		ctx.setReason("missing or-else expression")
		return nil, "", "", false
	}

	controlTemp := ctx.newTemp()
	valueDoneLabel := ctx.newTemp()
	valueCtx := ctx.child()
	valueCtx.controlCaptureVar = controlTemp
	valueCtx.controlCaptureLabel = valueDoneLabel
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(valueCtx, "", expr.Expression)
	if !ok {
		return nil, "", "", false
	}

	unionSuccessMember, unionFailureMember, valueErrorUnion := g.nativeUnionOrElseMembers(valueType)
	effectiveValueType := valueType
	if valueErrorUnion {
		effectiveValueType = unionSuccessMember.GoType
	}

	bindingName := ""
	bindingType := "runtime.Value"
	if expr.ErrorBinding != nil && expr.ErrorBinding.Name != "" {
		bindingName = expr.ErrorBinding.Name
		if valueErrorUnion && unionFailureMember != nil && unionFailureMember.GoType != "" {
			bindingType = unionFailureMember.GoType
		}
	}

	preferredType := expected
	if preferredType == "" {
		if innerType, ok := g.nativeNullableValueInnerType(effectiveValueType); ok {
			preferredType = innerType
		} else if effectiveValueType != "runtime.Value" {
			preferredType = effectiveValueType
		}
	}
	newHandlerCtx := func() *compileContext {
		handlerCtx := ctx.child()
		if bindingName != "" {
			handlerCtx.locals[bindingName] = paramInfo{Name: bindingName, GoName: sanitizeIdent(bindingName), GoType: bindingType}
		}
		return handlerCtx
	}
	handlerCtx := newHandlerCtx()
	handlerLines, handlerExpr, handlerType, ok := g.compileBlockExpression(handlerCtx, expr.Handler, preferredType)
	if !ok && expected == "" && preferredType != "" {
		handlerCtx = newHandlerCtx()
		handlerLines, handlerExpr, handlerType, ok = g.compileBlockExpression(handlerCtx, expr.Handler, "")
	}
	if !ok {
		return nil, "", "", false
	}
	valueTypeExpr, _ := g.inferExpressionTypeExpr(valueCtx, expr.Expression, valueType)
	handlerTypeExpr, _ := g.inferExpressionTypeExpr(handlerCtx, expr.Handler, handlerType)

	resultType := expected
	valueNullableInner, valueNullable := g.nativeNullableValueInnerType(effectiveValueType)
	valueJoinType := effectiveValueType
	if valueNullable {
		valueJoinType = valueNullableInner
	}
	if resultType == "" {
		switch {
		case valueType == "" && handlerType == "":
			resultType = "runtime.Value"
		case valueType == "":
			resultType = handlerType
		case handlerType == "":
			resultType = valueJoinType
		default:
			joinBranches := []joinBranchInfo{
				{
					GoType:   valueJoinType,
					Expr:     expr.Expression,
					TypeExpr: valueTypeExpr,
					SawNil:   g.joinBranchIsNilExpr(valueExpr, valueJoinType),
				},
				{
					GoType:   handlerType,
					Expr:     expr.Handler,
					TypeExpr: handlerTypeExpr,
					SawNil:   g.joinBranchIsNilExpr(handlerExpr, handlerType),
				},
			}
			if joinedType, ok := g.joinResultTypeFromBranches(ctx, joinBranches); ok {
				resultType = joinedType
			} else {
				resultType = "runtime.Value"
			}
		}
	}
	if resultType == "" {
		resultType = "runtime.Value"
	}

	handlerCoerceLines, handlerResultExpr, ok := g.coerceJoinBranch(ctx, resultType, handlerExpr, handlerType)
	if !ok {
		ctx.setReason("or-else type mismatch")
		return nil, "", "", false
	}

	resultTemp := ctx.newTemp()
	failedTemp := ctx.newTemp()
	valueTemp := ctx.newTemp()
	failureTemp := ""
	errorTemp := ""
	controlValueTemp := ""
	successTemp := ""
	successOkTemp := ""
	if bindingName != "" {
		failureTemp = ctx.newTemp()
		errorTemp = ctx.newTemp()
		if bindingType != "runtime.Value" {
			controlValueTemp = ctx.newTemp()
		}
	}
	if valueErrorUnion {
		successTemp = ctx.newTemp()
		successOkTemp = ctx.newTemp()
	}

	lines := []string{
		fmt.Sprintf("var %s %s", resultTemp, resultType),
		fmt.Sprintf("var %s bool", failedTemp),
		fmt.Sprintf("var %s %s", valueTemp, valueType),
		fmt.Sprintf("var %s *__ableControl", controlTemp),
	}
	if bindingName != "" {
		lines = append(lines, fmt.Sprintf("var %s %s", failureTemp, bindingType))
		lines = append(lines, fmt.Sprintf("var %s bool", errorTemp))
	}
	if successTemp != "" {
		lines = append(lines, fmt.Sprintf("var %s %s", successTemp, unionSuccessMember.GoType))
		lines = append(lines, fmt.Sprintf("var %s bool", successOkTemp))
	}
	lines = append(lines, "{")
	lines = append(lines, indentLines(valueLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", valueTemp, valueExpr))
	lines = append(lines, "}")
	lines = append(lines, fmt.Sprintf("if false { goto %s }", valueDoneLabel))
	lines = append(lines, fmt.Sprintf("%s:", valueDoneLabel))

	if bindingName != "" {
		if bindingType == "runtime.Value" {
			lines = append(lines, fmt.Sprintf("if %s != nil { %s = __able_control_value(%s); %s = true; %s = true }", controlTemp, failureTemp, controlTemp, failedTemp, errorTemp))
		} else {
			controlFailureLines, convertedFailure, ok := g.expectRuntimeValueExprLines(ctx, controlValueTemp, bindingType)
			if !ok {
				ctx.setReason("or-else binding type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, fmt.Sprintf("if %s != nil {", controlTemp))
			lines = append(lines, fmt.Sprintf("\t%s := __able_control_value(%s)", controlValueTemp, controlTemp))
			lines = append(lines, indentLines(controlFailureLines, 1)...)
			lines = append(lines, fmt.Sprintf("\t%s = %s", failureTemp, convertedFailure))
			lines = append(lines, fmt.Sprintf("\t%s = true", failedTemp))
			lines = append(lines, fmt.Sprintf("\t%s = true", errorTemp))
			lines = append(lines, "}")
		}
	} else {
		lines = append(lines, fmt.Sprintf("if %s != nil { %s = true }", controlTemp, failedTemp))
	}

	successExpr := valueTemp
	successType := effectiveValueType
	nullableCheckExpr := valueTemp
	if valueErrorUnion {
		failureNativeTemp := "_"
		if bindingName != "" {
			failureNativeTemp = ctx.newTemp()
		}
		failureOkTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("if %s == nil {", controlTemp))
		lines = append(lines, fmt.Sprintf("\t%s, %s := %s(%s)", failureNativeTemp, failureOkTemp, unionFailureMember.UnwrapHelper, valueTemp))
		if bindingName != "" {
			lines = append(lines, fmt.Sprintf("\tif %s {", failureOkTemp))
			if bindingType == "runtime.Value" {
				failureRuntimeLines, failureRuntimeExpr, ok := g.runtimeValueLines(ctx, failureNativeTemp, unionFailureMember.GoType)
				if !ok {
					ctx.setReason("or-else union error conversion mismatch")
					return nil, "", "", false
				}
				lines = append(lines, indentLines(failureRuntimeLines, 2)...)
				lines = append(lines, fmt.Sprintf("\t\t%s = %s", failureTemp, failureRuntimeExpr))
			} else {
				lines = append(lines, fmt.Sprintf("\t\t%s = %s", failureTemp, failureNativeTemp))
			}
			lines = append(lines, fmt.Sprintf("\t\t%s = true", failedTemp))
			lines = append(lines, fmt.Sprintf("\t\t%s = true", errorTemp))
			lines = append(lines, "\t}")
		} else {
			lines = append(lines, fmt.Sprintf("\tif %s { %s = true }", failureOkTemp, failedTemp))
		}
		invariantLines, ok := g.controlTransferLines(ctx, g.runtimeErrorControlExpr("nil", `fmt.Errorf("compiler: native union or-else success branch missing")`))
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, fmt.Sprintf("\tif !%s {", failureOkTemp))
		lines = append(lines, fmt.Sprintf("\t\t%s, %s = %s(%s)", successTemp, successOkTemp, unionSuccessMember.UnwrapHelper, valueTemp))
		lines = append(lines, fmt.Sprintf("\t\tif !%s {", successOkTemp))
		lines = append(lines, indentLines(invariantLines, 3)...)
		lines = append(lines, "\t\t}")
		lines = append(lines, "\t}")
		lines = append(lines, "}")
		successExpr = successTemp
		successType = unionSuccessMember.GoType
		nullableCheckExpr = successTemp
	}

	if valueNullable {
		successExpr = fmt.Sprintf("(*%s)", successExpr)
		successType = valueNullableInner
	}
	successConvLines, successExpr, ok := g.coerceJoinBranch(ctx, resultType, successExpr, successType)
	if !ok {
		ctx.setReason("or-else type mismatch")
		return nil, "", "", false
	}

	if valueNullable {
		if bindingName != "" {
			lines = append(lines, fmt.Sprintf("if %s == nil && %s == nil { %s = runtime.NilValue{}; %s = true }", nullableCheckExpr, controlTemp, failureTemp, failedTemp))
		} else {
			lines = append(lines, fmt.Sprintf("if %s == nil && %s == nil { %s = true }", nullableCheckExpr, controlTemp, failedTemp))
		}
	}
	if !valueErrorUnion && (valueType == "runtime.Value" || valueType == "any") {
		checkTemp := valueTemp
		if valueType == "any" {
			checkTemp = ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", checkTemp, valueTemp))
		}
		if bindingName != "" {
			lines = append(lines, fmt.Sprintf("if __able_is_nil(%s) && %s == nil { %s = runtime.NilValue{}; %s = true }", checkTemp, controlTemp, failureTemp, failedTemp))
			lines = append(lines, fmt.Sprintf("if __able_is_error(%s) && %s == nil { %s = %s; %s = true; %s = true }", checkTemp, controlTemp, failureTemp, checkTemp, failedTemp, errorTemp))
		} else {
			lines = append(lines, fmt.Sprintf("if __able_is_nil(%s) && %s == nil { %s = true }", checkTemp, controlTemp, failedTemp))
			lines = append(lines, fmt.Sprintf("if __able_is_error(%s) && %s == nil { %s = true }", checkTemp, controlTemp, failedTemp))
		}
	}
	lines = append(lines, fmt.Sprintf("if !%s {", failedTemp))
	lines = append(lines, indentLines(successConvLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, successExpr))
	lines = append(lines, "}")

	lines = append(lines, fmt.Sprintf("if %s {", failedTemp))
	if bindingName != "" {
		goName := sanitizeIdent(bindingName)
		lines = append(lines, fmt.Sprintf("\tvar %s %s", goName, bindingType))
		lines = append(lines, fmt.Sprintf("\tif %s { %s = %s }", errorTemp, goName, failureTemp))
		lines = append(lines, fmt.Sprintf("\t_ = %s", goName))
	}
	lines = append(lines, indentLines(handlerLines, 1)...)
	lines = append(lines, indentLines(handlerCoerceLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, handlerResultExpr))
	lines = append(lines, "}")
	return lines, resultTemp, resultType, true
}
