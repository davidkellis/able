package interpreter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var serialOnlyConcurrencyFixtures = map[string]struct{}{
	"concurrency/future_flush_fairness":       {},
	"concurrency/future_yield_flush":          {},
	"concurrency/future_executor_diagnostics": {},
}

func TestFixtureParityStringLiteral(t *testing.T) {
	root := filepath.Join(repositoryRoot(), "v11", "fixtures", "ast")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "ast")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("reading fixtures: %v", err)
	}

	t.Setenv(fixtureTypecheckEnv, "warn")

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
				relUnix := filepath.ToSlash(rel)
				if strings.HasPrefix(relUnix, "concurrency/") {
					if _, serialOnly := serialOnlyConcurrencyFixtures[relUnix]; serialOnly {
						runFixtureWithExecutor(t, dir, rel, nil)
						return
					}
					cases := []struct {
						name string
						exec func() Executor
					}{
						{name: "serial", exec: func() Executor { return nil }},
						{name: "goroutine", exec: func() Executor { return NewGoroutineExecutor(nil) }},
					}
					for _, tc := range cases {
						tc := tc
						t.Run(tc.name, func(t *testing.T) {
							runFixtureWithExecutor(t, dir, rel, tc.exec())
						})
					}
					return
				}
				runFixtureWithExecutor(t, dir, rel, nil)
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
