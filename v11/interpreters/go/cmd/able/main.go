package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"able/interpreter10-go/pkg/driver"
	"able/interpreter10-go/pkg/interpreter"
	"able/interpreter10-go/pkg/runtime"
)

const cliToolVersion = "able-cli 0.0.0-dev"

var errManifestNotFound = errors.New("package.yml not found")

type executionMode int

const (
	modeRun executionMode = iota
	modeCheck
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 1
	}

	switch args[0] {
	case "--help", "-h":
		printUsage()
		return 0
	case "--version", "-V", "version":
		fmt.Fprintln(os.Stdout, cliToolVersion)
		return 0
	case "run":
		return runEntry(args[1:])
	case "check":
		return runCheck(args[1:])
	case "deps":
		return runDeps(args[1:])
	default:
		return runEntry(args)
	}
}

func runEntry(args []string) int {
	return runEntryWithMode(args, modeRun)
}

func runCheck(args []string) int {
	return runEntryWithMode(args, modeCheck)
}

func runEntryWithMode(args []string, mode executionMode) int {
	var manifest *driver.Manifest
	var manifestErr error

	if len(args) <= 1 {
		manifest, manifestErr = loadManifestFrom(".")
		if manifestErr != nil {
			switch {
			case errors.Is(manifestErr, errManifestNotFound):
				// No manifest nearby; fall back to file-based invocation if possible.
				manifest = nil
			case len(args) == 1 && looksLikePathCandidate(args[0]):
				fmt.Fprintf(os.Stderr, "warning: unable to load manifest (%v); falling back to direct file execution\n", manifestErr)
				manifest = nil
			default:
				fmt.Fprintf(os.Stderr, "failed to load manifest: %v\n", manifestErr)
				return 1
			}
		}
	}

	if len(args) == 0 {
		if manifest == nil {
			fmt.Fprintf(os.Stderr, "%s requires a manifest target or source file (package.yml not found)\n", modeCommandLabel(mode))
			return 1
		}
		lock, err := loadLockfileForManifest(manifest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		target, err := manifest.DefaultTarget()
		if err != nil {
			fmt.Fprintf(os.Stderr, "manifest error: %v\n", err)
			return 1
		}
		entryPath, err := resolveTargetMain(manifest, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to resolve target entrypoint: %v\n", err)
			return 1
		}
		return executeEntry(entryPath, manifest, lock, mode)
	}

	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %s\n", strings.Join(args[1:], " "))
		return 1
	}

	candidate := args[0]
	activeManifest := manifest
	if manifest != nil {
		if target, ok := manifest.FindTarget(candidate); ok && target != nil {
			entryPath, err := resolveTargetMain(manifest, target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to resolve target %q: %v\n", target.OriginalName, err)
				return 1
			}
			lock, err := loadLockfileForManifest(manifest)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return 1
			}
			return executeEntry(entryPath, manifest, lock, mode)
		}
	}

	if absCandidate, err := filepath.Abs(candidate); err == nil {
		entryDir := filepath.Dir(absCandidate)
		if manifestPath, findErr := findManifest(entryDir); findErr == nil {
			if activeManifest == nil || filepath.Clean(activeManifest.Path) != filepath.Clean(manifestPath) {
				m, loadErr := driver.LoadManifest(manifestPath)
				if loadErr != nil {
					fmt.Fprintf(os.Stderr, "failed to read manifest for %s: %v\n", candidate, loadErr)
					return 1
				}
				activeManifest = m
			}
		} else if !errors.Is(findErr, errManifestNotFound) {
			fmt.Fprintf(os.Stderr, "failed to locate manifest for %s: %v\n", candidate, findErr)
			return 1
		}
	}

	// Treat the argument as a direct source file path.
	lock, err := loadLockfileForManifest(activeManifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	return executeEntry(candidate, activeManifest, lock, mode)
}

