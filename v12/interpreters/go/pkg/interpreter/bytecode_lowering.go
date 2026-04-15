package interpreter

import (
	"errors"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

var errBytecodeUnsupported = errors.New("bytecode lowering unsupported")

func bytecodeSimpleLookupName(name string) bool {
	return name != "" && !strings.Contains(name, ".")
}

func bytecodeIdentifierMemberName(expr ast.Expression) string {
	if ident, ok := expr.(*ast.Identifier); ok && ident != nil {
		return ident.Name
	}
	return ""
}

type bytecodeLoweringContext struct {
	instructions           []bytecodeInstruction
	scopeDepth             int
	loopStack              []loopContext
	allowPlaceholderLambda bool
	frameLayout            *bytecodeFrameLayout // non-nil = slot mode
	slotScopes             []map[string]int     // scope stack for slot lookups
	nextSlot               int                  // next available slot index
	selfCallName           string               // current function name for self-recursive call lowering
	selfCallSlot           int                  // reserved slot for self-recursive call fast path
}

type loopContext struct {
	start      int
	scopeDepth int
	breakJumps []int
}

func (i *Interpreter) lowerModuleToBytecode(module *ast.Module) (*bytecodeProgram, error) {
	if module == nil {
		return nil, fmt.Errorf("bytecode lowering module is nil")
	}
	ctx := &bytecodeLoweringContext{
		instructions:           make([]bytecodeInstruction, 0, len(module.Body)*2),
		allowPlaceholderLambda: true,
	}
	if len(module.Body) == 0 {
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
		ctx.emit(bytecodeInstruction{op: bytecodeOpReturn})
		return &bytecodeProgram{instructions: ctx.instructions}, nil
	}
	for idx, stmt := range module.Body {
		if stmt == nil {
			return nil, bytecodeUnsupported("nil statement in module body")
		}
		if err := emitStatement(ctx, i, stmt, idx == len(module.Body)-1); err != nil {
			return nil, err
		}
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpReturn})
	return &bytecodeProgram{instructions: ctx.instructions}, nil
}

func (i *Interpreter) lowerExpressionToBytecode(expr ast.Expression) (*bytecodeProgram, error) {
	return i.lowerExpressionToBytecodeWithOptions(expr, true)
}

func (i *Interpreter) lowerExpressionToBytecodeWithOptions(expr ast.Expression, allowPlaceholderLambda bool) (*bytecodeProgram, error) {
	if expr == nil {
		return nil, fmt.Errorf("bytecode lowering expression is nil")
	}
	ctx := &bytecodeLoweringContext{
		instructions:           make([]bytecodeInstruction, 0, 4),
		allowPlaceholderLambda: allowPlaceholderLambda,
	}
	if err := emitExpression(ctx, i, expr); err != nil {
		return nil, err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpReturn})
	return &bytecodeProgram{instructions: ctx.instructions}, nil
}

func (i *Interpreter) lowerBlockExpressionToBytecode(block *ast.BlockExpression, allowPlaceholderLambda bool) (*bytecodeProgram, error) {
	if block == nil {
		return nil, fmt.Errorf("bytecode lowering block is nil")
	}
	ctx := &bytecodeLoweringContext{
		instructions:           make([]bytecodeInstruction, 0, len(block.Body)*2),
		allowPlaceholderLambda: allowPlaceholderLambda,
	}
	if err := emitBlock(ctx, i, block); err != nil {
		return nil, err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpReturn})
	return &bytecodeProgram{instructions: ctx.instructions}, nil
}

