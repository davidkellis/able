package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/compiler"
	"able/interpreter-go/pkg/driver"
)

type buildConfig struct {
	OutputDir string
	BinPath   string
	WithTests bool
	ShowHelp  bool
}

func runBuild(args []string) int {
	config, remaining, err := parseBuildArguments(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able build: %v\n", err)
		return 1
	}
	if config.ShowHelp {
		printBuildUsage()
		return 0
	}
	if len(remaining) > 1 {
		fmt.Fprintf(os.Stderr, "able build expects at most one target or entry file (received %s)\n", strings.Join(remaining, " "))
		return 1
	}

	manifest, lock, entryPath, targetName, err := resolveBuildEntry(remaining)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able build: %v\n", err)
		return 1
	}

	entryAbs, err := filepath.Abs(entryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able build: resolve entry path: %v\n", err)
		return 1
	}

	extras, err := buildExecutionSearchPaths(manifest, lock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able build: failed to prepare build environment: %v\n", err)
		return 1
	}
	searchPaths := collectSearchPaths(filepath.Dir(entryAbs), extras...)

	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able build: failed to initialize loader: %v\n", err)
		return 1
	}
	defer loader.Close()

	program, err := loader.LoadWithOptions(entryAbs, driver.LoadOptions{IncludeTests: config.WithTests})
	if err != nil {
		var parseErr *driver.ParserDiagnosticError
		if errors.As(err, &parseErr) {
			fmt.Fprintln(os.Stderr, driver.DescribeParserDiagnostic(parseErr.Diagnostic))
			return 1
		}
		fmt.Fprintf(os.Stderr, "able build: failed to load program: %v\n", err)
		return 1
	}

	outputDir := config.OutputDir
	if outputDir == "" {
		outputDir = defaultBuildOutputDir(manifest, entryAbs, targetName, config.WithTests)
	}
	if outputDir, err = filepath.Abs(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "able build: resolve output dir: %v\n", err)
		return 1
	}

	comp := compiler.New(compiler.Options{
		PackageName: "main",
		EmitMain:    true,
		EntryPath:   entryAbs,
	})
	result, err := comp.Compile(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able build: compile failed: %v\n", err)
		return 1
	}
	for _, warning := range result.Warnings {
		fmt.Fprintln(os.Stderr, warning)
	}
	if err := result.Write(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "able build: write output: %v\n", err)
		return 1
	}
	if err := prepareBuildModule(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "able build: %v\n", err)
		return 1
	}

	binPath := config.BinPath
	if binPath == "" {
		binPath = filepath.Join(outputDir, defaultBuildBinaryName(manifest, targetName, entryAbs))
	}
	if binPath, err = filepath.Abs(binPath); err != nil {
		fmt.Fprintf(os.Stderr, "able build: resolve binary path: %v\n", err)
		return 1
	}

	cmd := exec.Command("go", "build", "-mod=mod", "-o", binPath, ".")
	cmd.Dir = outputDir
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "able build: go build failed: %v\n%s\n", err, string(output))
		return 1
	}

	fmt.Fprintf(os.Stdout, "built %s\n", binPath)
	return 0
}

func parseBuildArguments(args []string) (buildConfig, []string, error) {
	config := buildConfig{}
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h":
			config.ShowHelp = true
		case arg == "--out" || arg == "-o":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return buildConfig{}, nil, err
			}
			config.OutputDir = val
		case strings.HasPrefix(arg, "--out="):
			config.OutputDir = strings.TrimPrefix(arg, "--out=")
		case arg == "--with-tests":
			config.WithTests = true
		case arg == "--bin":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return buildConfig{}, nil, err
			}
			config.BinPath = val
		case strings.HasPrefix(arg, "--bin="):
			config.BinPath = strings.TrimPrefix(arg, "--bin=")
		case arg == "--":
			remaining = append(remaining, args[i+1:]...)
			return config, remaining, nil
		default:
			remaining = append(remaining, arg)
		}
	}
	return config, remaining, nil
}

