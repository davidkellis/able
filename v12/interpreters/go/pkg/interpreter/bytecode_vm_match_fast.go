package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execJumpIfNotNil(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode jump-if-not-nil missing instruction")
	}
	val, err := vm.pop()
	if err != nil {
		return err
	}
	if !isNilRuntimeValue(val) {
		vm.ip = instr.target
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execJumpIfNotTypedPattern(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode jump-if-not-typed-pattern missing instruction")
	}
	val, err := vm.pop()
	if err != nil {
		return err
	}
	if exactDef, ok := bytecodeInstructionExactStructDefinition(instr); ok {
		coerced, matched := bytecodeMatchExactStructDefinition(exactDef, val)
		if !matched {
			vm.ip = instr.target
			return nil
		}
		vm.appendTypedPatternMatchValue(coerced)
		vm.ip++
		return nil
	}
	if coerced, matched, decided := vm.bytecodeMatchExactGenericStructPatternInstruction(instr, val); decided {
		if !matched {
			vm.ip = instr.target
			return nil
		}
		vm.appendTypedPatternMatchValue(coerced)
		vm.ip++
		return nil
	}
	if coerced, matched, decided := bytecodeMatchSimpleTypedPattern(vm, instr.typeSimpleCheck, val); decided {
		if !matched {
			vm.ip = instr.target
			return nil
		}
		vm.appendTypedPatternMatchValue(coerced)
		vm.ip++
		return nil
	}
	if vm.interp != nil {
		if matched, decided := vm.interp.matchesTypeWithoutRuntimeValue(instr.typeExpr); decided {
			if !matched {
				vm.ip = instr.target
				return nil
			}
			vm.appendTypedPatternMatchSnapshot(val)
			vm.ip++
			return nil
		}
	}
	coerced, ok := vm.interp.matchTypedPatternValueInEnv(instr.typeExpr, val, vm.env)
	if !ok {
		vm.ip = instr.target
		return nil
	}
	if coerced == nil {
		coerced = val
	}
	vm.appendTypedPatternMatchValue(coerced)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) appendTypedPatternMatchValue(value runtime.Value) {
	if bytecodeIsRawIntegerCarrier(value) {
		if kind, raw, ok := bytecodeRawIntegerValueInfo(value); ok {
			vm.appendRawIntegerStack(kind, raw)
			return
		}
	}
	vm.appendStackValue(value)
}

func (vm *bytecodeVM) appendTypedPatternMatchSnapshot(value runtime.Value) {
	if bytecodeIsRawIntegerCarrier(value) {
		if kind, raw, ok := bytecodeRawIntegerValueInfo(value); ok {
			vm.appendRawIntegerStack(kind, raw)
			return
		}
	}
	vm.appendStackValue(bytecodeStackSnapshotValue(value))
}

func bytecodeMatchSimpleTypedPattern(vm *bytecodeVM, check bytecodeSimpleTypeCheck, value runtime.Value) (runtime.Value, bool, bool) {
	if check == bytecodeSimpleTypeCheckUnknown {
		return nil, false, false
	}
	if check == bytecodeSimpleTypeCheckIteratorEnd {
		return bytecodeMatchSimpleIteratorEndTypedPattern(value)
	}
	if coerced, matched, decided := bytecodeMatchSimpleRawIntegerTypedPattern(check, value); decided {
		return coerced, matched, true
	}
	value = bytecodeSlotReadValue(value)
	switch check {
	case bytecodeSimpleTypeCheckAnyInteger:
		if _, _, ok := bytecodeRawIntegerValueInfo(value); ok {
			return bytecodeStackSnapshotValue(value), true, true
		}
		switch v := value.(type) {
		case runtime.IntegerValue:
			return v, true, true
		case *runtime.IntegerValue:
			if v != nil {
				return v, true, true
			}
		}
		return nil, false, true
	case bytecodeSimpleTypeCheckAnyFloat:
		if _, _, ok := bytecodeDirectRawFloatValue(value); ok {
			return bytecodeMaterializeRawFloatValue(value), true, true
		}
		switch v := value.(type) {
		case runtime.FloatValue:
			return v, true, true
		case *runtime.FloatValue:
			if v != nil {
				return v, true, true
			}
		case runtime.IntegerValue, *runtime.IntegerValue:
			return nil, false, false
		}
		return nil, false, true
	case bytecodeSimpleTypeCheckString:
		switch v := value.(type) {
		case runtime.StringValue:
			return v, true, true
		case *runtime.StringValue:
			if v != nil {
				return v, true, true
			}
		}
		return nil, false, false
	case bytecodeSimpleTypeCheckBool:
		switch v := value.(type) {
		case runtime.BoolValue:
			return v, true, true
		case *runtime.BoolValue:
			if v != nil {
				return v, true, true
			}
		}
		return nil, false, false
	}
	if targetKind, ok := check.integerType(); ok {
		if coerced, ok := bytecodeCoerceRawIntegerPatternValue(value, targetKind); ok {
			return coerced, true, true
		}
		if coerced, ok := coerceIntegerValueToTargetKindIfInRange(value, targetKind); ok {
			return coerced, true, true
		}
		return nil, false, true
	}
	if targetKind, ok := check.floatType(); ok {
		if raw, kind, ok := bytecodeDirectRawFloatValue(value); ok {
			if kind == targetKind {
				return bytecodeMaterializeRawFloatValue(value), true, true
			}
			return runtime.FloatValue{
				Val:        normalizeFloat(targetKind, raw),
				TypeSuffix: targetKind,
			}, true, true
		}
		if coerced, ok, err := inlineCoerceValueBySimpleType(string(targetKind), vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonPattern, value)); err != nil {
			return nil, false, true
		} else if ok {
			return coerced, true, true
		}
		return nil, false, true
	}
	return nil, false, false
}

