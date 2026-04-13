package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRescueHigherOrderCallKeepsDynamicErrorValueBinding(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct ChannelNil {}",
		"",
		"fn capture(action: fn() -> String) -> String {",
		"  do { action(); \"ok\" } rescue {",
		"    case err => {",
		"      err.value match {",
		"        case ChannelNil {} => \"ChannelNil\",",
		"        case _ => \"Other\"",
		"      }",
		"    }",
		"  }",
		"}",
		"",
		"fn main() -> String {",
		"  capture({ => do { __able_channel_close(0); \"ok\" } })",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_capture")
	if !ok {
		t.Fatalf("could not find compiled capture function")
	}
	if strings.Contains(body, "var err string") {
		t.Fatalf("expected higher-order rescue binding to avoid String mis-inference:\n%s", body)
	}
}
