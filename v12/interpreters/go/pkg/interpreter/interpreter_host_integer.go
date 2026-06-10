package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func hostIntegerValue(val runtime.Value) (runtime.IntegerValue, bool) {
	return bytecodeIntegerValue(val)
}

func hostIntegerToInt64(val runtime.Value) (int64, error) {
	iv, ok := hostIntegerValue(val)
	if !ok {
		return 0, fmt.Errorf("expected integer value")
	}
	if n, ok := iv.ToInt64(); ok {
		return n, nil
	}
	return 0, fmt.Errorf("integer out of range for i64")
}
