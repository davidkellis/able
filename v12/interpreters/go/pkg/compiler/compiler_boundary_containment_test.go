package compiler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

type compilerBoundaryMarkerAudit struct {
	file  string
	lines []string
}

type compilerStaticLookupMarkers struct {
	MemberLookupCalls         int64
	InterfaceLookupCalls      int64
	GlobalLookupCalls         int64
	GlobalLookupEnvCalls      int64
	GlobalLookupRegistryCalls int64
}

func TestCompilerBoundaryExplicitHelperSetSourceAudit(t *testing.T) {
	root := filepath.Join(repositoryRoot(), "v12", "interpreters", "go", "pkg", "compiler")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read compiler dir: %v", err)
	}

	allowed := map[string]compilerBoundaryMarkerAudit{
		"generator_render_functions.go": {
			file:  "generator_render_functions.go",
			lines: []string{`__able_mark_boundary_explicit(\"call_original\", %q)`},
		},
		"generator_render_runtime_calls.go": {
			file:  "generator_render_runtime_calls.go",
			lines: []string{`__able_mark_boundary_explicit(\"call_value\", __able_boundary_name)`},
		},
		"generator_render_runtime_calls_tail.go": {
			file:  "generator_render_runtime_calls_tail.go",
			lines: []string{`__able_mark_boundary_explicit(\"call_named\", name)`},
		},
	}
	var failures []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		text := string(data)
		if !strings.Contains(text, "__able_mark_boundary_explicit(") {
			continue
		}
		audit, ok := allowed[name]
		if !ok {
			failures = append(failures, fmt.Sprintf("%s: unexpected explicit boundary marker emitter", name))
			continue
		}
		for _, line := range audit.lines {
			if !strings.Contains(text, line) {
				failures = append(failures, fmt.Sprintf("%s: missing explicit boundary marker line %q", name, line))
			}
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		t.Fatalf("boundary explicit-helper audit failed:\n%s", strings.Join(failures, "\n"))
	}
}

func TestCompilerNoBootstrapStaticFixturesStayBoundaryClean(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping no-bootstrap static boundary audit in short mode")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	for _, rel := range []string{
		"06_01_compiler_placeholder_lambda",
		"06_12_02_stdlib_array_helpers",
		"10_03_interface_type_dynamic_dispatch",
	} {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			runCompilerNoBootstrapBoundaryAuditFixture(t, root, rel)
		})
	}
}

func runCompilerNoBootstrapBoundaryAuditFixture(t *testing.T, root, rel string) {
	t.Helper()
	dir := filepath.Join(root, filepath.FromSlash(rel))
	manifest, err := interpreter.LoadFixtureManifest(dir)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if shouldSkipTarget(manifest.SkipTargets, "go") {
		return
	}
	if len(manifest.Expect.TypecheckDiagnostics) > 0 {
		return
	}
	entry := manifest.Entry
	if entry == "" {
		entry = "main.able"
	}
	entryPath := filepath.Join(dir, entry)

	searchPaths, err := buildExecSearchPaths(entryPath, dir, manifest)
	if err != nil {
		t.Fatalf("exec search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	program, err := loader.Load(entryPath)
	loader.Close()
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-noboot-boundary")
	comp := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: requireNoFallbacksForFixtureGates(t),
	})
	result, err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(workDir); err != nil {
		t.Fatalf("write output: %v", err)
	}

	harness := compilerHarnessSourceNoBootstrap(manifest.Executor)
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte(harness), 0o600); err != nil {
		t.Fatalf("write harness: %v", err)
	}

	binPath := filepath.Join(workDir, "compiled-fixture")
	if goruntime.GOOS == "windows" {
		binPath += ".exe"
	}
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = workDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binPath)
	cmdEnv := withEnv(os.Environ(), "ABLE_COMPILER_BOUNDARY_MARKER", "1")
	cmdEnv = withEnv(cmdEnv, "ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE", "1")
	cmdEnv = withEnv(cmdEnv, "ABLE_COMPILER_INTERFACE_LOOKUP_MARKER", "1")
	cmdEnv = withEnv(cmdEnv, "ABLE_COMPILER_GLOBAL_LOOKUP_MARKER", "1")
	cmd.Env = applyFixtureEnv(cmdEnv, manifest.Env)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("compiled fixture timed out after 60s")
	}

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("run error: %v", runErr)
		}
	}

	stderrLines := splitLines(stderr.String())
	filteredStderr, boundaryMarkers, err := extractBoundaryMarkers(stderrLines)
	if err != nil {
		t.Fatalf("extract boundary markers: %v\nstderr=%q stdout=%q", err, stderr.String(), stdout.String())
	}
	filteredStderr, lookupMarkers, err := extractStaticLookupMarkers(filteredStderr)
	if err != nil {
		t.Fatalf("extract lookup markers: %v\nstderr=%q stdout=%q", err, stderr.String(), stdout.String())
	}

	if boundaryMarkers.FallbackCount != 0 {
		t.Fatalf("expected no fallback boundary calls, got %d (%q)", boundaryMarkers.FallbackCount, boundaryMarkers.FallbackNames)
	}
	if boundaryMarkers.ExplicitCount != 0 {
		t.Fatalf("expected no explicit dynamic boundary calls, got %d (%q)", boundaryMarkers.ExplicitCount, boundaryMarkers.ExplicitNames)
	}
	if lookupMarkers.MemberLookupCalls != 0 || lookupMarkers.InterfaceLookupCalls != 0 {
		t.Fatalf("expected no interface/member lookup fallback calls, got total=%d interface=%d", lookupMarkers.MemberLookupCalls, lookupMarkers.InterfaceLookupCalls)
	}
	if lookupMarkers.GlobalLookupCalls != 0 || lookupMarkers.GlobalLookupEnvCalls != 0 || lookupMarkers.GlobalLookupRegistryCalls != 0 {
		t.Fatalf("expected no global lookup fallback calls, got total=%d env=%d registry=%d", lookupMarkers.GlobalLookupCalls, lookupMarkers.GlobalLookupEnvCalls, lookupMarkers.GlobalLookupRegistryCalls)
	}

	actualStdout := splitLines(stdout.String())
	if manifest.Expect.Stdout != nil {
		expectedStdout := expandFixtureLines(manifest.Expect.Stdout)
		if !reflect.DeepEqual(actualStdout, expectedStdout) {
			t.Fatalf("stdout mismatch: expected %v, got %v", expectedStdout, actualStdout)
		}
	}

	actualStderr := filteredStderr
	if manifest.Expect.Stderr != nil {
		expectedStderr := expandFixtureLines(manifest.Expect.Stderr)
		if !reflect.DeepEqual(actualStderr, expectedStderr) {
			t.Fatalf("stderr mismatch: expected %v, got %v", expectedStderr, actualStderr)
		}
	} else if len(actualStderr) > 0 {
		t.Fatalf("unexpected stderr: %v", actualStderr)
	}

	expectedExit := 0
	if manifest.Expect.Exit != nil {
		expectedExit = *manifest.Expect.Exit
	}
	if exitCode != expectedExit {
		t.Fatalf("exit mismatch: expected %d, got %d (stderr=%q stdout=%q)", expectedExit, exitCode, stderr.String(), stdout.String())
	}
}

