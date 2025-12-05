package driver

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/parser"
)

type RootKind int

const (
	RootUser RootKind = iota
	RootStdlib
)

// SearchPath describes a module search root.
type SearchPath struct {
	Path string
	Kind RootKind
}

// Module aggregates the Able source for a fully qualified package.
type Module struct {
	Package     string
	AST         *ast.Module
	Files       []string
	Imports     []string
	NodeOrigins map[ast.Node]string
}

// Program contains the entry package and dependency-ordered modules.
type Program struct {
	Entry   *Module
	Modules []*Module
}

type packageLocation struct {
	rootDir  string
	rootName string
	kind     RootKind
	files    []string
}

type packageOrigin struct {
	root string
	kind RootKind
}

type rootInfo struct {
	rootDir  string
	rootName string
	kind     RootKind
}

// Loader wires Able source files into aggregated modules.
type Loader struct {
	parser      *parser.ModuleParser
	searchPaths []SearchPath
}

// NewLoader constructs a loader with optional extra search paths (reserved for future use).
func NewLoader(searchPaths []SearchPath) (*Loader, error) {
	mp, err := parser.NewModuleParser()
	if err != nil {
		return nil, err
	}
	unique := make([]SearchPath, 0, len(searchPaths))
	seen := make(map[string]struct{}, len(searchPaths))
	for _, sp := range searchPaths {
		if sp.Path == "" {
			continue
		}
		abs, err := filepath.Abs(sp.Path)
		if err != nil {
			return nil, fmt.Errorf("loader: resolve search path %q: %w", sp.Path, err)
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		kind := sp.Kind
		if kind != RootStdlib {
			kind = RootUser
		}
		unique = append(unique, SearchPath{Path: abs, Kind: kind})
	}
	return &Loader{parser: mp, searchPaths: unique}, nil
}

// Close releases parser resources.
func (l *Loader) Close() {
	if l == nil {
		return
	}
	if l.parser != nil {
		l.parser.Close()
		l.parser = nil
	}
}

// Load aggregates the entry package and its dependencies according to the v10 package rules.
func (l *Loader) Load(entry string) (*Program, error) {
	if l == nil || l.parser == nil {
		return nil, fmt.Errorf("loader: closed")
	}
	if entry == "" {
		return nil, fmt.Errorf("loader: empty entry path")
	}
	entryPath, err := filepath.Abs(entry)
	if err != nil {
		return nil, fmt.Errorf("loader: resolve entry path: %w", err)
	}
	info, err := os.Stat(entryPath)
	if err != nil {
		return nil, fmt.Errorf("loader: stat entry %s: %w", entryPath, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("loader: entry path %s is a directory", entryPath)
	}

	rootDir, rootName, err := l.discoverRoot(entryPath)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(entryPath, rootDir) {
		return nil, fmt.Errorf("loader: entry file %s is outside package root %s", entryPath, rootDir)
	}

	entryKind := RootUser
	for _, sp := range l.searchPaths {
		if sp.Kind == RootStdlib && pathsOverlap(sp.Path, rootDir) {
			entryKind = RootStdlib
			break
		}
	}
	entryRoot := rootInfo{rootDir: rootDir, rootName: rootName, kind: entryKind}
	if ok, err := ensureNamespaceAllowed(entryRoot, false); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("loader: package namespace 'able.*' is reserved for the standard library (path: %s)", entryRoot.rootDir)
	}

	entryPackages, fileIndex, err := indexSourceFiles(rootDir, rootName, entryKind)
	if err != nil {
		return nil, err
	}
	entryPackage, ok := fileIndex[entryPath]
	if !ok {
		return nil, fmt.Errorf("loader: failed to resolve package for entry file %s", entryPath)
	}

	pkgIndex := make(map[string]*packageLocation, len(entryPackages))
	origins := make(map[string]packageOrigin)
	if err := registerPackages(pkgIndex, entryPackages, entryRoot, origins); err != nil {
		return nil, err
	}

	if err := l.indexAdditionalRoots(pkgIndex, origins, entryRoot); err != nil {
		return nil, err
	}

	loaded := make(map[string]*Module, len(pkgIndex))
	inProgress := make(map[string]bool)
	var ordered []*Module

	var loadPackage func(string) (*Module, error)
	loadPackage = func(name string) (*Module, error) {
		if mod, ok := loaded[name]; ok {
			return mod, nil
		}
		if inProgress[name] {
			return nil, fmt.Errorf("loader: import cycle detected at package %s", name)
		}
		loc, ok := pkgIndex[name]
		if !ok || loc == nil || len(loc.files) == 0 {
			return nil, fmt.Errorf("loader: package %s not found", name)
		}
		inProgress[name] = true
		defer delete(inProgress, name)

		fileMods := make([]*fileModule, 0, len(loc.files))
		for _, path := range loc.files {
			fm, err := l.parseFile(path, loc.rootDir, loc.rootName, loc.kind)
			if err != nil {
				return nil, err
			}
			if fm.packageName != name {
				return nil, fmt.Errorf("loader: file %s resolves to package %s, expected %s", path, fm.packageName, name)
			}
			fileMods = append(fileMods, fm)
		}

		mod, err := combinePackage(name, fileMods)
		if err != nil {
			return nil, err
		}
		loaded[name] = mod

		for _, dep := range mod.Imports {
			if dep == name {
				continue
			}
			if _, ok := pkgIndex[dep]; !ok {
				return nil, fmt.Errorf("loader: package %s imports unknown package %s", name, dep)
			}
			if _, err := loadPackage(dep); err != nil {
				return nil, err
			}
		}

		ordered = append(ordered, mod)
		return mod, nil
	}

	entryModule, err := loadPackage(entryPackage)
	if err != nil {
		return nil, err
	}

	return &Program{Entry: entryModule, Modules: ordered}, nil
}

type fileModule struct {
	path        string
	packageName string
	ast         *ast.Module
	imports     []string
}

func (l *Loader) indexAdditionalRoots(pkgIndex map[string]*packageLocation, origins map[string]packageOrigin, entryRoot rootInfo) error {
	if len(l.searchPaths) == 0 {
		return nil
	}
	used := make(map[string]struct{}, len(l.searchPaths)+1)
	usedList := make([]string, 0, len(l.searchPaths)+1)
	if entryRoot.rootDir != "" {
		clean := filepath.Clean(entryRoot.rootDir)
		used[clean] = struct{}{}
		usedList = append(usedList, clean)
	}

	for _, root := range l.searchPaths {
		abs, rootName, err := discoverRootForPath(root.Path)
		if err != nil {
			return err
		}
		clean := filepath.Clean(abs)
		overlaps := false
		for _, seen := range usedList {
			if pathsOverlap(seen, clean) {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}
		used[clean] = struct{}{}
		usedList = append(usedList, clean)
		kind := root.Kind
		if kind != RootStdlib {
			kind = RootUser
		}
		info := rootInfo{rootDir: abs, rootName: rootName, kind: kind}
		if ok, err := ensureNamespaceAllowed(info, true); err != nil {
			return err
		} else if !ok {
			continue
		}
		packages, _, err := indexSourceFiles(abs, rootName, kind)
		if err != nil {
			return err
		}
		if len(packages) == 0 {
			continue
		}
		if err := registerPackages(pkgIndex, packages, info, origins); err != nil {
			return err
		}
	}
	return nil
}

func ensureNamespaceAllowed(root rootInfo, allowSkip bool) (bool, error) {
	if root.rootName == "able" && root.kind != RootStdlib {
		if allowSkip {
			return false, nil
		}
		return false, fmt.Errorf("loader: package namespace 'able.*' is reserved for the standard library (path: %s)", root.rootDir)
	}
	return true, nil
}

func registerPackages(pkgIndex map[string]*packageLocation, packages map[string][]string, root rootInfo, origins map[string]packageOrigin) error {
	for name, files := range packages {
		if len(files) == 0 {
			continue
		}
		if existing, ok := origins[name]; ok {
			return fmt.Errorf("loader: package %s found in multiple roots (%s, %s)", name, existing.root, root.rootDir)
		}
		origins[name] = packageOrigin{root: root.rootDir, kind: root.kind}
		pkgIndex[name] = &packageLocation{
			rootDir:  root.rootDir,
			rootName: root.rootName,
			kind:     root.kind,
			files:    files,
		}
	}
	return nil
}

func pathsOverlap(a, b string) bool {
	aClean := filepath.Clean(a)
	bClean := filepath.Clean(b)
	return containsPathPrefix(aClean, bClean) || containsPathPrefix(bClean, aClean)
}

func containsPathPrefix(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != "..")
}

func (l *Loader) parseFile(path, rootDir, rootPackage string, kind RootKind) (*fileModule, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loader: read %s: %w", path, err)
	}
	moduleAST, err := l.parser.ParseModule(source)
	if err != nil {
		return nil, fmt.Errorf("loader: parse %s: %w", path, err)
	}

	segments, isPrivate, err := computePackageSegments(rootDir, rootPackage, path, moduleAST, kind)
	if err != nil {
		return nil, err
	}
	pkgName := strings.Join(segments, ".")

	moduleAST.Package = ast.NewPackageStatement(buildIdentifiers(segments), isPrivate)

	importSet := make(map[string]struct{})
	for _, imp := range moduleAST.Imports {
		if imp == nil {
			continue
		}
		name := joinIdentifiers(imp.PackagePath)
		if name == "" {
			continue
		}
		importSet[name] = struct{}{}
	}
	collectDynImports(moduleAST, importSet, make(map[uintptr]struct{}))
	imports := make([]string, 0, len(importSet))
	for name := range importSet {
		imports = append(imports, name)
	}
	sort.Strings(imports)

	return &fileModule{
		path:        path,
		packageName: pkgName,
		ast:         moduleAST,
		imports:     imports,
	}, nil
}

