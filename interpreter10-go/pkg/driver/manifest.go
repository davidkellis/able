package driver

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest represents the parsed contents of package.yml.
type Manifest struct {
	Path              string
	Name              string
	Version           string
	License           string
	Authors           []string
	Targets           map[string]*TargetSpec
	TargetOrder       []string
	Dependencies      map[string]*DependencySpec
	DevDependencies   map[string]*DependencySpec
	BuildDependencies map[string]*DependencySpec
	Workspace         map[string]any

	targetEntries []manifestTargetEntry
}

// TargetSpec describes a buildable target from the manifest.
type TargetSpec struct {
	Name         string
	OriginalName string
	Type         TargetType
	Main         string
	Dependencies map[string]*DependencySpec
}

type manifestTargetEntry struct {
	sanitized string
	spec      *TargetSpec
}

// TargetType enumerates supported target kinds.
type TargetType string

const (
	TargetTypeExecutable TargetType = "executable"
	TargetTypeLibrary    TargetType = "library"
	TargetTypeTest       TargetType = "test"
)

// DependencySpec describes a dependency descriptor in the manifest.
type DependencySpec struct {
	Version  string
	Git      string
	Rev      string
	Tag      string
	Branch   string
	Path     string
	Registry string
	Features []string
	Optional bool
}

// ValidationError aggregates manifest validation failures.
type ValidationError struct {
	Issues []string
}

func (e *ValidationError) Error() string {
	if len(e.Issues) == 0 {
		return "manifest: invalid configuration"
	}
	var b strings.Builder
	b.WriteString("manifest validation failed:")
	for _, issue := range e.Issues {
		b.WriteString("\n- ")
		b.WriteString(issue)
	}
	return b.String()
}

// LoadManifest parses package.yml from disk, returning a validated manifest.
func LoadManifest(path string) (*Manifest, error) {
	if path == "" {
		return nil, fmt.Errorf("manifest: empty path")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("manifest: resolve %s: %w", path, err)
	}
	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("manifest: open %s: %w", absPath, err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)

	var raw manifestFile
	if err := decoder.Decode(&raw); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("manifest: %s is empty", absPath)
		}
		return nil, fmt.Errorf("manifest: parse %s: %w", absPath, err)
	}

	manifest := raw.toManifest(absPath)
	if err := manifest.validate(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *Manifest) validate() error {
	var errs ValidationError
	if m.Name == "" {
		errs.Issues = append(errs.Issues, "name must be provided")
	}
	for i, author := range m.Authors {
		if author == "" {
			errs.Issues = append(errs.Issues, fmt.Sprintf("authors[%d] must be a non-empty string", i))
		}
	}

	targetNames := make(map[string]string, len(m.targetEntries))
	for _, entry := range m.targetEntries {
		if entry.spec == nil {
			continue
		}
		key := entry.sanitized
		target := entry.spec
		if target == nil {
			continue
		}
		if target.OriginalName == "" {
			errs.Issues = append(errs.Issues, "targets must not use empty keys")
			continue
		}
		if other, exists := targetNames[key]; exists {
			errs.Issues = append(errs.Issues, fmt.Sprintf("targets %q and %q collide after sanitization", other, target.OriginalName))
		} else {
			targetNames[key] = target.OriginalName
		}
		if target.Type == "" {
			errs.Issues = append(errs.Issues, fmt.Sprintf("target %q missing type", target.OriginalName))
		} else if !target.Type.IsValid() {
			errs.Issues = append(errs.Issues, fmt.Sprintf("target %q has unsupported type %q", target.OriginalName, target.Type))
		}
		if target.Type.RequiresMain() && target.Main == "" {
			errs.Issues = append(errs.Issues, fmt.Sprintf("target %q requires a main entrypoint", target.OriginalName))
		}
		for depName, dep := range target.Dependencies {
			if dep == nil {
				continue
			}
			dep.normalize()
			for _, issue := range dep.validate(false) {
				errs.Issues = append(errs.Issues, fmt.Sprintf("targets.%s.dependencies.%s: %s", target.OriginalName, depName, issue))
			}
		}
	}

	for groupName, deps := range map[string]map[string]*DependencySpec{
		"dependencies":       m.Dependencies,
		"dev_dependencies":   m.DevDependencies,
		"build_dependencies": m.BuildDependencies,
	} {
		for depName, dep := range deps {
			if dep == nil {
				continue
			}
			dep.normalize()
			for _, issue := range dep.validate(true) {
				errs.Issues = append(errs.Issues, fmt.Sprintf("%s.%s: %s", groupName, depName, issue))
			}
		}
	}

	if len(errs.Issues) > 0 {
		return &errs
	}
	return nil
}

