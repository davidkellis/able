package profilehook

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStartFromEnvNoopWhenUnset(t *testing.T) {
	t.Setenv(cpuProfileEnv, "")
	t.Setenv(memProfileEnv, "")

	stop, err := StartFromEnv()
	if err != nil {
		t.Fatalf("StartFromEnv() error = %v", err)
	}
	if stop != nil {
		t.Fatalf("expected nil stop function when profiling env vars are unset")
	}
}

func TestStartFromEnvWritesCPUAndHeapProfiles(t *testing.T) {
	dir := t.TempDir()
	cpuPath := filepath.Join(dir, "profiles", "cpu.pprof")
	memPath := filepath.Join(dir, "profiles", "heap.pprof")
	t.Setenv(cpuProfileEnv, cpuPath)
	t.Setenv(memProfileEnv, memPath)

	stop, err := StartFromEnv()
	if err != nil {
		t.Fatalf("StartFromEnv() error = %v", err)
	}
	if stop == nil {
		t.Fatalf("expected stop function when profiling env vars are set")
	}

	acc := 0
	for i := 0; i < 1_000_000; i++ {
		acc += i
	}
	if acc == 0 {
		t.Fatalf("unexpected accumulator result")
	}

	if err := stop(); err != nil {
		t.Fatalf("stop() error = %v", err)
	}

	for _, path := range []string{cpuPath, memPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected profile %s to exist: %v", path, err)
		}
		if info.Size() == 0 {
			t.Fatalf("expected profile %s to be non-empty", path)
		}
	}
}
