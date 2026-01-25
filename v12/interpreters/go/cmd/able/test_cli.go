package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"able/interpreter-go/pkg/interpreter"
)

func runTest(args []string) int {
	config, err := parseTestArguments(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 1
	}

	targets, err := resolveTestTargets(config.Targets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 1
	}

	testFiles, err := collectTestFiles(targets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 1
	}

	if len(testFiles) == 0 {
		fmt.Fprintln(os.Stdout, "able test: no test modules found")
		return 0
	}

	loadResult, err := loadTestPrograms(testFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}

	mode := resolveTestTypecheckMode()
	if ok, code := typecheckTestModules(loadResult.modules, mode); !ok {
		return code
	}

	interp := interpreter.New()
	registerPrint(interp)

	if ok, code := evaluateTestModules(interp, loadResult.modules); !ok {
		return code
	}

	cliModule, err := loadTestCliModule(interp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}

	discoveryRequest, err := buildDiscoveryRequest(interp, cliModule, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}

	descriptors, ok := callHarnessDiscover(interp, cliModule, discoveryRequest)
	if !ok {
		return 2
	}

	if config.ListOnly || config.DryRun {
		emitTestPlanList(interp, descriptors, config)
		return 0
	}

	if arrayLength(interp, descriptors) == 0 {
		fmt.Fprintln(os.Stdout, "able test: no tests to run")
		return 0
	}

	state := &TestEventState{}
	reporter, err := createTestReporter(interp, cliModule, config.ReporterFormat, state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}

	runOptions, err := buildRunOptions(interp, cliModule, config.Run)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}
	testPlan, err := buildTestPlan(cliModule, descriptors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}

	if callHarnessRun(interp, cliModule, testPlan, runOptions, reporter.reporter) != nil {
		return 2
	}

	if reporter.finish != nil {
		reporter.finish()
	}

	if state.FrameworkErrors > 0 {
		return 2
	}
	if state.Failed > 0 {
		return 1
	}
	return 0
}

func parseTestArguments(args []string) (TestCliConfig, error) {
	filters := TestCliFilters{
		IncludePaths: []string{},
		ExcludePaths: []string{},
		IncludeNames: []string{},
		ExcludeNames: []string{},
		IncludeTags:  []string{},
		ExcludeTags:  []string{},
	}
	run := TestRunOptions{
		FailFast:    false,
		Repeat:      1,
		Parallelism: 1,
	}
	format := reporterDoc
	listOnly := false
	dryRun := false
	var shuffleSeed *int64
	var targets []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--list":
			listOnly = true
		case "--dry-run":
			dryRun = true
			listOnly = true
		case "--path":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.IncludePaths = append(filters.IncludePaths, val)
		case "--exclude-path":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.ExcludePaths = append(filters.ExcludePaths, val)
		case "--name":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.IncludeNames = append(filters.IncludeNames, val)
		case "--exclude-name":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.ExcludeNames = append(filters.ExcludeNames, val)
		case "--tag":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.IncludeTags = append(filters.IncludeTags, val)
		case "--exclude-tag":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.ExcludeTags = append(filters.ExcludeTags, val)
		case "--format":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			parsed, err := parseReporterFormat(val)
			if err != nil {
				return TestCliConfig{}, err
			}
			format = parsed
		case "--fail-fast":
			run.FailFast = true
		case "--repeat":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			count, err := parsePositiveInt(val, arg, 1)
			if err != nil {
				return TestCliConfig{}, err
			}
			run.Repeat = count
		case "--parallel":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			count, err := parsePositiveInt(val, arg, 1)
			if err != nil {
				return TestCliConfig{}, err
			}
			run.Parallelism = count
		case "--shuffle":
			next := peekArg(args, i+1)
			if next != "" && !strings.HasPrefix(next, "-") {
				seed, err := parsePositiveInt(next, arg, 0)
				if err != nil {
					return TestCliConfig{}, err
				}
				seedVal := int64(seed)
				shuffleSeed = &seedVal
				i++
			} else {
				seed := generateShuffleSeed()
				shuffleSeed = &seed
			}
		default:
			if strings.HasPrefix(arg, "-") {
				return TestCliConfig{}, fmt.Errorf("unknown able test flag '%s'", arg)
			}
			targets = append(targets, arg)
		}
	}

	run.ShuffleSeed = shuffleSeed

	return TestCliConfig{
		Targets:        targets,
		Filters:        filters,
		Run:            run,
		ReporterFormat: format,
		ListOnly:       listOnly,
		DryRun:         dryRun,
	}, nil
}

