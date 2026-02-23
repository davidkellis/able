package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func (g *generator) renderCompiledImportSeedingFile() ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)
	imports := []string{
		"able/interpreter-go/pkg/runtime",
		"strings",
	}
	sort.Strings(imports)
	fmt.Fprintf(&buf, "import (\n")
	for _, imp := range imports {
		fmt.Fprintf(&buf, "\t%q\n", imp)
	}
	fmt.Fprintf(&buf, ")\n\n")
	fmt.Fprintf(&buf, "var (\n")
	fmt.Fprintf(&buf, "\t_ runtime.Value\n")
	fmt.Fprintf(&buf, "\t_ = strings.TrimSpace\n")
	fmt.Fprintf(&buf, ")\n\n")
	fmt.Fprintf(&buf, "func __able_seed_no_bootstrap_imports(__able_bootstrapped_metadata bool) {\n")
	fmt.Fprintf(&buf, "\t_ = __able_bootstrapped_metadata\n")
	packageList := g.registrationPackageList()
	g.renderNoBootstrapImportSeeding(&buf, packageList)
	fmt.Fprintf(&buf, "}\n")
	return formatSource(buf.Bytes())
}

func (g *generator) renderNoBootstrapImportSeeding(buf *bytes.Buffer, packageList []string) {
	if g == nil || buf == nil || len(packageList) == 0 || len(g.staticImports) == 0 {
		return
	}
	targetPackages := make([]string, 0, len(packageList))
	for _, pkgName := range packageList {
		if len(g.staticImportsForPackage(pkgName)) == 0 {
			continue
		}
		if _, ok := g.packageEnvVar(pkgName); !ok {
			continue
		}
		targetPackages = append(targetPackages, pkgName)
	}
	if len(targetPackages) == 0 {
		return
	}
	sort.Strings(targetPackages)
	sourceSet := make(map[string]struct{})
	for _, pkgName := range targetPackages {
		for _, binding := range g.staticImportsForPackage(pkgName) {
			if binding.SourcePackage == "" {
				continue
			}
			if _, ok := g.packageEnvVar(binding.SourcePackage); !ok {
				continue
			}
			sourceSet[binding.SourcePackage] = struct{}{}
		}
	}
	if len(sourceSet) == 0 {
		return
	}
	sourcePackages := make([]string, 0, len(sourceSet))
	for pkgName := range sourceSet {
		sourcePackages = append(sourcePackages, pkgName)
	}
	sort.Strings(sourcePackages)
	sourcePublicVars := make(map[string]string, len(sourcePackages))

	fmt.Fprintf(buf, "\tif !__able_bootstrapped_metadata {\n")
	fmt.Fprintf(buf, "\t\t__able_make_pkg_callable := func(pkgEnv *runtime.Environment, name string) runtime.Value {\n")
	fmt.Fprintf(buf, "\t\t\tif pkgEnv == nil || strings.TrimSpace(name) == \"\" {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn nil\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\tentry := __able_lookup_compiled_call(pkgEnv, name)\n")
	fmt.Fprintf(buf, "\t\t\tif entry == nil || entry.fn == nil {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn nil\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\tfn := &runtime.NativeFunctionValue{Name: name, Arity: entry.arity}\n")
	fmt.Fprintf(buf, "\t\t\tfn.Impl = func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
	fmt.Fprintf(buf, "\t\t\t\tif ctx == nil {\n")
	fmt.Fprintf(buf, "\t\t\t\t\tctx = &runtime.NativeCallContext{Env: pkgEnv}\n")
	fmt.Fprintf(buf, "\t\t\t\t} else if ctx.Env == nil {\n")
	fmt.Fprintf(buf, "\t\t\t\t\tctx.Env = pkgEnv\n")
	fmt.Fprintf(buf, "\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t\treturn entry.fn.Impl(ctx, args)\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\treturn fn\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\t_ = __able_make_pkg_callable\n")
	fmt.Fprintf(buf, "\t\t__able_define_struct_binding := func(env *runtime.Environment, name string, val runtime.Value) {\n")
	fmt.Fprintf(buf, "\t\t\tif env == nil || strings.TrimSpace(name) == \"\" || val == nil {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\tswitch def := val.(type) {\n")
	fmt.Fprintf(buf, "\t\t\tcase *runtime.StructDefinitionValue:\n")
	fmt.Fprintf(buf, "\t\t\t\tenv.DefineStruct(name, def)\n")
	fmt.Fprintf(buf, "\t\t\tcase runtime.StructDefinitionValue:\n")
	fmt.Fprintf(buf, "\t\t\t\tcopyDef := def\n")
	fmt.Fprintf(buf, "\t\t\t\tenv.DefineStruct(name, &copyDef)\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\t_ = __able_define_struct_binding\n")

	for idx, sourcePkg := range sourcePackages {
		sourceEnvVar, ok := g.packageEnvVar(sourcePkg)
		if !ok {
			continue
		}
		publicVar := fmt.Sprintf("__able_pkg_public_%d", idx)
		sourcePublicVars[sourcePkg] = publicVar
		fmt.Fprintf(buf, "\t\t%s := map[string]runtime.Value{}\n", publicVar)
		for _, name := range g.sortedPublicCallableNames(sourcePkg) {
			fmt.Fprintf(buf, "\t\tif callable := __able_make_pkg_callable(%s, %q); callable != nil {\n", sourceEnvVar, name)
			fmt.Fprintf(buf, "\t\t\t%s[%q] = callable\n", publicVar, name)
			fmt.Fprintf(buf, "\t\t}\n")
		}
		for _, name := range g.sortedPublicStructNames(sourcePkg) {
			fmt.Fprintf(buf, "\t\tif def, ok := %s.StructDefinition(%q); ok && def != nil {\n", sourceEnvVar, name)
			fmt.Fprintf(buf, "\t\t\t%s[%q] = def\n", publicVar, name)
			fmt.Fprintf(buf, "\t\t}\n")
		}
		for _, name := range g.sortedPublicInterfaceNames(sourcePkg) {
			fmt.Fprintf(buf, "\t\tif val, err := %s.Get(%q); err == nil && val != nil {\n", sourceEnvVar, name)
			fmt.Fprintf(buf, "\t\t\t%s[%q] = val\n", publicVar, name)
			fmt.Fprintf(buf, "\t\t}\n")
		}
		for _, name := range g.sortedPublicUnionNames(sourcePkg) {
			fmt.Fprintf(buf, "\t\tif val, err := %s.Get(%q); err == nil && val != nil {\n", sourceEnvVar, name)
			fmt.Fprintf(buf, "\t\t\t%s[%q] = val\n", publicVar, name)
			fmt.Fprintf(buf, "\t\t}\n")
		}
		for _, name := range g.sortedPublicImplNamespaceNames(sourcePkg) {
			fmt.Fprintf(buf, "\t\tif val, err := %s.Get(%q); err == nil && val != nil {\n", sourceEnvVar, name)
			fmt.Fprintf(buf, "\t\t\t%s[%q] = val\n", publicVar, name)
			fmt.Fprintf(buf, "\t\t}\n")
		}
		fmt.Fprintf(buf, "\t\t_ = %s\n", publicVar)
	}

	for _, targetPkg := range targetPackages {
		targetEnvVar, ok := g.packageEnvVar(targetPkg)
		if !ok {
			continue
		}
		for _, binding := range g.staticImportsForPackage(targetPkg) {
			publicVar, ok := sourcePublicVars[binding.SourcePackage]
			if !ok {
				continue
			}
			switch binding.Kind {
			case staticImportBindingPackage:
				namePath := packageNamePathLiteral(binding.SourcePackage)
				fmt.Fprintf(buf, "\t\t%s.Define(%q, runtime.PackageValue{Name: %q, NamePath: %s, Public: %s})\n",
					targetEnvVar, binding.LocalName, binding.SourcePackage, namePath, publicVar)
			case staticImportBindingSelector:
				sourceEnvVar, sourceOk := g.packageEnvVar(binding.SourcePackage)
				fmt.Fprintf(buf, "\t\tif !%s.HasInCurrentScope(%q) {\n", targetEnvVar, binding.LocalName)
				fmt.Fprintf(buf, "\t\t\tif val, ok := %s[%q]; ok && val != nil {\n", publicVar, binding.SourceName)
				fmt.Fprintf(buf, "\t\t\t\t%s.Define(%q, val)\n", targetEnvVar, binding.LocalName)
				fmt.Fprintf(buf, "\t\t\t\t__able_define_struct_binding(%s, %q, val)\n", targetEnvVar, binding.LocalName)
				if sourceOk {
					fmt.Fprintf(buf, "\t\t\t} else if def, ok := %s.StructDefinition(%q); ok && def != nil {\n", sourceEnvVar, binding.SourceName)
					fmt.Fprintf(buf, "\t\t\t\t%s.Define(%q, def)\n", targetEnvVar, binding.LocalName)
					fmt.Fprintf(buf, "\t\t\t\t%s.DefineStruct(%q, def)\n", targetEnvVar, binding.LocalName)
				}
				fmt.Fprintf(buf, "\t\t\t}\n")
				fmt.Fprintf(buf, "\t\t}\n")
			case staticImportBindingWildcard:
				for _, name := range g.sortedImportableNames(binding.SourcePackage) {
					fmt.Fprintf(buf, "\t\tif !%s.HasInCurrentScope(%q) {\n", targetEnvVar, name)
					fmt.Fprintf(buf, "\t\t\tif val, ok := %s[%q]; ok && val != nil {\n", publicVar, name)
					fmt.Fprintf(buf, "\t\t\t\t%s.Define(%q, val)\n", targetEnvVar, name)
					fmt.Fprintf(buf, "\t\t\t\t__able_define_struct_binding(%s, %q, val)\n", targetEnvVar, name)
					fmt.Fprintf(buf, "\t\t\t}\n")
					fmt.Fprintf(buf, "\t\t}\n")
				}
			}
		}
	}
	fmt.Fprintf(buf, "\t}\n")
}

func packageNamePathLiteral(pkgName string) string {
	trimmed := strings.TrimSpace(pkgName)
	if trimmed == "" {
		return "nil"
	}
	parts := strings.Split(trimmed, ".")
	literals := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		literals = append(literals, fmt.Sprintf("%q", part))
	}
	if len(literals) == 0 {
		return "nil"
	}
	return "[]string{" + strings.Join(literals, ", ") + "}"
}
