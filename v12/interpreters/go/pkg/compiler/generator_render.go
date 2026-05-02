package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func (g *generator) render() (map[string][]byte, error) {
	g.resolveStaticFunctionIntegerFacts()
	g.resolveStaticFunctionIntegerReturnFacts()
	if err := g.preparePackageInitBodies(); err != nil {
		return nil, err
	}
	if err := g.prepareGoHostSupport(); err != nil {
		return nil, err
	}
	files := make(map[string][]byte)
	compiled, err := g.renderCompiled()
	if err != nil {
		return nil, err
	}
	files["compiled.go"] = compiled
	if g.hasFunctions() {
		registerSrc, err := g.renderCompiledRegisterFile()
		if err != nil {
			return nil, err
		}
		files["compiled_register.go"] = registerSrc
		importSeedingSrc, err := g.renderCompiledImportSeedingFile()
		if err != nil {
			return nil, err
		}
		files["compiled_import_seeding.go"] = importSeedingSrc
		interfaceDispatchSrc, err := g.renderCompiledInterfaceDispatchFile()
		if err != nil {
			return nil, err
		}
		files["compiled_interface_dispatch.go"] = interfaceDispatchSrc
		methodImplFiles, err := g.renderCompiledPackageMethodImplFiles()
		if err != nil {
			return nil, err
		}
		for name, src := range methodImplFiles {
			files[name] = src
		}
		definitionFiles, err := g.renderCompiledPackageDefinitionFiles()
		if err != nil {
			return nil, err
		}
		for name, src := range definitionFiles {
			files[name] = src
		}
		registrarFiles, err := g.renderCompiledPackageRegistrarFiles()
		if err != nil {
			return nil, err
		}
		for name, src := range registrarFiles {
			files[name] = src
		}
		callableFiles, err := g.renderCompiledPackageCallableFiles()
		if err != nil {
			return nil, err
		}
		for name, src := range callableFiles {
			files[name] = src
		}
		compiledPackageAggregators, err := g.renderCompiledPackageAggregatorsFile()
		if err != nil {
			return nil, err
		}
		files["compiled_package_aggregators.go"] = compiledPackageAggregators
		packageInitSrc, err := g.renderCompiledPackageInitFile()
		if err != nil {
			return nil, err
		}
		if len(packageInitSrc) > 0 {
			files["compiled_package_inits.go"] = packageInitSrc
		}
	}
	if g.opts.EmitMain {
		mainSrc, err := g.renderMain()
		if err != nil {
			return nil, err
		}
		files["main.go"] = mainSrc
	}
	g.discardRedundantImplFallbackSpecializations()
	return files, nil
}

func (g *generator) renderCompiled() ([]byte, error) {
	// Render body first so compilation sets flags (e.g. needsStrconv)
	// that affect import selection.
	var body bytes.Buffer

	if g.hasFunctions() {
		fmt.Fprintf(&body, "const __able_experimental_mono_arrays = %t\n\n", g.monoArraysEnabled())
		fmt.Fprintf(&body, "var __able_runtime *bridge.Runtime\n\n")
		g.ensurePackageEnvVars()
		if len(g.packageEnvOrder) > 0 {
			for _, pkgName := range g.packageEnvOrder {
				if envVar, ok := g.packageEnvVars[pkgName]; ok {
					fmt.Fprintf(&body, "var %s *runtime.Environment\n", envVar)
				}
				if bootVar, ok := g.packageBootstrappedVars[pkgName]; ok {
					fmt.Fprintf(&body, "var %s bool\n", bootVar)
				}
			}
			fmt.Fprintf(&body, "\n")
		}
		g.renderRuntimeHelpers(&body)
		g.renderGoPreludeDecls(&body)
	}

	if g.hasFunctions() {
		renderedMethods, renderedFunctions := g.renderCompiledBodies(&body)
		g.renderMethodWrappers(&body)
		g.renderWrappers(&body)
		g.renderFunctionThunks(&body)
		g.renderOverloadDispatchers(&body)
		g.renderMethodThunks(&body)
		// Warm native interface discovery before the later body/fallback passes so
		// adapter/default-method specialization can settle, but defer the actual
		// code emission until after the compiled body graph has stabilized.
		var nativeInterfaceWarmup bytes.Buffer
		g.renderNativeInterfaces(&nativeInterfaceWarmup)
		g.nativeInterfaceRenderedAdapters = make(map[string]struct{})
		g.nativeInterfaceRenderedInfos = make(map[string]struct{})
		g.nativeInterfaceRenderedDispatches = make(map[string]struct{})
		g.nativeInterfaceRenderedApplyHelpers = make(map[string]struct{})
		g.renderIteratorCollectMonoArrayHelpers(&body)
		// Generic nominal specializations and carrier arrays are discovered while
		// compiling function bodies above, so emit their concrete type
		// declarations only after codegen has settled.
		g.renderMonoArrayTypes(&body)
		g.renderStructs(&body)
		g.renderStructConverters(&body)
		g.renderMonoArrayConverters(&body)
		for g.renderPendingNativeInterfaceConcreteAdapters(&body) {
		}
		g.renderAdditionalCompiledBodies(&body, renderedMethods, renderedFunctions)
		g.renderPendingCompiledMethodFallbacks(&body, renderedMethods)
		g.renderPendingCompiledFunctionFallbacks(&body, renderedFunctions)
		// Emit shared native interfaces and unions only after the compiled body
		// graph has stabilized so the generated adapter/default-method bodies are
		// final on first emission.
		g.renderNativeInterfaces(&body)
		g.renderNativeUnions(&body)
		// Boundary/apply helpers must wait until every body/specialization pass
		// has settled so their interface-adapter switch tables are complete.
		g.renderNativeInterfaceBoundaryHelpersFinal(&body)
		g.renderNativeInterfaceApplyRuntimeHelpers(&body)
		g.renderNativeCallables(&body)
		g.renderNominalCoercions(&body)
		g.renderDiagnosticGlobals(&body)
	} else {
		g.renderMonoArrayTypes(&body)
		g.renderStructs(&body)
		g.renderNativeCallables(&body)
		g.renderNativeInterfaces(&body)
		g.renderIteratorCollectMonoArrayHelpers(&body)
		g.renderNativeUnions(&body)
		for g.renderPendingNativeInterfaceConcreteAdapters(&body) {
		}
	}

	// Now render the header with imports (flags are set by body rendering).
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)
	imports := g.importsForCompiled(body.String())
	if len(imports) > 0 {
		fmt.Fprintf(&buf, "import (\n")
		for _, imp := range imports {
			fmt.Fprintf(&buf, "\t%s\n", formatCompiledImportSpec(imp))
		}
		fmt.Fprintf(&buf, ")\n\n")
	}
	buf.Write(body.Bytes())

	return formatSource(buf.Bytes())
}

