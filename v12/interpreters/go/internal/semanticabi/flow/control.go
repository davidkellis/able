package flow

import (
	"fmt"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/pkg/ast"
)

func (lower *lowerer) lowerIf(expression *ast.IfExpression) (uint32, bool, error) {
	condition, err := lower.lowerExpression(expression.IfCondition)
	if err != nil {
		return 0, false, err
	}
	result := lower.newRegister(lower.dynamicType())
	origin := lower.current
	thenBlock := lower.newBlock()
	elseBlock := lower.newBlock()
	if err := lower.emitTo(origin, expression, semanticabi.OpBranch, condition, thenBlock, elseBlock); err != nil {
		return 0, false, err
	}

	fallBlocks := make([]uint32, 0, 2)
	lower.setCurrent(thenBlock)
	thenValue, thenProduced, err := lower.lowerBlock(expression.IfBody)
	if err != nil {
		return 0, false, err
	}
	if !lower.currentBlock().terminated {
		if !thenProduced {
			thenValue, err = lower.emitVoid(expression.IfBody)
			if err != nil {
				return 0, false, err
			}
		}
		if err := lower.emit(expression.IfBody, semanticabi.OpMoveValue, result, thenValue); err != nil {
			return 0, false, err
		}
		fallBlocks = append(fallBlocks, lower.current)
	}

	lower.setCurrent(elseBlock)
	var elseValue uint32
	elseProduced := false
	if len(expression.ElseIfClauses) != 0 {
		return 0, false, fmt.Errorf("semanticabi flow: elsif lowering is not admitted")
	}
	if expression.ElseBody != nil {
		elseValue, elseProduced, err = lower.lowerBlock(expression.ElseBody)
		if err != nil {
			return 0, false, err
		}
	}
	if !lower.currentBlock().terminated {
		if !elseProduced {
			elseValue, err = lower.emitNil(expression)
			if err != nil {
				return 0, false, err
			}
		}
		if err := lower.emit(expression, semanticabi.OpMoveValue, result, elseValue); err != nil {
			return 0, false, err
		}
		fallBlocks = append(fallBlocks, lower.current)
	}
	if len(fallBlocks) == 0 {
		return 0, false, nil
	}
	merge := lower.newBlock()
	for _, blockID := range fallBlocks {
		if err := lower.emitTo(blockID, expression, semanticabi.OpJump, merge); err != nil {
			return 0, false, err
		}
	}
	lower.setCurrent(merge)
	return result, true, nil
}

func (lower *lowerer) lowerLoop(expression *ast.LoopExpression) (uint32, bool, error) {
	origin := lower.current
	result := lower.newRegister(lower.dynamicType())
	header := lower.newBlock()
	exit := lower.newBlock()
	if err := lower.emitTo(origin, expression, semanticabi.OpJump, header); err != nil {
		return 0, false, err
	}
	lower.loops = append(lower.loops, loopTarget{continueBlock: header, breakBlock: exit, resultRegister: result})
	lower.setCurrent(header)
	_, _, err := lower.lowerBlock(expression.Body)
	if err != nil {
		return 0, false, err
	}
	if !lower.currentBlock().terminated {
		if err := lower.emit(expression, semanticabi.OpJump, header); err != nil {
			return 0, false, err
		}
	}
	lower.loops = lower.loops[:len(lower.loops)-1]
	lower.setCurrent(exit)
	return result, true, nil
}

func (lower *lowerer) lowerBreak(statement *ast.BreakStatement) (uint32, bool, error) {
	if len(lower.loops) == 0 {
		return 0, false, fmt.Errorf("semanticabi flow: break outside loop")
	}
	if statement.Label != nil {
		return 0, false, fmt.Errorf("semanticabi flow: labeled break is not admitted")
	}
	loop := lower.loops[len(lower.loops)-1]
	var value uint32
	var err error
	if statement.Value == nil {
		value, err = lower.emitNil(statement)
	} else {
		value, err = lower.lowerExpression(statement.Value)
	}
	if err != nil {
		return 0, false, err
	}
	if err := lower.emit(statement, semanticabi.OpMoveValue, loop.resultRegister, value); err != nil {
		return 0, false, err
	}
	return 0, false, lower.emit(statement, semanticabi.OpJump, loop.breakBlock)
}

