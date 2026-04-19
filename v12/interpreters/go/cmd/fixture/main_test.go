package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/interpreter"
)

func TestFindFixtureStdlibRootPrefersCachedAbleHome(t *testing.T) {
	repoRoot := t.TempDir()
	cacheHome := filepath.Join(t.TempDir(), ".able")
	cacheSrc := filepath.Join(cacheHome, "pkg", "src", "able", "0.1.0", "src")
	siblingSrc := filepath.Join(repoRoot, "able-stdlib", "src")

	if err := os.MkdirAll(cacheSrc, 0o755); err != nil {
		t.Fatalf("mkdir cache src: %v", err)
	}
	if err := os.MkdirAll(siblingSrc, 0o755); err != nil {
		t.Fatalf("mkdir sibling src: %v", err)
	}

	t.Setenv("ABLE_STDLIB_ROOT", "")
	t.Setenv("ABLE_HOME", cacheHome)

	if got := findFixtureStdlibRoot(repoRoot); got != cacheSrc {
		t.Fatalf("findFixtureStdlibRoot() = %q, want %q", got, cacheSrc)
	}
}

func TestRunFixtureUsesSharedReplayHelper(t *testing.T) {
	fixtureDir := t.TempDir()
	writeFile := func(path, contents string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	writeFile(filepath.Join(fixtureDir, "source.able"), "package sample\n\nprint(\"hello\")\n\"done\"\n")

	outcome, err := runFixture(fixtureDir, "source.able", nil, nil)
	if err != nil {
		t.Fatalf("runFixture: %v", err)
	}
	if outcome.Error != "" {
		t.Fatalf("unexpected fixture error %q", outcome.Error)
	}
	if len(outcome.Stdout) != 1 || outcome.Stdout[0] != "hello" {
		t.Fatalf("stdout = %v, want [hello]", outcome.Stdout)
	}
	if outcome.Result == nil || outcome.Result.Kind != "String" || outcome.Result.Value != "done" {
		t.Fatalf("result = %#v, want String(done)", outcome.Result)
	}
	if outcome.TypecheckMode == "" {
		t.Fatalf("expected non-empty typecheck mode")
	}
}

func TestResolveFixtureInputUsesPositionalDirectory(t *testing.T) {
	fixtureDir := t.TempDir()

	dir, entry, err := resolveFixtureInput("", "", []string{fixtureDir})
	if err != nil {
		t.Fatalf("resolveFixtureInput: %v", err)
	}
	if dir != fixtureDir {
		t.Fatalf("dir = %q, want %q", dir, fixtureDir)
	}
	if entry != "" {
		t.Fatalf("entry = %q, want empty", entry)
	}
}

func TestResolveFixtureInputUsesPositionalFileAsEntry(t *testing.T) {
	fixtureDir := t.TempDir()
	entryPath := filepath.Join(fixtureDir, "source.able")
	if err := os.WriteFile(entryPath, []byte("package sample\n"), 0o644); err != nil {
		t.Fatalf("write source fixture: %v", err)
	}

	dir, entry, err := resolveFixtureInput("", "", []string{entryPath})
	if err != nil {
		t.Fatalf("resolveFixtureInput: %v", err)
	}
	if dir != fixtureDir {
		t.Fatalf("dir = %q, want %q", dir, fixtureDir)
	}
	if entry != "source.able" {
		t.Fatalf("entry = %q, want source.able", entry)
	}
}

func TestResolveFixtureInputRejectsDirAndPositionalTarget(t *testing.T) {
	fixtureDir := t.TempDir()

	_, _, err := resolveFixtureInput(fixtureDir, "", []string{fixtureDir})
	if err == nil {
		t.Fatalf("expected dir/target conflict")
	}
	if err.Error() != "fixture target cannot be used together with --dir" {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestResolveFixtureInputRejectsEntryFlagForPositionalFile(t *testing.T) {
	fixtureDir := t.TempDir()
	entryPath := filepath.Join(fixtureDir, "source.able")
	if err := os.WriteFile(entryPath, []byte("package sample\n"), 0o644); err != nil {
		t.Fatalf("write source fixture: %v", err)
	}

	_, _, err := resolveFixtureInput("", "module.json", []string{entryPath})
	if err == nil {
		t.Fatalf("expected entry/file conflict")
	}
	if err.Error() != "--entry cannot be used when the fixture target is already a file" {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestResolveFixtureInputUsesRepoRelativeDirectoryTarget(t *testing.T) {
	fixturesRoot := t.TempDir()
	fixtureDir := filepath.Join(fixturesRoot, "basics", "bool_literal")
	if err := os.MkdirAll(fixtureDir, 0o755); err != nil {
		t.Fatalf("mkdir fixture dir: %v", err)
	}

	dir, entry, err := resolveFixtureInputWithFixturesRoot("", "", []string{"basics/bool_literal"}, fixturesRoot)
	if err != nil {
		t.Fatalf("resolveFixtureInputWithFixturesRoot: %v", err)
	}
	if dir != fixtureDir {
		t.Fatalf("dir = %q, want %q", dir, fixtureDir)
	}
	if entry != "" {
		t.Fatalf("entry = %q, want empty", entry)
	}
}

func TestResolveFixtureInputUsesRepoRelativeFileTarget(t *testing.T) {
	fixturesRoot := t.TempDir()
	fixtureDir := filepath.Join(fixturesRoot, "basics", "bool_literal")
	if err := os.MkdirAll(fixtureDir, 0o755); err != nil {
		t.Fatalf("mkdir fixture dir: %v", err)
	}
	entryPath := filepath.Join(fixtureDir, "source.able")
	if err := os.WriteFile(entryPath, []byte("package sample\n"), 0o644); err != nil {
		t.Fatalf("write source fixture: %v", err)
	}

	dir, entry, err := resolveFixtureInputWithFixturesRoot("", "", []string{"basics/bool_literal/source.able"}, fixturesRoot)
	if err != nil {
		t.Fatalf("resolveFixtureInputWithFixturesRoot: %v", err)
	}
	if dir != fixtureDir {
		t.Fatalf("dir = %q, want %q", dir, fixtureDir)
	}
	if entry != "source.able" {
		t.Fatalf("entry = %q, want source.able", entry)
	}
}

func TestRepositoryRootFromFixtureSuffixFallback(t *testing.T) {
	root := t.TempDir()
	start := filepath.Join(root, "v12", "interpreters", "go", "cmd", "fixture")
	if err := os.MkdirAll(start, 0o755); err != nil {
		t.Fatalf("mkdir start: %v", err)
	}

	if got := repositoryRootFrom(start); got != root {
		t.Fatalf("repositoryRootFrom(%q) = %q, want %q", start, got, root)
	}
}

func TestListFixtureTargetsWithRootListsRelativeFixtureDirs(t *testing.T) {
	fixturesRoot := t.TempDir()
	for _, rel := range []string{
		"basics/bool_literal/source.able",
		"control/for_sum/source.able",
	} {
		path := filepath.Join(fixturesRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir fixture dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("package sample\n"), 0o644); err != nil {
			t.Fatalf("write fixture source: %v", err)
		}
	}

	var stdout bytes.Buffer
	if err := listFixtureTargetsWithRoot("", "", "", nil, outputFormatText, fixturesRoot, &stdout); err != nil {
		t.Fatalf("listFixtureTargetsWithRoot: %v", err)
	}
	got := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	want := []string{"basics/bool_literal", "control/for_sum"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("listed fixtures = %v, want %v", got, want)
	}
}

func TestListFixtureTargetsWithRootFiltersByPrefix(t *testing.T) {
	fixturesRoot := t.TempDir()
	for _, rel := range []string{
		"basics/bool_literal/source.able",
		"basics/char_literal/source.able",
		"control/for_sum/source.able",
	} {
		path := filepath.Join(fixturesRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir fixture dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("package sample\n"), 0o644); err != nil {
			t.Fatalf("write fixture source: %v", err)
		}
	}

	var stdout bytes.Buffer
	if err := listFixtureTargetsWithRoot("", "", "", []string{"basics"}, outputFormatText, fixturesRoot, &stdout); err != nil {
		t.Fatalf("listFixtureTargetsWithRoot: %v", err)
	}
	got := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	want := []string{"basics/bool_literal", "basics/char_literal"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("listed fixtures = %v, want %v", got, want)
	}
}

func TestListFixtureTargetsWithRootRejectsDirFlag(t *testing.T) {
	err := listFixtureTargetsWithRoot("basics/bool_literal", "", "serial", nil, outputFormatText, t.TempDir(), &bytes.Buffer{})
	if err == nil {
		t.Fatalf("expected --dir conflict")
	}
	if err.Error() != "--list cannot be used together with --dir" {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestListFixtureTargetsWithRootWritesJSON(t *testing.T) {
	fixturesRoot := t.TempDir()
	for _, rel := range []string{
		"basics/bool_literal/source.able",
		"basics/char_literal/source.able",
	} {
		path := filepath.Join(fixturesRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir fixture dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("package sample\n"), 0o644); err != nil {
			t.Fatalf("write fixture source: %v", err)
		}
	}

	var stdout bytes.Buffer
	if err := listFixtureTargetsWithRoot("", "", "", []string{"basics"}, outputFormatJSON, fixturesRoot, &stdout); err != nil {
		t.Fatalf("listFixtureTargetsWithRoot: %v", err)
	}
	var got []string
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal list json: %v", err)
	}
	want := []string{"basics/bool_literal", "basics/char_literal"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("listed fixtures = %v, want %v", got, want)
	}
}

func TestResolveFixtureBatchInputsWithRootUsesMultipleTargets(t *testing.T) {
	fixturesRoot := t.TempDir()
	for _, rel := range []string{
		"basics/bool_literal/source.able",
		"control/for_sum/source.able",
	} {
		path := filepath.Join(fixturesRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir fixture dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("package sample\n"), 0o644); err != nil {
			t.Fatalf("write fixture source: %v", err)
		}
	}

	targets, err := resolveFixtureBatchInputsWithFixturesRoot("", "", []string{"basics/bool_literal", "control/for_sum/source.able"}, fixturesRoot)
	if err != nil {
		t.Fatalf("resolveFixtureBatchInputsWithFixturesRoot: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("len(targets) = %d, want 2", len(targets))
	}
	if targets[0].Dir != filepath.Join(fixturesRoot, "basics", "bool_literal") || targets[0].EntryOverride != "" {
		t.Fatalf("target[0] = %#v", targets[0])
	}
	if targets[1].Dir != filepath.Join(fixturesRoot, "control", "for_sum") || targets[1].EntryOverride != "source.able" {
		t.Fatalf("target[1] = %#v", targets[1])
	}
}

func TestResolveFixtureBatchInputsRejectsEntryFlag(t *testing.T) {
	_, err := resolveFixtureBatchInputsWithFixturesRoot("", "module.json", []string{"basics/bool_literal"}, t.TempDir())
	if err == nil {
		t.Fatalf("expected --entry conflict")
	}
	if err.Error() != "--batch cannot be used together with --entry" {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestValidateTopLevelModesRejectsConflicts(t *testing.T) {
	cases := []struct {
		listMode     bool
		describeMode bool
		batchMode    bool
		want         string
	}{
		{listMode: true, describeMode: true, want: "--list cannot be used together with --describe"},
		{listMode: true, batchMode: true, want: "--list cannot be used together with --batch"},
		{describeMode: true, batchMode: true, want: "--batch cannot be used together with --describe"},
	}
	for _, tc := range cases {
		err := validateTopLevelModes(tc.listMode, tc.describeMode, tc.batchMode)
		if err == nil {
			t.Fatalf("expected mode conflict for %+v", tc)
		}
		if err.Error() != tc.want {
			t.Fatalf("error = %q, want %q", err.Error(), tc.want)
		}
	}
}

func TestResolveOutputFormatByMode(t *testing.T) {
	cases := []struct {
		name         string
		requested    string
		listMode     bool
		describeMode bool
		batchMode    bool
		want         string
		wantErr      string
	}{
		{name: "list default", listMode: true, want: outputFormatText},
		{name: "list json", requested: "json", listMode: true, want: outputFormatJSON},
		{name: "describe text", requested: "text", describeMode: true, want: outputFormatText},
		{name: "describe default", describeMode: true, want: outputFormatJSON},
		{name: "batch jsonl", requested: "jsonl", batchMode: true, want: outputFormatJSONL},
		{name: "single default", want: outputFormatJSON},
		{name: "list rejects jsonl", requested: "jsonl", listMode: true, wantErr: "--list supports --format text or json"},
		{name: "describe rejects jsonl", requested: "jsonl", describeMode: true, wantErr: "--describe supports --format json or text"},
		{name: "batch rejects text", requested: "text", batchMode: true, wantErr: "--batch supports --format json or jsonl"},
		{name: "single rejects text", requested: "text", wantErr: "single fixture replay only supports --format json"},
	}
	for _, tc := range cases {
		got, err := resolveOutputFormat(tc.requested, tc.listMode, tc.describeMode, tc.batchMode)
		if tc.wantErr != "" {
			if err == nil {
				t.Fatalf("%s: expected error", tc.name)
			}
			if err.Error() != tc.wantErr {
				t.Fatalf("%s: error = %q, want %q", tc.name, err.Error(), tc.wantErr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: resolveOutputFormat: %v", tc.name, err)
		}
		if got != tc.want {
			t.Fatalf("%s: format = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestResolveFixtureExecutorPrefersManifestExecutor(t *testing.T) {
	name, _, err := resolveFixtureExecutor("goroutine", "")
	if err != nil {
		t.Fatalf("resolveFixtureExecutor: %v", err)
	}
	if name != "goroutine" {
		t.Fatalf("executor = %q, want goroutine", name)
	}
}

func TestResolveFixtureExecutorOverrideWins(t *testing.T) {
	name, _, err := resolveFixtureExecutor("goroutine", "serial")
	if err != nil {
		t.Fatalf("resolveFixtureExecutor: %v", err)
	}
	if name != "serial" {
		t.Fatalf("executor = %q, want serial", name)
	}
}

func TestBuildFixtureDescriptionUsesResolvedMetadata(t *testing.T) {
	t.Setenv("ABLE_TYPECHECK_FIXTURES", "warn")

	manifest := interpreter.FixtureManifest{
		Description: "sample fixture",
		Entry:       "module.json",
		Setup:       []string{"setup.able"},
		Executor:    "goroutine",
		SkipTargets: []string{"go", "ts"},
	}

	desc := buildFixtureDescription("/tmp/fixture", manifest, "source.able", "goroutine")
	if desc.Description != "sample fixture" {
		t.Fatalf("description = %q", desc.Description)
	}
	if desc.FixtureDir != "/tmp/fixture" {
		t.Fatalf("fixtureDir = %q", desc.FixtureDir)
	}
	if desc.Entry != "source.able" {
		t.Fatalf("entry = %q", desc.Entry)
	}
	if desc.Executor != "goroutine" {
		t.Fatalf("executor = %q", desc.Executor)
	}
	if desc.TypecheckMode != "warn" {
		t.Fatalf("typecheckMode = %q", desc.TypecheckMode)
	}
	if !desc.Skipped {
		t.Fatalf("expected skipped=true")
	}
	if strings.Join(desc.Setup, ",") != "setup.able" {
		t.Fatalf("setup = %v", desc.Setup)
	}
	if strings.Join(desc.SkipTargets, ",") != "go,ts" {
		t.Fatalf("skipTargets = %v", desc.SkipTargets)
	}
}

func TestDescribeFixtureOutputJSONShape(t *testing.T) {
	t.Setenv("ABLE_TYPECHECK_FIXTURES", "strict")

	manifest := interpreter.FixtureManifest{
		Description: "sample fixture",
		Setup:       []string{"setup.able"},
		SkipTargets: []string{"go"},
	}
	out := buildFixtureDescription("/tmp/fixture", manifest, "module.json", "serial")
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal description: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal description: %v", err)
	}
	if decoded["fixtureDir"] != "/tmp/fixture" {
		t.Fatalf("fixtureDir = %#v", decoded["fixtureDir"])
	}
	if decoded["entry"] != "module.json" {
		t.Fatalf("entry = %#v", decoded["entry"])
	}
	if decoded["executor"] != "serial" {
		t.Fatalf("executor = %#v", decoded["executor"])
	}
	if decoded["typecheckMode"] != "strict" {
		t.Fatalf("typecheckMode = %#v", decoded["typecheckMode"])
	}
	if decoded["skipped"] != true {
		t.Fatalf("skipped = %#v", decoded["skipped"])
	}
}

func TestWriteDescribeTextOutputsReadableSummary(t *testing.T) {
	desc := fixtureDescriptionOutput{
		Description:   "sample fixture",
		FixtureDir:    "/tmp/fixture",
		Entry:         "module.json",
		Executor:      "goroutine",
		TypecheckMode: "warn",
		Skipped:       true,
		Setup:         []string{"setup.able"},
		SkipTargets:   []string{"go"},
	}

	var stdout bytes.Buffer
	if err := writeDescribeText(&stdout, desc); err != nil {
		t.Fatalf("writeDescribeText: %v", err)
	}
	text := stdout.String()
	for _, want := range []string{
		"description: sample fixture",
		"fixtureDir: /tmp/fixture",
		"entry: module.json",
		"executor: goroutine",
		"typecheckMode: warn",
		"skipped: true",
		"setup:\n  - setup.able",
		"skipTargets:\n  - go",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("describe text missing %q in %q", want, text)
		}
	}
}

func TestRunFixtureBatchReplaysMultipleTargets(t *testing.T) {
	fixtureA := t.TempDir()
	fixtureB := t.TempDir()
	for _, fixture := range []string{fixtureA, fixtureB} {
		if err := os.WriteFile(filepath.Join(fixture, "source.able"), []byte("package sample\n"), 0o644); err != nil {
			t.Fatalf("write source fixture: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(fixtureA, "source.able"), []byte("package sample\n\ntrue\n"), 0o644); err != nil {
		t.Fatalf("write fixtureA: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fixtureB, "source.able"), []byte("package sample\n\n\"done\"\n"), 0o644); err != nil {
		t.Fatalf("write fixtureB: %v", err)
	}

	results, err := runFixtureBatch([]resolvedFixtureTarget{
		{Dir: fixtureA, EntryOverride: "source.able"},
		{Dir: fixtureB, EntryOverride: "source.able"},
	}, "")
	if err != nil {
		t.Fatalf("runFixtureBatch: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].FixtureDir != fixtureA || results[0].Entry != "source.able" || results[0].Executor != "serial" {
		t.Fatalf("result[0] metadata = %#v", results[0])
	}
	if results[0].Result == nil || results[0].Result.Kind != "bool" || results[0].Result.Bool == nil || !*results[0].Result.Bool {
		t.Fatalf("result[0] payload = %#v", results[0].Result)
	}
	if results[1].FixtureDir != fixtureB || results[1].Entry != "source.able" || results[1].Executor != "serial" {
		t.Fatalf("result[1] metadata = %#v", results[1])
	}
	if results[1].Result == nil || results[1].Result.Kind != "String" || results[1].Result.Value != "done" {
		t.Fatalf("result[1] payload = %#v", results[1].Result)
	}
}

func TestWriteBatchJSONLWritesOneObjectPerLine(t *testing.T) {
	results := []fixtureBatchRunOutput{
		{FixtureDir: "/tmp/a", Entry: "module.json", Executor: "serial", parityOutput: parityOutput{TypecheckMode: "strict", Skipped: false}},
		{FixtureDir: "/tmp/b", Entry: "source.able", Executor: "goroutine", parityOutput: parityOutput{TypecheckMode: "warn", Skipped: true}},
	}

	var stdout bytes.Buffer
	if err := writeBatchJSONL(&stdout, results); err != nil {
		t.Fatalf("writeBatchJSONL: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("line count = %d, want 2", len(lines))
	}
	for i, line := range lines {
		var decoded map[string]any
		if err := json.Unmarshal([]byte(line), &decoded); err != nil {
			t.Fatalf("line %d unmarshal: %v", i, err)
		}
		if decoded["fixtureDir"] != results[i].FixtureDir {
			t.Fatalf("line %d fixtureDir = %#v, want %q", i, decoded["fixtureDir"], results[i].FixtureDir)
		}
	}
}
