package typechecker

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	diagRootOnce sync.Once
	diagRootPath string
)

func normalizeDiagnosticPath(raw string) string {
	if raw == "" {
		return ""
	}
	path := raw
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}
	root := diagnosticRoot()
	anchors := []string{}
	if root != "" {
		anchors = append(anchors, filepath.Join(root, "v11", "interpreters", "ts", "scripts", "export-fixtures"))
		anchors = append(anchors, filepath.Join(root, "v11", "interpreters", "ts", "scripts"))
		anchors = append(anchors, root)
	}
	for _, anchor := range anchors {
		if anchor == "" {
			continue
		}
		if rel, err := filepath.Rel(anchor, path); err == nil {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(path)
}

func diagnosticRoot() string {
	diagRootOnce.Do(func() {
		start := ""
		if _, file, _, ok := runtime.Caller(0); ok {
			start = filepath.Dir(file)
		} else if wd, err := os.Getwd(); err == nil {
			start = wd
		}
		dir := start
		for i := 0; i < 12 && dir != "" && dir != string(filepath.Separator); i++ {
			if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
				diagRootPath = dir
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	})
	return diagRootPath
}
