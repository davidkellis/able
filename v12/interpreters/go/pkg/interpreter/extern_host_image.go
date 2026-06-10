//go:build !(js && wasm)

package interpreter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"sort"
	"strings"
	"time"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

type externHostImageEntry struct {
	packageName string
	packageKey  string
	state       *externTargetState
	hash        string
}

// prepareExternHostImageForProgram registers every Go extern declaration in a
// dependency-ordered program, then loads one image for that complete known
// set. Individual package plugins remain the fallback for dynamic definitions
// that appear after this preparation step.
func (i *Interpreter) prepareExternHostImageForProgram(program *driver.Program) (int, error) {
	if program == nil {
		return 0, fmt.Errorf("interpreter: nil extern host program")
	}
	previousPackage := i.currentPackage
	defer func() { i.currentPackage = previousPackage }()
	for _, module := range program.Modules {
		if module == nil || module.AST == nil || module.Package == "" {
			continue
		}
		i.currentPackage = module.Package
		i.registerExternStatements(module.AST)
	}
	return i.prepareExternHostImageForPackageNames(programPackageNames(program.Modules))
}

func programPackageNames(modules []*driver.Module) []string {
	seen := make(map[string]struct{}, len(modules))
	for _, module := range modules {
		if module != nil && module.AST != nil && module.Package != "" {
			seen[module.Package] = struct{}{}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (i *Interpreter) prepareExternHostImageForPackageNames(packageNames []string) (int, error) {
	i.externHostMu.Lock()
	defer i.externHostMu.Unlock()

	entries := make([]externHostImageEntry, 0, len(packageNames))
	for _, packageName := range packageNames {
		pkg := i.externHostPackages[packageName]
		if pkg == nil {
			continue
		}
		state := pkg.targets[ast.HostTargetGo]
		if !hasPrewarmableGoExtern(i, state) {
			continue
		}
		entries = append(entries, externHostImageEntry{
			packageName: packageName,
			packageKey:  fmt.Sprintf("p%d", len(entries)),
			state:       state,
			hash:        cachedExternStateHash(ast.HostTargetGo, state, externHostCacheScope()),
		})
	}
	if len(entries) == 0 {
		return 0, nil
	}

	modules, err := buildExternHostImage(entries)
	if err != nil {
		return 0, err
	}
	for _, entry := range entries {
		pkg := i.externHostPackages[entry.packageName]
		if pkg == nil {
			continue
		}
		pkg.modules[ast.HostTargetGo] = modules[entry.packageName]
	}
	return 1, nil
}

func buildExternHostImage(entries []externHostImageEntry) (map[string]*externHostModule, error) {
	imageHash := externHostImageHash(entries)
	cacheDir := filepath.Join(externHostCacheRoot(), "image", imageHash)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("extern image cache mkdir: %w", err)
	}
	moduleName := "able_extern_image_" + imageHash
	if err := writeExternHostImageSources(cacheDir, moduleName, entries); err != nil {
		return nil, err
	}
	pluginPath := externPluginArtifactPath(cacheDir)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		if err := buildExternPlugin(cacheDir, pluginPath); err != nil {
			return nil, err
		}
	}
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		fallbackPath := filepath.Join(cacheDir, fmt.Sprintf("extern-rebuilt-%d.so", time.Now().UnixNano()))
		if buildErr := buildExternPlugin(cacheDir, fallbackPath); buildErr != nil {
			return nil, fmt.Errorf("extern host image plugin open: %w; rebuild failed: %v", err, buildErr)
		}
		plug, err = plugin.Open(fallbackPath)
		if err != nil {
			return nil, fmt.Errorf("extern host rebuilt image plugin open: %w", err)
		}
		_ = writeExternPluginArtifactPath(cacheDir, filepath.Base(fallbackPath))
	}

	modules := make(map[string]*externHostModule, len(entries))
	for _, entry := range entries {
		modules[entry.packageName] = &externHostModule{
			hash:            entry.hash,
			plugin:          plug,
			symbols:         make(map[string]reflect.Value),
			imagePackageKey: entry.packageKey,
		}
	}
	return modules, nil
}

func externHostImageHash(entries []externHostImageEntry) string {
	hasher := sha256.New()
	hasher.Write([]byte("extern-host-image-v1\n"))
	for _, entry := range entries {
		hasher.Write([]byte(entry.packageName))
		hasher.Write([]byte("\n"))
		hasher.Write([]byte(entry.hash))
		hasher.Write([]byte("\n"))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func writeExternHostImageSources(cacheDir, moduleName string, entries []externHostImageEntry) error {
	modulePath := filepath.Join(cacheDir, "go.mod")
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		if err := os.WriteFile(modulePath, []byte(fmt.Sprintf("module %s\n\ngo 1.22\n", moduleName)), 0o644); err != nil {
			return fmt.Errorf("extern image write go.mod: %w", err)
		}
	}
	for _, entry := range entries {
		packageDir := filepath.Join(cacheDir, entry.packageKey)
		if err := os.MkdirAll(packageDir, 0o755); err != nil {
			return fmt.Errorf("extern image package mkdir: %w", err)
		}
		sourcePath := filepath.Join(packageDir, "extern.go")
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			source, renderErr := renderGoHostPackage(entry.packageKey, entry.state)
			if renderErr != nil {
				return renderErr
			}
			if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
				return fmt.Errorf("extern image write package: %w", err)
			}
		}
	}
	mainPath := filepath.Join(cacheDir, "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		source, renderErr := renderGoHostImageMain(moduleName, entries)
		if renderErr != nil {
			return renderErr
		}
		if err := os.WriteFile(mainPath, []byte(source), 0o644); err != nil {
			return fmt.Errorf("extern image write main: %w", err)
		}
	}
	return nil
}

