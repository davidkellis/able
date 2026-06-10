package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

func runCompiledTestCacheCommand(args []string) int {
	if len(args) == 0 {
		printCompiledTestCacheUsage()
		return 1
	}
	switch args[0] {
	case "inspect":
		return runCompiledTestCacheInspect(args[1:])
	case "prune":
		return runCompiledTestCachePrune(args[1:])
	default:
		printCompiledTestCacheUsage()
		return 1
	}
}

func runCompiledTestCacheInspect(args []string) int {
	flags := flag.NewFlagSet("able cache compiled-tests inspect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var root string
	var jsonOutput bool
	var verbose bool
	flags.StringVar(&root, "dir", "", "compiled-test cache root")
	flags.BoolVar(&jsonOutput, "json", false, "emit JSON")
	flags.BoolVar(&verbose, "verbose", false, "list every classified entry")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		if err != nil && !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintf(os.Stderr, "able cache compiled-tests inspect: %v\n", err)
		}
		printCompiledTestCacheInspectUsage()
		return 1
	}
	cache, err := resolveCompiledTestCacheForCommand(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able cache compiled-tests inspect: %v\n", err)
		return 1
	}
	inventory, err := inspectCompiledTestCache(cache)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able cache compiled-tests inspect: %v\n", err)
		return 2
	}
	if jsonOutput {
		if err := writeCompiledTestCacheJSON(inventory); err != nil {
			fmt.Fprintf(os.Stderr, "able cache compiled-tests inspect: %v\n", err)
			return 2
		}
		return 0
	}
	printCompiledTestCacheInventory(inventory, verbose)
	return 0
}

func runCompiledTestCachePrune(args []string) int {
	flags := flag.NewFlagSet("able cache compiled-tests prune", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var root string
	var jsonOutput bool
	var dryRun bool
	var maxBytesRaw string
	var maxAgeRaw string
	flags.StringVar(&root, "dir", "", "compiled-test cache root")
	flags.BoolVar(&jsonOutput, "json", false, "emit JSON")
	flags.BoolVar(&dryRun, "dry-run", false, "report without deleting")
	flags.StringVar(&maxBytesRaw, "max-bytes", "", "maximum retained valid bytes")
	flags.StringVar(&maxAgeRaw, "max-age", "", "maximum valid-entry age")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		if err != nil && !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintf(os.Stderr, "able cache compiled-tests prune: %v\n", err)
		}
		printCompiledTestCachePruneUsage()
		return 1
	}
	options := compiledTestCachePruneOptions{DryRun: dryRun}
	if strings.TrimSpace(maxBytesRaw) != "" {
		maxBytes, err := parseCompiledTestCacheByteSize(maxBytesRaw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "able cache compiled-tests prune: %v\n", err)
			return 1
		}
		options.MaxBytes = maxBytes
		options.MaxBytesSet = true
	}
	if strings.TrimSpace(maxAgeRaw) != "" {
		maxAge, err := parseCompiledTestCacheDuration(maxAgeRaw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "able cache compiled-tests prune: %v\n", err)
			return 1
		}
		options.MaxAge = maxAge
		options.MaxAgeSet = true
	}
	cache, err := resolveCompiledTestCacheForCommand(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able cache compiled-tests prune: %v\n", err)
		return 1
	}
	result, err := pruneCompiledTestCache(cache, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able cache compiled-tests prune: %v\n", err)
		return 2
	}
	if jsonOutput {
		if err := writeCompiledTestCacheJSON(result); err != nil {
			fmt.Fprintf(os.Stderr, "able cache compiled-tests prune: %v\n", err)
			return 2
		}
	} else {
		printCompiledTestCachePruneResult(result)
	}
	if result.Busy {
		return 2
	}
	return 0
}

func resolveCompiledTestCacheForCommand(explicitRoot string) (*compiledTestCache, error) {
	root := strings.TrimSpace(explicitRoot)
	if root == "" {
		root = strings.TrimSpace(os.Getenv(compiledTestCacheDirEnv))
	}
	if root == "" {
		return nil, fmt.Errorf("cache root is required via --dir or %s", compiledTestCacheDirEnv)
	}
	return openCompiledTestCacheAt(root)
}

func printCompiledTestCacheInventory(inventory compiledTestCacheInventory, verbose bool) {
	fmt.Fprintf(os.Stdout, "compiled-test cache: %s\n", inventory.Root)
	fmt.Fprintf(os.Stdout, "  schema: %s\n", inventory.Schema)
	fmt.Fprintf(os.Stdout, "  valid: %d entries, %s\n",
		inventory.ValidEntries, formatCompiledTestCacheBytes(inventory.ValidBytes))
	fmt.Fprintf(os.Stdout, "  corrupt: %d entries, %s\n",
		inventory.CorruptEntries, formatCompiledTestCacheBytes(inventory.CorruptBytes))
	fmt.Fprintf(os.Stdout, "  staging: %d entries, %s\n",
		inventory.StagingEntries, formatCompiledTestCacheBytes(inventory.StagingBytes))
	fmt.Fprintf(os.Stdout, "  obsolete-schema: %d entries, %s\n",
		inventory.ObsoleteEntries, formatCompiledTestCacheBytes(inventory.ObsoleteBytes))
	fmt.Fprintf(os.Stdout, "  unknown: %d entries, %s\n",
		inventory.UnknownEntries, formatCompiledTestCacheBytes(inventory.UnknownBytes))
	if !verbose {
		return
	}
	for _, entry := range inventory.Entries {
		fmt.Fprintf(os.Stdout, "  %s %s %s %s",
			entry.Class,
			formatCompiledTestCacheBytes(entry.SizeBytes),
			entry.LastUsedAt.UTC().Format(time.RFC3339),
			entry.Path,
		)
		if entry.Reason != "" {
			fmt.Fprintf(os.Stdout, " (%s)", entry.Reason)
		}
		fmt.Fprintln(os.Stdout)
	}
}

