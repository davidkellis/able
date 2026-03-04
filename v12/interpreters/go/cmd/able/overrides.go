package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// loadGlobalOverrides reads $ABLE_HOME/overrides.yml.
// Returns an empty map if the file doesn't exist.
func loadGlobalOverrides() map[string]string {
	home, err := resolveAbleHome()
	if err != nil {
		return nil
	}
	return loadGlobalOverridesFrom(filepath.Join(home, "overrides.yml"))
}

func loadGlobalOverridesFrom(path string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	overrides := make(map[string]string)
	if err := yaml.Unmarshal(data, &overrides); err != nil {
		return nil
	}
	// Normalize all keys on load.
	normalized := make(map[string]string, len(overrides))
	for k, v := range overrides {
		normalized[normalizeGitURL(k)] = v
	}
	return normalized
}

// saveGlobalOverrides writes the override map to $ABLE_HOME/overrides.yml.
func saveGlobalOverrides(overrides map[string]string) error {
	home, err := resolveAbleHome()
	if err != nil {
		return err
	}
	return saveGlobalOverridesTo(overrides, filepath.Join(home, "overrides.yml"))
}

func saveGlobalOverridesTo(overrides map[string]string, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create override directory: %w", err)
	}
	// Sort keys for deterministic output.
	keys := make([]string, 0, len(overrides))
	for k := range overrides {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sorted := make(map[string]string, len(overrides))
	for _, k := range keys {
		sorted[k] = overrides[k]
	}
	data, err := yaml.Marshal(sorted)
	if err != nil {
		return fmt.Errorf("marshal overrides: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// sshGitPattern matches git@host:user/repo.git style URLs.
var sshGitPattern = regexp.MustCompile(`^git@([^:]+):(.+)$`)

// sshSchemePattern matches ssh://git@host/user/repo.git style URLs.
var sshSchemePattern = regexp.MustCompile(`^ssh://git@([^/]+)/(.+)$`)

// normalizeGitURL canonicalizes git URLs so that different forms of the
// same repo URL all match. The canonical form is https://host/path.git.
func normalizeGitURL(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return url
	}

	// git@github.com:user/repo.git → https://github.com/user/repo.git
	if m := sshGitPattern.FindStringSubmatch(url); m != nil {
		url = "https://" + m[1] + "/" + m[2]
	}

	// ssh://git@github.com/user/repo.git → https://github.com/user/repo.git
	if m := sshSchemePattern.FindStringSubmatch(url); m != nil {
		url = "https://" + m[1] + "/" + m[2]
	}

	// Ensure trailing .git
	if !strings.HasSuffix(url, ".git") {
		url += ".git"
	}

	return url
}

// resolvePackageSrcPath returns dir/src if it exists, otherwise dir.
func resolvePackageSrcPath(dir string) string {
	src := filepath.Join(dir, "src")
	if info, err := os.Stat(src); err == nil && info.IsDir() {
		return src
	}
	return dir
}