// IsValid reports whether the target type is recognised.
func (t TargetType) IsValid() bool {
	switch t {
	case TargetTypeExecutable, TargetTypeLibrary, TargetTypeTest:
		return true
	default:
		return false
	}
}

// RequiresMain reports if the target requires a main entrypoint.
func (t TargetType) RequiresMain() bool {
	switch t {
	case TargetTypeExecutable, TargetTypeTest:
		return true
	default:
		return false
	}
}

var ErrNoExecutableTarget = errors.New("manifest: no executable targets defined")

// DefaultExecutableTarget returns the first executable target in manifest order.
func (m *Manifest) DefaultExecutableTarget() (*TargetSpec, error) {
	if m == nil {
		return nil, ErrNoExecutableTarget
	}
	for _, entry := range m.targetEntries {
		if entry.spec == nil {
			continue
		}
		if entry.spec.Type == TargetTypeExecutable {
			return entry.spec, nil
		}
	}
	return nil, ErrNoExecutableTarget
}

// FindTarget looks up a target by sanitized or original name.
func (m *Manifest) FindTarget(name string) (*TargetSpec, bool) {
	if m == nil {
		return nil, false
	}
	key := sanitizeSegment(strings.TrimSpace(name))
	if key != "" {
		if target, ok := m.Targets[key]; ok && target != nil {
			return target, true
		}
	}
	for _, entry := range m.targetEntries {
		if entry.spec == nil {
			continue
		}
		if strings.EqualFold(entry.spec.OriginalName, strings.TrimSpace(name)) {
			return entry.spec, true
		}
	}
	return nil, false
}

func (d *DependencySpec) normalize() {
	if d == nil {
		return
	}
	if len(d.Features) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(d.Features))
	out := make([]string, 0, len(d.Features))
	for _, feature := range d.Features {
		feature = sanitizeSegment(feature)
		if feature == "" {
			continue
		}
		if _, ok := seen[feature]; ok {
			continue
		}
		seen[feature] = struct{}{}
		out = append(out, feature)
	}
	sort.Strings(out)
	d.Features = out
}

func (d *DependencySpec) validate(requireSource bool) []string {
	var errs []string
	if d == nil {
		return errs
	}

	if d.Path != "" && (d.Version != "" || d.Git != "") {
		errs = append(errs, "path overrides cannot specify version or git source")
	}
	if d.Git != "" && d.Version != "" {
		errs = append(errs, "git dependencies cannot also specify version")
	}
	if d.Registry != "" && (d.Git != "" || d.Path != "") {
		errs = append(errs, "registry overrides apply only to registry-based version dependencies")
	}

	hasSource := d.Version != "" || d.Git != "" || d.Path != ""
	if requireSource && !hasSource {
		errs = append(errs, "must specify version, git, or path")
	}

	if d.Version != "" && !isValidVersionConstraint(d.Version) {
		errs = append(errs, fmt.Sprintf("invalid version constraint %q", d.Version))
	}
	return errs
}

var versionConstraintPattern = regexp.MustCompile(`^(~>|>=|<=|>|<|=|\^)?\s*[0-9]+(\.[0-9]+){0,2}([0-9A-Za-z\-\+\.]*)?$`)

func isValidVersionConstraint(input string) bool {
	s := strings.TrimSpace(input)
	if s == "" {
		return false
	}
	if s == "*" {
		return true
	}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" || !versionConstraintPattern.MatchString(part) {
			return false
		}
	}
	return true
}

type manifestFile struct {
	Name              string         `yaml:"name"`
	Version           string         `yaml:"version"`
	License           string         `yaml:"license"`
	Authors           stringList     `yaml:"authors"`
	Targets           targetMap      `yaml:"targets"`
	Dependencies      dependencyMap  `yaml:"dependencies"`
	DevDependencies   dependencyMap  `yaml:"dev_dependencies"`
	BuildDependencies dependencyMap  `yaml:"build_dependencies"`
	Workspace         map[string]any `yaml:"workspace"`
}

