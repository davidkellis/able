package main

import (
	"os/exec"
	"testing"
)

func TestTestCommandCompiledDryRunMatchesInterpretedListFormat(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	dir := enterTempWorkingDir(t)
	writeMinimalTestCliWorkspace(t, dir)

	stdlibSrc := writeMinimalTestCliStdlib(t, dir)
	configureMinimalTestCliEnv(t, stdlibSrc, false)

	stdout := runCLIExpectSuccess(t,
		"test",
		"--compiled",
		"--dry-run",
		".",
	)
	expected := "demo.framework | pkg | demo-1 | example works | tags=fast | metadata=kind=demo"
	assertOutputContainsAll(t, stdout, expected)
}