func executeEntry(entry string, manifest *driver.Manifest, lock *driver.Lockfile, mode executionMode) int {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		fmt.Fprintf(os.Stderr, "%s requires a source file\n", modeCommandLabel(mode))
		return 1
	}

	entryAbs, err := filepath.Abs(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve entry path: %v\n", err)
		return 1
	}

	extras, err := buildExecutionSearchPaths(manifest, lock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to prepare execution environment: %v\n", err)
		return 1
	}
	searchPaths := collectSearchPaths(filepath.Dir(entryAbs), extras...)

	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize loader: %v\n", err)
		return 1
	}
	defer loader.Close()

	program, err := loader.Load(entryAbs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load program: %v\n", err)
		return 1
	}

	if mode == modeCheck {
		result, err := interpreter.TypecheckProgram(program)
		if err != nil {
			fmt.Fprintf(os.Stderr, "typecheck error: %v\n", err)
			return 1
		}
		if reportTypecheckDiagnostics(result) {
			return 1
		}
		fmt.Fprintln(os.Stdout, "typecheck: ok")
		return 0
	}

	interp := interpreter.New()
	registerPrint(interp)

	_, entryEnv, check, err := interp.EvaluateProgram(program, interpreter.ProgramEvaluationOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	if reportTypecheckDiagnostics(check) {
		return 1
	}

	mainValue, err := entryEnv.Get("main")
	if err != nil {
		fmt.Fprintln(os.Stderr, "entry module does not define a main function")
		return 1
	}

	if _, err := interp.CallFunction(mainValue, nil); err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
		return 1
	}
	return 0
}

func printPackageSummaries(w io.Writer, summaries map[string]interpreter.PackageSummary) {
	if len(summaries) == 0 {
		return
	}
	keys := make([]string, 0, len(summaries))
	for name := range summaries {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	fmt.Fprintln(w, "---- package export summary ----")
	for _, name := range keys {
		summary := summaries[name]
		label := name
		if summary.Visibility == "private" {
			label = fmt.Sprintf("%s (private)", name)
		}
		structs := formatSummaryList(summary.Structs)
		interfaces := formatSummaryList(summary.Interfaces)
		functions := formatSummaryList(summary.Functions)
		fmt.Fprintf(
			w,
			"package %s exports: structs=%s; interfaces=%s; functions=%s; impls=%d; method sets=%d\n",
			label,
			structs,
			interfaces,
			functions,
			len(summary.Implementations),
			len(summary.MethodSets),
		)
	}
}

func formatSummaryList[T any](items map[string]T) string {
	if len(items) == 0 {
		return "-"
	}
	names := make([]string, 0, len(items))
	for name := range items {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func reportTypecheckDiagnostics(result interpreter.ProgramCheckResult) bool {
	if len(result.Diagnostics) == 0 {
		return false
	}
	for _, diag := range result.Diagnostics {
		fmt.Fprintln(os.Stderr, interpreter.DescribeModuleDiagnostic(diag))
	}
	printPackageSummaries(os.Stderr, result.Packages)
	return true
}

func runDeps(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "able deps requires a subcommand (install, update)")
		return 1
	}
	switch args[0] {
	case "install":
		if len(args) > 1 {
			fmt.Fprintf(os.Stderr, "able deps install does not take arguments (received %s)\n", strings.Join(args[1:], " "))
			return 1
		}
		return runDepsInstall()
	case "update":
		return runDepsUpdate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown deps subcommand %q\n", args[0])
		return 1
	}
}

func collectSearchPaths(base string, extra ...string) []driver.SearchPath {
	seen := make(map[string]struct{})
	var paths []driver.SearchPath

	add := func(path string, kind driver.RootKind) {
		if path == "" {
			return
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return
		}
		if kind != driver.RootStdlib && looksLikeStdlibPathCLI(abs) {
			kind = driver.RootStdlib
		}
		if kind != driver.RootStdlib && looksLikeKernelPathCLI(abs) {
			kind = driver.RootStdlib
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			return
		}
		if _, ok := seen[abs]; ok {
			return
		}
		seen[abs] = struct{}{}
		paths = append(paths, driver.SearchPath{Path: abs, Kind: kind})
	}

	for _, path := range extra {
		add(path, driver.RootUser)
	}

	if base != "" {
		add(base, driver.RootUser)
	}

	if cwd, err := os.Getwd(); err == nil {
		add(cwd, driver.RootUser)
	}

	for _, part := range splitPathListEnv(os.Getenv("ABLE_PATH")) {
		add(part, driver.RootUser)
	}

	for _, part := range splitPathListEnv(os.Getenv("ABLE_MODULE_PATHS")) {
		add(part, driver.RootUser)
	}

	for _, path := range collectKernelPaths(base) {
		add(path, driver.RootStdlib)
	}

	for _, path := range collectStdlibPaths(base) {
		add(path, driver.RootStdlib)
	}

	return paths
}

func splitPathListEnv(value string) []string {
	if value == "" {
		return nil
	}
	raw := strings.Split(value, string(os.PathListSeparator))
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func modeCommandLabel(mode executionMode) string {
	switch mode {
	case modeCheck:
		return "able check"
	default:
		return "able run"
	}
}

func looksLikeStdlibPathCLI(path string) bool {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))
	for _, part := range parts {
		switch strings.ToLower(part) {
		case "stdlib", "stdlib_v11", "stdlib_v10", "able-stdlib", "able_stdlib":
			return true
		}
	}
	return false
}