func resolveBuildEntry(args []string) (*driver.Manifest, *driver.Lockfile, string, string, error) {
	var candidate string
	if len(args) > 0 {
		candidate = strings.TrimSpace(args[0])
	}

	if candidate == "" {
		manifest, err := loadManifestFrom(".")
		if err != nil {
			if errors.Is(err, errManifestNotFound) {
				return nil, nil, "", "", fmt.Errorf("no package.yml found; specify a target or entry file")
			}
			return nil, nil, "", "", err
		}
		lock, err := loadLockfileForManifest(manifest)
		if err != nil {
			return nil, nil, "", "", err
		}
		target, err := manifest.DefaultTarget()
		if err != nil {
			return nil, nil, "", "", err
		}
		entry, err := resolveTargetMain(manifest, target)
		if err != nil {
			return nil, nil, "", "", err
		}
		return manifest, lock, entry, target.OriginalName, nil
	}

	manifest, manifestErr := loadManifestFrom(".")
	if manifestErr != nil && !errors.Is(manifestErr, errManifestNotFound) {
		return nil, nil, "", "", manifestErr
	}
	if manifest != nil {
		if target, ok := manifest.FindTarget(candidate); ok && target != nil {
			lock, err := loadLockfileForManifest(manifest)
			if err != nil {
				return nil, nil, "", "", err
			}
			entry, err := resolveTargetMain(manifest, target)
			if err != nil {
				return nil, nil, "", "", err
			}
			return manifest, lock, entry, target.OriginalName, nil
		}
		if !looksLikePathCandidate(candidate) {
			return nil, nil, "", "", fmt.Errorf("unknown build target %q", candidate)
		}
	} else if !looksLikePathCandidate(candidate) {
		return nil, nil, "", "", fmt.Errorf("no package.yml found; specify a target or entry file")
	}

	activeManifest := manifest
	if absCandidate, err := filepath.Abs(candidate); err == nil {
		entryDir := filepath.Dir(absCandidate)
		if manifestPath, findErr := findManifest(entryDir); findErr == nil {
			if activeManifest == nil || filepath.Clean(activeManifest.Path) != filepath.Clean(manifestPath) {
				loaded, loadErr := driver.LoadManifest(manifestPath)
				if loadErr != nil {
					return nil, nil, "", "", loadErr
				}
				activeManifest = loaded
			}
		} else if !errors.Is(findErr, errManifestNotFound) {
			return nil, nil, "", "", findErr
		}
	}

	lock, err := loadLockfileForManifest(activeManifest)
	if err != nil {
		return nil, nil, "", "", err
	}

	return activeManifest, lock, candidate, "", nil
}

func defaultBuildOutputDir(manifest *driver.Manifest, entryPath string, targetName string, withTests bool) string {
	base := ""
	if manifest != nil && manifest.Path != "" {
		base = filepath.Dir(manifest.Path)
	} else {
		base = filepath.Dir(entryPath)
	}
	subdir := filepath.Join("target", "compiled")
	if withTests {
		subdir = filepath.Join("target", "test", "compiled")
	}
	if targetName != "" {
		subdir = filepath.Join(subdir, sanitizePathSegment(targetName))
	}
	return filepath.Join(base, subdir)
}

func defaultBuildBinaryName(manifest *driver.Manifest, targetName string, entryPath string) string {
	if targetName != "" {
		return sanitizePathSegment(targetName)
	}
	if manifest != nil && manifest.Name != "" {
		return sanitizePathSegment(manifest.Name)
	}
	base := strings.TrimSuffix(filepath.Base(entryPath), filepath.Ext(entryPath))
	if base == "" {
		return "compiled"
	}
	return sanitizePathSegment(base)
}

func printBuildUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  able build [target]")
	fmt.Fprintln(os.Stderr, "  able build <file.able>")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  -o, --out <dir>   output directory for generated Go code")
	fmt.Fprintln(os.Stderr, "      --bin <path>  output path for the compiled binary")
	fmt.Fprintln(os.Stderr, "      --with-tests  include test modules in the build")
}
