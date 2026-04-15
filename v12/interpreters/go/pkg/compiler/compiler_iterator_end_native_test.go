package compiler

import (
	"strings"
	"testing"
)

func generatorYieldIteratorEndSource() string {
	return strings.Join([]string{
		"package demo",
		"",
		"fn describe(next) -> String {",
		"  next match {",
		"    case IteratorEnd {} => \"end\",",
		"    case v: i32 => `value ${v}`,",
		"    case _ => \"other\"",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  iter := Iterator i32 { gen =>",
		"    print(\"gen-start\")",
		"    gen.yield(1)",
		"    print(\"gen-after-yield\")",
		"    gen.stop()",
		"    print(\"gen-after-stop\")",
		"    gen.yield(2)",
		"  }",
		"",
		"  print(`next1 ${describe(iter.next())}`)",
		"  print(`next2 ${describe(iter.next())}`)",
		"  print(`next3 ${describe(iter.next())}`)",
		"}",
		"",
	}, "\n")
}

func TestCompilerGeneratorYieldIteratorEndExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-generator-yield-iterator-end-", generatorYieldIteratorEndSource(), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(stdout) != strings.Join([]string{
		"gen-start",
		"next1 value 1",
		"gen-after-yield",
		"next2 end",
		"next3 end",
	}, "\n") {
		t.Fatalf("expected compiled generator IteratorEnd program to match fixture output, got %q", stdout)
	}
}
