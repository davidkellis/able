package bridge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/interpreter"
)

var _ Interpreter = (*interpreter.Interpreter)(nil)

func TestProductionBridgeDoesNotImportConcreteInterpreter(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read bridge package: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if strings.Contains(string(data), `"able/interpreter-go/pkg/interpreter"`) {
			t.Fatalf("production bridge file %s imports the concrete interpreter", name)
		}
	}
}
