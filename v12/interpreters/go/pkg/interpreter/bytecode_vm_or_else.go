package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execOrElse(instr bytecodeInstruction) (bool, error) {
	orElseExpr, ok := instr.node.(*ast.OrElseExpression)
	if !ok || orElseExpr == nil {
		return false, fmt.Errorf("bytecode or-else expects node")
	}
	val, err := vm.evalExpressionWithFallback(orElseExpr.Expression, vm.env)
	if err != nil {
		if rs, ok := err.(raiseSignal); ok {
			err = nil
			val = rs.value
		} else {
			if vm.handleLoopSignal(err) {
				return true, nil
			}
			return false, err
		}
	}
	failureKind := ""
	var failureValue runtime.Value
	if val == nil {
		failureKind = "nil"
	} else if val.Kind() == runtime.KindNil {
		failureKind = "nil"
	} else if errVal, ok := asErrorValue(val); ok {
		failureKind = "error"
		failureValue = errVal
	} else if vm.interp.matchesType(ast.Ty("Error"), val) {
		failureKind = "error"
		failureValue = val
	}
	if failureKind == "" {
		if val == nil {
			val = runtime.NilValue{}
		}
		vm.stack = append(vm.stack, val)
		vm.ip = instr.target
		return false, nil
	}
	vm.env = runtime.NewEnvironment(vm.env)
	if instr.name != "" && failureKind == "error" {
		vm.env.Define(instr.name, failureValue)
	}
	vm.ip++
	return false, nil
}
