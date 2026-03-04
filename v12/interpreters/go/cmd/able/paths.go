package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/driver"
)

const defaultStdlibVersion = "0.1.0"

type searchPathOptions struct {
	skipStdlibDiscovery bool // true when lockfile provides stdlib/kernel paths
}

func collectSearchPaths(base string, opts searchPathOptions, extra ...driver.SearchPath) []driver.SearchPath {
	seen := make(map[string]struct{})
	var paths []driver.SearchPath

	add := func(path string, kind driver.RootKind) {
		if path == "" {
			return
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return
		}
		if kind != driver.RootStdlib && looksLikeStdlibPathCLI(abs) {
			kind = driver.RootStdlib
		}
		if kind != driver.RootStdlib && looksLikeKernelPathCLI(abs) {
			kind = driver.RootStdlib
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			return
		}
		if _, ok := seen[abs]; ok {
			return
		}
		seen[abs] = struct{}{}
		paths = append(paths, driver.SearchPath{Path: abs, Kind: kind})
	}

	for _, sp := range extra {
		add(sp.Path, sp.Kind)
	}

	if base != "" {
		add(base, driver.RootUser)
	}

	if cwd, err := os.Getwd(); err == nil {
		add(cwd, driver.RootUser)
	}

	for _, part := range splitPathListEnv(os.Getenv("ABLE_PATH")) {
		add(part, driver.RootUser)
	}

	for _, part := range splitPathListEnv(os.Getenv("ABLE_MODULE_PATHS")) {
		add(part, driver.RootUser)
	}

	if !opts.skipStdlibDiscovery {
		overrides := loadGlobalOverrides()
		if stdlibPath, ok := overrides[normalizeGitURL(defaultStdlibGitURL)]; ok {
			add(resolvePackageSrcPath(stdlibPath), driver.RootStdlib)
		} else if cachedPath, err := ensureCachedStdlib(); err == nil {
			add(cachedPath, driver.RootStdlib)
		}
		// Kernel: always use filesystem walk / embedded (no overrides).
		for _, path := range collectKernelPaths(base) {
			add(path, driver.RootStdlib)
		}
	}

	return paths
}

// ensureCachedStdlib returns the path to the stdlib src directory in the
// global ABLE_HOME cache, downloading it via git if necessary.
func ensureCachedStdlib() (string, error) {
	cacheDir, err := resolveAbleHome()
	if err != nil {
		return "", fmt.Errorf("resolve ABLE_HOME for stdlib: %w", err)
	}
	srcDir := filepath.Join(cacheDir, "pkg", "src", "able", defaultStdlibVersion, "src")
	if info, err := os.Stat(srcDir); err == nil && info.IsDir() {
		return srcDir, nil
	}
	// Not cached yet — resolve via the dependency installer (override → cache → git fetch).
	installer := newDependencyInstaller(nil, cacheDir)
	installer.manifestRoot = cacheDir
	resolved, err := installer.resolveStdlibDependency(&driver.DependencySpec{Version: defaultStdlibVersion})
	if err != nil {
		return "", fmt.Errorf("auto-install stdlib %s: %w", defaultStdlibVersion, err)
	}
	if resolved == nil || resolved.pkg == nil {
		return "", fmt.Errorf("stdlib %s: resolver returned nil", defaultStdlibVersion)
	}
	// The resolved source may be a path: spec; extract the actual directory.
	// The path may point to the package root or the src/ subdirectory.
	if src := strings.TrimPrefix(resolved.pkg.Source, "path:"); src != "" {
		srcPath := resolvePackageSrcPath(src)
		if info, err := os.Stat(srcPath); err == nil && info.IsDir() {
			return srcPath, nil
		}
	}
	// Fall back to the expected cache location.
	if info, err := os.Stat(srcDir); err == nil && info.IsDir() {
		return srcDir, nil
	}
	return "", fmt.Errorf("stdlib %s: resolved but src directory missing", defaultStdlibVersion)
}

func splitPathListEnv(value string) []string {
	if value == "" {
		return nil
	}
	raw := strings.Split(value, string(os.PathListSeparator))
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func looksLikeStdlibPathCLI(path string) bool {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))
	for _, part := range parts {
		lower := strings.ToLower(part)
		if lower == "stdlib" || strings.HasPrefix(lower, "stdlib_") || lower == "able-stdlib" || lower == "able_stdlib" {
			return true
		}
	}
	return false
}

