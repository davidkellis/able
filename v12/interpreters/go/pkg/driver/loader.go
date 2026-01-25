package driver

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/parser"
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
	DynImports  []string
	NodeOrigins map[ast.Node]string
}

// Program contains the entry package and dependency-ordered modules.
type Program struct {
	Entry   *Module
	Modules []*Module
}

// LoadOptions configures optional loading behavior.
type LoadOptions struct {
	IncludePackages []string
}

type packageLocation struct {
	rootDir  string
	rootName string
	kind     RootKind
	files    []string
}

type packageOrigin struct {
	root     string
	rootName string
	kind     RootKind
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

// Load aggregates the entry package and its dependencies according to the v12 package rules.
func (l *Loader) Load(entry string) (*Program, error) {
	return l.LoadWithOptions(entry, LoadOptions{})
}

// LoadWithOptions aggregates the entry package and additional packages.
func (l *Loader) LoadWithOptions(entry string, options LoadOptions) (*Program, error) {
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
	if rootName == "able" || looksLikeStdlibPath(rootDir) || looksLikeKernelPath(rootDir) {
		entryKind = RootStdlib
	}
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
		for _, dep := range mod.DynImports {
			if dep == name {
				continue
			}
			if _, ok := pkgIndex[dep]; !ok {
				continue
			}
			if _, err := loadPackage(dep); err != nil {
				return nil, err
			}
		}

		ordered = append(ordered, mod)
		return mod, nil
	}

	include := make(map[string]struct{})
	for _, name := range options.IncludePackages {
		if name == "" {
			continue
		}
		include[name] = struct{}{}
	}
	for _, name := range collectKernelPackages(origins) {
		include[name] = struct{}{}
	}
	include[entryPackage] = struct{}{}

	for name := range include {
		if _, err := loadPackage(name); err != nil {
			return nil, err
		}
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
	dynImports  []string
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
		kind := root.Kind
		clean := filepath.Clean(abs)
		overlaps := false
		for _, seen := range usedList {
			if pathsOverlap(seen, clean) {
				overlaps = true
				break
			}
		}
		if overlaps {
			if kind == RootStdlib && entryRoot.kind == RootStdlib {
				continue
			}
			if kind != RootStdlib {
				continue
			}
		}
		used[clean] = struct{}{}
		usedList = append(usedList, clean)
		if kind != RootStdlib &&
			(rootName == "able" || looksLikeStdlibPath(abs) || looksLikeKernelPath(abs)) {
			kind = RootStdlib
		} else if kind != RootStdlib {
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
	if root.rootName == "able" {
		return true, nil
	}
	if root.kind == RootStdlib {
		return true, nil
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
		origins[name] = packageOrigin{root: root.rootDir, rootName: root.rootName, kind: root.kind}
		pkgIndex[name] = &packageLocation{
			rootDir:  root.rootDir,
			rootName: root.rootName,
			kind:     root.kind,
			files:    files,
		}
	}
	return nil
}

func collectKernelPackages(origins map[string]packageOrigin) []string {
	names := make([]string, 0, len(origins))
	for name, origin := range origins {
		if isKernelOrigin(origin) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func isKernelOrigin(origin packageOrigin) bool {
	if origin.rootName == "kernel" {
		return true
	}
	return looksLikeKernelPath(origin.root)
}

func pathsOverlap(a, b string) bool {
	aClean := filepath.Clean(a)
	bClean := filepath.Clean(b)
	return containsPathPrefix(aClean, bClean) || containsPathPrefix(bClean, aClean)
}

func looksLikeStdlibPath(path string) bool {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))
	for _, part := range parts {
		lower := strings.ToLower(part)
		if lower == "stdlib" || strings.HasPrefix(lower, "stdlib_") || lower == "able-stdlib" || lower == "able_stdlib" {
			return true
		}
	}
	return false
}

func looksLikeKernelPath(path string) bool {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))
	for _, part := range parts {
		switch strings.ToLower(part) {
		case "kernel", "ablekernel", "able_kernel":
			return true
		}
	}
	return false
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
		var parseErr *parser.ParseError
		if errors.As(err, &parseErr) {
			return nil, &ParserDiagnosticError{
				Diagnostic: ParserDiagnostic{
					Severity: SeverityError,
					Message:  parseErr.Message,
					Location: DiagnosticLocation{
						Path:      path,
						Line:      parseErr.Location.Line,
						Column:    parseErr.Location.Column,
						EndLine:   parseErr.Location.EndLine,
						EndColumn: parseErr.Location.EndColumn,
					},
				},
			}
		}
		return nil, fmt.Errorf("loader: parse %s: %w", path, err)
	}

	segments, isPrivate, err := computePackageSegments(rootDir, rootPackage, path, moduleAST, kind)
	if err != nil {
		return nil, err
	}
	pkgName := strings.Join(segments, ".")

	moduleAST.Package = ast.NewPackageStatement(buildIdentifiers(segments), isPrivate)

	importSet := make(map[string]struct{})
	dynImportSet := make(map[string]struct{})
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
	collectDynImports(moduleAST, dynImportSet, make(map[uintptr]struct{}))
	imports := make([]string, 0, len(importSet))
	for name := range importSet {
		imports = append(imports, name)
	}
	sort.Strings(imports)
	dynImports := make([]string, 0, len(dynImportSet))
	for name := range dynImportSet {
		dynImports = append(dynImports, name)
	}
	sort.Strings(dynImports)

	return &fileModule{
		path:        path,
		packageName: pkgName,
		ast:         moduleAST,
		imports:     imports,
		dynImports:  dynImports,
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
			if d.Name() == "quarantine" {
				return fs.SkipDir
			}
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
		if len(parts) > 1 {
			return nil, fmt.Errorf("loader: package declaration must be unqualified in %s", path)
		}
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
	return buildPackageSegmentsWithBase([]string{sanitizeSegment(rootPackage)}, rootDir, filePath, declared)
}

func buildPackageSegmentsWithBase(base []string, rootDir, filePath string, declared []string) ([]string, error) {
	rel, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		return nil, fmt.Errorf("loader: compute relative path for %s: %w", filePath, err)
	}
	rel = filepath.ToSlash(rel)
	relDir := filepath.Dir(rel)

	segments := append([]string{}, base...)
	if relDir != "." && relDir != "/" {
		for _, part := range strings.Split(relDir, "/") {
			part = strings.TrimSpace(part)
			if part == "" || part == "." {
				continue
			}
			segments = append(segments, sanitizeSegment(part))
		}
	}
	for _, part := range declared {
		clean := sanitizeSegment(part)
		if clean == "" {
			continue
		}
		segments = append(segments, clean)
	}
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
	if rootPackage == "kernel" || looksLikeKernelPath(rootDir) {
		return buildPackageSegmentsWithBase(
			[]string{sanitizeSegment("able"), sanitizeSegment("kernel")},
			rootDir,
			filePath,
			declared,
		)
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
	dynImportSeen := make(map[string]struct{})
	var dynImportNames []string
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
		for _, name := range fm.dynImports {
			if name == packageName {
				continue
			}
			if _, ok := dynImportSeen[name]; ok {
				continue
			}
			dynImportSeen[name] = struct{}{}
			dynImportNames = append(dynImportNames, name)
		}
		body = append(body, fm.ast.Body...)
	}
	sort.Strings(importNames)
	sort.Strings(dynImportNames)

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
		DynImports:  dynImportNames,
		NodeOrigins: origins,
	}, nil
}
