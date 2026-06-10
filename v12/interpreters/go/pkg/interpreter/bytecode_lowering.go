package interpreter

import (
	"errors"
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

var errBytecodeUnsupported = errors.New("bytecode lowering unsupported")

func emitStatement(ctx *bytecodeLoweringContext, i *Interpreter, stmt ast.Statement, isLast bool) error {
	resultDiscarded := false
	switch s := stmt.(type) {
	case *ast.FunctionDefinition:
		if s == nil {
			return bytecodeUnsupported("nil function definition")
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpDefineFunction, node: s})
	case *ast.StructDefinition:
		if s == nil {
			return bytecodeUnsupported("nil struct definition")
		}
		if s.ID != nil && s.ID.Name != "" {
			if ctx.structDefs == nil {
				ctx.structDefs = make(map[string]*ast.StructDefinition, 4)
			}
			ctx.structDefs[s.ID.Name] = s
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpDefineStruct, node: s})
	case *ast.UnionDefinition:
		if s == nil {
			return bytecodeUnsupported("nil union definition")
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpDefineUnion, node: s})
	case *ast.TypeAliasDefinition:
		if s == nil {
			return bytecodeUnsupported("nil type alias definition")
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpDefineTypeAlias, node: s})
	case *ast.MethodsDefinition:
		if s == nil {
			return bytecodeUnsupported("nil methods definition")
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpDefineMethods, node: s})
	case *ast.InterfaceDefinition:
		if s == nil {
			return bytecodeUnsupported("nil interface definition")
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpDefineInterface, node: s})
	case *ast.ImplementationDefinition:
		if s == nil {
			return bytecodeUnsupported("nil implementation definition")
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpDefineImplementation, node: s})
	case *ast.ExternFunctionBody:
		if s == nil {
			return bytecodeUnsupported("nil extern function body")
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpDefineExtern, node: s})
	case *ast.ImportStatement:
		if s == nil {
			return bytecodeUnsupported("nil import statement")
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpImport, node: s})
	case *ast.DynImportStatement:
		if s == nil {
			return bytecodeUnsupported("nil dynimport statement")
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpDynImport, node: s})
	case *ast.ExportStatement, *ast.PackageStatement, *ast.PreludeStatement:
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
	case *ast.ReturnStatement:
		if s == nil {
			return bytecodeUnsupported("nil return statement")
		}
		if s.Argument != nil {
			if emitted, err := bytecodeEmitFinalI32StackExpr(ctx, s.Argument); err != nil {
				return err
			} else if !emitted {
				if err := emitExpression(ctx, i, s.Argument); err != nil {
					return err
				}
			}
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.VoidValue{}})
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpReturn, node: s})
	case *ast.YieldStatement:
		if s == nil {
			return bytecodeUnsupported("nil yield statement")
		}
		if s.Expression != nil {
			if err := emitExpression(ctx, i, s.Expression); err != nil {
				return err
			}
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpYield})
	case *ast.RaiseStatement:
		if s == nil {
			return bytecodeUnsupported("nil raise statement")
		}
		if s.Expression != nil {
			if err := emitExpression(ctx, i, s.Expression); err != nil {
				return err
			}
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpRaise, node: s})
	case *ast.RethrowStatement:
		if s == nil {
			return bytecodeUnsupported("nil rethrow statement")
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpRethrow, node: s})
	case *ast.ForLoop:
		if err := emitForLoop(ctx, i, s); err != nil {
			return err
		}
	case *ast.WhileLoop:
		if err := emitWhileLoop(ctx, i, s); err != nil {
			return err
		}
	case *ast.BreakStatement:
		if err := emitBreakStatement(ctx, i, s); err != nil {
			return err
		}
	case *ast.ContinueStatement:
		if err := emitContinueStatement(ctx, i, s); err != nil {
			return err
		}
	case ast.Expression:
		if !isLast {
			if ifExpr, ok := s.(*ast.IfExpression); ok {
				return emitIfStatement(ctx, i, ifExpr)
			}
		}
		if isLast {
			if emitted, err := bytecodeEmitFinalI32StackExpr(ctx, s); err != nil {
				return err
			} else if emitted {
				return nil
			}
		}
		start := len(ctx.instructions)
		prevDiscard := ctx.discardExpressionValue
		prevDiscardNode := ctx.discardExpressionNode
		ctx.discardExpressionValue = !isLast
		if !isLast {
			ctx.discardExpressionNode = s
		}
		if err := emitExpression(ctx, i, s); err != nil {
			ctx.discardExpressionValue = prevDiscard
			ctx.discardExpressionNode = prevDiscardNode
			return err
		}
		ctx.discardExpressionValue = prevDiscard
		ctx.discardExpressionNode = prevDiscardNode
		resultDiscarded = !isLast && len(ctx.instructions) > start && ctx.instructions[len(ctx.instructions)-1].discardResult
	default:
		return bytecodeUnsupported("statement %T", stmt)
	}
	if !isLast && !resultDiscarded {
		ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	}
	return nil
}

