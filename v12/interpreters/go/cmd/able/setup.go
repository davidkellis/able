package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/driver"
)

func runSetup(args []string) int {
	prewarmExtern := false
	if len(args) == 1 && args[0] == "--prewarm-extern" {
		prewarmExtern = true
	} else if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "able setup accepts only --prewarm-extern (received %s)\n", strings.Join(args, " "))
		return 1
	}

	cacheDir, err := resolveAbleHome()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve ABLE_HOME: %v\n", err)
		return 1
	}

	// Extract embedded kernel.
	kernelPath, err := ensureEmbeddedKernel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract kernel: %v\n", err)
		return 1
	}
	kernelVersion := embeddedKernelVersion()
	fmt.Fprintf(os.Stdout, "kernel %s extracted to %s\n", kernelVersion, kernelPath)

	// Download stdlib via git.
	installer := newDependencyInstaller(nil, cacheDir)
	installer.manifestRoot = cacheDir
	resolved, err := installer.resolveStdlibDependency(&driver.DependencySpec{Version: defaultStdlibVersion})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to download stdlib: %v\n", err)
		return 1
	}
	for _, line := range installer.logs {
		fmt.Fprintln(os.Stdout, line)
	}

	stdlibPath := resolved.root
	if resolved.pkg != nil {
		if src := strings.TrimPrefix(resolved.pkg.Source, "path:"); src != "" {
			stdlibPath = src
		}
	}
	fmt.Fprintf(os.Stdout, "stdlib %s cached at %s\n", resolved.pkg.Version, stdlibPath)

	if prewarmExtern {
		searchPaths := setupPrewarmSearchPaths(resolvePackageSrcPath(stdlibPath), kernelPath)
		result, packageCount, prewarmErr := prewarmExternHostModules(searchPaths)
		if prewarmErr != nil {
			fmt.Fprintf(os.Stderr, "failed to prewarm extern host cache: %v\n", prewarmErr)
			return 1
		}
		fmt.Fprintf(os.Stdout, "prewarmed %d Go extern host modules from %d packages\n", result.Modules, packageCount)
	}

	// Write a marker lockfile so ad-hoc scripts know where to find stdlib.
	markerLock := driver.NewLockfile("_global", cliToolVersion)
	markerLock.Packages = []*driver.LockedPackage{
		{
			Name:    "kernel",
			Version: kernelVersion,
			Source:  fmt.Sprintf("path:%s", kernelPath),
		},
	}
	if resolved.pkg != nil {
		markerLock.Packages = append(markerLock.Packages, resolved.pkg)
	}
	markerPath := filepath.Join(cacheDir, "setup.lock")
	if err := driver.WriteLockfile(markerLock, markerPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to write setup marker: %v\n", err)
	}

	fmt.Fprintln(os.Stdout, "setup complete")
	return 0
}
