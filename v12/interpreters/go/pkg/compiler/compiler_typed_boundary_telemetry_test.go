package compiler

import (
	"strings"
	"testing"
)

func TestCompilerTypedBoundaryTelemetryIsOptIn(t *testing.T) {
	source := `
struct Pair { left: i32, right: i32 }

fn main() -> void {
  pair := Pair { left: 1, right: 2 }
  print(pair.left + pair.right)
}
`

	baseline := compileNoFallbackExecSourceWithOptions(t, "ablec-typed-boundary-telemetry-baseline", source, Options{
		PackageName: "main",
		EmitMain:    true,
		EntryPath:   "main.able",
	})
	baselineSource := combinedGeneratedSource(baseline)
	if strings.Contains(baselineSource, "__able_typed_boundary_telemetry") || strings.Contains(baselineSource, typedBoundaryTelemetryEnv) {
		t.Fatalf("normal generated output must not contain typed-boundary telemetry")
	}

	telemetry := compileNoFallbackExecSourceWithOptions(t, "ablec-typed-boundary-telemetry-enabled", source, Options{
		PackageName:                "main",
		EmitMain:                   true,
		EntryPath:                  "main.able",
		EmitTypedBoundaryTelemetry: true,
	})
	telemetrySource := combinedGeneratedSource(telemetry)
	for _, expected := range []string{
		"__able_typed_boundary_telemetry_snapshot",
		"__able_typed_boundary_telemetry_reset",
		"__able_telemetry_typed_boundary_struct_from_runtime",
		"__able_telemetry_typed_boundary_array_from_runtime",
		"__able_telemetry_typed_boundary_array_to_runtime",
		"__able_telemetry_typed_boundary_control_from_error",
		"func __able_control_from_error_with_node(node ast.Node, err error) *__ableControl {\n\t__able_telemetry_typed_boundary_control_from_error()",
		`{"categories":`,
		`"generated_function":%q`,
		`"able_source":%q`,
		`"carrier":%q`,
		`"immediate_consumer":%q`,
		`"reason":%q`,
		`generatedFunction: "__able_struct_Pair_from"`,
		`carrier: "runtime.Value"`,
		`immediateConsumer: "*Pair"`,
		`case "channel handle":`,
		`ableSource: "<compiler-runtime>::channel-handle"`,
		`immediateConsumer: "int64 channel handle"`,
		`ableSource: "<compiler-runtime>::channel-capacity"`,
		`ableSource: "<compiler-runtime>::mutex-handle"`,
		typedBoundaryTelemetryEnv,
	} {
		if !strings.Contains(telemetrySource, expected) {
			t.Fatalf("typed-boundary telemetry output missing %q", expected)
		}
	}
}

func TestCompilerTypedBoundaryTelemetryAttributesConcreteInterfaceLiftViaRuntime(t *testing.T) {
	source := `
interface Matcher T for Self {
  fn matches(self: Self, value: T) -> bool
}

struct BeWithinMatcher {}

impl Matcher f64 for BeWithinMatcher {
  fn matches(self: Self, value: f64) -> bool { true }
}

fn accept(value: Matcher (Result f64)) -> bool {
  value.matches(1.0)
}

fn main() -> bool {
  accept(BeWithinMatcher {})
}
`

	baseline := compileNoFallbackExecSourceWithOptions(t, "ablec-interface-lift-baseline", source, Options{
		PackageName: "main",
		EmitMain:    true,
		EntryPath:   "main.able",
	})
	if strings.Contains(combinedGeneratedSource(baseline), "interface_lift_via_runtime") {
		t.Fatalf("normal generated output must not contain interface-lift telemetry")
	}

	telemetry := compileNoFallbackExecSourceWithOptions(t, "ablec-interface-lift-telemetry", source, Options{
		PackageName:                "main",
		EmitMain:                   true,
		EntryPath:                  "main.able",
		EmitTypedBoundaryTelemetry: true,
	})
	compiled := combinedGeneratedSource(telemetry)
	for _, expected := range []string{
		`category: "interface_lift_via_runtime"`,
		`carrier: "*BeWithinMatcher [BeWithinMatcher]"`,
		`immediateConsumer: "__able_iface_Matcher_Result_f64_ [Matcher<Result<f64>>]"`,
		`reason: "statically known concrete carrier boxed through runtime.Value to recover a lifted native interface"`,
		"__able_telemetry_typed_boundary_interface_lift_via_runtime()",
		"__able_struct_BeWithinMatcher_to(",
		"__able_iface_Matcher_Result_f64__from_value(",
	} {
		if !strings.Contains(compiled, expected) {
			t.Fatalf("interface-lift telemetry output missing %q", expected)
		}
	}
	marker := "__able_telemetry_typed_boundary_interface_lift_via_runtime()"
	if got := strings.Count(compiled, marker); got != 2 {
		t.Fatalf("expected one executable runtime-lift marker plus its definition; occurrences = %d", got)
	}
}
