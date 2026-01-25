package main

import (
	"encoding/json"
	"flag"
	"fmt"
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
	flag.Parse()

	root := *rootFlag
	if root == "" {
		root = defaultFixturesRoot()
	}
	if root == "" {
		fmt.Fprintln(os.Stderr, "fixture exporter: unable to locate repository root")
		os.Exit(1)
	}

	if err := exportFixtures(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func exportFixtures(root string) error {
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("fixture exporter: read fixtures root: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("fixture exporter: %s is not a directory", root)
	}

	count := 0
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() != "source.able" {
			return nil
		}
		module, err := parseModule(path)
		if err != nil {
			return fmt.Errorf("fixture exporter: parse %s: %w", path, err)
		}
		parser.NormalizeFixtureModule(module)
		serialized, err := json.MarshalIndent(module, "", "  ")
		if err != nil {
			return fmt.Errorf("fixture exporter: marshal %s: %w", path, err)
		}
		outPath := filepath.Join(filepath.Dir(path), "module.json")
		if err := os.WriteFile(outPath, append(serialized, '\n'), 0o644); err != nil {
			return fmt.Errorf("fixture exporter: write %s: %w", outPath, err)
		}
		count++
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %d fixture module(s) under %s\n", count, root)
	return nil
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
		return filepath.Clean(filepath.Join(start, "..", "..", "..", ".."))
	}
	return ""
}