func bytecodeMatchSimpleIteratorEndTypedPattern(value runtime.Value) (runtime.Value, bool, bool) {
	switch v := value.(type) {
	case runtime.IteratorEndValue:
		return v, true, true
	case *runtime.StructDefinitionValue:
		if v != nil && v.Node != nil && v.Node.ID != nil && isSingletonStructDef(v.Node) && v.Node.ID.Name == "IteratorEnd" {
			return value, true, true
		}
		return nil, false, true
	case runtime.StructDefinitionValue:
		if v.Node != nil && v.Node.ID != nil && isSingletonStructDef(v.Node) && v.Node.ID.Name == "IteratorEnd" {
			return value, true, true
		}
		return nil, false, true
	case *runtime.StructInstanceValue:
		if v != nil && v.Definition != nil && v.Definition.Node != nil && v.Definition.Node.ID != nil && v.Definition.Node.ID.Name == "IteratorEnd" {
			return value, true, true
		}
		return nil, false, true
	case *runtime.InterfaceValue:
		return nil, false, false
	default:
		if value != nil && value.Kind() == runtime.KindIteratorEnd {
			return value, true, true
		}
		return nil, false, true
	}
}

func bytecodeMatchSimpleRawIntegerTypedPattern(check bytecodeSimpleTypeCheck, value runtime.Value) (runtime.Value, bool, bool) {
	sourceKind, raw, ok := bytecodeRawIntegerValueInfo(value)
	if !ok || !bytecodeIsRawIntegerCarrier(value) {
		return nil, false, false
	}
	switch check {
	case bytecodeSimpleTypeCheckAnyInteger:
		return value, true, true
	case bytecodeSimpleTypeCheckAnyFloat,
		bytecodeSimpleTypeCheckString,
		bytecodeSimpleTypeCheckBool:
		return nil, false, false
	}
	targetKind, ok := check.integerType()
	if !ok {
		if _, ok := check.floatType(); ok {
			return nil, false, false
		}
		return nil, false, false
	}
	if sourceKind == targetKind {
		return value, true, true
	}
	if raw < 0 && (sourceKind == runtime.IntegerU64 || sourceKind == runtime.IntegerUsize) {
		return nil, false, true
	}
	if err := ensureFitsInt64Type(targetKind, raw); err != nil {
		return nil, false, true
	}
	return bytecodeRawIntegerResultValue(targetKind, raw), true, true
}

func bytecodeCoerceRawIntegerPatternValue(value runtime.Value, targetKind runtime.IntegerType) (runtime.Value, bool) {
	if !bytecodeIsRawIntegerCarrier(value) {
		return nil, false
	}
	sourceKind, raw, ok := bytecodeRawIntegerValueInfo(value)
	if !ok {
		return nil, false
	}
	if sourceKind == targetKind {
		return value, true
	}
	if raw < 0 && (sourceKind == runtime.IntegerU64 || sourceKind == runtime.IntegerUsize) {
		return nil, false
	}
	if err := ensureFitsInt64Type(targetKind, raw); err != nil {
		return nil, false
	}
	return bytecodeRawIntegerResultValue(targetKind, raw), true
}

func bytecodeInstructionExactStructDefinition(instr *bytecodeInstruction) (*runtime.StructDefinitionValue, bool) {
	if instr == nil || instr.value == nil {
		return nil, false
	}
	switch def := instr.value.(type) {
	case *runtime.StructDefinitionValue:
		return def, def != nil && def.Node != nil
	case runtime.StructDefinitionValue:
		if def.Node == nil {
			return nil, false
		}
		return &def, true
	default:
		return nil, false
	}
}

