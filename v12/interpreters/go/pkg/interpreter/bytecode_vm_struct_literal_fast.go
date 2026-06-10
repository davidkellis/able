package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execStructLiteral(instr *bytecodeInstruction) error {
	lit, ok := instr.node.(*ast.StructLiteral)
	if !ok || lit == nil {
		return fmt.Errorf("bytecode struct literal expects node")
	}
	val, err := vm.interp.evaluateStructLiteral(lit, vm.env)
	if err != nil {
		return err
	}
	if val == nil {
		val = runtime.NilValue{}
	}
	if arr, isArray := val.(*runtime.ArrayValue); isArray {
		vm.trackBytecodeArrayOwnershipCreation(arr)
	}
	vm.appendStackValue(val)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execStructLiteralNamedFast(instr *bytecodeInstruction, program *bytecodeProgram) error {
	if vm == nil || instr == nil {
		return fmt.Errorf("bytecode named struct literal missing instruction")
	}
	lit, ok := instr.node.(*ast.StructLiteral)
	if !ok || lit == nil || lit.StructType == nil || lit.StructType.Name == "" {
		return fmt.Errorf("bytecode named struct literal expects node")
	}
	if !simpleNamedStructLiteralSyntacticEligible(lit) {
		return fmt.Errorf("bytecode named struct literal fast path requires simple named struct literal")
	}
	if instr.argCount < 0 || instr.argCount != len(lit.Fields) || instr.argCount > vm.stackDepth() {
		return fmt.Errorf("bytecode named struct literal stack underflow")
	}
	plan, planOK := bytecodeNamedStructLiteralPlanAt(program, vm.ip)
	structDefVal, err := vm.resolveNamedStructLiteralDefinition(lit.StructType.Name, plan)
	if err != nil {
		return err
	}
	if structDefVal == nil || structDefVal.Node == nil {
		return fmt.Errorf("struct definition '%s' unavailable", lit.StructType.Name)
	}
	if singletonValue, ok, err := vm.interp.singletonStructLiteralValueIfPossible(structDefVal, lit.StructType.Name, lit.TypeArguments); ok {
		if err != nil {
			return err
		}
		base := vm.stackDepth() - instr.argCount
		vm.truncateStack(base)
		vm.appendStackValue(singletonValue)
		vm.ip++
		return nil
	}
	if structDefVal.Node.Kind == ast.StructKindPositional {
		return fmt.Errorf("Named struct literal not allowed for positional struct '%s'", lit.StructType.Name)
	}
	if !simpleNamedStructLiteralDefinitionEligible(structDefVal.Node, lit.StructType.Name) {
		return fmt.Errorf("bytecode named struct literal fast path requires simple named struct '%s'", lit.StructType.Name)
	}
	base := vm.stackDepth() - instr.argCount
	fieldCount := len(structDefVal.Node.Fields)
	if lit.StructType.Name == "Array" {
		var inline [3]runtime.Value
		values := inline[:]
		if fieldCount > len(inline) {
			values = make([]runtime.Value, fieldCount)
		} else {
			values = values[:fieldCount]
		}
		if err := vm.fillStructLiteralNamedFastValues(values, lit, structDefVal, plan, planOK, base); err != nil {
			return err
		}
		typeArgs, err := vm.interp.resolveStructTypeArguments(structDefVal.Node, lit.TypeArguments, nil, nil, values)
		if err != nil {
			return err
		}
		if err := vm.interp.enforceStructConstraints(structDefVal.Node, typeArgs, lit.StructType.Name); err != nil {
			return err
		}
		arr, err := vm.interp.arrayValueFromStructFieldValues(structDefVal.Node.Fields, values)
		if err != nil {
			return err
		}
		vm.trackBytecodeArrayOwnershipCreation(arr)
		vm.truncateStack(base)
		vm.appendStackValue(arr)
		vm.ip++
		return nil
	}
	inst, values := runtime.NewStructInstancePositionalSized(structDefVal, len(structDefVal.Node.Fields), nil)
	if err := vm.fillStructLiteralNamedFastValues(values, lit, structDefVal, plan, planOK, base); err != nil {
		return err
	}
	typeArgs, err := vm.interp.resolveStructTypeArguments(structDefVal.Node, lit.TypeArguments, nil, nil, values)
	if err != nil {
		return err
	}
	if err := vm.interp.enforceStructConstraints(structDefVal.Node, typeArgs, lit.StructType.Name); err != nil {
		return err
	}
	vm.markBytecodeArrayOwnershipValuesEscaped(values, bytecodeArrayOwnershipEscapeAggregate)
	inst.TypeArguments = typeArgs
	vm.truncateStack(base)
	vm.appendStackValue(inst)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) fillStructLiteralNamedFastValues(values []runtime.Value, lit *ast.StructLiteral, structDefVal *runtime.StructDefinitionValue, plan bytecodeNamedStructLiteralPlan, planOK bool, base int) error {
	if planOK && len(plan.fieldOrder) == len(lit.Fields) {
		for idx, fieldIndex := range plan.fieldOrder {
			if fieldIndex < 0 || fieldIndex >= len(values) {
				return fmt.Errorf("Invalid field plan for struct '%s'", lit.StructType.Name)
			}
			// Struct fields outlive the current instruction and may be observed
			// after later inline calls. Raw VM scalar carriers are scratch values,
			// so stabilize them at this aggregate boundary.
			values[fieldIndex] = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonCollection, vm.stackValue(base+idx))
		}
	} else {
		runtimePlan, err := vm.interp.namedStructLiteralPlanCached(lit, structDefVal.Node)
		if err != nil {
			return err
		}
		for idx, fieldIndex := range runtimePlan.fieldOrder {
			values[fieldIndex] = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonCollection, vm.stackValue(base+idx))
		}
	}
	return nil
}

func bytecodeNamedStructLiteralPlanAt(program *bytecodeProgram, ip int) (bytecodeNamedStructLiteralPlan, bool) {
	if program == nil || ip < 0 || program.namedStructLiterals == nil {
		return bytecodeNamedStructLiteralPlan{}, false
	}
	plan, ok := program.namedStructLiterals[ip]
	return plan, ok
}

func (vm *bytecodeVM) resolveNamedStructLiteralDefinition(name string, plan bytecodeNamedStructLiteralPlan) (*runtime.StructDefinitionValue, error) {
	if plan.definition != nil {
		return plan.definition, nil
	}
	structDefVal, ok := vm.env.StructDefinition(name)
	if !ok {
		defValue, found := vm.env.Lookup(name)
		if !found {
			return nil, fmt.Errorf("Undefined variable '%s'", name)
		}
		var err error
		structDefVal, err = toStructDefinitionValue(defValue, name)
		if err != nil {
			return nil, err
		}
	}
	return structDefVal, nil
}
