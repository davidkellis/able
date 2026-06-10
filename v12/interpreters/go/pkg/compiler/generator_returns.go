package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) compileReturnStatement(ctx *compileContext, returnType string, ret *ast.ReturnStatement, lines []string) ([]string, string, bool) {
	if ret == nil {
		ctx.setReason("missing return")
		return nil, "", false
	}
	if ret.Argument == nil {
		if g.isVoidType(returnType) {
			return lines, "struct{}{}", true
		}
		if g.isNilType(returnType) {
			return lines, "runtime.NilValue{}", true
		}
		if successExpr, ok := g.nativeResultVoidSuccessExpr(ctx, returnType); ok {
			return lines, successExpr, true
		}
		ctx.setReason("missing return expression")
		return nil, "", false
	}
	if g.isVoidType(returnType) {
		stmtLines, valueExpr, valueType, ok := g.compileDiscardedTailExpression(ctx, ret.Argument)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, stmtLines...)
		if valueExpr != "" {
			lines, ok = g.discardStatementResult(ctx, lines, valueExpr, valueType)
			if !ok {
				return nil, "", false
			}
		}
		return lines, "struct{}{}", true
	}
	previousExpectedTypeExpr := ctx.expectedTypeExpr
	ctx.expectedTypeExpr = g.concretizedExpectedTypeExpr(ctx, returnType, ctx.returnTypeExpr)
	previousTailExpr := ctx.callerOwnedTailExpr
	if ctx.callerOwnedResultSlot != "" {
		ctx.callerOwnedTailExpr = ret.Argument
	}
	exprLines, expr, exprType, ok := g.compileTailExpression(ctx, returnType, ret.Argument)
	ctx.callerOwnedTailExpr = previousTailExpr
	ctx.expectedTypeExpr = previousExpectedTypeExpr
	if !ok {
		return nil, "", false
	}
	if returnType == "runtime.Value" {
		if ifaceType, ok := g.interfaceTypeExpr(ctx.returnTypeExpr); ok {
			if exprType != "runtime.Value" {
				convLines, converted, ok := g.lowerRuntimeValue(ctx, expr, exprType)
				if !ok {
					ctx.setReason("return type mismatch")
					return nil, "", false
				}
				exprLines = append(exprLines, convLines...)
				expr = converted
			}
			ifaceLines, coerced, ok := g.interfaceReturnExprLines(ctx, expr, ifaceType, ctx.genericNames)
			if !ok {
				ctx.setReason("return type mismatch")
				return nil, "", false
			}
			exprLines = append(exprLines, ifaceLines...)
			expr = coerced
		}
	}
	lines = append(lines, exprLines...)
	return lines, expr, true
}

func (g *generator) compileImplicitReturn(ctx *compileContext, returnType string, expr ast.Expression, lines []string) ([]string, string, bool) {
	if g.isVoidType(returnType) {
		var (
			stmtLines []string
			valueExpr string
			valueType string
			ok        bool
		)
		if assign, isAssignment := expr.(*ast.AssignmentExpression); isAssignment {
			stmtLines, valueExpr, valueType, ok = g.compileAssignmentMode(ctx, assign, true)
		} else {
			stmtLines, valueExpr, valueType, ok = g.compileDiscardedTailExpression(ctx, expr)
		}
		if !ok {
			return nil, "", false
		}
		lines = append(lines, stmtLines...)
		if valueExpr != "" {
			lines, ok = g.discardStatementResult(ctx, lines, valueExpr, valueType)
			if !ok {
				return nil, "", false
			}
		}
		return lines, "struct{}{}", true
	}
	if g.isNilType(returnType) {
		if _, ok := expr.(*ast.NilLiteral); ok {
			return lines, "runtime.NilValue{}", true
		}
	}
	previousExpectedTypeExpr := ctx.expectedTypeExpr
	ctx.expectedTypeExpr = g.concretizedExpectedTypeExpr(ctx, returnType, ctx.returnTypeExpr)
	previousTailExpr := ctx.callerOwnedTailExpr
	if ctx.callerOwnedResultSlot != "" {
		ctx.callerOwnedTailExpr = expr
	}
	stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, returnType, expr)
	ctx.callerOwnedTailExpr = previousTailExpr
	ctx.expectedTypeExpr = previousExpectedTypeExpr
	if !ok {
		return nil, "", false
	}
	if returnType == "runtime.Value" && valueType != "runtime.Value" {
		convLines, converted, ok := g.lowerRuntimeValue(ctx, valueExpr, valueType)
		if !ok {
			ctx.setReason("return type mismatch")
			return nil, "", false
		}
		stmtLines = append(stmtLines, convLines...)
		valueExpr = converted
		valueType = "runtime.Value"
	} else if returnType != "" && returnType != "runtime.Value" && returnType != "any" && returnType != valueType && g.canCoerceStaticExpr(returnType, valueType) {
		coercedLines, coercedExpr, coercedType, ok := g.lowerCoerceExpectedStaticExpr(ctx, stmtLines, valueExpr, valueType, returnType)
		if !ok {
			ctx.setReason("assignment return type mismatch")
			return nil, "", false
		}
		stmtLines = coercedLines
		valueExpr = coercedExpr
		valueType = coercedType
	} else if !g.typeMatches(returnType, valueType) {
		if returnType != "" && returnType != "runtime.Value" && returnType != "any" && g.canCoerceStaticExpr(returnType, valueType) {
			coercedLines, coercedExpr, coercedType, ok := g.lowerCoerceExpectedStaticExpr(ctx, stmtLines, valueExpr, valueType, returnType)
			if !ok {
				ctx.setReason("assignment return type mismatch")
				return nil, "", false
			}
			stmtLines = coercedLines
			valueExpr = coercedExpr
			valueType = coercedType
		} else {
			ctx.setReason("assignment return type mismatch")
			return nil, "", false
		}
	}
	if returnType == "runtime.Value" {
		if ifaceType, ok := g.interfaceTypeExpr(ctx.returnTypeExpr); ok {
			ifaceLines, coerced, ok := g.interfaceReturnExprLines(ctx, valueExpr, ifaceType, ctx.genericNames)
			if !ok {
				ctx.setReason("return type mismatch")
				return nil, "", false
			}
			stmtLines = append(stmtLines, ifaceLines...)
			valueExpr = coerced
		}
	}
	lines = append(lines, stmtLines...)
	return lines, valueExpr, true
}