func looksLikeKernelPathCLI(path string) bool {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))
	for _, part := range parts {
		switch strings.ToLower(part) {
		case "kernel", "ablekernel", "able_kernel":
			return true
		}
	}
	return false
}

func collectKernelPaths(base string) []string {
	var paths []string

	add := func(candidate string) {
		if candidate == "" {
			return
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return
		}
		for _, existing := range paths {
			if existing == abs {
				return
			}
		}
		paths = append(paths, abs)
	}

	// Discover bundled kernel relative to the entry directory first.
	if base != "" {
		for _, found := range findKernelRoots(base) {
			add(found)
		}
		if len(paths) > 0 {
			return paths
		}
	}

	for _, entry := range splitPathListEnv(os.Getenv("ABLE_MODULE_PATHS")) {
		add(entry)
	}
	for _, entry := range splitPathListEnv(os.Getenv("ABLE_PATH")) {
		add(entry)
	}
	if len(paths) > 0 {
		return paths
	}

	// Discover bundled kernel relative to the working directory.
	if cwd, err := os.Getwd(); err == nil {
		for _, found := range findKernelRoots(cwd) {
			add(found)
		}
		if len(paths) > 0 {
			return paths
		}
	}

	// Also probe relative to the executable for installed builds.
	if exe, err := os.Executable(); err == nil {
		for _, found := range findKernelRoots(filepath.Dir(exe)) {
			add(found)
		}
	}

	// Fall back to embedded kernel extracted to ABLE_HOME cache.
	if len(paths) == 0 {
		if kernelPath, err := ensureEmbeddedKernel(); err == nil {
			add(kernelPath)
		}
	}

	return paths
}

func collectStdlibPaths(base string) []string {
	var paths []string

	add := func(candidate string) {
		if candidate == "" {
			return
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return
		}
		for _, existing := range paths {
			if existing == abs {
				return
			}
		}
		paths = append(paths, abs)
	}

	// Discover bundled stdlib relative to the entry directory first.
	if base != "" {
		for _, found := range findStdlibRoots(base) {
			add(found)
		}
		if len(paths) > 0 {
			return paths
		}
	}

	for _, entry := range splitPathListEnv(os.Getenv("ABLE_MODULE_PATHS")) {
		add(entry)
	}
	for _, entry := range splitPathListEnv(os.Getenv("ABLE_PATH")) {
		add(entry)
	}
	if len(paths) > 0 {
		return paths
	}

	// Discover bundled stdlib relative to the working directory.
	if cwd, err := os.Getwd(); err == nil {
		for _, found := range findStdlibRoots(cwd) {
			add(found)
		}
		if len(paths) > 0 {
			return paths
		}
	}

	// Also probe relative to the executable for installed builds.
	if exe, err := os.Executable(); err == nil {
		for _, found := range findStdlibRoots(filepath.Dir(exe)) {
			add(found)
		}
	}

	return paths
}

func resolveReplEntryPath(base string) (string, error) {
	// Try override first.
	overrides := loadGlobalOverrides()
	if stdlibPath, ok := overrides[normalizeGitURL(defaultStdlibGitURL)]; ok {
		candidate := filepath.Join(resolvePackageSrcPath(stdlibPath), "repl.able")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	// Try cached stdlib.
	if cachedPath, err := ensureCachedStdlib(); err == nil {
		candidate := filepath.Join(cachedPath, "repl.able")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	// Fall back to filesystem walk and env vars.
	for _, root := range collectStdlibPaths(base) {
		candidate := filepath.Join(root, "repl.able")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("unable to locate stdlib repl.able (set ABLE_PATH or ABLE_MODULE_PATHS)")
}

func findKernelRoots(start string) []string {
	var roots []string
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			roots = append(roots, candidate)
		}
	}

	dir := start
	for {
		for _, candidate := range []string{
			filepath.Join(dir, "kernel", "src"),
			filepath.Join(dir, "v12", "kernel", "src"),
			filepath.Join(dir, "ablekernel", "src"),
			filepath.Join(dir, "able_kernel", "src"),
		} {
			add(candidate)
		}
		if len(roots) > 0 {
			return roots
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return roots
}

func findStdlibRoots(start string) []string {
	var roots []string
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			roots = append(roots, candidate)
		}
	}

	dir := start
	for {
		for _, candidate := range []string{
			filepath.Join(dir, "stdlib", "src"),
			filepath.Join(dir, "able-stdlib", "src"),
			filepath.Join(dir, "able_stdlib", "src"),
		} {
			add(candidate)
		}
		if len(roots) > 0 {
			return roots
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return roots
}
