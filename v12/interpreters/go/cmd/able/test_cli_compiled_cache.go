package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"able/interpreter-go/pkg/compiler"
	"able/interpreter-go/pkg/driver"
)

const (
	compiledTestCacheDirEnv   = "ABLE_TEST_COMPILED_CACHE_DIR"
	compiledTestCacheSaltEnv  = "ABLE_TEST_COMPILED_CACHE_SALT"
	compiledTestCacheTraceEnv = "ABLE_TEST_COMPILED_CACHE_TRACE"
	compiledTestCacheSchema   = "able-compiled-test-v1"
)

type compiledTestCache struct {
	root      string
	tracePath string
}

type compiledTestCacheKeyInput struct {
	Program         *driver.Program
	EntryPath       string
	RunnerSource    string
	HarnessSource   string
	SearchPaths     []driver.SearchPath
	Packages        []string
	CompilerOptions compiler.Options
	ModuleRoot      string
	BuildIdentity   string
	Salt            string
}

type compiledTestCacheManifest struct {
	Schema       string `json:"schema"`
	Key          string `json:"key"`
	BinarySHA256 string `json:"binary_sha256"`
}

type compiledTestCacheValidation struct {
	BinaryPath string
	Manifest   compiledTestCacheManifest
	Reason     string
	Valid      bool
}

func openCompiledTestCache() (*compiledTestCache, error) {
	rawRoot := strings.TrimSpace(os.Getenv(compiledTestCacheDirEnv))
	if rawRoot == "" {
		return nil, nil
	}
	return openCompiledTestCacheAt(rawRoot)
}

func openCompiledTestCacheAt(rawRoot string) (*compiledTestCache, error) {
	root, err := filepath.Abs(rawRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", compiledTestCacheDirEnv, err)
	}
	root = filepath.Clean(root)
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create compiled-test cache: %w", err)
	}
	return &compiledTestCache{
		root:      root,
		tracePath: strings.TrimSpace(os.Getenv(compiledTestCacheTraceEnv)),
	}, nil
}

