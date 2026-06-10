package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type compiledTestCacheEntryClass string

const (
	compiledTestCacheEntryValid    compiledTestCacheEntryClass = "valid"
	compiledTestCacheEntryCorrupt  compiledTestCacheEntryClass = "corrupt"
	compiledTestCacheEntryStaging  compiledTestCacheEntryClass = "staging"
	compiledTestCacheEntryObsolete compiledTestCacheEntryClass = "obsolete-schema"
	compiledTestCacheEntryUnknown  compiledTestCacheEntryClass = "unknown"
)

type compiledTestCacheEntryInfo struct {
	Path       string                      `json:"path"`
	Class      compiledTestCacheEntryClass `json:"class"`
	SizeBytes  int64                       `json:"size_bytes"`
	LastUsedAt time.Time                   `json:"last_used_at"`
	Reason     string                      `json:"reason,omitempty"`
	fullPath   string
}

type compiledTestCacheInventory struct {
	Root            string                       `json:"root"`
	Schema          string                       `json:"schema"`
	ScannedAt       time.Time                    `json:"scanned_at"`
	ValidEntries    int                          `json:"valid_entries"`
	ValidBytes      int64                        `json:"valid_bytes"`
	CorruptEntries  int                          `json:"corrupt_entries"`
	CorruptBytes    int64                        `json:"corrupt_bytes"`
	StagingEntries  int                          `json:"staging_entries"`
	StagingBytes    int64                        `json:"staging_bytes"`
	ObsoleteEntries int                          `json:"obsolete_schema_entries"`
	ObsoleteBytes   int64                        `json:"obsolete_schema_bytes"`
	UnknownEntries  int                          `json:"unknown_entries"`
	UnknownBytes    int64                        `json:"unknown_bytes"`
	Entries         []compiledTestCacheEntryInfo `json:"entries"`
}

type compiledTestCachePruneOptions struct {
	MaxBytes    int64
	MaxBytesSet bool
	MaxAge      time.Duration
	MaxAgeSet   bool
	DryRun      bool
	Now         time.Time
}

type compiledTestCachePruneResult struct {
	Root               string                       `json:"root"`
	DryRun             bool                         `json:"dry_run"`
	Busy               bool                         `json:"busy"`
	MaxBytes           *int64                       `json:"max_bytes,omitempty"`
	MaxAgeSeconds      *float64                     `json:"max_age_seconds,omitempty"`
	RemovedEntries     int                          `json:"removed_entries"`
	RemovedBytes       int64                        `json:"removed_bytes"`
	RetainedValid      int                          `json:"retained_valid_entries"`
	RetainedValidBytes int64                        `json:"retained_valid_bytes"`
	Entries            []compiledTestCacheEntryInfo `json:"entries"`
}

func inspectCompiledTestCache(cache *compiledTestCache) (compiledTestCacheInventory, error) {
	if cache == nil {
		return compiledTestCacheInventory{}, fmt.Errorf("compiled-test cache is not configured")
	}
	lock, acquired, err := acquireCompiledTestCacheFileLock(cache.root, false, false)
	if err != nil {
		return compiledTestCacheInventory{}, err
	}
	if !acquired {
		return compiledTestCacheInventory{}, fmt.Errorf("compiled-test cache: failed to acquire inspection lock")
	}
	defer func() { _ = lock.release() }()
	return scanCompiledTestCache(cache, time.Now())
}

func scanCompiledTestCache(
	cache *compiledTestCache,
	now time.Time,
) (compiledTestCacheInventory, error) {
	inventory := compiledTestCacheInventory{
		Root:      cache.root,
		Schema:    compiledTestCacheSchema,
		ScannedAt: now,
	}
	rootEntries, err := os.ReadDir(cache.root)
	if err != nil {
		return inventory, fmt.Errorf("compiled-test cache: inspect root: %w", err)
	}
	for _, rootEntry := range rootEntries {
		name := rootEntry.Name()
		if name == compiledTestCacheLockName {
			continue
		}
		fullPath := filepath.Join(cache.root, name)
		if name == compiledTestCacheSchema && rootEntry.IsDir() {
			entries, err := scanCompiledTestCacheCurrentSchema(cache, fullPath)
			if err != nil {
				return inventory, err
			}
			inventory.Entries = append(inventory.Entries, entries...)
			continue
		}
		class := compiledTestCacheEntryUnknown
		reason := "not managed by the compiled-test cache"
		if rootEntry.IsDir() && strings.HasPrefix(name, "able-compiled-test-") {
			class = compiledTestCacheEntryObsolete
			reason = "cache schema is not current"
		}
		entry, err := compiledTestCacheInventoryEntry(fullPath, name, class, reason)
		if err != nil {
			return inventory, err
		}
		inventory.Entries = append(inventory.Entries, entry)
	}
	sort.Slice(inventory.Entries, func(i, j int) bool {
		if inventory.Entries[i].Class != inventory.Entries[j].Class {
			return inventory.Entries[i].Class < inventory.Entries[j].Class
		}
		return inventory.Entries[i].Path < inventory.Entries[j].Path
	})
	summarizeCompiledTestCacheInventory(&inventory)
	return inventory, nil
}

