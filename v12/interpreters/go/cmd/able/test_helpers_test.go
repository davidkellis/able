package main

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/driver"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func initGitRepo(t *testing.T, dir string) string {
	t.Helper()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("PlainInit: %v", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}
	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == filepath.Join(dir, ".git") {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, ".git/") {
			return nil
		}
		if _, err := worktree.Add(rel); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("stage files: %v", err)
	}
	hash, err := worktree.Commit("init", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Able CLI",
			Email: "able@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	return hash.String()
}

func containsPath(paths []string, target string) bool {
	for _, path := range paths {
		if path == target {
			return true
		}
	}
	return false
}

func containsSearchPath(paths []driver.SearchPath, target string) bool {
	for _, sp := range paths {
		if filepath.Clean(sp.Path) == filepath.Clean(target) {
			return true
		}
	}
	return false
}

func findLockedPackage(pkgs []*driver.LockedPackage, name string) *driver.LockedPackage {
	for _, pkg := range pkgs {
		if pkg != nil && pkg.Name == name {
			return pkg
		}
	}
	return nil
}

func requireLockedPackage(t *testing.T, pkgs []*driver.LockedPackage, name string) *driver.LockedPackage {
	t.Helper()

	pkg := findLockedPackage(pkgs, name)
	if pkg == nil {
		t.Fatalf("missing %s entry: %#v", name, pkgs)
	}
	return pkg
}

func requireLockedStdlibAndKernel(t *testing.T, pkgs []*driver.LockedPackage) (*driver.LockedPackage, *driver.LockedPackage) {
	t.Helper()

	return requireLockedPackage(t, pkgs, "able"), requireLockedPackage(t, pkgs, "kernel")
}

func repoStdlibPath(t *testing.T) string {
	t.Helper()
	// Prefer the checked-out canonical source for tests. Cache discovery may
	// bootstrap through git, which would make unrelated test runs network-bound.
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	base := filepath.Dir(current)
	repoRoot := filepath.Clean(filepath.Join(base, "..", "..", "..", "..", ".."))
	for _, candidate := range []string{
		filepath.Join(repoRoot, "able-stdlib", "src"),
		filepath.Join(filepath.Dir(repoRoot), "able-stdlib", "src"),
	} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	// A checkout is not always available (for example, downstream module
	// consumers); retain the runtime-like cache and override fallbacks there.
	if path, err := ensureCachedStdlib(); err == nil {
		return path
	}
	overrides := loadGlobalOverrides()
	if stdlibPath, ok := overrides[normalizeGitURL(defaultStdlibGitURL)]; ok {
		src := resolvePackageSrcPath(stdlibPath)
		if info, err := os.Stat(src); err == nil && info.IsDir() {
			return src
		}
	}
	t.Fatalf("stdlib path not found via cache, override, or sibling directory")
	return ""
}

func repoKernelPath(t *testing.T) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	base := filepath.Dir(current)
	repoRoot := filepath.Clean(filepath.Join(base, "..", "..", "..", "..", ".."))
	path := filepath.Join(repoRoot, "v12", "kernel", "src")
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		t.Fatalf("kernel path %s invalid: %v", path, err)
	}
	return path
}

const compiledCLIIntegrationEnv = "ABLE_RUN_COMPILED_CLI_INTEGRATION"

func compiledCLIExecutionRequiresIntegrationLane(args []string) bool {
	if len(args) == 0 || args[0] != "test" {
		return false
	}
	compiled := false
	for _, arg := range args[1:] {
		switch arg {
		case "--compiled":
			compiled = true
		case "--dry-run":
			return false
		}
	}
	return compiled
}

func compiledCLIIntegrationEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(compiledCLIIntegrationEnv))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func captureCLI(t *testing.T, args []string) (int, string, string) {
	t.Helper()
	if compiledCLIExecutionRequiresIntegrationLane(args) && !compiledCLIIntegrationEnabled() {
		t.Skipf("generated-Go CLI integration test; rerun with %s=1", compiledCLIIntegrationEnv)
	}

	stdout := os.Stdout
	stderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	code := run(args)

	if err := wOut.Close(); err != nil {
		t.Fatalf("stdout close: %v", err)
	}
	if err := wErr.Close(); err != nil {
		t.Fatalf("stderr close: %v", err)
	}

	os.Stdout = stdout
	os.Stderr = stderr

	outBytes, err := io.ReadAll(rOut)
	if err != nil {
		t.Fatalf("stdout read: %v", err)
	}
	errBytes, err := io.ReadAll(rErr)
	if err != nil {
		t.Fatalf("stderr read: %v", err)
	}

	if err := rOut.Close(); err != nil {
		t.Fatalf("stdout pipe close: %v", err)
	}
	if err := rErr.Close(); err != nil {
		t.Fatalf("stderr pipe close: %v", err)
	}

	return code, string(outBytes), string(errBytes)
}
