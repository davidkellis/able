package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/typechecker"
)

func TestDebugArrayBoundary(t *testing.T) {
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	dir := filepath.Join(root, "06_12_01_stdlib_string_helpers")
	manifest, err := interpreter.LoadFixtureManifest(dir)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	entryFile := manifest.Entry
	if entryFile == "" {
		entryFile = "main.able"
	}
	entryPath := filepath.Join(dir, entryFile)
	searchPaths, err := buildExecSearchPaths(entryPath, dir, manifest)
	if err != nil {
		t.Fatalf("search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	defer loader.Close()
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}
	outDir := filepath.Join(repositoryRoot(), "v12", "interpreters", "go", "tmp", "able-debug")
	os.MkdirAll(outDir, 0o755)

	// Build generator directly so we can inspect it
	checker := typechecker.NewProgramChecker()
	if _, err := checker.Check(program); err != nil {
		t.Fatalf("typecheck: %v", err)
	}
	gen := newGenerator(Options{PackageName: "main"})
	if err := gen.collect(program); err != nil {
		t.Fatalf("collect: %v", err)
	}
	dynamicReport, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("dynamic features: %v", err)
	}
	gen.setDynamicFeatureReport(dynamicReport)
	gen.resolveCompileableFunctions()
	gen.resolveCompileableMethods()

	files, err := gen.render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for name, content := range files {
		if writeErr := os.WriteFile(filepath.Join(outDir, name), content, 0o600); writeErr != nil {
			t.Fatalf("write %s: %v", name, writeErr)
		}
	}
	t.Logf("wrote compiled output to %s", outDir)

	type diagEntry struct {
		name   string
		reason string
	}
	var nonCompilable []diagEntry
	for _, info := range gen.allFunctionInfos() {
		if info == nil || info.Compileable || !info.SupportedTypes {
			continue
		}
		reason := info.Reason
		if reason == "" {
			reason = "(no reason)"
		}
		nonCompilable = append(nonCompilable, diagEntry{info.Name, reason})
	}
	for _, method := range gen.methodList {
		if method == nil || method.Info == nil || method.Info.Compileable || !method.Info.SupportedTypes {
			continue
		}
		reason := method.Info.Reason
		if reason == "" {
			reason = "(no reason)"
		}
		nonCompilable = append(nonCompilable, diagEntry{method.Info.Name, reason})
	}
	sort.Slice(nonCompilable, func(i, j int) bool {
		return nonCompilable[i].name < nonCompilable[j].name
	})

	reasonCounts := map[string]int{}
	for _, e := range nonCompilable {
		reasonCounts[e.reason]++
	}
	type rc struct {
		reason string
		count  int
	}
	var reasons []rc
	for r, c := range reasonCounts {
		reasons = append(reasons, rc{r, c})
	}
	sort.Slice(reasons, func(i, j int) bool {
		return reasons[i].count > reasons[j].count
	})

	fmt.Printf("\n=== Non-compilable (type-supported, body fails) ===\n")
	for _, e := range nonCompilable {
		fmt.Printf("  %-60s %s\n", e.name, e.reason)
	}
	fmt.Printf("\n=== Reason summary ===\n")
	for _, r := range reasons {
		fmt.Printf("  %3d  %s\n", r.count, r.reason)
	}
	fmt.Printf("Total: %d\n", len(nonCompilable))
}
