package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/compiler"
	"able/interpreter-go/pkg/driver"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("ablec", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	monoArraysEnabled, err := resolveAblecExperimentalMonoArraysFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	outputDir := fs.String("o", "", "output directory for generated Go code")
	pkgName := fs.String("pkg", "", "Go package name for generated code")
	emitMain := fs.Bool("main", false, "emit a runnable main.go wrapper (package must be main)")
	buildBin := fs.Bool("build", false, "build a native binary after emitting Go code (forces -pkg=main)")
	binPath := fs.String("bin", "", "output path for built binary (defaults to <output dir>/compiled)")
	requireNoFallbacks := fs.Bool("no-fallbacks", false, "fail compilation when any interpreter fallback is required")
	experimentalMonoArrays := fs.Bool("experimental-mono-arrays", monoArraysEnabled, "legacy compatibility flag; native static Array lowering is always enabled")
	noExperimentalMonoArrays := fs.Bool("no-experimental-mono-arrays", false, "legacy compatibility flag; native static Array lowering remains enabled")
	experimentalExecutionContext := fs.Bool("experimental-execution-context", false, "force generated-call execution-context propagation for diagnostic comparison")
	dynamicBoundaryTelemetry := fs.Bool("dynamic-boundary-telemetry", false, "emit debug-only dynamic-boundary counters in generated code")
	callPathTelemetry := fs.Bool("call-path-telemetry", false, "emit debug-only generated call-path counters in generated code")
	typedBoundaryTelemetry := fs.Bool("typed-boundary-telemetry", false, "emit debug-only typed/runtime boundary counters in generated code")
	nominalEffectsJSON := fs.String("nominal-effects-json", "", "write conservative typed nominal callable effects to this JSON file")
	nominalOwnershipJSON := fs.String("nominal-ownership-json", "", "write fail-closed nominal ownership-transfer proofs to this JSON file")
	experimentalNominalOwnership := fs.Bool("experimental-nominal-ownership", false, "legacy compatibility flag; proven caller-owned nominal-result lowering is enabled by default")
	noNominalOwnership := fs.Bool("no-nominal-ownership", false, "disable proven caller-owned nominal-result lowering for diagnostic comparison")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *noExperimentalMonoArrays {
		*experimentalMonoArrays = true
	}
	if *experimentalNominalOwnership && *noNominalOwnership {
		fmt.Fprintln(os.Stderr, "ablec: --experimental-nominal-ownership and --no-nominal-ownership are mutually exclusive")
		return 2
	}

	entry := fs.Arg(0)
	if entry == "" {
		fmt.Fprintln(os.Stderr, "usage: ablec [options] <entry.able>")
		fs.PrintDefaults()
		return 2
	}

	absEntry, err := filepath.Abs(entry)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if *buildBin {
		*emitMain = true
		if *pkgName != "" && *pkgName != "main" {
			fmt.Fprintln(os.Stderr, "ablec: -build requires -pkg=main")
			return 2
		}
		*pkgName = "main"
	}

	if *emitMain {
		if *pkgName != "" && *pkgName != "main" {
			fmt.Fprintln(os.Stderr, "ablec: -main requires -pkg=main")
			return 2
		}
		*pkgName = "main"
	}

	if *outputDir == "" {
		*outputDir = filepath.Join("target", "compiled")
	}

	searchPaths := collectSearchPaths(filepath.Dir(absEntry))
	searchPaths, err = finalizeSearchPaths(searchPaths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer loader.Close()

	program, err := loader.Load(absEntry)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	comp := compiler.New(compiler.Options{
		PackageName:                  *pkgName,
		EmitMain:                     *emitMain,
		EntryPath:                    absEntry,
		RequireNoFallbacks:           *requireNoFallbacks,
		RequireStaticNoFallbacks:     *requireNoFallbacks,
		ExperimentalMonoArrays:       *experimentalMonoArrays,
		ExperimentalMonoArraysSet:    true,
		ExperimentalExecutionContext: *experimentalExecutionContext,
		EmitDynamicBoundaryTelemetry: *dynamicBoundaryTelemetry,
		EmitCallPathTelemetry:        *callPathTelemetry,
		EmitTypedBoundaryTelemetry:   *typedBoundaryTelemetry,
		CollectNominalEffects:        *nominalEffectsJSON != "",
		CollectNominalOwnership:      *nominalOwnershipJSON != "",
		ExperimentalNominalOwnership: *experimentalNominalOwnership,
		DisableNominalOwnership:      *noNominalOwnership,
	})
	result, err := comp.Compile(program)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, warning := range result.Warnings {
		fmt.Fprintln(os.Stderr, warning)
	}
	if *nominalEffectsJSON != "" {
		if err := writeNominalEffectsJSON(*nominalEffectsJSON, result.NominalEffects); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	if *nominalOwnershipJSON != "" {
		if err := writeNominalOwnershipJSON(*nominalOwnershipJSON, result.NominalOwnership); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	if err := result.Write(*outputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := prepareBuildModule(*outputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if *buildBin {
		out := *binPath
		if out == "" {
			out = filepath.Join(*outputDir, "compiled")
		}
		cmd := exec.Command("go", "build", "-mod=mod", "-o", out, ".")
		cmd.Dir = *outputDir
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "ablec: go build failed: %v\n%s\n", err, string(output))
			return 1
		}
	}

	return 0
}

func writeNominalEffectsJSON(path string, report *compiler.NominalEffectReport) error {
	if report == nil {
		return fmt.Errorf("ablec: nominal effect report was not collected")
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("ablec: encode nominal effects: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("ablec: write nominal effects: %w", err)
	}
	return nil
}

func writeNominalOwnershipJSON(path string, report *compiler.NominalOwnershipReport) error {
	if report == nil {
		return fmt.Errorf("ablec: nominal ownership report was not collected")
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("ablec: encode nominal ownership: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("ablec: write nominal ownership: %w", err)
	}
	return nil
}

func resolveAblecExperimentalMonoArraysFromEnv() (bool, error) {
	raw, ok := os.LookupEnv("ABLE_EXPERIMENTAL_MONO_ARRAYS")
	if !ok {
		return true, nil
	}
	normalized := strings.TrimSpace(strings.ToLower(raw))
	switch normalized {
	case "", "0", "false", "no", "off", "1", "true", "yes", "on":
		return true, nil
	default:
		return false, fmt.Errorf("invalid ABLE_EXPERIMENTAL_MONO_ARRAYS value %q (expected one of: 1,true,yes,on,0,false,no,off)", raw)
	}
}