func emitStatement(ctx *bytecodeLoweringContext, i *Interpreter, stmt ast.Statement, isLast bool) error {
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
	case *ast.PackageStatement, *ast.PreludeStatement:
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
	case *ast.ReturnStatement:
		if s == nil {
			return bytecodeUnsupported("nil return statement")
		}
		if s.Argument != nil {
			if err := emitExpression(ctx, i, s.Argument); err != nil {
				return err
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
		if err := emitExpression(ctx, i, s); err != nil {
			return err
		}
	default:
		return bytecodeUnsupported("statement %T", stmt)
	}
	if !isLast {
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
		val := bigFromLiteral(n.Value)
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NewBigIntValue(val, suffix)})
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
		if slot, ok := ctx.lookupSlot(n.Name); ok {
			ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: slot, name: n.Name, node: n})
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpLoadName, name: n.Name, nameSimple: bytecodeSimpleLookupName(n.Name), node: n})
		}
		return nil
	case *ast.MemberAccessExpression:
		if err := emitExpression(ctx, i, n.Object); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{
			op:            bytecodeOpMemberAccess,
			name:          bytecodeIdentifierMemberName(n.Member),
			node:          n,
			safe:          n.Safe,
			preferMethods: false,
		})
		return nil
	case *ast.IndexExpression:
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
			if err := emitExpression(ctx, i, member.Object); err != nil {
				return err
			}
			if member.Safe {
				ctx.emit(bytecodeInstruction{op: bytecodeOpDup})
				jumpToNil := ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfNil, target: -1})
				ctx.emit(bytecodeInstruction{
					op:            bytecodeOpMemberAccess,
					name:          bytecodeIdentifierMemberName(member.Member),
					node:          member,
					preferMethods: true,
				})
				for _, arg := range n.Arguments {
					if err := emitExpression(ctx, i, arg); err != nil {
						return err
					}
				}
				ctx.emit(bytecodeInstruction{op: bytecodeOpCall, argCount: len(n.Arguments), node: n})
				jumpToEnd := ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1})
				ctx.patchJump(jumpToNil, len(ctx.instructions))
				ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
				ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
				ctx.patchJump(jumpToEnd, len(ctx.instructions))
				return nil
			}
			ctx.emit(bytecodeInstruction{
				op:            bytecodeOpMemberAccess,
				name:          bytecodeIdentifierMemberName(member.Member),
				node:          member,
				preferMethods: true,
			})
			for _, arg := range n.Arguments {
				if err := emitExpression(ctx, i, arg); err != nil {
					return err
				}
			}
			ctx.emit(bytecodeInstruction{op: bytecodeOpCall, argCount: len(n.Arguments), node: n})
			return nil
		}
		if ident, ok := n.Callee.(*ast.Identifier); ok && ident != nil {
			if slot, found := ctx.lookupSlot(ident.Name); found {
				ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: slot, name: ident.Name, node: ident})
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
			for _, arg := range n.Arguments {
				if err := emitExpression(ctx, i, arg); err != nil {
					return err
				}
			}
			ctx.emit(bytecodeInstruction{op: bytecodeOpCallName, name: ident.Name, nameSimple: bytecodeSimpleLookupName(ident.Name), argCount: len(n.Arguments), node: n})
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
			ctx.emit(bytecodeInstruction{op: op, operator: n.Operator, node: n})
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
			ctx.emit(bytecodeInstruction{op: bytecodeOpMemberSet, operator: string(n.Operator), node: memberExpr})
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
		useTypedSlotDeclare := typedSimple && ok && n.Operator == ast.AssignmentDeclare && ctx.frameLayout != nil
		if pattern, ok := n.Left.(ast.Pattern); ok && pattern != nil {
			if _, simple := resolveAssignmentTargetName(n.Left); !simple || (typedSimple && !useTypedSlotDeclare) {
				if err := emitExpression(ctx, i, n.Right); err != nil {
					return err
				}
				ctx.emit(bytecodeInstruction{op: bytecodeOpAssignPattern, operator: string(n.Operator), node: n})
				return nil
			}
		}
		if ok && n.Operator != ast.AssignmentDeclare && n.Operator != ast.AssignmentAssign {
			if _, isCompound := binaryOpForAssignment(n.Operator); isCompound {
				if err := emitExpression(ctx, i, n.Right); err != nil {
					return err
				}
				if slot, found := ctx.lookupSlot(name); found {
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
		if err := emitExpression(ctx, i, n.Right); err != nil {
			return err
		}
		if ctx.frameLayout != nil && ok {
			if n.Operator == ast.AssignmentDeclare {
				slot := ctx.declareSlot(name)
				ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: slot, name: name, node: n})
			} else if slot, found := ctx.lookupSlot(name); found {
				ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlot, target: slot, name: name, node: n})
			} else {
				ctx.emit(bytecodeInstruction{op: bytecodeOpAssignName, name: name, node: n})
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
		ctx.emit(bytecodeInstruction{op: bytecodeOpPropagation})
		return nil
	case *ast.AwaitExpression:
		if err := emitExpression(ctx, i, n.Expression); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpAwait, node: n})
		return nil
	case *ast.SpawnExpression:
		ctx.emit(bytecodeInstruction{op: bytecodeOpSpawn, node: n})
		return nil
	case *ast.ImplicitMemberExpression:
		ctx.emit(bytecodeInstruction{op: bytecodeOpImplicitMember, node: n})
		return nil
	case *ast.IteratorLiteral:
		var program *bytecodeProgram
		if n != nil {
			module := ast.NewModule(n.Body, nil, nil)
			lowered, err := i.lowerModuleToBytecode(module)
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

func placeholderPlanForExpression(expr ast.Expression) (placeholderPlan, bool, error) {
	if expr == nil {
		return placeholderPlan{}, false, nil
	}
	switch n := expr.(type) {
	case *ast.AssignmentExpression:
		return placeholderPlan{}, false, nil
	case *ast.BinaryExpression:
		if n.Operator == "|>" || n.Operator == "|>>" {
			return placeholderPlan{}, false, nil
		}
	}
	plan, ok, err := analyzePlaceholderExpression(expr)
	if err != nil || !ok {
		return plan, ok, err
	}
	if call, isCall := expr.(*ast.FunctionCall); isCall {
		calleeHas := expressionContainsPlaceholder(call.Callee)
		argsHave := false
		for _, arg := range call.Arguments {
			if expressionContainsPlaceholder(arg) {
				argsHave = true
				break
			}
		}
		if calleeHas && !argsHave {
			return placeholderPlan{}, false, nil
		}
	}
	return plan, true, nil
}

func emitBlock(ctx *bytecodeLoweringContext, i *Interpreter, block *ast.BlockExpression) error {
	if block == nil {
		return bytecodeUnsupported("nil block")
	}
	ctx.enterScope()
	if len(block.Body) == 0 {
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.VoidValue{}})
		ctx.exitScope()
		return nil
	}
	for idx, stmt := range block.Body {
		if stmt == nil {
			return bytecodeUnsupported("nil statement in block")
		}
		if err := emitStatement(ctx, i, stmt, idx == len(block.Body)-1); err != nil {
			return err
		}
	}
	ctx.exitScope()
	return nil
}

