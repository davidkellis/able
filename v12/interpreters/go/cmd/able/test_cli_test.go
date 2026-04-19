package main

import "testing"

func TestTestCommandReportsEmptyWorkspaceInListMode(t *testing.T) {
	enterTempWorkingDir(t)

	stdout := runCLIExpectSuccess(t, "test", "--list")
	assertOutputContainsAll(t, stdout, "able test: no test modules found")
}

func TestTestCommandParsesFiltersAndTargets(t *testing.T) {
	dir := enterTempWorkingDir(t)
	writeMinimalTestCliWorkspace(t, dir)

	stdlibSrc := writeMinimalTestCliStdlib(t, dir)
	configureMinimalTestCliEnv(t, stdlibSrc, true)

	stdout := runCLIExpectSuccess(t,
		"test",
		"--path", "pkg",
		"--exclude-path", "tmp",
		"--name", "example works",
		"--exclude-name", "skip",
		"--tag", "fast",
		"--exclude-tag", "flaky",
		"--format", "progress",
		"--fail-fast",
		"--repeat", "3",
		"--parallel", "2",
		"--shuffle", "123",
		"--dry-run",
		".",
	)
	assertOutputContainsAll(t, stdout, "demo.framework", "example", "tags=", "metadata=")
}

func TestTestCommandCompiledRuns(t *testing.T) {
	dir := enterTempWorkingDir(t)
	writeCompiledSampleTestModule(t, dir)

	stdout := runCompiledSampleTests(t, dir)
	assertCompiledSampleSuccess(t, stdout)
}

func TestTestCommandCompiledRejectsInvalidRequireNoFallbacksEnv(t *testing.T) {
	dir := enterTempWorkingDir(t)
	writeCompiledSampleTestModule(t, dir)

	requireGoToolchain(t)
	configureRepoCompiledEnv(t)
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "sometimes")

	_, stderr := runCLIExpectFailureCode(t, 2, "test", "--compiled", dir)
	assertTextContainsAll(t, stderr, "invalid ABLE_COMPILER_REQUIRE_NO_FALLBACKS value")
}
