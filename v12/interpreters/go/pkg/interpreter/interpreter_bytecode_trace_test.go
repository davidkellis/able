package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestBytecodeTraceSnapshotSortsAndResets(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_TRACE", "1")
	interp := NewBytecode()
	node := ast.CallExpr(ast.ID("slow"))
	ast.SetSpan(node, ast.Span{Start: ast.Position{Line: 3, Column: 5}})

	interp.recordBytecodeCallTrace("call_name", "slow", "name", "generic", node)
	interp.recordBytecodeCallTrace("call_name", "slow", "name", "generic", node)
	interp.recordBytecodeCallTrace("call_member", "fast", "resolved_method", "inline", node)

	snapshot := interp.BytecodeTrace(1)
	if !snapshot.Enabled {
		t.Fatalf("expected bytecode trace to be enabled")
	}
	if snapshot.TotalHits != 3 {
		t.Fatalf("unexpected total hits: got %d want 3", snapshot.TotalHits)
	}
	if len(snapshot.Entries) != 1 {
		t.Fatalf("unexpected limited entry count: got %d want 1", len(snapshot.Entries))
	}
	if snapshot.Entries[0].Name != "slow" || snapshot.Entries[0].Hits != 2 {
		t.Fatalf("unexpected top entry: %#v", snapshot.Entries[0])
	}

	interp.ResetBytecodeTrace()
	if got := interp.BytecodeTrace(0); got.TotalHits != 0 || len(got.Entries) != 0 {
		t.Fatalf("expected trace reset, got %#v", got)
	}
}

func TestBytecodeTraceCapturesCallNameAndMemberSites(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_TRACE", "1")
	module := mustParseModuleSource(t, `
struct Box {
  value: i32
}

methods Box {
  fn next(self: Self) -> i32 {
    self.value + 1
  }
}

fn inc(x: i32) -> i32 {
  x + 1
}

fn main() -> i32 {
  value := Box { value: 1 }.next()
  inc(value)
}

main()
`)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode trace fixture mismatch: got=%#v want=%#v", got, want)
	}

	snapshot := interp.BytecodeTrace(0)
	if snapshot.TotalHits == 0 {
		t.Fatalf("expected bytecode trace hits, got %#v", snapshot)
	}
	var foundInc bool
	var foundNext bool
	for _, entry := range snapshot.Entries {
		if entry.Op == "call_name" && entry.Name == "inc" {
			foundInc = true
			if entry.Line == 0 || entry.Column == 0 {
				t.Fatalf("expected source location for call_name trace, got %#v", entry)
			}
		}
		if entry.Op == "call_member" && entry.Name == "next" {
			foundNext = true
			if entry.Line == 0 || entry.Column == 0 {
				t.Fatalf("expected source location for call_member trace, got %#v", entry)
			}
		}
	}
	if !foundInc {
		t.Fatalf("expected trace entry for inc call, got %#v", snapshot.Entries)
	}
	if !foundNext {
		t.Fatalf("expected trace entry for next call, got %#v", snapshot.Entries)
	}
}

func TestBytecodeTracePrimitiveMemberCallsUseResolvedMethodPath(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_TRACE", "1")
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	program := mustLoadAbleProgramFromSource(t, `
import able.core.interfaces.{Less}

fn is_less(left: i32, right: i32) -> bool {
  left.cmp(right) == Less
}

fn main() -> bool {
  is_less(1, 2) && is_less(1, 2)
}

main()
`)

	want, _, _, err := New().EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("tree evaluation failed: %v", err)
	}
	interp := NewBytecode()
	got, _, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("bytecode evaluation failed: %v", err)
	}
	if !valuesEqual(got, want) {
		t.Fatalf("primitive member-call trace mismatch: got=%#v want=%#v", got, want)
	}

	snapshot := interp.BytecodeTrace(0)
	var resolvedCmp bool
	for _, entry := range snapshot.Entries {
		if entry.Op != "call_member" || entry.Name != "cmp" {
			continue
		}
		if entry.Lookup == "member_access" {
			t.Fatalf("primitive cmp call should avoid member_access path, got %#v", entry)
		}
		if entry.Lookup == "resolved_method" {
			resolvedCmp = true
		}
	}
	if !resolvedCmp {
		t.Fatalf("expected resolved_method trace entry for primitive cmp call, got %#v", snapshot.Entries)
	}
}

func TestBytecodeTraceStaticMemberCallsUseResolvedMethodPath(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_TRACE", "1")

	module := mustParseModuleSource(t, `
struct Math

methods Math {
  fn twice(value: i32) -> i32 {
    value + value
  }
}

fn main() -> i32 {
  Math.twice(3) + Math.twice(4)
}

main()
`)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("static member-call trace mismatch: got=%#v want=%#v", got, want)
	}

	snapshot := interp.BytecodeTrace(0)
	var resolvedTwice bool
	for _, entry := range snapshot.Entries {
		if entry.Op != "call_member" || entry.Name != "twice" {
			continue
		}
		if entry.Lookup == "member_access" {
			t.Fatalf("static member call should avoid member_access path, got %#v", entry)
		}
		if entry.Lookup == "resolved_method" {
			resolvedTwice = true
		}
	}
	if !resolvedTwice {
		t.Fatalf("expected resolved_method trace entry for static member call, got %#v", snapshot.Entries)
	}
}
