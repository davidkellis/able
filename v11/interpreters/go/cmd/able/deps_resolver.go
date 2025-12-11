package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"able/interpreter10-go/pkg/driver"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type resolvedPackage struct {
	pkg      *driver.LockedPackage
	manifest *driver.Manifest
	root     string
}

func buildExecutionSearchPaths(manifest *driver.Manifest, lock *driver.Lockfile) ([]string, error) {
	var extras []string
	var manifestRoot string
	if manifest != nil {
		manifestRoot = filepath.Dir(manifest.Path)
		extras = append(extras, manifestRoot)
	}
	if lock == nil || len(lock.Packages) == 0 {
		return extras, nil
	}

	cacheDir, err := resolveAbleHome()
	if err != nil {
		return nil, err
	}

	for _, pkg := range lock.Packages {
		if pkg == nil {
			continue
		}
		if source := strings.TrimSpace(pkg.Source); source != "" {
			if resolved, ok := resolvePackageSourcePath(source, manifestRoot, cacheDir); ok {
				extras = append(extras, resolved)
				continue
			}
		}
		if pkg.Name == "" || pkg.Version == "" {
			continue
		}
		extras = append(extras, filepath.Join(cacheDir, "pkg", "src", pkg.Name, sanitizePathSegment(pkg.Version)))
	}
	return extras, nil
}

func resolvePackageSourcePath(source, manifestRoot, cacheDir string) (string, bool) {
	if strings.HasPrefix(source, "path:") {
		pathSpec := strings.TrimSpace(strings.TrimPrefix(source, "path:"))
		if pathSpec == "" {
			return "", false
		}
		if filepath.IsAbs(pathSpec) {
			return filepath.Clean(pathSpec), true
		}
		base := manifestRoot
		if base == "" {
			base = cacheDir
		}
		return filepath.Join(base, filepath.FromSlash(pathSpec)), true
	}
	if strings.HasPrefix(source, "registry:") {
		pathSpec := strings.TrimSpace(strings.TrimPrefix(source, "registry:"))
		if pathSpec == "" {
			return "", false
		}
		parts := strings.Split(pathSpec, "/")
		if len(parts) >= 2 {
			name := parts[len(parts)-2]
			version := parts[len(parts)-1]
			return filepath.Join(cacheDir, "pkg", "src", filepath.FromSlash(name), filepath.FromSlash(version)), true
		}
		return filepath.Join(cacheDir, "pkg", "src", filepath.FromSlash(pathSpec)), true
	}
	if strings.HasPrefix(source, "git:") {
		pathSpec := strings.TrimSpace(strings.TrimPrefix(source, "git:"))
		if pathSpec == "" {
			return "", false
		}
		return filepath.Join(cacheDir, "pkg", "src", filepath.FromSlash(pathSpec)), true
	}
	return "", false
}

type dependencyInstaller struct {
	manifest     *driver.Manifest
	manifestRoot string
	cacheDir     string
	logs         []string
	registry     *registryFetcher
	git          *gitFetcher
	resolved     map[string]*driver.LockedPackage
	aliases      map[string]string
	resolving    map[string]bool
	resolvingPkg map[string]bool
}

func newDependencyInstaller(manifest *driver.Manifest, cacheDir string) *dependencyInstaller {
	var root string
	if manifest != nil {
		root = filepath.Dir(manifest.Path)
	}
	return &dependencyInstaller{
		manifest:     manifest,
		manifestRoot: root,
		cacheDir:     cacheDir,
		logs:         []string{},
		registry:     newRegistryFetcher(cacheDir),
		git:          newGitFetcher(cacheDir),
		resolved:     make(map[string]*driver.LockedPackage),
		aliases:      make(map[string]string),
		resolving:    make(map[string]bool),
		resolvingPkg: make(map[string]bool),
	}
}

