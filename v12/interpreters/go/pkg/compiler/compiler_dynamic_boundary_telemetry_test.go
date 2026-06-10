package compiler

import (
	"strings"
	"testing"
)

func TestCompilerDynamicBoundaryTelemetryIsOptIn(t *testing.T) {
	source := `
fn main() -> void {
  print("telemetry")
}
`

	baseline := compileNoFallbackExecSourceWithOptions(t, "ablec-boundary-telemetry-baseline", source, Options{
		PackageName: "main",
		EmitMain:    true,
		EntryPath:   "main.able",
	})
	baselineSource := combinedGeneratedSource(baseline)
	if strings.Contains(baselineSource, "__able_dynamic_boundary_telemetry") || strings.Contains(baselineSource, dynamicBoundaryTelemetryEnv) {
		t.Fatalf("normal generated output must not contain dynamic-boundary telemetry")
	}

	telemetry := compileNoFallbackExecSourceWithOptions(t, "ablec-boundary-telemetry-enabled", source, Options{
		PackageName:                  "main",
		EmitMain:                     true,
		EntryPath:                    "main.able",
		EmitDynamicBoundaryTelemetry: true,
	})
	telemetrySource := combinedGeneratedSource(telemetry)
	for _, expected := range []string{
		"__able_dynamic_boundary_telemetry_snapshot",
		"__able_telemetry_explicit_dynamic_call",
		"__able_telemetry_residual_polymorphic_call",
		"__able_telemetry_host_abi_call",
		"__able_telemetry_runtime_service_call",
		dynamicBoundaryTelemetryEnv,
	} {
		if !strings.Contains(telemetrySource, expected) {
			t.Fatalf("telemetry output missing %q", expected)
		}
	}
}
