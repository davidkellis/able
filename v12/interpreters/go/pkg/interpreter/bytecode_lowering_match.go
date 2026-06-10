package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeSlotMatchPatternKind uint8

const (
	bytecodeSlotMatchPatternWildcard bytecodeSlotMatchPatternKind = iota
	bytecodeSlotMatchPatternNil
	bytecodeSlotMatchPatternIdentifier
	bytecodeSlotMatchPatternTyped
	bytecodeSlotMatchPatternNamedStructFields
)

type bytecodeSlotMatchClausePlan struct {
	kind          bytecodeSlotMatchPatternKind
	typeExpr      ast.TypeExpression
	bindingName   string
	slotKind      bytecodeCellKind
	structDef     *runtime.StructDefinitionValue
	fieldBindings []bytecodeSlotMatchFieldBinding
}

type bytecodeSlotMatchFieldBinding struct {
	fieldName      string
	fieldIndex     int
	bindingName    string
	slotKind       bytecodeCellKind
	exactStructDef *runtime.StructDefinitionValue
}

type bytecodeSlotMatchFieldBindingNames struct {
	first  string
	second string
	count  int
}

func (names bytecodeSlotMatchFieldBindingNames) at(index int) string {
	switch index {
	case 0:
		return names.first
	case 1:
		return names.second
	default:
		return ""
	}
}

func (plan bytecodeSlotMatchClausePlan) hasSlotBindings() bool {
	return plan.bindingName != "" || len(plan.fieldBindings) != 0
}

func emitSlotMatch(ctx *bytecodeLoweringContext, i *Interpreter, expr *ast.MatchExpression) (bool, error) {
	if ctx == nil || ctx.frameLayout == nil || expr == nil {
		return false, nil
	}
	plans, ok := bytecodeSlotMatchClausePlansForExpression(ctx, expr)
	if !ok {
		return false, nil
	}
	if err := emitExpression(ctx, i, expr.Subject); err != nil {
		return false, err
	}
	subjectName := fmt.Sprintf("$match_subject_%d", ctx.nextSlot)
	subjectSlot := ctx.declareSlotWithKind(subjectName, bytecodeCellKindValue)
	ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: subjectSlot, name: subjectName, node: expr})
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})

	endJumps := make([]int, 0, len(expr.Clauses))
	for idx, clause := range expr.Clauses {
		if clause == nil {
			continue
		}
		plan := plans[idx]
		nextJump := emitSlotMatchPatternTest(ctx, subjectSlot, plan, clause)
		ctx.enterScope(false, plan.hasSlotBindings(), false)
		emitSlotMatchPatternBinding(ctx, subjectSlot, plan, clause)
		if err := emitExpression(ctx, i, clause.Body); err != nil {
			return false, err
		}
		ctx.exitScope()
		endJumps = append(endJumps, ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1}))
		if nextJump >= 0 {
			ctx.patchJump(nextJump, len(ctx.instructions))
		}
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpMatchNoClause, node: expr})
	end := len(ctx.instructions)
	for _, jump := range endJumps {
		ctx.patchJump(jump, end)
	}
	return true, nil
}

func emitEnvMatch(ctx *bytecodeLoweringContext, i *Interpreter, expr *ast.MatchExpression) (bool, error) {
	if ctx == nil || ctx.frameLayout != nil || expr == nil {
		return false, nil
	}
	plans, ok := bytecodeEnvMatchClausePlansForExpression(ctx, expr)
	if !ok {
		return false, nil
	}
	if err := emitExpression(ctx, i, expr.Subject); err != nil {
		return false, err
	}

	endJumps := make([]int, 0, len(expr.Clauses))
	for idx, clause := range expr.Clauses {
		if clause == nil {
			continue
		}
		plan := plans[idx]
		nextJump := emitEnvMatchPatternTest(ctx, plan, clause)
		ctx.enterScope(true, false, true)
		emitEnvMatchPatternBinding(ctx, plan, clause)
		ctx.emit(bytecodeInstruction{op: bytecodeOpPop, node: clause})
		if err := emitExpression(ctx, i, clause.Body); err != nil {
			return false, err
		}
		ctx.exitScope()
		endJumps = append(endJumps, ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1}))
		if nextJump >= 0 {
			ctx.patchJump(nextJump, len(ctx.instructions))
		}
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop, node: expr})
	ctx.emit(bytecodeInstruction{op: bytecodeOpMatchNoClause, node: expr})
	end := len(ctx.instructions)
	for _, jump := range endJumps {
		ctx.patchJump(jump, end)
	}
	return true, nil
}

