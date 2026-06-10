package flow

import (
	"fmt"
	"strings"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/pkg/ast"
)

func (lower *lowerer) lowerBlock(block *ast.BlockExpression) (uint32, bool, error) {
	if block == nil {
		return 0, false, fmt.Errorf("semanticabi flow: nil block")
	}
	lower.pushScope()
	defer lower.popScope()
	var result uint32
	hasResult := false
	for _, statement := range block.Body {
		if lower.currentBlock().terminated {
			break
		}
		value, produced, err := lower.lowerStatement(statement)
		if err != nil {
			return 0, false, err
		}
		if produced {
			result, hasResult = value, true
		}
	}
	return result, hasResult, nil
}

func (lower *lowerer) lowerStatement(statement ast.Statement) (uint32, bool, error) {
	switch value := statement.(type) {
	case *ast.AssignmentExpression:
		result, err := lower.lowerAssignment(value)
		return result, true, err
	case *ast.IfExpression:
		return lower.lowerIf(value)
	case *ast.LoopExpression:
		return lower.lowerLoop(value)
	case *ast.MatchExpression:
		return lower.lowerMatch(value)
	case *ast.BreakStatement:
		return lower.lowerBreak(value)
	case *ast.ContinueStatement:
		return lower.lowerContinue(value)
	case *ast.ReturnStatement:
		return lower.lowerReturn(value)
	case *ast.RaiseStatement:
		return lower.lowerRaise(value)
	case ast.Expression:
		result, err := lower.lowerExpression(value)
		return result, true, err
	default:
		return 0, false, fmt.Errorf("semanticabi flow: unsupported statement %T", statement)
	}
}

func (lower *lowerer) lowerExpression(expression ast.Expression) (uint32, error) {
	if register, handled, err := lower.emitLiteral(expression); handled {
		return register, err
	}
	switch value := expression.(type) {
	case *ast.Identifier:
		return lower.lowerIdentifier(value)
	case *ast.AssignmentExpression:
		return lower.lowerAssignment(value)
	case *ast.UnaryExpression:
		return lower.lowerUnary(value)
	case *ast.BinaryExpression:
		return lower.lowerBinary(value)
	case *ast.TypeCastExpression:
		return lower.lowerCast(value)
	case *ast.FunctionCall:
		return lower.lowerCall(value)
	case *ast.MemberAccessExpression:
		return lower.lowerMember(value)
	case *ast.IfExpression:
		result, produced, err := lower.lowerIf(value)
		if !produced && err == nil {
			return 0, fmt.Errorf("semanticabi flow: terminating if used as a value")
		}
		return result, err
	case *ast.LoopExpression:
		result, _, err := lower.lowerLoop(value)
		return result, err
	case *ast.MatchExpression:
		result, produced, err := lower.lowerMatch(value)
		if !produced && err == nil {
			return 0, fmt.Errorf("semanticabi flow: terminating match used as a value")
		}
		return result, err
	case *ast.BlockExpression:
		result, produced, err := lower.lowerBlock(value)
		if err != nil {
			return 0, err
		}
		if !produced {
			return lower.emitVoid(value)
		}
		return result, nil
	default:
		return 0, fmt.Errorf("semanticabi flow: unsupported expression %T", expression)
	}
}

func (lower *lowerer) lowerIdentifier(identifier *ast.Identifier) (uint32, error) {
	if value, ok := lower.lookup(identifier.Name); ok {
		return value.register, nil
	}
	if receiverName, memberName, ok := splitQualified(identifier.Name); ok {
		if receiver, found := lower.lookup(receiverName); found {
			resultType := lower.dynamicType()
			result := lower.newRegister(resultType)
			if err := lower.emit(identifier, semanticabi.OpGetMemberValue, result, resultType, receiver.register, lower.tables.symbol(memberName)); err != nil {
				return 0, err
			}
			return result, nil
		}
	}
	typeName := lower.globalTypes[identifier.Name]
	typeID := lower.tables.typeIndex(typeName)
	result := lower.newRegister(typeID)
	if err := lower.emit(identifier, semanticabi.OpLoadGlobal, result, typeID, lower.tables.symbol(identifier.Name)); err != nil {
		return 0, err
	}
	return result, nil
}

