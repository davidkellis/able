package profilehook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestStartFromEnvNoopWhenUnset(t *testing.T) {
	t.Setenv(cpuProfileEnv, "")
	t.Setenv(memProfileEnv, "")
	t.Setenv(allocProfileEnv, "")

	stop, err := StartFromEnv()
	if err != nil {
		t.Fatalf("StartFromEnv() error = %v", err)
	}
	if stop != nil {
		t.Fatalf("expected nil stop function when profiling env vars are unset")
	}
}

func TestPhaseProfilerNoopWhenUnset(t *testing.T) {
	t.Setenv(cpuProfileEnv, "")
	t.Setenv(phaseProfileDirEnv, "")
	t.Setenv(phaseCPUProfileDirEnv, "")
	t.Setenv(phaseAllocProfileDirEnv, "")
	t.Setenv(phaseStatsDirEnv, "")

	profiler, err := NewPhaseProfilerFromEnv()
	if err != nil {
		t.Fatalf("NewPhaseProfilerFromEnv() error = %v", err)
	}
	if profiler != nil {
		t.Fatal("expected nil phase profiler when phase profiling is unset")
	}
}

func TestPhaseCPUOnlyProfilerWritesProfilesWithoutAllocationSnapshots(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(cpuProfileEnv, "")
	t.Setenv(phaseProfileDirEnv, "")
	t.Setenv(phaseCPUProfileDirEnv, dir)
	t.Setenv(phaseAllocProfileDirEnv, "")
	t.Setenv(phaseStatsDirEnv, "")

	before := runtime.MemProfileRate
	profiler, err := NewPhaseProfilerFromEnv()
	if err != nil {
		t.Fatalf("NewPhaseProfilerFromEnv() error = %v", err)
	}
	if profiler == nil {
		t.Fatal("expected phase CPU-only profiler")
	}
	if profiler.captureAllocationSnapshots {
		t.Fatal("CPU-only profiler unexpectedly captures allocation snapshots")
	}
	if got := runtime.MemProfileRate; got != before {
		t.Fatalf("MemProfileRate after CPU-only setup = %d, want %d", got, before)
	}
	if err := profiler.StartBootstrap(); err != nil {
		t.Fatalf("StartBootstrap() error = %v", err)
	}
	busyProfileWork()
	if err := profiler.StartMain(); err != nil {
		t.Fatalf("StartMain() error = %v", err)
	}
	busyProfileWork()
	if err := profiler.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	for _, name := range []string{"bootstrap.cpu.pprof", "main.cpu.pprof"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("expected profile %s to exist: %v", name, err)
		}
		if info.Size() == 0 {
			t.Fatalf("expected profile %s to be non-empty", name)
		}
	}
	for _, name := range []string{
		"bootstrap-start.allocs.pprof", "bootstrap-end.allocs.pprof",
		"main-start.allocs.pprof", "main-end.allocs.pprof", "phase-stats.json",
	} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected no allocation artifact %s, got %v", name, err)
		}
	}
	if got := runtime.MemProfileRate; got != before {
		t.Fatalf("MemProfileRate after CPU-only stop = %d, want %d", got, before)
	}
}

func TestPhaseAllocationOnlyProfilerWritesSnapshotsWithoutCPUProfiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(cpuProfileEnv, "")
	t.Setenv(phaseProfileDirEnv, "")
	t.Setenv(phaseCPUProfileDirEnv, "")
	t.Setenv(phaseAllocProfileDirEnv, dir)
	t.Setenv(phaseStatsDirEnv, "")

	before := runtime.MemProfileRate
	profiler, err := NewPhaseProfilerFromEnv()
	if err != nil {
		t.Fatalf("NewPhaseProfilerFromEnv() error = %v", err)
	}
	if profiler == nil {
		t.Fatal("expected phase allocation-only profiler")
	}
	if profiler.captureCPUProfiles {
		t.Fatal("allocation-only profiler unexpectedly captures CPU profiles")
	}
	if !profiler.captureAllocationSnapshots {
		t.Fatal("allocation-only profiler does not capture allocation snapshots")
	}
	if got := runtime.MemProfileRate; got != 1 {
		t.Fatalf("MemProfileRate after allocation-only setup = %d, want 1", got)
	}
	if err := profiler.StartBootstrap(); err != nil {
		t.Fatalf("StartBootstrap() error = %v", err)
	}
	busyProfileWork()
	if err := profiler.StartMain(); err != nil {
		t.Fatalf("StartMain() error = %v", err)
	}
	busyProfileWork()
	if err := profiler.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	for _, name := range []string{
		"bootstrap-start.allocs.pprof", "bootstrap-end.allocs.pprof",
		"main-start.allocs.pprof", "main-end.allocs.pprof", "phase-stats.json",
	} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected allocation artifact %s: %v", name, err)
		}
	}
	for _, name := range []string{"bootstrap.cpu.pprof", "main.cpu.pprof"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected no CPU artifact %s, got %v", name, err)
		}
	}
	if got := runtime.MemProfileRate; got != before {
		t.Fatalf("MemProfileRate after allocation-only stop = %d, want %d", got, before)
	}
}

