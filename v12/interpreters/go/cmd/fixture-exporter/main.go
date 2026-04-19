package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/parser"
)

func main() {
	rootFlag := flag.String("root", "", "path to v12/fixtures/ast")
	checkFlag := flag.Bool("check", false, "verify fixture module.json files are current without rewriting them")
	flag.Parse()

	root := *rootFlag
	if root == "" {
		root = defaultFixturesRoot()
	}
	if root == "" {
		fmt.Fprintln(os.Stderr, "fixture exporter: unable to locate repository root")
		os.Exit(1)
	}

	if err := exportFixtures(root, *checkFlag, flag.Args(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func exportFixtures(root string, check bool, targets []string, stdout io.Writer) error {
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("fixture exporter: read fixtures root: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("fixture exporter: %s is not a directory", root)
	}

	sourceFiles, err := collectFixtureSources(root, targets)
	if err != nil {
		return err
	}

	count := 0
	stale := make([]string, 0)
	for _, path := range sourceFiles {
		module, err := parseModule(path)
		if err != nil {
			return fmt.Errorf("fixture exporter: parse %s: %w", path, err)
		}
		parser.NormalizeFixtureModule(module)
		serialized, err := json.MarshalIndent(module, "", "  ")
		if err != nil {
			return fmt.Errorf("fixture exporter: marshal %s: %w", path, err)
		}
		serialized = append(serialized, '\n')
		outPath := filepath.Join(filepath.Dir(path), "module.json")
		if check {
			existing, err := os.ReadFile(outPath)
			if err != nil {
				if os.IsNotExist(err) {
					stale = append(stale, outPath)
					count++
					continue
				}
				return fmt.Errorf("fixture exporter: read %s: %w", outPath, err)
			}
			if !bytes.Equal(existing, serialized) {
				stale = append(stale, outPath)
			}
			count++
			continue
		}
		if err := os.WriteFile(outPath, serialized, 0o644); err != nil {
			return fmt.Errorf("fixture exporter: write %s: %w", outPath, err)
		}
		count++
	}
	if check {
		if len(stale) > 0 {
			return fmt.Errorf(
				"fixture exporter: %d stale fixture module(s) under %s:\n  %s",
				len(stale),
				root,
				strings.Join(stale, "\n  "),
			)
		}
		fmt.Fprintf(stdout, "Verified %d fixture module(s) under %s\n", count, root)
		return nil
	}
	fmt.Fprintf(stdout, "Wrote %d fixture module(s) under %s\n", count, root)
	return nil
}

func collectFixtureSources(root string, targets []string) ([]string, error) {
	if len(targets) == 0 {
		return collectAllFixtureSources(root)
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("fixture exporter: resolve fixtures root: %w", err)
	}

	seen := make(map[string]struct{})
	out := make([]string, 0, len(targets))
	for _, target := range targets {
		resolved, err := resolveFixtureTarget(rootAbs, target)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return nil, fmt.Errorf("fixture exporter: read target %s: %w", target, err)
		}
		if info.IsDir() {
			files, err := collectAllFixtureSources(resolved)
			if err != nil {
				return nil, err
			}
			for _, path := range files {
				if _, ok := seen[path]; ok {
					continue
				}
				seen[path] = struct{}{}
				out = append(out, path)
			}
			continue
		}
		if filepath.Base(resolved) != "source.able" {
			return nil, fmt.Errorf("fixture exporter: target %s must be a fixture directory or source.able", target)
		}
		if _, ok := seen[resolved]; ok {
			continue
		}
		seen[resolved] = struct{}{}
		out = append(out, resolved)
	}
	return out, nil
}

func collectAllFixtureSources(root string) ([]string, error) {
	out := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "source.able" {
			return nil
		}
		out = append(out, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fixture exporter: walk fixtures root: %w", err)
	}
	return out, nil
}

func resolveFixtureTarget(root, target string) (string, error) {
	candidates := make([]string, 0, 2)
	if filepath.IsAbs(target) {
		candidates = append(candidates, filepath.Clean(target))
	} else {
		candidates = append(candidates, filepath.Join(root, target))
		candidates = append(candidates, target)
	}
	for _, candidate := range candidates {
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return "", fmt.Errorf("fixture exporter: resolve target %s: %w", target, err)
		}
		if !isWithinRoot(abs, root) {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
	}
	return "", fmt.Errorf("fixture exporter: target %s not found under %s", target, root)
}

func isWithinRoot(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func parseModule(path string) (*ast.Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p, err := parser.NewModuleParser()
	if err != nil {
		return nil, err
	}
	defer p.Close()
	return p.ParseModule(data)
}

func defaultFixturesRoot() string {
	root := repositoryRoot()
	if root == "" {
		return ""
	}
	return filepath.Join(root, "v12", "fixtures", "ast")
}

func repositoryRoot() string {
	start := ""
	if _, file, _, ok := runtime.Caller(0); ok {
		start = filepath.Dir(file)
	} else if wd, err := os.Getwd(); err == nil {
		start = wd
	}
	return repositoryRootFrom(start)
}

func repositoryRootFrom(start string) string {
	dir := start
	for i := 0; i < 12 && dir != "" && dir != string(filepath.Separator); i++ {
		if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if start == "" {
		return ""
	}
	if strings.HasSuffix(start, filepath.Join("v12", "interpreters", "go", "cmd", "fixture-exporter")) {
		return filepath.Clean(filepath.Join(start, "..", "..", "..", "..", ".."))
	}
	return ""
}