func bytecodeEnvMatchClausePlansForExpression(ctx *bytecodeLoweringContext, expr *ast.MatchExpression) ([]bytecodeSlotMatchClausePlan, bool) {
	if ctx == nil || expr == nil || len(expr.Clauses) == 0 {
		return nil, false
	}
	plans := make([]bytecodeSlotMatchClausePlan, len(expr.Clauses))
	for idx, clause := range expr.Clauses {
		if clause == nil || clause.Guard != nil {
			return nil, false
		}
		plan, ok := bytecodeEnvMatchClausePlanForPattern(ctx, clause.Pattern)
		if !ok {
			return nil, false
		}
		plans[idx] = plan
	}
	return plans, true
}

func bytecodeEnvMatchClausePlanForPattern(ctx *bytecodeLoweringContext, pattern ast.Pattern) (bytecodeSlotMatchClausePlan, bool) {
	plan, ok := bytecodeSlotMatchClausePlanForPattern(ctx, pattern)
	if !ok || plan.kind == bytecodeSlotMatchPatternNamedStructFields {
		return bytecodeSlotMatchClausePlan{}, false
	}
	return plan, true
}

func bytecodeCanLowerSlotMatch(ctx *bytecodeLoweringContext, expr *ast.MatchExpression) bool {
	_, ok := bytecodeSlotMatchClausePlansForExpression(ctx, expr)
	return ok
}

func bytecodeSlotMatchClausePlansForExpression(ctx *bytecodeLoweringContext, expr *ast.MatchExpression) ([]bytecodeSlotMatchClausePlan, bool) {
	if ctx == nil || expr == nil {
		return nil, false
	}
	if len(expr.Clauses) == 0 {
		return nil, false
	}
	plans := make([]bytecodeSlotMatchClausePlan, len(expr.Clauses))
	for idx, clause := range expr.Clauses {
		if clause == nil || clause.Guard != nil {
			return nil, false
		}
		plan, ok := bytecodeSlotMatchClausePlanForPattern(ctx, clause.Pattern)
		if !ok {
			return nil, false
		}
		plans[idx] = plan
	}
	return plans, true
}

func bytecodeCanLowerSlotMatchInEnv(expr *ast.MatchExpression, env *runtime.Environment) bool {
	if expr == nil {
		return false
	}
	if len(expr.Clauses) == 0 {
		return false
	}
	for _, clause := range expr.Clauses {
		if clause == nil || clause.Guard != nil {
			return false
		}
		if !bytecodeCanLowerSlotMatchPatternInEnv(clause.Pattern, env) {
			return false
		}
	}
	return true
}

