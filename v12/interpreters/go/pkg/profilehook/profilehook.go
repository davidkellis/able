package profilehook

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
)

const (
	cpuProfileEnv           = "ABLE_GO_CPU_PROFILE"
	memProfileEnv           = "ABLE_GO_MEM_PROFILE"
	allocProfileEnv         = "ABLE_GO_ALLOC_PROFILE"
	phaseProfileDirEnv      = "ABLE_GO_PHASE_PROFILE_DIR"
	phaseCPUProfileDirEnv   = "ABLE_GO_PHASE_CPU_PROFILE_DIR"
	phaseAllocProfileDirEnv = "ABLE_GO_PHASE_ALLOC_PROFILE_DIR"
	phaseStatsDirEnv        = "ABLE_GO_PHASE_STATS_DIR"
)

var stopHooks struct {
	sync.Mutex
	nextID uint64
	hooks  map[uint64]func()
}

// RegisterStopHook registers a callback that runs when StartFromEnv handles
// an interrupt. It is intended for opt-in diagnostic snapshots which must be
// written before the profiling signal path exits the process. The returned
// function unregisters the callback and is safe to call more than once.
func RegisterStopHook(hook func()) func() {
	if hook == nil {
		return func() {}
	}
	stopHooks.Lock()
	if stopHooks.hooks == nil {
		stopHooks.hooks = make(map[uint64]func())
	}
	stopHooks.nextID++
	id := stopHooks.nextID
	stopHooks.hooks[id] = hook
	stopHooks.Unlock()

	var unregisterOnce sync.Once
	return func() {
		unregisterOnce.Do(func() {
			stopHooks.Lock()
			delete(stopHooks.hooks, id)
			stopHooks.Unlock()
		})
	}
}

func runStopHooks() {
	stopHooks.Lock()
	hooks := make([]func(), 0, len(stopHooks.hooks))
	for _, hook := range stopHooks.hooks {
		hooks = append(hooks, hook)
	}
	stopHooks.Unlock()
	for _, hook := range hooks {
		hook()
	}
}

// PhaseAllocationStats is the allocation delta observed while one generated
// binary phase is active. Heap fields are signed because live heap state can
// decrease while a phase runs; allocation counters are process-monotonic.
type PhaseAllocationStats struct {
	Phase              string `json:"phase"`
	AllocatedBytes     uint64 `json:"allocated_bytes"`
	Allocations        uint64 `json:"allocations"`
	Frees              uint64 `json:"frees"`
	HeapAllocatedDelta int64  `json:"heap_allocated_delta"`
	HeapObjectsDelta   int64  `json:"heap_objects_delta"`
	GCCount            uint32 `json:"gc_count"`
}

type phaseAllocationSnapshot struct {
	totalAlloc  uint64
	mallocs     uint64
	frees       uint64
	heapAlloc   uint64
	heapObjects uint64
	numGC       uint32
}

type phaseProfileStats struct {
	Version int                    `json:"version"`
	Phases  []PhaseAllocationStats `json:"phases"`
}

// PhaseProfiler writes independent CPU profiles for the generated binary
// launcher bootstrap and its registered entry function. The allocation-capable
// mode also writes boundary allocation snapshots. It is opt-in and does not
// alter normal process execution when no phase-profile variable is set.
type PhaseProfiler struct {
	dir                        string
	captureCPUProfiles         bool
	captureAllocationSnapshots bool
	captureAllocationProfiles  bool
	currentFile                *os.File
	currentPhase               string
	currentStart               phaseAllocationSnapshot
	phases                     []PhaseAllocationStats
	previousMemProfileRate     int
	stopped                    bool
}

