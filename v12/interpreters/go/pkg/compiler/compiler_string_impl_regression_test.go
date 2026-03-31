package compiler

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestCompilerNoFallbacksStringDefaultImplStaticEmpty(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"interface Default {",
		"  fn default() -> Self",
		"}",
		"",
		"struct String {",
		"  n: i32",
		"}",
		"",
		"methods String {",
		"  fn empty() -> String { String { n: 0 } }",
		"}",
		"",
		"impl Default for String {",
		"  fn default() -> String { String.empty() }",
		"}",
		"",
		"fn main() -> void {}",
		"",
	}, "\n")
	entryPath := filepath.Join(root, "main.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	result, err := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile with no fallbacks: %v", err)
	}
	if len(result.Files) == 0 {
		t.Fatalf("expected generated output files")
	}
}

func TestCompilerNoFallbacksStringBuilderIteratorCloneSpecializesArraySelfReturn(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	entryPath := filepath.Join(root, "main.able")
	source := strings.Join([]string{
		"package demo",
		"import able.text.string.{StringBuilder}",
		"",
		"fn main() -> void {",
		"  builder := StringBuilder.new()",
		"  builder.iterator()",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(current), "..", "..", "..", "..", ".."))
	stdlibRoot := filepath.Join(filepath.Dir(repoRoot), "able-stdlib", "src")
	kernelRoot := filepath.Join(repoRoot, "v12", "kernel", "src")
	loader, err := driver.NewLoader([]driver.SearchPath{
		{Path: root, Kind: driver.RootUser},
		{Path: stdlibRoot, Kind: driver.RootStdlib},
		{Path: kernelRoot, Kind: driver.RootStdlib},
	})
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	result, err := New(Options{
		PackageName:            "main",
		RequireNoFallbacks:     true,
		ExperimentalMonoArrays: true,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile with no fallbacks: %v", err)
	}
	if len(result.Fallbacks) != 0 {
		t.Fatalf("expected no fallbacks, got %v", result.Fallbacks)
	}
	if len(result.Files) == 0 {
		t.Fatalf("expected generated output files")
	}
}

func TestCompilerNoFallbacksGraphemeCloneKeepsNativeArrayClone(t *testing.T) {
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-grapheme-clone-native", strings.Join([]string{
		"package demo",
		"import able.kernel.{Array}",
		"import able.collections.array",
		"import able.text.string.{Grapheme}",
		"",
		"fn main() -> u64 {",
		"  bytes: Array u8 = Array.new()",
		"  bytes.push(1)",
		"  grapheme := Grapheme { bytes: bytes }",
		"  grapheme.clone().len_bytes()",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		RequireNoFallbacks:     true,
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	graphemeClone := regexp.MustCompile(`func __able_compiled_impl_Clone_clone_[^(]*\(self \*Grapheme\)[\s\S]*?\n}`).FindString(compiledSrc)
	if graphemeClone == "" {
		t.Fatalf("expected compiled Grapheme.clone impl in output")
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_array_u8_to(__able_runtime, self.Bytes)",
		"__able_array_u8_from(",
	} {
		if strings.Contains(graphemeClone, fragment) {
			t.Fatalf("expected Grapheme.clone to keep Array<u8>.clone on the native static path without %q:\n%s", fragment, graphemeClone)
		}
	}
	for _, fragment := range []string{
		"type ArrayIterator_T struct {",
		"__able_iface_Iterator_u8_wrap_ptr_ArrayIterator(value *ArrayIterator)",
		"__able_iface_Iterator_char_wrap_ptr_ArrayIterator(value *ArrayIterator)",
		"__able_iface_Iterator_Grapheme_wrap_ptr_ArrayIterator(value *ArrayIterator)",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected Grapheme.clone to avoid malformed generic ArrayIterator specializations (%q):\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerNoFallbacksStringBuilderUsesNativeArrayPushAll(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	entryPath := filepath.Join(root, "main.able")
	source := strings.Join([]string{
		"package demo",
		"import able.core.interfaces.{Error}",
		"import able.text.string.{String, StringBuilder}",
		"",
		"fn make_string(value: String) -> String {",
		"  String.from_builtin(value) match {",
		"    case s: String => s,",
		"    case err: Error => { raise err }",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  builder := StringBuilder.new()",
		"  builder.push_string(make_string(\"Hello\"))",
		"  builder.push_string(make_string(\" World\"))",
		"  builder.finish() match {",
		"    case result: String => print(result),",
		"    case err: Error => { raise err }",
		"  }",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(current), "..", "..", "..", "..", ".."))
	stdlibRoot := filepath.Join(filepath.Dir(repoRoot), "able-stdlib", "src")
	kernelRoot := filepath.Join(repoRoot, "v12", "kernel", "src")
	loader, err := driver.NewLoader([]driver.SearchPath{
		{Path: root, Kind: driver.RootUser},
		{Path: stdlibRoot, Kind: driver.RootStdlib},
		{Path: kernelRoot, Kind: driver.RootStdlib},
	})
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	result, err := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile with no fallbacks: %v", err)
	}
	compiledSrc := string(result.Files["compiled.go"])
	pushStringBody := extractCompiledFunctionBody(compiledSrc, "method_StringBuilder_push_string")
	if pushStringBody == "" {
		t.Fatalf("StringBuilder.push_string body not found")
	}
	if strings.Contains(pushStringBody, "__able_method_call_node(") || strings.Contains(pushStringBody, "\"push_all\"") {
		t.Fatalf("expected StringBuilder.push_string to stay on native array lowering:\n%s", pushStringBody)
	}
	pushBytesBody := extractCompiledFunctionBody(compiledSrc, "method_StringBuilder_push_bytes")
	if pushBytesBody == "" {
		t.Fatalf("StringBuilder.push_bytes body not found")
	}
	if strings.Contains(pushBytesBody, "__able_method_call_node(") || strings.Contains(pushBytesBody, "\"push_all\"") {
		t.Fatalf("expected StringBuilder.push_bytes to stay on native array lowering:\n%s", pushBytesBody)
	}

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-stringbuilder-native-push-all-exec", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(stdout) != "Hello World" {
		t.Fatalf("expected StringBuilder native push_all exec to print Hello World, got %q", stdout)
	}
}

func TestCompilerNoFallbacksVectorStringCompiledBuildKeepsNativeStringReceiver(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-vector-string-native-receiver", strings.Join([]string{
		"package demo",
		"import able.collections.vector.{Vector}",
		"",
		"fn main() -> void {",
		"  base: Vector String := Vector.new()",
		"  base = base.push(\"zero\")",
		"  base = base.push(\"one\")",
		"  base = base.push(\"two\")",
		"  updated := base.set(1, \"ONE\")",
		"  print(updated.get(1))",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(stdout) != "ONE" {
		t.Fatalf("expected compiled Vector String regression to print ONE, got %q", stdout)
	}
}
