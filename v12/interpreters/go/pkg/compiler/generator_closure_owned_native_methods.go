package compiler

import (
	"bytes"
	"fmt"
	"strings"
)

var closureOwnedNativeMethodLocals = [...]string{
	"isReady",
	"register",
	"commit",
	"isDefault",
	"wakeFn",
	"cancelMethod",
}

func (g *generator) renderClosureOwnedNativeMethodField(
	buf *bytes.Buffer,
	field string,
	methodLocal string,
) {
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\tinst.Fields[%q] = %s\n", field, methodLocal)
		return
	}
	fmt.Fprintf(
		buf,
		"\tinst.Fields[%q] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: %s}\n",
		field,
		methodLocal,
	)
}

func (g *generator) receiverFreeClosureOwnedNativeMethods(source string) string {
	if !g.callableExecutionContextsEnabled() {
		return source
	}
	for _, methodLocal := range closureOwnedNativeMethodLocals {
		bound := fmt.Sprintf(
			"&runtime.NativeBoundMethodValue{Receiver: inst, Method: %s}",
			methodLocal,
		)
		source = strings.ReplaceAll(source, bound, methodLocal)
	}
	return source
}
