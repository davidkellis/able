package compiler

import (
	"bytes"
	"fmt"
)

func (g *generator) renderRuntimeBuiltinHelpers(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "func __able_slice_len[T any](items []T) int {\n")
	fmt.Fprintf(buf, "\treturn len(items)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_slice_cap[T any](items []T) int {\n")
	fmt.Fprintf(buf, "\treturn cap(items)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_string_len_bytes(value string) int {\n")
	fmt.Fprintf(buf, "\treturn len(value)\n")
	fmt.Fprintf(buf, "}\n\n")
}
