package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strings"

	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
	"able/interpreter-go/pkg/stdlibpath"
)

type parityValue struct {
	Kind  string `json:"kind"`
	Value string `json:"value,omitempty"`
	Bool  *bool  `json:"bool,omitempty"`
}

type parityOutput struct {
	Result        *parityValue `json:"result,omitempty"`
	Stdout        []string     `json:"stdout,omitempty"`
	Error         string       `json:"error,omitempty"`
	Diagnostics   []string     `json:"diagnostics,omitempty"`
	TypecheckMode string       `json:"typecheckMode"`
	Skipped       bool         `json:"skipped"`
}

type fixtureDescriptionOutput struct {
	Description   string   `json:"description,omitempty"`
	FixtureDir    string   `json:"fixtureDir"`
	Entry         string   `json:"entry"`
	Setup         []string `json:"setup,omitempty"`
	Executor      string   `json:"executor"`
	SkipTargets   []string `json:"skipTargets,omitempty"`
	TypecheckMode string   `json:"typecheckMode"`
	Skipped       bool     `json:"skipped"`
}

type fixtureBatchRunOutput struct {
	FixtureDir string `json:"fixtureDir"`
	Entry      string `json:"entry"`
	Executor   string `json:"executor"`
	parityOutput
}

type resolvedFixtureTarget struct {
	Dir           string
	EntryOverride string
}

const (
	outputFormatJSON  = "json"
	outputFormatText  = "text"
	outputFormatJSONL = "jsonl"
)

