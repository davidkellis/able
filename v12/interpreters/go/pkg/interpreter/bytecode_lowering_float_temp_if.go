package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func emitFloatMulAddMulCompareConstTempIfStatement(ctx *bytecodeLoweringContext, i *Interpreter, statements []ast.Statement) (bool, error) {
	if ctx == nil || ctx.frameLayout == nil || len(statements) < 4 {
		return false, nil
	}
	firstAssign, ok := statements[0].(*ast.AssignmentExpression)
	if !ok || firstAssign == nil || firstAssign.Operator != ast.AssignmentDeclare {
		return false, nil
	}
	secondAssign, ok := statements[1].(*ast.AssignmentExpression)
	if !ok || secondAssign == nil || secondAssign.Operator != ast.AssignmentDeclare {
		return false, nil
	}
	ifExpr, ok := statements[2].(*ast.IfExpression)
	if !ok || ifExpr == nil || len(ifExpr.ElseIfClauses) != 0 || ifExpr.ElseBody != nil {
		return false, nil
	}

	leftName, ok := bytecodeSimpleDeclaredIdentifierName(firstAssign.Left)
	if !ok {
		return false, nil
	}
	rightName, ok := bytecodeSimpleDeclaredIdentifierName(secondAssign.Left)
	if !ok || leftName == rightName {
		return false, nil
	}
	if !bytecodeBlockTerminatesWithoutFallingThrough(ifExpr.IfBody) {
		return false, nil
	}
	if bytecodeBlockReferencesAnyIdentifier(ifExpr.IfBody, leftName, rightName) {
		return false, nil
	}
	if !bytecodeStatementsReferenceAnyIdentifier(statements[3:], leftName, rightName) {
		return false, nil
	}

	leftMulLeftSlot, leftMulRightSlot, ok := bytecodeFloatIdentifierMultiplySlots(ctx, firstAssign.Right)
	if !ok {
		return false, nil
	}
	rightMulLeftSlot, rightMulRightSlot, ok := bytecodeFloatIdentifierMultiplySlots(ctx, secondAssign.Right)
	if !ok {
		return false, nil
	}

	leftTargetSlot := ctx.declareSlot(leftName)
	rightTargetSlot := ctx.declareSlot(rightName)
	instr, plan, ok := bytecodeJumpIfFalseFloatMulAddMulCompareConstTempInstruction(ifExpr.IfCondition, leftName, rightName)
	if !ok {
		return false, nil
	}
	plan.leftMulLeftSlot = leftMulLeftSlot
	plan.leftMulRightSlot = leftMulRightSlot
	plan.rightMulLeftSlot = rightMulLeftSlot
	plan.rightMulRightSlot = rightMulRightSlot
	plan.storeProducts = true
	plan.leftTargetSlot = leftTargetSlot
	plan.rightTargetSlot = rightTargetSlot

	jump := ctx.emit(instr)
	ctx.setFloatMulAddMulCompareConstJumpPlan(jump, plan)

	bodyStart := len(ctx.instructions)
	if err := emitBlock(ctx, i, ifExpr.IfBody); err != nil {
		return false, err
	}
	if !bytecodeDiscardTrailingBlockResult(ctx, ifExpr.IfBody, bodyStart) {
		ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	}
	ctx.patchJump(jump, len(ctx.instructions))
	return true, nil
}

func bytecodeJumpIfFalseFloatMulAddMulCompareConstTempInstruction(expr ast.Expression, leftName string, rightName string) (bytecodeInstruction, bytecodeFloatMulAddMulCompareConstJumpPlan, bool) {
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil || !bytecodeComparisonOperator(binary.Operator) {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}
	add, ok := binary.Left.(*ast.BinaryExpression)
	if !ok || add == nil || add.Operator != "+" {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}
	leftIdent, ok := add.Left.(*ast.Identifier)
	if !ok || leftIdent == nil || leftIdent.Name != leftName {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}
	rightIdent, ok := add.Right.(*ast.Identifier)
	if !ok || rightIdent == nil || rightIdent.Name != rightName {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}
	rightLiteral, ok := binary.Right.(*ast.FloatLiteral)
	if !ok || rightLiteral == nil {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}

	rightKind := runtime.FloatF64
	if rightLiteral.FloatType != nil {
		rightKind = runtime.FloatType(*rightLiteral.FloatType)
	}
	if rightKind != runtime.FloatF32 && rightKind != runtime.FloatF64 {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}

	return bytecodeInstruction{
			op:       bytecodeOpJumpIfFloatMulAddMulCompareConstFalse,
			target:   -1,
			operator: binary.Operator,
			node:     binary,
		}, bytecodeFloatMulAddMulCompareConstJumpPlan{
			rightImmediate: runtime.FloatValue{
				Val:        normalizeFloat(rightKind, rightLiteral.Value),
				TypeSuffix: rightKind,
			},
		}, true
}

