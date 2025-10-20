package interpreter

import (
	"path/filepath"
	"testing"
)

func TestConcurrencyFixturesWithGoroutineExecutor(t *testing.T) {
	fixtures := map[string]struct{}{
		"concurrency/proc_cancel_value":            {},
		"concurrency/future_memoization":           {},
		"concurrency/proc_cancelled_outside_error": {},
		"concurrency/proc_cancelled_helper":        {},
	}

	root := filepath.Join("..", "..", "..", "fixtures", "ast")

	walkFixtures(t, root, func(dir string) {
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			t.Fatalf("failed to compute relative path: %v", err)
		}
		if _, ok := fixtures[rel]; !ok {
			return
		}
		t.Run(rel, func(t *testing.T) {
			runFixtureWithExecutor(t, dir, NewGoroutineExecutor(nil))
		})
	})
}
