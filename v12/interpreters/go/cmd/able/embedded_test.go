package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureEmbeddedKernelRefreshesStaleSameVersionCache(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ABLE_HOME", home)

	target, err := ensureEmbeddedKernel()
	if err != nil {
		t.Fatalf("initial ensureEmbeddedKernel: %v", err)
	}
	kernelPath := filepath.Join(target, "kernel.able")
	if err := os.WriteFile(kernelPath, []byte("stale kernel"), 0o644); err != nil {
		t.Fatalf("write stale kernel: %v", err)
	}

	refreshed, err := ensureEmbeddedKernel()
	if err != nil {
		t.Fatalf("refresh ensureEmbeddedKernel: %v", err)
	}
	if refreshed != target {
		t.Fatalf("refreshed target = %q, want %q", refreshed, target)
	}
	if !embeddedKernelCacheCurrent(target, filepath.Dir(target)) {
		t.Fatal("cached embedded kernel was not refreshed")
	}
}
