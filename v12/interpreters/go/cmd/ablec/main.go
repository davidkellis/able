package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"able/interpreter-go/pkg/compiler"
	"able/interpreter-go/pkg/driver"
)

func main() {
	outputDir := flag.String("o", "", "output directory for generated Go code")
	pkgName := flag.String("pkg", "", "Go package name for generated code")
	emitMain := flag.Bool("main", false, "emit a runnable main.go wrapper (package must be main)")
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

	loader, err := driver.NewLoader(nil)
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
}