func formatCompiledImportSpec(spec string) string {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return `""`
	}
	if strings.HasPrefix(spec, `"`) || strings.Contains(spec, " ") || strings.HasPrefix(spec, ". ") || strings.HasPrefix(spec, "_ ") {
		return spec
	}
	return strconv.Quote(spec)
}

func (g *generator) renderDiagnosticGlobals(buf *bytes.Buffer) {
	if g == nil || buf == nil {
		return
	}
	if len(g.diagNodes) > 0 {
		for _, info := range g.diagNodes {
			initExpr := ""
			switch {
			case info.CallName != "":
				initExpr = fmt.Sprintf("&ast.FunctionCall{Callee: ast.NewIdentifier(%q)}", info.CallName)
			case info.CallMember != "":
				initExpr = fmt.Sprintf("&ast.FunctionCall{Callee: ast.NewMemberAccessExpression(ast.NewIdentifier(\"\"), ast.NewIdentifier(%q))}", info.CallMember)
			default:
				goType := info.GoType
				if strings.HasPrefix(goType, "*") {
					goType = "&" + strings.TrimPrefix(goType, "*")
				}
				initExpr = fmt.Sprintf("%s{}", goType)
			}
			fmt.Fprintf(buf, "var %s = %s\n", info.Name, initExpr)
		}
		fmt.Fprintf(buf, "\n")
	}
	if len(g.awaitExprs) > 0 {
		for _, name := range g.awaitExprs {
			fmt.Fprintf(buf, "var %s = &ast.AwaitExpression{}\n", name)
		}
		fmt.Fprintf(buf, "\n")
	}
}

func (g *generator) importsForCompiled(compiledBody string) []string {
	importSet := map[string]struct{}{}
	needsRuntime := g.hasFunctions() || g.structUsesRuntimeValue()
	if g.hasFunctions() {
		importSet["context"] = struct{}{}
		importSet["errors"] = struct{}{}
		importSet["fmt"] = struct{}{}
		importSet["math"] = struct{}{}
		importSet["math/big"] = struct{}{}
		importSet["math/bits"] = struct{}{}
		importSet["sort"] = struct{}{}
		importSet["strings"] = struct{}{}
		importSet["sync"] = struct{}{}
		importSet["sync/atomic"] = struct{}{}
		importSet["time"] = struct{}{}
		importSet["unicode/utf8"] = struct{}{}
		importSet["able/interpreter-go/pkg/compiler/bridge"] = struct{}{}
		importSet["able/interpreter-go/pkg/ast"] = struct{}{}
		importSet["able/interpreter-go/pkg/interpreter"] = struct{}{}
	}
	if needsRuntime {
		importSet["able/interpreter-go/pkg/runtime"] = struct{}{}
	}
	for _, imp := range g.goPreludeImports {
		if strings.TrimSpace(imp) == "" {
			continue
		}
		importSet[imp] = struct{}{}
	}
	if g.needsIterator {
		importSet["errors"] = struct{}{}
	}
	if g.hasFunctions() && g.needsIterator {
		importSet["sync"] = struct{}{}
	}
	if g.needsStrconv && strings.Contains(compiledBody, "strconv.") {
		importSet["strconv"] = struct{}{}
	}
	imports := make([]string, 0, len(importSet))
	for imp := range importSet {
		imports = append(imports, imp)
	}
	sort.Strings(imports)
	return imports
}

func (g *generator) importsForCompiledPackageAggregators() []string {
	imports := []string{
		"able/interpreter-go/pkg/compiler/bridge",
		"able/interpreter-go/pkg/interpreter",
		"able/interpreter-go/pkg/runtime",
	}
	sort.Strings(imports)
	return imports
}

func (g *generator) importsForCompiledRegister() []string {
	imports := []string{
		"able/interpreter-go/pkg/ast",
		"able/interpreter-go/pkg/compiler/bridge",
		"able/interpreter-go/pkg/interpreter",
		"able/interpreter-go/pkg/runtime",
		"fmt",
	}
	sort.Strings(imports)
	return imports
}

func (g *generator) structUsesRuntimeValue() bool {
	for _, info := range g.allStructInfos() {
		if info == nil {
			continue
		}
		for _, field := range info.Fields {
			if field.GoType == "runtime.Value" {
				return true
			}
		}
	}
	return false
}
