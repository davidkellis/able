package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"able/interpreter-go/pkg/driver"
)

func resolveBuildPrecompileStdlibFromEnv() (bool, error) {
	raw, ok := os.LookupEnv("ABLE_BUILD_PRECOMPILE_STDLIB")
	if !ok {
		return false, nil
	}
	normalized := strings.TrimSpace(strings.ToLower(raw))
	switch normalized {
	case "", "0", "false", "no", "off":
		return false, nil
	case "1", "true", "yes", "on":
		return true, nil
	default:
		return false, fmt.Errorf("invalid ABLE_BUILD_PRECOMPILE_STDLIB value %q (expected one of: 1,true,yes,on,0,false,no,off)", raw)
	}
}

func discoverPrecompilePackages(searchPaths []driver.SearchPath, includeTests bool) ([]string, error) {
	roots := make(map[string]struct{})
	packages := make(map[string]struct{})
	for _, sp := range searchPaths {
		if sp.Path == "" || sp.Kind != driver.RootStdlib {
			continue
		}
		root := filepath.Clean(sp.Path)
		if _, ok := roots[root]; ok {
			continue
		}
		roots[root] = struct{}{}

		rootName, err := discoverBuildRootName(root)
		if err != nil {
			return nil, err
		}
		found, err := discoverPackagesUnderRoot(root, rootName, includeTests)
		if err != nil {
			return nil, err
		}
		for _, name := range found {
			packages[name] = struct{}{}
		}
	}

	out := make([]string, 0, len(packages))
	for name := range packages {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}

func discoverBuildRootName(root string) (string, error) {
	manifestPath, err := findManifest(root)
	switch {
	case err == nil:
		manifest, loadErr := driver.LoadManifest(manifestPath)
		if loadErr != nil {
			return "", loadErr
		}
		name := sanitizeBuildSegment(manifest.Name)
		if name != "" {
			return name, nil
		}
	case errors.Is(err, errManifestNotFound):
	default:
		return "", err
	}

	name := sanitizeBuildSegment(filepath.Base(root))
	if name == "" {
		name = "pkg"
	}
	return name, nil
}

func discoverPackagesUnderRoot(root string, rootName string, includeTests bool) ([]string, error) {
	out := make(map[string]struct{})
	base := []string{sanitizeBuildSegment(rootName)}
	if rootName == "kernel" || buildLooksLikeKernelPath(root) {
		base = []string{"able", "kernel"}
	}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if shouldSkipBuildPackageDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".able") {
			return nil
		}
		if !includeTests && strings.HasSuffix(name, ".test.able") {
			return nil
		}
		declared, err := readBuildDeclaredPackage(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relDir := filepath.Dir(filepath.ToSlash(rel))
		segments := append([]string{}, base...)
		if relDir != "." && relDir != "/" {
			for _, part := range strings.Split(relDir, "/") {
				part = sanitizeBuildSegment(part)
				if part == "" {
					continue
				}
				segments = append(segments, part)
			}
		}
		segments = append(segments, declared...)
		if len(segments) > 0 {
			out[strings.Join(segments, ".")] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	packages := make([]string, 0, len(out))
	for name := range out {
		packages = append(packages, name)
	}
	sort.Strings(packages)
	return packages, nil
}

func readBuildDeclaredPackage(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read package declaration %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "##") {
			continue
		}
		if !strings.HasPrefix(line, "package ") {
			return nil, nil
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, "package "))
		if rest == "" {
			return nil, nil
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			return nil, nil
		}
		name := strings.TrimSpace(fields[0])
		if strings.Contains(name, ".") {
			return nil, fmt.Errorf("package declaration must be unqualified in %s", path)
		}
		name = sanitizeBuildSegment(name)
		if name == "" {
			return nil, nil
		}
		return []string{name}, nil
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read package declaration %s: %w", path, err)
	}
	return nil, nil
}

func sanitizeBuildSegment(seg string) string {
	seg = strings.TrimSpace(seg)
	seg = strings.ReplaceAll(seg, "-", "_")
	return seg
}

func buildLooksLikeKernelPath(path string) bool {
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

func shouldSkipBuildPackageDir(name string) bool {
	switch name {
	case ".git", ".tmp", "tmp", "target", "node_modules", "vendor":
		return true
	default:
		return false
	}
}