func extractStaticLookupMarkers(lines []string) ([]string, compilerStaticLookupMarkers, error) {
	var markers compilerStaticLookupMarkers
	var rawMember string
	var rawInterface string
	var rawGlobal string
	var rawGlobalEnv string
	var rawGlobalRegistry string
	filtered := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		switch {
		case strings.HasPrefix(trimmed, "__ABLE_MEMBER_LOOKUP_CALLS="):
			rawMember = strings.TrimPrefix(trimmed, "__ABLE_MEMBER_LOOKUP_CALLS=")
		case strings.HasPrefix(trimmed, "__ABLE_MEMBER_LOOKUP_INTERFACE_CALLS="):
			rawInterface = strings.TrimPrefix(trimmed, "__ABLE_MEMBER_LOOKUP_INTERFACE_CALLS=")
		case strings.HasPrefix(trimmed, "__ABLE_MEMBER_LOOKUP_NAMES="):
			continue
		case strings.HasPrefix(trimmed, "__ABLE_GLOBAL_LOOKUP_FALLBACK_CALLS="):
			rawGlobal = strings.TrimPrefix(trimmed, "__ABLE_GLOBAL_LOOKUP_FALLBACK_CALLS=")
		case strings.HasPrefix(trimmed, "__ABLE_GLOBAL_LOOKUP_FALLBACK_ENV_CALLS="):
			rawGlobalEnv = strings.TrimPrefix(trimmed, "__ABLE_GLOBAL_LOOKUP_FALLBACK_ENV_CALLS=")
		case strings.HasPrefix(trimmed, "__ABLE_GLOBAL_LOOKUP_FALLBACK_REGISTRY_CALLS="):
			rawGlobalRegistry = strings.TrimPrefix(trimmed, "__ABLE_GLOBAL_LOOKUP_FALLBACK_REGISTRY_CALLS=")
		case strings.HasPrefix(trimmed, "__ABLE_GLOBAL_LOOKUP_FALLBACK_NAMES="):
			continue
		default:
			filtered = append(filtered, trimmed)
		}
	}

	if rawMember == "" || rawInterface == "" || rawGlobal == "" || rawGlobalEnv == "" || rawGlobalRegistry == "" {
		return nil, compilerStaticLookupMarkers{}, fmt.Errorf("missing lookup markers: member=%q interface=%q global=%q env=%q registry=%q", rawMember, rawInterface, rawGlobal, rawGlobalEnv, rawGlobalRegistry)
	}

	var err error
	if markers.MemberLookupCalls, err = strconv.ParseInt(rawMember, 10, 64); err != nil {
		return nil, compilerStaticLookupMarkers{}, fmt.Errorf("invalid member lookup marker %q: %w", rawMember, err)
	}
	if markers.InterfaceLookupCalls, err = strconv.ParseInt(rawInterface, 10, 64); err != nil {
		return nil, compilerStaticLookupMarkers{}, fmt.Errorf("invalid interface lookup marker %q: %w", rawInterface, err)
	}
	if markers.GlobalLookupCalls, err = strconv.ParseInt(rawGlobal, 10, 64); err != nil {
		return nil, compilerStaticLookupMarkers{}, fmt.Errorf("invalid global lookup marker %q: %w", rawGlobal, err)
	}
	if markers.GlobalLookupEnvCalls, err = strconv.ParseInt(rawGlobalEnv, 10, 64); err != nil {
		return nil, compilerStaticLookupMarkers{}, fmt.Errorf("invalid global lookup env marker %q: %w", rawGlobalEnv, err)
	}
	if markers.GlobalLookupRegistryCalls, err = strconv.ParseInt(rawGlobalRegistry, 10, 64); err != nil {
		return nil, compilerStaticLookupMarkers{}, fmt.Errorf("invalid global lookup registry marker %q: %w", rawGlobalRegistry, err)
	}

	return filtered, markers, nil
}