func bytecodeMatchExactStructDefinition(def *runtime.StructDefinitionValue, value runtime.Value) (runtime.Value, bool) {
	if def == nil || def.Node == nil {
		return nil, false
	}
	if bytecodeIsRawIntegerCarrier(value) {
		return nil, false
	}
	if _, _, ok := bytecodeDirectRawFloatValue(value); ok {
		return nil, false
	}
	value = bytecodeMaterializeRawValue(bytecodeSlotReadValue(value))
	switch v := value.(type) {
	case runtime.IteratorEndValue:
		if def.Node.ID != nil && def.Node.ID.Name == "IteratorEnd" {
			return value, true
		}
	case *runtime.IteratorEndValue:
		if v != nil && def.Node.ID != nil && def.Node.ID.Name == "IteratorEnd" {
			return value, true
		}
	case *runtime.StructInstanceValue:
		if v != nil && bytecodeSameStructDefinition(v.Definition, def) {
			return value, true
		}
	case *runtime.StructDefinitionValue:
		if v != nil && isSingletonStructDef(def.Node) && bytecodeSameStructDefinition(v, def) {
			return value, true
		}
	case runtime.StructDefinitionValue:
		if isSingletonStructDef(def.Node) && bytecodeSameStructDefinition(&v, def) {
			return value, true
		}
	}
	return nil, false
}

type bytecodeGenericStructPatternPlan struct {
	baseName string
	args     []bytecodeGenericStructPatternArgPlan
}

type bytecodeGenericStructPatternArgPlan struct {
	expr            ast.TypeExpression
	skip            bool
	simpleName      string
	simplePrimitive bool
	simpleOpenParam bool
}

func bytecodeGenericStructPatternPlanForTypeExpr(typeExpr ast.TypeExpression) *bytecodeGenericStructPatternPlan {
	return bytecodeGenericStructPatternPlanForTypeExprWithDefinition(typeExpr, nil)
}

func bytecodeGenericStructPatternPlanForTypeExprWithDefinition(typeExpr ast.TypeExpression, def *runtime.StructDefinitionValue) *bytecodeGenericStructPatternPlan {
	generic, ok := typeExpr.(*ast.GenericTypeExpression)
	if !ok || generic == nil {
		return nil
	}
	base, ok := generic.Base.(*ast.SimpleTypeExpression)
	if !ok || base == nil || base.Name == nil || base.Name.Name == "" {
		return nil
	}
	plan := &bytecodeGenericStructPatternPlan{
		baseName: base.Name.Name,
		args:     make([]bytecodeGenericStructPatternArgPlan, len(generic.Arguments)),
	}
	paramNames := bytecodeGenericStructDefinitionParamNames(def)
	for idx, want := range generic.Arguments {
		arg := bytecodeGenericStructPatternArgPlan{expr: want}
		switch t := want.(type) {
		case nil:
			arg.skip = true
		case *ast.WildcardTypeExpression:
			arg.skip = true
		case *ast.SimpleTypeExpression:
			if t != nil && t.Name != nil {
				arg.simpleName = t.Name.Name
				arg.simplePrimitive = isPrimitiveName(arg.simpleName)
				if !arg.simplePrimitive && paramNames != nil {
					_, arg.simpleOpenParam = paramNames[arg.simpleName]
				}
			}
		}
		plan.args[idx] = arg
	}
	return plan
}

func bytecodeGenericStructDefinitionParamNames(def *runtime.StructDefinitionValue) map[string]struct{} {
	if def == nil || def.Node == nil || len(def.Node.GenericParams) == 0 {
		return nil
	}
	names := make(map[string]struct{}, len(def.Node.GenericParams))
	for _, param := range def.Node.GenericParams {
		if param == nil || param.Name == nil || param.Name.Name == "" {
			continue
		}
		names[param.Name.Name] = struct{}{}
	}
	return names
}

func (vm *bytecodeVM) bytecodeMatchExactGenericStructPatternInstruction(instr *bytecodeInstruction, value runtime.Value) (runtime.Value, bool, bool) {
	if instr == nil {
		return nil, false, false
	}
	if instr.genericStructMatch != nil {
		return vm.bytecodeMatchExactGenericStructPatternPlan(instr.genericStructMatch, value)
	}
	return vm.bytecodeMatchExactGenericStructPattern(instr.typeExpr, value)
}

