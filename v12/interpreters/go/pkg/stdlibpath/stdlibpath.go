package stdlibpath

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ResolveInstalledSrc prefers explicit stdlib overrides and installed cache
// entries over repository-relative discovery.
func ResolveInstalledSrc() string {
	if candidate := normalizeSrcDir(os.Getenv("ABLE_STDLIB_ROOT")); candidate != "" {
		return candidate
	}
	if candidate := resolveCachedSrc(resolveAbleHome()); candidate != "" {
		return candidate
	}
	return ""
}

// ResolveRepoOrInstalledSrc returns the first available stdlib source root
// using installed/cached state first, then local repository-style layouts.
func ResolveRepoOrInstalledSrc(repoRoot string) string {
	if candidate := ResolveInstalledSrc(); candidate != "" {
		return candidate
	}
	if candidate := FindSiblingSrc(repoRoot); candidate != "" {
		return candidate
	}
	return filepath.Join(repoRoot, "stdlib", "src")
}

// FindSiblingSrc probes local repository-style stdlib layouts relative to the
// provided root.
func FindSiblingSrc(repoRoot string) string {
	for _, candidate := range siblingCandidates(repoRoot) {
		if candidate := normalizeSrcDir(candidate); candidate != "" {
			return candidate
		}
	}
	return ""
}

func siblingCandidates(repoRoot string) []string {
	candidates := []string{
		filepath.Join(repoRoot, "stdlib", "src"),
		filepath.Join(repoRoot, "able-stdlib", "src"),
		filepath.Join(repoRoot, "able_stdlib", "src"),
	}
	if parent := filepath.Dir(repoRoot); parent != "" && parent != repoRoot {
		candidates = append(candidates,
			filepath.Join(parent, "able-stdlib", "src"),
			filepath.Join(parent, "able_stdlib", "src"),
		)
	}
	return candidates
}

func resolveAbleHome() string {
	if home := strings.TrimSpace(os.Getenv("ABLE_HOME")); home != "" {
		return home
	}
	userHome, err := os.UserHomeDir()
	if err != nil || userHome == "" {
		return ""
	}
	return filepath.Join(userHome, ".able")
}

func resolveCachedSrc(home string) string {
	if home == "" {
		return ""
	}
	cacheBase := filepath.Join(home, "pkg", "src", "able")
	entries, err := os.ReadDir(cacheBase)
	if err != nil {
		return ""
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	for i := len(names) - 1; i >= 0; i-- {
		if candidate := normalizeSrcDir(filepath.Join(cacheBase, names[i], "src")); candidate != "" {
			return candidate
		}
	}
	return ""
}

func normalizeSrcDir(candidate string) string {
	trimmed := strings.TrimSpace(candidate)
	if trimmed == "" {
		return ""
	}
	abs, err := filepath.Abs(trimmed)
	if err != nil {
		return ""
	}
	src := filepath.Join(abs, "src")
	if info, err := os.Stat(src); err == nil && info.IsDir() {
		return src
	}
	if info, err := os.Stat(abs); err == nil && info.IsDir() {
		return abs
	}
	return ""
}
