package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func (g *generator) render() (map[string][]byte, error) {
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
	}
	if g.opts.EmitMain {
		mainSrc, err := g.renderMain()
		if err != nil {
			return nil, err
		}
		files["main.go"] = mainSrc
	}
	return files, nil
}

func (g *generator) renderCompiled() ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)

	imports := g.importsForCompiled()
	if len(imports) > 0 {
		fmt.Fprintf(&buf, "import (\n")
		for _, imp := range imports {
			fmt.Fprintf(&buf, "\t%q\n", imp)
		}
		fmt.Fprintf(&buf, ")\n\n")
	}

	if g.hasFunctions() {
		fmt.Fprintf(&buf, "var __able_runtime *bridge.Runtime\n\n")
		g.ensurePackageEnvVars()
		if len(g.packageEnvOrder) > 0 {
			for _, pkgName := range g.packageEnvOrder {
				if envVar, ok := g.packageEnvVars[pkgName]; ok {
					fmt.Fprintf(&buf, "var %s *runtime.Environment\n", envVar)
				}
			}
			fmt.Fprintf(&buf, "\n")
		}
		g.renderRuntimeHelpers(&buf)
	}

	g.renderStructs(&buf)
	if g.hasFunctions() {
		g.renderStructConverters(&buf)
		g.renderCompiledMethods(&buf)
		g.renderCompiledFunctions(&buf)
		g.renderMethodWrappers(&buf)
		g.renderWrappers(&buf)
		g.renderFunctionThunks(&buf)
		g.renderOverloadDispatchers(&buf)
		g.renderMethodThunks(&buf)
		g.renderDiagnosticGlobals(&buf)
	}

	return formatSource(buf.Bytes())
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

func (g *generator) importsForCompiled() []string {
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
	if g.needsIterator {
		importSet["errors"] = struct{}{}
	}
	if g.hasFunctions() && g.needsIterator {
		importSet["sync"] = struct{}{}
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
		"able/interpreter-go/pkg/compiler/bridge",
		"able/interpreter-go/pkg/interpreter",
		"able/interpreter-go/pkg/runtime",
		"fmt",
	}
	sort.Strings(imports)
	return imports
}

func (g *generator) structUsesRuntimeValue() bool {
	for _, info := range g.structs {
		for _, field := range info.Fields {
			if field.GoType == "runtime.Value" {
				return true
			}
		}
	}
	return false
}
