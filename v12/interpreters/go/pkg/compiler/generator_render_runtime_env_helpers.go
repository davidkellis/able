package compiler

import (
	"bytes"
	"fmt"
)

func writeRuntimeEnvSwapIfNeeded(buf *bytes.Buffer, indent string, runtimeExpr string, envExpr string, extraGuard string) {
	if buf == nil || runtimeExpr == "" || envExpr == "" {
		return
	}
	fmt.Fprintf(buf, "%sif %s != nil", indent, runtimeExpr)
	if extraGuard != "" {
		fmt.Fprintf(buf, " && %s", extraGuard)
	}
	fmt.Fprintf(buf, " && %s != nil {\n", envExpr)
	fmt.Fprintf(buf, "%s\tif __able_prev_env, __able_swapped_env := bridge.SwapEnvIfNeeded(%s, %s); __able_swapped_env {\n", indent, runtimeExpr, envExpr)
	fmt.Fprintf(buf, "%s\t\tdefer bridge.RestoreEnvIfNeeded(%s, __able_prev_env, __able_swapped_env)\n", indent, runtimeExpr)
	fmt.Fprintf(buf, "%s\t}\n", indent)
	fmt.Fprintf(buf, "%s}\n", indent)
}
