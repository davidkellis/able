package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type fixedCallArg2 struct {
	values [2]runtime.Value
}

func callableAllowsEphemeralArgs(callee runtime.Value) bool {
	switch fn := callee.(type) {
	case *runtime.FunctionValue:
		return fn != nil
	case *runtime.FunctionOverloadValue:
		return fn != nil
	case runtime.NativeFunctionValue:
		return fn.BorrowArgs
	case *runtime.NativeFunctionValue:
		return fn != nil && fn.BorrowArgs
	case runtime.NativeBoundMethodValue:
		return fn.Method.BorrowArgs
	case *runtime.NativeBoundMethodValue:
		return fn != nil && fn.Method.BorrowArgs
	case runtime.BoundMethodValue:
		return callableAllowsEphemeralArgs(fn.Method)
	case *runtime.BoundMethodValue:
		return fn != nil && callableAllowsEphemeralArgs(fn.Method)
	default:
		return false
	}
}

func (i *Interpreter) acquireFixedCallArg2() *fixedCallArg2 {
	if i == nil {
		return &fixedCallArg2{}
	}
	raw := i.fixedCallArg2Pool.Get()
	args, _ := raw.(*fixedCallArg2)
	if args == nil {
		args = &fixedCallArg2{}
	}
	return args
}

func (i *Interpreter) releaseFixedCallArg2(args *fixedCallArg2) {
	if args == nil {
		return
	}
	clear(args.values[:])
	if i != nil {
		i.fixedCallArg2Pool.Put(args)
	}
}

func (i *Interpreter) callCallableValue2Mutable(callee runtime.Value, first runtime.Value, second runtime.Value, env *runtime.Environment, call *ast.FunctionCall) (runtime.Value, error) {
	if !callableAllowsEphemeralArgs(callee) {
		return i.callCallableValueMutable(callee, []runtime.Value{first, second}, env, call)
	}
	args := i.acquireFixedCallArg2()
	args.values[0] = first
	args.values[1] = second
	result, err := i.callCallableValueMutable(callee, args.values[:], env, call)
	i.releaseFixedCallArg2(args)
	return result, err
}