func printCompiledTestCachePruneResult(result compiledTestCachePruneResult) {
	if result.Busy {
		fmt.Fprintln(os.Stdout, "compiled-test cache is busy; no entries pruned")
		return
	}
	action := "pruned"
	if result.DryRun {
		action = "would prune"
	}
	fmt.Fprintf(os.Stdout, "compiled-test cache: %s\n", result.Root)
	fmt.Fprintf(os.Stdout, "  %s: %d entries, %s\n",
		action, result.RemovedEntries, formatCompiledTestCacheBytes(result.RemovedBytes))
	fmt.Fprintf(os.Stdout, "  retained valid: %d entries, %s\n",
		result.RetainedValid, formatCompiledTestCacheBytes(result.RetainedValidBytes))
}

func writeCompiledTestCacheJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	return nil
}

func parseCompiledTestCacheByteSize(raw string) (int64, error) {
	normalized := strings.TrimSpace(strings.ToUpper(raw))
	if normalized == "" {
		return 0, fmt.Errorf("max bytes must not be empty")
	}
	multipliers := []struct {
		suffix     string
		multiplier float64
	}{
		{suffix: "TIB", multiplier: 1 << 40},
		{suffix: "GIB", multiplier: 1 << 30},
		{suffix: "MIB", multiplier: 1 << 20},
		{suffix: "KIB", multiplier: 1 << 10},
		{suffix: "TB", multiplier: 1e12},
		{suffix: "GB", multiplier: 1e9},
		{suffix: "MB", multiplier: 1e6},
		{suffix: "KB", multiplier: 1e3},
		{suffix: "B", multiplier: 1},
	}
	multiplier := float64(1)
	number := normalized
	for _, candidate := range multipliers {
		if strings.HasSuffix(normalized, candidate.suffix) {
			multiplier = candidate.multiplier
			number = strings.TrimSpace(strings.TrimSuffix(normalized, candidate.suffix))
			break
		}
	}
	value, err := strconv.ParseFloat(number, 64)
	if err != nil || value < 0 || value > float64(math.MaxInt64)/multiplier {
		return 0, fmt.Errorf("invalid max bytes %q (examples: 1073741824, 1GiB, 1.5GB)", raw)
	}
	return int64(value * multiplier), nil
}

func parseCompiledTestCacheDuration(raw string) (time.Duration, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if strings.HasSuffix(normalized, "d") {
		days, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(normalized, "d")), 64)
		if err != nil || days < 0 || days > float64(math.MaxInt64)/float64(24*time.Hour) {
			return 0, fmt.Errorf("invalid max age %q (examples: 24h, 7d)", raw)
		}
		return time.Duration(days * float64(24*time.Hour)), nil
	}
	duration, err := time.ParseDuration(normalized)
	if err != nil || duration < 0 {
		return 0, fmt.Errorf("invalid max age %q (examples: 24h, 7d)", raw)
	}
	return duration, nil
}

func formatCompiledTestCacheBytes(size int64) string {
	const (
		kib = int64(1 << 10)
		mib = int64(1 << 20)
		gib = int64(1 << 30)
		tib = int64(1 << 40)
	)
	switch {
	case size >= tib:
		return fmt.Sprintf("%.2f TiB", float64(size)/float64(tib))
	case size >= gib:
		return fmt.Sprintf("%.2f GiB", float64(size)/float64(gib))
	case size >= mib:
		return fmt.Sprintf("%.2f MiB", float64(size)/float64(mib))
	case size >= kib:
		return fmt.Sprintf("%.2f KiB", float64(size)/float64(kib))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func printCompiledTestCacheUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  able cache compiled-tests inspect [--dir PATH] [--json] [--verbose]")
	fmt.Fprintln(os.Stderr, "  able cache compiled-tests prune [--dir PATH] [--max-bytes SIZE] [--max-age DURATION] [--dry-run] [--json]")
}

func printCompiledTestCacheInspectUsage() {
	fmt.Fprintln(os.Stderr, "usage: able cache compiled-tests inspect [--dir PATH] [--json] [--verbose]")
}

func printCompiledTestCachePruneUsage() {
	fmt.Fprintln(os.Stderr, "usage: able cache compiled-tests prune [--dir PATH] [--max-bytes SIZE] [--max-age DURATION] [--dry-run] [--json]")
}
