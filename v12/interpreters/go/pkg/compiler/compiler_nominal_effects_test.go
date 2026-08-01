package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNominalEffectsPropagateDirectAndMonomorphicLambdaCalls(t *testing.T) {
	source := `
struct Record {
  value: i32
}

fn score(record: Record) -> i32 {
  record.value
}

fn score_twice(record: Record) -> i32 {
  score(record) + score(record)
}

fn main() -> void {
  scorer := { record => score_twice(record) }
  nested := { current => scorer(current) }
  print(nested(Record { value: 21_i32 }))
}
`
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-effects-positive", source, Options{
		CollectNominalEffects: true,
	})
	report := result.NominalEffects
	if report == nil {
		t.Fatal("expected nominal effect report")
	}
	for _, name := range []string{"score", "score_twice"} {
		parameter := requireNominalEffectParameter(t, report, name, "record")
		if !parameter.ReadOnlyNonEscaping || len(parameter.Blockers) != 0 {
			t.Fatalf("%s(record) effect = %+v, want read-only/non-escaping", name, parameter)
		}
	}
	if summary := requireNominalEffectCallable(t, report, "score"); summary.GeneratedGoName == "" {
		t.Fatal("expected generated Go name for score")
	}
	var safeNominalLambdas int
	for _, callable := range report.Callables {
		if callable.Kind != "lambda" {
			continue
		}
		for _, parameter := range callable.Parameters {
			if parameter.Nominal == "Record" && parameter.ReadOnlyNonEscaping {
				safeNominalLambdas++
			}
		}
	}
	if safeNominalLambdas != 2 {
		t.Fatalf("safe Record lambda parameters = %d, want 2\nreport: %s", safeNominalLambdas, report)
	}
}

func TestCompilerNominalEffectsFailClosedForMutationCaptureReturnAndUnknownCall(t *testing.T) {
	source := `
struct Record {
  value: i32
}

struct Holder {
  record: Record
}

fn mutate(record: Record) -> void {
  record.value = record.value + 1_i32
}

fn transitive_mutate(record: Record) -> void {
  mutate(record)
}

fn capture(record: Record) -> (() -> i32) {
  { => record.value }
}

fn return_alias(record: Record) -> Record {
  record
}

fn conditional_alias(record: Record, choose: bool) -> Record {
  if choose {
    record
  } else {
    record
  }
}

fn composite_capture(record: Record) -> Holder {
  Holder { record: record }
}

fn unknown_call(record: Record, callback: (Record -> i32)) -> i32 {
  callback(record)
}

fn store(record: Record, records: Array Record) -> void {
  records.write_slot(0, record)
}

fn main() -> void {
  record := Record { value: 1_i32 }
  records: Array Record := Array.with_capacity(1)
  records.push(record)
  store(record, records)
  transitive_mutate(record)
  reader := capture(record)
  print(reader())
  print(return_alias(record).value)
  print(conditional_alias(record, true).value)
  print(composite_capture(record).record.value)
  print(unknown_call(record, { current => current.value }))
}
`
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-effects-negative", source, Options{
		CollectNominalEffects: true,
	})
	report := result.NominalEffects
	if report == nil {
		t.Fatal("expected nominal effect report")
	}
	assertNominalEffect(t, report, "mutate", "record", nominalEffectMutation)
	assertNominalEffect(t, report, "transitive_mutate", "record", nominalEffectMutation)
	assertNominalEffect(t, report, "capture", "record", nominalEffectCapture)
	assertNominalEffect(t, report, "return_alias", "record", nominalEffectReturnAlias)
	assertNominalEffect(t, report, "conditional_alias", "record", nominalEffectReturnAlias)
	assertNominalEffect(t, report, "composite_capture", "record", nominalEffectCapture)
	assertNominalEffect(t, report, "unknown_call", "record", nominalEffectUnknownCall)
	assertNominalEffect(t, report, "store", "record", nominalEffectCapture)
}