func (l *Loader) discoverRoot(entryPath string) (string, string, error) {
	dir := filepath.Dir(entryPath)
	for {
		cfgPath := filepath.Join(dir, "package.yml")
		if _, err := os.Stat(cfgPath); err == nil {
			name, err := readPackageName(cfgPath)
			if err != nil {
				return "", "", err
			}
			if name == "" {
				return "", "", fmt.Errorf("loader: package.yml at %s missing name", cfgPath)
			}
			name = sanitizeSegment(name)
			return dir, name, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	fallbackRoot := filepath.Dir(entryPath)
	fallbackName := sanitizeSegment(filepath.Base(fallbackRoot))
	return fallbackRoot, fallbackName, nil
}

func readPackageName(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("loader: read package.yml %s: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "name:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			value = strings.Trim(value, "\"'")
			return value, nil
		}
	}
	return "", nil
}

func indexSourceFiles(rootDir, rootPackage string, kind RootKind) (map[string][]string, map[string]string, error) {
	packages := make(map[string][]string)
	fileToPackage := make(map[string]string)
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".able" {
			return nil
		}
		declared, err := scanPackageDeclaration(path)
		if err != nil {
			return err
		}
		segments, err := resolvePackageSegments(rootDir, rootPackage, path, declared, kind)
		if err != nil {
			return err
		}
		pkgName := strings.Join(segments, ".")
		packages[pkgName] = append(packages[pkgName], path)
		fileToPackage[path] = pkgName
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("loader: traverse %s: %w", rootDir, err)
	}
	for pkg, files := range packages {
		sort.Strings(files)
		packages[pkg] = files
	}
	return packages, fileToPackage, nil
}