func (lower *lowerer) lowerAssignment(assignment *ast.AssignmentExpression) (uint32, error) {
	right, err := lower.lowerExpression(assignment.Right)
	if err != nil {
		return 0, err
	}
	name, annotation, ok := assignmentTargetInfo(assignment.Left)
	if !ok {
		return 0, fmt.Errorf("semanticabi flow: unsupported assignment target %T", assignment.Left)
	}
	typeID := lower.registerType(right)
	if annotation != nil {
		typeID = lower.tables.typeIndex(typeName(annotation))
	}
	target, found := lower.lookup(name)
	if !found || assignment.Operator == ast.AssignmentDeclare && !lower.boundInCurrentScope(name) {
		target = binding{register: lower.newRegister(typeID), typeID: typeID}
		lower.bind(name, target)
	}
	if err := lower.emit(assignment, semanticabi.OpMoveValue, target.register, right); err != nil {
		return 0, err
	}
	return target.register, nil
}

func (lower *lowerer) boundInCurrentScope(name string) bool {
	_, ok := lower.scopes[len(lower.scopes)-1][name]
	return ok
}

func (lower *lowerer) lowerUnary(expression *ast.UnaryExpression) (uint32, error) {
	operand, err := lower.lowerExpression(expression.Operand)
	if err != nil {
		return 0, err
	}
	typeID := lower.registerType(operand)
	result := lower.newRegister(typeID)
	return result, lower.emit(expression, semanticabi.OpUnaryValue, result, typeID, lower.tables.symbol(string(expression.Operator)), operand)
}

func (lower *lowerer) lowerBinary(expression *ast.BinaryExpression) (uint32, error) {
	if expression.Operator == "&&" || expression.Operator == "||" {
		return 0, fmt.Errorf("semanticabi flow: short-circuit operator %q requires dedicated lowering", expression.Operator)
	}
	left, err := lower.lowerExpression(expression.Left)
	if err != nil {
		return 0, err
	}
	right, err := lower.lowerExpression(expression.Right)
	if err != nil {
		return 0, err
	}
	typeID := lower.joinTypes(lower.registerType(left), lower.registerType(right))
	switch expression.Operator {
	case "<", ">", "<=", ">=", "==", "!=":
		typeID = lower.boolType()
	}
	result := lower.newRegister(typeID)
	return result, lower.emit(expression, semanticabi.OpBinaryValue, result, typeID, lower.tables.symbol(expression.Operator), left, right)
}

func (lower *lowerer) lowerCast(expression *ast.TypeCastExpression) (uint32, error) {
	operand, err := lower.lowerExpression(expression.Expression)
	if err != nil {
		return 0, err
	}
	typeID := lower.tables.typeIndex(typeName(expression.TargetType))
	result := lower.newRegister(typeID)
	return result, lower.emit(expression, semanticabi.OpCastValue, result, typeID, operand)
}

func (lower *lowerer) lowerMember(expression *ast.MemberAccessExpression) (uint32, error) {
	receiver, err := lower.lowerExpression(expression.Object)
	if err != nil {
		return 0, err
	}
	member, err := requireIdentifier(expression.Member)
	if err != nil {
		return 0, err
	}
	typeID := lower.dynamicType()
	result := lower.newRegister(typeID)
	return result, lower.emit(expression, semanticabi.OpGetMemberValue, result, typeID, receiver, lower.tables.symbol(member.Name))
}

