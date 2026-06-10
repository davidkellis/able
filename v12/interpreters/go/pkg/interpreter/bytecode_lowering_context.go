package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeSimpleLookupName(name string) bool {
	return name != "" && !strings.Contains(name, ".")
}

func bytecodeIdentifierMemberName(expr ast.Expression) string {
	if ident, ok := expr.(*ast.Identifier); ok && ident != nil {
		return ident.Name
	}
	return ""
}

func bytecodeDottedSlotMemberCall(ctx *bytecodeLoweringContext, ident *ast.Identifier) (int, string, bool) {
	if ctx == nil || ident == nil || ident.Name == "" {
		return 0, "", false
	}
	head, tail, ok := strings.Cut(ident.Name, ".")
	if !ok || head == "" || tail == "" {
		return 0, "", false
	}
	slot, found := ctx.lookupSlot(head)
	if !found {
		return 0, "", false
	}
	return slot, tail, true
}

func bytecodeCallMemberOpcodeForName(name string, argCount int) bytecodeOp {
	switch {
	case name == "get" && argCount == 1:
		return bytecodeOpCallMemberArrayGet
	case name == "next" && argCount == 0:
		return bytecodeOpCallMemberNext
	case name == "new" && argCount == 0:
		return bytecodeOpCallMemberArrayNew
	case bytecodeArraySlotCallShape(name, argCount):
		return bytecodeOpCallMemberArraySlot
	default:
		return bytecodeOpCallMember
	}
}

func bytecodeCallMemberInstructionForName(name string, argCount int, node ast.Node) bytecodeInstruction {
	instr := bytecodeInstruction{
		op:       bytecodeCallMemberOpcodeForName(name, argCount),
		name:     name,
		argCount: argCount,
		node:     node,
	}
	if instr.op == bytecodeOpCallMemberArraySlot {
		if kind, ok := bytecodeArraySlotCallFastPathForName(name, argCount); ok {
			instr.memberFastPath = kind
		}
	}
	return instr
}

type bytecodeLoweringContext struct {
	instructions              []bytecodeInstruction
	scopeDepth                int
	scopeStack                []bytecodeLexicalScope
	loopStack                 []loopContext
	allowPlaceholderLambda    bool
	frameLayout               *bytecodeFrameLayout // non-nil = slot mode
	definitionEnv             *runtime.Environment
	slotScopes                []map[string]int   // scope stack for slot lookups
	implicitSlotScopes        []map[string]int   // runtime-guarded `=` local slots
	slotKinds                 []bytecodeCellKind // typed-cell kind by slot while lowering
	slotSimpleChecks          []bytecodeSimpleTypeCheck
	collectScalarProofs       bool
	scalarProofChecks         []bytecodeSimpleTypeCheck
	nextSlot                  int    // next available slot index
	currentFunctionName       string // current function name, if any
	selfCallName              string // current function name for self-recursive call lowering
	selfCallSlot              int    // reserved slot for self-recursive call fast path
	discardExpressionValue    bool
	discardExpressionNode     ast.Expression
	methodSet                 *runtime.MethodSet
	structDefs                map[string]*ast.StructDefinition
	structDefValues           map[string]*runtime.StructDefinitionValue
	slotExactStructDefs       []*runtime.StructDefinitionValue
	namedStructLiterals       map[int]bytecodeNamedStructLiteralPlan
	namedStructMembers        map[int]bytecodeNamedStructMemberPlan
	f64DotLoops               map[int]bytecodeF64DotLoopPlan
	f64MatrixRowLoops         map[int]bytecodeF64MatrixRowLoopPlan
	f64AffineRowLoops         map[int]bytecodeF64AffineRowLoopPlan
	f64TransposeRowLoops      map[int]bytecodeF64TransposeRowLoopPlan
	f64AffinePushes           map[int]bytecodeF64AffineProductPushPlan
	f64NestedGetPushes        map[int]bytecodeF64NestedArrayGetPushPlan
	floatMulAddMulJumps       map[int]bytecodeFloatMulAddMulCompareConstJumpPlan
	floatAddCompareConstJumps map[int]bytecodeFloatAddCompareConstJumpPlan
	floatAffineStores         map[int]bytecodeStoreSlotFloatAffinePlan
	floatUpdatePairs          map[int]bytecodeFloatUpdatePairPlan
	floatRegions              []bytecodeFloatRegionPlan
}