func TestCompilerNominalEffectsRejectAmbiguousLocalCallableAndInterfaceDispatch(t *testing.T) {
	source := `
struct Record {
  value: i32
}

interface Consumer {
  fn consume(self: Self, record: Record) -> i32
}

struct Reader {}

impl Consumer for Reader {
  fn consume(self: Self, record: Record) -> i32 {
    record.value
  }
}

fn ambiguous(record: Record, choose: bool) -> i32 {
  if choose {
    callback := { current => current.value }
    return callback(record)
  }
  callback := { current => current.value + 1_i32 }
  callback(record)
}

fn interface_call(record: Record, consumer: Consumer) -> i32 {
  consumer.consume(record)
}

fn main() -> void {
  record := Record { value: 4_i32 }
  print(ambiguous(record, true))
  print(interface_call(record, Reader {}))
}
`
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-effects-ambiguous", source, Options{
		CollectNominalEffects: true,
	})
	report := result.NominalEffects
	if report == nil {
		t.Fatal("expected nominal effect report")
	}
	assertNominalEffect(t, report, "ambiguous", "record", nominalEffectUnknownCall)
	assertNominalEffect(t, report, "interface_call", "record", nominalEffectUnknownCall)
}

func TestCompilerNominalEffectsRemainOptInAndDoNotChangeGeneratedGo(t *testing.T) {
	source := `
struct Record {
  value: i32
}

fn score(record: Record) -> i32 {
  record.value
}

fn main() -> void {
  print(score(Record { value: 7_i32 }))
}
`
	candidate := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-effects-on", source, Options{
		CollectNominalEffects: true,
	})
	if candidate.NominalEffects == nil {
		t.Fatal("expected opt-in nominal effects")
	}
	for name, source := range candidate.Files {
		if strings.Contains(string(source), "nominal_effect") ||
			strings.Contains(string(source), "nominal effect") {
			t.Fatalf("diagnostic nominal effects leaked into generated file %q", name)
		}
	}
	baseline := compileNoFallbackExecSourceWithOptions(t, "ablec-nominal-effects-off", source, Options{})
	if baseline.NominalEffects != nil {
		t.Fatal("nominal effects should not be collected by default")
	}
}

func requireNominalEffectParameter(t *testing.T, report *NominalEffectReport, callableName, parameterName string) NominalParameterEffectSummary {
	t.Helper()
	for _, callable := range report.Callables {
		if callable.Callable != callableName &&
			!strings.HasSuffix(callable.Callable, "."+callableName) &&
			!strings.HasSuffix(callable.Callable, "::"+callableName) {
			continue
		}
		for _, parameter := range callable.Parameters {
			if parameter.Name == parameterName {
				return parameter
			}
		}
	}
	t.Fatalf("missing nominal effect for %s(%s)", callableName, parameterName)
	return NominalParameterEffectSummary{}
}

func requireNominalEffectCallable(t *testing.T, report *NominalEffectReport, callableName string) NominalCallableEffectSummary {
	t.Helper()
	for _, callable := range report.Callables {
		if callable.Callable == callableName ||
			strings.HasSuffix(callable.Callable, "."+callableName) ||
			strings.HasSuffix(callable.Callable, "::"+callableName) {
			return callable
		}
	}
	t.Fatalf("missing nominal effect callable %s", callableName)
	return NominalCallableEffectSummary{}
}

func assertNominalEffect(t *testing.T, report *NominalEffectReport, callableName, parameterName string, want nominalParameterEffect) {
	t.Helper()
	parameter := requireNominalEffectParameter(t, report, callableName, parameterName)
	got := nominalParameterEffect(0)
	if parameter.FieldOrIndexWrite {
		got |= nominalEffectMutation
	}
	if parameter.CapturedOrStored {
		got |= nominalEffectCapture
	}
	if parameter.ReturnAlias {
		got |= nominalEffectReturnAlias
	}
	if parameter.UnknownOrDynamicCall {
		got |= nominalEffectUnknownCall
	}
	if got&want == 0 {
		t.Fatalf("%s(%s) effect = %+v, want bit %d", callableName, parameterName, parameter, want)
	}
	if parameter.ReadOnlyNonEscaping {
		t.Fatalf("%s(%s) unexpectedly admitted: %+v", callableName, parameterName, parameter)
	}
}