func looksLikeKernelPathCLI(path string) bool {
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

func loadManifestFrom(start string) (*driver.Manifest, error) {
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("resolve working directory: %w", err)
		}
		start = cwd
	}
	absStart, err := filepath.Abs(start)
	if err != nil {
		return nil, fmt.Errorf("resolve manifest search path %q: %w", start, err)
	}
	if info, statErr := os.Stat(absStart); statErr == nil && !info.IsDir() {
		absStart = filepath.Dir(absStart)
	}
	manifestPath, err := findManifest(absStart)
	if err != nil {
		return nil, err
	}
	manifest, err := driver.LoadManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func resolveTargetMain(manifest *driver.Manifest, target *driver.TargetSpec) (string, error) {
	if manifest == nil || target == nil {
		return "", fmt.Errorf("missing manifest or target")
	}
	mainPath := strings.TrimSpace(target.Main)
	if mainPath == "" {
		return "", fmt.Errorf("target %q missing main entrypoint", target.OriginalName)
	}
	if filepath.IsAbs(mainPath) {
		return filepath.Clean(mainPath), nil
	}
	base := filepath.Dir(manifest.Path)
	if base == "" {
		return filepath.Clean(filepath.FromSlash(mainPath)), nil
	}
	return filepath.Join(base, filepath.FromSlash(mainPath)), nil
}

func looksLikePathCandidate(arg string) bool {
	if arg == "" {
		return false
	}
	if strings.Contains(arg, string(os.PathSeparator)) {
		return true
	}
	// Support forward/backward slashes regardless of host OS.
	if strings.Contains(arg, "/") || strings.Contains(arg, "\\") {
		return true
	}
	if filepath.Ext(arg) == ".able" {
		return true
	}
	if strings.HasPrefix(arg, ".") {
		return true
	}
	return false
}

func runDepsInstall() int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to determine working directory: %v\n", err)
		return 1
	}
	manifestPath, err := findManifest(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to locate package.yml: %v\n", err)
		return 1
	}
	manifest, err := driver.LoadManifest(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read manifest: %v\n", err)
		return 1
	}
	cacheDir, err := resolveAbleHome()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve ABLE_HOME: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stdout, "Manifest: %s\n", manifest.Path)
	fmt.Fprintf(os.Stdout, "Root package: %s\n", manifest.Name)
	fmt.Fprintf(os.Stdout, "Dependencies: %d\n", len(manifest.Dependencies))
	fmt.Fprintf(os.Stdout, "Cache directory: %s\n", cacheDir)

	lockPath := filepath.Join(filepath.Dir(manifest.Path), "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	lockCreated := false
	switch {
	case err == nil:
		if lock.Root != manifest.Name {
			fmt.Fprintf(os.Stderr, "lockfile root %q does not match manifest name %q\n", lock.Root, manifest.Name)
			return 1
		}
	case errors.Is(err, os.ErrNotExist):
		lock = driver.NewLockfile(manifest.Name, cliToolVersion)
		lock.Path = lockPath
		lockCreated = true
	default:
		fmt.Fprintf(os.Stderr, "failed to read lockfile: %v\n", err)
		return 1
	}

	lock.Path = lockPath
	lock.Tool = cliToolVersion

	installer := newDependencyInstaller(manifest, cacheDir)
	changed, logs, err := installer.Install(lock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve dependencies: %v\n", err)
		return 1
	}
	for _, line := range logs {
		fmt.Fprintln(os.Stdout, line)
	}

	if changed || lockCreated {
		action := "Updated"
		if lockCreated {
			action = "Created"
		}
		if err := driver.WriteLockfile(lock, lockPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write lockfile: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stdout, "%s package.lock: %s\n", action, lock.Path)
	} else {
		fmt.Fprintf(os.Stdout, "package.lock already up to date: %s\n", lock.Path)
	}

	fmt.Fprintln(os.Stdout, "Dependencies installed.")
	return 0
}

