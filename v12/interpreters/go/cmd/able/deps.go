package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/driver"
)

func runDeps(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "able deps requires a subcommand (install, update)")
		return 1
	}
	switch args[0] {
	case "install":
		if len(args) > 1 {
			fmt.Fprintf(os.Stderr, "able deps install does not take arguments (received %s)\n", strings.Join(args[1:], " "))
			return 1
		}
		return runDepsInstall()
	case "update":
		return runDepsUpdate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown deps subcommand %q\n", args[0])
		return 1
	}
}

func runDepsInstall() int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to determine working directory: %v\n", err)
		return 1
	}
	manifestPath, err := findManifest(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to locate package.yml: %v\n", err)
		return 1
	}
	manifest, err := driver.LoadManifest(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read manifest: %v\n", err)
		return 1
	}
	cacheDir, err := resolveAbleHome()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve ABLE_HOME: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stdout, "Manifest: %s\n", manifest.Path)
	fmt.Fprintf(os.Stdout, "Root package: %s\n", manifest.Name)
	fmt.Fprintf(os.Stdout, "Dependencies: %d\n", len(manifest.Dependencies))
	fmt.Fprintf(os.Stdout, "Cache directory: %s\n", cacheDir)

	lockPath := filepath.Join(filepath.Dir(manifest.Path), "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	lockCreated := false
	switch {
	case err == nil:
		if lock.Root != manifest.Name {
			fmt.Fprintf(os.Stderr, "lockfile root %q does not match manifest name %q\n", lock.Root, manifest.Name)
			return 1
		}
	case errors.Is(err, os.ErrNotExist):
		lock = driver.NewLockfile(manifest.Name, cliToolVersion)
		lock.Path = lockPath
		lockCreated = true
	default:
		fmt.Fprintf(os.Stderr, "failed to read lockfile: %v\n", err)
		return 1
	}

	lock.Path = lockPath
	lock.Tool = cliToolVersion

	installer := newDependencyInstaller(manifest, cacheDir)
	changed, logs, err := installer.Install(lock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve dependencies: %v\n", err)
		return 1
	}
	for _, line := range logs {
		fmt.Fprintln(os.Stdout, line)
	}

	if changed || lockCreated {
		action := "Updated"
		if lockCreated {
			action = "Created"
		}
		if err := driver.WriteLockfile(lock, lockPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write lockfile: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stdout, "%s package.lock: %s\n", action, lock.Path)
	} else {
		fmt.Fprintf(os.Stdout, "package.lock already up to date: %s\n", lock.Path)
	}

	fmt.Fprintln(os.Stdout, "Dependencies installed.")
	return 0
}

func runDepsUpdate(targets []string) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to determine working directory: %v\n", err)
		return 1
	}
	manifestPath, err := findManifest(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to locate package.yml: %v\n", err)
		return 1
	}
	manifest, err := driver.LoadManifest(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read manifest: %v\n", err)
		return 1
	}
	cacheDir, err := resolveAbleHome()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve ABLE_HOME: %v\n", err)
		return 1
	}

	updateSet := make(map[string]struct{})
	if len(targets) > 0 {
		manifestDeps := make(map[string]struct{}, len(manifest.Dependencies))
		for name := range manifest.Dependencies {
			manifestDeps[sanitizeName(name)] = struct{}{}
		}
		for _, target := range targets {
			sanitized := sanitizeName(target)
			if _, ok := manifestDeps[sanitized]; !ok {
				fmt.Fprintf(os.Stderr, "dependency %q not declared in manifest\n", target)
				return 1
			}
			updateSet[sanitized] = struct{}{}
		}
	}

	lockPath := filepath.Join(filepath.Dir(manifest.Path), "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	lockCreated := false
	switch {
	case err == nil:
		if lock.Root != manifest.Name {
			fmt.Fprintf(os.Stderr, "lockfile root %q does not match manifest name %q\n", lock.Root, manifest.Name)
			return 1
		}
	case errors.Is(err, os.ErrNotExist):
		lock = driver.NewLockfile(manifest.Name, cliToolVersion)
		lock.Path = lockPath
		lockCreated = true
	default:
		fmt.Fprintf(os.Stderr, "failed to read lockfile: %v\n", err)
		return 1
	}

	if len(updateSet) == 0 {
		lock.Packages = nil
	} else {
		filtered := make([]*driver.LockedPackage, 0, len(lock.Packages))
		for _, pkg := range lock.Packages {
			if pkg == nil {
				continue
			}
			if _, ok := updateSet[sanitizeName(pkg.Name)]; ok {
				continue
			}
			filtered = append(filtered, pkg)
		}
		lock.Packages = filtered
	}

	installer := newDependencyInstaller(manifest, cacheDir)
	changed, logs, err := installer.Install(lock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to update dependencies: %v\n", err)
		return 1
	}
	for _, line := range logs {
		fmt.Fprintln(os.Stdout, line)
	}

	lock.Path = lockPath
	lock.Tool = cliToolVersion

	if changed || lockCreated {
		if err := driver.WriteLockfile(lock, lockPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write lockfile: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stdout, "Updated package.lock: %s\n", lock.Path)
	} else {
		fmt.Fprintln(os.Stdout, "Dependencies already up to date.")
	}
	return 0
}
