package compiler

import (
	"os"
	"strings"
	"testing"
)

const compilerFixtureRequireNoFallbacksEnv = "ABLE_COMPILER_FIXTURE_REQUIRE_NO_FALLBACKS"

func requireNoFallbacksForFixtureGates(t *testing.T) bool {
	t.Helper()
	raw, ok := os.LookupEnv(compilerFixtureRequireNoFallbacksEnv)
	if !ok {
		return true
	}
	normalized := strings.TrimSpace(strings.ToLower(raw))
	switch normalized {
	case "", "0", "false", "no", "off":
		return false
	case "1", "true", "yes", "on":
		return true
	default:
		t.Fatalf("invalid %s value %q (expected one of: 1,true,yes,on,0,false,no,off)", compilerFixtureRequireNoFallbacksEnv, raw)
		return true
	}
}
