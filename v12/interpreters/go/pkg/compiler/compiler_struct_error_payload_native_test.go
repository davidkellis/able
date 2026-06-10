package compiler

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCompilerArrayHelperFixtureIndexErrorStaysBoundaryClean(t *testing.T) {
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	runCompilerNoBootstrapBoundaryAuditFixture(t, root, "06_12_02_stdlib_array_helpers")
}

func TestCompilerStructErrorPayloadHelpersUseSharedStructInstanceUnwrap(t *testing.T) {
	result := compileExecFixtureResult(t, "06_12_02_stdlib_array_helpers")

	tryFrom, ok := findCompiledFunction(result, "__able_struct_IndexError_try_from")
	if !ok {
		t.Fatalf("could not find __able_struct_IndexError_try_from")
	}
	if !strings.Contains(tryFrom, "inst := __able_struct_instance(current)") {
		t.Fatalf("expected IndexError try_from helper to unwrap error payloads through __able_struct_instance:\n%s", tryFrom)
	}
	if strings.Contains(tryFrom, "current.(*runtime.StructInstanceValue)") {
		t.Fatalf("expected IndexError try_from helper to avoid direct raw struct assertions:\n%s", tryFrom)
	}

	from, ok := findCompiledFunction(result, "__able_struct_IndexError_from")
	if !ok {
		t.Fatalf("could not find __able_struct_IndexError_from")
	}
	if !strings.Contains(from, "inst := __able_struct_instance(current)") {
		t.Fatalf("expected IndexError from helper to unwrap error payloads through __able_struct_instance:\n%s", from)
	}
	if strings.Contains(from, "current.(*runtime.StructInstanceValue)") {
		t.Fatalf("expected IndexError from helper to avoid direct raw struct assertions:\n%s", from)
	}

	toSeen, ok := findCompiledFunction(result, "__able_struct_IndexError_to_seen")
	if !ok {
		t.Fatalf("could not find __able_struct_IndexError_to_seen")
	}
	if !strings.Contains(toSeen, "runtime.NewStructInstancePositionalSized(def, ") {
		t.Fatalf("expected IndexError runtime encoding to use shared positional storage:\n%s", toSeen)
	}
	if strings.Contains(toSeen, "Fields: make(map[string]runtime.Value") {
		t.Fatalf("expected IndexError runtime encoding to avoid map-backed field storage:\n%s", toSeen)
	}
}

func TestCompilerPositionalErrorPayloadPreservesObservableSemantics(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"struct RootError { message: String }",
		"struct OuterError { message: String, cause: ?Error }",
		"",
		"impl Error for RootError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"impl Error for OuterError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { self.cause }",
		"}",
		"",
		"fn fail(root: bool) -> String {",
		"  if root {",
		"    raise RootError { message: \"root\" }",
		"  } else {",
		"    raise OuterError { message: \"outer\", cause: RootError { message: \"root\" } }",
		"  }",
		"  \"unreachable\"",
		"}",
		"",
		"fn relay(root: bool) -> String {",
		"  fail(root) rescue {",
		"    case err: Error => { raise err }",
		"  }",
		"}",
		"",
		"fn inspect(root: bool) -> String {",
		"  relay(root) rescue {",
		"    case err: Error => {",
		"      kind := err.value match {",
		"        case RootError { message::seen } => seen,",
		"        case OuterError { message::seen } => seen,",
		"        case _ => \"unknown\"",
		"      }",
		"      nested := err.cause() match {",
		"        case cause: Error => cause.message(),",
		"        case _ => \"none\"",
		"      }",
		"      `${kind}:${err.message()}:${nested}`",
		"    }",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  print(inspect(true))",
		"  print(inspect(false))",
		"}",
		"",
	}, "\n")

	result := compileNoFallbackSource(t, source)
	for _, helper := range []string{
		"__able_struct_RootError_to_seen",
		"__able_struct_OuterError_to_seen",
	} {
		body, ok := findCompiledFunction(result, helper)
		if !ok {
			t.Fatalf("could not find %s", helper)
		}
		if !strings.Contains(body, "runtime.NewStructInstancePositionalSized(def, ") {
			t.Fatalf("expected %s to use shared positional storage:\n%s", helper, body)
		}
	}

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-positional-error-payload-", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(stdout) != "root:root:none\nouter:outer:root" {
		t.Fatalf("unexpected positional error payload output %q", stdout)
	}
}
