package compiler

import (
	"os"
	"strings"
	"testing"
)

// compilerFixtureExperimentalExecutionContextEnv enables the candidate ABI
// only for fixture/execution test harnesses. It is intentionally test-only:
// production defaults remain governed solely by Options and the CLI flag.
const compilerFixtureExperimentalExecutionContextEnv = "ABLE_COMPILER_FIXTURE_EXPERIMENTAL_EXECUTION_CONTEXT"

func compilerFixtureOptions(t *testing.T, opts Options) Options {
	t.Helper()
	if compilerFixtureExperimentalExecutionContext(t) {
		opts.ExperimentalExecutionContext = true
	}
	return opts
}

func compilerFixtureExperimentalExecutionContext(t *testing.T) bool {
	t.Helper()
	raw, ok := os.LookupEnv(compilerFixtureExperimentalExecutionContextEnv)
	if !ok {
		return false
	}
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", "0", "false", "no", "off":
		return false
	case "1", "true", "yes", "on":
		return true
	default:
		t.Fatalf("invalid %s value %q (expected one of: 1,true,yes,on,0,false,no,off)", compilerFixtureExperimentalExecutionContextEnv, raw)
		return false
	}
}

func TestCompilerFixtureExperimentalExecutionContextOption(t *testing.T) {
	t.Setenv(compilerFixtureExperimentalExecutionContextEnv, "1")
	if got := compilerFixtureOptions(t, Options{}).ExperimentalExecutionContext; !got {
		t.Fatal("fixture option gate did not enable the experimental execution context")
	}
}
