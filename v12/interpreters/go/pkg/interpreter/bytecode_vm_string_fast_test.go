package interpreter

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CanonicalStringMethodsUseFastPathsFromSource(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_TRACE", "1")

	source := `
import able.core.interfaces.{Error}
import able.core.iteration.{IteratorEnd}
import able.text.string.{String, StringBuilder}

fn build_text() -> String {
  builder := StringBuilder.with_capacity(24)
  builder.push_string("able")
  builder.push_string("-bench")
  builder.finish() match {
    case text: String => text,
    case err: Error => { raise err }
  }
}

fn byte_sum(text: String) -> i32 {
  iter := text.bytes().iterator()
  total := 0
  loop {
    iter.next() match {
      case IteratorEnd {} => { break },
      case value: u8 => total = total + (value as i32)
    }
  }
  total
}

fn char_count(text: String) -> i32 {
  iter := text.chars().iterator()
  total := 0
  loop {
    iter.next() match {
      case IteratorEnd {} => { break },
      case value: char => {
        total = total + 1
      }
    }
  }
  total
}

fn main() -> i64 {
  text := build_text()
  total := text.len_bytes() as i64
  if text.contains("able") {
    total = total + 10_i64
  }
  replaced := text.replace("able", "ABLE")
  total = total + (replaced.len_bytes() as i64)
  if replaced.contains("ABLE") {
    total = total + 20_i64
  }
  total = total + (byte_sum(replaced) as i64)
  total = total + (char_count(replaced) as i64)
  total
}

main()
`

	program := mustLoadAbleProgramFromSource(t, source)
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
		t.Fatalf("canonical string trace fixture mismatch: got=%#v want=%#v", got, want)
	}

	snapshot := interp.BytecodeTrace(0)
	var foundContains bool
	var foundReplace bool
	var foundLenBytes bool
	var foundBytes bool
	var foundChars bool
	var foundByteIterNext bool
	var foundCharIterNext bool
	for _, entry := range snapshot.Entries {
		switch entry.Dispatch {
		case "string_contains_fast":
			foundContains = true
		case "string_replace_fast":
			foundReplace = true
		case "string_len_bytes_fast":
			foundLenBytes = true
		case "string_bytes_fast":
			foundBytes = true
		case "string_chars_fast":
			foundChars = true
		case "string_byte_iter_next_fast":
			foundByteIterNext = true
		case "string_char_iter_next_fast":
			foundCharIterNext = true
		}
	}
	if !foundContains || !foundReplace || !foundLenBytes || !foundBytes || !foundChars || !foundByteIterNext || !foundCharIterNext {
		t.Fatalf("expected canonical string fast-path trace entries, got %#v", snapshot.Entries)
	}
}

func TestBytecodeVM_CanonicalStringCharsFastPathFallsBackForInvalidUTF8(t *testing.T) {
	interp := mustLoadBytecodeInterpreterFromSource(t, `
import able.core.interfaces.{Error}
import able.text.string.{String, StringBuilder}

fn main() -> String {
  builder := StringBuilder.new()
  builder.push_string("ok")
  builder.finish() match {
    case text: String => text,
    case err: Error => { raise err }
  }
}

main()
`)
	stringDef, ok := interp.lookupStructDefinition("String")
	if !ok || stringDef == nil {
		t.Fatalf("String definition missing after stdlib setup")
	}
	value := &runtime.StructInstanceValue{
		Definition: stringDef,
		Fields: map[string]runtime.Value{
			"bytes":     interp.newU8ArrayValueFromBytes([]byte{0xff}),
			"len_bytes": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
	}

	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{value}
	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringChars,
		bytecodeInstruction{name: "chars", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("canonical String.chars invalid UTF-8 fallback failed: %v", err)
	}
	if handled {
		t.Fatalf("canonical String.chars fast path should fall back for invalid UTF-8")
	}
}

func mustLoadAbleProgramFromSource(t *testing.T, source string) *driver.Program {
	t.Helper()

	root := t.TempDir()
	entryPath := filepath.Join(root, "main.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write able source: %v", err)
	}
	searchPaths, err := buildExecSearchPaths(entryPath, root, fixtureManifest{})
	if err != nil {
		t.Fatalf("build search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("new loader: %v", err)
	}
	defer loader.Close()
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load able source: %v", err)
	}
	return program
}

func mustLoadBytecodeInterpreterFromSource(t *testing.T, source string) *Interpreter {
	t.Helper()

	program := mustLoadAbleProgramFromSource(t, source)
	interp := NewBytecode()
	if _, _, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{}); err != nil {
		t.Fatalf("bytecode evaluation failed: %v", err)
	}
	return interp
}