// NewPhaseProfilerFromEnv creates a phase profiler when one phase-profile
// directory variable is set. ABLE_GO_PHASE_PROFILE_DIR records CPU profiles
// and exact allocation snapshots, ABLE_GO_PHASE_CPU_PROFILE_DIR records only
// CPU profiles, ABLE_GO_PHASE_ALLOC_PROFILE_DIR records only exact allocation
// snapshots, and ABLE_GO_PHASE_STATS_DIR records only lightweight MemStats
// deltas without enabling one-object allocation profiling.
// A phase profile cannot run together with ABLE_GO_CPU_PROFILE because Go
// permits only one active CPU profiler.
func NewPhaseProfilerFromEnv() (*PhaseProfiler, error) {
	allocationDir := strings.TrimSpace(os.Getenv(phaseProfileDirEnv))
	cpuOnlyDir := strings.TrimSpace(os.Getenv(phaseCPUProfileDirEnv))
	allocationOnlyDir := strings.TrimSpace(os.Getenv(phaseAllocProfileDirEnv))
	statsOnlyDir := strings.TrimSpace(os.Getenv(phaseStatsDirEnv))
	modeCount := 0
	for _, dir := range []string{allocationDir, cpuOnlyDir, allocationOnlyDir, statsOnlyDir} {
		if dir != "" {
			modeCount++
		}
	}
	if modeCount > 1 {
		return nil, fmt.Errorf("profilehook: %s, %s, %s, and %s are mutually exclusive", phaseProfileDirEnv, phaseCPUProfileDirEnv, phaseAllocProfileDirEnv, phaseStatsDirEnv)
	}
	if modeCount == 0 {
		return nil, nil
	}
	if strings.TrimSpace(os.Getenv(cpuProfileEnv)) != "" {
		return nil, fmt.Errorf("profilehook: phase profiling cannot be combined with %s", cpuProfileEnv)
	}
	dir := allocationDir
	dirEnv := phaseProfileDirEnv
	captureAllocationSnapshots := true
	captureAllocationProfiles := true
	captureCPUProfiles := true
	if dir == "" {
		dir = cpuOnlyDir
		dirEnv = phaseCPUProfileDirEnv
		captureAllocationSnapshots = false
		captureAllocationProfiles = false
	}
	if dir == "" {
		dir = allocationOnlyDir
		dirEnv = phaseAllocProfileDirEnv
		captureAllocationSnapshots = true
		captureAllocationProfiles = true
		captureCPUProfiles = false
	}
	if dir == "" {
		dir = statsOnlyDir
		dirEnv = phaseStatsDirEnv
		captureAllocationSnapshots = true
		captureAllocationProfiles = false
		captureCPUProfiles = false
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("profilehook: resolve %s: %w", dirEnv, err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("profilehook: prepare %s: %w", dirEnv, err)
	}
	profiler := &PhaseProfiler{
		dir:                        abs,
		captureCPUProfiles:         captureCPUProfiles,
		captureAllocationSnapshots: captureAllocationSnapshots,
		captureAllocationProfiles:  captureAllocationProfiles,
	}
	if !captureAllocationProfiles {
		return profiler, nil
	}
	// Boundary allocation profiles must be exact for subtraction to be useful.
	// The default sampled profile can add old samples only after a later GC,
	// making a start/end diff appear to attribute package initialization to the
	// active phase. This is opt-in diagnostic instrumentation, so use exact
	// allocation sampling while this profiler is active and restore the process
	// setting from Stop.
	profiler.previousMemProfileRate = runtime.MemProfileRate
	runtime.MemProfileRate = 1
	return profiler, nil
}

// StartBootstrap begins the launcher/bootstrap phase profile.
func (p *PhaseProfiler) StartBootstrap() error {
	if p == nil {
		return nil
	}
	if p.currentPhase != "" {
		return fmt.Errorf("profilehook: bootstrap phase already started")
	}
	return p.start("bootstrap")
}

// StartMain closes the bootstrap profile and begins the registered main phase
// profile. Generated launchers call it immediately before RunRegisteredMain.
func (p *PhaseProfiler) StartMain() error {
	if p == nil {
		return nil
	}
	if p.currentPhase != "bootstrap" {
		return fmt.Errorf("profilehook: main phase requires an active bootstrap phase")
	}
	return p.start("main")
}

// Stop closes the active phase profile. It is safe to call after an early
// launcher error, where only the bootstrap phase may be active.
func (p *PhaseProfiler) Stop() error {
	if p == nil || p.stopped {
		return nil
	}
	p.stopped = true
	if p.captureAllocationProfiles {
		defer p.restoreMemProfileRate()
	}
	stopErr := p.stopCurrent()
	statsErr := p.writeStats()
	if stopErr != nil {
		return stopErr
	}
	return statsErr
}

func (p *PhaseProfiler) start(phase string) error {
	if p.stopped {
		return fmt.Errorf("profilehook: start %s phase after stop", phase)
	}
	if err := p.stopCurrent(); err != nil {
		return err
	}
	if p.captureAllocationProfiles {
		if err := p.writeAllocationSnapshot(phase, "start"); err != nil {
			return err
		}
	}
	var start phaseAllocationSnapshot
	if p.captureAllocationSnapshots {
		start = readPhaseAllocationSnapshot()
	}
	var file *os.File
	if p.captureCPUProfiles {
		path := filepath.Join(p.dir, phase+".cpu.pprof")
		var err error
		file, err = os.Create(path)
		if err != nil {
			return fmt.Errorf("profilehook: create %s phase: %w", phase, err)
		}
		if err := pprof.StartCPUProfile(file); err != nil {
			_ = file.Close()
			return fmt.Errorf("profilehook: start %s phase: %w", phase, err)
		}
	}
	p.currentFile = file
	p.currentPhase = phase
	p.currentStart = start
	return nil
}

func (p *PhaseProfiler) stopCurrent() error {
	if p.currentPhase == "" {
		return nil
	}
	phase := p.currentPhase
	start := p.currentStart
	if p.captureCPUProfiles {
		pprof.StopCPUProfile()
	}
	var end phaseAllocationSnapshot
	var snapshotErr error
	if p.captureAllocationProfiles {
		end = readPhaseAllocationSnapshot()
		snapshotErr = p.writeAllocationSnapshot(phase, "end")
	} else if p.captureAllocationSnapshots {
		end = readPhaseAllocationSnapshot()
	}
	var closeErr error
	if p.currentFile != nil {
		closeErr = p.currentFile.Close()
	}
	p.currentFile = nil
	p.currentPhase = ""
	p.currentStart = phaseAllocationSnapshot{}
	if p.captureAllocationSnapshots {
		p.phases = append(p.phases, phaseAllocationDelta(phase, start, end))
	}
	if snapshotErr != nil {
		return snapshotErr
	}
	if closeErr != nil {
		return fmt.Errorf("profilehook: close phase profile: %w", closeErr)
	}
	return nil
}

func (p *PhaseProfiler) writeAllocationSnapshot(phase string, boundary string) error {
	if p == nil {
		return nil
	}
	// Runtime allocation profiles are updated during sweeping and may otherwise
	// lag by up to two GC cycles. Drain that lag at both boundaries so the
	// cumulative start/end profiles are comparable.
	runtime.GC()
	runtime.GC()
	profile := pprof.Lookup("allocs")
	if profile == nil {
		return fmt.Errorf("profilehook: lookup allocs profile")
	}
	path := filepath.Join(p.dir, phase+"-"+boundary+".allocs.pprof")
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("profilehook: create %s %s allocation snapshot: %w", phase, boundary, err)
	}
	if err := profile.WriteTo(file, 0); err != nil {
		_ = file.Close()
		return fmt.Errorf("profilehook: write %s %s allocation snapshot: %w", phase, boundary, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("profilehook: close %s %s allocation snapshot: %w", phase, boundary, err)
	}
	return nil
}

func (p *PhaseProfiler) restoreMemProfileRate() {
	if p == nil {
		return
	}
	runtime.MemProfileRate = p.previousMemProfileRate
}

func readPhaseAllocationSnapshot() phaseAllocationSnapshot {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return phaseAllocationSnapshot{
		totalAlloc:  stats.TotalAlloc,
		mallocs:     stats.Mallocs,
		frees:       stats.Frees,
		heapAlloc:   stats.HeapAlloc,
		heapObjects: stats.HeapObjects,
		numGC:       stats.NumGC,
	}
}

func phaseAllocationDelta(phase string, start phaseAllocationSnapshot, end phaseAllocationSnapshot) PhaseAllocationStats {
	return PhaseAllocationStats{
		Phase:              phase,
		AllocatedBytes:     counterDelta(start.totalAlloc, end.totalAlloc),
		Allocations:        counterDelta(start.mallocs, end.mallocs),
		Frees:              counterDelta(start.frees, end.frees),
		HeapAllocatedDelta: signedCounterDelta(start.heapAlloc, end.heapAlloc),
		HeapObjectsDelta:   signedCounterDelta(start.heapObjects, end.heapObjects),
		GCCount:            end.numGC - start.numGC,
	}
}

func counterDelta(start uint64, end uint64) uint64 {
	if end < start {
		return 0
	}
	return end - start
}

func signedCounterDelta(start uint64, end uint64) int64 {
	if end >= start {
		return int64(end - start)
	}
	return -int64(start - end)
}

func (p *PhaseProfiler) writeStats() error {
	if p == nil || !p.captureAllocationSnapshots {
		return nil
	}
	payload, err := json.MarshalIndent(phaseProfileStats{Version: 1, Phases: p.phases}, "", "  ")
	if err != nil {
		return fmt.Errorf("profilehook: encode phase allocation stats: %w", err)
	}
	path := filepath.Join(p.dir, "phase-stats.json")
	if err := os.WriteFile(path, append(payload, '\n'), 0o644); err != nil {
		return fmt.Errorf("profilehook: write phase allocation stats: %w", err)
	}
	return nil
}

// StartFromEnv enables optional Go CPU, heap, and allocation profiling when
// the matching env vars are set. It is inert when all profile env vars are empty.
func StartFromEnv() (func() error, error) {
	cpuPath := strings.TrimSpace(os.Getenv(cpuProfileEnv))
	memPath := strings.TrimSpace(os.Getenv(memProfileEnv))
	allocPath := strings.TrimSpace(os.Getenv(allocProfileEnv))
	if cpuPath == "" && memPath == "" && allocPath == "" {
		return nil, nil
	}

	var (
		cpuFile    *os.File
		stopOnce   sync.Once
		stopErr    error
		interrupts chan os.Signal
	)

	if cpuPath != "" {
		if err := ensureParentDir(cpuPath); err != nil {
			return nil, fmt.Errorf("profilehook: prepare %s: %w", cpuProfileEnv, err)
		}
		file, err := os.Create(cpuPath)
		if err != nil {
			return nil, fmt.Errorf("profilehook: create %s: %w", cpuProfileEnv, err)
		}
		if err := pprof.StartCPUProfile(file); err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("profilehook: start cpu profile: %w", err)
		}
		cpuFile = file
	}

	stop := func() error {
		stopOnce.Do(func() {
			if interrupts != nil {
				signal.Stop(interrupts)
				close(interrupts)
			}
			stopErr = stopProfiles(cpuFile, memPath, allocPath)
		})
		return stopErr
	}

	interrupts = make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	go func() {
		_, ok := <-interrupts
		if !ok {
			return
		}
		_ = stop()
		runStopHooks()
		os.Exit(130)
	}()

	return stop, nil
}

