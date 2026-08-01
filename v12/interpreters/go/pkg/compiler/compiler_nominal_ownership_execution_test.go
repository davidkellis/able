package compiler

import (
	"strings"
	"testing"
)

func nominalOwnershipExecutionSource() string {
	return `
struct State {
  value: i64,
  check: i64
}

struct Step {
  state: State,
  observed: i64
}

methods State {
  fn advance(self: Self, delta: i64) -> State {
    State { value: self.value + delta, check: self.check + (delta * 3_i64) }
  }
}

interface Updater {
  fn update(self: Self, state: State, delta: i64) -> Step
}

struct Add {}
struct Double {}

impl Updater for Add {
  fn update(self: Self, state: State, delta: i64) -> Step {
    Step { state: state.advance(delta), observed: state.value }
  }
}

impl Updater for Double {
  fn update(self: Self, state: State, delta: i64) -> Step {
    Step { state: state.advance(delta * 2_i64), observed: state.value }
  }
}

fn direct() -> State {
  state := State { value: 1_i64, check: 10_i64 }
  state = state.advance(2_i64)
  state
}

fn through_interface(updater: Updater) -> State {
  state := State { value: 3_i64, check: 20_i64 }
  step := updater.update(state, 4_i64)
  print(step.observed)
  state = step.state
  state
}

fn main() -> void {
  first := direct()
  second := through_interface(Add {})
  third := through_interface(Double {})
  print(first.value)
  print(first.check)
  print(second.value)
  print(second.check)
  print(third.value)
  print(third.check)
}
`
}

func TestCompilerNominalOwnershipDefaultEmitsDirectAndEmbeddedInterfaceStoragePaths(t *testing.T) {
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-ownership-execution", nominalOwnershipExecutionSource(), Options{})
	compiled := combinedGeneratedSource(result)
	if !strings.Contains(compiled, "_owned(") {
		t.Fatal("expected opt-in caller-owned execution variants")
	}
	if !strings.Contains(compiled, "__able_owned_iface_Updater_update_State") {
		t.Fatal("expected typed ownership interface dispatch helper")
	}
	direct := mustCompiledFunctionBody(t, result, "__able_compiled_fn_direct")
	if !strings.Contains(direct, "__able_compiled_method_State_advance_owned(") {
		t.Fatalf("direct replacement did not select owned variant:\n%s", direct)
	}
	through := mustCompiledFunctionBody(t, result, "__able_compiled_fn_through_interface")
	if !strings.Contains(through, "__able_owned_iface_Updater_update_State(") {
		t.Fatalf("interface replacement did not select owned dispatch:\n%s", through)
	}
	compiled = combinedGeneratedSource(result)
	if !strings.Contains(compiled, "return &Step{State: __able_owned, Observed: __able_owned_outer.Observed}, nil") {
		t.Fatal("embedded ownership variant must reconstruct the outer result around owned storage")
	}

	baseline := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-ownership-execution-off", nominalOwnershipExecutionSource(), Options{
		DisableNominalOwnership: true,
	})
	if strings.Contains(combinedGeneratedSource(baseline), "__able_owned_iface_Updater_update_State") {
		t.Fatal("ownership execution helper must remain absent with the diagnostic opt-out")
	}
}

func TestCompilerNominalOwnershipDefaultExecutesDirectAndEmbeddedInterfacePaths(t *testing.T) {
	stdout := compileAndRunSourceWithOptions(t, "ablec-nominal-ownership-execution-run", nominalOwnershipExecutionSource(), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.Join(strings.Fields(stdout), " "); got != "3 3 3 16 7 32 11 44" {
		t.Fatalf("unexpected ownership execution output %q", stdout)
	}
}

func TestCompilerNominalOwnershipDefaultPreservesRetainedAndConditionalIdentities(t *testing.T) {
	source := `
struct State { value: i64, check: i64 }

fn advance(state: State) -> State {
  State { value: state.value + 1_i64, check: state.check + 3_i64 }
}

fn main() -> void {
  retained_state := State { value: 10_i64, check: 20_i64 }
  old := retained_state
  retained_state = advance(retained_state)
  print(old.value)
  print(old.check)

  conditional_state := State { value: 30_i64, check: 40_i64 }
  candidate := advance(conditional_state)
  if true { conditional_state = candidate }
  print(conditional_state.value)
  print(conditional_state.check)
}
`
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-ownership-identity-structure", source, Options{})
	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if strings.Contains(mainBody, "__able_compiled_fn_advance_owned(") {
		t.Fatalf("blocked replacements must stay on fresh-result calls:\n%s", mainBody)
	}
	stdout := compileAndRunSourceWithOptions(t, "ablec-nominal-ownership-identity-run", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.Join(strings.Fields(stdout), " "); got != "10 20 31 43" {
		t.Fatalf("ownership execution changed a retained or conditional identity: %q", stdout)
	}
}