func (vm *bytecodeVM) bytecodeMatchExactGenericStructPattern(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, bool, bool) {
	generic, ok := typeExpr.(*ast.GenericTypeExpression)
	if !ok || generic == nil {
		return nil, false, false
	}
	base, ok := generic.Base.(*ast.SimpleTypeExpression)
	if !ok || base == nil || base.Name == nil || base.Name.Name == "" {
		return nil, false, false
	}
	baseName := base.Name.Name
	inst, ok := vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonPattern, bytecodeSlotReadValue(value)).(*runtime.StructInstanceValue)
	if !ok || inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
		return nil, false, false
	}
	if inst.Definition.Node.ID.Name != baseName {
		return nil, false, false
	}
	return value, bytecodeGenericStructTypeArgumentsMatch(vm.interp, generic.Arguments, inst.TypeArguments, inst.Definition.Node.GenericParams), true
}

func (vm *bytecodeVM) bytecodeMatchExactGenericStructPatternPlan(plan *bytecodeGenericStructPatternPlan, value runtime.Value) (runtime.Value, bool, bool) {
	if plan == nil || plan.baseName == "" {
		return nil, false, false
	}
	inst, ok := vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonPattern, bytecodeSlotReadValue(value)).(*runtime.StructInstanceValue)
	if !ok || inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
		return nil, false, false
	}
	if inst.Definition.Node.ID.Name != plan.baseName {
		return nil, false, false
	}
	return value, bytecodeGenericStructTypeArgumentsMatchPlan(vm.interp, plan.args, inst.TypeArguments, inst.Definition.Node.GenericParams), true
}

func bytecodeGenericStructTypeArgumentsMatchPlan(interp *Interpreter, expected []bytecodeGenericStructPatternArgPlan, actual []ast.TypeExpression, params []*ast.GenericParameter) bool {
	if len(expected) == 0 || len(actual) == 0 {
		return true
	}
	if len(expected) != len(actual) {
		return false
	}
	for idx, want := range expected {
		if !bytecodeGenericStructTypeArgumentMatchesPlan(interp, want, actual[idx], params) {
			return false
		}
	}
	return true
}

func bytecodeGenericStructTypeArgumentMatchesPlan(interp *Interpreter, want bytecodeGenericStructPatternArgPlan, got ast.TypeExpression, params []*ast.GenericParameter) bool {
	if want.skip || want.expr == nil || got == nil {
		return true
	}
	if _, ok := got.(*ast.WildcardTypeExpression); ok {
		return true
	}
	if want.simpleOpenParam {
		return true
	}
	if want.simpleName != "" && !want.simplePrimitive && bytecodeGenericStructParamNameMatches(want.simpleName, params) {
		return true
	}
	if bytecodeGenericStructTypeArgumentIsOpen(interp, got, params) {
		return true
	}
	if want.simpleName != "" && !want.simplePrimitive {
		if interp == nil || !interp.isKnownTypeName(want.simpleName) {
			return true
		}
	}
	return typeExpressionsEqual(want.expr, got)
}

func bytecodeGenericStructTypeArgumentsMatch(interp *Interpreter, expected []ast.TypeExpression, actual []ast.TypeExpression, params []*ast.GenericParameter) bool {
	if len(expected) == 0 || len(actual) == 0 {
		return true
	}
	if len(expected) != len(actual) {
		return false
	}
	for idx, want := range expected {
		got := actual[idx]
		if want == nil || got == nil {
			continue
		}
		if _, ok := got.(*ast.WildcardTypeExpression); ok {
			continue
		}
		if _, ok := want.(*ast.WildcardTypeExpression); ok {
			continue
		}
		if bytecodeGenericStructTypeArgumentIsOpen(interp, got, params) || bytecodeGenericStructTypeArgumentIsOpen(interp, want, params) {
			continue
		}
		if !typeExpressionsEqual(want, got) {
			return false
		}
	}
	return true
}

func bytecodeGenericStructTypeArgumentIsOpen(interp *Interpreter, expr ast.TypeExpression, params []*ast.GenericParameter) bool {
	simple, ok := expr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return false
	}
	name := simple.Name.Name
	if name == "" || isPrimitiveName(name) {
		return false
	}
	if bytecodeGenericStructParamNameMatches(name, params) {
		return true
	}
	if interp != nil && interp.isKnownTypeName(name) {
		return false
	}
	return true
}

func bytecodeGenericStructParamNameMatches(name string, params []*ast.GenericParameter) bool {
	if name == "" {
		return false
	}
	for _, param := range params {
		if param != nil && param.Name != nil && param.Name.Name == name {
			return true
		}
	}
	return false
}

func bytecodeSameStructDefinition(left *runtime.StructDefinitionValue, right *runtime.StructDefinitionValue) bool {
	if left == nil || right == nil || left.Node == nil || right.Node == nil {
		return false
	}
	if left == right || left.Node == right.Node {
		return true
	}
	if left.Node.ID == nil || right.Node.ID == nil {
		return false
	}
	return left.Node.ID.Name != "" && left.Node.ID.Name == right.Node.ID.Name
}
