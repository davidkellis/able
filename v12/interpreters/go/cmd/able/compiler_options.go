package main

import (
	"fmt"
	"os"
	"strings"
)

func resolveCompilerRequireNoFallbacksFromEnv() (bool, error) {
	raw, ok := os.LookupEnv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS")
	if !ok {
		return false, nil
	}
	value, err := parseCompilerBoolEnv(raw)
	if err != nil {
		return false, fmt.Errorf("invalid ABLE_COMPILER_REQUIRE_NO_FALLBACKS value %q (expected one of: 1,true,yes,on,0,false,no,off)", raw)
	}
	return value, nil
}

func resolveCompilerExperimentalMonoArraysFromEnv() (bool, error) {
	raw, ok := os.LookupEnv("ABLE_EXPERIMENTAL_MONO_ARRAYS")
	if !ok {
		return true, nil
	}
	value, err := parseCompilerBoolEnv(raw)
	if err != nil {
		return false, fmt.Errorf("invalid ABLE_EXPERIMENTAL_MONO_ARRAYS value %q (expected one of: 1,true,yes,on,0,false,no,off)", raw)
	}
	return value, nil
}

func parseCompilerBoolEnv(raw string) (bool, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	switch normalized {
	case "", "0", "false", "no", "off":
		return false, nil
	case "1", "true", "yes", "on":
		return true, nil
	default:
		return false, fmt.Errorf("invalid boolean value")
	}
}
