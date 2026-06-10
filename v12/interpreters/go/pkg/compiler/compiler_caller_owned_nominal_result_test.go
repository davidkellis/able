package compiler

import (
	"strings"
	"testing"
)

func callerOwnedNominalResultSource() string {
	return strings.Join([]string{
		"struct Pair {",
		"  left: i64,",
		"  right: i64",
		"}",
		"",
		"fn fresh(left: i64, right: i64) -> Pair {",
		"  Pair { left, right }",
		"}",
		"",
		"fn relay(left: i64, right: i64) -> Pair {",
		"  fresh(left, right)",
		"}",
		"",
		"fn existing(value: Pair) -> Pair { value }",
		"",
		"fn checked(ok: bool) -> Pair {",
		"  if !ok { raise(\"boom\") }",
		"  Pair { left: 21, right: 22 }",
		"}",
		"",
		"fn early(flag: bool) -> Pair {",
		"  if flag { return Pair { left: 1, right: 2 } }",
		"  Pair { left: 3, right: 4 }",
		"}",
		"",
		"fn main() {",
		"  first := relay(1, 2)",
		"  alias := first",
		"  alias.left = 9",
		"  print(first.left)",
		"",
		"  escaped := [relay(3, 4)]",
		"  item := escaped[0]",
		"  item.right = 12",
		"  print(escaped[0].right)",
		"",
		"  original := Pair { left: 7, right: 8 }",
		"  same := existing(original)",
		"  same.left = 15",
		"  print(original.left)",
		"  print(early(true).right)",
		"  recovered := checked(false) rescue {",
		"    case _ => Pair { left: 31, right: 32 }",
		"  }",
		"  print(recovered.left)",
		"}",
		"",
	}, "\n")
}

func TestCompilerCallerOwnedNominalResultIsStructuralAndPreservesAliases(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-caller-owned-result", callerOwnedNominalResultSource())
	compiled := combinedGeneratedSource(result)
	for _, function := range []string{"fresh", "relay", "checked"} {
		if !strings.Contains(compiled, "func __able_compiled_fn_"+function+"_into(") {
			t.Fatalf("expected structural caller-owned variant for %s", function)
		}
	}
	if strings.Contains(compiled, "func __able_compiled_fn_existing_into(") {
		t.Fatalf("returning an existing parameter must preserve its original pointer")
	}
	if strings.Contains(compiled, "func __able_compiled_fn_early_into(") {
		t.Fatalf("an explicit early nominal return must remain on the ordinary pointer ABI")
	}
	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(mainBody, "__able_compiled_fn_relay_into(") {
		t.Fatalf("expected static caller to provide result storage:\n%s", mainBody)
	}
	wrapper := mustCompiledFunctionBody(t, result, "__able_wrap_fn_relay")
	if strings.Contains(wrapper, "_into(") || !strings.Contains(wrapper, "__able_compiled_fn_relay(") {
		t.Fatalf("dynamic wrapper must retain the ordinary pointer-return ABI:\n%s", wrapper)
	}
}

func TestCompilerCallerOwnedNominalResultExecutesAliasAndEscapeCases(t *testing.T) {
	stdout := compileAndRunSourceWithOptions(t, "ablec-caller-owned-result-exec", callerOwnedNominalResultSource(), Options{
		RequireNoFallbacks: true,
		PackageName:        "main",
		EmitMain:           true,
	})
	if got := strings.Fields(stdout); strings.Join(got, " ") != "9 12 15 2 31" {
		t.Fatalf("unexpected caller-owned nominal result output %q", stdout)
	}
}

func TestCompilerLoopCarriedNominalKeepsRetainedOldResultDistinct(t *testing.T) {
	stdout := compileAndRunSourceWithOptions(t, "ablec-loop-carried-old-result", strings.Join([]string{
		"struct State { value: i64, check: i64 }",
		"",
		"fn advance(state: State) -> State {",
		"  State { value: state.value + 1, check: state.check + 3 }",
		"}",
		"",
		"fn main() {",
		"  state := State { value: 10, check: 20 }",
		"  retained := state",
		"  state = advance(state)",
		"  state = advance(state)",
		"  print(retained.value)",
		"  print(retained.check)",
		"  print(state.value)",
		"  print(state.check)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.Join(strings.Fields(stdout), " "); got != "10 20 12 26" {
		t.Fatalf("loop-carried nominal reuse changed retained identity: %q", stdout)
	}
}

func TestCompilerLoopCarriedNominalPreservesCalleeCaptureAndConditionalCandidate(t *testing.T) {
	source := strings.Join([]string{
		"struct State { value: i64, check: i64 }",
		"",
		"fn advance(state: State, delta: i64) -> State {",
		"  State { value: state.value + delta, check: state.check + (delta * 3) }",
		"}",
		"",
		"fn retain_and_advance(state: State, history: Array State) -> State {",
		"  history.push(state)",
		"  advance(state, 1)",
		"}",
		"",
		"fn main() {",
		"  history: Array State := Array.new()",
		"  state := State { value: 10, check: 20 }",
		"  i := 0",
		"  while i < 2 {",
		"    state = retain_and_advance(state, history)",
		"    i += 1",
		"  }",
		"  oldest := history.get(0)!",
		"  print(oldest.value)",
		"  print(oldest.check)",
		"  print(state.value)",
		"  print(state.check)",
		"",
		"  best := State { value: 100, check: 200 }",
		"  remembered := best",
		"  round := 0",
		"  while round < 3 {",
		"    candidate := advance(best, 1)",
		"    if round == 0 {",
		"      best = candidate",
		"      remembered = best",
		"    }",
		"    round += 1",
		"  }",
		"  print(remembered.value)",
		"  print(remembered.check)",
		"  print(best.value)",
		"  print(best.check)",
		"}",
		"",
	}, "\n")
	result := compileNoFallbackExecSource(t, "ablec-loop-carried-capture-boundaries", source)
	compiled := combinedGeneratedSource(result)
	for _, function := range []string{"advance", "retain_and_advance"} {
		if !strings.Contains(compiled, "func __able_compiled_fn_"+function+"_into(") {
			t.Fatalf("expected %s to exercise the caller-owned result family", function)
		}
	}
	stdout := compileAndRunSourceWithOptions(t, "ablec-loop-carried-capture-boundaries-exec", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.Join(strings.Fields(stdout), " "); got != "10 20 12 26 101 203 101 203" {
		t.Fatalf("loop-carried nominal reuse changed a captured or conditionally retained identity: %q", stdout)
	}
}
