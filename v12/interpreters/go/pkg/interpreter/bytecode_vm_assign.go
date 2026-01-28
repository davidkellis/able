package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execAssignPattern(instr bytecodeInstruction) error {
	assignExpr, ok := instr.node.(*ast.AssignmentExpression)
	if !ok || assignExpr == nil {
		return fmt.Errorf("bytecode assign pattern expects assignment expression")
	}
	pattern, ok := assignExpr.Left.(ast.Pattern)
	if !ok || pattern == nil {
		return fmt.Errorf("bytecode assign pattern expects pattern target")
	}
	val, err := vm.pop()
	if err != nil {
		return err
	}
	op := ast.AssignmentOperator(instr.operator)
	_, isCompound := binaryOpForAssignment(op)
	if isCompound {
		err := fmt.Errorf("compound assignment not supported with patterns")
		return vm.interp.attachRuntimeContext(err, assignExpr, vm.interp.stateFromEnv(vm.env))
	}
	switch op {
	case ast.AssignmentDeclare:
		newNames, hasAny := analyzePatternDeclarationNames(vm.env, pattern)
		if !hasAny || len(newNames) == 0 {
			err := fmt.Errorf(":= requires at least one new binding")
			return vm.interp.attachRuntimeContext(err, assignExpr, vm.interp.stateFromEnv(vm.env))
		}
		intent := &bindingIntent{declarationNames: newNames}
		if err := vm.interp.assignPattern(pattern, val, vm.env, true, intent); err != nil {
			return vm.interp.attachRuntimeContext(err, assignExpr, vm.interp.stateFromEnv(vm.env))
		}
	case ast.AssignmentAssign:
		intent := &bindingIntent{allowFallback: true}
		if err := vm.interp.assignPattern(pattern, val, vm.env, false, intent); err != nil {
			return vm.interp.attachRuntimeContext(err, assignExpr, vm.interp.stateFromEnv(vm.env))
		}
	default:
		err := fmt.Errorf("unsupported assignment operator %s", op)
		return vm.interp.attachRuntimeContext(err, assignExpr, vm.interp.stateFromEnv(vm.env))
	}
	if val == nil {
		val = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, val)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execAssignNameCompound(instr bytecodeInstruction) error {
	if instr.name == "" {
		return fmt.Errorf("bytecode compound assignment missing target name")
	}
	val, err := vm.pop()
	if err != nil {
		return err
	}
	op := ast.AssignmentOperator(instr.operator)
	binaryOp, isCompound := binaryOpForAssignment(op)
	if !isCompound {
		err := fmt.Errorf("unsupported assignment operator %s", op)
		if instr.node != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	current, err := vm.env.Get(instr.name)
	if err != nil {
		if instr.node != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	computed, err := applyBinaryOperator(vm.interp, binaryOp, current, val)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	if err := vm.env.Assign(instr.name, computed); err != nil {
		if instr.node != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	if computed == nil {
		computed = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, computed)
	vm.ip++
	return nil
}
