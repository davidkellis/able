package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) lowerFunctionDefinitionBytecodeWithEnv(def *ast.FunctionDefinition, env *runtime.Environment) (*bytecodeProgram, error) {
	return i.lowerFunctionDefinitionBytecodeWithMethodSetEnv(def, env, nil)
}

func (i *Interpreter) lowerFunctionDefinitionBytecodeWithMethodSetEnv(def *ast.FunctionDefinition, env *runtime.Environment, methodSet *runtime.MethodSet) (*bytecodeProgram, error) {
	if def == nil || def.Body == nil {
		return nil, nil
	}
	layout := analyzeFrameLayoutWithEnvAndMethodSet(i, def, env, methodSet)
	if layout == nil {
		ctx := &bytecodeLoweringContext{
			instructions:           make([]bytecodeInstruction, 0, len(def.Body.Body)*2),
			allowPlaceholderLambda: true,
			collectScalarProofs:    i != nil && i.bytecodeStatsEnabled,
			definitionEnv:          env,
			methodSet:              methodSet,
		}
		seedBytecodeLoweringStructDefs(ctx, env)
		if err := emitBlock(ctx, i, def.Body); err != nil {
			return nil, err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpReturn})
		bytecodeFuseImplicitReturnBinary(ctx.instructions, nil)
		program := finalizeBytecodeProgramMetadata(&bytecodeProgram{
			instructions:              ctx.instructions,
			namedStructLiterals:       ctx.namedStructLiterals,
			namedStructMembers:        ctx.namedStructMembers,
			f64DotLoops:               ctx.f64DotLoops,
			f64MatrixRowLoops:         ctx.f64MatrixRowLoops,
			f64AffineRowLoops:         ctx.f64AffineRowLoops,
			f64TransposeRowLoops:      ctx.f64TransposeRowLoops,
			f64AffinePushes:           ctx.f64AffinePushes,
			f64NestedGetPushes:        ctx.f64NestedGetPushes,
			floatMulAddMulJumps:       ctx.floatMulAddMulJumps,
			floatAddCompareConstJumps: ctx.floatAddCompareConstJumps,
			floatAffineStores:         ctx.floatAffineStores,
			floatUpdatePairs:          ctx.floatUpdatePairs,
			floatRegions:              ctx.floatRegions,
		})
		return i.annotateBytecodeProgramReachWithScalarChecks(program, "function", bytecodeFunctionReachName(def), def, ctx.scalarProofChecks), nil
	}
	ctx := &bytecodeLoweringContext{
		instructions:           make([]bytecodeInstruction, 0, len(def.Body.Body)*2),
		allowPlaceholderLambda: true,
		collectScalarProofs:    i != nil && i.bytecodeStatsEnabled,
		frameLayout:            layout,
		definitionEnv:          env,
		slotKinds:              append([]bytecodeCellKind(nil), layout.paramKinds...),
		slotSimpleChecks:       append([]bytecodeSimpleTypeCheck(nil), layout.paramSimpleChecks...),
		slotExactStructDefs:    append([]*runtime.StructDefinitionValue(nil), layout.paramExactStructDef...),
		nextSlot:               layout.paramSlots,
		selfCallSlot:           -1,
		methodSet:              methodSet,
	}
	if def.ID != nil {
		ctx.currentFunctionName = def.ID.Name
	}
	seedBytecodeLoweringStructDefs(ctx, env)
	for idx, paramType := range layout.paramTypes {
		if def := bytecodeNominalNamedStructDefinitionForTypeExpr(ctx, paramType); def != nil {
			ctx.setSlotExactStructDef(idx, def)
		}
	}
	if canUseSelfCallSlot(def) {
		layout.selfCallSlot = ctx.nextSlot
		ctx.selfCallSlot = ctx.nextSlot
		ctx.nextSlot++
		ctx.setSlotKind(layout.selfCallSlot, bytecodeCellKindValue)
		ctx.selfCallName = ctx.currentFunctionName
	}
	paramScope := make(map[string]int, layout.paramSlots)
	for idx, param := range def.Params {
		if ident, ok := param.Name.(*ast.Identifier); ok {
			paramScope[ident.Name] = idx
		}
	}
	ctx.slotScopes = []map[string]int{paramScope}
	ctx.implicitSlotScopes = []map[string]int{make(map[string]int)}
	if err := emitBlock(ctx, i, def.Body); err != nil {
		return nil, err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpReturn})
	bytecodeFuseImplicitReturnBinary(ctx.instructions, layout)
	layout.slotCount = ctx.nextSlot
	layout.slotKinds = make([]bytecodeCellKind, layout.slotCount)
	copy(layout.slotKinds, ctx.slotKinds)
	layout.hasTypedSlots = false
	for _, kind := range layout.slotKinds {
		if kind != bytecodeCellKindValue {
			layout.hasTypedSlots = true
			break
		}
	}
	layout.i32RegisterFrame = layout.i32RegisterFrame && layout.hasTypedSlots
	bytecodeAttachI32FrameProof(layout, def, ctx.instructions, i.bytecodeInferenceFactsSnapshot())
	program := finalizeBytecodeProgramMetadata(&bytecodeProgram{
		instructions:              ctx.instructions,
		frameLayout:               layout,
		namedStructLiterals:       ctx.namedStructLiterals,
		namedStructMembers:        ctx.namedStructMembers,
		f64DotLoops:               ctx.f64DotLoops,
		f64MatrixRowLoops:         ctx.f64MatrixRowLoops,
		f64AffineRowLoops:         ctx.f64AffineRowLoops,
		f64TransposeRowLoops:      ctx.f64TransposeRowLoops,
		f64AffinePushes:           ctx.f64AffinePushes,
		f64NestedGetPushes:        ctx.f64NestedGetPushes,
		floatMulAddMulJumps:       ctx.floatMulAddMulJumps,
		floatAddCompareConstJumps: ctx.floatAddCompareConstJumps,
		floatAffineStores:         ctx.floatAffineStores,
		floatUpdatePairs:          ctx.floatUpdatePairs,
		floatRegions:              ctx.floatRegions,
	})
	program = i.annotateBytecodeProgramReachWithScalarChecks(program, "function", bytecodeFunctionReachName(def), def, ctx.scalarProofChecks)
	program.i32RecurrenceKernel = bytecodeDetectI32RecurrenceKernel(program)
	return program, nil
}

func bytecodeFunctionReachName(def *ast.FunctionDefinition) string {
	if def != nil && def.ID != nil && def.ID.Name != "" {
		return def.ID.Name
	}
	return "<anonymous-function>"
}

func seedBytecodeLoweringStructDefs(ctx *bytecodeLoweringContext, env *runtime.Environment) {
	if ctx == nil || env == nil {
		return
	}
	for cur := env; cur != nil; cur = cur.Parent() {
		cur.ForEachCurrentStructDefinition(func(name string, def *runtime.StructDefinitionValue) bool {
			if def == nil || def.Node == nil || name == "" {
				return true
			}
			if ctx.structDefs == nil {
				ctx.structDefs = make(map[string]*ast.StructDefinition)
			}
			if _, exists := ctx.structDefs[name]; exists {
				return true
			}
			ctx.structDefs[name] = def.Node
			if ctx.structDefValues == nil {
				ctx.structDefValues = make(map[string]*runtime.StructDefinitionValue)
			}
			ctx.structDefValues[name] = def
			return true
		})
	}
}
