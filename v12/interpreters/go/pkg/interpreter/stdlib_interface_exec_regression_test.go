package interpreter

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func runStdlibExecSource(t *testing.T, source string, mode testExecMode) ([]string, error) {
	executor := NewSerialExecutor(nil)
	defer executor.Close()
	return runStdlibExecSourceWithExecutor(t, source, mode, executor)
}

func runStdlibExecSourceWithExecutor(t *testing.T, source string, mode testExecMode, executor Executor) ([]string, error) {
	t.Helper()

	root := t.TempDir()
	entryPath := filepath.Join(root, "main.able")
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	searchPaths, err := buildExecSearchPaths(entryPath, root, fixtureManifest{})
	if err != nil {
		t.Fatalf("build search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	interp := newTestInterpreter(t, mode, executor)
	typecheckMode := configureFixtureTypechecker(interp)
	var stdout []string
	registerPrint(interp, &stdout)

	entryEnv := interp.GlobalEnvironment()
	_, entryEnv, _, err = interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    typecheckMode == typecheckModeOff,
		AllowDiagnostics: typecheckMode != typecheckModeOff,
	})
	if err != nil {
		return stdout, err
	}

	mainValue, err := entryEnv.Get("main")
	if err != nil {
		return stdout, err
	}
	_, err = interp.CallFunction(mainValue, nil)
	return stdout, err
}

func TestPublicMutexAwaitLockSurvivesRepeatedGoroutineContentionInBothModes(t *testing.T) {
	source := `
package demo

import able.kernel.{Mutex}

fn main() -> void {
  rounds := 48_i64
  mutex := Mutex.new()
  commits := 0_i64

  first := spawn {
    subtotal := 0_i64
    index := 0_i64
    loop {
      if index >= rounds { break }
      committed := await [mutex.await_lock(fn() -> i64 {
        do {
          commits = commits + 1_i64
          index + 1_i64
        } ensure {
          mutex.unlock()
        }
      })]
      subtotal = subtotal + committed
      index = index + 1_i64
    }
    subtotal
  }
  second := spawn {
    subtotal := 0_i64
    index := 0_i64
    loop {
      if index >= rounds { break }
      committed := await [mutex.await_lock(fn() -> i64 {
        do {
          commits = commits + 1_i64
          index + 1_i64
        } ensure {
          mutex.unlock()
        }
      })]
      subtotal = subtotal + committed
      index = index + 1_i64
    }
    subtotal
  }

  print(first.value()! as i64)
  print(second.value()! as i64)
  print(commits)
}
`

	want := []string{"1176", "1176", "96"}
	for _, mode := range []testExecMode{testExecTreewalker, testExecBytecode} {
		t.Run(string(mode), func(t *testing.T) {
			got, err := runStdlibExecSourceWithExecutor(t, source, mode, NewGoroutineExecutor(nil))
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("stdout mismatch: got %v want %v", got, want)
			}
		})
	}
}

func TestStdlibLinkedListEnumerableMethodsExecuteInBothModes(t *testing.T) {
	source := `
package demo

import able.collections.linked_list.{LinkedList}

fn build(size: i32) -> LinkedList i32 {
  values: LinkedList i32 = LinkedList.new()
  i := 0
  loop {
    if i >= size { break }
    values.push_back(i)
    i = i + 1
  }
  values
}

fn main() -> void {
  values := build(8)
  mapped := values.map<i64>({ value => (value as i64) * 3_i64 })
  filtered := mapped.filter({ value => value >= 6_i64 })
  print(filtered.reduce<i64>(0_i64, { acc, value => acc + value }))
}
`

	want := []string{"81"}
	for _, mode := range []testExecMode{testExecTreewalker, testExecBytecode} {
		t.Run(string(mode), func(t *testing.T) {
			got, err := runStdlibExecSource(t, source, mode)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("stdout mismatch: got %v want %v", got, want)
			}
		})
	}
}

