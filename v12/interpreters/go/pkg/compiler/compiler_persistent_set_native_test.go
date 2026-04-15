package compiler

import (
	"strings"
	"testing"
)

func persistentSetAssocSource() string {
	return strings.Join([]string{
		"package demo",
		"",
		"import able.collections.persistent_map.{PersistentSet}",
		"",
		"fn main() -> void {",
		"  set: PersistentSet i32 = PersistentSet.empty()",
		"  set = set.insert(1)",
		"  set = set.insert(2)",
		"  print(set.contains(1))",
		"  print(set.contains(2))",
		"}",
		"",
	}, "\n")
}

func TestCompilerPersistentSetInsertExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-persistent-set-assoc-", persistentSetAssocSource(), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(stdout) != "true\ntrue" && strings.TrimSpace(stdout) != "true\r\ntrue" {
		t.Fatalf("expected compiled PersistentSet insert program to print true/true, got %q", stdout)
	}
}

func TestCompilerPersistentSetInsertAvoidsOuterExpectedTypeLeakInAssocPatternBindings(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-persistent-set-assoc-audit-", persistentSetAssocSource())
	compiled := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"var existing_key *HamtAssocResult",
		"var existing_value *HamtAssocResult",
	} {
		if strings.Contains(compiled, fragment) {
			t.Fatalf("expected PersistentSet assoc pattern bindings to avoid leaking the outer HamtAssocResult expected type into nested field bindings (%q):\n%s", fragment, compiled)
		}
	}
}