func scanPackageDeclaration(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("loader: open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if idx := strings.Index(trimmed, "##"); idx >= 0 {
			trimmed = strings.TrimSpace(trimmed[:idx])
		}
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "private ") {
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "private"))
		}
		if !strings.HasPrefix(trimmed, "package ") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "package"))
		rest = strings.TrimSuffix(rest, ";")
		rest = strings.TrimSpace(rest)
		if rest == "" {
			return nil, nil
		}
		parts := strings.Split(rest, ".")
		segments := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			segments = append(segments, part)
		}
		return segments, nil
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("loader: read %s: %w", path, err)
	}
	return nil, nil
}

func buildPackageSegments(rootDir, rootPackage, filePath string, declared []string) ([]string, error) {
	rel, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		return nil, fmt.Errorf("loader: compute relative path for %s: %w", filePath, err)
	}
	rel = filepath.ToSlash(rel)
	relDir := filepath.Dir(rel)

	segments := []string{sanitizeSegment(rootPackage)}
	if relDir != "." && relDir != "/" {
		for _, part := range strings.Split(relDir, "/") {
			part = strings.TrimSpace(part)
			if part == "" || part == "." {
				continue
			}
			segments = append(segments, sanitizeSegment(part))
		}
	}
	segments = append(segments, declared...)
	return segments, nil
}