func nextArg(args []string, index *int) string {
	*index = *index + 1
	if *index >= len(args) {
		return ""
	}
	return args[*index]
}

func peekArg(args []string, index int) string {
	if index < 0 || index >= len(args) {
		return ""
	}
	return args[index]
}

func expectFlagValue(flag string, value string) (string, error) {
	if value == "" || strings.HasPrefix(value, "-") {
		return "", fmt.Errorf("%s expects a value", flag)
	}
	return value, nil
}

func parseReporterFormat(value string) (TestReporterFormat, error) {
	switch value {
	case "doc":
		return reporterDoc, nil
	case "progress":
		return reporterProgress, nil
	case "tap":
		return reporterTap, nil
	case "json":
		return reporterJSON, nil
	default:
		return "", fmt.Errorf("unknown --format value '%s' (expected doc, progress, tap, or json)", value)
	}
}

func parsePositiveInt(value string, flag string, min int) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < min {
		return 0, fmt.Errorf("%s expects an integer >= %d", flag, min)
	}
	return parsed, nil
}

func generateShuffleSeed() int64 {
	now := time.Now().UnixMilli()
	str := fmt.Sprintf("%d", now)
	if len(str) > 9 {
		str = str[len(str)-9:]
	}
	parsed, _ := strconv.ParseInt(str, 10, 64)
	return parsed
}

func resolveTestTargets(targets []string) ([]string, error) {
	rawTargets := targets
	if len(rawTargets) == 0 {
		rawTargets = []string{"."}
	}
	seen := make(map[string]struct{})
	var resolved []string
	for _, target := range rawTargets {
		abs, err := filepath.Abs(target)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve %s: %w", target, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("unable to access %s: %w", abs, err)
		}
		if info.IsDir() {
			if _, ok := seen[abs]; !ok {
				seen[abs] = struct{}{}
				resolved = append(resolved, abs)
			}
			continue
		}
		if info.Mode().IsRegular() {
			if isTestFile(abs) {
				if _, ok := seen[abs]; !ok {
					seen[abs] = struct{}{}
					resolved = append(resolved, abs)
				}
			} else {
				dir := filepath.Dir(abs)
				if _, ok := seen[dir]; !ok {
					seen[dir] = struct{}{}
					resolved = append(resolved, dir)
				}
			}
			continue
		}
		return nil, fmt.Errorf("unsupported test target: %s", abs)
	}
	return resolved, nil
}

func collectTestFiles(targets []string) ([]string, error) {
	found := make(map[string]struct{})
	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			if err := walkTestFiles(target, found); err != nil {
				return nil, err
			}
			continue
		}
		if isTestFile(target) {
			found[filepath.Clean(target)] = struct{}{}
		}
	}
	files := make([]string, 0, len(found))
	for file := range found {
		files = append(files, file)
	}
	sort.Strings(files)
	return files, nil
}

func walkTestFiles(root string, found map[string]struct{}) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "quarantine", "node_modules", ".git":
				return fs.SkipDir
			default:
				return nil
			}
		}
		if d.Type().IsRegular() && isTestFile(d.Name()) {
			found[filepath.Clean(path)] = struct{}{}
		}
		return nil
	})
}

func isTestFile(path string) bool {
	return strings.HasSuffix(path, ".test.able") || strings.HasSuffix(path, ".spec.able")
}