func bytecodeSimpleDeclaredIdentifierName(target ast.AssignmentTarget) (string, bool) {
	ident, ok := target.(*ast.Identifier)
	if !ok || ident == nil || ident.Name == "" {
		return "", false
	}
	return ident.Name, true
}

func bytecodeBlockTerminatesWithoutFallingThrough(block *ast.BlockExpression) bool {
	if block == nil || len(block.Body) == 0 {
		return false
	}
	return bytecodeStatementTerminatesWithoutFallingThrough(block.Body[len(block.Body)-1])
}

func bytecodeStatementTerminatesWithoutFallingThrough(stmt ast.Statement) bool {
	switch stmt.(type) {
	case *ast.BreakStatement, *ast.ContinueStatement, *ast.ReturnStatement, *ast.RaiseStatement, *ast.RethrowStatement:
		return true
	default:
		return false
	}
}

func bytecodeStatementsReferenceAnyIdentifier(statements []ast.Statement, names ...string) bool {
	nameSet := bytecodeIdentifierNameSet(names...)
	if len(nameSet) == 0 {
		return false
	}
	for _, stmt := range statements {
		if bytecodeStatementReferencesIdentifierSet(stmt, nameSet) {
			return true
		}
	}
	return false
}

func bytecodeBlockReferencesAnyIdentifier(block *ast.BlockExpression, names ...string) bool {
	if block == nil {
		return false
	}
	return bytecodeStatementsReferenceAnyIdentifier(block.Body, names...)
}

func bytecodeStatementReferencesIdentifierSet(stmt ast.Statement, names map[string]struct{}) bool {
	if stmt == nil {
		return false
	}
	if expr, ok := stmt.(ast.Expression); ok {
		return bytecodeExpressionReferencesIdentifierSet(expr, names)
	}
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		return bytecodeExpressionReferencesIdentifierSet(s.Argument, names)
	case *ast.RaiseStatement:
		return bytecodeExpressionReferencesIdentifierSet(s.Expression, names)
	case *ast.ForLoop:
		if bytecodePatternReferencesIdentifierSet(s.Pattern, names) {
			return true
		}
		return bytecodeExpressionReferencesIdentifierSet(s.Iterable, names) || bytecodeBlockReferencesIdentifierSet(s.Body, names)
	case *ast.WhileLoop:
		return bytecodeExpressionReferencesIdentifierSet(s.Condition, names) || bytecodeBlockReferencesIdentifierSet(s.Body, names)
	case *ast.BreakStatement:
		return bytecodeExpressionReferencesIdentifierSet(s.Value, names)
	case *ast.YieldStatement:
		return bytecodeExpressionReferencesIdentifierSet(s.Expression, names)
	case *ast.ContinueStatement, *ast.PreludeStatement, *ast.ExternFunctionBody, *ast.ImportStatement, *ast.DynImportStatement, *ast.PackageStatement:
		return false
	default:
		return true
	}
}

func bytecodeBlockReferencesIdentifierSet(block *ast.BlockExpression, names map[string]struct{}) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.Body {
		if bytecodeStatementReferencesIdentifierSet(stmt, names) {
			return true
		}
	}
	return false
}

