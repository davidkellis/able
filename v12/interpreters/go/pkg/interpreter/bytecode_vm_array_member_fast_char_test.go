package interpreter

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func monoCharArrayValueForTest(t *testing.T, values ...rune) *runtime.ArrayValue {
	t.Helper()
	handle := runtime.ArrayStoreMonoNewWithCapacityChar(len(values))
	for idx, value := range values {
		if err := runtime.ArrayStoreMonoWriteChar(handle, idx, value); err != nil {
			t.Fatalf("write mono char value %d: %v", idx, err)
		}
	}
	return &runtime.ArrayValue{Handle: handle, TrackedHandle: handle}
}

func TestBytecodeVM_ArrayMemberFastPathMonoCharGetSkipsBoxedState(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	arr := monoCharArrayValueForTest(t, 'a', 'b')
	vm := newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}

	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayGet,
		bytecodeInstruction{name: "get", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("mono char array get fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected mono char array get fast path to handle call")
	}
	charVal, ok := vm.stack[0].(runtime.CharValue)
	if !ok || charVal.Val != 'b' {
		t.Fatalf("mono char get result = %#v, want char 'b'", vm.stack[0])
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono char array get should not materialize boxed state")
	}
}

func TestBytecodeVM_ArrayPushMemberFastAppendsMonoCharWithoutMaterializingState(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	arr := monoCharArrayValueForTest(t)
	vm := newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr, runtime.CharValue{Val: 'x'}}

	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayPush,
		bytecodeInstruction{name: "push", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("mono char array push fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected mono char array push fast path to handle call")
	}
	if _, ok := vm.stack[0].(runtime.VoidValue); !ok {
		t.Fatalf("push result = %#v, want void", vm.stack[0])
	}
	raw, ok, err := runtime.ArrayStoreMonoReadCharIfAvailable(arr.Handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadCharIfAvailable: %v", err)
	}
	if !ok || raw != 'x' {
		t.Fatalf("mono char push read = (%q, %v), want ('x', true)", raw, ok)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono char push should not materialize boxed state")
	}
}

func TestBytecodeVM_ArrayReadSlotMemberFastReadsMonoCharWithoutMaterializingState(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	arr := monoCharArrayValueForTest(t, 'm', 'n')
	vm := newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(0, runtime.IntegerI32)}

	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayReadSlot,
		bytecodeInstruction{name: "read_slot", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("mono char read_slot fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected mono char read_slot fast path to handle call")
	}
	charVal, ok := vm.stack[0].(runtime.CharValue)
	if !ok || charVal.Val != 'm' {
		t.Fatalf("mono char read_slot result = %#v, want char 'm'", vm.stack[0])
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono char read_slot should not materialize boxed state")
	}
}

func TestBytecodeVM_ExactNativeArrayReadFastReadsMonoCharWithoutMaterializingState(t *testing.T) {
	interp := NewBytecode()
	arr := monoCharArrayValueForTest(t, 'r', 's')
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	result, handled, err := vm.tryExecExactNativeArrayReadFast("__able_array_read", []runtime.Value{
		runtime.NewSmallInt(arr.Handle, runtime.IntegerI64),
		runtime.NewSmallInt(1, runtime.IntegerI32),
	})
	if err != nil {
		t.Fatalf("exact native array read fast failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected exact native array read fast path to handle call")
	}
	charVal, ok := result.(runtime.CharValue)
	if !ok || charVal.Val != 's' {
		t.Fatalf("exact native array read result = %#v, want char 's'", result)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("exact native array read should not materialize boxed state")
	}
}

func TestBytecodeVM_ZigzagCharShapeUsesMonoCharFastPathsFromSource(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_TRACE", "1")
	root := t.TempDir()
	entryPath := filepath.Join(root, "main.able")
	source := `
import able.kernel.{Array}
import able.collections.array

fn build_input(length: i32) -> Array char {
  pattern: Array char = ['A', 'B', 'L', 'E']
  pattern_len := pattern.len()
  out: Array char = Array.with_capacity(length)
  i := 0
  loop {
    if i >= length { break }
    out.push(pattern[(i % pattern_len) as i32]!)
    i = i + 1
  }
  out
}

fn zigzag(chars: Array char, rows: i32) -> !Array char {
  buckets: Array (Array char) = Array.with_capacity(rows)
  r := 0
  loop {
    if r >= rows { break }
    buckets.push(Array.new())
    r = r + 1
  }

  current_row := 0
  direction := 1
  i := 0
  length := chars.len()
  loop {
    if i >= length { break }
    buckets[current_row]!.push(chars[i]!)
    current_row = current_row + direction
    if current_row == rows {
      current_row = rows - 2
      direction = -1
    } elsif current_row < 0 {
      current_row = 1
      direction = 1
    }
    i = i + 1
  }

  result: Array char = Array.new()
  i = 0
  count := buckets.len()
  loop {
    if i >= count { break }
    row := buckets[i]!
    j := 0
    row_len := row.len()
    loop {
      if j >= row_len { break }
      result.push(row[j]!)
      j = j + 1
    }
    i = i + 1
  }
  result
}

fn main() -> i32 {
  current := build_input(64)
  total := 0
  iter := 0
  loop {
    if iter >= 4 { break }
    current = zigzag(current, 5)!
    total = total + current.len()
    iter = iter + 1
  }
  total
}

main()
`
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write zigzag trace fixture: %v", err)
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
		t.Fatalf("load zigzag trace fixture: %v", err)
	}

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
		t.Fatalf("zigzag char trace fixture mismatch: got=%#v want=%#v", got, want)
	}

	snapshot := interp.BytecodeTrace(0)
	var foundCharPush bool
	var foundNativeArrayRead bool
	for _, entry := range snapshot.Entries {
		if entry.Dispatch == "array_push_char_mono_fast" {
			foundCharPush = true
		}
		if entry.Name == "__able_array_read" && entry.Dispatch == "exact_native" {
			foundNativeArrayRead = true
		}
	}
	if !foundCharPush {
		t.Fatalf("expected mono char push promotion trace entry, got %#v", snapshot.Entries)
	}
	if foundNativeArrayRead {
		t.Fatalf("expected direct char array indexing to avoid kernel array read fallback, got %#v", snapshot.Entries)
	}
}

func TestBytecodeVM_ZigzagCharReducedProfileTargetParity(t *testing.T) {
	root := repositoryRoot()
	if root == "" {
		t.Fatalf("repository root not found")
	}
	entryPath := filepath.Join(root, "v12", "interpreters", "go", "pkg", "interpreter", "testdata", "bytecode_runtime", "zigzag_char_reduced.able")
	searchPaths, err := buildExecSearchPaths(entryPath, filepath.Dir(entryPath), fixtureManifest{})
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
		t.Fatalf("load reduced zigzag profile target: %v", err)
	}

	want, _, _, err := New().EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("tree evaluation failed: %v", err)
	}
	got, _, _, err := NewBytecode().EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("bytecode evaluation failed: %v", err)
	}
	if !valuesEqual(got, want) {
		t.Fatalf("reduced zigzag profile target mismatch: got=%#v want=%#v", got, want)
	}
}
