package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type buildModuleInfo struct {
	Root string
	Path string
}

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

func loadBuildModuleInfo() (buildModuleInfo, error) {
	moduleRoot, err := findGoModuleRoot()
	if err != nil {
		return buildModuleInfo{}, err
	}
	modulePath, err := readGoModulePath(moduleRoot)
	if err != nil {
		return buildModuleInfo{}, err
	}
	return buildModuleInfo{Root: moduleRoot, Path: modulePath}, nil
}

func writeBuildGoMod(outputDir string, info buildModuleInfo, replaceOverride string) error {
	outDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("resolve build output: %w", err)
	}
	replacePath := replaceOverride
	if replacePath == "" {
		replacePath = info.Root
		if rel, err := filepath.Rel(outDir, info.Root); err == nil {
			replacePath = rel
		}
	}
	replacePath = filepath.ToSlash(replacePath)
	content := fmt.Sprintf("module able/compiled\n\ngo 1.22\n\nrequire %s v0.0.0\n\nreplace %s => %s\n",
		info.Path,
		info.Path,
		replacePath,
	)
	if err := os.WriteFile(filepath.Join(outDir, "go.mod"), []byte(content), 0o600); err != nil {
		return fmt.Errorf("write go.mod: %w", err)
	}
	return nil
}

func prepareBuildModule(outputDir string) error {
	info, err := loadBuildModuleInfo()
	if err != nil {
		return err
	}
	outDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("resolve build output: %w", err)
	}
	moduleRoot, err := filepath.Abs(info.Root)
	if err != nil {
		return fmt.Errorf("resolve module root: %w", err)
	}
	if isWithinDir(outDir, moduleRoot) {
		return writeBuildGoMod(outDir, info, "")
	}

	v12Root := filepath.Dir(filepath.Dir(moduleRoot))
	copyTarget := filepath.Join(outDir, "v12", "interpreters", "go")
	if err := copyModuleTree(moduleRoot, copyTarget); err != nil {
		return fmt.Errorf("copy interpreter module: %w", err)
	}
	parserSrc := filepath.Join(v12Root, "parser", "tree-sitter-able")
	parserDst := filepath.Join(outDir, "v12", "parser", "tree-sitter-able")
	if err := copyModuleTree(parserSrc, parserDst); err != nil {
		return fmt.Errorf("copy parser sources: %w", err)
	}
	return writeBuildGoMod(outDir, info, "./v12/interpreters/go")
}

func isWithinDir(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	prefix := ".." + string(os.PathSeparator)
	return rel != ".." && !strings.HasPrefix(rel, prefix)
}

func copyModuleTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		if entry.IsDir() {
			if shouldSkipModuleDir(entry.Name()) {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return copyBuildFile(path, filepath.Join(dst, rel), info.Mode().Perm())
	})
}

func shouldSkipModuleDir(name string) bool {
	switch name {
	case ".git", ".gocache", "tmp", "target", "node_modules", "vendor", ".tmp":
		return true
	default:
		return false
	}
}

func copyBuildFile(src, dst string, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
