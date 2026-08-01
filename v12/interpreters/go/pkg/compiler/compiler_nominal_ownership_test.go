package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNominalOwnershipProvesDirectInterfaceAndEmbeddedTransfers(t *testing.T) {
	source := `
struct State {
  value: i64
}

struct Step {
  state: State,
  observed: i64
}

methods State {
  fn advance(self: Self, delta: i64) -> State {
    State { value: self.value + delta }
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
    next := state.advance(delta * 2_i64)
    Step { state: next, observed: state.value }
  }
}

fn direct() -> State {
  state := State { value: 1_i64 }
  state = state.advance(2_i64)
  state
}

fn through_interface(updater: Updater) -> State {
  state := State { value: 3_i64 }
  step := updater.update(state, 4_i64)
  state = step.state
  state
}

fn main() -> void {
  print(direct().value)
  print(through_interface(Add {}).value)
  print(through_interface(Double {}).value)
}
`
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-ownership-positive", source, Options{
		CollectNominalOwnership: true,
	})
	report := result.NominalOwnership
	if report == nil {
		t.Fatal("expected nominal ownership report")
	}
	if report.Totals.EligibleTransfers < 2 {
		t.Fatalf("eligible transfers = %d, want at least direct and interface embedded\nsuccessors=%#v\nsites=%#v",
			report.Totals.EligibleTransfers, report.Successors, report.CallSites)
	}
	requireOwnershipSite(t, report, "direct", "state", "direct-replacement", "direct", true)
	site := requireOwnershipSite(t, report, "through_interface", "state", "embedded-field-replacement", "native-interface", true)
	if strings.Join(site.ResultPath, ".") != "state" || len(site.Targets) != 2 {
		t.Fatalf("interface ownership site = %+v, want two targets and result path state", site)
	}
}

func TestCompilerNominalOwnershipFailsClosedForRetainedCaptureAndConditionalReplacement(t *testing.T) {
	source := `
struct State {
  value: i64
}

fn advance(state: State) -> State {
  State { value: state.value + 1_i64 }
}

fn retain_and_advance(state: State, history: Array State) -> State {
  history.push(state)
  advance(state)
}

fn retained() -> State {
  state := State { value: 1_i64 }
  old := state
  state = advance(state)
  print(old.value)
  state
}

fn captured(history: Array State) -> State {
  state := State { value: 2_i64 }
  state = retain_and_advance(state, history)
  state
}

fn external_origin(state: State) -> State {
  state = advance(state)
  state
}

fn conditional(flag: bool) -> State {
  state := State { value: 3_i64 }
  candidate := advance(state)
  if flag {
    state = candidate
  }
  state
}

fn main() -> void {
  history: Array State := Array.new()
  print(retained().value)
  print(captured(history).value)
  print(external_origin(State { value: 9_i64 }).value)
  print(conditional(true).value)
}
`
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-ownership-negative", source, Options{
		CollectNominalOwnership: true,
	})
	report := result.NominalOwnership
	if report == nil {
		t.Fatal("expected nominal ownership report")
	}
	retained := requireOwnershipSite(t, report, "retained", "state", "direct-replacement", "direct", false)
	if !containsOwnershipBlocker(retained.Blockers, "retained-alias") {
		t.Fatalf("retained site blockers = %v, want retained-alias", retained.Blockers)
	}
	conditional := requireOwnershipSite(t, report, "conditional", "state", "embedded-field-replacement", "direct", false)
	if !containsOwnershipBlocker(conditional.Blockers, "conditional-or-nonstraight-replacement") {
		t.Fatalf("conditional site blockers = %v, want conditional blocker", conditional.Blockers)
	}
	external := requireOwnershipSite(t, report, "external_origin", "state", "direct-replacement", "direct", false)
	if !containsOwnershipBlocker(external.Blockers, "source-not-locally-fresh") {
		t.Fatalf("external-origin site blockers = %v, want source origin blocker", external.Blockers)
	}
	for _, site := range report.CallSites {
		if strings.HasSuffix(site.Caller, "captured") && site.Eligible {
			t.Fatalf("capturing callee must not produce an eligible transfer: %+v", site)
		}
	}
}

func TestCompilerNominalOwnershipIsOptInAndDoesNotChangeGeneratedGo(t *testing.T) {
	source := `
struct State { value: i64 }
fn advance(state: State) -> State { State { value: state.value + 1_i64 } }
fn main() -> void {
  state := State { value: 1_i64 }
  state = advance(state)
  print(state.value)
}
`
	candidate := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-ownership-on", source, Options{
		CollectNominalOwnership: true,
	})
	if candidate.NominalOwnership == nil {
		t.Fatal("expected opt-in nominal ownership report")
	}
	for name, generated := range candidate.Files {
		if strings.Contains(string(generated), "nominal_ownership") ||
			strings.Contains(string(generated), "ownership transfer") {
			t.Fatalf("ownership diagnostics leaked into generated file %q", name)
		}
	}
	baseline := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-ownership-off", source, Options{})
	if baseline.NominalOwnership != nil {
		t.Fatal("nominal ownership should not be collected by default")
	}
	for name, generated := range baseline.Files {
		if strings.Contains(string(generated), "nominal_ownership") ||
			strings.Contains(string(generated), "ownership transfer") {
			t.Fatalf("baseline generated file %q contains ownership diagnostics", name)
		}
	}
}

func requireOwnershipSite(
	t *testing.T,
	report *NominalOwnershipReport,
	callerSuffix, binding, replacement, dispatch string,
	eligible bool,
) NominalOwnershipCallSite {
	t.Helper()
	for _, site := range report.CallSites {
		if (site.Caller == callerSuffix || strings.HasSuffix(site.Caller, "."+callerSuffix)) &&
			site.SourceBinding == binding &&
			site.Replacement == replacement &&
			site.Dispatch == dispatch &&
			site.Eligible == eligible {
			return site
		}
	}
	t.Fatalf("missing ownership site caller=%s binding=%s replacement=%s dispatch=%s eligible=%t\n%+v",
		callerSuffix, binding, replacement, dispatch, eligible, report.CallSites)
	return NominalOwnershipCallSite{}
}

func containsOwnershipBlocker(blockers []string, want string) bool {
	for _, blocker := range blockers {
		if blocker == want {
			return true
		}
	}
	return false
}