func runDepsUpdate(targets []string) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to determine working directory: %v\n", err)
		return 1
	}
	manifestPath, err := findManifest(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to locate package.yml: %v\n", err)
		return 1
	}
	manifest, err := driver.LoadManifest(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read manifest: %v\n", err)
		return 1
	}
	cacheDir, err := resolveAbleHome()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve ABLE_HOME: %v\n", err)
		return 1
	}

	updateSet := make(map[string]struct{})
	if len(targets) > 0 {
		manifestDeps := make(map[string]struct{}, len(manifest.Dependencies))
		for name := range manifest.Dependencies {
			manifestDeps[sanitizeName(name)] = struct{}{}
		}
		for _, target := range targets {
			sanitized := sanitizeName(target)
			if _, ok := manifestDeps[sanitized]; !ok {
				fmt.Fprintf(os.Stderr, "dependency %q not declared in manifest\n", target)
				return 1
			}
			updateSet[sanitized] = struct{}{}
		}
	}

	lockPath := filepath.Join(filepath.Dir(manifest.Path), "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	lockCreated := false
	switch {
	case err == nil:
		if lock.Root != manifest.Name {
			fmt.Fprintf(os.Stderr, "lockfile root %q does not match manifest name %q\n", lock.Root, manifest.Name)
			return 1
		}
	case errors.Is(err, os.ErrNotExist):
		lock = driver.NewLockfile(manifest.Name, cliToolVersion)
		lock.Path = lockPath
		lockCreated = true
	default:
		fmt.Fprintf(os.Stderr, "failed to read lockfile: %v\n", err)
		return 1
	}

	if len(updateSet) == 0 {
		lock.Packages = nil
	} else {
		filtered := make([]*driver.LockedPackage, 0, len(lock.Packages))
		for _, pkg := range lock.Packages {
			if pkg == nil {
				continue
			}
			if _, ok := updateSet[sanitizeName(pkg.Name)]; ok {
				continue
			}
			filtered = append(filtered, pkg)
		}
		lock.Packages = filtered
	}

	installer := newDependencyInstaller(manifest, cacheDir)
	changed, logs, err := installer.Install(lock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to update dependencies: %v\n", err)
		return 1
	}
	for _, line := range logs {
		fmt.Fprintln(os.Stdout, line)
	}

	lock.Path = lockPath
	lock.Tool = cliToolVersion

	if changed || lockCreated {
		if err := driver.WriteLockfile(lock, lockPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write lockfile: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stdout, "Updated package.lock: %s\n", lock.Path)
	} else {
		fmt.Fprintln(os.Stdout, "Dependencies already up to date.")
	}
	return 0
}

func findManifest(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve start directory %q: %w", start, err)
	}
	if info, statErr := os.Stat(dir); statErr == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	origin := dir
	for {
		candidate := filepath.Join(dir, "package.yml")
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no package.yml found from %s upwards: %w", origin, errManifestNotFound)
		}
		dir = parent
	}
}