func bytecodeSlotMatchClausePlanForPattern(ctx *bytecodeLoweringContext, pattern ast.Pattern) (bytecodeSlotMatchClausePlan, bool) {
	switch p := pattern.(type) {
	case *ast.Identifier:
		if !bytecodeSlotMatchBindingIdentifierCanLower(ctx, p) {
			return bytecodeSlotMatchClausePlan{}, false
		}
		return bytecodeSlotMatchClausePlan{
			kind:        bytecodeSlotMatchPatternIdentifier,
			bindingName: p.Name,
			slotKind:    bytecodeCellKindValue,
		}, true
	case *ast.WildcardPattern:
		if p == nil {
			return bytecodeSlotMatchClausePlan{}, false
		}
		return bytecodeSlotMatchClausePlan{kind: bytecodeSlotMatchPatternWildcard}, true
	case *ast.LiteralPattern:
		if p == nil {
			return bytecodeSlotMatchClausePlan{}, false
		}
		if _, ok := p.Literal.(*ast.NilLiteral); ok {
			return bytecodeSlotMatchClausePlan{kind: bytecodeSlotMatchPatternNil}, true
		}
	case *ast.StructPattern:
		if typeExpr, ok := bytecodeSlotMatchZeroFieldStructPatternTypeExprInContext(ctx, p); ok {
			return bytecodeSlotMatchClausePlan{
				kind:      bytecodeSlotMatchPatternTyped,
				typeExpr:  typeExpr,
				slotKind:  bytecodeCellKindForTypeExpr(typeExpr),
				structDef: bytecodeNamedStructDefinitionForTypeExpr(ctx, typeExpr),
			}, true
		}
		return bytecodeSlotMatchNamedStructFieldPlanInContext(ctx, p)
	case *ast.TypedPattern:
		if p == nil || p.TypeAnnotation == nil {
			return bytecodeSlotMatchClausePlan{}, false
		}
		plan := bytecodeSlotMatchClausePlan{
			kind:     bytecodeSlotMatchPatternTyped,
			typeExpr: p.TypeAnnotation,
			slotKind: bytecodeCellKindForTypeExpr(p.TypeAnnotation),
		}
		switch inner := p.Pattern.(type) {
		case *ast.Identifier:
			if inner != nil && inner.Name != "" && inner.Name != "_" {
				if !bytecodeSlotMatchBindingIdentifierCanLower(ctx, inner) {
					return bytecodeSlotMatchClausePlan{}, false
				}
				plan.bindingName = inner.Name
			}
			return plan, true
		case *ast.WildcardPattern:
			return plan, inner != nil
		}
	}
	return bytecodeSlotMatchClausePlan{}, false
}

func bytecodeCanLowerSlotMatchPatternInEnv(pattern ast.Pattern, env *runtime.Environment) bool {
	switch p := pattern.(type) {
	case *ast.Identifier:
		return bytecodeSlotMatchBindingIdentifierCanLowerInEnv(p, env)
	case *ast.WildcardPattern:
		return p != nil
	case *ast.LiteralPattern:
		if p == nil {
			return false
		}
		_, ok := p.Literal.(*ast.NilLiteral)
		return ok
	case *ast.StructPattern:
		if _, ok := bytecodeSlotMatchZeroFieldStructPatternTypeExprInEnv(p, env); ok {
			return true
		}
		return bytecodeCanLowerSlotMatchNamedStructFieldPatternInEnv(p, env)
	case *ast.TypedPattern:
		if p == nil || p.TypeAnnotation == nil {
			return false
		}
		switch inner := p.Pattern.(type) {
		case *ast.Identifier:
			if inner == nil {
				return false
			}
			if inner.Name == "" || inner.Name == "_" {
				return inner.Name == "_"
			}
			return bytecodeSlotMatchBindingIdentifierCanLowerInEnv(inner, env)
		case *ast.WildcardPattern:
			return inner != nil
		}
	}
	return false
}

func bytecodeSlotMatchZeroFieldStructPatternTypeExprInContext(ctx *bytecodeLoweringContext, pattern *ast.StructPattern) (ast.TypeExpression, bool) {
	name, ok := bytecodeSlotMatchZeroFieldStructPatternName(pattern)
	if !ok {
		return nil, false
	}
	if name == "IteratorEnd" {
		return ast.Ty(name), true
	}
	return bytecodeSlotMatchZeroFieldStructPatternTypeExprForDef(name, bytecodeSlotMatchStructDefinitionInContext(ctx, name))
}

func bytecodeSlotMatchZeroFieldStructPatternTypeExprInEnv(pattern *ast.StructPattern, env *runtime.Environment) (ast.TypeExpression, bool) {
	name, ok := bytecodeSlotMatchZeroFieldStructPatternName(pattern)
	if !ok {
		return nil, false
	}
	if name == "IteratorEnd" {
		return ast.Ty(name), true
	}
	if env == nil {
		return nil, false
	}
	def, ok := env.StructDefinition(name)
	if !ok || def == nil {
		return nil, false
	}
	return bytecodeSlotMatchZeroFieldStructPatternTypeExprForDef(name, def.Node)
}