func (lower *lowerer) lowerCall(call *ast.FunctionCall) (uint32, error) {
	if capabilityName, returnName, ok := lower.hostCapability(call.Callee); ok {
		arguments := make([]uint32, 0, len(call.Arguments))
		for _, argument := range call.Arguments {
			value, err := lower.lowerExpression(argument)
			if err != nil {
				return 0, err
			}
			arguments = append(arguments, value)
		}
		return lower.lowerHostEffect(call, capabilityName, returnName, arguments)
	}
	targetKind, packageName, ownerType, targetName, returnType, arguments, err := lower.resolveCall(call)
	if err != nil {
		return 0, err
	}
	target := lower.tables.callTarget(targetKind, packageName, ownerType, targetName, uint32(len(arguments)), returnType)
	result := lower.newRegister(returnType)
	operands := []uint32{result, target, returnType}
	operands = append(operands, arguments...)
	return result, lower.emit(call, semanticabi.OpInvoke, operands...)
}

func (lower *lowerer) resolveCall(call *ast.FunctionCall) (semanticabi.CallTargetKind, string, uint32, string, uint32, []uint32, error) {
	name := expressionName(call.Callee)
	arguments := make([]uint32, 0, len(call.Arguments)+1)
	if qualifier, member, ok := splitQualified(name); ok {
		if receiver, found := lower.lookup(qualifier); found {
			arguments = append(arguments, receiver.register)
			for _, argument := range call.Arguments {
				value, err := lower.lowerExpression(argument)
				if err != nil {
					return 0, "", 0, "", 0, nil, err
				}
				arguments = append(arguments, value)
			}
			return semanticabi.CallTargetMember, "", receiver.typeID, member, lower.dynamicType(), arguments, nil
		}
		if packageName, imported := lower.imports[qualifier]; imported {
			for _, argument := range call.Arguments {
				value, err := lower.lowerExpression(argument)
				if err != nil {
					return 0, "", 0, "", 0, nil, err
				}
				arguments = append(arguments, value)
			}
			ownerType := lower.tables.typeIndex(qualifier)
			return semanticabi.CallTargetImported, packageName, ownerType, member, lower.dynamicType(), arguments, nil
		}
	}
	if memberCall, ok := call.Callee.(*ast.MemberAccessExpression); ok {
		receiver, err := lower.lowerExpression(memberCall.Object)
		if err != nil {
			return 0, "", 0, "", 0, nil, err
		}
		member, err := requireIdentifier(memberCall.Member)
		if err != nil {
			return 0, "", 0, "", 0, nil, err
		}
		arguments = append(arguments, receiver)
		for _, argument := range call.Arguments {
			value, err := lower.lowerExpression(argument)
			if err != nil {
				return 0, "", 0, "", 0, nil, err
			}
			arguments = append(arguments, value)
		}
		ownerType := lower.registerType(receiver)
		return semanticabi.CallTargetMember, "", ownerType, member.Name, lower.dynamicType(), arguments, nil
	}
	for _, argument := range call.Arguments {
		value, err := lower.lowerExpression(argument)
		if err != nil {
			return 0, "", 0, "", 0, nil, err
		}
		arguments = append(arguments, value)
	}
	if returnName, ok := lower.localReturns[name]; ok {
		return semanticabi.CallTargetLocal, lower.packageName, semanticabi.NoIndex, name, lower.tables.typeIndex(returnName), arguments, nil
	}
	return semanticabi.CallTargetBuiltin, "", semanticabi.NoIndex, name, lower.dynamicType(), arguments, nil
}

func (lower *lowerer) lowerHostEffect(call *ast.FunctionCall, name, returnName string, arguments []uint32) (uint32, error) {
	returnType := lower.tables.typeIndex(returnName)
	result := lower.newRegister(returnType)
	capability := lower.tables.capability(name)
	origin := lower.current
	continuation := lower.newBlock()
	operands := []uint32{result, capability, returnType, continuation}
	operands = append(operands, arguments...)
	if err := lower.emitTo(origin, call, semanticabi.OpHostEffectResume, operands...); err != nil {
		return 0, err
	}
	lower.setCurrent(continuation)
	lower.coverage.HostEffects = appendUnique(lower.coverage.HostEffects, name)
	return result, nil
}

func splitQualified(name string) (string, string, bool) {
	index := strings.LastIndexByte(name, '.')
	if index <= 0 || index == len(name)-1 {
		return "", "", false
	}
	return name[:index], name[index+1:], true
}
