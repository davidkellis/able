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
	outputDir := flag.String("o", "", "output directory for generated Go code")
	pkgName := flag.String("pkg", "", "Go package name for generated code")
	emitMain := flag.Bool("main", false, "emit a runnable main.go wrapper (package must be main)")
	buildBin := flag.Bool("build", false, "build a native binary after emitting Go code (forces -pkg=main)")
	binPath := flag.String("bin", "", "output path for built binary (defaults to <output dir>/compiled)")
	flag.Parse()

	entry := flag.Arg(0)
	if entry == "" {
		fmt.Fprintln(os.Stderr, "usage: ablec [options] <entry.able>")
		flag.PrintDefaults()
		os.Exit(2)
	}

	absEntry, err := filepath.Abs(entry)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *buildBin {
		*emitMain = true
		if *pkgName != "" && *pkgName != "main" {
			fmt.Fprintln(os.Stderr, "ablec: -build requires -pkg=main")
			os.Exit(2)
		}
		*pkgName = "main"
	}

	if *emitMain {
		if *pkgName != "" && *pkgName != "main" {
			fmt.Fprintln(os.Stderr, "ablec: -main requires -pkg=main")
			os.Exit(2)
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
		os.Exit(1)
	}
	defer loader.Close()

	program, err := loader.Load(absEntry)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	comp := compiler.New(compiler.Options{
		PackageName: *pkgName,
		EmitMain:    *emitMain,
		EntryPath:   absEntry,
	})
	result, err := comp.Compile(program)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, warning := range result.Warnings {
		fmt.Fprintln(os.Stderr, warning)
	}
	if err := result.Write(*outputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := writeBuildGoMod(*outputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
			os.Exit(1)
		}
	}
}