func bytecodeSlotMatchZeroFieldStructPatternName(pattern *ast.StructPattern) (string, bool) {
	if pattern == nil || pattern.StructType == nil || pattern.StructType.Name == "" || len(pattern.Fields) != 0 {
		return "", false
	}
	return pattern.StructType.Name, true
}

func bytecodeSlotMatchStructDefinitionInContext(ctx *bytecodeLoweringContext, name string) *ast.StructDefinition {
	if ctx == nil || name == "" {
		return nil
	}
	if def := ctx.structDefs[name]; def != nil {
		return def
	}
	if def := ctx.structDefValues[name]; def != nil {
		return def.Node
	}
	return nil
}

func bytecodeSlotMatchZeroFieldStructPatternTypeExprForDef(name string, def *ast.StructDefinition) (ast.TypeExpression, bool) {
	if !bytecodeSlotMatchZeroFieldStructPatternCanUseTypedPattern(def) {
		return nil, false
	}
	return ast.Ty(name), true
}

func bytecodeSlotMatchZeroFieldStructPatternCanUseTypedPattern(def *ast.StructDefinition) bool {
	if def == nil {
		return false
	}
	switch def.Kind {
	case ast.StructKindNamed:
		return true
	case ast.StructKindSingleton:
		return true
	case ast.StructKindPositional:
		return len(def.Fields) == 0
	default:
		return false
	}
}

func bytecodeSlotMatchNamedStructFieldPlanInContext(ctx *bytecodeLoweringContext, pattern *ast.StructPattern) (bytecodeSlotMatchClausePlan, bool) {
	name, def, ok := bytecodeSlotMatchNamedStructFieldPatternDefinition(ctx, pattern)
	if !ok {
		return bytecodeSlotMatchClausePlan{}, false
	}
	bindings := make([]bytecodeSlotMatchFieldBinding, 0, len(pattern.Fields))
	for _, field := range pattern.Fields {
		fieldIndex, ok := bytecodeSlotMatchNamedStructFieldIndex(field, def)
		if !ok {
			return bytecodeSlotMatchClausePlan{}, false
		}
		bindingNames, ok := bytecodeSlotMatchFieldBindingNamesInContext(ctx, field)
		if !ok {
			return bytecodeSlotMatchClausePlan{}, false
		}
		var fieldType ast.TypeExpression
		if fieldDef := def.Fields[fieldIndex]; fieldDef != nil {
			fieldType = fieldDef.FieldType
		}
		fieldName := field.FieldName.Name
		for bindingIdx := 0; bindingIdx < bindingNames.count; bindingIdx++ {
			bindingName := bindingNames.at(bindingIdx)
			bindings = append(bindings, bytecodeSlotMatchFieldBinding{
				fieldName:      fieldName,
				fieldIndex:     fieldIndex,
				bindingName:    bindingName,
				slotKind:       bytecodeCellKindForTypeExpr(fieldType),
				exactStructDef: bytecodeNominalNamedStructDefinitionForTypeExpr(ctx, fieldType),
			})
		}
	}
	typeExpr := ast.Ty(name)
	return bytecodeSlotMatchClausePlan{
		kind:          bytecodeSlotMatchPatternNamedStructFields,
		typeExpr:      typeExpr,
		slotKind:      bytecodeCellKindForTypeExpr(typeExpr),
		structDef:     bytecodeNamedStructLiteralDefinition(ctx, name),
		fieldBindings: bindings,
	}, true
}

func bytecodeCanLowerSlotMatchNamedStructFieldPatternInEnv(pattern *ast.StructPattern, env *runtime.Environment) bool {
	if env == nil {
		return false
	}
	_, def, ok := bytecodeSlotMatchNamedStructFieldPatternDefinitionInEnv(pattern, env)
	if !ok {
		return false
	}
	for _, field := range pattern.Fields {
		if _, ok := bytecodeSlotMatchNamedStructFieldIndex(field, def); !ok {
			return false
		}
		if _, ok := bytecodeSlotMatchFieldBindingNamesInEnv(field, env); !ok {
			return false
		}
	}
	return true
}