func scanCompiledTestCacheCurrentSchema(
	cache *compiledTestCache,
	schemaRoot string,
) ([]compiledTestCacheEntryInfo, error) {
	children, err := os.ReadDir(schemaRoot)
	if err != nil {
		return nil, fmt.Errorf("compiled-test cache: inspect current schema: %w", err)
	}
	entries := make([]compiledTestCacheEntryInfo, 0, len(children))
	for _, child := range children {
		name := child.Name()
		fullPath := filepath.Join(schemaRoot, name)
		relativePath := filepath.ToSlash(filepath.Join(compiledTestCacheSchema, name))
		if strings.HasPrefix(name, ".publish-") {
			entry, err := compiledTestCacheInventoryEntry(
				fullPath,
				relativePath,
				compiledTestCacheEntryStaging,
				"incomplete atomic publication",
			)
			if err != nil {
				return nil, err
			}
			entries = append(entries, entry)
			continue
		}
		if !child.IsDir() {
			entry, err := compiledTestCacheInventoryEntry(
				fullPath,
				relativePath,
				compiledTestCacheEntryCorrupt,
				"current-schema entry is not a directory",
			)
			if err != nil {
				return nil, err
			}
			entries = append(entries, entry)
			continue
		}
		validation := cache.validateEntry(name)
		class := compiledTestCacheEntryValid
		reason := ""
		if !validation.Valid {
			class = compiledTestCacheEntryCorrupt
			reason = validation.Reason
		}
		entry, err := compiledTestCacheInventoryEntry(fullPath, relativePath, class, reason)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func compiledTestCacheInventoryEntry(
	fullPath string,
	relativePath string,
	class compiledTestCacheEntryClass,
	reason string,
) (compiledTestCacheEntryInfo, error) {
	info, err := os.Lstat(fullPath)
	if err != nil {
		return compiledTestCacheEntryInfo{}, fmt.Errorf("compiled-test cache: stat %s: %w", fullPath, err)
	}
	size, err := compiledTestCacheManagedPathSize(fullPath)
	if err != nil {
		return compiledTestCacheEntryInfo{}, err
	}
	return compiledTestCacheEntryInfo{
		Path:       filepath.ToSlash(relativePath),
		Class:      class,
		SizeBytes:  size,
		LastUsedAt: info.ModTime(),
		Reason:     reason,
		fullPath:   fullPath,
	}, nil
}

func compiledTestCacheManagedPathSize(root string) (int64, error) {
	var total int64
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			total += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("compiled-test cache: measure %s: %w", root, err)
	}
	return total, nil
}

func summarizeCompiledTestCacheInventory(inventory *compiledTestCacheInventory) {
	if inventory == nil {
		return
	}
	for _, entry := range inventory.Entries {
		switch entry.Class {
		case compiledTestCacheEntryValid:
			inventory.ValidEntries++
			inventory.ValidBytes += entry.SizeBytes
		case compiledTestCacheEntryCorrupt:
			inventory.CorruptEntries++
			inventory.CorruptBytes += entry.SizeBytes
		case compiledTestCacheEntryStaging:
			inventory.StagingEntries++
			inventory.StagingBytes += entry.SizeBytes
		case compiledTestCacheEntryObsolete:
			inventory.ObsoleteEntries++
			inventory.ObsoleteBytes += entry.SizeBytes
		case compiledTestCacheEntryUnknown:
			inventory.UnknownEntries++
			inventory.UnknownBytes += entry.SizeBytes
		}
	}
}

func pruneCompiledTestCache(
	cache *compiledTestCache,
	options compiledTestCachePruneOptions,
) (compiledTestCachePruneResult, error) {
	if cache == nil {
		return compiledTestCachePruneResult{}, fmt.Errorf("compiled-test cache is not configured")
	}
	result := compiledTestCachePruneResult{Root: cache.root, DryRun: options.DryRun}
	if options.MaxBytesSet {
		maxBytes := options.MaxBytes
		result.MaxBytes = &maxBytes
	}
	if options.MaxAgeSet {
		maxAgeSeconds := options.MaxAge.Seconds()
		result.MaxAgeSeconds = &maxAgeSeconds
	}
	lock, acquired, err := acquireCompiledTestCacheFileLock(cache.root, true, true)
	if err != nil {
		return result, err
	}
	if !acquired {
		result.Busy = true
		return result, nil
	}
	defer func() { _ = lock.release() }()

	now := options.Now
	if now.IsZero() {
		now = time.Now()
	}
	inventory, err := scanCompiledTestCache(cache, now)
	if err != nil {
		return result, err
	}
	candidates := selectCompiledTestCachePruneCandidates(inventory, options, now)
	for _, entry := range candidates {
		result.RemovedEntries++
		result.RemovedBytes += entry.SizeBytes
		result.Entries = append(result.Entries, entry)
		if options.DryRun {
			continue
		}
		if err := removeCompiledTestCacheManagedPath(cache.root, entry.fullPath); err != nil {
			return result, err
		}
	}
	if options.DryRun {
		result.RetainedValid, result.RetainedValidBytes = retainedCompiledTestCacheValid(inventory, candidates)
		return result, nil
	}
	after, err := scanCompiledTestCache(cache, now)
	if err != nil {
		return result, err
	}
	result.RetainedValid = after.ValidEntries
	result.RetainedValidBytes = after.ValidBytes
	return result, nil
}

func selectCompiledTestCachePruneCandidates(
	inventory compiledTestCacheInventory,
	options compiledTestCachePruneOptions,
	now time.Time,
) []compiledTestCacheEntryInfo {
	selected := make(map[string]struct{})
	var candidates []compiledTestCacheEntryInfo
	add := func(entry compiledTestCacheEntryInfo) {
		if _, exists := selected[entry.fullPath]; exists {
			return
		}
		selected[entry.fullPath] = struct{}{}
		candidates = append(candidates, entry)
	}
	var valid []compiledTestCacheEntryInfo
	var retainedBytes int64
	for _, entry := range inventory.Entries {
		switch entry.Class {
		case compiledTestCacheEntryCorrupt, compiledTestCacheEntryStaging, compiledTestCacheEntryObsolete:
			add(entry)
		case compiledTestCacheEntryValid:
			if options.MaxAgeSet && now.Sub(entry.LastUsedAt) > options.MaxAge {
				add(entry)
				continue
			}
			valid = append(valid, entry)
			retainedBytes += entry.SizeBytes
		}
	}
	if options.MaxBytesSet && retainedBytes > options.MaxBytes {
		sort.Slice(valid, func(i, j int) bool {
			if !valid[i].LastUsedAt.Equal(valid[j].LastUsedAt) {
				return valid[i].LastUsedAt.Before(valid[j].LastUsedAt)
			}
			return valid[i].Path < valid[j].Path
		})
		for _, entry := range valid {
			if retainedBytes <= options.MaxBytes {
				break
			}
			add(entry)
			retainedBytes -= entry.SizeBytes
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if !candidates[i].LastUsedAt.Equal(candidates[j].LastUsedAt) {
			return candidates[i].LastUsedAt.Before(candidates[j].LastUsedAt)
		}
		return candidates[i].Path < candidates[j].Path
	})
	return candidates
}

func retainedCompiledTestCacheValid(
	inventory compiledTestCacheInventory,
	candidates []compiledTestCacheEntryInfo,
) (int, int64) {
	removed := make(map[string]struct{}, len(candidates))
	for _, entry := range candidates {
		removed[entry.fullPath] = struct{}{}
	}
	var count int
	var bytes int64
	for _, entry := range inventory.Entries {
		if entry.Class != compiledTestCacheEntryValid {
			continue
		}
		if _, selected := removed[entry.fullPath]; selected {
			continue
		}
		count++
		bytes += entry.SizeBytes
	}
	return count, bytes
}

func removeCompiledTestCacheManagedPath(root, target string) error {
	root = filepath.Clean(root)
	target = filepath.Clean(target)
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == "." || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("compiled-test cache: refuse to prune path outside cache root: %s", target)
	}
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("compiled-test cache: prune %s: %w", target, err)
	}
	return nil
}
