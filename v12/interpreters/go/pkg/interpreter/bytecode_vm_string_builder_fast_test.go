package interpreter

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_StringBuilderMethodsUseFastPathsFromSource(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_TRACE", "1")

	root := t.TempDir()
	entryPath := filepath.Join(root, "main.able")
	source := `
import able.core.interfaces.{Error}
import able.text.string.{StringBuilder}

fn main() -> String {
  builder := StringBuilder.with_capacity(16)
  builder.push_string(` + "`ab`" + `)
  builder.push_char('c')
  builder.push_byte(100_u8)
  builder.finish() match {
    case text: String => text,
    case err: Error => { raise err }
  }
}

main()
`
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write string-builder trace fixture: %v", err)
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
		t.Fatalf("load string-builder trace fixture: %v", err)
	}

	interp := NewBytecode()
	got, _, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("bytecode evaluation failed: %v", err)
	}
	if text := mustStringValueText(t, got); text != "abcd" {
		t.Fatalf("string builder trace fixture result = %q, want %q", text, "abcd")
	}

	snapshot := interp.BytecodeTrace(0)
	var foundPushString bool
	var foundPushChar bool
	var foundPushByte bool
	var foundFinish bool
	for _, entry := range snapshot.Entries {
		switch entry.Dispatch {
		case "string_builder_push_string_fast":
			foundPushString = true
		case "string_builder_push_char_fast":
			foundPushChar = true
		case "string_builder_push_byte_fast":
			foundPushByte = true
		case "string_builder_finish_fast":
			foundFinish = true
		}
	}
	if !foundPushString || !foundPushChar || !foundPushByte || !foundFinish {
		t.Fatalf("expected string builder fast-path trace entries, got %#v", snapshot.Entries)
	}
}

func TestBytecodeVM_StringBuilderPushBytesAndAppendBuilderFastSemantics(t *testing.T) {
	interp, builderDef := setupCanonicalStringBuilderFast(t)

	builder := &runtime.StructInstanceValue{
		Definition: builderDef,
		Fields: map[string]runtime.Value{
			"buffer": interp.newU8ArrayValueFromString("ab"),
		},
	}
	sourceBytes := interp.newU8ArrayValueFromString("cd")

	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{builder, sourceBytes}
	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringBuilderPushBytes,
		bytecodeInstruction{name: "push_bytes", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("StringBuilder.push_bytes fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected StringBuilder.push_bytes fast path to handle call")
	}
	if got := mustStringBuilderBufferText(t, builder); got != "abcd" {
		t.Fatalf("buffer after push_bytes = %q, want %q", got, "abcd")
	}

	other := &runtime.StructInstanceValue{
		Definition: builderDef,
		Fields: map[string]runtime.Value{
			"buffer": interp.newU8ArrayValueFromString("ef"),
		},
	}
	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{builder, other}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringBuilderAppendBuilder,
		bytecodeInstruction{name: "append_builder", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("StringBuilder.append_builder fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected StringBuilder.append_builder fast path to handle call")
	}
	if got := mustStringBuilderBufferText(t, builder); got != "abcdef" {
		t.Fatalf("buffer after append_builder = %q, want %q", got, "abcdef")
	}
}

func TestBytecodeVM_StringBuilderFinishFastPathFallsBackForInvalidUTF8(t *testing.T) {
	interp, builderDef := setupCanonicalStringBuilderFast(t)
	builder := &runtime.StructInstanceValue{
		Definition: builderDef,
		Fields: map[string]runtime.Value{
			"buffer": interp.newU8ArrayValueFromBytes([]byte{0xff}),
		},
	}

	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{builder}
	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringBuilderFinish,
		bytecodeInstruction{name: "finish", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("StringBuilder.finish invalid UTF-8 fallback failed: %v", err)
	}
	if handled {
		t.Fatalf("StringBuilder.finish fast path should fall back for invalid UTF-8")
	}
}

