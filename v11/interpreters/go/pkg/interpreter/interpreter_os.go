package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) initOsBuiltins() {
	if i.osReady {
		return
	}

	osArgsFn := runtime.NativeFunctionValue{
		Name:  "__able_os_args",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("__able_os_args expects no arguments")
			}
			values := make([]runtime.Value, 0, len(i.osArgs))
			for _, arg := range i.osArgs {
				values = append(values, runtime.StringValue{Val: arg})
			}
			return i.newArrayValue(values, len(values)), nil
		},
	}

	osExitFn := runtime.NativeFunctionValue{
		Name:  "__able_os_exit",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_os_exit expects one argument")
			}
			code64, err := i.int64FromValue(args[0], "exit code")
			if err != nil {
				return nil, err
			}
			if code64 < 0 {
				return nil, fmt.Errorf("exit code must be non-negative")
			}
			if code64 > int64(^uint(0)>>1) {
				return nil, fmt.Errorf("exit code is out of range")
			}
			return nil, exitSignal{code: int(code64)}
		},
	}

	i.global.Define("__able_os_args", osArgsFn)
	i.global.Define("__able_os_exit", osExitFn)
	i.osReady = true
}