func renderGoHostPackage(packageKey string, state *externTargetState) (string, error) {
	var builder strings.Builder
	builder.WriteString("package ")
	builder.WriteString(packageKey)
	builder.WriteString("\n\n")
	writeGoHostPreamble(&builder, state)
	for _, extern := range state.externs {
		if extern == nil || extern.Signature == nil || extern.Signature.ID == nil {
			continue
		}
		fn, err := renderGoExternFunction(extern)
		if err != nil {
			return "", err
		}
		builder.WriteString(fn)
		builder.WriteString("\n\n")
	}
	return builder.String(), nil
}

func renderGoHostImageMain(moduleName string, entries []externHostImageEntry) (string, error) {
	var builder strings.Builder
	builder.WriteString("package main\n\n")
	builder.WriteString("import (\n")
	if externHostImageNeedsBigInt(entries) {
		builder.WriteString("\t\"math/big\"\n")
	}
	for _, entry := range entries {
		builder.WriteString("\t")
		builder.WriteString(entry.packageKey)
		builder.WriteString(" \"")
		builder.WriteString(moduleName)
		builder.WriteString("/")
		builder.WriteString(entry.packageKey)
		builder.WriteString("\"\n")
	}
	builder.WriteString(")\n\n")
	builder.WriteString("type IoHandle = interface{}\n")
	builder.WriteString("type ProcHandle = interface{}\n\n")

	for _, entry := range entries {
		for _, extern := range entry.state.externs {
			if extern == nil || extern.Signature == nil || extern.Signature.ID == nil {
				continue
			}
			wrapper, err := renderGoHostImageWrapper(entry.packageKey, extern)
			if err != nil {
				return "", err
			}
			builder.WriteString(wrapper)
			builder.WriteString("\n\n")
		}
	}
	return builder.String(), nil
}

func externHostImageNeedsBigInt(entries []externHostImageEntry) bool {
	for _, entry := range entries {
		if needsBigInt(entry.state) {
			return true
		}
	}
	return false
}

func renderGoHostImageWrapper(packageKey string, extern *ast.ExternFunctionBody) (string, error) {
	name := extern.Signature.ID.Name
	params := make([]string, 0, len(extern.Signature.Params))
	argNames := make([]string, 0, len(extern.Signature.Params))
	for idx, param := range extern.Signature.Params {
		paramName := externParamName(param, idx)
		typ, err := goTypeForExpr(param.ParamType)
		if err != nil {
			return "", err
		}
		params = append(params, fmt.Sprintf("%s %s", paramName, typ))
		argNames = append(argNames, paramName)
	}
	retType, err := goReturnTypeForExpr(extern.Signature.ReturnType)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("func ")
	builder.WriteString(externImageSymbolName(packageKey, name))
	builder.WriteString("(")
	builder.WriteString(strings.Join(params, ", "))
	builder.WriteString(")")
	if retType != "" {
		builder.WriteString(" ")
		builder.WriteString(retType)
	}
	builder.WriteString(" {\n\t")
	if retType != "" {
		builder.WriteString("return ")
	}
	builder.WriteString(packageKey)
	builder.WriteString(".")
	builder.WriteString(externSymbolName(name))
	builder.WriteString("(")
	builder.WriteString(strings.Join(argNames, ", "))
	builder.WriteString(")\n}")
	return builder.String(), nil
}

func writeGoHostPreamble(builder *strings.Builder, state *externTargetState) {
	if state != nil {
		for _, prelude := range state.preludes {
			builder.WriteString(prelude)
			builder.WriteString("\n")
		}
	}
	if needsBigInt(state) {
		builder.WriteString("import \"math/big\"\n\n")
	}
	builder.WriteString("type IoHandle = interface{}\n")
	builder.WriteString("type ProcHandle = interface{}\n")
	builder.WriteString("type hostError struct{ message string }\n")
	builder.WriteString("func (e hostError) Error() string { return e.message }\n")
	builder.WriteString("func host_error[T any](message string) (T, error) { var zero T; return zero, hostError{message} }\n\n")
	builder.WriteString("func able_borrowed_bytes(data []byte) []byte { return data }\n\n")
}
