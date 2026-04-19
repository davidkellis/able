package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestNormalizeGitURL(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"https://github.com/user/repo.git", "https://github.com/user/repo.git"},
		{"https://github.com/user/repo", "https://github.com/user/repo.git"},
		{"git@github.com:user/repo.git", "https://github.com/user/repo.git"},
		{"git@github.com:user/repo", "https://github.com/user/repo.git"},
		{"ssh://git@github.com/user/repo.git", "https://github.com/user/repo.git"},
		{"ssh://git@github.com/user/repo", "https://github.com/user/repo.git"},
		{"  https://github.com/user/repo.git  ", "https://github.com/user/repo.git"},
	}
	for _, tc := range cases {
		got := normalizeGitURL(tc.input)
		if got != tc.want {
			t.Errorf("normalizeGitURL(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestOverrideAddRemoveList(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", root)

	// Create a fake package directory.
	pkgDir := filepath.Join(root, "mypkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(pkgDir, "package.yml"), "name: mypkg\nversion: 0.1.0\n")

	// Add an override.
	code := runOverrideAdd("https://github.com/someone/mypkg.git", pkgDir)
	if code != 0 {
		t.Fatalf("runOverrideAdd returned %d", code)
	}

	// Verify override file was created.
	overrides := loadGlobalOverrides()
	normalized := normalizeGitURL("https://github.com/someone/mypkg.git")
	if overrides[normalized] != pkgDir {
		t.Fatalf("expected override %s → %s, got %v", normalized, pkgDir, overrides)
	}

	// List should show it.
	code = runOverrideList()
	if code != 0 {
		t.Fatalf("runOverrideList returned %d", code)
	}

	// Remove it.
	code = runOverrideRemove("https://github.com/someone/mypkg.git")
	if code != 0 {
		t.Fatalf("runOverrideRemove returned %d", code)
	}

	overrides = loadGlobalOverrides()
	if _, ok := overrides[normalized]; ok {
		t.Fatalf("expected override to be removed, got %v", overrides)
	}

	// Remove again should fail.
	code = runOverrideRemove("https://github.com/someone/mypkg.git")
	if code == 0 {
		t.Fatalf("expected failure removing non-existent override")
	}
}

func TestOverrideAddSSHFormMatchesHTTPS(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", root)

	pkgDir := filepath.Join(root, "mypkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(pkgDir, "package.yml"), "name: mypkg\nversion: 0.1.0\n")

	// Add using SSH form.
	code := runOverrideAdd("git@github.com:someone/mypkg.git", pkgDir)
	if code != 0 {
		t.Fatalf("runOverrideAdd returned %d", code)
	}

	// Should be retrievable using HTTPS form.
	overrides := loadGlobalOverrides()
	httpsKey := normalizeGitURL("https://github.com/someone/mypkg")
	if overrides[httpsKey] != pkgDir {
		t.Fatalf("SSH override not matchable via HTTPS: %v", overrides)
	}
}

func TestOverrideAddValidatesPackageYml(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", root)

	// Directory without package.yml.
	emptyDir := filepath.Join(root, "empty")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	code := runOverrideAdd("https://github.com/someone/pkg.git", emptyDir)
	if code == 0 {
		t.Fatalf("expected failure when package.yml is missing")
	}
}

func TestDepsInstallRespectsGlobalOverride(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", filepath.Join(root, "home"))

	// Create the override package.
	overrideDir := filepath.Join(root, "override-pkg")
	if err := os.MkdirAll(filepath.Join(overrideDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir override: %v", err)
	}
	writeFile(t, filepath.Join(overrideDir, "package.yml"), "name: coolpkg\nversion: 0.5.0-dev\n")

	// Create a fake git repo URL for the override.
	gitURL := "https://github.com/someone/coolpkg.git"

	// Set up the override.
	code := runOverrideAdd(gitURL, overrideDir)
	if code != 0 {
		t.Fatalf("runOverrideAdd returned %d", code)
	}

	// Create the main app with a git dependency.
	appDir := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(appDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir app: %v", err)
	}
	writeFile(t, filepath.Join(appDir, "package.yml"), `
name: testapp
version: 0.1.0
dependencies:
  coolpkg:
    git: https://github.com/someone/coolpkg.git
`)

	manifest, err := driver.LoadManifest(filepath.Join(appDir, "package.yml"))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	lock := driver.NewLockfile(manifest.Name, cliToolVersion)
	installer := newDependencyInstaller(manifest, filepath.Join(root, "home"))

	changed, logs, err := installer.Install(lock)
	if err != nil {
		t.Fatalf("Install error: %v (logs: %v)", err, logs)
	}
	if !changed {
		t.Fatalf("expected lockfile change")
	}

	pkg := requireLockedPackage(t, lock.Packages, "coolpkg")
	if pkg.Version != "0.5.0-dev" {
		t.Fatalf("expected overridden version 0.5.0-dev, got %q", pkg.Version)
	}
	if !strings.HasPrefix(pkg.Source, "path:") {
		t.Fatalf("expected path source for override, got %q", pkg.Source)
	}

	// Check that logs mention the override.
	assertAnyTextContainsAll(t, logs, "using override", "coolpkg")
}

func TestDepsInstallStdlibRespectsGlobalOverride(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", filepath.Join(root, "home"))

	// Create a local stdlib.
	stdlibDir := filepath.Join(root, "local-stdlib")
	stdlibSrc := filepath.Join(stdlibDir, "src")
	if err := os.MkdirAll(stdlibSrc, 0o755); err != nil {
		t.Fatalf("mkdir stdlib: %v", err)
	}
	writeFile(t, filepath.Join(stdlibDir, "package.yml"), "name: able\nversion: 99.0.0-local\n")

	// Set up override for default stdlib URL.
	code := runOverrideAdd(defaultStdlibGitURL, stdlibDir)
	if code != 0 {
		t.Fatalf("runOverrideAdd returned %d", code)
	}

	// Create the app (no explicit stdlib dependency — it's implicit).
	appDir := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(appDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir app: %v", err)
	}
	writeFile(t, filepath.Join(appDir, "package.yml"), "name: testapp\nversion: 0.1.0\n")

	manifest, err := driver.LoadManifest(filepath.Join(appDir, "package.yml"))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	lock := driver.NewLockfile(manifest.Name, cliToolVersion)
	installer := newDependencyInstaller(manifest, filepath.Join(root, "home"))

	changed, logs, err := installer.Install(lock)
	if err != nil {
		t.Fatalf("Install error: %v (logs: %v)", err, logs)
	}
	if !changed {
		t.Fatalf("expected lockfile change")
	}

	stdlib := requireLockedPackage(t, lock.Packages, "able")
	if stdlib.Version != "99.0.0-local" {
		t.Fatalf("expected overridden stdlib version 99.0.0-local, got %q", stdlib.Version)
	}
	if !strings.HasPrefix(stdlib.Source, "path:") {
		t.Fatalf("expected path source for stdlib override, got %q", stdlib.Source)
	}

	assertAnyTextContainsAll(t, logs, "using override", "stdlib")
}

func TestDepsInstallProjectPathBeatsOverride(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", filepath.Join(root, "home"))

	// Set up a global override for a git URL.
	overrideDir := filepath.Join(root, "global-override")
	if err := os.MkdirAll(filepath.Join(overrideDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir override: %v", err)
	}
	writeFile(t, filepath.Join(overrideDir, "package.yml"), "name: mypkg\nversion: 1.0.0-global\n")

	code := runOverrideAdd("https://github.com/someone/mypkg.git", overrideDir)
	if code != 0 {
		t.Fatalf("runOverrideAdd returned %d", code)
	}

	// Set up a per-project path dependency that should win.
	localDir := filepath.Join(root, "local-dep")
	if err := os.MkdirAll(filepath.Join(localDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir local: %v", err)
	}
	writeFile(t, filepath.Join(localDir, "package.yml"), "name: mypkg\nversion: 2.0.0-local\n")

	// The app uses path: which should beat the global override.
	appDir := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(appDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir app: %v", err)
	}
	writeFile(t, filepath.Join(appDir, "package.yml"), `
name: testapp
version: 0.1.0
dependencies:
  mypkg:
    path: ../local-dep
`)

	manifest, err := driver.LoadManifest(filepath.Join(appDir, "package.yml"))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	lock := driver.NewLockfile(manifest.Name, cliToolVersion)
	installer := newDependencyInstaller(manifest, filepath.Join(root, "home"))

	_, _, err = installer.Install(lock)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}

	pkg := findLockedPackage(lock.Packages, "mypkg")
	if pkg == nil {
		t.Fatalf("missing mypkg in lock: %#v", lock.Packages)
	}
	// Per-project path should win — version is from local-dep, not global-override.
	if pkg.Version != "2.0.0-local" {
		t.Fatalf("expected local path version 2.0.0-local, got %q (global override should not apply)", pkg.Version)
	}
}

func TestCollectSearchPathsUsesStdlibOverride(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", filepath.Join(root, "home"))

	// Create a local stdlib.
	stdlibDir := filepath.Join(root, "my-stdlib")
	stdlibSrc := filepath.Join(stdlibDir, "src")
	if err := os.MkdirAll(stdlibSrc, 0o755); err != nil {
		t.Fatalf("mkdir stdlib: %v", err)
	}
	writeFile(t, filepath.Join(stdlibDir, "package.yml"), "name: able\nversion: 1.0.0\n")

	// Set up override for default stdlib URL.
	code := runOverrideAdd(defaultStdlibGitURL, stdlibDir)
	if code != 0 {
		t.Fatalf("runOverrideAdd returned %d", code)
	}

	// Collect search paths for an ad-hoc script (no lockfile → skipStdlibDiscovery=false).
	projectDir := filepath.Join(root, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	paths := collectSearchPaths(projectDir, searchPathOptions{})
	if !containsSearchPath(paths, stdlibSrc) {
		t.Fatalf("expected search paths to contain stdlib override %s, got %v", stdlibSrc, paths)
	}
}

func TestOverrideUpdateExistingEntry(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", root)

	// Create two package directories.
	pkgDir1 := filepath.Join(root, "pkg-v1")
	pkgDir2 := filepath.Join(root, "pkg-v2")
	for _, dir := range []string{pkgDir1, pkgDir2} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		writeFile(t, filepath.Join(dir, "package.yml"), "name: mypkg\nversion: 0.1.0\n")
	}

	gitURL := "https://github.com/someone/mypkg.git"

	// Add first override.
	code := runOverrideAdd(gitURL, pkgDir1)
	if code != 0 {
		t.Fatalf("first add returned %d", code)
	}
	overrides := loadGlobalOverrides()
	if overrides[normalizeGitURL(gitURL)] != pkgDir1 {
		t.Fatalf("expected first path, got %v", overrides)
	}

	// Re-add with different path — should update.
	code = runOverrideAdd(gitURL, pkgDir2)
	if code != 0 {
		t.Fatalf("second add returned %d", code)
	}
	overrides = loadGlobalOverrides()
	if overrides[normalizeGitURL(gitURL)] != pkgDir2 {
		t.Fatalf("expected updated path %s, got %v", pkgDir2, overrides)
	}
}

func TestOverrideCLIErrorPaths(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", root)

	// No subcommand.
	if code := runOverride(nil); code == 0 {
		t.Fatalf("expected failure with no subcommand")
	}

	// Unknown subcommand.
	if code := runOverride([]string{"bogus"}); code == 0 {
		t.Fatalf("expected failure with unknown subcommand")
	}

	// Add with wrong number of args.
	if code := runOverride([]string{"add"}); code == 0 {
		t.Fatalf("expected failure with missing add args")
	}
	if code := runOverride([]string{"add", "url-only"}); code == 0 {
		t.Fatalf("expected failure with one add arg")
	}

	// Remove with no args.
	if code := runOverride([]string{"remove"}); code == 0 {
		t.Fatalf("expected failure with missing remove args")
	}

	// List with extra args.
	if code := runOverride([]string{"list", "extra"}); code == 0 {
		t.Fatalf("expected failure with extra list args")
	}

	// Add with non-existent path.
	if code := runOverrideAdd("https://github.com/x/y.git", "/nonexistent/path"); code == 0 {
		t.Fatalf("expected failure with non-existent path")
	}

	// Add with file instead of directory.
	filePath := filepath.Join(root, "afile")
	writeFile(t, filePath, "not a dir")
	if code := runOverrideAdd("https://github.com/x/y.git", filePath); code == 0 {
		t.Fatalf("expected failure when path is a file")
	}
}

func TestOverrideAddResolvesRelativePath(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", root)

	pkgDir := filepath.Join(root, "mypkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(pkgDir, "package.yml"), "name: mypkg\nversion: 0.1.0\n")

	// Change to root so relative path works.
	enterWorkingDir(t, root)

	// Add with relative path.
	code := runOverrideAdd("https://github.com/someone/mypkg.git", "mypkg")
	if code != 0 {
		t.Fatalf("runOverrideAdd returned %d", code)
	}

	overrides := loadGlobalOverrides()
	resolved := overrides[normalizeGitURL("https://github.com/someone/mypkg.git")]
	if !filepath.IsAbs(resolved) {
		t.Fatalf("expected absolute path, got %q", resolved)
	}
	if resolved != pkgDir {
		t.Fatalf("expected %s, got %s", pkgDir, resolved)
	}
}

func TestOverrideListEmpty(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", root)

	// List with no overrides should succeed.
	code := runOverrideList()
	if code != 0 {
		t.Fatalf("expected success listing empty overrides, got %d", code)
	}
}

func TestLoadGlobalOverridesMissingFile(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", root)

	overrides := loadGlobalOverrides()
	if len(overrides) != 0 {
		t.Fatalf("expected empty overrides for missing file, got %v", overrides)
	}
}

func TestLoadGlobalOverridesCorruptFile(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", root)

	// Write invalid YAML.
	writeFile(t, filepath.Join(root, "overrides.yml"), "not: [valid: yaml: {{{")

	overrides := loadGlobalOverrides()
	if len(overrides) != 0 {
		t.Fatalf("expected empty overrides for corrupt file, got %v", overrides)
	}
}

func TestNormalizeGitURLEmpty(t *testing.T) {
	if got := normalizeGitURL(""); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
	if got := normalizeGitURL("   "); got != "" {
		t.Fatalf("expected empty string for whitespace, got %q", got)
	}
}

func TestOverrideDoesNotAffectKernel(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ABLE_HOME", filepath.Join(root, "home"))

	// Set up stdlib override so we don't need network access.
	stdlibDir := filepath.Join(root, "stdlib")
	if err := os.MkdirAll(filepath.Join(stdlibDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir stdlib: %v", err)
	}
	writeFile(t, filepath.Join(stdlibDir, "package.yml"), "name: able\nversion: 1.0.0\n")
	code := runOverrideAdd(defaultStdlibGitURL, stdlibDir)
	if code != 0 {
		t.Fatalf("runOverrideAdd stdlib returned %d", code)
	}

	// Create app.
	appDir := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(appDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir app: %v", err)
	}
	writeFile(t, filepath.Join(appDir, "package.yml"), "name: testapp\nversion: 0.1.0\n")

	manifest, err := driver.LoadManifest(filepath.Join(appDir, "package.yml"))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	lock := driver.NewLockfile(manifest.Name, cliToolVersion)
	installer := newDependencyInstaller(manifest, filepath.Join(root, "home"))

	_, _, err = installer.Install(lock)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}

	// Kernel should be resolved via filesystem walk or embedded, never via override.
	_, kernel := requireLockedStdlibAndKernel(t, lock.Packages)
	// Kernel source should NOT reference any override path.
	if strings.Contains(kernel.Source, "override") {
		t.Fatalf("kernel should not use overrides, got source %q", kernel.Source)
	}
}

func TestResolvePackageSrcPath(t *testing.T) {
	root := t.TempDir()

	// With src/ subdirectory.
	withSrc := filepath.Join(root, "with-src")
	if err := os.MkdirAll(filepath.Join(withSrc, "src"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if got := resolvePackageSrcPath(withSrc); got != filepath.Join(withSrc, "src") {
		t.Fatalf("expected %s/src, got %s", withSrc, got)
	}

	// Without src/ subdirectory.
	withoutSrc := filepath.Join(root, "without-src")
	if err := os.MkdirAll(withoutSrc, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if got := resolvePackageSrcPath(withoutSrc); got != withoutSrc {
		t.Fatalf("expected %s, got %s", withoutSrc, got)
	}
}
