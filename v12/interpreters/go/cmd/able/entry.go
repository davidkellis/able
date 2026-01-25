package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

func runEntry(args []string) int {
	return runEntryWithMode(args, modeRun)
}

func runCheck(args []string) int {
	return runEntryWithMode(args, modeCheck)
}

func runRepl(args []string) int {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "able repl does not take arguments (received %s)\n", strings.Join(args, " "))
		return 1
	}
	manifest, err := loadManifestFrom(".")
	if err != nil {
		if !errors.Is(err, errManifestNotFound) {
			fmt.Fprintf(os.Stderr, "failed to load manifest: %v\n", err)
			return 1
		}
		manifest = nil
	}
	lock, err := loadLockfileForManifest(manifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	base := "."
	if manifest != nil && manifest.Path != "" {
		base = filepath.Dir(manifest.Path)
	} else if cwd, cwdErr := os.Getwd(); cwdErr == nil {
		base = cwd
	}
	entryPath, err := resolveReplEntryPath(base)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	return executeEntry(entryPath, manifest, lock, modeRun, nil)
}

func runEntryWithMode(args []string, mode executionMode) int {
	var manifest *driver.Manifest
	var manifestErr error
	programArgs := []string{}

	if len(args) > 1 {
		if mode != modeRun {
			fmt.Fprintf(os.Stderr, "unexpected arguments: %s\n", strings.Join(args[1:], " "))
			return 1
		}
		programArgs = append([]string{}, args[1:]...)
	}

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
		return executeEntry(entryPath, manifest, lock, mode, programArgs)
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
			return executeEntry(entryPath, manifest, lock, mode, programArgs)
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
	return executeEntry(candidate, activeManifest, lock, mode, programArgs)
}

func executeEntry(entry string, manifest *driver.Manifest, lock *driver.Lockfile, mode executionMode, programArgs []string) int {
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
		var parseErr *driver.ParserDiagnosticError
		if errors.As(err, &parseErr) {
			fmt.Fprintln(os.Stderr, driver.DescribeParserDiagnostic(parseErr.Diagnostic))
			return 1
		}
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
	interp.SetArgs(programArgs)
	registerPrint(interp)

	_, entryEnv, check, err := interp.EvaluateProgram(program, interpreter.ProgramEvaluationOptions{})
	if err != nil {
		if code, ok := interpreter.ExitCodeFromError(err); ok {
			return code
		}
		fmt.Fprintln(os.Stderr, interpreter.DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(err)))
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
		if code, ok := interpreter.ExitCodeFromError(err); ok {
			return code
		}
		fmt.Fprintln(os.Stderr, interpreter.DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(err)))
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