func TestPhaseStatsOnlyProfilerWritesStatsWithoutProfiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(cpuProfileEnv, "")
	t.Setenv(phaseProfileDirEnv, "")
	t.Setenv(phaseCPUProfileDirEnv, "")
	t.Setenv(phaseAllocProfileDirEnv, "")
	t.Setenv(phaseStatsDirEnv, dir)

	before := runtime.MemProfileRate
	profiler, err := NewPhaseProfilerFromEnv()
	if err != nil {
		t.Fatalf("NewPhaseProfilerFromEnv() error = %v", err)
	}
	if profiler == nil || !profiler.captureAllocationSnapshots {
		t.Fatal("expected phase stats-only profiler")
	}
	if profiler.captureCPUProfiles || profiler.captureAllocationProfiles {
		t.Fatal("stats-only profiler unexpectedly captures a profile")
	}
	if got := runtime.MemProfileRate; got != before {
		t.Fatalf("MemProfileRate after stats-only setup = %d, want %d", got, before)
	}
	if err := profiler.StartBootstrap(); err != nil {
		t.Fatalf("StartBootstrap() error = %v", err)
	}
	busyProfileWork()
	if err := profiler.StartMain(); err != nil {
		t.Fatalf("StartMain() error = %v", err)
	}
	busyProfileWork()
	if err := profiler.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "phase-stats.json"))
	if err != nil {
		t.Fatalf("read phase allocation stats: %v", err)
	}
	var stats phaseProfileStats
	if err := json.Unmarshal(data, &stats); err != nil {
		t.Fatalf("decode phase allocation stats: %v", err)
	}
	if len(stats.Phases) != 2 || stats.Phases[1].Phase != "main" {
		t.Fatalf("unexpected phase stats: %+v", stats)
	}
	for _, name := range []string{
		"bootstrap.cpu.pprof", "main.cpu.pprof",
		"bootstrap-start.allocs.pprof", "bootstrap-end.allocs.pprof",
		"main-start.allocs.pprof", "main-end.allocs.pprof",
	} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected no profile artifact %s, got %v", name, err)
		}
	}
}

func TestPhaseProfilerRejectsBothPhaseModes(t *testing.T) {
	t.Setenv(cpuProfileEnv, "")
	t.Setenv(phaseProfileDirEnv, t.TempDir())
	t.Setenv(phaseCPUProfileDirEnv, t.TempDir())
	t.Setenv(phaseAllocProfileDirEnv, "")
	t.Setenv(phaseStatsDirEnv, "")

	if _, err := NewPhaseProfilerFromEnv(); err == nil {
		t.Fatal("expected phase profiler to reject both phase modes")
	}
}

func TestStartFromEnvWritesCPUHeapAndAllocationProfiles(t *testing.T) {
	dir := t.TempDir()
	cpuPath := filepath.Join(dir, "profiles", "cpu.pprof")
	memPath := filepath.Join(dir, "profiles", "heap.pprof")
	allocPath := filepath.Join(dir, "profiles", "allocs.pprof")
	t.Setenv(cpuProfileEnv, cpuPath)
	t.Setenv(memProfileEnv, memPath)
	t.Setenv(allocProfileEnv, allocPath)

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

	for _, path := range []string{cpuPath, memPath, allocPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected profile %s to exist: %v", path, err)
		}
		if info.Size() == 0 {
			t.Fatalf("expected profile %s to be non-empty", path)
		}
	}
}

