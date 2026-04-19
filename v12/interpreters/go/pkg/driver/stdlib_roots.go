package driver

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// StdlibSourceClass records how a visible stdlib root entered resolution.
type StdlibSourceClass int

const (
	StdlibSourceUnknown StdlibSourceClass = iota
	StdlibSourceLockfile
	StdlibSourceOverride
	StdlibSourceEnv
	StdlibSourceCache
	StdlibSourceWorkspace
)

func (c StdlibSourceClass) String() string {
	switch c {
	case StdlibSourceLockfile:
		return "lockfile"
	case StdlibSourceOverride:
		return "override"
	case StdlibSourceEnv:
		return "env"
	case StdlibSourceCache:
		return "cache"
	case StdlibSourceWorkspace:
		return "workspace"
	default:
		return "unknown"
	}
}

func normalizeStdlibSourceClass(source StdlibSourceClass) StdlibSourceClass {
	switch source {
	case StdlibSourceLockfile, StdlibSourceOverride, StdlibSourceEnv, StdlibSourceCache, StdlibSourceWorkspace:
		return source
	default:
		return StdlibSourceUnknown
	}
}

// CanonicalizeStdlibCandidateRoot normalizes a discovered stdlib candidate to a
// comparable source-root path. Package roots collapse onto their src/
// directory when present.
func CanonicalizeStdlibCandidateRoot(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("driver: empty stdlib root")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("driver: resolve stdlib root %q: %w", path, err)
	}
	clean := filepath.Clean(abs)
	if info, err := os.Stat(filepath.Join(clean, "src")); err == nil && info.IsDir() {
		return filepath.Join(clean, "src"), nil
	}
	return clean, nil
}

func determineEntryRootMetadata(rootDir, rootName string, searchPaths []SearchPath) (RootKind, StdlibSourceClass) {
	kind := RootUser
	source := StdlibSourceUnknown
	if rootName == "able" {
		kind = RootStdlib
		source = StdlibSourceWorkspace
	} else if looksLikeStdlibPath(rootDir) || looksLikeKernelPath(rootDir) {
		kind = RootStdlib
	}
	for _, sp := range searchPaths {
		if sp.Kind != RootStdlib || !pathsOverlap(sp.Path, rootDir) {
			continue
		}
		kind = RootStdlib
		if normalized := normalizeStdlibSourceClass(sp.StdlibSource); normalized != StdlibSourceUnknown {
			source = normalized
		}
	}
	return kind, source
}

func determineSearchRootMetadata(root SearchPath, rootDir, rootName string) (RootKind, StdlibSourceClass) {
	kind := root.Kind
	source := normalizeStdlibSourceClass(root.StdlibSource)
	if kind != RootStdlib &&
		(rootName == "able" || looksLikeStdlibPath(rootDir) || looksLikeKernelPath(rootDir)) {
		kind = RootStdlib
	} else if kind != RootStdlib {
		kind = RootUser
	}
	return kind, source
}

type stdlibRootCandidate struct {
	index         int
	searchPath    SearchPath
	rootDir       string
	rootName      string
	canonicalRoot string
}

type stdlibRootGroup struct {
	canonicalRoot string
	candidates    []stdlibRootCandidate
	sources       map[StdlibSourceClass]struct{}
}

func (g *stdlibRootGroup) sortedSources() []StdlibSourceClass {
	if g == nil || len(g.sources) == 0 {
		return nil
	}
	out := make([]StdlibSourceClass, 0, len(g.sources))
	for source := range g.sources {
		out = append(out, source)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i] == out[j] {
			return false
		}
		return stdlibSourcePrecedence(out[i]) < stdlibSourcePrecedence(out[j])
	})
	return out
}

func (g *stdlibRootGroup) sourceLabel() string {
	sources := g.sortedSources()
	if len(sources) == 0 {
		return StdlibSourceUnknown.String()
	}
	names := make([]string, 0, len(sources))
	for _, source := range sources {
		names = append(names, source.String())
	}
	return strings.Join(names, ", ")
}

func ResolveCanonicalStdlibSearchPaths(searchPaths []SearchPath, manifestBased bool) ([]SearchPath, error) {
	candidates, err := collectVisibleStdlibCandidates(searchPaths)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return append([]SearchPath(nil), searchPaths...), nil
	}

	groups := make(map[string]*stdlibRootGroup, len(candidates))
	for _, candidate := range candidates {
		group, ok := groups[candidate.canonicalRoot]
		if !ok {
			group = &stdlibRootGroup{
				canonicalRoot: candidate.canonicalRoot,
				sources:       make(map[StdlibSourceClass]struct{}),
			}
			groups[candidate.canonicalRoot] = group
		}
		group.candidates = append(group.candidates, candidate)
		group.sources[normalizeStdlibSourceClass(candidate.searchPath.StdlibSource)] = struct{}{}
	}

	selected, err := selectCanonicalStdlibGroup(groups, manifestBased)
	if err != nil {
		return nil, err
	}
	if selected == nil {
		return append([]SearchPath(nil), searchPaths...), nil
	}

	conflicts := conflictingStdlibGroups(groups, selected.canonicalRoot)
	if len(conflicts) > 0 {
		return nil, buildStdlibCollisionError(selected, conflicts)
	}

	return filterCanonicalStdlibSearchPaths(searchPaths, candidates, selected.canonicalRoot), nil
}

