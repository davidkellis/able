package driver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderPhaseObserver(t *testing.T) {
	root := t.TempDir()
	entry := filepath.Join(root, "main.able")
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: phase_probe\n"), 0o600); err != nil {
		t.Fatalf("write package manifest: %v", err)
	}
	source := []byte("package phase_probe\nfn main() { 1 }\n")
	if err := os.WriteFile(entry, source, 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	loader, err := NewLoader(nil)
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()
	counts := make(map[LoaderPhase]int)
	loader.SetPhaseObserver(func(sample LoaderPhaseSample) {
		if sample.Duration <= 0 {
			t.Fatalf("%s duration must be positive", sample.Phase)
		}
		counts[sample.Phase]++
	})
	if _, err := loader.Load(entry); err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, phase := range []LoaderPhase{LoaderPhaseNativeParse, LoaderPhaseASTMapping, LoaderPhaseOriginAnnotation} {
		if counts[phase] == 0 {
			t.Fatalf("missing %s phase sample", phase)
		}
	}
}
