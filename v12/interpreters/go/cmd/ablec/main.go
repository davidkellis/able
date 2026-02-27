package main

import (
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
	experimentalMonoArrays := fs.Bool("experimental-mono-arrays", monoArraysEnabled, "enable staged monomorphized array lowering (experimental, default on)")
	noExperimentalMonoArrays := fs.Bool("no-experimental-mono-arrays", false, "disable staged monomorphized array lowering")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *noExperimentalMonoArrays {
		*experimentalMonoArrays = false
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
		PackageName:            *pkgName,
		EmitMain:               *emitMain,
		EntryPath:              absEntry,
		ExperimentalMonoArrays: *experimentalMonoArrays,
	})
	result, err := comp.Compile(program)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, warning := range result.Warnings {
		fmt.Fprintln(os.Stderr, warning)
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

func resolveAblecExperimentalMonoArraysFromEnv() (bool, error) {
	raw, ok := os.LookupEnv("ABLE_EXPERIMENTAL_MONO_ARRAYS")
	if !ok {
		return true, nil
	}
	normalized := strings.TrimSpace(strings.ToLower(raw))
	switch normalized {
	case "", "0", "false", "no", "off":
		return false, nil
	case "1", "true", "yes", "on":
		return true, nil
	default:
		return false, fmt.Errorf("invalid ABLE_EXPERIMENTAL_MONO_ARRAYS value %q (expected one of: 1,true,yes,on,0,false,no,off)", raw)
	}
}