func emitIf(ctx *bytecodeLoweringContext, i *Interpreter, expr *ast.IfExpression) error {
	if expr == nil {
		return bytecodeUnsupported("nil if expression")
	}
	if err := emitExpression(ctx, i, expr.IfCondition); err != nil {
		return err
	}
	jumpToElse := ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})
	if err := emitBlock(ctx, i, expr.IfBody); err != nil {
		return err
	}
	jumpToEnd := []int{ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1})}
	ctx.patchJump(jumpToElse, len(ctx.instructions))

	for _, clause := range expr.ElseIfClauses {
		if clause == nil {
			return bytecodeUnsupported("nil elsif clause")
		}
		if err := emitExpression(ctx, i, clause.Condition); err != nil {
			return err
		}
		jumpToNext := ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})
		if err := emitBlock(ctx, i, clause.Body); err != nil {
			return err
		}
		jumpToEnd = append(jumpToEnd, ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1}))
		ctx.patchJump(jumpToNext, len(ctx.instructions))
	}

	if expr.ElseBody != nil {
		if err := emitBlock(ctx, i, expr.ElseBody); err != nil {
			return err
		}
	} else {
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
	}

	end := len(ctx.instructions)
	for _, idx := range jumpToEnd {
		ctx.patchJump(idx, end)
	}
	return nil
}