func collectVisibleStdlibCandidates(searchPaths []SearchPath) ([]stdlibRootCandidate, error) {
	candidates := make([]stdlibRootCandidate, 0)
	for idx, searchPath := range searchPaths {
		rootDir, rootName, err := discoverRootForPath(searchPath.Path)
		if err != nil {
			return nil, err
		}
		if rootName != "able" {
			continue
		}
		canonicalRoot, err := CanonicalizeStdlibCandidateRoot(rootDir)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, stdlibRootCandidate{
			index:         idx,
			searchPath:    searchPath,
			rootDir:       rootDir,
			rootName:      rootName,
			canonicalRoot: canonicalRoot,
		})
	}
	return candidates, nil
}

func selectCanonicalStdlibGroup(groups map[string]*stdlibRootGroup, manifestBased bool) (*stdlibRootGroup, error) {
	var precedence []StdlibSourceClass
	if manifestBased {
		precedence = []StdlibSourceClass{
			StdlibSourceLockfile,
			StdlibSourceOverride,
			StdlibSourceEnv,
			StdlibSourceCache,
		}
	} else {
		precedence = []StdlibSourceClass{
			StdlibSourceOverride,
			StdlibSourceEnv,
			StdlibSourceCache,
		}
	}

	for _, source := range precedence {
		matches := groupsForSource(groups, source)
		if len(matches) == 0 {
			continue
		}
		if len(matches) > 1 {
			return nil, buildMultipleStdlibRootsError(source, matches)
		}
		return matches[0], nil
	}
	return nil, nil
}

func groupsForSource(groups map[string]*stdlibRootGroup, source StdlibSourceClass) []*stdlibRootGroup {
	out := make([]*stdlibRootGroup, 0)
	for _, group := range groups {
		if _, ok := group.sources[source]; !ok {
			continue
		}
		out = append(out, group)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].canonicalRoot < out[j].canonicalRoot
	})
	return out
}

func conflictingStdlibGroups(groups map[string]*stdlibRootGroup, selectedCanonical string) []*stdlibRootGroup {
	out := make([]*stdlibRootGroup, 0)
	for canonical, group := range groups {
		if canonical == selectedCanonical {
			continue
		}
		out = append(out, group)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].canonicalRoot < out[j].canonicalRoot
	})
	return out
}

func filterCanonicalStdlibSearchPaths(searchPaths []SearchPath, candidates []stdlibRootCandidate, selectedCanonical string) []SearchPath {
	selectedByIndex := make(map[int]string, len(candidates))
	for _, candidate := range candidates {
		selectedByIndex[candidate.index] = candidate.canonicalRoot
	}

	filtered := make([]SearchPath, 0, len(searchPaths))
	seenCanonical := make(map[string]struct{})
	for idx, searchPath := range searchPaths {
		canonical, ok := selectedByIndex[idx]
		if !ok {
			filtered = append(filtered, searchPath)
			continue
		}
		if canonical != selectedCanonical {
			continue
		}
		if _, ok := seenCanonical[canonical]; ok {
			continue
		}
		seenCanonical[canonical] = struct{}{}
		searchPath.Path = canonical
		filtered = append(filtered, searchPath)
	}
	return filtered
}

func buildMultipleStdlibRootsError(source StdlibSourceClass, groups []*stdlibRootGroup) error {
	var builder strings.Builder
	fmt.Fprintf(&builder, "stdlib collision: multiple %s-provided `name: able` roots are visible:", source.String())
	for _, group := range groups {
		fmt.Fprintf(&builder, "\n  %s", group.canonicalRoot)
	}
	if source == StdlibSourceEnv {
		fmt.Fprintf(&builder, "\nset only one stdlib root in ABLE_MODULE_PATHS / ABLE_PATH")
	}
	return fmt.Errorf("%s", builder.String())
}

func buildStdlibCollisionError(selected *stdlibRootGroup, conflicts []*stdlibRootGroup) error {
	var builder strings.Builder
	fmt.Fprintf(
		&builder,
		"stdlib collision: selected canonical stdlib root (%s):\n  %s",
		selected.sourceLabel(),
		selected.canonicalRoot,
	)
	for _, conflict := range conflicts {
		fmt.Fprintf(
			&builder,
			"\nconflicts with distinct visible stdlib root (%s):\n  %s",
			conflict.sourceLabel(),
			conflict.canonicalRoot,
		)
	}
	return fmt.Errorf("%s", builder.String())
}

func stdlibSourcePrecedence(source StdlibSourceClass) int {
	switch source {
	case StdlibSourceLockfile:
		return 0
	case StdlibSourceOverride:
		return 1
	case StdlibSourceEnv:
		return 2
	case StdlibSourceCache:
		return 3
	case StdlibSourceWorkspace:
		return 4
	default:
		return 5
	}
}