func bytecodeSlotMatchNamedStructFieldPatternDefinition(ctx *bytecodeLoweringContext, pattern *ast.StructPattern) (string, *ast.StructDefinition, bool) {
	if ctx == nil || pattern == nil || pattern.StructType == nil || pattern.StructType.Name == "" || pattern.IsPositional || len(pattern.Fields) == 0 {
		return "", nil, false
	}
	name := pattern.StructType.Name
	def := bytecodeSlotMatchStructDefinitionInContext(ctx, name)
	if def == nil || def.Kind != ast.StructKindNamed {
		return "", nil, false
	}
	return name, def, true
}

func bytecodeSlotMatchNamedStructFieldPatternDefinitionInEnv(pattern *ast.StructPattern, env *runtime.Environment) (string, *ast.StructDefinition, bool) {
	if env == nil || pattern == nil || pattern.StructType == nil || pattern.StructType.Name == "" || pattern.IsPositional || len(pattern.Fields) == 0 {
		return "", nil, false
	}
	name := pattern.StructType.Name
	defVal, ok := env.StructDefinition(name)
	if !ok || defVal == nil || defVal.Node == nil || defVal.Node.Kind != ast.StructKindNamed {
		return "", nil, false
	}
	return name, defVal.Node, true
}

func bytecodeSlotMatchNamedStructFieldOrder(pattern *ast.StructPattern, def *ast.StructDefinition) ([]int, bool) {
	plan, ok := buildNamedStructPatternPlan(pattern, def)
	if !ok || len(plan.fieldOrder) != len(pattern.Fields) {
		return nil, false
	}
	return plan.fieldOrder, true
}

func bytecodeSlotMatchNamedStructFieldIndex(field *ast.StructPatternField, def *ast.StructDefinition) (int, bool) {
	if field == nil || field.FieldName == nil || field.FieldName.Name == "" {
		return 0, false
	}
	return namedStructFieldIndex(def, field.FieldName.Name)
}

func bytecodeSlotMatchFieldBindingNamesInContext(ctx *bytecodeLoweringContext, field *ast.StructPatternField) (bytecodeSlotMatchFieldBindingNames, bool) {
	return bytecodeSlotMatchCollectFieldBindingNames(field, func(ident *ast.Identifier) bool {
		return bytecodeSlotMatchBindingIdentifierCanLower(ctx, ident)
	})
}

func bytecodeSlotMatchFieldBindingNamesInEnv(field *ast.StructPatternField, env *runtime.Environment) (bytecodeSlotMatchFieldBindingNames, bool) {
	return bytecodeSlotMatchCollectFieldBindingNames(field, func(ident *ast.Identifier) bool {
		return bytecodeSlotMatchBindingIdentifierCanLowerInEnv(ident, env)
	})
}

func bytecodeSlotMatchCollectFieldBindingNames(field *ast.StructPatternField, canBind func(*ast.Identifier) bool) (bytecodeSlotMatchFieldBindingNames, bool) {
	if field == nil || field.FieldName == nil || field.FieldName.Name == "" {
		return bytecodeSlotMatchFieldBindingNames{}, false
	}
	var bindings bytecodeSlotMatchFieldBindingNames
	switch p := field.Pattern.(type) {
	case *ast.Identifier:
		if p == nil {
			return bytecodeSlotMatchFieldBindingNames{}, false
		}
		if p.Name != "" && p.Name != "_" {
			if canBind == nil || !canBind(p) {
				return bytecodeSlotMatchFieldBindingNames{}, false
			}
			bindings.first = p.Name
			bindings.count = 1
		}
	case *ast.WildcardPattern:
		if p == nil {
			return bytecodeSlotMatchFieldBindingNames{}, false
		}
	default:
		return bytecodeSlotMatchFieldBindingNames{}, false
	}
	if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
		if canBind == nil || !canBind(field.Binding) {
			return bytecodeSlotMatchFieldBindingNames{}, false
		}
		if bindings.count == 0 {
			bindings.first = field.Binding.Name
			bindings.count = 1
		} else if bindings.first != field.Binding.Name {
			bindings.second = field.Binding.Name
			bindings.count = 2
		}
	}
	return bindings, true
}

