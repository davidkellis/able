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
	result, err := vm.interp.assignPatternExpression(pattern, val, vm.env, op)
	if err != nil {
		return vm.interp.attachRuntimeContext(err, assignExpr, vm.interp.stateFromEnv(vm.env))
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, result)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execBindPattern(instr bytecodeInstruction) (bool, error) {
	var pattern ast.Pattern
	contextNode := instr.node
	isForLoop := false
	switch node := instr.node.(type) {
	case ast.Pattern:
		pattern = node
	case *ast.ForLoop:
		if node != nil {
			pattern = node.Pattern
			contextNode = node
			isForLoop = true
		}
	}
	if pattern == nil {
		return false, fmt.Errorf("bytecode bind pattern expects pattern node")
	}
	val, err := vm.pop()
	if err != nil {
		return false, err
	}
	if isForLoop {
		assigned, err := vm.interp.assignPatternForLoop(pattern, val, vm.env)
		if err != nil {
			err = vm.interp.attachRuntimeContext(err, contextNode, vm.interp.stateFromEnv(vm.env))
			if vm.handleLoopSignal(err) {
				return true, nil
			}
			return false, err
		}
		if errVal, ok := asErrorValue(assigned); ok {
			if len(vm.loopStack) == 0 {
				return false, fmt.Errorf("bytecode loop frame missing for pattern mismatch")
			}
			frame := vm.loopStack[len(vm.loopStack)-1]
			vm.env = frame.env
			vm.stack = append(vm.stack, errVal)
			vm.ip = frame.breakTarget
			return true, nil
		}
		vm.ip++
		return true, nil
	}
	if err := vm.interp.assignPattern(pattern, val, vm.env, true, nil); err != nil {
		err = vm.interp.attachRuntimeContext(err, contextNode, vm.interp.stateFromEnv(vm.env))
		if vm.handleLoopSignal(err) {
			return true, nil
		}
		return false, err
	}
	vm.ip++
	return true, nil
}

func (vm *bytecodeVM) execCompoundAssignSlot(instr bytecodeInstruction) error {
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
	current := vm.slots[instr.target]
	computed, err := applyBinaryOperator(vm.interp, binaryOp, current, val)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	vm.slots[instr.target] = computed
	if instr.target == 0 {
		vm.setSelfFastSlot0I32Value(computed)
	}
	if computed == nil {
		computed = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, computed)
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