func setupCanonicalStringBuilderFast(t *testing.T) (*Interpreter, *runtime.StructDefinitionValue) {
	t.Helper()

	interp := NewBytecode()
	module := mustParseModuleSource(t, `
import able.kernel.{Array}

struct StringBuilder {
  buffer: Array u8
}

methods StringBuilder {
  fn push_char(self: Self, value: char) -> void {}
  fn push_byte(self: Self, value: u8) -> void {}
  fn push_bytes(self: Self, bytes: Array u8) -> void {}
  fn push_string(self: Self, value: String) -> void {}
  fn append_builder(self: Self, other: StringBuilder) -> void {}
  fn finish(self: Self) -> !String { "" }
}
`)
	if _, _, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("evaluate StringBuilder setup: %v", err)
	}
	builderDef, found := interp.lookupStructDefinition("StringBuilder")
	if !found || builderDef == nil || builderDef.Node == nil {
		t.Fatalf("StringBuilder definition missing after setup")
	}

	origins := map[ast.Node]string{
		builderDef.Node: "/tmp/able-stdlib/src/text/string.able",
	}
	for _, stmt := range module.Body {
		methods, ok := stmt.(*ast.MethodsDefinition)
		if !ok || methods == nil {
			continue
		}
		for _, def := range methods.Definitions {
			if def != nil {
				origins[def] = "/tmp/able-stdlib/src/text/string.able"
			}
		}
	}
	interp.SetNodeOrigins(origins)
	return interp, builderDef
}

func mustStringBuilderBufferText(t *testing.T, builder *runtime.StructInstanceValue) string {
	t.Helper()

	arr, ok := builder.Fields["buffer"].(*runtime.ArrayValue)
	if !ok || arr == nil {
		t.Fatalf("StringBuilder buffer = %#v, want Array u8", builder.Fields["buffer"])
	}
	bytes, ok, err := runtime.ArrayStoreMonoBorrowedU8BytesIfAvailable(arr.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoBorrowedU8BytesIfAvailable: %v", err)
	}
	if !ok {
		t.Fatalf("expected mono u8 buffer handle for %#v", arr)
	}
	return string(bytes)
}

func mustStringValueText(t *testing.T, value runtime.Value) string {
	t.Helper()

	switch typed := value.(type) {
	case runtime.StringValue:
		return typed.Val
	case *runtime.StringValue:
		if typed == nil {
			t.Fatalf("string value is nil")
		}
		return typed.Val
	case *runtime.StructInstanceValue:
		if typed == nil || typed.Definition == nil || typed.Definition.Node == nil || typed.Definition.Node.ID == nil || typed.Definition.Node.ID.Name != "String" {
			t.Fatalf("value = %#v, want String", value)
		}
		var bytesValue runtime.Value
		if typed.Fields != nil {
			bytesValue = typed.Fields["bytes"]
		} else if len(typed.Positional) > 0 {
			bytesValue = typed.Positional[0]
		}
		arr, ok := bytesValue.(*runtime.ArrayValue)
		if !ok || arr == nil {
			t.Fatalf("String bytes = %#v, want Array u8", bytesValue)
		}
		bytes, ok, err := runtime.ArrayStoreMonoBorrowedU8BytesIfAvailable(arr.Handle)
		if err != nil {
			t.Fatalf("ArrayStoreMonoBorrowedU8BytesIfAvailable: %v", err)
		}
		if ok {
			return string(bytes)
		}
		size, err := runtime.ArrayStoreSize(arr.Handle)
		if err != nil {
			t.Fatalf("ArrayStoreSize: %v", err)
		}
		materialized := make([]byte, 0, size)
		for idx := 0; idx < size; idx++ {
			current, err := runtime.ArrayStoreRead(arr.Handle, idx)
			if err != nil {
				t.Fatalf("ArrayStoreRead(%d): %v", idx, err)
			}
			intVal, ok := current.(runtime.IntegerValue)
			if !ok || intVal.TypeSuffix != runtime.IntegerU8 {
				t.Fatalf("array element %d = %#v, want u8", idx, current)
			}
			raw, ok := intVal.ToInt64()
			if !ok || raw < 0 || raw > 255 {
				t.Fatalf("array element %d raw = %d/%v, want 0..255", idx, raw, ok)
			}
			materialized = append(materialized, byte(raw))
		}
		return string(materialized)
	default:
		t.Fatalf("value = %#v, want String", value)
		return ""
	}
}