func (d *dependencyInstaller) Install(lock *driver.Lockfile) (bool, []string, error) {
	if d.manifest == nil {
		return false, d.logs, nil
	}

	d.resolved = make(map[string]*driver.LockedPackage)
	d.aliases = make(map[string]string)
	d.resolving = make(map[string]bool)
	d.resolvingPkg = make(map[string]bool)

	hasStdlibDep := false
	hasKernelDep := false
	names := make([]string, 0, len(d.manifest.Dependencies))
	for name := range d.manifest.Dependencies {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		spec := d.manifest.Dependencies[name]
		if spec == nil {
			return false, d.logs, fmt.Errorf("dependency %q has no descriptor", name)
		}
		if sanitizeName(name) == "able" {
			hasStdlibDep = true
		}
		if sanitizeName(name) == "kernel" {
			hasKernelDep = true
		}
		if err := d.installDependency(name, cloneDependencySpec(spec)); err != nil {
			return false, d.logs, err
		}
	}

	if !hasStdlibDep {
		if err := d.installDependency("able", &driver.DependencySpec{}); err != nil {
			return false, d.logs, fmt.Errorf("resolve stdlib: %w", err)
		}
	}

	if !hasKernelDep {
		if err := d.installDependency("kernel", &driver.DependencySpec{}); err != nil {
			return false, d.logs, fmt.Errorf("resolve kernel: %w", err)
		}
	}

	desired := make([]*driver.LockedPackage, 0, len(d.resolved))
	for _, pkg := range d.resolved {
		if pkg == nil {
			continue
		}
		desired = append(desired, pkg)
	}
	sort.SliceStable(desired, func(i, j int) bool {
		if desired[i].Name == desired[j].Name {
			return desired[i].Version < desired[j].Version
		}
		return desired[i].Name < desired[j].Name
	})

	existing := make(map[string]*driver.LockedPackage, len(lock.Packages))
	for _, pkg := range lock.Packages {
		if pkg == nil {
			continue
		}
		existing[pkg.Name] = pkg
	}

	changed := len(desired) != len(existing)
	for _, pkg := range desired {
		if current, ok := existing[pkg.Name]; ok {
			if !lockedPackageEqual(current, pkg) {
				changed = true
			}
		} else {
			changed = true
		}
	}

	lock.Packages = desired
	return changed, d.logs, nil
}

func (d *dependencyInstaller) installDependency(name string, spec *driver.DependencySpec) error {
	if spec == nil {
		return fmt.Errorf("dependency %q has no descriptor", name)
	}
	alias := sanitizeName(name)
	if canonical, ok := d.aliases[alias]; ok {
		if _, exists := d.resolved[canonical]; exists {
			return nil
		}
		if d.resolvingPkg[canonical] {
			return fmt.Errorf("dependency cycle detected at %s", canonical)
		}
	}
	if d.resolving[alias] {
		return fmt.Errorf("dependency cycle detected at %s", alias)
	}
	d.resolving[alias] = true
	defer delete(d.resolving, alias)

	resolvedPkg, err := d.resolveDependency(name, spec)
	if err != nil {
		return err
	}
	if resolvedPkg == nil || resolvedPkg.pkg == nil {
		return nil
	}

	pkg := resolvedPkg.pkg
	canonical := pkg.Name
	if canonical == "" {
		canonical = alias
	}

	if d.resolvingPkg[canonical] {
		return fmt.Errorf("dependency cycle detected at %s", canonical)
	}
	d.resolvingPkg[canonical] = true
	defer delete(d.resolvingPkg, canonical)

	d.aliases[alias] = canonical
	if _, exists := d.resolved[canonical]; exists {
		return nil
	}

	pkg.Dependencies = nil

	if resolvedPkg.manifest != nil && len(resolvedPkg.manifest.Dependencies) > 0 {
		childNames := make([]string, 0, len(resolvedPkg.manifest.Dependencies))
		for childName, childSpec := range resolvedPkg.manifest.Dependencies {
			if childSpec == nil {
				return fmt.Errorf("dependency %s lists %s without descriptor", pkg.Name, childName)
			}
			if childSpec.Optional {
				continue
			}
			childNames = append(childNames, childName)
		}
		sort.Strings(childNames)
		seen := make(map[string]struct{}, len(childNames))
		for _, childName := range childNames {
			childSpec := cloneDependencySpec(resolvedPkg.manifest.Dependencies[childName])
			if childSpec == nil {
				return fmt.Errorf("dependency %s lists %s without descriptor", pkg.Name, childName)
			}
			if childSpec.Path != "" && !filepath.IsAbs(childSpec.Path) {
				base := resolvedPkg.root
				if base == "" {
					base = d.manifestRoot
				}
				if base != "" {
					childSpec.Path = filepath.Clean(filepath.Join(base, childSpec.Path))
				}
			}
			if err := d.installDependency(childName, childSpec); err != nil {
				return err
			}
			childAlias := sanitizeName(childName)
			canonicalChild := d.aliases[childAlias]
			if canonicalChild == "" {
				canonicalChild = childAlias
			}
			childPkg, ok := d.resolved[canonicalChild]
			if !ok {
				return fmt.Errorf("resolved child package %s missing from cache", childName)
			}
			if _, dup := seen[childPkg.Name]; dup {
				continue
			}
			seen[childPkg.Name] = struct{}{}
			pkg.Dependencies = append(pkg.Dependencies, driver.LockedDependency{
				Name:    childPkg.Name,
				Version: childPkg.Version,
			})
		}
		sort.SliceStable(pkg.Dependencies, func(i, j int) bool {
			if pkg.Dependencies[i].Name == pkg.Dependencies[j].Name {
				return pkg.Dependencies[i].Version < pkg.Dependencies[j].Version
			}
			return pkg.Dependencies[i].Name < pkg.Dependencies[j].Name
		})
	}

	d.resolved[canonical] = pkg
	return nil
}