func main() {
	dirFlag := flag.String("dir", "", "Path to fixture directory (optional if a fixture directory or entry file is passed positionally; relative paths also resolve under v12/fixtures/ast when available)")
	entryFlag := flag.String("entry", "", "Override manifest entry file")
	execFlag := flag.String("executor", "", "Executor override (serial|goroutine); defaults to the fixture manifest executor or serial")
	listFlag := flag.Bool("list", false, "List fixture directories under v12/fixtures/ast (optionally filtered by prefix/path) and exit")
	describeFlag := flag.Bool("describe", false, "Print resolved fixture metadata (dir, entry, setup, skip targets, executor, typecheck mode) and exit")
	batchFlag := flag.Bool("batch", false, "Replay multiple fixture targets and emit a JSON array of per-target results")
	formatFlag := flag.String("format", "", "Output format: describe supports json|text, batch supports json|jsonl")
	flag.Parse()

	if err := validateTopLevelModes(*listFlag, *describeFlag, *batchFlag); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	format, err := resolveOutputFormat(*formatFlag, *listFlag, *describeFlag, *batchFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *listFlag {
		if err := listFixtureTargets(*dirFlag, *entryFlag, *execFlag, flag.Args(), format, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *batchFlag {
		targets, err := resolveFixtureBatchInputs(*dirFlag, *entryFlag, flag.Args())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		outcomes, err := runFixtureBatch(targets, *execFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "evaluation failed: %v\n", err)
			os.Exit(1)
		}
		writeBatchOutput(outcomes, format)
		return
	}

	dir, entryOverride, err := resolveFixtureInput(*dirFlag, *entryFlag, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	outcome, desc, err := evaluateFixtureTarget(dir, entryOverride, *execFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "evaluation failed: %v\n", err)
		os.Exit(1)
	}
	if *describeFlag {
		writeDescribeOutput(desc, format)
		return
	}
	writeJSON(outcome)
}

func validateTopLevelModes(listMode, describeMode, batchMode bool) error {
	if listMode && describeMode {
		return fmt.Errorf("--list cannot be used together with --describe")
	}
	if listMode && batchMode {
		return fmt.Errorf("--list cannot be used together with --batch")
	}
	if describeMode && batchMode {
		return fmt.Errorf("--batch cannot be used together with --describe")
	}
	return nil
}

func resolveOutputFormat(requested string, listMode, describeMode, batchMode bool) (string, error) {
	format := strings.TrimSpace(strings.ToLower(requested))
	switch {
	case listMode:
		switch format {
		case "", outputFormatText:
			return outputFormatText, nil
		case outputFormatJSON:
			return outputFormatJSON, nil
		}
		return "", fmt.Errorf("--list supports --format text or json")
	case describeMode:
		switch format {
		case "", outputFormatJSON:
			return outputFormatJSON, nil
		case outputFormatText:
			return outputFormatText, nil
		default:
			return "", fmt.Errorf("--describe supports --format json or text")
		}
	case batchMode:
		switch format {
		case "", outputFormatJSON:
			return outputFormatJSON, nil
		case outputFormatJSONL:
			return outputFormatJSONL, nil
		default:
			return "", fmt.Errorf("--batch supports --format json or jsonl")
		}
	default:
		switch format {
		case "", outputFormatJSON:
			return outputFormatJSON, nil
		default:
			return "", fmt.Errorf("single fixture replay only supports --format json")
		}
	}
}

func listFixtureTargets(dirFlag, entryFlag, execFlag string, args []string, format string, stdout io.Writer) error {
	return listFixtureTargetsWithRoot(dirFlag, entryFlag, execFlag, args, format, defaultFixturesRoot(), stdout)
}

func listFixtureTargetsWithRoot(dirFlag, entryFlag, execFlag string, args []string, format string, fixturesRoot string, stdout io.Writer) error {
	if dirFlag != "" {
		return fmt.Errorf("--list cannot be used together with --dir")
	}
	if entryFlag != "" {
		return fmt.Errorf("--list cannot be used together with --entry")
	}
	if execFlag != "" {
		return fmt.Errorf("--list cannot be used together with --executor")
	}
	if fixturesRoot == "" {
		return fmt.Errorf("unable to locate v12/fixtures/ast")
	}

	filters, err := normalizeFixtureListFilters(args, fixturesRoot)
	if err != nil {
		return err
	}
	fixtures, err := collectFixtureDirs(fixturesRoot)
	if err != nil {
		return err
	}
	matches := filterFixtureDirs(fixtures, filters)
	if len(matches) == 0 {
		if len(filters) == 0 {
			return fmt.Errorf("no fixture directories found under %s", fixturesRoot)
		}
		return fmt.Errorf("no fixture directories matched %s", strings.Join(filters, ", "))
	}
	return writeListOutput(stdout, matches, format)
}

func resolveFixtureInput(dirFlag, entryFlag string, args []string) (string, string, error) {
	return resolveFixtureInputWithFixturesRoot(dirFlag, entryFlag, args, defaultFixturesRoot())
}

func resolveFixtureBatchInputs(dirFlag, entryFlag string, args []string) ([]resolvedFixtureTarget, error) {
	return resolveFixtureBatchInputsWithFixturesRoot(dirFlag, entryFlag, args, defaultFixturesRoot())
}

func resolveFixtureInputWithFixturesRoot(dirFlag, entryFlag string, args []string, fixturesRoot string) (string, string, error) {
	if dirFlag != "" {
		if len(args) > 0 {
			return "", "", fmt.Errorf("fixture target cannot be used together with --dir")
		}
		dir, err := resolveFixtureTarget(dirFlag, fixturesRoot)
		if err != nil {
			return "", "", err
		}
		return dir, entryFlag, nil
	}
	if len(args) == 0 {
		return "", "", fmt.Errorf("--dir or a fixture directory / entry file argument is required")
	}
	if len(args) > 1 {
		return "", "", fmt.Errorf("expected at most one fixture target argument")
	}

	target := args[0]
	resolved, err := resolveFixtureTarget(target, fixturesRoot)
	if err != nil {
		return "", "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", "", fmt.Errorf("read fixture target %s: %w", target, err)
	}
	if info.IsDir() {
		return resolved, entryFlag, nil
	}
	if entryFlag != "" {
		return "", "", fmt.Errorf("--entry cannot be used when the fixture target is already a file")
	}
	return filepath.Dir(resolved), filepath.Base(resolved), nil
}

func resolveFixtureBatchInputsWithFixturesRoot(dirFlag, entryFlag string, args []string, fixturesRoot string) ([]resolvedFixtureTarget, error) {
	if dirFlag != "" {
		return nil, fmt.Errorf("--batch cannot be used together with --dir")
	}
	if entryFlag != "" {
		return nil, fmt.Errorf("--batch cannot be used together with --entry")
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("--batch requires at least one fixture target")
	}
	targets := make([]resolvedFixtureTarget, 0, len(args))
	for _, arg := range args {
		dir, entry, err := resolveFixtureInputWithFixturesRoot("", "", []string{arg}, fixturesRoot)
		if err != nil {
			return nil, err
		}
		targets = append(targets, resolvedFixtureTarget{Dir: dir, EntryOverride: entry})
	}
	return targets, nil
}

func resolveFixtureTarget(target string, fixturesRoot string) (string, error) {
	candidates := make([]string, 0, 2)
	if filepath.IsAbs(target) {
		candidates = append(candidates, target)
	} else {
		candidates = append(candidates, target)
		if fixturesRoot != "" {
			candidates = append(candidates, filepath.Join(fixturesRoot, target))
		}
	}
	for _, candidate := range candidates {
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return "", fmt.Errorf("resolve fixture target %s: %w", target, err)
		}
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
	}
	if fixturesRoot != "" {
		return "", fmt.Errorf("fixture target %s not found (checked current path and %s)", target, fixturesRoot)
	}
	return "", fmt.Errorf("fixture target %s not found", target)
}

func normalizeFixtureListFilters(args []string, fixturesRoot string) ([]string, error) {
	if len(args) == 0 {
		return nil, nil
	}
	filters := make([]string, 0, len(args))
	for _, arg := range args {
		filter, err := normalizeFixtureListFilter(arg, fixturesRoot)
		if err != nil {
			return nil, err
		}
		filters = append(filters, filter)
	}
	return filters, nil
}

func normalizeFixtureListFilter(arg string, fixturesRoot string) (string, error) {
	candidates := []string{arg}
	if !filepath.IsAbs(arg) {
		candidates = append(candidates, filepath.Join(fixturesRoot, arg))
	}
	for _, candidate := range candidates {
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return "", fmt.Errorf("resolve fixture filter %s: %w", arg, err)
		}
		if !isWithinRoot(abs, fixturesRoot) {
			continue
		}
		rel, err := filepath.Rel(fixturesRoot, abs)
		if err != nil {
			return "", fmt.Errorf("resolve fixture filter %s: %w", arg, err)
		}
		return filepath.ToSlash(filepath.Clean(rel)), nil
	}
	return filepath.ToSlash(filepath.Clean(arg)), nil
}

