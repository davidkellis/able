package interpreter

import (
	"flag"
	"strings"
	"testing"
)

type testExecMode string

const (
	testExecTreewalker testExecMode = "treewalker"
	testExecBytecode   testExecMode = "bytecode"
)

var execModeFlag = flag.String("exec-mode", string(testExecTreewalker), "execution mode for interpreter tests (treewalker|bytecode)")

func resolveTestExecMode(t *testing.T) testExecMode {
	t.Helper()
	switch strings.ToLower(strings.TrimSpace(*execModeFlag)) {
	case string(testExecTreewalker), "":
		return testExecTreewalker
	case string(testExecBytecode):
		return testExecBytecode
	default:
		t.Fatalf("unknown exec mode %q (expected treewalker or bytecode)", *execModeFlag)
		return testExecTreewalker
	}
}

func newTestInterpreter(t *testing.T, mode testExecMode, executor Executor) *Interpreter {
	t.Helper()
	switch mode {
	case testExecTreewalker:
		return NewWithExecutor(executor)
	case testExecBytecode:
		// TODO: wire bytecode backend once available.
		return NewWithExecutor(executor)
	default:
		return NewWithExecutor(executor)
	}
}
