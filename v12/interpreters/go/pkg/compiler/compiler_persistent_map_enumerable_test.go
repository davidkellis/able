package compiler

import (
	"strings"
	"testing"
)

func persistentMapEachSource() string {
	return strings.Join([]string{
		"package demo",
		"",
		"import able.collections.map.{MapEntry}",
		"import able.collections.persistent_map.{PersistentMap}",
		"",
		"fn main() -> void {",
		"  map: PersistentMap String i32 = PersistentMap.empty()",
		"  map = map.set(\"a\", 1)",
		"  map = map.set(\"b\", 2)",
		"  total := 0",
		"  map.each(fn(entry: MapEntry String i32) -> void {",
		"    total = total + entry.value",
		"  })",
		"  print(total)",
		"}",
		"",
	}, "\n")
}

func persistentMapEachNestedCallbackSource() string {
	return strings.Join([]string{
		"package demo",
		"",
		"import able.collections.persistent_map.{PersistentMap}",
		"",
		"struct ExampleContext",
		"",
		"fn with_context(run: ExampleContext -> void) -> void {",
		"  run(ExampleContext)",
		"}",
		"",
		"fn main() -> void {",
		"  map: PersistentMap String i32 = PersistentMap.empty()",
		"  map = map.set(\"a\", 1)",
		"  map = map.set(\"b\", 2)",
		"  total := 0",
		"  with_context(fn(_ctx) -> void {",
		"    map.each(fn(entry) -> void {",
		"      total = total + entry.value",
		"    })",
		"  })",
		"  print(total)",
		"}",
		"",
	}, "\n")
}

func TestCompilerPersistentMapEachExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-persistent-map-each-", persistentMapEachSource(), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(stdout) != "3" {
		t.Fatalf("expected compiled PersistentMap.each program to print 3, got %q", stdout)
	}
}

func TestCompilerPersistentMapEachInfersClosureParamType(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-persistent-map-each-inferred-", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.persistent_map.{PersistentMap}",
		"",
		"fn main() -> void {",
		"  map: PersistentMap String i32 = PersistentMap.empty()",
		"  map = map.set(\"a\", 1)",
		"  map = map.set(\"b\", 2)",
		"  total := 0",
		"  map.each(fn(entry) -> void {",
		"    total = total + entry.value",
		"  })",
		"  print(total)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(stdout) != "3" {
		t.Fatalf("expected compiled PersistentMap.each inferred-closure program to print 3, got %q", stdout)
	}
}

func TestCompilerPersistentMapEachInferredClosureIgnoresOuterCallbackExpectedType(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-persistent-map-each-nested-", persistentMapEachNestedCallbackSource(), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(stdout) != "3" {
		t.Fatalf("expected compiled nested PersistentMap.each inferred-closure program to print 3, got %q", stdout)
	}
}