func collectFixtureDirs(fixturesRoot string) ([]string, error) {
	dirs := make(map[string]struct{})
	err := filepath.WalkDir(fixturesRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		switch entry.Name() {
		case "source.able", "module.json", "manifest.json":
			dir := filepath.Dir(path)
			rel, err := filepath.Rel(fixturesRoot, dir)
			if err != nil {
				return err
			}
			dirs[filepath.ToSlash(rel)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list fixtures under %s: %w", fixturesRoot, err)
	}
	out := make([]string, 0, len(dirs))
	for dir := range dirs {
		out = append(out, dir)
	}
	sort.Strings(out)
	return out, nil
}

func filterFixtureDirs(fixtures []string, filters []string) []string {
	if len(filters) == 0 {
		return fixtures
	}
	out := make([]string, 0, len(fixtures))
	for _, fixture := range fixtures {
		for _, filter := range filters {
			if fixture == filter || strings.HasPrefix(fixture, filter+"/") {
				out = append(out, fixture)
				break
			}
		}
	}
	return out
}

func isWithinRoot(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func defaultFixturesRoot() string {
	root := repositoryRoot()
	if root == "" {
		return ""
	}
	return filepath.Join(root, "v12", "fixtures", "ast")
}

func repositoryRoot() string {
	start := ""
	if _, file, _, ok := goruntime.Caller(0); ok {
		start = filepath.Dir(file)
	} else if wd, err := os.Getwd(); err == nil {
		start = wd
	}
	return repositoryRootFrom(start)
}

func repositoryRootFrom(start string) string {
	dir := start
	for i := 0; i < 12 && dir != "" && dir != string(filepath.Separator); i++ {
		if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if start == "" {
		return ""
	}
	if strings.HasSuffix(start, filepath.Join("v12", "interpreters", "go", "cmd", "fixture")) {
		return filepath.Clean(filepath.Join(start, "..", "..", "..", "..", ".."))
	}
	return ""
}

func selectExecutor(name string) (interpreter.Executor, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "serial":
		return interpreter.NewSerialExecutor(nil), nil
	case "goroutine":
		return interpreter.NewGoroutineExecutor(nil), nil
	default:
		return nil, fmt.Errorf("unknown executor %q", name)
	}
}

func resolveFixtureEntry(manifestEntry, entryOverride string) string {
	if entryOverride != "" {
		return entryOverride
	}
	if manifestEntry != "" {
		return manifestEntry
	}
	return "module.json"
}

func resolveFixtureExecutor(manifestExecutor, execOverride string) (string, interpreter.Executor, error) {
	name := strings.TrimSpace(execOverride)
	if name == "" {
		name = strings.TrimSpace(manifestExecutor)
	}
	if name == "" {
		name = "serial"
	}
	executor, err := selectExecutor(name)
	if err != nil {
		return "", nil, err
	}
	return name, executor, nil
}

func buildFixtureDescription(dir string, manifest interpreter.FixtureManifest, entry string, executorName string) fixtureDescriptionOutput {
	return fixtureDescriptionOutput{
		Description:   manifest.Description,
		FixtureDir:    dir,
		Entry:         entry,
		Setup:         append([]string(nil), manifest.Setup...),
		Executor:      executorName,
		SkipTargets:   append([]string(nil), manifest.SkipTargets...),
		TypecheckMode: fixtureTypecheckModeStringFromEnv(),
		Skipped:       shouldSkipTarget(manifest.SkipTargets, "go"),
	}
}

func fixtureTypecheckModeStringFromEnv() string {
	modeVal, ok := os.LookupEnv("ABLE_TYPECHECK_FIXTURES")
	if !ok {
		return "strict"
	}
	switch mode := strings.TrimSpace(strings.ToLower(modeVal)); mode {
	case "", "0", "off", "false":
		return "off"
	case "strict", "fail", "error", "1", "true":
		return "strict"
	case "warn", "warning":
		return "warn"
	default:
		return "warn"
	}
}

func shouldSkipTarget(skipTargets []string, target string) bool {
	for _, entry := range skipTargets {
		if strings.EqualFold(strings.TrimSpace(entry), target) {
			return true
		}
	}
	return false
}

func evaluateFixtureTarget(dir, entryOverride, execOverride string) (parityOutput, fixtureDescriptionOutput, error) {
	var output parityOutput
	var desc fixtureDescriptionOutput

	manifest, err := interpreter.LoadFixtureManifest(dir)
	if err != nil {
		return output, desc, fmt.Errorf("read manifest: %w", err)
	}
	entry := resolveFixtureEntry(manifest.Entry, entryOverride)
	executorName, executor, err := resolveFixtureExecutor(manifest.Executor, execOverride)
	if err != nil {
		return output, desc, err
	}
	desc = buildFixtureDescription(dir, manifest, entry, executorName)
	if desc.Skipped {
		return parityOutput{
			TypecheckMode: desc.TypecheckMode,
			Skipped:       true,
		}, desc, nil
	}
	output, err = runFixture(dir, entry, manifest.Setup, executor)
	if err != nil {
		return output, desc, err
	}
	return output, desc, nil
}

func runFixtureBatch(targets []resolvedFixtureTarget, execOverride string) ([]fixtureBatchRunOutput, error) {
	out := make([]fixtureBatchRunOutput, 0, len(targets))
	for _, target := range targets {
		result, desc, err := evaluateFixtureTarget(target.Dir, target.EntryOverride, execOverride)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", target.Dir, err)
		}
		out = append(out, fixtureBatchRunOutput{
			FixtureDir:   desc.FixtureDir,
			Entry:        desc.Entry,
			Executor:     desc.Executor,
			parityOutput: result,
		})
	}
	return out, nil
}

func runFixture(dir, entry string, setup []string, executor interpreter.Executor) (parityOutput, error) {
	var output parityOutput
	replayed, err := interpreter.ReplayFixture(dir, entry, setup, executor)
	if err != nil {
		return output, err
	}
	output.TypecheckMode = replayed.TypecheckMode
	output.Diagnostics = replayed.Diagnostics
	output.Stdout = replayed.Stdout
	if replayed.RuntimeError != "" {
		output.Error = replayed.RuntimeError
		return output, nil
	}
	if replayed.Value != nil {
		normalized := normalizeValue(replayed.Value)
		output.Result = &normalized
	}
	return output, nil
}

func findFixtureStdlibRoot(repoRoot string) string {
	return stdlibpath.ResolveRepoOrInstalledSrc(repoRoot)
}

func normalizeValue(val runtime.Value) parityValue {
	switch v := val.(type) {
	case runtime.StringValue:
		return parityValue{Kind: "String", Value: v.Val}
	case runtime.BoolValue:
		return parityValue{Kind: "bool", Bool: &v.Val}
	case runtime.CharValue:
		return parityValue{Kind: "char", Value: string(v.Val)}
	case runtime.IntegerValue:
		return parityValue{Kind: string(v.TypeSuffix), Value: v.Val.String()}
	case runtime.FloatValue:
		return parityValue{Kind: string(v.TypeSuffix), Value: fmt.Sprintf("%g", v.Val)}
	case runtime.NilValue:
		return parityValue{Kind: "nil"}
	default:
		return parityValue{Kind: v.Kind().String()}
	}
}

func writeJSON(out any) {
	if err := writeJSONTo(os.Stdout, out); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}

func writeJSONTo(out io.Writer, value any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "")
	return enc.Encode(value)
}

func writeDescribeOutput(desc fixtureDescriptionOutput, format string) {
	var err error
	switch format {
	case outputFormatText:
		err = writeDescribeText(os.Stdout, desc)
	default:
		err = writeJSONTo(os.Stdout, desc)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write describe output: %v\n", err)
		os.Exit(1)
	}
}

func writeListOutput(out io.Writer, fixtures []string, format string) error {
	switch format {
	case outputFormatJSON:
		return writeJSONTo(out, fixtures)
	default:
		for _, fixture := range fixtures {
			if _, err := fmt.Fprintln(out, fixture); err != nil {
				return err
			}
		}
		return nil
	}
}

func writeBatchOutput(outcomes []fixtureBatchRunOutput, format string) {
	var err error
	switch format {
	case outputFormatJSONL:
		err = writeBatchJSONL(os.Stdout, outcomes)
	default:
		err = writeJSONTo(os.Stdout, outcomes)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write batch output: %v\n", err)
		os.Exit(1)
	}
}

func writeDescribeText(out io.Writer, desc fixtureDescriptionOutput) error {
	if desc.Description != "" {
		if _, err := fmt.Fprintf(out, "description: %s\n", desc.Description); err != nil {
			return err
		}
	}
	lines := []string{
		fmt.Sprintf("fixtureDir: %s", desc.FixtureDir),
		fmt.Sprintf("entry: %s", desc.Entry),
		fmt.Sprintf("executor: %s", desc.Executor),
		fmt.Sprintf("typecheckMode: %s", desc.TypecheckMode),
		fmt.Sprintf("skipped: %t", desc.Skipped),
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	if len(desc.Setup) > 0 {
		if _, err := fmt.Fprintln(out, "setup:"); err != nil {
			return err
		}
		for _, entry := range desc.Setup {
			if _, err := fmt.Fprintf(out, "  - %s\n", entry); err != nil {
				return err
			}
		}
	}
	if len(desc.SkipTargets) > 0 {
		if _, err := fmt.Fprintln(out, "skipTargets:"); err != nil {
			return err
		}
		for _, entry := range desc.SkipTargets {
			if _, err := fmt.Fprintf(out, "  - %s\n", entry); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeBatchJSONL(out io.Writer, outcomes []fixtureBatchRunOutput) error {
	for _, outcome := range outcomes {
		if err := writeJSONTo(out, outcome); err != nil {
			return err
		}
	}
	return nil
}