type targetYAML struct {
	Type         TargetType    `yaml:"type"`
	Main         string        `yaml:"main"`
	Dependencies dependencyMap `yaml:"dependencies"`
}

type targetMap struct {
	items []targetMapEntry
}

type targetMapEntry struct {
	name string
	spec *targetYAML
}

func (tm *targetMap) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == 0 {
		tm.items = nil
		return nil
	}
	if value.Kind == yaml.ScalarNode && value.Tag == "!!null" {
		tm.items = nil
		return nil
	}
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("manifest: targets must be a mapping")
	}
	items := make([]targetMapEntry, 0, len(value.Content)/2)
	for i := 0; i < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valueNode := value.Content[i+1]

		var key string
		if err := keyNode.Decode(&key); err != nil {
			return err
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("manifest: targets must not use empty keys")
		}
		entry := new(targetYAML)
		if err := valueNode.Decode(entry); err != nil {
			return fmt.Errorf("manifest: target %q: %w", key, err)
		}
		items = append(items, targetMapEntry{
			name: key,
			spec: entry,
		})
	}
	tm.items = items
	return nil
}

type dependencyMap map[string]*DependencySpec

type stringList []string

func (mf manifestFile) toManifest(path string) *Manifest {
	targetCapacity := len(mf.Targets.items)
	result := &Manifest{
		Path:              path,
		Name:              sanitizeSegment(strings.TrimSpace(mf.Name)),
		Version:           strings.TrimSpace(mf.Version),
		License:           strings.TrimSpace(mf.License),
		Authors:           mf.Authors.Clone(),
		Targets:           make(map[string]*TargetSpec, targetCapacity),
		TargetOrder:       make([]string, 0, targetCapacity),
		Dependencies:      cloneDependencyMap(mf.Dependencies),
		DevDependencies:   cloneDependencyMap(mf.DevDependencies),
		BuildDependencies: cloneDependencyMap(mf.BuildDependencies),
		Workspace:         mf.Workspace,
		targetEntries:     make([]manifestTargetEntry, 0, targetCapacity),
	}

	for _, dep := range result.Dependencies {
		if dep != nil {
			dep.Version = strings.TrimSpace(dep.Version)
			dep.Git = strings.TrimSpace(dep.Git)
			dep.Rev = strings.TrimSpace(dep.Rev)
			dep.Tag = strings.TrimSpace(dep.Tag)
			dep.Branch = strings.TrimSpace(dep.Branch)
			dep.Path = strings.TrimSpace(dep.Path)
			dep.Registry = strings.TrimSpace(dep.Registry)
		}
	}
	for _, dep := range result.DevDependencies {
		if dep != nil {
			dep.Version = strings.TrimSpace(dep.Version)
			dep.Git = strings.TrimSpace(dep.Git)
			dep.Rev = strings.TrimSpace(dep.Rev)
			dep.Tag = strings.TrimSpace(dep.Tag)
			dep.Branch = strings.TrimSpace(dep.Branch)
			dep.Path = strings.TrimSpace(dep.Path)
			dep.Registry = strings.TrimSpace(dep.Registry)
		}
	}
	for _, dep := range result.BuildDependencies {
		if dep != nil {
			dep.Version = strings.TrimSpace(dep.Version)
			dep.Git = strings.TrimSpace(dep.Git)
			dep.Rev = strings.TrimSpace(dep.Rev)
			dep.Tag = strings.TrimSpace(dep.Tag)
			dep.Branch = strings.TrimSpace(dep.Branch)
			dep.Path = strings.TrimSpace(dep.Path)
			dep.Registry = strings.TrimSpace(dep.Registry)
		}
	}

	seenTargets := make(map[string]struct{}, targetCapacity)
	for _, item := range mf.Targets.items {
		target := item.spec
		if target == nil {
			continue
		}
		original := strings.TrimSpace(item.name)
		if original == "" {
			continue
		}
		sanitized := sanitizeSegment(original)
		spec := &TargetSpec{
			Name:         sanitized,
			OriginalName: original,
			Type:         target.Type,
			Main:         strings.TrimSpace(target.Main),
			Dependencies: cloneDependencyMap(target.Dependencies),
		}
		if _, exists := result.Targets[sanitized]; !exists {
			result.Targets[sanitized] = spec
		}
		if _, exists := seenTargets[sanitized]; !exists {
			result.TargetOrder = append(result.TargetOrder, sanitized)
			seenTargets[sanitized] = struct{}{}
		}
		result.targetEntries = append(result.targetEntries, manifestTargetEntry{
			sanitized: sanitized,
			spec:      spec,
		})
	}
	return result
}