func computePackageSegments(rootDir, rootPackage, filePath string, module *ast.Module, kind RootKind) ([]string, bool, error) {
	var declared []string
	isPrivate := false
	if module.Package != nil {
		isPrivate = module.Package.IsPrivate
		for _, part := range module.Package.NamePath {
			if part == nil || part.Name == "" {
				continue
			}
			declared = append(declared, part.Name)
		}
	}
	segments, err := resolvePackageSegments(rootDir, rootPackage, filePath, declared, kind)
	if err != nil {
		return nil, false, err
	}
	return segments, isPrivate, nil
}

func resolvePackageSegments(rootDir, rootPackage, filePath string, declared []string, kind RootKind) ([]string, error) {
	if kind == RootStdlib && len(declared) > 0 {
		segments := []string{sanitizeSegment(rootPackage)}
		for _, part := range declared {
			if part == "" {
				continue
			}
			segments = append(segments, sanitizeSegment(part))
		}
		return segments, nil
	}
	return buildPackageSegments(rootDir, rootPackage, filePath, declared)
}

func discoverRootForPath(path string) (string, string, error) {
	if path == "" {
		return "", "", fmt.Errorf("loader: empty search path")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("loader: resolve search path %q: %w", path, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", "", fmt.Errorf("loader: stat search path %s: %w", abs, err)
	}
	if !info.IsDir() {
		return "", "", fmt.Errorf("loader: search path %s is not a directory", abs)
	}
	name, err := findManifestName(abs)
	if err != nil {
		return "", "", err
	}
	if name == "" {
		name = sanitizeSegment(filepath.Base(abs))
		if name == "" {
			name = "pkg"
		}
	} else {
		name = sanitizeSegment(name)
	}
	return abs, name, nil
}

func findManifestName(start string) (string, error) {
	dir := filepath.Clean(start)
	for {
		candidate := filepath.Join(dir, "package.yml")
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			name, readErr := readPackageName(candidate)
			if readErr != nil {
				return "", readErr
			}
			return name, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("loader: stat %s: %w", candidate, err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", nil
}

func combinePackage(packageName string, files []*fileModule) (*Module, error) {
	if len(files) == 0 {
		return nil, errors.New("loader: combinePackage called with no files")
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].path < files[j].path
	})

	var pkgStmt *ast.PackageStatement
	importSeen := make(map[string]struct{})
	importNodeSeen := make(map[string]struct{})
	var importNodes []*ast.ImportStatement
	var importNames []string
	var body []ast.Statement
	filePaths := make([]string, 0, len(files))
	origins := make(map[ast.Node]string)

	for _, fm := range files {
		filePaths = append(filePaths, fm.path)
		ast.AnnotateOrigins(fm.ast, fm.path, origins)
		if fm.ast.Package != nil && pkgStmt == nil {
			pkgStmt = ast.NewPackageStatement(copyIdentifiers(fm.ast.Package.NamePath), fm.ast.Package.IsPrivate)
		}
		for _, imp := range fm.ast.Imports {
			if imp == nil {
				continue
			}
			key := importKey(imp)
			if _, ok := importNodeSeen[key]; ok {
				continue
			}
			importNodeSeen[key] = struct{}{}
			importNodes = append(importNodes, imp)
		}
		for _, name := range fm.imports {
			if name == packageName {
				continue
			}
			if _, ok := importSeen[name]; ok {
				continue
			}
			importSeen[name] = struct{}{}
			importNames = append(importNames, name)
		}
		body = append(body, fm.ast.Body...)
	}
	sort.Strings(importNames)

	if pkgStmt == nil {
		pkgStmt = ast.NewPackageStatement(buildIdentifiers(strings.Split(packageName, ".")), false)
	}

	module := ast.NewModule(body, importNodes, pkgStmt)
	primaryPath := ""
	if len(filePaths) > 0 {
		primaryPath = filePaths[0]
	}
	ast.AnnotateOrigins(module, primaryPath, origins)
	return &Module{
		Package:     packageName,
		AST:         module,
		Files:       filePaths,
		Imports:     importNames,
		NodeOrigins: origins,
	}, nil
}

func collectDynImports(node ast.Node, into map[string]struct{}, visited map[uintptr]struct{}) {
	collectDynImportsFromValue(reflect.ValueOf(node), into, visited)
}

func collectDynImportsFromValue(val reflect.Value, into map[string]struct{}, visited map[uintptr]struct{}) {
	if !val.IsValid() {
		return
	}
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return
		}
		if ptr := val.Pointer(); ptr != 0 {
			if _, ok := visited[ptr]; ok {
				return
			}
			visited[ptr] = struct{}{}
		}
		if node, ok := val.Interface().(ast.Node); ok {
			if dyn, ok := node.(*ast.DynImportStatement); ok && dyn != nil {
				if name := joinIdentifiers(dyn.PackagePath); name != "" {
					into[name] = struct{}{}
				}
			}
		}
		collectDynImportsFromValue(val.Elem(), into, visited)
		return
	}
	if val.Kind() == reflect.Interface {
		if val.IsNil() {
			return
		}
		collectDynImportsFromValue(val.Elem(), into, visited)
		return
	}
	switch val.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			collectDynImportsFromValue(val.Field(i), into, visited)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			collectDynImportsFromValue(val.Index(i), into, visited)
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			collectDynImportsFromValue(val.MapIndex(key), into, visited)
		}
	}
}

