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

// writeExecutionContextPackageEnv binds both the fixed execution context and
// the dynamic bridge to a package environment. Static same-package calls only
// pass the context pointer; the conditional bridge swap is needed when a
// compiled body later reaches an interpreter-backed dynamic boundary.
func writeExecutionContextPackageEnv(buf *bytes.Buffer, indent string, contextExpr string, runtimeExpr string, envExpr string) {
	if buf == nil || contextExpr == "" || envExpr == "" {
		return
	}
	localContext := contextExpr + "_package"
	fmt.Fprintf(buf, "%svar %s __able_execution_context\n", indent, localContext)
	fmt.Fprintf(buf, "%s%s = __able_context_with_environment(%s, %s, &%s)\n", indent, contextExpr, contextExpr, envExpr, localContext)
	writeRuntimeEnvSwapIfNeeded(buf, indent, runtimeExpr, envExpr, "")
}
