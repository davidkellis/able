package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/profilehook"
)

const bytecodeStatsOutputEnv = "ABLE_BYTECODE_STATS_OUT"
const bytecodeStatsMainOnlyEnv = "ABLE_BYTECODE_STATS_MAIN_ONLY"

func bytecodeStatsMainOnlyEnabled() bool {
	return strings.TrimSpace(os.Getenv(bytecodeStatsMainOnlyEnv)) != ""
}

// newBytecodeStatsOutput installs an optional snapshot sink for one CLI
// invocation. A profiling interrupt takes the profilehook exit path, so the
// hook writes the same snapshot before that path terminates the process.
func newBytecodeStatsOutput(interp *interpreter.Interpreter) func() {
	path := strings.TrimSpace(os.Getenv(bytecodeStatsOutputEnv))
	if path == "" || interp == nil {
		return func() {}
	}

	var writeOnce sync.Once
	write := func() {
		writeOnce.Do(func() {
			if err := writeBytecodeStatsSnapshot(path, interp.BytecodeStats()); err != nil {
				fmt.Fprintf(os.Stderr, "bytecode stats: %v\n", err)
			}
		})
	}
	unregister := profilehook.RegisterStopHook(write)
	return func() {
		write()
		unregister()
	}
}

func writeBytecodeStatsSnapshot(path string, snapshot interpreter.BytecodeStatsSnapshot) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", bytecodeStatsOutputEnv, err)
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fmt.Errorf("prepare %s: %w", bytecodeStatsOutputEnv, err)
	}
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	if err := os.WriteFile(abs, append(payload, '\n'), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", bytecodeStatsOutputEnv, err)
	}
	return nil
}