func (d *dependencyInstaller) resolveDependency(name string, spec *driver.DependencySpec) (*resolvedPackage, error) {
	if spec.Path != "" {
		return d.resolvePathDependency(name, spec)
	}
	if sanitizeName(name) == "able" {
		return d.resolveStdlibDependency(spec)
	}
	if sanitizeName(name) == "kernel" {
		return d.resolveKernelDependency(spec)
	}
	if spec.Git != "" {
		return d.resolveGitDependency(name, spec)
	}
	if spec.Version != "" {
		return d.resolveRegistryDependency(name, spec)
	}
	return nil, fmt.Errorf("dependency %q: unsupported descriptor", name)
}

func (d *dependencyInstaller) resolvePathDependency(name string, spec *driver.DependencySpec) (*resolvedPackage, error) {
	pathSpec := spec.Path
	if !filepath.IsAbs(pathSpec) {
		pathSpec = filepath.Join(d.manifestRoot, pathSpec)
	}
	abs, err := filepath.Abs(pathSpec)
	if err != nil {
		return nil, fmt.Errorf("dependency %q: resolve path %q: %w", name, spec.Path, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("dependency %q: stat %s: %w", name, abs, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("dependency %q: expected directory at %s", name, abs)
	}

	manifestPath := filepath.Join(abs, "package.yml")
	depManifest, err := driver.LoadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("dependency %q: load manifest %s: %w", name, manifestPath, err)
	}

	version := strings.TrimSpace(depManifest.Version)
	if version == "" {
		version = "0.0.0-dev"
	}
	pkgName := depManifest.Name
	if pkgName == "" {
		pkgName = sanitizeName(name)
	}

	d.logs = append(d.logs, fmt.Sprintf("linked %s %s (%s)", pkgName, version, d.displayPath(abs)))

	lock := &driver.LockedPackage{
		Name:     pkgName,
		Version:  version,
		Source:   fmt.Sprintf("path:%s", abs),
		Checksum: "",
	}

	return &resolvedPackage{
		pkg:      lock,
		manifest: depManifest,
		root:     abs,
	}, nil
}

func (d *dependencyInstaller) resolveStdlibDependency(spec *driver.DependencySpec) (*resolvedPackage, error) {
	paths := collectStdlibPaths(d.manifestRoot)
	for _, candidate := range paths {
		root, manifestPath := ascendToManifest(candidate)
		if manifestPath == "" {
			continue
		}
		stdManifest, err := driver.LoadManifest(manifestPath)
		if err != nil {
			continue
		}
		pkgName := sanitizeName(stdManifest.Name)
		if pkgName == "" {
			pkgName = "able"
		}
		if pkgName != "able" {
			continue
		}
		version := strings.TrimSpace(stdManifest.Version)
		if version == "" {
			version = "0.0.0"
		}
		if spec.Version != "" && !constraintContainsVersion(spec.Version, version) {
			continue
		}
		src := candidate
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			// already src directory
		} else {
			src = filepath.Join(root, "src")
		}
		d.logs = append(d.logs, fmt.Sprintf("using stdlib %s (%s)", version, d.displayPath(src)))
		lock := &driver.LockedPackage{
			Name:    "able",
			Version: version,
			Source:  fmt.Sprintf("path:%s", src),
		}
		return &resolvedPackage{
			pkg:      lock,
			manifest: stdManifest,
			root:     root,
		}, nil
	}
	return nil, fmt.Errorf("stdlib dependency requested but no stdlib distribution found")
}

func (d *dependencyInstaller) resolveKernelDependency(spec *driver.DependencySpec) (*resolvedPackage, error) {
	paths := collectKernelPaths(d.manifestRoot)
	for _, candidate := range paths {
		root, manifestPath := ascendToManifest(candidate)
		if manifestPath == "" {
			continue
		}
		kernelManifest, err := driver.LoadManifest(manifestPath)
		if err != nil {
			continue
		}
		pkgName := sanitizeName(kernelManifest.Name)
		if pkgName == "" {
			pkgName = "kernel"
		}
		if pkgName != "kernel" {
			continue
		}
		version := strings.TrimSpace(kernelManifest.Version)
		if version == "" {
			version = "0.0.0"
		}
		if spec.Version != "" && !constraintContainsVersion(spec.Version, version) {
			continue
		}
		src := candidate
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			// already src directory
		} else {
			src = filepath.Join(root, "src")
		}
		d.logs = append(d.logs, fmt.Sprintf("using kernel %s (%s)", version, d.displayPath(src)))
		lock := &driver.LockedPackage{
			Name:    "kernel",
			Version: version,
			Source:  fmt.Sprintf("path:%s", src),
		}
		return &resolvedPackage{
			pkg:      lock,
			manifest: kernelManifest,
			root:     root,
		}, nil
	}
	return nil, fmt.Errorf("kernel dependency requested but no kernel distribution found")
}