func compiledTestCacheKey(input compiledTestCacheKeyInput) (string, error) {
	if input.Program == nil || input.Program.Entry == nil {
		return "", fmt.Errorf("compiled-test cache: missing loaded program")
	}
	metadataDigest, err := compiledTestCacheMetadataDigest(input)
	if err != nil {
		return "", err
	}
	programDigest, err := compiledTestCacheProgramDigest(input)
	if err != nil {
		return "", err
	}
	digest := sha256.New()
	writeCompiledTestCacheField(digest, "schema", compiledTestCacheSchema)
	writeCompiledTestCacheField(digest, "metadata", metadataDigest)
	writeCompiledTestCacheField(digest, "program", programDigest)
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func compiledTestCacheMetadataDigest(input compiledTestCacheKeyInput) (string, error) {
	digest := sha256.New()
	writeCompiledTestCacheField(digest, "module-root", cleanAbsolutePath(input.ModuleRoot))
	writeCompiledTestCacheField(digest, "build-identity", input.BuildIdentity)
	writeCompiledTestCacheField(digest, "salt", input.Salt)
	writeCompiledTestCacheField(digest, "runner-source", input.RunnerSource)
	writeCompiledTestCacheField(digest, "harness-source", input.HarnessSource)

	options, err := json.Marshal(input.CompilerOptions)
	if err != nil {
		return "", fmt.Errorf("compiled-test cache: encode compiler options: %w", err)
	}
	writeCompiledTestCacheBytes(digest, "compiler-options", options)

	for _, searchPath := range input.SearchPaths {
		writeCompiledTestCacheField(digest, "search-kind", fmt.Sprintf("%d", searchPath.Kind))
		writeCompiledTestCacheField(digest, "search-stdlib-source", fmt.Sprintf("%d", searchPath.StdlibSource))
	}
	packages := append([]string(nil), input.Packages...)
	sort.Strings(packages)
	for _, pkg := range packages {
		writeCompiledTestCacheField(digest, "included-package", pkg)
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func compiledTestCacheProgramDigest(input compiledTestCacheKeyInput) (string, error) {
	digest := sha256.New()
	modules, err := compiledTestCacheModuleDigests(input)
	if err != nil {
		return "", err
	}
	for _, module := range modules {
		writeCompiledTestCacheField(digest, "module-package", module.pkg)
		writeCompiledTestCacheField(digest, "module-digest", module.digest)
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

type compiledTestModuleDigest struct {
	pkg    string
	digest string
}

func compiledTestCacheModuleDigests(input compiledTestCacheKeyInput) ([]compiledTestModuleDigest, error) {
	entryPath := cleanAbsolutePath(input.EntryPath)
	modules := make([]compiledTestModuleDigest, 0, len(input.Program.Modules))
	for _, module := range input.Program.Modules {
		moduleDigest := sha256.New()
		if module == nil {
			writeCompiledTestCacheField(moduleDigest, "module", "<nil>")
			modules = append(modules, compiledTestModuleDigest{
				pkg:    "<nil>",
				digest: hex.EncodeToString(moduleDigest.Sum(nil)),
			})
			continue
		}
		writeCompiledTestCacheField(moduleDigest, "module", module.Package)
		writeCompiledTestCacheStrings(moduleDigest, "module-import", module.Imports)
		writeCompiledTestCacheStrings(moduleDigest, "module-dyn-import", module.DynImports)
		files := compiledTestCacheSourceFiles(module.Files)
		for _, source := range files {
			if source.actualPath == entryPath {
				writeCompiledTestCacheField(moduleDigest, "source-path", "@compiled-test-runner")
				writeCompiledTestCacheField(moduleDigest, "source-content", input.RunnerSource)
				continue
			}
			content, err := os.ReadFile(source.actualPath)
			if err != nil {
				return nil, fmt.Errorf("compiled-test cache: read source %s: %w", source.actualPath, err)
			}
			writeCompiledTestCacheField(moduleDigest, "source-path", source.keyPath)
			writeCompiledTestCacheBytes(moduleDigest, "source-content", content)
		}
		modules = append(modules, compiledTestModuleDigest{
			pkg:    module.Package,
			digest: hex.EncodeToString(moduleDigest.Sum(nil)),
		})
	}
	sort.Slice(modules, func(i, j int) bool {
		if modules[i].pkg != modules[j].pkg {
			return modules[i].pkg < modules[j].pkg
		}
		return modules[i].digest < modules[j].digest
	})
	return modules, nil
}

type compiledTestCacheSource struct {
	actualPath string
	keyPath    string
}

func compiledTestCacheSourceFiles(sourcePaths []string) []compiledTestCacheSource {
	files := make([]compiledTestCacheSource, 0, len(sourcePaths))
	for _, sourcePath := range sourcePaths {
		actualPath := cleanAbsolutePath(sourcePath)
		files = append(files, compiledTestCacheSource{
			actualPath: actualPath,
			keyPath:    filepath.ToSlash(actualPath),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].keyPath != files[j].keyPath {
			return files[i].keyPath < files[j].keyPath
		}
		return files[i].actualPath < files[j].actualPath
	})
	return files
}

func normalizeCompiledTestRunnerOrigins(program *driver.Program, entryPath string) {
	if program == nil {
		return
	}
	cleanEntryPath := cleanAbsolutePath(entryPath)
	modules := append([]*driver.Module{program.Entry}, program.Modules...)
	for _, module := range modules {
		if module == nil {
			continue
		}
		for node, origin := range module.NodeOrigins {
			if cleanAbsolutePath(origin) == cleanEntryPath {
				module.NodeOrigins[node] = "compiled-test-runner/runner.able"
			}
		}
	}
}

func compiledTestBuildIdentity(moduleRoot string) (string, error) {
	digest := sha256.New()
	goExecutable, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf("compiled-test cache: locate Go toolchain: %w", err)
	}
	if err := writeCompiledTestCacheFile(digest, "go-executable", goExecutable); err != nil {
		return "", err
	}
	goEnv := exec.Command("go", "env", "-json",
		"GOOS", "GOARCH", "GOVERSION", "CGO_ENABLED", "GOFLAGS",
		"GOEXPERIMENT", "GOTOOLCHAIN", "CC", "CXX",
	)
	goEnv.Dir = moduleRoot
	output, err := goEnv.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("compiled-test cache: identify Go toolchain: %w\n%s", err, output)
	}
	writeCompiledTestCacheBytes(digest, "go-env", output)

	buildFiles, err := compiledTestModuleBuildFiles(moduleRoot)
	if err != nil {
		return "", err
	}
	for _, sourcePath := range buildFiles {
		if err := writeCompiledTestCacheFile(digest, "module-build-source", sourcePath); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func compiledTestModuleBuildFiles(moduleRoot string) ([]string, error) {
	root, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("compiled-test cache: resolve module root: %w", err)
	}
	var files []string
	for _, name := range []string{"go.mod", "go.sum"} {
		path := filepath.Join(root, name)
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			files = append(files, path)
		} else if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("compiled-test cache: stat %s: %w", path, err)
		}
	}
	for _, relRoot := range []string{"cmd", "pkg"} {
		walkRoot := filepath.Join(root, relRoot)
		err := filepath.WalkDir(walkRoot, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			name := entry.Name()
			extension := filepath.Ext(name)
			include := extension == ".go" || extension == ".c" || extension == ".h" || extension == ".s"
			if extension == ".go" && strings.HasSuffix(name, "_test.go") {
				include = false
			}
			if path == filepath.Join(root, "cmd", "able", "embedded", "kernel", "src", "kernel.able") ||
				path == filepath.Join(root, "cmd", "able", "embedded", "kernel", "package.yml") {
				include = true
			}
			if include {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("compiled-test cache: inspect module build sources: %w", err)
		}
	}
	sort.Strings(files)
	return files, nil
}

func (cache *compiledTestCache) lookup(key string) (string, bool) {
	if cache == nil || key == "" {
		return "", false
	}
	validation := cache.validateEntry(key)
	if !validation.Valid {
		return "", false
	}
	return validation.BinaryPath, true
}

func (cache *compiledTestCache) validateEntry(key string) compiledTestCacheValidation {
	entryDir := filepath.Join(cache.root, compiledTestCacheSchema, key)
	manifestBytes, err := os.ReadFile(filepath.Join(entryDir, "manifest.json"))
	if err != nil {
		return compiledTestCacheValidation{Reason: fmt.Sprintf("read manifest: %v", err)}
	}
	var manifest compiledTestCacheManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return compiledTestCacheValidation{Reason: fmt.Sprintf("decode manifest: %v", err)}
	}
	if manifest.Schema != compiledTestCacheSchema || manifest.Key != key || manifest.BinarySHA256 == "" {
		return compiledTestCacheValidation{
			Manifest: manifest,
			Reason:   "manifest schema, key, or binary checksum does not match entry",
		}
	}
	binaryPath := filepath.Join(entryDir, "able-test")
	info, err := os.Stat(binaryPath)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o100 == 0 {
		return compiledTestCacheValidation{
			Manifest: manifest,
			Reason:   "cached executable is missing, non-regular, or non-executable",
		}
	}
	digest, err := compiledTestFileSHA256(binaryPath)
	if err != nil || digest != manifest.BinarySHA256 {
		return compiledTestCacheValidation{
			Manifest: manifest,
			Reason:   "cached executable checksum does not match manifest",
		}
	}
	return compiledTestCacheValidation{
		BinaryPath: binaryPath,
		Manifest:   manifest,
		Valid:      true,
	}
}

func (cache *compiledTestCache) markUsed(key string) error {
	if cache == nil || key == "" {
		return nil
	}
	now := time.Now()
	entryDir := filepath.Join(cache.root, compiledTestCacheSchema, key)
	if err := os.Chtimes(entryDir, now, now); err != nil {
		return fmt.Errorf("compiled-test cache: update entry usage time: %w", err)
	}
	return nil
}

func (cache *compiledTestCache) publish(key, sourceBinary string) (string, error) {
	if cache == nil {
		return "", nil
	}
	if existing, ok := cache.lookup(key); ok {
		if err := cache.markUsed(key); err != nil {
			return "", err
		}
		return existing, nil
	}
	schemaRoot := filepath.Join(cache.root, compiledTestCacheSchema)
	if err := os.MkdirAll(schemaRoot, 0o700); err != nil {
		return "", fmt.Errorf("compiled-test cache: create schema directory: %w", err)
	}
	stagingDir, err := os.MkdirTemp(schemaRoot, ".publish-")
	if err != nil {
		return "", fmt.Errorf("compiled-test cache: create staging directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(stagingDir) }()

	stagedBinary := filepath.Join(stagingDir, "able-test")
	binaryDigest, err := copyCompiledTestCacheBinary(stagedBinary, sourceBinary)
	if err != nil {
		return "", err
	}
	manifest := compiledTestCacheManifest{
		Schema:       compiledTestCacheSchema,
		Key:          key,
		BinarySHA256: binaryDigest,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("compiled-test cache: encode manifest: %w", err)
	}
	manifestBytes = append(manifestBytes, '\n')
	if err := os.WriteFile(filepath.Join(stagingDir, "manifest.json"), manifestBytes, 0o600); err != nil {
		return "", fmt.Errorf("compiled-test cache: write manifest: %w", err)
	}
	entryDir := filepath.Join(schemaRoot, key)
	if err := os.Rename(stagingDir, entryDir); err != nil {
		if existing, ok := cache.lookup(key); ok {
			if touchErr := cache.markUsed(key); touchErr != nil {
				return "", touchErr
			}
			return existing, nil
		}
		if removeErr := os.RemoveAll(entryDir); removeErr != nil {
			return "", fmt.Errorf("compiled-test cache: remove invalid entry: %w", removeErr)
		}
		if retryErr := os.Rename(stagingDir, entryDir); retryErr != nil {
			if existing, ok := cache.lookup(key); ok {
				if touchErr := cache.markUsed(key); touchErr != nil {
					return "", touchErr
				}
				return existing, nil
			}
			return "", fmt.Errorf("compiled-test cache: publish entry: %w", retryErr)
		}
	}
	return filepath.Join(entryDir, "able-test"), nil
}

func (cache *compiledTestCache) trace(status, key string) {
	if cache == nil || cache.tracePath == "" {
		return
	}
	tracePath, err := filepath.Abs(cache.tracePath)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(tracePath), 0o700); err != nil {
		return
	}
	file, err := os.OpenFile(tracePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(file, "%s %s\n", status, key)
	_ = file.Close()
}

func copyCompiledTestCacheBinary(destination, source string) (string, error) {
	input, err := os.Open(source)
	if err != nil {
		return "", fmt.Errorf("compiled-test cache: open built binary: %w", err)
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		return "", fmt.Errorf("compiled-test cache: create cached binary: %w", err)
	}
	digest := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(output, digest), input)
	syncErr := output.Sync()
	closeErr := output.Close()
	if copyErr != nil {
		return "", fmt.Errorf("compiled-test cache: copy built binary: %w", copyErr)
	}
	if syncErr != nil {
		return "", fmt.Errorf("compiled-test cache: sync built binary: %w", syncErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("compiled-test cache: close built binary: %w", closeErr)
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func compiledTestFileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func writeCompiledTestCacheFile(digest hash.Hash, label, path string) error {
	cleanPath := cleanAbsolutePath(path)
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("compiled-test cache: read identity file %s: %w", cleanPath, err)
	}
	writeCompiledTestCacheField(digest, label+"-path", cleanPath)
	writeCompiledTestCacheBytes(digest, label+"-content", content)
	return nil
}

func writeCompiledTestCacheStrings(digest hash.Hash, label string, values []string) {
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	for _, value := range sorted {
		writeCompiledTestCacheField(digest, label, value)
	}
}

func writeCompiledTestCacheField(digest hash.Hash, label, value string) {
	writeCompiledTestCacheBytes(digest, label, []byte(value))
}

func writeCompiledTestCacheBytes(digest hash.Hash, label string, value []byte) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(label)))
	_, _ = digest.Write(size[:])
	_, _ = digest.Write([]byte(label))
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = digest.Write(size[:])
	_, _ = digest.Write(value)
}

func cleanAbsolutePath(path string) string {
	if absolute, err := filepath.Abs(path); err == nil {
		return filepath.Clean(absolute)
	}
	return filepath.Clean(path)
}
