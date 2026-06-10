package compiler

import (
	"bytes"
	"fmt"
)

const callPathTelemetryEnv = "ABLE_COMPILER_CALL_PATH_TELEMETRY"

// callPathTelemetryEnabled gates generated diagnostic instrumentation. It is
// deliberately distinct from dynamic-boundary telemetry: these counters are
// used to select CPU-profile candidates, not to claim a dynamic crossing.
func (g *generator) callPathTelemetryEnabled() bool {
	return g != nil && g.opts.EmitCallPathTelemetry
}

func (g *generator) renderCallPathTelemetryHelpers(buf *bytes.Buffer) {
	if !g.callPathTelemetryEnabled() {
		return
	}
	fmt.Fprintf(buf, "var __able_call_path_fast_method_calls int64\n")
	fmt.Fprintf(buf, "var __able_call_path_generic_union_method_calls int64\n")
	fmt.Fprintf(buf, "var __able_call_path_generic_union_fallbacks int64\n\n")
	fmt.Fprintf(buf, "func __able_telemetry_fast_method_call() { atomic.AddInt64(&__able_call_path_fast_method_calls, 1) }\n")
	fmt.Fprintf(buf, "func __able_telemetry_generic_union_method_call() { atomic.AddInt64(&__able_call_path_generic_union_method_calls, 1) }\n")
	fmt.Fprintf(buf, "func __able_telemetry_generic_union_fallback() { atomic.AddInt64(&__able_call_path_generic_union_fallbacks, 1) }\n\n")
	fmt.Fprintf(buf, "func __able_call_path_telemetry_snapshot() string {\n")
	fmt.Fprintf(buf, "\treturn fmt.Sprintf(`{\"fast_method_call\":%%d,\"generic_union_method_call\":%%d,\"generic_union_fallback\":%%d}`,\n")
	fmt.Fprintf(buf, "\t\tatomic.LoadInt64(&__able_call_path_fast_method_calls),\n")
	fmt.Fprintf(buf, "\t\tatomic.LoadInt64(&__able_call_path_generic_union_method_calls),\n")
	fmt.Fprintf(buf, "\t\tatomic.LoadInt64(&__able_call_path_generic_union_fallbacks),\n")
	fmt.Fprintf(buf, "\t)\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) emitCallPathTelemetry(buf *bytes.Buffer, category string, indent string) {
	if !g.callPathTelemetryEnabled() {
		return
	}
	switch category {
	case "fast-method":
		fmt.Fprintf(buf, "%s__able_telemetry_fast_method_call()\n", indent)
	case "generic-union":
		fmt.Fprintf(buf, "%s__able_telemetry_generic_union_method_call()\n", indent)
	case "generic-union-fallback":
		fmt.Fprintf(buf, "%s__able_telemetry_generic_union_fallback()\n", indent)
	}
}
