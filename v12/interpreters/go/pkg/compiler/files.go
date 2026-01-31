package compiler

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeFiles(dir string, files map[string][]byte) error {
	if dir == "" {
		return fmt.Errorf("compiler: empty output dir")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("compiler: create output dir: %w", err)
	}
	for name, data := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return fmt.Errorf("compiler: write %s: %w", name, err)
		}
	}
	return nil
}
