package interpreter

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	ableruntime "able/interpreter-go/pkg/runtime"
)

func TestBytecodeOperandStack_BoxedOperationsPreserveValues(t *testing.T) {
	vm := &bytecodeVM{}
	first := ableruntime.StringValue{Val: "first"}
	second := ableruntime.BoolValue{Val: true}
	rawFloat := bytecodeRawF64SlotValue(3.5)

	vm.appendStackValue(first)
	vm.appendStackPair(second, rawFloat)
	if got := vm.stackDepth(); got != 3 {
		t.Fatalf("stack depth = %d, want 3", got)
	}
	if got := vm.stackValue(0); got != first {
		t.Fatalf("first stack value = %#v, want %#v", got, first)
	}
	if got := vm.stackValues(1, 3); len(got) != 2 || got[0] != second || got[1] != rawFloat {
		t.Fatalf("stack range = %#v, want [%#v %#v]", got, second, rawFloat)
	}
	if _, ok := vm.stackValue(2).(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("raw float was materialized: %T", vm.stackValue(2))
	}

	vm.setStackValue(1, first)
	vm.truncateStack(2)
	got, err := vm.pop()
	if err != nil {
		t.Fatalf("pop: %v", err)
	}
	if got != first || vm.stackDepth() != 1 {
		t.Fatalf("pop = %#v, depth = %d; want %#v, 1", got, vm.stackDepth(), first)
	}
	vm.appendStackPair(second, rawFloat)
	if got := vm.stackValuesFrom(1); len(got) != 2 || got[0] != second || got[1] != rawFloat {
		t.Fatalf("stack suffix = %#v, want [%#v %#v]", got, second, rawFloat)
	}
	vm.clearStackFrom(1)
	if got := vm.stackValuesFrom(1); got[0] != nil || got[1] != nil {
		t.Fatalf("cleared stack suffix = %#v, want nil values", got)
	}
	vm.truncateStack(1)
}

func TestBytecodeOperandStack_CoreMigrationHasNoDirectAccess(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current source file")
	}
	dir := filepath.Dir(currentFile)
	files := []string{
		"bytecode_vm_call_frames.go",
		"bytecode_vm_calls.go",
		"bytecode_vm_controlflow.go",
		"bytecode_vm_ops.go",
		"bytecode_vm_return.go",
		"bytecode_vm_run.go",
	}
	assertBytecodeOperandStackFilesHaveNoDirectAccess(t, dir, files)
}

func TestBytecodeOperandStack_NumericSlotMigrationHasNoDirectAccess(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current source file")
	}
	dir := filepath.Dir(currentFile)
	patterns := []string{
		"bytecode_vm_assign.go",
		"bytecode_vm_cast*.go",
		"bytecode_vm_float*.go",
		"bytecode_vm_i32*.go",
		"bytecode_vm_int*.go",
		"bytecode_vm_integer*.go",
		"bytecode_vm_jump_ops.go",
		"bytecode_vm_literals.go",
		"bytecode_vm_slot*.go",
	}
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		for _, match := range matches {
			name := filepath.Base(match)
			if strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, "_bench_test.go") {
				continue
			}
			files = append(files, name)
		}
	}
	assertBytecodeOperandStackFilesHaveNoDirectAccess(t, dir, files)
}

func TestBytecodeOperandStack_AggregateMigrationHasNoDirectAccess(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current source file")
	}
	dir := filepath.Dir(currentFile)
	patterns := []string{
		"bytecode_vm_array*.go",
		"bytecode_vm_f64*.go",
		"bytecode_vm_index*.go",
		"bytecode_vm_iterator*.go",
		"bytecode_vm_match_fast.go",
		"bytecode_vm_members.go",
		"bytecode_vm_propagation.go",
		"bytecode_vm_string*.go",
		"bytecode_vm_struct_literal_fast.go",
	}
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		for _, match := range matches {
			name := filepath.Base(match)
			if strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, "_bench_test.go") {
				continue
			}
			files = append(files, name)
		}
	}
	assertBytecodeOperandStackFilesHaveNoDirectAccess(t, dir, files)
}

func TestBytecodeOperandStack_CallControlMigrationHasNoDirectAccess(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current source file")
	}
	dir := filepath.Dir(currentFile)
	patterns := []string{
		"bytecode_vm_call*.go",
		"bytecode_vm_static_member*.go",
		"bytecode_vm_ensure.go",
		"bytecode_vm_loops.go",
		"bytecode_vm_or_else.go",
		"bytecode_vm_rescue.go",
		"bytecode_vm_run_prepare.go",
	}
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		for _, match := range matches {
			name := filepath.Base(match)
			if strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, "_bench_test.go") {
				continue
			}
			files = append(files, name)
		}
	}
	assertBytecodeOperandStackFilesHaveNoDirectAccess(t, dir, files)
}

func TestBytecodeOperandStack_OnlyCentralImplementationAccessesStorage(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current source file")
	}
	dir := filepath.Dir(currentFile)
	matches, err := filepath.Glob(filepath.Join(dir, "bytecode_vm_*.go"))
	if err != nil {
		t.Fatalf("glob bytecode VM sources: %v", err)
	}
	var files []string
	for _, match := range matches {
		name := filepath.Base(match)
		if name == "bytecode_vm_stack.go" || strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, "_bench_test.go") {
			continue
		}
		files = append(files, name)
	}
	assertBytecodeOperandStackFilesHaveNoDirectAccess(t, dir, files)
}

func assertBytecodeOperandStackFilesHaveNoDirectAccess(t *testing.T, dir string, files []string) {
	t.Helper()
	forbidden := []string{
		"vm.stack[",
		"len(vm.stack)",
		"cap(vm.stack)",
		"clear(vm.stack)",
		"append(vm.stack,",
		"vm.stack =",
	}
	for _, name := range files {
		source, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, direct := range forbidden {
			if strings.Contains(string(source), direct) {
				t.Errorf("%s contains direct operand-stack access %q", name, direct)
			}
		}
	}
}
