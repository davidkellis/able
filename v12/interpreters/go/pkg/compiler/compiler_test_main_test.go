package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	cleanStaleTmpDirs()
	os.Exit(m.Run())
}

// cleanStaleTmpDirs removes orphaned ablec-* directories from tmp/ that were
// left behind by interrupted test runs. Directories older than 10 minutes are
// considered stale.
func cleanStaleTmpDirs() {
	moduleRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		return
	}
	tmpRoot := filepath.Join(moduleRoot, "tmp")
	entries, err := os.ReadDir(tmpRoot)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-10 * time.Minute)
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "ablec-") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.RemoveAll(filepath.Join(tmpRoot, entry.Name()))
		}
	}
}