func TestPhaseProfilerWritesBootstrapAndMainProfiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(cpuProfileEnv, "")
	t.Setenv(phaseProfileDirEnv, dir)
	t.Setenv(phaseCPUProfileDirEnv, "")
	t.Setenv(phaseAllocProfileDirEnv, "")
	t.Setenv(phaseStatsDirEnv, "")

	profiler, err := NewPhaseProfilerFromEnv()
	if err != nil {
		t.Fatalf("NewPhaseProfilerFromEnv() error = %v", err)
	}
	if profiler == nil {
		t.Fatal("expected phase profiler")
	}
	if err := profiler.StartBootstrap(); err != nil {
		t.Fatalf("StartBootstrap() error = %v", err)
	}
	busyProfileWork()
	if err := profiler.StartMain(); err != nil {
		t.Fatalf("StartMain() error = %v", err)
	}
	busyProfileWork()
	if err := profiler.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	for _, name := range []string{
		"bootstrap.cpu.pprof",
		"bootstrap-start.allocs.pprof",
		"bootstrap-end.allocs.pprof",
		"main.cpu.pprof",
		"main-start.allocs.pprof",
		"main-end.allocs.pprof",
		"phase-stats.json",
	} {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected profile %s to exist: %v", path, err)
		}
		if info.Size() == 0 {
			t.Fatalf("expected profile %s to be non-empty", path)
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "phase-stats.json"))
	if err != nil {
		t.Fatalf("read phase allocation stats: %v", err)
	}
	var stats phaseProfileStats
	if err := json.Unmarshal(data, &stats); err != nil {
		t.Fatalf("decode phase allocation stats: %v", err)
	}
	if stats.Version != 1 || len(stats.Phases) != 2 {
		t.Fatalf("unexpected phase allocation stats: %+v", stats)
	}
	if stats.Phases[0].Phase != "bootstrap" || stats.Phases[1].Phase != "main" {
		t.Fatalf("unexpected phase order: %+v", stats.Phases)
	}
	for _, phase := range stats.Phases {
		if phase.AllocatedBytes == 0 || phase.Allocations == 0 {
			t.Fatalf("expected allocation activity for %s, got %+v", phase.Phase, phase)
		}
	}
}

func TestPhaseProfilerRejectsStandaloneCPUProfile(t *testing.T) {
	t.Setenv(cpuProfileEnv, filepath.Join(t.TempDir(), "cpu.pprof"))
	t.Setenv(phaseProfileDirEnv, t.TempDir())
	t.Setenv(phaseCPUProfileDirEnv, "")
	t.Setenv(phaseAllocProfileDirEnv, "")
	t.Setenv(phaseStatsDirEnv, "")

	if _, err := NewPhaseProfilerFromEnv(); err == nil {
		t.Fatal("expected phase profiler to reject an active standalone CPU profile")
	}
}

func TestPhaseProfilerRestoresMemProfileRate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(cpuProfileEnv, "")
	t.Setenv(phaseProfileDirEnv, dir)
	t.Setenv(phaseCPUProfileDirEnv, "")
	t.Setenv(phaseAllocProfileDirEnv, "")
	t.Setenv(phaseStatsDirEnv, "")

	before := runtime.MemProfileRate
	profiler, err := NewPhaseProfilerFromEnv()
	if err != nil {
		t.Fatalf("NewPhaseProfilerFromEnv() error = %v", err)
	}
	if got := runtime.MemProfileRate; got != 1 {
		t.Fatalf("MemProfileRate after phase-profiler setup = %d, want 1", got)
	}
	if err := profiler.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if got := runtime.MemProfileRate; got != before {
		t.Fatalf("MemProfileRate after phase-profiler stop = %d, want %d", got, before)
	}
}

func TestPhaseProfilerWritesBootstrapStatsOnEarlyStop(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(cpuProfileEnv, "")
	t.Setenv(phaseProfileDirEnv, dir)
	t.Setenv(phaseCPUProfileDirEnv, "")
	t.Setenv(phaseAllocProfileDirEnv, "")
	t.Setenv(phaseStatsDirEnv, "")

	profiler, err := NewPhaseProfilerFromEnv()
	if err != nil {
		t.Fatalf("NewPhaseProfilerFromEnv() error = %v", err)
	}
	if err := profiler.StartBootstrap(); err != nil {
		t.Fatalf("StartBootstrap() error = %v", err)
	}
	busyProfileWork()
	if err := profiler.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "phase-stats.json"))
	if err != nil {
		t.Fatalf("read phase allocation stats: %v", err)
	}
	var stats phaseProfileStats
	if err := json.Unmarshal(data, &stats); err != nil {
		t.Fatalf("decode phase allocation stats: %v", err)
	}
	if len(stats.Phases) != 1 || stats.Phases[0].Phase != "bootstrap" {
		t.Fatalf("expected bootstrap-only early-stop stats, got %+v", stats)
	}
	for _, name := range []string{"bootstrap-start.allocs.pprof", "bootstrap-end.allocs.pprof"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected bootstrap allocation snapshot %s: %v", name, err)
		}
	}
	for _, name := range []string{"main-start.allocs.pprof", "main-end.allocs.pprof"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected no main allocation snapshot %s, got %v", name, err)
		}
	}
}

func busyProfileWork() {
	chunks := make([][]byte, 64)
	acc := 0
	for i := 0; i < len(chunks); i++ {
		chunks[i] = make([]byte, 256)
		chunks[i][0] = byte(i)
	}
	for i := 0; i < 1_000_000; i++ {
		acc += i
	}
	for _, chunk := range chunks {
		acc += int(chunk[0])
	}
	if acc == 0 {
		panic("unexpected profile accumulator")
	}
}