func (lower *lowerer) lowerContinue(statement *ast.ContinueStatement) (uint32, bool, error) {
	if len(lower.loops) == 0 {
		return 0, false, fmt.Errorf("semanticabi flow: continue outside loop")
	}
	if statement.Label != nil {
		return 0, false, fmt.Errorf("semanticabi flow: labeled continue is not admitted")
	}
	target := lower.loops[len(lower.loops)-1].continueBlock
	return 0, false, lower.emit(statement, semanticabi.OpJump, target)
}

func (lower *lowerer) lowerReturn(statement *ast.ReturnStatement) (uint32, bool, error) {
	var value uint32
	var err error
	if statement.Argument == nil {
		value, err = lower.emitVoid(statement)
	} else {
		value, err = lower.lowerExpression(statement.Argument)
	}
	if err != nil {
		return 0, false, err
	}
	return 0, false, lower.emit(statement, semanticabi.OpReturnValue, value)
}

func (lower *lowerer) lowerRaise(statement *ast.RaiseStatement) (uint32, bool, error) {
	value, err := lower.lowerExpression(statement.Expression)
	if err != nil {
		return 0, false, err
	}
	return 0, false, lower.emit(statement, semanticabi.OpRaiseValue, value)
}

func (lower *lowerer) lowerMatch(expression *ast.MatchExpression) (uint32, bool, error) {
	subject, err := lower.lowerExpression(expression.Subject)
	if err != nil {
		return 0, false, err
	}
	result := lower.newRegister(lower.dynamicType())
	testBlock := lower.current
	fallBlocks := make([]uint32, 0, len(expression.Clauses))
	for _, clause := range expression.Clauses {
		pattern, ok := clause.Pattern.(*ast.TypedPattern)
		if !ok || pattern == nil {
			return 0, false, fmt.Errorf("semanticabi flow: match pattern %T is not a typed pattern", clause.Pattern)
		}
		if clause.Guard != nil {
			return 0, false, fmt.Errorf("semanticabi flow: guarded match clause is not admitted")
		}
		typeID := lower.tables.typeIndex(typeName(pattern.TypeAnnotation))
		condition := lower.newRegister(lower.boolType())
		bodyBlock := lower.newBlock()
		nextBlock := lower.newBlock()
		if err := lower.emitTo(testBlock, clause, semanticabi.OpTypeTest, condition, subject, typeID); err != nil {
			return 0, false, err
		}
		if err := lower.emitTo(testBlock, clause, semanticabi.OpBranch, condition, bodyBlock, nextBlock); err != nil {
			return 0, false, err
		}

		lower.setCurrent(bodyBlock)
		lower.pushScope()
		if name, ok := patternBindingName(pattern.Pattern); ok {
			bound := binding{register: lower.newRegister(typeID), typeID: typeID}
			lower.bind(name, bound)
			if err := lower.emit(pattern, semanticabi.OpMoveValue, bound.register, subject); err != nil {
				lower.popScope()
				return 0, false, err
			}
		}
		bodyValue, produced, err := lower.lowerMatchBody(clause.Body)
		lower.popScope()
		if err != nil {
			return 0, false, err
		}
		if !lower.currentBlock().terminated {
			if !produced {
				bodyValue, err = lower.emitVoid(clause)
				if err != nil {
					return 0, false, err
				}
			}
			if err := lower.emit(clause, semanticabi.OpMoveValue, result, bodyValue); err != nil {
				return 0, false, err
			}
			fallBlocks = append(fallBlocks, lower.current)
		}
		testBlock = nextBlock
	}
	lower.setCurrent(testBlock)
	if err := lower.emit(expression, semanticabi.OpMatchFail); err != nil {
		return 0, false, err
	}
	if len(fallBlocks) == 0 {
		return 0, false, nil
	}
	merge := lower.newBlock()
	for _, blockID := range fallBlocks {
		if err := lower.emitTo(blockID, expression, semanticabi.OpJump, merge); err != nil {
			return 0, false, err
		}
	}
	lower.setCurrent(merge)
	return result, true, nil
}

func (lower *lowerer) lowerMatchBody(expression ast.Expression) (uint32, bool, error) {
	if block, ok := expression.(*ast.BlockExpression); ok {
		return lower.lowerBlock(block)
	}
	value, err := lower.lowerExpression(expression)
	return value, err == nil, err
}
