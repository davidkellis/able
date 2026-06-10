package bridge

import (
	"fmt"
	"strconv"

	"able/interpreter-go/pkg/runtime"
)

// stringifyPrimitive preserves the language's scalar display semantics in
// standalone binaries, where the generated program deliberately has no
// interpreter available for the general Stringify fallback.
func stringifyPrimitive(value runtime.Value) (string, bool) {
	switch typed := value.(type) {
	case runtime.StringValue:
		return typed.Val, true
	case runtime.BoolValue:
		return strconv.FormatBool(typed.Val), true
	case runtime.CharValue:
		return string(typed.Val), true
	case runtime.IntegerValue:
		return typed.String(), true
	case runtime.FloatValue:
		return fmt.Sprintf("%g", typed.Val), true
	case runtime.NilValue:
		return "nil", true
	default:
		return "", false
	}
}