func bytecodeExpressionReferencesIdentifierSet(expr ast.Expression, names map[string]struct{}) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		_, ok := names[e.Name]
		return ok
	case *ast.BinaryExpression:
		return bytecodeExpressionReferencesIdentifierSet(e.Left, names) || bytecodeExpressionReferencesIdentifierSet(e.Right, names)
	case *ast.UnaryExpression:
		return bytecodeExpressionReferencesIdentifierSet(e.Operand, names)
	case *ast.TypeCastExpression:
		return bytecodeExpressionReferencesIdentifierSet(e.Expression, names)
	case *ast.FunctionCall:
		if bytecodeExpressionReferencesIdentifierSet(e.Callee, names) {
			return true
		}
		for _, arg := range e.Arguments {
			if bytecodeExpressionReferencesIdentifierSet(arg, names) {
				return true
			}
		}
		return false
	case *ast.MemberAccessExpression:
		if bytecodeExpressionReferencesIdentifierSet(e.Object, names) {
			return true
		}
		if memberExpr, ok := e.Member.(ast.Expression); ok {
			return bytecodeExpressionReferencesIdentifierSet(memberExpr, names)
		}
		return false
	case *ast.ImplicitMemberExpression:
		return false
	case *ast.IndexExpression:
		return bytecodeExpressionReferencesIdentifierSet(e.Object, names) || bytecodeExpressionReferencesIdentifierSet(e.Index, names)
	case *ast.BlockExpression:
		return bytecodeBlockReferencesIdentifierSet(e, names)
	case *ast.LoopExpression:
		return bytecodeBlockReferencesIdentifierSet(e.Body, names)
	case *ast.AssignmentExpression:
		if bytecodeAssignmentTargetReferencesIdentifierSet(e.Left, names) {
			return true
		}
		return bytecodeExpressionReferencesIdentifierSet(e.Right, names)
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			if bytecodeExpressionReferencesIdentifierSet(part, names) {
				return true
			}
		}
		return false
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			if field == nil {
				continue
			}
			if field.IsShorthand && field.Value == nil && field.Name != nil {
				if _, ok := names[field.Name.Name]; ok {
					return true
				}
			}
			if bytecodeExpressionReferencesIdentifierSet(field.Value, names) {
				return true
			}
		}
		for _, src := range e.FunctionalUpdateSources {
			if bytecodeExpressionReferencesIdentifierSet(src, names) {
				return true
			}
		}
		return false
	case *ast.ArrayLiteral:
		for _, el := range e.Elements {
			if bytecodeExpressionReferencesIdentifierSet(el, names) {
				return true
			}
		}
		return false
	case *ast.RangeExpression:
		return bytecodeExpressionReferencesIdentifierSet(e.Start, names) || bytecodeExpressionReferencesIdentifierSet(e.End, names)
	case *ast.MatchExpression:
		if bytecodeExpressionReferencesIdentifierSet(e.Subject, names) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil && bytecodeExpressionReferencesIdentifierSet(clause.Guard, names) {
				return true
			}
			if bytecodeExpressionReferencesIdentifierSet(clause.Body, names) {
				return true
			}
		}
		return false
	case *ast.OrElseExpression:
		return bytecodeExpressionReferencesIdentifierSet(e.Expression, names) || bytecodeBlockReferencesIdentifierSet(e.Handler, names)
	case *ast.RescueExpression:
		if bytecodeExpressionReferencesIdentifierSet(e.MonitoredExpression, names) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil && bytecodeExpressionReferencesIdentifierSet(clause.Guard, names) {
				return true
			}
			if bytecodeExpressionReferencesIdentifierSet(clause.Body, names) {
				return true
			}
		}
		return false
	case *ast.EnsureExpression:
		return bytecodeExpressionReferencesIdentifierSet(e.TryExpression, names) || bytecodeBlockReferencesIdentifierSet(e.EnsureBlock, names)
	case *ast.IfExpression:
		if bytecodeExpressionReferencesIdentifierSet(e.IfCondition, names) || bytecodeBlockReferencesIdentifierSet(e.IfBody, names) {
			return true
		}
		for _, clause := range e.ElseIfClauses {
			if clause == nil {
				continue
			}
			if bytecodeExpressionReferencesIdentifierSet(clause.Condition, names) || bytecodeBlockReferencesIdentifierSet(clause.Body, names) {
				return true
			}
		}
		return bytecodeBlockReferencesIdentifierSet(e.ElseBody, names)
	case *ast.PropagationExpression:
		return bytecodeExpressionReferencesIdentifierSet(e.Expression, names)
	case *ast.AwaitExpression:
		return bytecodeExpressionReferencesIdentifierSet(e.Expression, names)
	case *ast.SpawnExpression:
		return true
	case *ast.LambdaExpression, *ast.IteratorLiteral:
		return true
	case *ast.IntegerLiteral,
		*ast.FloatLiteral,
		*ast.BooleanLiteral,
		*ast.StringLiteral,
		*ast.CharLiteral,
		*ast.NilLiteral,
		*ast.PlaceholderExpression:
		return false
	default:
		return true
	}
}

func bytecodeAssignmentTargetReferencesIdentifierSet(target ast.AssignmentTarget, names map[string]struct{}) bool {
	if target == nil {
		return false
	}
	if ident, ok := target.(*ast.Identifier); ok && ident != nil {
		_, found := names[ident.Name]
		return found
	}
	if expr, ok := target.(ast.Expression); ok && expr != nil {
		return bytecodeExpressionReferencesIdentifierSet(expr, names)
	}
	if pattern, ok := target.(ast.Pattern); ok {
		return bytecodePatternReferencesIdentifierSet(pattern, names)
	}
	return true
}

func bytecodePatternReferencesIdentifierSet(pattern ast.Pattern, names map[string]struct{}) bool {
	if pattern == nil {
		return false
	}
	switch p := pattern.(type) {
	case *ast.Identifier:
		_, ok := names[p.Name]
		return ok
	case *ast.TypedPattern:
		return bytecodePatternReferencesIdentifierSet(p.Pattern, names)
	case *ast.ArrayPattern:
		for _, element := range p.Elements {
			if bytecodePatternReferencesIdentifierSet(element, names) {
				return true
			}
		}
		return bytecodePatternReferencesIdentifierSet(p.RestPattern, names)
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil {
				if _, ok := names[field.Binding.Name]; ok {
					return true
				}
			}
			if bytecodePatternReferencesIdentifierSet(field.Pattern, names) {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func bytecodeIdentifierNameSet(names ...string) map[string]struct{} {
	if len(names) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		if name != "" {
			set[name] = struct{}{}
		}
	}
	return set
}