func emitExpression(ctx *bytecodeLoweringContext, i *Interpreter, expr ast.Expression) error {
	if expr == nil {
		return bytecodeUnsupported("nil expression")
	}
	if ctx.allowPlaceholderLambda {
		if plan, ok, err := placeholderPlanForExpression(expr); err != nil {
			return err
		} else if ok {
			ctx.emit(bytecodeInstruction{
				op:       bytecodeOpPlaceholderLambda,
				node:     expr,
				argCount: plan.paramCount,
			})
			return nil
		}
	}
	switch n := expr.(type) {
	case *ast.StringLiteral:
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.StringValue{Val: n.Value}})
		return nil
	case *ast.BooleanLiteral:
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.BoolValue{Val: n.Value}})
		return nil
	case *ast.CharLiteral:
		if len(n.Value) == 0 {
			return fmt.Errorf("empty char literal")
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.CharValue{Val: []rune(n.Value)[0]}})
		return nil
	case *ast.NilLiteral:
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
		return nil
	case *ast.IntegerLiteral:
		suffix := runtime.IntegerI32
		if n.IntegerType != nil {
			suffix = runtime.IntegerType(*n.IntegerType)
		}
		if n.Value != nil && n.Value.IsInt64() {
			ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NewSmallInt(n.Value.Int64(), suffix), node: n})
			return nil
		}
		val := bigFromLiteral(n.Value)
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NewBigIntValue(val, suffix), node: n})
		return nil
	case *ast.FloatLiteral:
		suffix := runtime.FloatF64
		if n.FloatType != nil {
			suffix = runtime.FloatType(*n.FloatType)
		}
		val := n.Value
		if suffix == runtime.FloatF32 {
			val = float64(float32(val))
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.FloatValue{Val: val, TypeSuffix: suffix}})
		return nil
	case *ast.Identifier:
		if slot, implicit, ok := ctx.lookupAnySlot(n.Name); ok && implicit {
			ctx.emit(bytecodeInstruction{op: bytecodeOpLoadImplicitSlot, target: slot, name: n.Name, nameSimple: bytecodeSimpleLookupName(n.Name), node: n})
		} else if ok {
			ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: slot, name: n.Name, node: n})
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpLoadName, name: n.Name, nameSimple: bytecodeSimpleLookupName(n.Name), node: n})
		}
		return nil
	case *ast.MemberAccessExpression:
		if instr, plan, ok := bytecodeLoadSlotStructFieldInstruction(ctx, n); ok {
			memberIP := ctx.emit(instr)
			bytecodeStoreNamedStructFieldMemberPlan(ctx, memberIP, plan.definition, plan.fieldIndex)
			return nil
		}
		if err := emitExpression(ctx, i, n.Object); err != nil {
			return err
		}
		memberIP := ctx.emit(bytecodeInstruction{
			op:            bytecodeOpMemberAccess,
			name:          bytecodeIdentifierMemberName(n.Member),
			node:          n,
			safe:          n.Safe,
			preferMethods: false,
		})
		bytecodeStoreNamedStructMemberPlan(ctx, memberIP, n)
		return nil
	case *ast.IndexExpression:
		if instr, ok := bytecodeArrayIndexGetSlotInstruction(ctx, n); ok {
			ctx.emit(instr)
			return nil
		}
		if err := emitExpression(ctx, i, n.Object); err != nil {
			return err
		}
		if err := emitExpression(ctx, i, n.Index); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpIndexGet, node: n})
		return nil
	case *ast.FunctionCall:
		if member, ok := n.Callee.(*ast.MemberAccessExpression); ok && member != nil {
			memberName := bytecodeIdentifierMemberName(member.Member)
			if instr, ok := bytecodeArrayReadSlotInstruction(ctx, n, member, memberName); ok {
				ctx.emit(instr)
				return nil
			}
			staticReceiver := false
			if instr, ok := bytecodeStaticReceiverInstruction(ctx, member, memberName, len(n.Arguments)); ok {
				ctx.emit(instr)
				staticReceiver = true
			} else if err := emitExpression(ctx, i, member.Object); err != nil {
				return err
			}
			if member.Safe {
				ctx.emit(bytecodeInstruction{op: bytecodeOpDup})
				jumpToNil := ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfNil, target: -1})
				for _, arg := range n.Arguments {
					if err := emitExpression(ctx, i, arg); err != nil {
						return err
					}
				}
				if memberName != "" {
					op := bytecodeOpCallMember
					if i.bytecodeGenericUnionMethodCallProven(n) {
						op = bytecodeOpCallGenericUnionMember
					}
					ctx.emit(bytecodeInstruction{op: op, name: memberName, argCount: len(n.Arguments), node: n, safe: true})
				} else {
					ctx.emit(bytecodeInstruction{
						op:            bytecodeOpMemberAccess,
						name:          memberName,
						node:          member,
						preferMethods: true,
					})
					ctx.emit(bytecodeInstruction{op: bytecodeOpCall, argCount: len(n.Arguments), node: n})
				}
				jumpToEnd := ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1})
				ctx.patchJump(jumpToNil, len(ctx.instructions))
				ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
				ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
				ctx.patchJump(jumpToEnd, len(ctx.instructions))
				return nil
			}
			for _, arg := range n.Arguments {
				if err := emitExpression(ctx, i, arg); err != nil {
					return err
				}
			}
			if memberName != "" {
				if staticReceiver {
					ctx.emit(bytecodeStaticMemberCallInstruction(memberName, len(n.Arguments), n))
				} else if i.bytecodeGenericUnionMethodCallProven(n) {
					ctx.emit(bytecodeInstruction{op: bytecodeOpCallGenericUnionMember, name: memberName, argCount: len(n.Arguments), node: n})
				} else {
					ctx.emit(bytecodeCallMemberInstructionForName(memberName, len(n.Arguments), n))
				}
			} else {
				ctx.emit(bytecodeInstruction{
					op:            bytecodeOpMemberAccess,
					name:          memberName,
					node:          member,
					preferMethods: true,
				})
				ctx.emit(bytecodeInstruction{op: bytecodeOpCall, argCount: len(n.Arguments), node: n})
			}
			return nil
		}
		if ident, ok := n.Callee.(*ast.Identifier); ok && ident != nil {
			if slot, memberName, ok := bytecodeDottedSlotMemberCall(ctx, ident); ok {
				ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: slot, name: ident.Name, node: ident})
				for _, arg := range n.Arguments {
					if err := emitExpression(ctx, i, arg); err != nil {
						return err
					}
				}
				ctx.emit(bytecodeCallMemberInstructionForName(memberName, len(n.Arguments), n))
				return nil
			}
			if slot, implicit, found := ctx.lookupAnySlot(ident.Name); found {
				op := bytecodeOpLoadSlot
				if implicit {
					op = bytecodeOpLoadImplicitSlot
				}
				ctx.emit(bytecodeInstruction{op: op, target: slot, name: ident.Name, nameSimple: bytecodeSimpleLookupName(ident.Name), node: ident})
				for _, arg := range n.Arguments {
					if err := emitExpression(ctx, i, arg); err != nil {
						return err
					}
				}
				ctx.emit(bytecodeInstruction{op: bytecodeOpCall, argCount: len(n.Arguments), node: n})
				return nil
			}
			if ctx.selfCallSlot >= 0 && ctx.selfCallName != "" && ident.Name == ctx.selfCallName {
				if instr, ok := bytecodeSelfCallSlotConstInstruction(ctx, n); ok {
					ctx.emit(instr)
					return nil
				}
				for _, arg := range n.Arguments {
					if err := emitExpression(ctx, i, arg); err != nil {
						return err
					}
				}
				ctx.emit(bytecodeInstruction{op: bytecodeOpCallSelf, target: ctx.selfCallSlot, argCount: len(n.Arguments), node: n})
				return nil
			}
			if instr, ok := bytecodeCallNameSlotArgsInstruction(ctx, ident.Name, n); ok {
				ctx.emit(instr)
				return nil
			}
			for _, arg := range n.Arguments {
				if err := emitExpression(ctx, i, arg); err != nil {
					return err
				}
			}
			ctx.emit(bytecodeInstruction{
				op:         bytecodeOpCallName,
				name:       ident.Name,
				nameSimple: bytecodeSimpleLookupName(ident.Name),
				argCount:   len(n.Arguments),
				node:       n,
			})
			return nil
		}
		if err := emitExpression(ctx, i, n.Callee); err != nil {
			return err
		}
		for _, arg := range n.Arguments {
			if err := emitExpression(ctx, i, arg); err != nil {
				return err
			}
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpCall, argCount: len(n.Arguments), node: n})
		return nil
	case *ast.LambdaExpression:
		ctx.emit(bytecodeInstruction{op: bytecodeOpMakeFunction, node: n})
		return nil
	case *ast.StructLiteral:
		if emitted, err := emitSimpleNamedStructLiteral(ctx, i, n); err != nil {
			return err
		} else if emitted {
			return nil
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpStructLiteral, node: n})
		return nil
	case *ast.MapLiteral:
		ctx.emit(bytecodeInstruction{op: bytecodeOpMapLiteral, node: n})
		return nil
	case *ast.ArrayLiteral:
		for _, el := range n.Elements {
			if err := emitExpression(ctx, i, el); err != nil {
				return err
			}
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpArrayLiteral, argCount: len(n.Elements), node: n})
		return nil
	case *ast.StringInterpolation:
		for _, part := range n.Parts {
			if err := emitExpression(ctx, i, part); err != nil {
				return err
			}
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpStringInterpolation, argCount: len(n.Parts), node: n})
		return nil
	case *ast.TypeCastExpression:
		if lowered, err := bytecodeEmitIntegerDivCast(ctx, i, n); err != nil {
			return err
		} else if lowered {
			return nil
		}
		if err := emitExpression(ctx, i, n.Expression); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpCast, node: n})
		return nil
	case *ast.RangeExpression:
		if err := emitExpression(ctx, i, n.Start); err != nil {
			return err
		}
		if err := emitExpression(ctx, i, n.End); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpRange, node: n})
		return nil
	case *ast.BinaryExpression:
		switch n.Operator {
		case "&&":
			if err := emitExpression(ctx, i, n.Left); err != nil {
				return err
			}
			ctx.emit(bytecodeInstruction{op: bytecodeOpDup})
			jumpToEnd := ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})
			ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
			if err := emitExpression(ctx, i, n.Right); err != nil {
				return err
			}
			ctx.patchJump(jumpToEnd, len(ctx.instructions))
			return nil
		case "||":
			if err := emitExpression(ctx, i, n.Left); err != nil {
				return err
			}
			ctx.emit(bytecodeInstruction{op: bytecodeOpDup})
			jumpToRight := ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})
			jumpToEnd := ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1})
			ctx.patchJump(jumpToRight, len(ctx.instructions))
			ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
			if err := emitExpression(ctx, i, n.Right); err != nil {
				return err
			}
			ctx.patchJump(jumpToEnd, len(ctx.instructions))
			return nil
		case "|>", "|>>":
			if err := emitExpression(ctx, i, n.Left); err != nil {
				return err
			}
			ctx.emit(bytecodeInstruction{op: bytecodeOpPipe, node: n.Right})
			return nil
		default:
			if instr, ok := bytecodeBinaryCastSlotFloatConstInstruction(ctx, n); ok {
				ctx.emit(instr)
				return nil
			}
			if instr, ok := bytecodeBinaryFloatMulSlotConstInstruction(ctx, n); ok {
				ctx.emit(instr)
				return nil
			}
			if instr, ok := bytecodeBinarySlotConstInstruction(ctx, n); ok {
				ctx.emit(instr)
				return nil
			}
			if err := emitExpression(ctx, i, n.Left); err != nil {
				return err
			}
			if err := emitExpression(ctx, i, n.Right); err != nil {
				return err
			}
			op := bytecodeBinaryOpcodeForOperator(n.Operator)
			ctx.emit(bytecodeInstruction{
				op:                  op,
				operator:            n.Operator,
				bitwiseRawCandidate: bytecodeDottedBitwiseOperator(n.Operator),
				node:                n,
			})
			return nil
		}
	case *ast.UnaryExpression:
		if err := emitExpression(ctx, i, n.Operand); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpUnary, operator: string(n.Operator), node: n})
		return nil
	case *ast.AssignmentExpression:
		if idxExpr, ok := n.Left.(*ast.IndexExpression); ok {
			if err := emitExpression(ctx, i, n.Right); err != nil {
				return err
			}
			if instr, ok := bytecodeArrayIndexSetSlotInstruction(ctx, n, idxExpr); ok {
				ctx.emit(instr)
				return nil
			}
			if err := emitExpression(ctx, i, idxExpr.Object); err != nil {
				return err
			}
			if err := emitExpression(ctx, i, idxExpr.Index); err != nil {
				return err
			}
			ctx.emit(bytecodeInstruction{op: bytecodeOpIndexSet, operator: string(n.Operator), node: n})
			return nil
		}
		if memberExpr, ok := n.Left.(*ast.MemberAccessExpression); ok {
			if err := emitExpression(ctx, i, n.Right); err != nil {
				return err
			}
			if err := emitExpression(ctx, i, memberExpr.Object); err != nil {
				return err
			}
			memberIP := ctx.emit(bytecodeInstruction{
				op:       bytecodeOpMemberSet,
				name:     bytecodeIdentifierMemberName(memberExpr.Member),
				operator: string(n.Operator),
				node:     memberExpr,
				safe:     memberExpr.Safe,
			})
			bytecodeStoreNamedStructMemberPlan(ctx, memberIP, memberExpr)
			return nil
		}
		if implicitExpr, ok := n.Left.(*ast.ImplicitMemberExpression); ok {
			if err := emitExpression(ctx, i, n.Right); err != nil {
				return err
			}
			ctx.emit(bytecodeInstruction{op: bytecodeOpImplicitMemberSet, operator: string(n.Operator), node: implicitExpr})
			return nil
		}
		name, ok := resolveAssignmentTargetName(n.Left)
		_, typedSimple := n.Left.(*ast.TypedPattern)
		useTypedSlotTarget := typedSimple && ok && ctx.frameLayout != nil && (n.Operator == ast.AssignmentDeclare || n.Operator == ast.AssignmentAssign)
		resultSimpleCheck := bytecodeExpressionSimpleTypeCheck(ctx, n.Right)
		if pattern, ok := n.Left.(ast.Pattern); ok && pattern != nil {
			if _, simple := resolveAssignmentTargetName(n.Left); !simple || (typedSimple && !useTypedSlotTarget) {
				if err := emitExpression(ctx, i, n.Right); err != nil {
					return err
				}
				ctx.emit(bytecodeInstruction{op: bytecodeOpAssignPattern, operator: string(n.Operator), node: n})
				return nil
			}
		}
		if ok && n.Operator != ast.AssignmentDeclare && n.Operator != ast.AssignmentAssign {
			if _, isCompound := binaryOpForAssignment(n.Operator); isCompound {
				if ctx.frameLayout != nil && bytecodeCanEmitRawI32CompoundAssign(n.Operator) && bytecodeCanEmitRawI32StackExprWithSlots(ctx, n.Right) {
					if slot, found := ctx.lookupSlot(name); found && ctx.slotKind(slot) == bytecodeCellKindI32 {
						bytecodeEmitRawI32StackExpr(ctx, n.Right)
						ctx.setSlotSimpleCheck(slot, bytecodeSimpleTypeCheckI32)
						ctx.emit(bytecodeInstruction{op: bytecodeOpCompoundAssignSlotI32, target: slot, name: name, operator: string(n.Operator), node: n, discardResult: ctx.discardExpressionValue && ctx.discardExpressionNode == n})
						return nil
					}
				}
				if err := emitExpression(ctx, i, n.Right); err != nil {
					return err
				}
				if slot, implicit, found := ctx.lookupAnySlot(name); found && implicit {
					ctx.setSlotSimpleCheck(slot, bytecodeSimpleTypeCheckUnknown)
					ctx.emit(bytecodeInstruction{op: bytecodeOpCompoundAssignImplicitSlot, target: slot, name: name, operator: string(n.Operator), node: n})
				} else if found {
					ctx.setSlotSimpleCheck(slot, bytecodeSimpleTypeCheckUnknown)
					ctx.emit(bytecodeInstruction{op: bytecodeOpCompoundAssignSlot, target: slot, name: name, operator: string(n.Operator), node: n})
				} else {
					ctx.emit(bytecodeInstruction{op: bytecodeOpAssignNameCompound, name: name, operator: string(n.Operator), node: n})
				}
				return nil
			}
		}
		if n.Operator != ast.AssignmentDeclare && n.Operator != ast.AssignmentAssign || !ok {
			return bytecodeUnsupported("assignment expression operator %q target %T", n.Operator, n.Left)
		}
		typedPattern, hasTypedStore := typedIdentifierPatternFromTarget(n.Left)
		if handled, err := emitGuardedImplicitSlotAssignment(ctx, i, n, name, typedPattern, hasTypedStore); handled || err != nil {
			return err
		}
		if ctx.frameLayout != nil && ok {
			if !hasTypedStore {
				if plan, ok := bytecodeStoreSlotFloatAffineInstruction(ctx, n.Right, n); ok {
					plan.instr.discardResult = ctx.discardExpressionValue && ctx.discardExpressionNode == n
					switch n.Operator {
					case ast.AssignmentDeclare:
						slot := ctx.declareSlotWithKind(name, bytecodeCellKindValue)
						ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
						plan.instr.target = slot
						plan.instr.name = name
						ip := ctx.emit(plan.instr)
						ctx.setFloatAffineStorePlan(ip, plan.plan)
						return nil
					case ast.AssignmentAssign:
						if slot, found := ctx.lookupSlot(name); found {
							ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
							plan.instr.target = slot
							plan.instr.name = name
							ip := ctx.emit(plan.instr)
							ctx.setFloatAffineStorePlan(ip, plan.plan)
							return nil
						}
					}
				}
				if instr, ok := bytecodeStoreSlotCastSlotFloatConstDivInstruction(ctx, n.Right, n); ok {
					instr.discardResult = ctx.discardExpressionValue && ctx.discardExpressionNode == n
					if n.Operator == ast.AssignmentDeclare {
						slot := ctx.declareSlotWithKind(name, bytecodeCellKindValue)
						ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
						instr.target = slot
						instr.name = name
						ctx.emit(instr)
						return nil
					}
					if n.Operator == ast.AssignmentAssign {
						if slot, found := ctx.lookupSlot(name); found {
							ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
							instr.target = slot
							instr.name = name
							ctx.emit(instr)
							return nil
						}
					}
				}
				if _, simpleIdent := n.Left.(*ast.Identifier); simpleIdent {
					if instr, ok := bytecodeStoreSlotFloatBinaryInstruction(ctx, n.Right, n); ok {
						instr.discardResult = ctx.discardExpressionValue && ctx.discardExpressionNode == n
						if n.Operator == ast.AssignmentDeclare {
							slot := ctx.declareSlotWithKind(name, bytecodeCellKindValue)
							ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
							instr.target = slot
							instr.name = name
							ctx.emit(instr)
							return nil
						}
						if n.Operator == ast.AssignmentAssign {
							if slot, found := ctx.lookupSlot(name); found {
								ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
								instr.target = slot
								instr.name = name
								ctx.emit(instr)
								return nil
							}
						}
					}
					if bytecodeTryEmitDeclaredFloatRegion(ctx, name, n, resultSimpleCheck) {
						return nil
					}
				}
			}
			if n.Operator == ast.AssignmentAssign {
				if _, simpleIdent := n.Left.(*ast.Identifier); simpleIdent {
					if plan, ok := bytecodeStoreSlotFloatAddMulArrayGetPlan(ctx, name, n.Right, n); ok {
						ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: plan.leftReceiverSlot, name: plan.leftReceiverName, node: n.Right})
						ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: plan.leftIndexSlot, name: plan.leftIndexName, node: n.Right})
						ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: plan.rightReceiverSlot, name: plan.rightReceiverName, node: n.Right})
						ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: plan.rightIndexSlot, name: plan.rightIndexName, node: n.Right})
						plan.instr.discardResult = ctx.discardExpressionValue && ctx.discardExpressionNode == n
						if slot, found := ctx.lookupSlot(name); found {
							ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
						}
						ctx.emit(plan.instr)
						return nil
					}
					if plan, ok := bytecodeStoreSlotFloatAddMulSlotPlan(ctx, name, n.Right, n); ok {
						if err := emitExpression(ctx, i, plan.stackExpr); err != nil {
							return err
						}
						plan.instr.discardResult = ctx.discardExpressionValue && ctx.discardExpressionNode == n
						if slot, found := ctx.lookupSlot(name); found {
							ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
						}
						ctx.emit(plan.instr)
						return nil
					}
					if plan, ok := bytecodeStoreSlotFloatAddMulPlan(ctx, name, n.Right, n); ok {
						ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: plan.baseSlot, name: plan.baseName, node: n.Right})
						if err := emitExpression(ctx, i, plan.mulLeft); err != nil {
							return err
						}
						if err := emitExpression(ctx, i, plan.mulRight); err != nil {
							return err
						}
						plan.instr.discardResult = ctx.discardExpressionValue && ctx.discardExpressionNode == n
						if slot, found := ctx.lookupSlot(name); found {
							ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
						}
						ctx.emit(plan.instr)
						return nil
					}
					if plan, ok := bytecodeStoreSlotFloatAddSubPlan(ctx, name, n.Right, n); ok {
						ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: plan.baseSlot, name: plan.baseName, node: n.Right})
						ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: plan.subLeftSlot, name: plan.subLeftName, node: n.Right})
						ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: plan.subRightSlot, name: plan.subRightName, node: n.Right})
						plan.instr.discardResult = ctx.discardExpressionValue && ctx.discardExpressionNode == n
						if slot, found := ctx.lookupSlot(name); found {
							ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
						}
						ctx.emit(plan.instr)
						return nil
					}
					if bytecodeTryEmitAssignedFloatRegion(ctx, name, n, resultSimpleCheck) {
						return nil
					}
					if plan, ok := bytecodeStoreSlotIntMulConstAddPlan(ctx, name, n.Right, n); ok {
						if plan.loadBase {
							ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: plan.targetSlot, name: name, node: n.Right})
						}
						if err := emitExpression(ctx, i, plan.addend); err != nil {
							return err
						}
						plan.instr.discardResult = ctx.discardExpressionValue && ctx.discardExpressionNode == n
						if slot, found := ctx.lookupSlot(name); found {
							ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
						}
						ctx.emit(plan.instr)
						return nil
					}
					if instr, ok := bytecodeStoreSlotIntMulConstModConstInstruction(ctx, name, n.Right, n); ok {
						instr.discardResult = ctx.discardExpressionValue && ctx.discardExpressionNode == n
						if slot, found := ctx.lookupSlot(name); found {
							ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
						}
						ctx.emit(instr)
						return nil
					}
					if instr, ok := bytecodeStoreSlotBinarySlotConstInstruction(ctx, name, n.Right, n); ok {
						instr.discardResult = ctx.discardExpressionValue && ctx.discardExpressionNode == n
						if slot, found := ctx.lookupSlot(name); found {
							ctx.setSlotSimpleCheck(slot, resultSimpleCheck)
						}
						ctx.emit(instr)
						return nil
					}
				}
			}
			if n.Operator == ast.AssignmentDeclare && hasTypedStore && bytecodeCellKindForTypeExpr(typedPattern.TypeAnnotation) == bytecodeCellKindI32 {
				if instr, ok := bytecodeArrayReadSlotI32Instruction(ctx, n.Right); ok {
					ctx.emit(instr)
				} else if bytecodeCanEmitRawI32StackExprWithSlots(ctx, n.Right) {
					bytecodeEmitRawI32StackExpr(ctx, n.Right)
				} else {
					if err := emitExpression(ctx, i, n.Right); err != nil {
						return err
					}
					ctx.emit(bytecodeInstruction{op: bytecodeOpUnboxI32, node: n})
				}
				slot := ctx.declareSlotWithKind(name, bytecodeCellKindI32)
				ctx.setSlotSimpleCheck(slot, bytecodeSimpleTypeCheckForName(cachedSimpleTypeName(typedPattern.TypeAnnotation)))
				ctx.setSlotExactStructDef(slot, bytecodeNominalNamedStructDefinitionForTypeExpr(ctx, typedPattern.TypeAnnotation))
				ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlotI32, target: slot, name: name, node: n, discardResult: ctx.discardExpressionValue && ctx.discardExpressionNode == n})
				return nil
			}
			if n.Operator == ast.AssignmentAssign {
				if slot, found := ctx.lookupSlot(name); found && ctx.slotKind(slot) == bytecodeCellKindI32 {
					if instr, ok := bytecodeArrayReadSlotI32Instruction(ctx, n.Right); ok {
						ctx.emit(instr)
					} else if bytecodeCanEmitRawI32StackExprWithSlots(ctx, n.Right) {
						bytecodeEmitRawI32StackExpr(ctx, n.Right)
					} else {
						if err := emitExpression(ctx, i, n.Right); err != nil {
							return err
						}
						ctx.emit(bytecodeInstruction{op: bytecodeOpUnboxI32, node: n})
					}
					ctx.setSlotSimpleCheck(slot, bytecodeSimpleTypeCheckI32)
					ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlotI32, target: slot, name: name, node: n, discardResult: ctx.discardExpressionValue && ctx.discardExpressionNode == n})
					return nil
				}
			}
		}
		if err := emitExpression(ctx, i, n.Right); err != nil {
			return err
		}
		if ctx.frameLayout != nil && ok {
			if n.Operator == ast.AssignmentDeclare {
				slotKind := bytecodeCellKindValue
				if hasTypedStore {
					slotKind = bytecodeCellKindForTypeExpr(typedPattern.TypeAnnotation)
				}
				slot := ctx.declareSlotWithKind(name, slotKind)
				if hasTypedStore {
					ctx.setSlotSimpleCheck(slot, bytecodeSimpleTypeCheckForName(cachedSimpleTypeName(typedPattern.TypeAnnotation)))
				} else {
					ctx.setSlotSimpleCheck(slot, bytecodeExpressionSimpleTypeCheck(ctx, n.Right))
				}
				if hasTypedStore {
					ctx.setSlotExactStructDef(slot, bytecodeNominalNamedStructDefinitionForTypeExpr(ctx, typedPattern.TypeAnnotation))
				} else {
					ctx.setSlotExactStructDef(slot, bytecodeNominalNamedStructDefinitionForExpr(ctx, n.Right))
				}
				instr := bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: slot, name: name, node: n}
				if hasTypedStore {
					instr.storeTyped = true
					instr.typeExpr = typedPattern.TypeAnnotation
					instr.discardResult = ctx.discardExpressionValue &&
						ctx.discardExpressionNode == n &&
						bytecodeCellKindForTypeExpr(typedPattern.TypeAnnotation) == bytecodeCellKindI32
				}
				ctx.emit(instr)
			} else if slot, found := ctx.lookupSlot(name); found {
				if hasTypedStore {
					ctx.setSlotSimpleCheck(slot, bytecodeSimpleTypeCheckForName(cachedSimpleTypeName(typedPattern.TypeAnnotation)))
				} else {
					ctx.setSlotSimpleCheck(slot, bytecodeExpressionSimpleTypeCheck(ctx, n.Right))
				}
				instr := bytecodeInstruction{op: bytecodeOpStoreSlot, target: slot, name: name, node: n}
				if hasTypedStore {
					instr.storeTyped = true
					instr.typeExpr = typedPattern.TypeAnnotation
					instr.discardResult = ctx.discardExpressionValue &&
						ctx.discardExpressionNode == n &&
						bytecodeCellKindForTypeExpr(typedPattern.TypeAnnotation) == bytecodeCellKindI32
				}
				ctx.emit(instr)
			} else if emitGuardedImplicitSlotDeclaration(ctx, n, name, typedPattern, hasTypedStore) {
			} else {
				if hasTypedStore {
					ctx.emit(bytecodeInstruction{op: bytecodeOpAssignPattern, operator: string(n.Operator), node: n})
				} else {
					ctx.emit(bytecodeInstruction{op: bytecodeOpAssignName, name: name, node: n})
				}
			}
		} else {
			op := bytecodeOpAssignName
			if n.Operator == ast.AssignmentDeclare {
				op = bytecodeOpDeclareName
			}
			ctx.emit(bytecodeInstruction{op: op, name: name, node: n})
		}
		return nil
	case *ast.BlockExpression:
		return emitBlock(ctx, i, n)
	case *ast.IfExpression:
		return emitIf(ctx, i, n)
	case *ast.MatchExpression:
		if emitted, err := emitSlotMatch(ctx, i, n); err != nil {
			return err
		} else if emitted {
			return nil
		}
		if emitted, err := emitEnvMatch(ctx, i, n); err != nil {
			return err
		} else if emitted {
			return nil
		}
		if err := emitExpression(ctx, i, n.Subject); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpMatch, node: n})
		return nil
	case *ast.RescueExpression:
		ctx.emit(bytecodeInstruction{op: bytecodeOpRescue, node: n})
		return nil
	case *ast.EnsureExpression:
		ctx.emit(bytecodeInstruction{op: bytecodeOpEnsure, node: n})
		if n.EnsureBlock != nil {
			if err := emitExpression(ctx, i, n.EnsureBlock); err != nil {
				return err
			}
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpEnsureEnd, node: n})
		return nil
	case *ast.OrElseExpression:
		if n.Handler == nil {
			return bytecodeUnsupported("or-else missing handler")
		}
		bindingName := ""
		if n.ErrorBinding != nil {
			bindingName = n.ErrorBinding.Name
		}
		jumpToEnd := ctx.emit(bytecodeInstruction{op: bytecodeOpOrElse, node: n, name: bindingName, target: -1})
		if err := emitExpression(ctx, i, n.Handler); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpExitScope})
		ctx.patchJump(jumpToEnd, len(ctx.instructions))
		return nil
	case *ast.PropagationExpression:
		if err := emitExpression(ctx, i, n.Expression); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpPropagation, node: n})
		return nil
	case *ast.AwaitExpression:
		if err := emitExpression(ctx, i, n.Expression); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpAwait, node: n})
		return nil
	case *ast.SpawnExpression:
		var program *bytecodeProgram
		if n != nil && n.Expression != nil {
			lowered, err := i.lowerExpressionToBytecode(n.Expression)
			if err != nil {
				return err
			}
			program = lowered
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpSpawn, node: n, program: program})
		return nil
	case *ast.ImplicitMemberExpression:
		ctx.emit(bytecodeInstruction{op: bytecodeOpImplicitMember, node: n})
		return nil
	case *ast.IteratorLiteral:
		var program *bytecodeProgram
		if n != nil {
			lowered, _, err := i.lowerIteratorLiteralBodyToBytecode(n, ctx.definitionEnv)
			if err != nil {
				return err
			}
			program = lowered
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpIteratorLiteral, node: n, program: program})
		return nil
	case *ast.BreakpointExpression:
		ctx.emit(bytecodeInstruction{op: bytecodeOpBreakpoint, node: n})
		return nil
	case *ast.PlaceholderExpression:
		ctx.emit(bytecodeInstruction{op: bytecodeOpPlaceholderValue, node: n})
		return nil
	case *ast.LoopExpression:
		return emitLoopExpression(ctx, i, n)
	default:
		return bytecodeUnsupported("expression %T", expr)
	}
}

func (ctx *bytecodeLoweringContext) emit(instr bytecodeInstruction) int {
	ctx.instructions = append(ctx.instructions, instr)
	if ctx.collectScalarProofs {
		check := bytecodeSimpleTypeCheckUnknown
		if bytecodeScalarProofUsesSlotCheck(instr.op) {
			check = ctx.slotSimpleCheck(instr.target)
		}
		ctx.scalarProofChecks = append(ctx.scalarProofChecks, check)
	}
	return len(ctx.instructions) - 1
}

func (ctx *bytecodeLoweringContext) patchJump(index int, target int) {
	if index < 0 || index >= len(ctx.instructions) {
		return
	}
	ctx.instructions[index].target = target
}

func (ctx *bytecodeLoweringContext) patchLoopTargets(index int, breakTarget int, continueTarget int) {
	if index < 0 || index >= len(ctx.instructions) {
		return
	}
	ctx.instructions[index].loopBreak = breakTarget
	ctx.instructions[index].loopContinue = continueTarget
}
