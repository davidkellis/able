package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/driver"
)

func loadManifestFrom(start string) (*driver.Manifest, error) {
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("resolve working directory: %w", err)
		}
		start = cwd
	}
	absStart, err := filepath.Abs(start)
	if err != nil {
		return nil, fmt.Errorf("resolve manifest search path %q: %w", start, err)
	}
	if info, statErr := os.Stat(absStart); statErr == nil && !info.IsDir() {
		absStart = filepath.Dir(absStart)
	}
	manifestPath, err := findManifest(absStart)
	if err != nil {
		return nil, err
	}
	manifest, err := driver.LoadManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func resolveTargetMain(manifest *driver.Manifest, target *driver.TargetSpec) (string, error) {
	if manifest == nil || target == nil {
		return "", fmt.Errorf("missing manifest or target")
	}
	mainPath := strings.TrimSpace(target.Main)
	if mainPath == "" {
		return "", fmt.Errorf("target %q missing main entrypoint", target.OriginalName)
	}
	if filepath.IsAbs(mainPath) {
		return filepath.Clean(mainPath), nil
	}
	base := filepath.Dir(manifest.Path)
	if base == "" {
		return filepath.Clean(filepath.FromSlash(mainPath)), nil
	}
	return filepath.Join(base, filepath.FromSlash(mainPath)), nil
}

func looksLikePathCandidate(arg string) bool {
	if arg == "" {
		return false
	}
	if strings.Contains(arg, string(os.PathSeparator)) {
		return true
	}
	// Support forward/backward slashes regardless of host OS.
	if strings.Contains(arg, "/") || strings.Contains(arg, "\\") {
		return true
	}
	if filepath.Ext(arg) == ".able" {
		return true
	}
	if strings.HasPrefix(arg, ".") {
		return true
	}
	return false
}

func findManifest(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve start directory %q: %w", start, err)
	}
	if info, statErr := os.Stat(dir); statErr == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	origin := dir
	for {
		candidate := filepath.Join(dir, "package.yml")
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no package.yml found from %s upwards: %w", origin, errManifestNotFound)
		}
		dir = parent
	}
}

func resolveAbleHome() (string, error) {
	if home := strings.TrimSpace(os.Getenv("ABLE_HOME")); home != "" {
		if abs, err := filepath.Abs(home); err == nil {
			return abs, nil
		} else {
			return "", fmt.Errorf("resolve ABLE_HOME %q: %w", home, err)
		}
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(userHome, ".able"), nil
}

func loadLockfileForManifest(manifest *driver.Manifest) (*driver.Lockfile, error) {
	if manifest == nil {
		return nil, nil
	}
	lockPath := filepath.Join(filepath.Dir(manifest.Path), "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if manifestHasDependencies(manifest) {
				return nil, fmt.Errorf("package.lock missing for %q; run `able deps install`", manifest.Name)
			}
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read lockfile %s: %w", lockPath, err)
	}
	if lock.Root != manifest.Name {
		return nil, fmt.Errorf("lockfile root %q does not match manifest name %q", lock.Root, manifest.Name)
	}
	return lock, nil
}

func manifestHasDependencies(manifest *driver.Manifest) bool {
	if manifest == nil {
		return false
	}
	return len(manifest.Dependencies) > 0 ||
		len(manifest.DevDependencies) > 0 ||
		len(manifest.BuildDependencies) > 0
}
