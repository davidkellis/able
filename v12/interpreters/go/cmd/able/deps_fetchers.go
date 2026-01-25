package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/driver"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type registryFetcher struct {
	base string
}

var nullRegistryFetcher *registryFetcher

func newRegistryFetcher(cacheDir string) *registryFetcher {
	if cacheDir == "" {
		return nullRegistryFetcher
	}
	return &registryFetcher{
		base: cacheDir,
	}
}

func (r *registryFetcher) Fetch(registry, name, version string) (*driver.LockedPackage, string, error) {
	if r == nil {
		return nil, "", errors.New("registry fetcher not initialised")
	}
	registryDir := os.Getenv("ABLE_REGISTRY")
	if registryDir == "" {
		registryDir = filepath.Join(r.base, "registry")
	}
	packageDir := filepath.Join(registryDir, registry, name, version)
	info, err := os.Stat(packageDir)
	if err != nil {
		return nil, "", fmt.Errorf("registry: package %s@%s not found in %s: %w", name, version, packageDir, err)
	}
	if !info.IsDir() {
		return nil, "", fmt.Errorf("registry: expected directory at %s", packageDir)
	}

	srcDir := filepath.Join(packageDir, "src")
	if _, err := os.Stat(srcDir); err != nil {
		return nil, "", fmt.Errorf("registry: package %s@%s missing src directory: %w", name, version, err)
	}

	cacheSrc := filepath.Join(r.base, "pkg", "src", sanitizeName(name), version)
	if err := copyOrSyncDir(srcDir, cacheSrc); err != nil {
		return nil, "", fmt.Errorf("registry: copy %s -> %s: %w", srcDir, cacheSrc, err)
	}

	checksum, err := dirChecksum(srcDir)
	if err != nil {
		return nil, "", fmt.Errorf("registry: checksum %s: %w", srcDir, err)
	}

	return &driver.LockedPackage{
		Name:     sanitizeName(name),
		Version:  version,
		Source:   fmt.Sprintf("registry:%s/%s/%s", registry, sanitizeName(name), version),
		Checksum: checksum,
	}, packageDir, nil
}

func copyOrSyncDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Remove stale files from destination.
	dstEntries, err := os.ReadDir(dst)
	if err == nil {
		for _, entry := range dstEntries {
			found := false
			for _, srcEntry := range entries {
				if srcEntry.Name() == entry.Name() {
					found = true
					break
				}
			}
			if !found {
				if err := os.RemoveAll(filepath.Join(dst, entry.Name())); err != nil {
					return err
				}
			}
		}
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyOrSyncDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func dirChecksum(path string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		h.Write([]byte(filepath.Base(p)))
		h.Write(data)
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

type gitFetcher struct {
	cacheDir string
}

func newGitFetcher(cacheDir string) *gitFetcher {
	if cacheDir == "" {
		return nil
	}
	return &gitFetcher{cacheDir: cacheDir}
}

func (g *gitFetcher) Fetch(name string, spec *driver.DependencySpec) (*driver.LockedPackage, string, error) {
	if g == nil {
		return nil, "", errors.New("git fetcher unavailable")
	}
	url := strings.TrimSpace(spec.Git)
	if url == "" {
		return nil, "", fmt.Errorf("dependency %q: git URL required", name)
	}

	baseDir := filepath.Join(g.cacheDir, "pkg", "src", sanitizeName(name))
	version, commit, err := ensureGitCheckout(baseDir, url, spec)
	if err != nil {
		return nil, "", err
	}

	checkoutDir := filepath.Join(baseDir, sanitizePathSegment(version))
	checksum, err := dirChecksum(checkoutDir)
	if err != nil {
		return nil, "", err
	}

	return &driver.LockedPackage{
		Name:     sanitizeName(name),
		Version:  version,
		Source:   fmt.Sprintf("git+%s@%s", url, commit),
		Checksum: checksum,
	}, commit, nil
}

func ensureGitCheckout(baseDir, url string, spec *driver.DependencySpec) (string, string, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", "", err
	}

	revision, descriptor, err := gitRevisionFromSpec(spec)
	if err != nil {
		return "", "", err
	}

	explicitRev := strings.TrimSpace(spec.Rev)
	if explicitRev != "" {
		existing := filepath.Join(baseDir, sanitizePathSegment(explicitRev))
		if _, err := os.Stat(existing); err == nil {
			return explicitRev, explicitRev, nil
		}
	}

	tmpDir, err := os.MkdirTemp(baseDir, "git-fetch-*")
	if err != nil {
		return "", "", err
	}
	if err := os.RemoveAll(tmpDir); err != nil {
		return "", "", err
	}

	repo, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:               url,
		Depth:             0,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", "", fmt.Errorf("git clone %s: %w", url, err)
	}

	hash, err := repo.ResolveRevision(revision)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", "", fmt.Errorf("resolve revision %s: %w", revision, err)
	}

	version := gitPinnedVersion(descriptor, hash.String())
	targetDir := filepath.Join(baseDir, sanitizePathSegment(version))
	if _, err := os.Stat(targetDir); err == nil {
		_ = os.RemoveAll(tmpDir)
		return version, hash.String(), nil
	}

	worktree, err := repo.Worktree()
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", "", err
	}
	if err := worktree.Checkout(&git.CheckoutOptions{
		Hash:  *hash,
		Force: true,
	}); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", "", fmt.Errorf("git checkout %s: %w", revision, err)
	}

	if err := os.Rename(tmpDir, targetDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", "", err
	}
	return version, hash.String(), nil
}

func gitPinnedVersion(descriptor, commit string) string {
	commit = strings.TrimSpace(commit)
	descriptor = strings.TrimSpace(descriptor)
	if commit == "" {
		return descriptor
	}
	if descriptor == "" || descriptor == commit {
		return commit
	}
	return fmt.Sprintf("%s@%s", descriptor, commit)
}

func gitRevisionFromSpec(spec *driver.DependencySpec) (plumbing.Revision, string, error) {
	if rev := strings.TrimSpace(spec.Rev); rev != "" {
		return plumbing.Revision(rev), rev, nil
	}
	if tag := strings.TrimSpace(spec.Tag); tag != "" {
		return plumbing.Revision("refs/tags/" + tag), tag, nil
	}
	if branch := strings.TrimSpace(spec.Branch); branch != "" {
		return plumbing.Revision("refs/heads/" + branch), branch, nil
	}
	return "", "", fmt.Errorf("git dependencies require rev, tag, or branch")
}

func sanitizePathSegment(segment string) string {
	segment = strings.TrimSpace(segment)
	if segment == "" {
		return "head"
	}
	var b strings.Builder
	for _, r := range segment {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	result := b.String()
	if result == "" {
		return "head"
	}
	return result
}