func TestStdlibLinkedListSatisfiesBaseIterableInterfaceInBothModes(t *testing.T) {
	source := `
package demo

import able.collections.linked_list.{LinkedList}
import able.core.iteration.{Iterable}

fn build(size: i32) -> LinkedList i32 {
  values: LinkedList i32 = LinkedList.new()
  i := 0
  loop {
    if i >= size { break }
    values.push_back(i)
    i = i + 1
  }
  values
}

fn checksum(values: Iterable i32) -> i64 {
  total: i64 = 0_i64
  for value in values {
    total = total + (value as i64)
  }
  total
}

fn main() -> void {
  values := build(4)
  view: Iterable i32 = values
  print(checksum(values))
  print(checksum(view))
}
`

	want := []string{"6", "6"}
	for _, mode := range []testExecMode{testExecTreewalker, testExecBytecode} {
		t.Run(string(mode), func(t *testing.T) {
			got, err := runStdlibExecSource(t, source, mode)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("stdout mismatch: got %v want %v", got, want)
			}
		})
	}
}

func TestStdlibLinkedListConcreteForLoopStillExecutesInBothModes(t *testing.T) {
	source := `
package demo

import able.collections.linked_list.{LinkedList}

fn build(size: i32) -> LinkedList i32 {
  values: LinkedList i32 = LinkedList.new()
  i := 0
  loop {
    if i >= size { break }
    values.push_back(i)
    i = i + 1
  }
  values
}

fn main() -> void {
  values := build(4)
  total: i64 = 0_i64
  for value in values {
    total = total + (value as i64)
  }
  print(total)
}
`

	want := []string{"6"}
	for _, mode := range []testExecMode{testExecTreewalker, testExecBytecode} {
		t.Run(string(mode), func(t *testing.T) {
			got, err := runStdlibExecSource(t, source, mode)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("stdout mismatch: got %v want %v", got, want)
			}
		})
	}
}

func TestImplMethodForwarderPrefersInherentMethodOnSameConcreteReceiver(t *testing.T) {
	source := `
package demo

import able.core.interfaces.{Add}

struct Box { value: i32 }

methods Box {
  fn add(self: Self, other: Box) -> Box {
    Box { value: self.value + other.value }
  }
}

impl Add Box Box for Box {
  fn add(self: Self, rhs: Box) -> Box { self.add(rhs) }
}

fn main() -> void {
  sum := Box { value: 1 } + Box { value: 2 }
  print(sum.value)
}
`

	want := []string{"3"}
	for _, mode := range []testExecMode{testExecTreewalker, testExecBytecode} {
		t.Run(string(mode), func(t *testing.T) {
			got, err := runStdlibExecSource(t, source, mode)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("stdout mismatch: got %v want %v", got, want)
			}
		})
	}
}

func TestOrdImplComparisonDoesNotDuplicateCandidatesAcrossInheritedLookup(t *testing.T) {
	source := `
package demo

import able.core.interfaces.{Ord, Ordering, Less, Equal, Greater}

struct Box { value: i32 }

impl Ord for Box {
  fn partial_cmp(self: Self, other: Self) -> Ordering { self.cmp(other) }
  fn cmp(self: Self, other: Self) -> Ordering {
    if self.value < other.value { return Less {} }
    if self.value > other.value { return Greater {} }
    Equal {}
  }
}

fn main() -> void {
  left := Box { value: 10 }
  right := Box { value: 3 }
  print(left > right)
  print(left >= right)
  print(left < right)
  print(left <= right)
}
`

	want := []string{"true", "true", "false", "false"}
	for _, mode := range []testExecMode{testExecTreewalker, testExecBytecode} {
		t.Run(string(mode), func(t *testing.T) {
			got, err := runStdlibExecSource(t, source, mode)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("stdout mismatch: got %v want %v", got, want)
			}
		})
	}
}

var _ runtime.Value
