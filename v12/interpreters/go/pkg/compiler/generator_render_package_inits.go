package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type compiledPackageInitImportSpec struct {
	Token string
	Path  string
	Alias string
}

func (g *generator) packageInitFuncName(pkgName string, idx int) string {
	trimmed := strings.TrimSpace(pkgName)
	if trimmed == "" {
		return fmt.Sprintf("__able_run_compiled_package_init_%d", idx)
	}
	return fmt.Sprintf("__able_run_compiled_package_init_%s_%d", sanitizeIdent(trimmed), idx)
}

func (g *generator) renderCompiledPackageInitFile() ([]byte, error) {
	if g == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)
	importSet := map[string]struct{}{
		"able/interpreter-go/pkg/compiler/bridge": {},
		"able/interpreter-go/pkg/interpreter":     {},
		"able/interpreter-go/pkg/runtime":         {},
	}
	if g.hasCompiledPackageInitBodies() {
		importSet["fmt"] = struct{}{}
	}
	for _, spec := range g.compiledPackageInitImports() {
		importSet[spec.Path] = struct{}{}
	}
	imports := make([]string, 0, len(importSet))
	for imp := range importSet {
		imports = append(imports, imp)
	}
	sort.Strings(imports)
	fmt.Fprintf(&buf, "import (\n")
	for _, imp := range imports {
		if spec, ok := compiledPackageInitImportSpecForPath(imp); ok {
			fmt.Fprintf(&buf, "\t%s %q\n", spec.Alias, imp)
			continue
		}
		fmt.Fprintf(&buf, "\t%q\n", imp)
	}
	fmt.Fprintf(&buf, ")\n\n")
	fmt.Fprintf(&buf, "func __able_run_compiled_package_inits(rt *bridge.Runtime, interp *interpreter.Interpreter, entryEnv *runtime.Environment, __able_bootstrapped_metadata bool) error {\n")
	fmt.Fprintf(&buf, "\t_ = interp\n")
	fmt.Fprintf(&buf, "\t_ = entryEnv\n")
	fmt.Fprintf(&buf, "\tif __able_bootstrapped_metadata {\n")
	fmt.Fprintf(&buf, "\t\treturn nil\n")
	fmt.Fprintf(&buf, "\t}\n")
	for idx, pkgName := range g.packageInitOrder {
		if _, ok := g.packageEnvVar(pkgName); !ok {
			continue
		}
		fmt.Fprintf(&buf, "\tif err := %s(rt); err != nil {\n", g.packageInitFuncName(pkgName, idx))
		fmt.Fprintf(&buf, "\t\treturn err\n")
		fmt.Fprintf(&buf, "\t}\n")
	}
	fmt.Fprintf(&buf, "\treturn nil\n")
	fmt.Fprintf(&buf, "}\n\n")
	for idx, pkgName := range g.packageInitOrder {
		envVar, ok := g.packageEnvVar(pkgName)
		if !ok {
			continue
		}
		lines := g.packageInitCompiled[pkgName]
		if len(lines) == 0 {
			continue
		}
		if err := g.renderCompiledPackageInitFunc(&buf, pkgName, idx, envVar, lines); err != nil {
			return nil, err
		}
	}
	return formatSource(buf.Bytes())
}

func (g *generator) compiledPackageInitImports() []compiledPackageInitImportSpec {
	if g == nil {
		return nil
	}
	needed := map[string]compiledPackageInitImportSpec{}
	for _, pkgName := range g.packageInitOrder {
		for _, line := range g.packageInitCompiled[pkgName] {
			for _, spec := range compiledPackageInitImportSpecs {
				if compiledPackageInitUsesQualifier(line, spec.Token) {
					needed[spec.Path] = spec
				}
			}
		}
	}
	imports := make([]compiledPackageInitImportSpec, 0, len(needed))
	for _, spec := range needed {
		imports = append(imports, spec)
	}
	sort.Slice(imports, func(i, j int) bool { return imports[i].Path < imports[j].Path })
	return imports
}

var compiledPackageInitImportSpecs = []compiledPackageInitImportSpec{
	{Token: "ast.", Path: "able/interpreter-go/pkg/ast", Alias: "ast"},
	{Token: "big.", Path: "math/big", Alias: "__able_big"},
	{Token: "bits.", Path: "math/bits", Alias: "__able_bits"},
	{Token: "context.", Path: "context", Alias: "__able_context"},
	{Token: "math.", Path: "math", Alias: "__able_math"},
	{Token: "sort.", Path: "sort", Alias: "__able_sort"},
	{Token: "strconv.", Path: "strconv", Alias: "__able_strconv"},
	{Token: "strings.", Path: "strings", Alias: "__able_strings"},
	{Token: "sync.", Path: "sync", Alias: "__able_sync"},
	{Token: "atomic.", Path: "sync/atomic", Alias: "__able_atomic"},
	{Token: "time.", Path: "time", Alias: "__able_time"},
	{Token: "utf8.", Path: "unicode/utf8", Alias: "__able_utf8"},
}