func bytecodeSlotMatchBindingIdentifierCanLower(ctx *bytecodeLoweringContext, ident *ast.Identifier) bool {
	if ctx == nil || ident == nil || ident.Name == "" || ident.Name == "_" {
		return false
	}
	if _, found := ctx.lookupSlot(ident.Name); found {
		return true
	}
	return !bytecodeSlotMatchIdentifierResolvesToSingletonStruct(ctx, ident.Name)
}

func bytecodeSlotMatchBindingIdentifierCanLowerInEnv(ident *ast.Identifier, env *runtime.Environment) bool {
	if ident == nil || ident.Name == "" || ident.Name == "_" {
		return false
	}
	if env == nil {
		return true
	}
	if existing, ok := env.Lookup(ident.Name); ok {
		switch defVal := existing.(type) {
		case *runtime.StructDefinitionValue:
			return defVal == nil || !isSingletonStructDef(defVal.Node)
		case runtime.StructDefinitionValue:
			return !isSingletonStructDef(defVal.Node)
		default:
			return true
		}
	}
	defVal, ok := env.StructDefinition(ident.Name)
	if !ok {
		return true
	}
	return !isSingletonStructDef(defVal.Node)
}

func bytecodeSlotMatchIdentifierResolvesToSingletonStruct(ctx *bytecodeLoweringContext, name string) bool {
	if ctx == nil || name == "" {
		return false
	}
	if def := ctx.structDefs[name]; isSingletonStructDef(def) {
		return true
	}
	if def := ctx.structDefValues[name]; def != nil {
		return isSingletonStructDef(def.Node)
	}
	return false
}

func emitSlotMatchPatternTest(ctx *bytecodeLoweringContext, subjectSlot int, plan bytecodeSlotMatchClausePlan, clause *ast.MatchClause) int {
	switch plan.kind {
	case bytecodeSlotMatchPatternWildcard:
		return -1
	case bytecodeSlotMatchPatternNil:
		ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: subjectSlot, node: clause})
		return ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfNotNil, target: -1, node: clause})
	case bytecodeSlotMatchPatternIdentifier:
		ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: subjectSlot, node: clause})
		return -1
	case bytecodeSlotMatchPatternTyped, bytecodeSlotMatchPatternNamedStructFields:
		ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: subjectSlot, node: clause})
		genericStructDef := plan.structDef
		if genericStructDef == nil {
			genericStructDef = bytecodeNominalNamedStructDefinitionForTypeExpr(ctx, plan.typeExpr)
		}
		return ctx.emit(bytecodeInstruction{
			op:                 bytecodeOpJumpIfNotTypedPattern,
			target:             -1,
			typeExpr:           plan.typeExpr,
			typeSimpleCheck:    bytecodeSimpleTypeCheckForName(cachedSimpleTypeName(plan.typeExpr)),
			value:              plan.structDef,
			genericStructMatch: bytecodeGenericStructPatternPlanForTypeExprWithDefinition(plan.typeExpr, genericStructDef),
			node:               clause,
		})
	default:
		return -1
	}
}

func emitEnvMatchPatternTest(ctx *bytecodeLoweringContext, plan bytecodeSlotMatchClausePlan, clause *ast.MatchClause) int {
	switch plan.kind {
	case bytecodeSlotMatchPatternWildcard, bytecodeSlotMatchPatternIdentifier:
		return -1
	case bytecodeSlotMatchPatternNil:
		ctx.emit(bytecodeInstruction{op: bytecodeOpDup, node: clause})
		return ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfNotNil, target: -1, node: clause})
	case bytecodeSlotMatchPatternTyped:
		ctx.emit(bytecodeInstruction{op: bytecodeOpDup, node: clause})
		genericStructDef := plan.structDef
		if genericStructDef == nil {
			genericStructDef = bytecodeNominalNamedStructDefinitionForTypeExpr(ctx, plan.typeExpr)
		}
		return ctx.emit(bytecodeInstruction{
			op:                 bytecodeOpJumpIfNotTypedPattern,
			target:             -1,
			typeExpr:           plan.typeExpr,
			typeSimpleCheck:    bytecodeSimpleTypeCheckForName(cachedSimpleTypeName(plan.typeExpr)),
			value:              plan.structDef,
			genericStructMatch: bytecodeGenericStructPatternPlanForTypeExprWithDefinition(plan.typeExpr, genericStructDef),
			node:               clause,
		})
	default:
		return -1
	}
}

