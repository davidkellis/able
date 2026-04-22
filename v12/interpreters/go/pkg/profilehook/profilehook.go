package profilehook

import (
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
	cpuProfileEnv = "ABLE_GO_CPU_PROFILE"
	memProfileEnv = "ABLE_GO_MEM_PROFILE"
)

// StartFromEnv enables optional Go CPU and heap profiling when the matching
// env vars are set. It is inert when both env vars are empty.
func StartFromEnv() (func() error, error) {
	cpuPath := strings.TrimSpace(os.Getenv(cpuProfileEnv))
	memPath := strings.TrimSpace(os.Getenv(memProfileEnv))
	if cpuPath == "" && memPath == "" {
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
			stopErr = stopProfiles(cpuFile, memPath)
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
		os.Exit(130)
	}()

	return stop, nil
}

func stopProfiles(cpuFile *os.File, memPath string) error {
	if cpuFile != nil {
		pprof.StopCPUProfile()
		if err := cpuFile.Close(); err != nil {
			return fmt.Errorf("profilehook: close cpu profile: %w", err)
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

func ensureParentDir(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Dir(abs), 0o755)
}
