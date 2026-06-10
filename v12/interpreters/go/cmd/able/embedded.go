package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// embeddedKernelVersion parses the version from the embedded kernel package.yml.
func embeddedKernelVersion() string {
	data, err := embeddedKernelFS.ReadFile("embedded/kernel/package.yml")
	if err != nil {
		return "0.0.0"
	}
	var manifest struct {
		Version string `yaml:"version"`
	}
	if err := yaml.Unmarshal(data, &manifest); err != nil || manifest.Version == "" {
		return "0.0.0"
	}
	return manifest.Version
}

// ensureEmbeddedKernel extracts the embedded kernel to the ABLE_HOME cache
// and returns the path to the kernel src directory.
func ensureEmbeddedKernel() (string, error) {
	version := embeddedKernelVersion()
	cacheDir, err := resolveAbleHome()
	if err != nil {
		return "", fmt.Errorf("resolve ABLE_HOME for embedded kernel: %w", err)
	}
	target := filepath.Join(cacheDir, "pkg", "src", "kernel", version, "src")
	parentDir := filepath.Dir(target)
	if embeddedKernelCacheCurrent(target, parentDir) {
		return target, nil
	}
	if err := extractEmbeddedFS(embeddedKernelFS, "embedded/kernel/src", target); err != nil {
		return "", fmt.Errorf("extract embedded kernel: %w", err)
	}
	// Also extract package.yml to parent directory for manifest loading.
	if err := extractEmbeddedFS(embeddedKernelFS, "embedded/kernel/package.yml", filepath.Join(parentDir, "package.yml")); err != nil {
		return "", fmt.Errorf("extract embedded kernel manifest: %w", err)
	}
	return target, nil
}

// embeddedKernelCacheCurrent reports whether the cached embedded files still
// match this CLI build. The kernel version intentionally changes less often
// than its declarations, so versioned cache paths alone cannot detect updates.
func embeddedKernelCacheCurrent(srcDir, packageDir string) bool {
	return embeddedFileMatches(filepath.Join(srcDir, "kernel.able"), "embedded/kernel/src/kernel.able") &&
		embeddedFileMatches(filepath.Join(packageDir, "package.yml"), "embedded/kernel/package.yml")
}

func embeddedFileMatches(path, embeddedPath string) bool {
	cached, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	embedded, err := embeddedKernelFS.ReadFile(embeddedPath)
	return err == nil && bytes.Equal(cached, embedded)
}

// extractEmbeddedFS extracts files from an embed.FS subtree to a target path.
// If srcPath points to a single file, dstPath is treated as the destination file path.
// If srcPath points to a directory, dstPath is treated as the destination directory.
func extractEmbeddedFS(fsys fs.FS, srcPath, dstPath string) error {
	info, err := fs.Stat(fsys, srcPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		// Single file extraction
		data, err := fs.ReadFile(fsys, srcPath)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0o644)
	}
	// Directory extraction
	return fs.WalkDir(fsys, srcPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}
		rel = strings.ReplaceAll(rel, "/", string(filepath.Separator))
		target := filepath.Join(dstPath, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, readErr := fs.ReadFile(fsys, path)
		if readErr != nil {
			return readErr
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