func emitSlotMatchPatternBinding(ctx *bytecodeLoweringContext, subjectSlot int, plan bytecodeSlotMatchClausePlan, clause *ast.MatchClause) {
	if plan.kind == bytecodeSlotMatchPatternNamedStructFields {
		emitSlotMatchNamedStructFieldBindings(ctx, subjectSlot, plan, clause)
		return
	}
	if plan.kind != bytecodeSlotMatchPatternTyped && plan.kind != bytecodeSlotMatchPatternIdentifier {
		return
	}
	if plan.bindingName == "" {
		if plan.kind == bytecodeSlotMatchPatternTyped {
			ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
		}
		return
	}
	slot := ctx.declareSlotWithKind(plan.bindingName, plan.slotKind)
	switch plan.kind {
	case bytecodeSlotMatchPatternTyped:
		ctx.setSlotExactStructDef(slot, bytecodeNominalNamedStructDefinitionForTypeExpr(ctx, plan.typeExpr))
	case bytecodeSlotMatchPatternIdentifier:
		ctx.setSlotExactStructDef(slot, ctx.slotExactStructDef(subjectSlot))
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: slot, name: plan.bindingName, node: clause})
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
}

func emitEnvMatchPatternBinding(ctx *bytecodeLoweringContext, plan bytecodeSlotMatchClausePlan, clause *ast.MatchClause) {
	switch plan.kind {
	case bytecodeSlotMatchPatternIdentifier:
		if plan.bindingName == "" {
			return
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpDup, node: clause})
		ctx.emit(bytecodeInstruction{op: bytecodeOpDeclareName, name: plan.bindingName, node: clause})
		ctx.emit(bytecodeInstruction{op: bytecodeOpPop, node: clause})
	case bytecodeSlotMatchPatternTyped:
		if plan.bindingName != "" {
			ctx.emit(bytecodeInstruction{op: bytecodeOpDeclareName, name: plan.bindingName, node: clause})
			ctx.emit(bytecodeInstruction{op: bytecodeOpPop, node: clause})
			return
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpPop, node: clause})
	}
}

func emitSlotMatchNamedStructFieldBindings(ctx *bytecodeLoweringContext, subjectSlot int, plan bytecodeSlotMatchClausePlan, clause *ast.MatchClause) {
	if ctx == nil {
		return
	}
	if len(plan.fieldBindings) == 0 {
		ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
		return
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlot, target: subjectSlot, node: clause, discardResult: true})
	for _, binding := range plan.fieldBindings {
		slot := ctx.declareSlotWithKind(binding.bindingName, binding.slotKind)
		ctx.setSlotExactStructDef(slot, binding.exactStructDef)
		memberIP := ctx.emit(bytecodeInstruction{
			op:     bytecodeOpLoadSlotStructField,
			target: subjectSlot,
			name:   binding.fieldName,
			node:   clause,
		})
		bytecodeStoreNamedStructFieldMemberPlan(ctx, memberIP, plan.structDef, binding.fieldIndex)
		ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: slot, name: binding.bindingName, node: clause})
		ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	}
}

func bytecodeStoreNamedStructFieldMemberPlan(ctx *bytecodeLoweringContext, ip int, def *runtime.StructDefinitionValue, fieldIndex int) {
	if ctx == nil || ip < 0 || def == nil || fieldIndex < 0 {
		return
	}
	plan := bytecodeNamedStructMemberPlan{
		definition: def,
		fieldIndex: fieldIndex,
	}
	if ctx.namedStructMembers == nil {
		ctx.namedStructMembers = make(map[int]bytecodeNamedStructMemberPlan, 1)
	}
	ctx.namedStructMembers[ip] = plan
}
