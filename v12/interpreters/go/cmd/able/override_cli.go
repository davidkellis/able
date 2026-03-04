package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func runOverride(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "able override requires a subcommand (add, remove, list)")
		return 1
	}
	switch args[0] {
	case "add":
		if len(args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: able override add <git-url> <local-path>")
			return 1
		}
		return runOverrideAdd(args[1], args[2])
	case "remove":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: able override remove <git-url>")
			return 1
		}
		return runOverrideRemove(args[1])
	case "list":
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "able override list does not take arguments")
			return 1
		}
		return runOverrideList()
	default:
		fmt.Fprintf(os.Stderr, "unknown override subcommand %q\n", args[0])
		return 1
	}
}

func runOverrideAdd(gitURL, localPath string) int {
	// Resolve the local path to absolute.
	abs, err := filepath.Abs(localPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve path %q: %v\n", localPath, err)
		return 1
	}

	// Validate that the directory exists.
	info, err := os.Stat(abs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "path %s does not exist: %v\n", abs, err)
		return 1
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "path %s is not a directory\n", abs)
		return 1
	}

	// Validate that package.yml exists in the directory.
	manifestPath := filepath.Join(abs, "package.yml")
	if _, err := os.Stat(manifestPath); err != nil {
		fmt.Fprintf(os.Stderr, "no package.yml found at %s\n", manifestPath)
		return 1
	}

	normalized := normalizeGitURL(gitURL)
	overrides := loadGlobalOverrides()
	if overrides == nil {
		overrides = make(map[string]string)
	}
	overrides[normalized] = abs

	if err := saveGlobalOverrides(overrides); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save overrides: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stdout, "override added: %s → %s\n", normalized, abs)
	return 0
}

func runOverrideRemove(gitURL string) int {
	normalized := normalizeGitURL(gitURL)
	overrides := loadGlobalOverrides()
	if len(overrides) == 0 {
		fmt.Fprintf(os.Stderr, "no override found for %s\n", normalized)
		return 1
	}
	if _, ok := overrides[normalized]; !ok {
		fmt.Fprintf(os.Stderr, "no override found for %s\n", normalized)
		return 1
	}
	delete(overrides, normalized)
	if err := saveGlobalOverrides(overrides); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save overrides: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stdout, "override removed: %s\n", normalized)
	return 0
}

func runOverrideList() int {
	overrides := loadGlobalOverrides()
	if len(overrides) == 0 {
		fmt.Fprintln(os.Stdout, "no overrides configured")
		return 0
	}
	keys := make([]string, 0, len(overrides))
	for k := range overrides {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(os.Stdout, "%s → %s\n", k, overrides[k])
	}
	return 0
}