func joinIdentifiers(ids []*ast.Identifier) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == nil || id.Name == "" {
			continue
		}
		parts = append(parts, id.Name)
	}
	return strings.Join(parts, ".")
}

func buildIdentifiers(parts []string) []*ast.Identifier {
	out := make([]*ast.Identifier, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		out = append(out, ast.NewIdentifier(part))
	}
	return out
}

func copyIdentifiers(ids []*ast.Identifier) []*ast.Identifier {
	if len(ids) == 0 {
		return nil
	}
	out := make([]*ast.Identifier, 0, len(ids))
	for _, id := range ids {
		if id == nil {
			continue
		}
		out = append(out, ast.NewIdentifier(id.Name))
	}
	return out
}

func importKey(imp *ast.ImportStatement) string {
	if imp == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(joinIdentifiers(imp.PackagePath))
	sb.WriteString("|")
	if imp.IsWildcard {
		sb.WriteString("*")
	}
	sb.WriteString("|")
	if imp.Alias != nil {
		sb.WriteString(imp.Alias.Name)
	}
	if len(imp.Selectors) > 0 {
		sb.WriteString("|")
		for _, sel := range imp.Selectors {
			if sel == nil || sel.Name == nil {
				continue
			}
			sb.WriteString(sel.Name.Name)
			if sel.Alias != nil {
				sb.WriteString("::")
				sb.WriteString(sel.Alias.Name)
			}
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func sanitizeSegment(seg string) string {
	seg = strings.TrimSpace(seg)
	seg = strings.ReplaceAll(seg, "-", "_")
	return seg
}
