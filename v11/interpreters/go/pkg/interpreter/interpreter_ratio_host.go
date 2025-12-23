package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) initRatioBuiltins() {
	if i.ratioReady {
		return
	}
	ratioFromFloat := runtime.NativeFunctionValue{
		Name:  "__able_ratio_from_float",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_ratio_from_float expects one argument")
			}
			val, ok := args[0].(runtime.FloatValue)
			if !ok {
				return nil, fmt.Errorf("__able_ratio_from_float expects float input")
			}
			parts, err := ratioFromFloatValue(val)
			if err != nil {
				return nil, err
			}
			return i.makeRatioValue(parts)
		},
	}
	i.global.Define("__able_ratio_from_float", ratioFromFloat)
	i.ratioReady = true
}
