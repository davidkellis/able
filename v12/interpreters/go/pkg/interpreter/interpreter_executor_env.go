package interpreter

import (
	"fmt"
	"os"
	"strings"
)

const ExecutorEnvVar = "ABLE_EXECUTOR"

// ExecutorKindFromEnvironment resolves the configured executor kind from the
// process environment. The empty value defaults to the serial executor.
func ExecutorKindFromEnvironment() (string, error) {
	return normalizeExecutorKind(os.Getenv(ExecutorEnvVar))
}

// NewExecutorFromEnvironment constructs the configured executor from the
// process environment. The empty value defaults to the serial executor.
func NewExecutorFromEnvironment() (Executor, error) {
	kind, err := ExecutorKindFromEnvironment()
	if err != nil {
		return nil, err
	}
	return newExecutorForKind(kind), nil
}

func normalizeExecutorKind(raw string) (string, error) {
	kind := strings.ToLower(strings.TrimSpace(raw))
	switch kind {
	case "", "serial":
		return "serial", nil
	case "goroutine":
		return "goroutine", nil
	default:
		return "", fmt.Errorf("unknown %s value %q (want serial or goroutine)", ExecutorEnvVar, raw)
	}
}

func newExecutorForKind(kind string) Executor {
	if strings.EqualFold(kind, "goroutine") {
		return NewGoroutineExecutor(nil)
	}
	return NewSerialExecutor(nil)
}
