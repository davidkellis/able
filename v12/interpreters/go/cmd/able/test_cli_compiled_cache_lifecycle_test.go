package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCompiledTestCacheLifecycleClassifiesAndPrunesByBytes(t *testing.T) {
	root := t.TempDir()
	cache := &compiledTestCache{root: root}
	oldKey := strings.Repeat("a", 64)
	newKey := strings.Repeat("b", 64)
	publishLifecycleTestEntry(t, cache, oldKey, 128)
	publishLifecycleTestEntry(t, cache, newKey, 256)

	now := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	setLifecycleEntryTime(t, cache, oldKey, now.Add(-48*time.Hour))
	setLifecycleEntryTime(t, cache, newKey, now.Add(-time.Hour))
	writeLifecycleCorruptEntry(t, cache, strings.Repeat("c", 64))
	writeLifecycleFile(t,
		filepath.Join(root, compiledTestCacheSchema, ".publish-interrupted", "partial"),
		"partial",
	)
	writeLifecycleFile(t,
		filepath.Join(root, "able-compiled-test-v0", "old-entry", "able-test"),
		"old",
	)
	writeLifecycleFile(t, filepath.Join(root, "leave-me-alone.txt"), "unknown")

	before, err := scanCompiledTestCache(cache, now)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	assertLifecycleInventoryCounts(t, before, 2, 1, 1, 1, 1)
	newEntry := lifecycleInventoryEntry(t, before, compiledTestCacheEntryValid, newKey)

	options := compiledTestCachePruneOptions{
		MaxBytes:    newEntry.SizeBytes,
		MaxBytesSet: true,
		DryRun:      true,
		Now:         now,
	}
	preview, err := pruneCompiledTestCache(cache, options)
	if err != nil {
		t.Fatalf("dry-run prune: %v", err)
	}
	if preview.Busy {
		t.Fatal("dry-run prune unexpectedly reported busy")
	}
	if preview.RemovedEntries != 4 {
		t.Fatalf("dry-run removed entries = %d, want 4", preview.RemovedEntries)
	}
	if preview.RetainedValid != 1 || preview.RetainedValidBytes != newEntry.SizeBytes {
		t.Fatalf("dry-run retained valid = %d/%d, want 1/%d",
			preview.RetainedValid, preview.RetainedValidBytes, newEntry.SizeBytes)
	}
	if _, err := os.Stat(filepath.Join(root, compiledTestCacheSchema, oldKey)); err != nil {
		t.Fatalf("dry-run removed old valid entry: %v", err)
	}

	options.DryRun = false
	pruned, err := pruneCompiledTestCache(cache, options)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if pruned.RemovedEntries != 4 || pruned.RetainedValid != 1 {
		t.Fatalf("prune result = removed %d retained %d, want 4/1",
			pruned.RemovedEntries, pruned.RetainedValid)
	}
	after, err := scanCompiledTestCache(cache, now)
	if err != nil {
		t.Fatalf("scan after prune: %v", err)
	}
	assertLifecycleInventoryCounts(t, after, 1, 0, 0, 0, 1)
	if _, err := os.Stat(filepath.Join(root, compiledTestCacheSchema, newKey)); err != nil {
		t.Fatalf("new valid entry was pruned: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "leave-me-alone.txt")); err != nil {
		t.Fatalf("unknown root file was pruned: %v", err)
	}
}

func TestCompiledTestCacheLifecyclePrunesByAgeAndRefreshesUsage(t *testing.T) {
	root := t.TempDir()
	cache := &compiledTestCache{root: root}
	oldKey := strings.Repeat("d", 64)
	usedKey := strings.Repeat("e", 64)
	publishLifecycleTestEntry(t, cache, oldKey, 64)
	publishLifecycleTestEntry(t, cache, usedKey, 64)

	now := time.Now()
	oldTime := now.Add(-30 * 24 * time.Hour)
	setLifecycleEntryTime(t, cache, oldKey, oldTime)
	setLifecycleEntryTime(t, cache, usedKey, oldTime)
	if _, ok := cache.lookup(usedKey); !ok {
		t.Fatal("expected used entry lookup to succeed")
	}
	if err := cache.markUsed(usedKey); err != nil {
		t.Fatalf("mark used: %v", err)
	}

	result, err := pruneCompiledTestCache(cache, compiledTestCachePruneOptions{
		MaxAge:    7 * 24 * time.Hour,
		MaxAgeSet: true,
		Now:       now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("age prune: %v", err)
	}
	if result.RemovedEntries != 1 || result.RetainedValid != 1 {
		t.Fatalf("age prune = removed %d retained %d, want 1/1",
			result.RemovedEntries, result.RetainedValid)
	}
	if _, err := os.Stat(filepath.Join(root, compiledTestCacheSchema, oldKey)); !os.IsNotExist(err) {
		t.Fatalf("old entry still exists or stat failed unexpectedly: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, compiledTestCacheSchema, usedKey)); err != nil {
		t.Fatalf("recently used entry was pruned: %v", err)
	}
}

func TestCompiledTestCachePruneRefusesCrossProcessActiveUse(t *testing.T) {
	if os.Getenv("ABLE_TEST_CACHE_LOCK_HELPER") == "1" {
		runCompiledTestCacheLockHelper(t)
		return
	}
	root := t.TempDir()
	cache := &compiledTestCache{root: root}
	key := strings.Repeat("f", 64)
	publishLifecycleTestEntry(t, cache, key, 128)
	readyPath := filepath.Join(root, "helper.ready")
	releasePath := filepath.Join(root, "helper.release")

	command := exec.Command(os.Args[0], "-test.run=^TestCompiledTestCachePruneRefusesCrossProcessActiveUse$")
	command.Env = append(os.Environ(),
		"ABLE_TEST_CACHE_LOCK_HELPER=1",
		"ABLE_TEST_CACHE_LOCK_ROOT="+root,
		"ABLE_TEST_CACHE_LOCK_READY="+readyPath,
		"ABLE_TEST_CACHE_LOCK_RELEASE="+releasePath,
	)
	var helperOutput bytes.Buffer
	command.Stdout = &helperOutput
	command.Stderr = &helperOutput
	if err := command.Start(); err != nil {
		t.Fatalf("start lock helper: %v", err)
	}
	helperExited := false
	t.Cleanup(func() {
		if !helperExited && command.Process != nil {
			_ = command.Process.Kill()
			_ = command.Wait()
		}
	})
	waitForLifecyclePath(t, readyPath, 5*time.Second)

	busy, err := pruneCompiledTestCache(cache, compiledTestCachePruneOptions{
		MaxBytes:    0,
		MaxBytesSet: true,
	})
	if err != nil {
		t.Fatalf("busy prune: %v", err)
	}
	if !busy.Busy || busy.RemovedEntries != 0 {
		t.Fatalf("busy prune = busy %v removed %d, want true/0", busy.Busy, busy.RemovedEntries)
	}
	if _, err := os.Stat(filepath.Join(root, compiledTestCacheSchema, key)); err != nil {
		t.Fatalf("busy prune removed active entry: %v", err)
	}

	writeLifecycleFile(t, releasePath, "release")
	if err := command.Wait(); err != nil {
		t.Fatalf("lock helper: %v\n%s", err, helperOutput.String())
	}
	helperExited = true
	pruned, err := pruneCompiledTestCache(cache, compiledTestCachePruneOptions{
		MaxBytes:    0,
		MaxBytesSet: true,
	})
	if err != nil {
		t.Fatalf("post-release prune: %v", err)
	}
	if pruned.Busy || pruned.RemovedEntries != 1 || pruned.RetainedValid != 0 {
		t.Fatalf("post-release prune = busy %v removed %d retained %d, want false/1/0",
			pruned.Busy, pruned.RemovedEntries, pruned.RetainedValid)
	}
}

func TestCompiledTestCacheLifecycleCLIInspectAndDryRunJSON(t *testing.T) {
	root := t.TempDir()
	cache := &compiledTestCache{root: root}
	key := strings.Repeat("1", 64)
	publishLifecycleTestEntry(t, cache, key, 32)

	code, stdout, stderr := captureCLI(t, []string{
		"cache", "compiled-tests", "inspect", "--dir", root, "--json",
	})
	if code != 0 || strings.TrimSpace(stderr) != "" {
		t.Fatalf("inspect = code %d stderr %q", code, stderr)
	}
	var inventory compiledTestCacheInventory
	if err := json.Unmarshal([]byte(stdout), &inventory); err != nil {
		t.Fatalf("decode inspect JSON: %v\n%s", err, stdout)
	}
	if inventory.ValidEntries != 1 || inventory.Root != cleanAbsolutePath(root) {
		t.Fatalf("inspect inventory = valid %d root %q", inventory.ValidEntries, inventory.Root)
	}

	code, stdout, stderr = captureCLI(t, []string{
		"cache", "compiled-tests", "prune", "--dir", root,
		"--max-bytes", "0", "--dry-run", "--json",
	})
	if code != 0 || strings.TrimSpace(stderr) != "" {
		t.Fatalf("dry-run prune = code %d stderr %q", code, stderr)
	}
	var result compiledTestCachePruneResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode prune JSON: %v\n%s", err, stdout)
	}
	if !result.DryRun || result.RemovedEntries != 1 || result.RetainedValid != 0 {
		t.Fatalf("dry-run result = dry %v removed %d retained %d",
			result.DryRun, result.RemovedEntries, result.RetainedValid)
	}
	if _, err := os.Stat(filepath.Join(root, compiledTestCacheSchema, key)); err != nil {
		t.Fatalf("CLI dry-run removed entry: %v", err)
	}
}

func TestCompiledTestCacheLifecycleParsesBounds(t *testing.T) {
	for _, testCase := range []struct {
		raw  string
		want int64
	}{
		{raw: "1024", want: 1024},
		{raw: "1KiB", want: 1 << 10},
		{raw: "1.5MiB", want: 1572864},
		{raw: "2GB", want: 2e9},
	} {
		got, err := parseCompiledTestCacheByteSize(testCase.raw)
		if err != nil || got != testCase.want {
			t.Fatalf("parse bytes %q = %d, %v; want %d", testCase.raw, got, err, testCase.want)
		}
	}
	if _, err := parseCompiledTestCacheByteSize("-1"); err == nil {
		t.Fatal("negative byte size unexpectedly succeeded")
	}
	if got, err := parseCompiledTestCacheDuration("7d"); err != nil || got != 7*24*time.Hour {
		t.Fatalf("parse 7d = %v, %v", got, err)
	}
	if got, err := parseCompiledTestCacheDuration("90m"); err != nil || got != 90*time.Minute {
		t.Fatalf("parse 90m = %v, %v", got, err)
	}
}

func TestCompiledTestCacheLifecycleRefusesOutsideRemoval(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside")
	if err := removeCompiledTestCacheManagedPath(root, outside); err == nil {
		t.Fatal("outside removal unexpectedly succeeded")
	}
}

func runCompiledTestCacheLockHelper(t *testing.T) {
	root := os.Getenv("ABLE_TEST_CACHE_LOCK_ROOT")
	readyPath := os.Getenv("ABLE_TEST_CACHE_LOCK_READY")
	releasePath := os.Getenv("ABLE_TEST_CACHE_LOCK_RELEASE")
	lock, acquired, err := acquireCompiledTestCacheFileLock(root, false, false)
	if err != nil || !acquired {
		t.Fatalf("helper acquire shared lock = %v, %v", acquired, err)
	}
	defer func() { _ = lock.release() }()
	if err := os.WriteFile(readyPath, []byte("ready"), 0o600); err != nil {
		t.Fatalf("helper ready: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(releasePath); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("helper timed out waiting for release")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func publishLifecycleTestEntry(t *testing.T, cache *compiledTestCache, key string, size int) {
	t.Helper()
	source := filepath.Join(t.TempDir(), "able-test")
	if err := os.WriteFile(source, []byte(strings.Repeat("x", size)), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.publish(key, source); err != nil {
		t.Fatalf("publish %s: %v", key, err)
	}
}

func setLifecycleEntryTime(t *testing.T, cache *compiledTestCache, key string, timestamp time.Time) {
	t.Helper()
	entryDir := filepath.Join(cache.root, compiledTestCacheSchema, key)
	if err := os.Chtimes(entryDir, timestamp, timestamp); err != nil {
		t.Fatalf("set entry time: %v", err)
	}
}

func writeLifecycleCorruptEntry(t *testing.T, cache *compiledTestCache, key string) {
	t.Helper()
	writeLifecycleFile(t,
		filepath.Join(cache.root, compiledTestCacheSchema, key, "manifest.json"),
		"{not-json",
	)
	writeLifecycleFile(t,
		filepath.Join(cache.root, compiledTestCacheSchema, key, "able-test"),
		"corrupt",
	)
}

func writeLifecycleFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func lifecycleInventoryEntry(
	t *testing.T,
	inventory compiledTestCacheInventory,
	class compiledTestCacheEntryClass,
	name string,
) compiledTestCacheEntryInfo {
	t.Helper()
	for _, entry := range inventory.Entries {
		if entry.Class == class && strings.HasSuffix(entry.Path, name) {
			return entry
		}
	}
	t.Fatalf("missing %s entry ending in %q", class, name)
	return compiledTestCacheEntryInfo{}
}

func assertLifecycleInventoryCounts(
	t *testing.T,
	inventory compiledTestCacheInventory,
	valid int,
	corrupt int,
	staging int,
	obsolete int,
	unknown int,
) {
	t.Helper()
	if inventory.ValidEntries != valid ||
		inventory.CorruptEntries != corrupt ||
		inventory.StagingEntries != staging ||
		inventory.ObsoleteEntries != obsolete ||
		inventory.UnknownEntries != unknown {
		t.Fatalf(
			"inventory counts = valid %d corrupt %d staging %d obsolete %d unknown %d; want %d/%d/%d/%d/%d",
			inventory.ValidEntries,
			inventory.CorruptEntries,
			inventory.StagingEntries,
			inventory.ObsoleteEntries,
			inventory.UnknownEntries,
			valid,
			corrupt,
			staging,
			obsolete,
			unknown,
		)
	}
}

func waitForLifecyclePath(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", path)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
