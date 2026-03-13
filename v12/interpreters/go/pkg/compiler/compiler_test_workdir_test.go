package compiler

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func compilerTestWorkDir(tb testing.TB, prefix string) (string, string) {
	return compilerTestWorkDirWithCleanup(tb, prefix, true)
}

func compilerTestWorkDirNoCleanup(tb testing.TB, prefix string) (string, string) {
	return compilerTestWorkDirWithCleanup(tb, prefix, false)
}

func compilerTestWorkDirWithCleanup(tb testing.TB, prefix string, cleanup bool) (string, string) {
	tb.Helper()
	moduleRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		tb.Fatalf("module root: %v", err)
	}
	tmpRoot := filepath.Join(moduleRoot, "tmp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		tb.Fatalf("mkdir tmp: %v", err)
	}
	dirName := compilerTestTempDirName(prefix, tb.Name())
	workDir := filepath.Join(tmpRoot, dirName)
	if err := os.RemoveAll(workDir); err != nil {
		tb.Fatalf("reset temp dir: %v", err)
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		tb.Fatalf("mkdir temp dir: %v", err)
	}
	if cleanup {
		tb.Cleanup(func() { _ = os.RemoveAll(workDir) })
	}
	return moduleRoot, workDir
}

func compilerTestTempDirName(prefix string, key string) string {
	trimmed := strings.TrimSuffix(strings.TrimSpace(prefix), "-")
	if trimmed == "" {
		trimmed = "ablec"
	}
	sum := sha256.Sum256([]byte(trimmed + "\x00" + key))
	return trimmed + "-" + hex.EncodeToString(sum[:6])
}
