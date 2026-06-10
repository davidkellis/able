package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func emitSimpleNamedStructLiteral(ctx *bytecodeLoweringContext, i *Interpreter, lit *ast.StructLiteral) (bool, error) {
	if ctx == nil || lit == nil || !simpleNamedStructLiteralEligible(ctx, lit) {
		return false, nil
	}
	plan := bytecodeNamedStructLiteralPlan{
		definition: bytecodeNamedStructLiteralDefinition(ctx, lit.StructType.Name),
	}
	if def := ctx.structDefs[lit.StructType.Name]; def != nil {
		fieldOrder, err := namedStructLiteralFieldOrder(lit, def)
		if err != nil {
			return false, err
		}
		plan.fieldOrder = fieldOrder
	}
	for _, field := range lit.Fields {
		value, ok := simpleNamedStructLiteralFieldValue(field)
		if !ok {
			return false, bytecodeUnsupported("nil struct field initializer")
		}
		if err := emitExpression(ctx, i, value); err != nil {
			return false, err
		}
	}
	fastIP := ctx.emit(bytecodeInstruction{
		op:       bytecodeOpStructLiteralNamedFast,
		name:     lit.StructType.Name,
		argCount: len(lit.Fields),
		node:     lit,
	})
	if ctx.namedStructLiterals == nil {
		ctx.namedStructLiterals = make(map[int]bytecodeNamedStructLiteralPlan, 1)
	}
	ctx.namedStructLiterals[fastIP] = plan
	return true, nil
}

func bytecodeNamedStructLiteralFieldOrder(ctx *bytecodeLoweringContext, lit *ast.StructLiteral) ([]int, error) {
	if ctx == nil || lit == nil || lit.StructType == nil {
		return nil, bytecodeUnsupported("nil named struct literal plan")
	}
	def, ok := ctx.structDefs[lit.StructType.Name]
	if !ok || def == nil {
		return nil, bytecodeUnsupported("missing named struct definition for fast literal")
	}
	fieldOrder, err := namedStructLiteralFieldOrder(lit, def)
	if err != nil {
		return nil, bytecodeUnsupported(err.Error())
	}
	return fieldOrder, nil
}

func bytecodeNamedStructLiteralDefinition(ctx *bytecodeLoweringContext, name string) *runtime.StructDefinitionValue {
	if ctx == nil || name == "" || ctx.structDefValues == nil {
		return nil
	}
	return ctx.structDefValues[name]
}

func simpleNamedStructLiteralEligible(ctx *bytecodeLoweringContext, lit *ast.StructLiteral) bool {
	if ctx == nil || !simpleNamedStructLiteralSyntacticEligible(lit) {
		return false
	}
	def := ctx.structDefs[lit.StructType.Name]
	if def == nil {
		// Function bytecode can be lowered with a closure environment that has
		// the definition while the lowering context has not retained its AST.
		// The named-literal opcode resolves that definition at execution time;
		// evaluating fields here is still required so local slot values do not
		// fall back to tree evaluation through the closure environment.
		return true
	}
	if isSingletonStructDef(def) {
		return true
	}
	return simpleNamedStructLiteralEligibleForDefinition(lit, def)
}
