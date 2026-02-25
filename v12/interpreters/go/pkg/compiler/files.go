package compiler

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeFiles(dir string, files map[string][]byte) error {
	if dir == "" {
		return fmt.Errorf("compiler: empty output dir")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("compiler: create output dir: %w", err)
	}

	// Remove stale .go files not in the new output set.
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") {
				continue
			}
			if _, keep := files[name]; !keep {
				_ = os.Remove(filepath.Join(dir, name))
			}
		}
	}

	// Write files, skipping unchanged ones to preserve mtime for go build cache.
	for name, data := range files {
		path := filepath.Join(dir, name)
		existing, err := os.ReadFile(path)
		if err == nil && bytes.Equal(existing, data) {
			continue
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return fmt.Errorf("compiler: write %s: %w", name, err)
		}
	}
	return nil
}
