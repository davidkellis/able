package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func findGoModuleRoot() (string, error) {
	if env := os.Getenv("ABLE_GO_MODULE_ROOT"); env != "" {
		root, err := filepath.Abs(env)
		if err != nil {
			return "", fmt.Errorf("resolve ABLE_GO_MODULE_ROOT: %w", err)
		}
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return root, nil
		}
		return "", fmt.Errorf("ABLE_GO_MODULE_ROOT has no go.mod at %s", root)
	}

	if root, err := findGoModFrom("."); err == nil {
		return root, nil
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		if root, err := findGoModFrom(filepath.Dir(file)); err == nil {
			return root, nil
		}
	}
	if exe, err := os.Executable(); err == nil {
		if root, err := findGoModFrom(filepath.Dir(exe)); err == nil {
			return root, nil
		}
	}
	return "", fmt.Errorf("unable to locate go.mod (set ABLE_GO_MODULE_ROOT to override)")
}

func findGoModFrom(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("stat go.mod: %w", err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errors.New("go.mod not found")
}

func readGoModulePath(root string) (string, error) {
	path := filepath.Join(root, "go.mod")
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open go.mod: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	return "", fmt.Errorf("go.mod missing module path")
}

func writeBuildGoMod(outputDir string) error {
	moduleRoot, err := findGoModuleRoot()
	if err != nil {
		return err
	}
	modulePath, err := readGoModulePath(moduleRoot)
	if err != nil {
		return err
	}
	outDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("resolve build output: %w", err)
	}
	replacePath := moduleRoot
	if rel, err := filepath.Rel(outDir, moduleRoot); err == nil {
		replacePath = rel
	}
	replacePath = filepath.ToSlash(replacePath)
	content := fmt.Sprintf("module able/compiled\n\ngo 1.22\n\nrequire %s v0.0.0\n\nreplace %s => %s\n",
		modulePath,
		modulePath,
		replacePath,
	)
	if err := os.WriteFile(filepath.Join(outDir, "go.mod"), []byte(content), 0o600); err != nil {
		return fmt.Errorf("write go.mod: %w", err)
	}
	return nil
}