func stopProfiles(cpuFile *os.File, memPath string, allocPath string) error {
	if cpuFile != nil {
		pprof.StopCPUProfile()
		if err := cpuFile.Close(); err != nil {
			return fmt.Errorf("profilehook: close cpu profile: %w", err)
		}
	}
	if allocPath != "" {
		if err := writeNamedProfile(allocProfileEnv, allocPath, "allocs"); err != nil {
			return err
		}
	}
	if memPath != "" {
		if err := ensureParentDir(memPath); err != nil {
			return fmt.Errorf("profilehook: prepare %s: %w", memProfileEnv, err)
		}
		file, err := os.Create(memPath)
		if err != nil {
			return fmt.Errorf("profilehook: create %s: %w", memProfileEnv, err)
		}
		runtime.GC()
		if err := pprof.WriteHeapProfile(file); err != nil {
			_ = file.Close()
			return fmt.Errorf("profilehook: write heap profile: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("profilehook: close heap profile: %w", err)
		}
	}
	return nil
}

func writeNamedProfile(envName string, path string, profileName string) error {
	if err := ensureParentDir(path); err != nil {
		return fmt.Errorf("profilehook: prepare %s: %w", envName, err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("profilehook: create %s: %w", envName, err)
	}
	profile := pprof.Lookup(profileName)
	if profile == nil {
		_ = file.Close()
		return fmt.Errorf("profilehook: lookup %s profile", profileName)
	}
	if err := profile.WriteTo(file, 0); err != nil {
		_ = file.Close()
		return fmt.Errorf("profilehook: write %s profile: %w", profileName, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("profilehook: close %s profile: %w", profileName, err)
	}
	return nil
}

func ensureParentDir(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Dir(abs), 0o755)
}
