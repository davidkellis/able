package interpreter

import "able/interpreter-go/pkg/runtime"

// mergeFunctionLike inserts a function or overload into the bucket, merging with
// any existing function-like value.
func mergeFunctionLike(bucket map[string]runtime.Value, name string, fn runtime.Value) {
	if bucket == nil {
		return
	}
	if existing, ok := bucket[name]; ok {
		if merged, ok := runtime.MergeFunctionValues(existing, fn); ok {
			bucket[name] = merged
			return
		}
	}
	bucket[name] = fn
}

// firstFunction returns the first concrete function in a function-like value.
func firstFunction(fn runtime.Value) *runtime.FunctionValue {
	if fn == nil {
		return nil
	}
	overloads := runtime.FlattenFunctionOverloads(fn)
	if len(overloads) == 0 {
		return nil
	}
	return overloads[0]
}

// functionOverloads expands a function-like value into its concrete overload list.
func functionOverloads(fn runtime.Value) []*runtime.FunctionValue {
	if fn == nil {
		return nil
	}
	return runtime.FlattenFunctionOverloads(fn)
}

// functionOverloadsView avoids flattening allocations when the value already
// stores a stable overload slice with no nil entries.
func functionOverloadsView(fn runtime.Value) []*runtime.FunctionValue {
	switch v := fn.(type) {
	case *runtime.FunctionOverloadValue:
		if v == nil || len(v.Overloads) == 0 {
			return nil
		}
		for _, entry := range v.Overloads {
			if entry == nil {
				return runtime.FlattenFunctionOverloads(fn)
			}
		}
		return v.Overloads
	default:
		return functionOverloads(fn)
	}
}
