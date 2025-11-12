package driver

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Lockfile models the package.lock contents.
type Lockfile struct {
	Path      string
	Root      string
	Generated string
	Tool      string
	Packages  []*LockedPackage
}

// LockedPackage captures a single resolved dependency entry.
type LockedPackage struct {
	Name         string
	Version      string
	Source       string
	Checksum     string
	Dependencies []LockedDependency
}

// LockedDependency identifies a dependency edge in the resolved graph.
type LockedDependency struct {
	Name    string
	Version string
}

// NewLockfile constructs a lockfile with metadata seeded for the provided root.
func NewLockfile(root, tool string) *Lockfile {
	return &Lockfile{
		Root:      sanitizeSegment(root),
		Generated: time.Now().UTC().Format(time.RFC3339),
		Tool:      strings.TrimSpace(tool),
		Packages:  []*LockedPackage{},
	}
}

// LoadLockfile parses package.lock from disk.
func LoadLockfile(path string) (*Lockfile, error) {
	if path == "" {
		return nil, fmt.Errorf("lockfile: empty path")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("lockfile: resolve %s: %w", path, err)
	}
	file, err := os.Open(abs)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var raw lockfileDisk
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("lockfile: parse %s: %w", abs, err)
	}

	lock := raw.toLockfile()
	lock.Path = abs
	lock.normalize()
	return lock, nil
}

// WriteLockfile serialises the lockfile back to disk, refreshing metadata.
func WriteLockfile(lock *Lockfile, path string) error {
	if lock == nil {
		return fmt.Errorf("lockfile: nil lockfile")
	}
	if path == "" {
		if lock.Path == "" {
			return fmt.Errorf("lockfile: missing path")
		}
		path = lock.Path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("lockfile: resolve %s: %w", path, err)
	}

	if lock.Generated == "" {
		lock.Generated = time.Now().UTC().Format(time.RFC3339)
	}
	lock.Path = abs
	lock.normalize()

	data := lock.toDisk()
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("lockfile: marshal %s: %w", abs, err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("lockfile: encoder close: %w", err)
	}
	if err := os.WriteFile(abs, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("lockfile: write %s: %w", abs, err)
	}
	return nil
}

func (l *Lockfile) normalize() {
	if l == nil {
		return
	}
	l.Root = sanitizeSegment(l.Root)
	l.Tool = strings.TrimSpace(l.Tool)
	sort.SliceStable(l.Packages, func(i, j int) bool {
		return l.Packages[i].Name < l.Packages[j].Name
	})
	for _, pkg := range l.Packages {
		if pkg == nil {
			continue
		}
		pkg.Name = sanitizeSegment(pkg.Name)
		pkg.Version = strings.TrimSpace(pkg.Version)
		pkg.Source = strings.TrimSpace(pkg.Source)
		pkg.Checksum = strings.TrimSpace(pkg.Checksum)
		sort.SliceStable(pkg.Dependencies, func(i, j int) bool {
			if pkg.Dependencies[i].Name == pkg.Dependencies[j].Name {
				return pkg.Dependencies[i].Version < pkg.Dependencies[j].Version
			}
			return pkg.Dependencies[i].Name < pkg.Dependencies[j].Name
		})
		for k := range pkg.Dependencies {
			pkg.Dependencies[k].Name = sanitizeSegment(pkg.Dependencies[k].Name)
			pkg.Dependencies[k].Version = strings.TrimSpace(pkg.Dependencies[k].Version)
		}
	}
}

func (l *Lockfile) toDisk() lockfileDisk {
	pkgs := make([]lockfilePackage, 0, len(l.Packages))
	for _, pkg := range l.Packages {
		if pkg == nil {
			continue
		}
		deps := make([]lockfileDependency, 0, len(pkg.Dependencies))
		for _, dep := range pkg.Dependencies {
			deps = append(deps, lockfileDependency{
				Name:    dep.Name,
				Version: dep.Version,
			})
		}
		pkgs = append(pkgs, lockfilePackage{
			Name:         pkg.Name,
			Version:      pkg.Version,
			Source:       pkg.Source,
			Checksum:     pkg.Checksum,
			Dependencies: deps,
		})
	}
	return lockfileDisk{
		Root:      l.Root,
		Generated: l.Generated,
		Tool:      l.Tool,
		Packages:  pkgs,
	}
}

type lockfileDisk struct {
	Root      string            `yaml:"root"`
	Generated string            `yaml:"generated"`
	Tool      string            `yaml:"tool"`
	Packages  []lockfilePackage `yaml:"packages"`
}

type lockfilePackage struct {
	Name         string               `yaml:"name"`
	Version      string               `yaml:"version"`
	Source       string               `yaml:"source"`
	Checksum     string               `yaml:"checksum"`
	Dependencies []lockfileDependency `yaml:"dependencies"`
}

type lockfileDependency struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func (d lockfileDisk) toLockfile() *Lockfile {
	lock := &Lockfile{
		Root:      sanitizeSegment(d.Root),
		Generated: strings.TrimSpace(d.Generated),
		Tool:      strings.TrimSpace(d.Tool),
		Packages:  make([]*LockedPackage, 0, len(d.Packages)),
	}
	for _, pkg := range d.Packages {
		deps := make([]LockedDependency, 0, len(pkg.Dependencies))
		for _, dep := range pkg.Dependencies {
			deps = append(deps, LockedDependency{
				Name:    sanitizeSegment(dep.Name),
				Version: strings.TrimSpace(dep.Version),
			})
		}
		lock.Packages = append(lock.Packages, &LockedPackage{
			Name:         sanitizeSegment(pkg.Name),
			Version:      strings.TrimSpace(pkg.Version),
			Source:       strings.TrimSpace(pkg.Source),
			Checksum:     strings.TrimSpace(pkg.Checksum),
			Dependencies: deps,
		})
	}
	lock.normalize()
	return lock
}
