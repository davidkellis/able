package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execRescue(instr bytecodeInstruction) error {
	rescueExpr, ok := instr.node.(*ast.RescueExpression)
	if !ok || rescueExpr == nil {
		return fmt.Errorf("bytecode rescue expects node")
	}
	val, err := vm.evalExpressionBytecode(rescueExpr.MonitoredExpression, vm.env)
	if err == nil {
		if val == nil {
			val = runtime.NilValue{}
		}
		vm.stack = append(vm.stack, val)
		vm.ip++
		return nil
	}
	rs, ok := err.(raiseSignal)
	if !ok {
		return err
	}
	for _, clause := range rescueExpr.Clauses {
		if clause == nil {
			continue
		}
		clauseEnv, matched := vm.interp.matchPattern(clause.Pattern, rs.value, vm.env)
		if !matched {
			continue
		}
		state := vm.interp.stateFromEnv(clauseEnv)
		state.pushRaise(rs.value)
		if clause.Guard != nil {
			guardVal, err := vm.evalExpressionBytecode(clause.Guard, clauseEnv)
			if err != nil {
				state.popRaise()
				return err
			}
			if !vm.interp.isTruthy(guardVal) {
				state.popRaise()
				continue
			}
		}
		val, bodyErr := vm.evalExpressionBytecode(clause.Body, clauseEnv)
		state.popRaise()
		if bodyErr != nil {
			return bodyErr
		}
		if val == nil {
			val = runtime.NilValue{}
		}
		vm.stack = append(vm.stack, val)
		vm.ip++
		return nil
	}
	return rs
}
