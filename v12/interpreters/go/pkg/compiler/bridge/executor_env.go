package bridge

import (
	"fmt"
	"os"
	"strings"
)

const ExecutorEnvVar = "ABLE_EXECUTOR"

// ExecutorKindFromEnvironment resolves the executor policy needed by a static
// compiled runtime without constructing or importing a concrete interpreter.
func ExecutorKindFromEnvironment() (string, error) {
	kind := strings.ToLower(strings.TrimSpace(os.Getenv(ExecutorEnvVar)))
	switch kind {
	case "", "serial":
		return "serial", nil
	case "goroutine":
		return "goroutine", nil
	default:
		return "", fmt.Errorf("unknown %s value %q (want serial or goroutine)", ExecutorEnvVar, os.Getenv(ExecutorEnvVar))
	}
}
