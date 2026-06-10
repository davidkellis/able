package compiler

import (
	"bytes"
	"fmt"
)

const dynamicBoundaryTelemetryEnv = "ABLE_COMPILER_DYNAMIC_BOUNDARY_TELEMETRY"

func (g *generator) dynamicBoundaryTelemetryEnabled() bool {
	return g != nil && g.opts.EmitDynamicBoundaryTelemetry
}

func (g *generator) renderDynamicBoundaryTelemetryHelpers(buf *bytes.Buffer) {
	if !g.dynamicBoundaryTelemetryEnabled() {
		return
	}
	fmt.Fprintf(buf, "var __able_dynamic_boundary_explicit_calls int64\n")
	fmt.Fprintf(buf, "var __able_dynamic_boundary_residual_polymorphic_calls int64\n")
	fmt.Fprintf(buf, "var __able_dynamic_boundary_host_abi_calls int64\n")
	fmt.Fprintf(buf, "var __able_dynamic_boundary_runtime_service_calls int64\n\n")
	fmt.Fprintf(buf, "func __able_telemetry_explicit_dynamic_call() { atomic.AddInt64(&__able_dynamic_boundary_explicit_calls, 1) }\n")
	fmt.Fprintf(buf, "func __able_telemetry_residual_polymorphic_call() { atomic.AddInt64(&__able_dynamic_boundary_residual_polymorphic_calls, 1) }\n")
	fmt.Fprintf(buf, "func __able_telemetry_host_abi_call() { atomic.AddInt64(&__able_dynamic_boundary_host_abi_calls, 1) }\n")
	fmt.Fprintf(buf, "func __able_telemetry_runtime_service_call() { atomic.AddInt64(&__able_dynamic_boundary_runtime_service_calls, 1) }\n\n")
	fmt.Fprintf(buf, "func __able_dynamic_boundary_telemetry_snapshot() string {\n")
	fmt.Fprintf(buf, "\treturn fmt.Sprintf(`{\"explicit_dynamic_call\":%%d,\"residual_polymorphic_call\":%%d,\"host_abi_call\":%%d,\"runtime_service_call\":%%d}`,\n")
	fmt.Fprintf(buf, "\t\tatomic.LoadInt64(&__able_dynamic_boundary_explicit_calls),\n")
	fmt.Fprintf(buf, "\t\tatomic.LoadInt64(&__able_dynamic_boundary_residual_polymorphic_calls),\n")
	fmt.Fprintf(buf, "\t\tatomic.LoadInt64(&__able_dynamic_boundary_host_abi_calls),\n")
	fmt.Fprintf(buf, "\t\tatomic.LoadInt64(&__able_dynamic_boundary_runtime_service_calls),\n")
	fmt.Fprintf(buf, "\t)\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) emitDynamicBoundaryTelemetry(buf *bytes.Buffer, category string) {
	if !g.dynamicBoundaryTelemetryEnabled() {
		return
	}
	switch category {
	case "explicit":
		fmt.Fprintf(buf, "\t__able_telemetry_explicit_dynamic_call()\n")
	case "residual":
		fmt.Fprintf(buf, "\t__able_telemetry_residual_polymorphic_call()\n")
	case "host":
		fmt.Fprintf(buf, "\t__able_telemetry_host_abi_call()\n")
	case "service":
		fmt.Fprintf(buf, "\t__able_telemetry_runtime_service_call()\n")
	}
}
