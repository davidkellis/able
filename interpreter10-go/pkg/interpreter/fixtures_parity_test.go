package interpreter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureParityStringLiteral(t *testing.T) {
	root := filepath.Join("..", "..", "..", "fixtures", "ast")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("reading fixtures: %v", err)
	}

	t.Setenv(fixtureTypecheckEnv, "warn")

	concurrencyFixtures := map[string]struct{}{
		"concurrency/proc_cancel_value":             {},
		"concurrency/proc_value_memoization":        {},
		"concurrency/proc_value_cancel_memoization": {},
		"concurrency/future_memoization":            {},
		"concurrency/proc_cancelled_outside_error":  {},
		"concurrency/proc_cancelled_helper":         {},
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		fixtureDir := filepath.Join(root, entry.Name())
		walkFixtures(t, fixtureDir, func(dir string) {
			rel, err := filepath.Rel(root, dir)
			if err != nil {
				t.Fatalf("computing relative path for %s: %v", dir, err)
			}
			t.Run(rel, func(t *testing.T) {
				var exec Executor
				if _, ok := concurrencyFixtures[rel]; ok {
					exec = NewGoroutineExecutor(nil)
				}
				runFixtureWithExecutor(t, dir, rel, exec)
			})
		})
	}
}

func walkFixtures(t *testing.T, dir string, fn func(string)) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	hasModule := false
	for _, entry := range entries {
		if entry.Type().IsRegular() && entry.Name() == "module.json" {
			hasModule = true
		}
	}
	if hasModule {
		fn(dir)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			walkFixtures(t, filepath.Join(dir, entry.Name()), fn)
		}
	}
}
