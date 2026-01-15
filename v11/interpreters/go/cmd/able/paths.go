package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/driver"
)

func collectSearchPaths(base string, extra ...string) []driver.SearchPath {
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

	for _, path := range extra {
		add(path, driver.RootUser)
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

	for _, path := range collectKernelPaths(base) {
		add(path, driver.RootStdlib)
	}

	for _, path := range collectStdlibPaths(base) {
		add(path, driver.RootStdlib)
	}

	return paths
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
	}

	for _, entry := range splitPathListEnv(os.Getenv("ABLE_MODULE_PATHS")) {
		add(entry)
	}
	for _, entry := range splitPathListEnv(os.Getenv("ABLE_PATH")) {
		add(entry)
	}

	// Discover bundled kernel relative to the working directory.
	if cwd, err := os.Getwd(); err == nil {
		for _, found := range findKernelRoots(cwd) {
			add(found)
		}
	}

	// Also probe relative to the executable for installed builds.
	if exe, err := os.Executable(); err == nil {
		for _, found := range findKernelRoots(filepath.Dir(exe)) {
			add(found)
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

	for _, entry := range splitPathListEnv(os.Getenv("ABLE_MODULE_PATHS")) {
		add(entry)
	}
	for _, entry := range splitPathListEnv(os.Getenv("ABLE_PATH")) {
		add(entry)
	}

	// Discover bundled stdlib relative to the entry directory first.
	if base != "" {
		for _, found := range findStdlibRoots(base) {
			add(found)
		}
	}

	// Discover bundled stdlib relative to the working directory.
	if cwd, err := os.Getwd(); err == nil {
		for _, found := range findStdlibRoots(cwd) {
			add(found)
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
			filepath.Join(dir, "v11", "kernel", "src"),
			filepath.Join(dir, "ablekernel", "src"),
			filepath.Join(dir, "able_kernel", "src"),
		} {
			add(candidate)
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
			filepath.Join(dir, "v11", "stdlib", "src"),
			filepath.Join(dir, "stdlib", "v11", "src"),
			filepath.Join(dir, "able-stdlib", "src"),
			filepath.Join(dir, "able_stdlib", "src"),
		} {
			add(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return roots
}
