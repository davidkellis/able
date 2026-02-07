package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"able/interpreter-go/pkg/compiler"
	"able/interpreter-go/pkg/driver"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("ablec", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	outputDir := fs.String("o", "", "output directory for generated Go code")
	pkgName := fs.String("pkg", "", "Go package name for generated code")
	emitMain := fs.Bool("main", false, "emit a runnable main.go wrapper (package must be main)")
	buildBin := fs.Bool("build", false, "build a native binary after emitting Go code (forces -pkg=main)")
	binPath := fs.String("bin", "", "output path for built binary (defaults to <output dir>/compiled)")

	if err := fs.Parse(args); err != nil {
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
		PackageName: *pkgName,
		EmitMain:    *emitMain,
		EntryPath:   absEntry,
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