func compiledPackageInitImportSpecForPath(path string) (compiledPackageInitImportSpec, bool) {
	for _, spec := range compiledPackageInitImportSpecs {
		if spec.Path == path {
			return spec, true
		}
	}
	return compiledPackageInitImportSpec{}, false
}

func compiledPackageInitUsesQualifier(line string, token string) bool {
	idx := 0
	for idx < len(line) {
		rel := strings.Index(line[idx:], token)
		if rel < 0 {
			return false
		}
		pos := idx + rel
		if pos == 0 || !isCompiledPackageInitIdentChar(line[pos-1]) {
			return true
		}
		idx = pos + len(token)
	}
	return false
}

func rewriteCompiledPackageInitQualifiers(line string) string {
	out := line
	for _, spec := range compiledPackageInitImportSpecs {
		out = rewriteCompiledPackageInitQualifier(out, spec.Token, spec.Alias+".")
	}
	return out
}

func rewriteCompiledPackageInitQualifier(line string, token string, replacement string) string {
	idx := 0
	for idx < len(line) {
		rel := strings.Index(line[idx:], token)
		if rel < 0 {
			break
		}
		pos := idx + rel
		if pos > 0 && isCompiledPackageInitIdentChar(line[pos-1]) {
			idx = pos + len(token)
			continue
		}
		line = line[:pos] + replacement + line[pos+len(token):]
		idx = pos + len(replacement)
	}
	return line
}

func isCompiledPackageInitIdentChar(b byte) bool {
	return b == '_' ||
		(b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9')
}

func (g *generator) hasCompiledPackageInitBodies() bool {
	if g == nil {
		return false
	}
	for _, pkgName := range g.packageInitOrder {
		if len(g.packageInitCompiled[pkgName]) > 0 {
			return true
		}
	}
	return false
}

func (g *generator) renderCompiledPackageInitFunc(buf *bytes.Buffer, pkgName string, idx int, envVar string, lines []string) error {
	fnName := g.packageInitFuncName(pkgName, idx)
	fmt.Fprintf(buf, "func %s(rt *bridge.Runtime) error {\n", fnName)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"compiler: missing runtime\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif %s == nil {\n", envVar)
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"compiler: missing package environment for %s\")\n", pkgName)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\t__able_runtime = rt\n")
	fmt.Fprintf(buf, "\trt.SetEnv(%s)\n", envVar)
	writeRuntimeEnvSwapIfNeeded(buf, "\t", "rt", envVar, "")
	for _, line := range lines {
		fmt.Fprintf(buf, "\t%s\n", rewriteCompiledPackageInitQualifiers(line))
	}
	fmt.Fprintf(buf, "\treturn nil\n")
	fmt.Fprintf(buf, "}\n\n")
	return nil
}

func (g *generator) preparePackageInitBodies() error {
	if g == nil || len(g.packageInitOrder) == 0 {
		return nil
	}
	g.packageInitCompiled = make(map[string][]string, len(g.packageInitOrder))
	for _, pkgName := range g.packageInitOrder {
		stmts := g.packageInitStatements[pkgName]
		if len(stmts) == 0 {
			continue
		}
		ctx := newCompileContext(g, nil, g.functionsForPackage(pkgName), g.overloadsForPackage(pkgName), pkgName, nil)
		ctx.controlMode = compileControlModeErrorOnly
		ctx.breakpointGoLabels = make(map[string]string)
		ctx.breakpointResultTemps = make(map[string]string)
		ctx.breakpointResultTypes = make(map[string]string)
		ctx.breakpointResultProbes = make(map[string]*controlFlowResultProbe)
		lines := make([]string, 0)
		for _, stmt := range stmts {
			stmtLines, ok := g.compileStatement(ctx, stmt)
			if !ok {
				reason := ctx.reason
				if reason == "" {
					reason = "unsupported package init statement"
				}
				return fmt.Errorf("compiler: package init %s: %s", pkgName, reason)
			}
			lines = append(lines, stmtLines...)
		}
		g.packageInitCompiled[pkgName] = lines
	}
	return nil
}