type bytecodeLexicalScope struct {
	runtime bool
	slots   bool
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
		collectScalarProofs:    i != nil && i.bytecodeStatsEnabled,
	}
	if len(module.Body) == 0 {
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
		ctx.emit(bytecodeInstruction{op: bytecodeOpReturn})
		program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: ctx.instructions})
		return i.annotateBytecodeProgramReachWithScalarChecks(program, "module", "<module>", module, ctx.scalarProofChecks), nil
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
	bytecodeFuseImplicitReturnBinary(ctx.instructions, nil)
	program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: ctx.instructions, namedStructLiterals: ctx.namedStructLiterals, namedStructMembers: ctx.namedStructMembers, f64DotLoops: ctx.f64DotLoops, f64MatrixRowLoops: ctx.f64MatrixRowLoops, f64AffineRowLoops: ctx.f64AffineRowLoops, f64TransposeRowLoops: ctx.f64TransposeRowLoops, f64AffinePushes: ctx.f64AffinePushes, f64NestedGetPushes: ctx.f64NestedGetPushes, floatMulAddMulJumps: ctx.floatMulAddMulJumps, floatAddCompareConstJumps: ctx.floatAddCompareConstJumps, floatAffineStores: ctx.floatAffineStores, floatUpdatePairs: ctx.floatUpdatePairs, floatRegions: ctx.floatRegions})
	return i.annotateBytecodeProgramReachWithScalarChecks(program, "module", "<module>", module, ctx.scalarProofChecks), nil
}

func (i *Interpreter) lowerExpressionToBytecode(expr ast.Expression) (*bytecodeProgram, error) {
	return i.lowerExpressionToBytecodeWithOptions(expr, true)
}

func (i *Interpreter) lowerExpressionToBytecodeWithOptions(expr ast.Expression, allowPlaceholderLambda bool) (*bytecodeProgram, error) {
	if expr == nil {
		return nil, fmt.Errorf("bytecode lowering expression is nil")
	}
	if cached, ok := i.lookupCachedExpressionBytecode(expr, allowPlaceholderLambda); ok {
		i.recordBytecodeExpressionCacheHit()
		return cached, nil
	}
	i.recordBytecodeExpressionCacheMiss()
	ctx := &bytecodeLoweringContext{
		instructions:           make([]bytecodeInstruction, 0, 4),
		allowPlaceholderLambda: allowPlaceholderLambda,
		collectScalarProofs:    i != nil && i.bytecodeStatsEnabled,
	}
	if emitted, err := bytecodeEmitFinalI32StackExpr(ctx, expr); err != nil {
		return nil, err
	} else if !emitted {
		if err := emitExpression(ctx, i, expr); err != nil {
			return nil, err
		}
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpReturn})
	bytecodeFuseImplicitReturnBinary(ctx.instructions, nil)
	program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: ctx.instructions, namedStructLiterals: ctx.namedStructLiterals, namedStructMembers: ctx.namedStructMembers, f64DotLoops: ctx.f64DotLoops, f64MatrixRowLoops: ctx.f64MatrixRowLoops, f64AffineRowLoops: ctx.f64AffineRowLoops, f64TransposeRowLoops: ctx.f64TransposeRowLoops, f64AffinePushes: ctx.f64AffinePushes, f64NestedGetPushes: ctx.f64NestedGetPushes, floatMulAddMulJumps: ctx.floatMulAddMulJumps, floatAddCompareConstJumps: ctx.floatAddCompareConstJumps, floatAffineStores: ctx.floatAffineStores, floatUpdatePairs: ctx.floatUpdatePairs, floatRegions: ctx.floatRegions})
	program = i.annotateBytecodeProgramReachWithScalarChecks(program, "expression", "<expression>", expr, ctx.scalarProofChecks)
	return i.cacheExpressionBytecode(expr, allowPlaceholderLambda, program), nil
}

func (i *Interpreter) lowerBlockExpressionToBytecode(block *ast.BlockExpression, allowPlaceholderLambda bool) (*bytecodeProgram, error) {
	if block == nil {
		return nil, fmt.Errorf("bytecode lowering block is nil")
	}
	ctx := &bytecodeLoweringContext{
		instructions:           make([]bytecodeInstruction, 0, len(block.Body)*2),
		allowPlaceholderLambda: allowPlaceholderLambda,
		collectScalarProofs:    i != nil && i.bytecodeStatsEnabled,
	}
	if err := emitBlock(ctx, i, block); err != nil {
		return nil, err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpReturn})
	bytecodeFuseImplicitReturnBinary(ctx.instructions, nil)
	program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: ctx.instructions, namedStructLiterals: ctx.namedStructLiterals, namedStructMembers: ctx.namedStructMembers, f64DotLoops: ctx.f64DotLoops, f64MatrixRowLoops: ctx.f64MatrixRowLoops, f64AffineRowLoops: ctx.f64AffineRowLoops, f64TransposeRowLoops: ctx.f64TransposeRowLoops, f64AffinePushes: ctx.f64AffinePushes, f64NestedGetPushes: ctx.f64NestedGetPushes, floatMulAddMulJumps: ctx.floatMulAddMulJumps, floatAddCompareConstJumps: ctx.floatAddCompareConstJumps, floatAffineStores: ctx.floatAffineStores, floatUpdatePairs: ctx.floatUpdatePairs, floatRegions: ctx.floatRegions})
	return i.annotateBytecodeProgramReachWithScalarChecks(program, "block", "<block>", block, ctx.scalarProofChecks), nil
}