func emitLoopExpression(ctx *bytecodeLoweringContext, i *Interpreter, loop *ast.LoopExpression) error {
	if loop == nil {
		return bytecodeUnsupported("nil loop expression")
	}
	if loop.Body == nil {
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.VoidValue{}})
		return nil
	}
	loopEnter := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopEnter, loopBreak: -1, loopContinue: -1})
	loopStart := len(ctx.instructions)
	ctx.pushLoop(loopStart)
	if err := emitBlock(ctx, i, loop.Body); err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: loopStart})
	loopExit := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopExit})
	ctx.popLoop(loopExit)
	ctx.patchLoopTargets(loopEnter, loopExit, loopStart)
	return nil
}

func emitWhileLoop(ctx *bytecodeLoweringContext, i *Interpreter, loop *ast.WhileLoop) error {
	if loop == nil {
		return bytecodeUnsupported("nil while loop")
	}
	if loop.Condition == nil || loop.Body == nil {
		return bytecodeUnsupported("while loop missing condition/body")
	}
	loopEnter := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopEnter, loopBreak: -1, loopContinue: -1})
	loopStart := len(ctx.instructions)
	ctx.pushLoop(loopStart)
	if err := emitExpression(ctx, i, loop.Condition); err != nil {
		return err
	}
	jumpToNoBreak := ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})
	if err := emitBlock(ctx, i, loop.Body); err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: loopStart})
	noBreak := len(ctx.instructions)
	ctx.patchJump(jumpToNoBreak, noBreak)
	ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.VoidValue{}})
	loopExit := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopExit})
	ctx.popLoop(loopExit)
	ctx.patchLoopTargets(loopEnter, loopExit, loopStart)
	return nil
}

func emitForLoop(ctx *bytecodeLoweringContext, i *Interpreter, loop *ast.ForLoop) error {
	if loop == nil {
		return bytecodeUnsupported("nil for loop")
	}
	if loop.Iterable == nil || loop.Body == nil {
		return bytecodeUnsupported("for loop missing iterable/body")
	}
	if err := emitExpression(ctx, i, loop.Iterable); err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpIterInit})
	loopEnter := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopEnter, loopBreak: -1, loopContinue: -1})
	loopStart := len(ctx.instructions)
	ctx.pushLoop(loopStart)

	ctx.emit(bytecodeInstruction{op: bytecodeOpIterNext})
	jumpToBody := ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})

	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	ctx.emit(bytecodeInstruction{op: bytecodeOpIterClose})
	ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.VoidValue{}})
	jumpToEnd := ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1})

	bodyStart := len(ctx.instructions)
	ctx.patchJump(jumpToBody, bodyStart)
	ctx.enterScope()
	if ctx.frameLayout != nil {
		if ident, ok := loop.Pattern.(*ast.Identifier); ok {
			slot := ctx.declareSlot(ident.Name)
			ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: slot, name: ident.Name})
			ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpBindPattern, node: loop})
		}
	} else {
		ctx.emit(bytecodeInstruction{op: bytecodeOpBindPattern, node: loop})
	}
	if err := emitBlock(ctx, i, loop.Body); err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	ctx.exitScope()
	ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: loopStart})

	breakCleanup := len(ctx.instructions)
	ctx.emit(bytecodeInstruction{op: bytecodeOpIterClose})
	loopExit := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopExit})

	ctx.patchJump(jumpToEnd, loopExit)
	ctx.popLoopWithBreakTarget(loopExit, breakCleanup)
	ctx.patchLoopTargets(loopEnter, breakCleanup, loopStart)
	return nil
}

func emitBreakStatement(ctx *bytecodeLoweringContext, i *Interpreter, stmt *ast.BreakStatement) error {
	if stmt == nil {
		return bytecodeUnsupported("nil break statement")
	}
	if stmt.Label != nil {
		if stmt.Value != nil {
			if err := emitExpression(ctx, i, stmt.Value); err != nil {
				return err
			}
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpBreakLabel, name: stmt.Label.Name, node: stmt})
		return nil
	}
	if len(ctx.loopStack) == 0 {
		if stmt.Value != nil {
			if err := emitExpression(ctx, i, stmt.Value); err != nil {
				return err
			}
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpBreakSignal, node: stmt})
		return nil
	}
	if stmt.Value != nil {
		if err := emitExpression(ctx, i, stmt.Value); err != nil {
			return err
		}
	} else {
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
	}
	exitCount, err := ctx.loopExitCount()
	if err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpExitScope, argCount: exitCount})
	jumpIdx := ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1})
	ctx.appendBreakJump(jumpIdx)
	return nil
}

