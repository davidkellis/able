package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

func runCache(args []string) int {
	if len(args) > 0 && args[0] == "compiled-tests" {
		return runCompiledTestCacheCommand(args[1:])
	}
	if len(args) != 1 || args[0] != "prewarm" {
		fmt.Fprintln(os.Stderr, "usage: able cache prewarm")
		printCompiledTestCacheUsage()
		return 1
	}
	searchPaths, err := cachePrewarmSearchPaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve canonical extern sources: %v\n", err)
		return 1
	}
	result, packageCount, err := prewarmExternHostModules(searchPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to prewarm extern host cache: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stdout, "prewarmed %d Go extern host modules from %d packages\n", result.Modules, packageCount)
	return 0
}

func cachePrewarmSearchPaths() ([]driver.SearchPath, error) {
	paths := collectSearchPaths("", searchPathOptions{skipStdlibDiscovery: true})
	if configuredStdlib := strings.TrimSpace(os.Getenv("ABLE_STDLIB_ROOT")); configuredStdlib != "" {
		paths = append([]driver.SearchPath{{
			Path:         configuredStdlib,
			Kind:         driver.RootStdlib,
			StdlibSource: driver.StdlibSourceEnv,
		}}, paths...)
	}
	return finalizeSearchPaths(paths, false)
}

func prewarmExternHostModules(searchPaths []driver.SearchPath) (interpreter.ExternHostPrewarmResult, int, error) {
	sources, err := driver.DiscoverPackages(searchPaths, false)
	if err != nil {
		return interpreter.ExternHostPrewarmResult{}, 0, err
	}
	if len(sources) == 0 {
		return interpreter.ExternHostPrewarmResult{}, 0, fmt.Errorf("no Able packages found in canonical search paths")
	}

	entryPath := ""
	packageNames := make([]string, 0, len(sources))
	for _, source := range sources {
		if len(source.Files) == 0 {
			continue
		}
		if entryPath == "" || source.Files[0] < entryPath {
			entryPath = source.Files[0]
		}
		packageNames = append(packageNames, source.Name)
	}
	if entryPath == "" {
		return interpreter.ExternHostPrewarmResult{}, 0, fmt.Errorf("no Able source files found in canonical search paths")
	}

	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		return interpreter.ExternHostPrewarmResult{}, 0, err
	}
	defer loader.Close()
	program, err := loader.LoadWithOptions(entryPath, driver.LoadOptions{IncludePackages: packageNames})
	if err != nil {
		return interpreter.ExternHostPrewarmResult{}, 0, err
	}
	interp, err := newInterpreter(interpreterBytecode)
	if err != nil {
		return interpreter.ExternHostPrewarmResult{}, 0, err
	}
	result, err := interp.PrewarmExternHostModules(program)
	if err != nil {
		return interpreter.ExternHostPrewarmResult{}, 0, err
	}
	return result, len(sources), nil
}

func setupPrewarmSearchPaths(stdlibPath, kernelPath string) []driver.SearchPath {
	paths := make([]driver.SearchPath, 0, 2)
	if stdlibPath != "" {
		paths = append(paths, driver.SearchPath{
			Path:         filepath.Clean(stdlibPath),
			Kind:         driver.RootStdlib,
			StdlibSource: driver.StdlibSourceCache,
		})
	}
	if kernelPath != "" {
		paths = append(paths, driver.SearchPath{
			Path: filepath.Clean(kernelPath),
			Kind: driver.RootStdlib,
		})
	}
	return paths
}
