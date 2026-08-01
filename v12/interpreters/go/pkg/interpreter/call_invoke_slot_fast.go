package interpreter

import "able/interpreter-go/pkg/runtime"

func ensureMutableCallArgs(args []runtime.Value, argsMutable bool) ([]runtime.Value, bool) {
	if argsMutable {
		return args, true
	}
	return append([]runtime.Value(nil), args...), true
}

func invokeFunctionBindArgsForSlotLayout(i *Interpreter, fn *runtime.FunctionValue, layout *bytecodeFrameLayout, bindArgs []runtime.Value, argsMutable bool) ([]runtime.Value, bool, error) {
	if layout == nil || len(bindArgs) == 0 || !layout.anyParamCoercion {
		return bindArgs, argsMutable, nil
	}
	limit := len(bindArgs)
	if layout.paramSlots < limit {
		limit = layout.paramSlots
	}
	for idx := 0; idx < limit; idx++ {
		if !inlineParamNeedsRuntimeCoercion(layout, idx, fn) {
			continue
		}
		arg := bindArgs[idx]
		paramType := inlineParamType(layout, idx)
		if inlineParamCoercionUnnecessary(i, layout, idx, paramType, arg) {
			continue
		}
		coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg)
		if err != nil {
			return nil, argsMutable, err
		}
		if !ok {
			coerced, err = i.coerceValueToTypeInEnv(paramType, arg, fn.Closure)
			if err != nil {
				return nil, argsMutable, err
			}
		}
		bindArgs, argsMutable = ensureMutableCallArgs(bindArgs, argsMutable)
		bindArgs[idx] = coerced
	}
	return bindArgs, argsMutable, nil
}

func slotLayoutUsesImplicitReceiver(layout *bytecodeFrameLayout, hasImplicit bool) bool {
	if !hasImplicit {
		return false
	}
	if layout == nil {
		return true
	}
	return layout.usesImplicitMember
}