func emitContinueStatement(ctx *bytecodeLoweringContext, _ *Interpreter, stmt *ast.ContinueStatement) error {
	if stmt == nil {
		return bytecodeUnsupported("nil continue statement")
	}
	if stmt.Label != nil {
		return bytecodeUnsupported("labeled continue not supported")
	}
	if len(ctx.loopStack) == 0 {
		ctx.emit(bytecodeInstruction{op: bytecodeOpContinueSignal, node: stmt})
		return nil
	}
	exitCount, err := ctx.loopExitCount()
	if err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpExitScope, argCount: exitCount})
	ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: ctx.currentLoopStart()})
	return nil
}

func (ctx *bytecodeLoweringContext) emit(instr bytecodeInstruction) int {
	ctx.instructions = append(ctx.instructions, instr)
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

func (ctx *bytecodeLoweringContext) enterScope() {
	if ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes {
		ctx.emit(bytecodeInstruction{op: bytecodeOpEnterScope})
	}
	ctx.scopeDepth++
	if ctx.frameLayout != nil {
		ctx.slotScopes = append(ctx.slotScopes, make(map[string]int))
	}
}

func (ctx *bytecodeLoweringContext) exitScope() {
	if ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes {
		ctx.emit(bytecodeInstruction{op: bytecodeOpExitScope})
	}
	if ctx.scopeDepth > 0 {
		ctx.scopeDepth--
	}
	if ctx.frameLayout != nil && len(ctx.slotScopes) > 1 {
		ctx.slotScopes = ctx.slotScopes[:len(ctx.slotScopes)-1]
	}
}
func (ctx *bytecodeLoweringContext) lookupSlot(name string) (int, bool) {
	for i := len(ctx.slotScopes) - 1; i >= 0; i-- {
		if slot, ok := ctx.slotScopes[i][name]; ok {
			return slot, true
		}
	}
	return 0, false
}
func (ctx *bytecodeLoweringContext) declareSlot(name string) int {
	slot := ctx.nextSlot
	ctx.nextSlot++
	if len(ctx.slotScopes) > 0 {
		ctx.slotScopes[len(ctx.slotScopes)-1][name] = slot
	}
	return slot
}

func (ctx *bytecodeLoweringContext) pushLoop(start int) {
	ctx.loopStack = append(ctx.loopStack, loopContext{
		start:      start,
		scopeDepth: ctx.scopeDepth,
	})
}

func (ctx *bytecodeLoweringContext) popLoop(loopEnd int) {
	ctx.popLoopWithBreakTarget(loopEnd, loopEnd)
}

func (ctx *bytecodeLoweringContext) popLoopWithBreakTarget(loopEnd int, breakTarget int) {
	if len(ctx.loopStack) == 0 {
		return
	}
	loop := ctx.loopStack[len(ctx.loopStack)-1]
	ctx.loopStack = ctx.loopStack[:len(ctx.loopStack)-1]
	for _, idx := range loop.breakJumps {
		ctx.patchJump(idx, breakTarget)
	}
}

func (ctx *bytecodeLoweringContext) appendBreakJump(index int) {
	if len(ctx.loopStack) == 0 {
		return
	}
	last := len(ctx.loopStack) - 1
	ctx.loopStack[last].breakJumps = append(ctx.loopStack[last].breakJumps, index)
}

func (ctx *bytecodeLoweringContext) currentLoopStart() int {
	if len(ctx.loopStack) == 0 {
		return -1
	}
	return ctx.loopStack[len(ctx.loopStack)-1].start
}

func (ctx *bytecodeLoweringContext) loopExitCount() (int, error) {
	if len(ctx.loopStack) == 0 {
		return 0, bytecodeUnsupported("break/continue outside loop")
	}
	loop := ctx.loopStack[len(ctx.loopStack)-1]
	exitCount := ctx.scopeDepth - loop.scopeDepth
	if exitCount <= 0 {
		return 0, bytecodeUnsupported("loop scope mismatch")
	}
	return exitCount, nil
}