func (d *dependencyInstaller) resolveRegistryDependency(name string, spec *driver.DependencySpec) (*resolvedPackage, error) {
	if d.registry == nullRegistryFetcher {
		return nil, fmt.Errorf("dependency %q: registry support unavailable", name)
	}
	regName := spec.Registry
	if regName == "" {
		regName = "default"
	}
	version := strings.TrimSpace(spec.Version)
	if version == "" {
		return nil, fmt.Errorf("dependency %q: registry dependencies must specify a version", name)
	}

	pkg, packageDir, err := d.registry.Fetch(regName, name, version)
	if err != nil {
		return nil, err
	}

	d.logs = append(d.logs, fmt.Sprintf("downloaded %s %s from registry %s", pkg.Name, pkg.Version, regName))

	manifestPath := filepath.Join(packageDir, "package.yml")
	var manifest *driver.Manifest
	if data, err := driver.LoadManifest(manifestPath); err == nil {
		manifest = data
		cachePkgDir := filepath.Join(d.cacheDir, "pkg", "src", pkg.Name, sanitizePathSegment(pkg.Version))
		cacheManifest := filepath.Join(cachePkgDir, "package.yml")
		if err := copyFile(manifestPath, cacheManifest); err != nil {
			return nil, fmt.Errorf("dependency %q: cache manifest %s: %w", name, cacheManifest, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("dependency %q: load manifest %s: %w", name, manifestPath, err)
	}

	return &resolvedPackage{
		pkg:      pkg,
		manifest: manifest,
		root:     packageDir,
	}, nil
}

func (d *dependencyInstaller) resolveGitDependency(name string, spec *driver.DependencySpec) (*resolvedPackage, error) {
	if d.git == nil {
		return nil, fmt.Errorf("dependency %q: git support unavailable", name)
	}
	result, _, err := d.git.Fetch(name, spec)
	if err != nil {
		return nil, err
	}
	d.logs = append(d.logs, fmt.Sprintf("fetched git dependency %s (%s)", result.Name, result.Version))
	rootDir := filepath.Join(d.git.cacheDir, "pkg", "src", sanitizeName(name), sanitizePathSegment(result.Version))
	manifestPath := filepath.Join(rootDir, "package.yml")
	var manifest *driver.Manifest
	if data, err := driver.LoadManifest(manifestPath); err == nil {
		manifest = data
		if manifest.Name != "" {
			result.Name = sanitizeName(manifest.Name)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("dependency %q: load manifest %s: %w", name, manifestPath, err)
	}
	return &resolvedPackage{
		pkg:      result,
		manifest: manifest,
		root:     rootDir,
	}, nil
}

func (d *dependencyInstaller) displayPath(path string) string {
	if d.manifestRoot != "" {
		if rel, err := filepath.Rel(d.manifestRoot, path); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return path
}

func lockedPackageEqual(a, b *driver.LockedPackage) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Name != b.Name || a.Version != b.Version || a.Source != b.Source || a.Checksum != b.Checksum {
		return false
	}
	if len(a.Dependencies) != len(b.Dependencies) {
		return false
	}
	for i := range a.Dependencies {
		if a.Dependencies[i].Name != b.Dependencies[i].Name || a.Dependencies[i].Version != b.Dependencies[i].Version {
			return false
		}
	}
	return true
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

func cloneDependencySpec(spec *driver.DependencySpec) *driver.DependencySpec {
	if spec == nil {
		return nil
	}
	clone := *spec
	if len(spec.Features) > 0 {
		clone.Features = append([]string{}, spec.Features...)
	}
	return &clone
}

func ascendToManifest(start string) (string, string) {
	dir := filepath.Clean(start)
	for {
		candidate := filepath.Join(dir, "package.yml")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return dir, candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", ""
}

func constraintContainsVersion(constraint, version string) bool {
	constraint = strings.TrimSpace(constraint)
	if constraint == "" || constraint == "*" {
		return true
	}
	if constraint == version {
		return true
	}
	return strings.Contains(constraint, version)
}

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