func resolveAbleHome() (string, error) {
	if home := strings.TrimSpace(os.Getenv("ABLE_HOME")); home != "" {
		if abs, err := filepath.Abs(home); err == nil {
			return abs, nil
		} else {
			return "", fmt.Errorf("resolve ABLE_HOME %q: %w", home, err)
		}
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(userHome, ".able"), nil
}

func collectKernelPaths(base string) []string {
	var paths []string

	add := func(candidate string) {
		if candidate == "" {
			return
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return
		}
		for _, existing := range paths {
			if existing == abs {
				return
			}
		}
		paths = append(paths, abs)
	}

	// Discover bundled kernel relative to the entry directory first.
	if base != "" {
		for _, found := range findKernelRoots(base) {
			add(found)
		}
	}

	for _, entry := range splitPathListEnv(os.Getenv("ABLE_MODULE_PATHS")) {
		add(entry)
	}
	for _, entry := range splitPathListEnv(os.Getenv("ABLE_PATH")) {
		add(entry)
	}

	// Discover bundled kernel relative to the working directory.
	if cwd, err := os.Getwd(); err == nil {
		for _, found := range findKernelRoots(cwd) {
			add(found)
		}
	}

	// Also probe relative to the executable for installed builds.
	if exe, err := os.Executable(); err == nil {
		for _, found := range findKernelRoots(filepath.Dir(exe)) {
			add(found)
		}
	}

	return paths
}

func collectStdlibPaths(base string) []string {
	var paths []string

	add := func(candidate string) {
		if candidate == "" {
			return
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return
		}
		for _, existing := range paths {
			if existing == abs {
				return
			}
		}
		paths = append(paths, abs)
	}

	for _, entry := range splitPathListEnv(os.Getenv("ABLE_MODULE_PATHS")) {
		add(entry)
	}
	for _, entry := range splitPathListEnv(os.Getenv("ABLE_PATH")) {
		add(entry)
	}

	// Discover bundled stdlib relative to the entry directory first.
	if base != "" {
		for _, found := range findStdlibRoots(base) {
			add(found)
		}
	}

	// Discover bundled stdlib relative to the working directory.
	if cwd, err := os.Getwd(); err == nil {
		for _, found := range findStdlibRoots(cwd) {
			add(found)
		}
	}

	// Also probe relative to the executable for installed builds.
	if exe, err := os.Executable(); err == nil {
		for _, found := range findStdlibRoots(filepath.Dir(exe)) {
			add(found)
		}
	}

	return paths
}

func findKernelRoots(start string) []string {
	var roots []string
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			roots = append(roots, candidate)
		}
	}

	dir := start
	for {
		for _, candidate := range []string{
			filepath.Join(dir, "kernel", "src"),
			filepath.Join(dir, "v11", "kernel", "src"),
			filepath.Join(dir, "ablekernel", "src"),
			filepath.Join(dir, "able_kernel", "src"),
		} {
			add(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return roots
}

func findStdlibRoots(start string) []string {
	var roots []string
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			roots = append(roots, candidate)
		}
	}

	dir := start
	for {
		for _, candidate := range []string{
			filepath.Join(dir, "stdlib", "src"),
			filepath.Join(dir, "v11", "stdlib", "src"),
			filepath.Join(dir, "stdlib", "v11", "src"),
			filepath.Join(dir, "stdlib", "v10", "src"),
			filepath.Join(dir, "able-stdlib", "src"),
			filepath.Join(dir, "able_stdlib", "src"),
		} {
			add(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return roots
}

func registerPrint(interp *interpreter.Interpreter) {
	printFn := runtime.NativeFunctionValue{
		Name:  "print",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			var parts []string
			for _, arg := range args {
				parts = append(parts, formatRuntimeValue(arg))
			}
			fmt.Fprintln(os.Stdout, strings.Join(parts, " "))
			return runtime.NilValue{}, nil
		},
	}
	interp.GlobalEnvironment().Define("print", printFn)
}

func formatRuntimeValue(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val
	case runtime.BoolValue:
		if v.Val {
			return "true"
		}
		return "false"
	case runtime.IntegerValue:
		return v.Val.String()
	case runtime.FloatValue:
		return fmt.Sprintf("%g", v.Val)
	case runtime.CharValue:
		return string(v.Val)
	case runtime.NilValue:
		return "nil"
	default:
		return fmt.Sprintf("[%s]", v.Kind())
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  able run [target]")
	fmt.Fprintln(os.Stderr, "  able run <file.able>")
	fmt.Fprintln(os.Stderr, "  able <file.able>")
	fmt.Fprintln(os.Stderr, "  able check [target]")
	fmt.Fprintln(os.Stderr, "  able check <file.able>")
	fmt.Fprintln(os.Stderr, "  able deps install")
	fmt.Fprintln(os.Stderr, "  able deps update [dependency ...]")
}

func loadLockfileForManifest(manifest *driver.Manifest) (*driver.Lockfile, error) {
	if manifest == nil {
		return nil, nil
	}
	lockPath := filepath.Join(filepath.Dir(manifest.Path), "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if manifestHasDependencies(manifest) {
				return nil, fmt.Errorf("package.lock missing for %q; run `able deps install`", manifest.Name)
			}
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read lockfile %s: %w", lockPath, err)
	}
	if lock.Root != manifest.Name {
		return nil, fmt.Errorf("lockfile root %q does not match manifest name %q", lock.Root, manifest.Name)
	}
	return lock, nil
}

func manifestHasDependencies(manifest *driver.Manifest) bool {
	if manifest == nil {
		return false
	}
	return len(manifest.Dependencies) > 0 ||
		len(manifest.DevDependencies) > 0 ||
		len(manifest.BuildDependencies) > 0
}