func cloneDependencyMap(src dependencyMap) map[string]*DependencySpec {
	if len(src) == 0 {
		return map[string]*DependencySpec{}
	}
	out := make(map[string]*DependencySpec, len(src))
	for name, dep := range src {
		if dep == nil {
			continue
		}
		copy := dep.clone()
		out[name] = copy
	}
	return out
}

func (d *DependencySpec) clone() *DependencySpec {
	if d == nil {
		return nil
	}
	copy := *d
	if len(d.Features) > 0 {
		copy.Features = append([]string{}, d.Features...)
	}
	return &copy
}

func (l stringList) Clone() []string {
	if len(l) == 0 {
		return nil
	}
	out := make([]string, 0, len(l))
	for _, item := range l {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func (l *stringList) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		if value.Tag == "!!null" || strings.TrimSpace(value.Value) == "" {
			*l = nil
			return nil
		}
		*l = stringList{strings.TrimSpace(value.Value)}
		return nil
	case yaml.SequenceNode:
		items := make([]string, 0, len(value.Content))
		for _, node := range value.Content {
			var str string
			if err := node.Decode(&str); err != nil {
				return err
			}
			str = strings.TrimSpace(str)
			if str == "" {
				continue
			}
			items = append(items, str)
		}
		*l = stringList(items)
		return nil
	case yaml.AliasNode:
		return l.UnmarshalYAML(value.Alias)
	case 0:
		*l = nil
		return nil
	default:
		return fmt.Errorf("manifest: expected string or sequence for list but found %s", value.ShortTag())
	}
}

func (dm *dependencyMap) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == 0 {
		*dm = make(dependencyMap)
		return nil
	}
	if value.Kind == yaml.ScalarNode && value.Tag == "!!null" {
		*dm = make(dependencyMap)
		return nil
	}
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("manifest: dependencies must be a mapping")
	}
	result := make(dependencyMap, len(value.Content)/2)
	for i := 0; i < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valNode := value.Content[i+1]

		var key string
		if err := keyNode.Decode(&key); err != nil {
			return err
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("manifest: dependency names must be non-empty")
		}
		var dep DependencySpec
		if err := dep.unmarshalYAML(valNode); err != nil {
			return fmt.Errorf("manifest: dependency %q: %w", key, err)
		}
		result[key] = dep.clone()
	}
	*dm = result
	return nil
}

func (d *DependencySpec) unmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		if value.Tag == "!!null" || strings.TrimSpace(value.Value) == "" {
			*d = DependencySpec{}
			return nil
		}
		*d = DependencySpec{Version: strings.TrimSpace(value.Value)}
		return nil
	case yaml.MappingNode:
		var raw struct {
			Version  string     `yaml:"version"`
			Git      string     `yaml:"git"`
			Rev      string     `yaml:"rev"`
			Tag      string     `yaml:"tag"`
			Branch   string     `yaml:"branch"`
			Path     string     `yaml:"path"`
			Registry string     `yaml:"registry"`
			Features stringList `yaml:"features"`
			Optional bool       `yaml:"optional"`
		}
		if err := value.Decode(&raw); err != nil {
			return err
		}
		*d = DependencySpec{
			Version:  strings.TrimSpace(raw.Version),
			Git:      strings.TrimSpace(raw.Git),
			Rev:      strings.TrimSpace(raw.Rev),
			Tag:      strings.TrimSpace(raw.Tag),
			Branch:   strings.TrimSpace(raw.Branch),
			Path:     strings.TrimSpace(raw.Path),
			Registry: strings.TrimSpace(raw.Registry),
			Features: raw.Features.Clone(),
			Optional: raw.Optional,
		}
		return nil
	case yaml.AliasNode:
		return d.unmarshalYAML(value.Alias)
	default:
		return fmt.Errorf("expected string or mapping, found %s", value.ShortTag())
	}
}
