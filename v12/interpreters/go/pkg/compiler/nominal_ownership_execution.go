package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"able/interpreter-go/pkg/ast"
)

const nominalOwnershipVariantSuffix = "_owned"

type nominalOwnershipExecutionSite struct {
	call            *ast.FunctionCall
	caller          *nominalEffectCallable
	sourceParameter int
	sourceArgument  int
	path            []string
	targets         []*nominalEffectCallable
	dispatch        string
	method          string
	nominal         *structInfo
	dispatchName    string
}

type nominalOwnershipVariant struct {
	parameter int
	path      []string
	nominal   *structInfo
}

type nominalOwnershipInterfaceDispatch struct {
	name   string
	site   *nominalOwnershipExecutionSite
	info   *nativeInterfaceInfo
	method *nativeInterfaceMethod
}

func nominalOwnershipVariantName(name string) string {
	if name == "" {
		return ""
	}
	return name + nominalOwnershipVariantSuffix
}

func (g *generator) prepareNominalOwnershipExecution() *NominalOwnershipReport {
	analyzer := g.nominalOwnershipAnalyzer()
	if analyzer == nil {
		return nil
	}
	g.nominalOwnershipExecutionSites = make(map[*ast.FunctionCall]*nominalOwnershipExecutionSite)
	g.nominalOwnershipVariants = make(map[*functionInfo]*nominalOwnershipVariant)
	g.nominalOwnershipInterfaceDispatches = make(map[string]*nominalOwnershipInterfaceDispatch)
	for callable, summaries := range analyzer.byParam {
		if callable == nil || callable.info == nil || !callable.info.Compileable || len(summaries) != 1 {
			continue
		}
		for parameter, value := range summaries {
			g.prepareNominalOwnershipSuccessorVariant(callable, parameter, value.path)
		}
	}
	for call, site := range analyzer.executionSites {
		if g.prepareNominalOwnershipExecutionSite(site) {
			g.nominalOwnershipExecutionSites[call] = site
		}
	}
	return analyzer.report()
}

func (g *generator) prepareNominalOwnershipSuccessorVariant(
	target *nominalEffectCallable,
	parameter int,
	path []string,
) bool {
	if target == nil || target.info == nil || parameter < 0 || parameter >= len(target.params) {
		return false
	}
	typeExpr := callableParameterType(target, parameter, target.params[parameter])
	info, ok := g.structInfoForTypeExpr(callablePackage(target), typeExpr)
	if !ok || info == nil || !info.Supported || info.Kind == ast.StructKindSingleton ||
		!g.nominalOwnershipResultPathSupported(target.info, path, info) {
		return false
	}
	if existing := g.nominalOwnershipVariantInfo(target.info); existing != nil {
		return existing.nominal.GoName == info.GoName &&
			existing.parameter == parameter &&
			ownershipPathsEqual(existing.path, path)
	}
	g.nominalOwnershipVariants[target.info] = &nominalOwnershipVariant{
		parameter: parameter,
		path:      append([]string(nil), path...),
		nominal:   info,
	}
	return true
}

func (g *generator) prepareNominalOwnershipExecutionSite(site *nominalOwnershipExecutionSite) bool {
	if g == nil || site == nil || site.call == nil || site.caller == nil || len(site.targets) == 0 {
		return false
	}
	var nominal *structInfo
	for _, target := range site.targets {
		if target == nil || target.info == nil || !target.info.Compileable ||
			site.sourceParameter < 0 || site.sourceParameter >= len(target.params) {
			return false
		}
		typeExpr := callableParameterType(target, site.sourceParameter, target.params[site.sourceParameter])
		info, ok := g.structInfoForTypeExpr(callablePackage(target), typeExpr)
		if !ok || info == nil || !info.Supported || info.Kind == ast.StructKindSingleton {
			return false
		}
		if nominal != nil && nominal.GoName != info.GoName {
			return false
		}
		nominal = info
		if !g.prepareNominalOwnershipSuccessorVariant(target, site.sourceParameter, site.path) {
			return false
		}
	}
	site.nominal = nominal
	return nominal != nil
}

func (g *generator) nominalOwnershipVariantInfo(info *functionInfo) *nominalOwnershipVariant {
	if g == nil || info == nil {
		return nil
	}
	if variant := g.nominalOwnershipVariants[info]; variant != nil {
		return variant
	}
	for candidate, variant := range g.nominalOwnershipVariants {
		if sameNominalOwnershipFunction(candidate, info) {
			return variant
		}
	}
	return nil
}

func sameNominalOwnershipFunction(left, right *functionInfo) bool {
	if left == nil || right == nil {
		return false
	}
	return left == right ||
		(left.Definition != nil && left.Definition == right.Definition &&
			left.GoName == right.GoName && left.ReturnType == right.ReturnType)
}

func (g *generator) nominalOwnershipResultPathSupported(info *functionInfo, path []string, nominal *structInfo) bool {
	if info == nil || nominal == nil || len(path) > 1 {
		return false
	}
	finalType, _, ok := g.nominalOwnershipResultField(info.ReturnType, path)
	return ok && finalType == "*"+nominal.GoName
}

func (g *generator) nominalOwnershipResultField(goType string, path []string) (string, []string, bool) {
	if len(path) == 0 {
		return goType, nil, goType != ""
	}
	currentType := goType
	goPath := make([]string, 0, len(path))
	for _, name := range path {
		info := g.structInfoByGoName(currentType)
		if info == nil {
			return "", nil, false
		}
		field, ok := nominalOwnershipField(info, name)
		if !ok {
			return "", nil, false
		}
		currentType = field.GoType
		goPath = append(goPath, field.GoName)
	}
	return currentType, goPath, true
}

func nominalOwnershipField(info *structInfo, name string) (fieldInfo, bool) {
	if info == nil {
		return fieldInfo{}, false
	}
	if strings.HasPrefix(name, "#") {
		index, err := strconv.Atoi(strings.TrimPrefix(name, "#"))
		if err == nil && index >= 0 && index < len(info.Fields) {
			return info.Fields[index], info.Fields[index].Supported
		}
		return fieldInfo{}, false
	}
	for _, field := range info.Fields {
		if field.Name == name {
			return field, field.Supported
		}
	}
	return fieldInfo{}, false
}

func (g *generator) nominalOwnershipStaticCall(
	ctx *compileContext,
	call *ast.FunctionCall,
	info *functionInfo,
	args []string,
	callTarget string,
) ([]string, string, []string, bool) {
	if g == nil || call == nil || info == nil || len(args) == 0 {
		return args, callTarget, nil, false
	}
	variant := g.nominalOwnershipVariantInfo(info)
	if ctx != nil && ctx.callerOwnedResultSlot != "" && ctx.callerOwnedTailExpr == call &&
		ctx.nominalOwnershipResultType != "" && variant != nil && len(variant.path) == 0 &&
		variant.nominal != nil && "*"+variant.nominal.GoName == ctx.nominalOwnershipResultType {
		args = append(args, ctx.callerOwnedResultSlot)
		return args, nominalOwnershipVariantName(callTarget), nil, true
	}
	if ctx != nil && ctx.nominalOwnershipVariantActive && variant != nil &&
		len(variant.path) == 0 && variant.nominal != nil {
		slotTemp := ctx.newTemp()
		args = append(args, "&"+slotTemp)
		lines := []string{fmt.Sprintf("var %s %s", slotTemp, variant.nominal.GoName)}
		return args, nominalOwnershipVariantName(callTarget), lines, true
	}
	site := g.nominalOwnershipExecutionSites[call]
	if site == nil || site.dispatch != "direct" || variant == nil ||
		site.sourceParameter < 0 || site.sourceParameter >= len(args) ||
		variant.nominal == nil || site.nominal == nil ||
		variant.nominal.GoName != site.nominal.GoName {
		return args, callTarget, nil, false
	}
	args = append(args, args[site.sourceParameter])
	return args, nominalOwnershipVariantName(callTarget), nil, true
}

func (g *generator) nominalOwnershipInterfaceCall(
	ctx *compileContext,
	call *ast.FunctionCall,
	receiverExpr string,
	receiverType string,
	method *nativeInterfaceMethod,
	args []string,
) (string, bool) {
	if g == nil || ctx == nil || call == nil || method == nil {
		return "", false
	}
	site := g.nominalOwnershipExecutionSites[call]
	if site == nil || site.dispatch != "native-interface" || site.nominal == nil ||
		site.sourceArgument < 0 || site.sourceArgument >= len(args) {
		return "", false
	}
	info := g.nativeInterfaceInfoForGoType(receiverType)
	if info == nil {
		return "", false
	}
	key := info.Key + "::" + method.Name + "::" + site.nominal.GoName + "::" + strings.Join(site.path, ".")
	dispatch := g.nominalOwnershipInterfaceDispatches[key]
	if dispatch == nil {
		name := g.mangler.unique(fmt.Sprintf(
			"__able_owned_iface_%s_%s_%s",
			sanitizeIdent(method.InterfaceName),
			sanitizeIdent(method.Name),
			sanitizeIdent(site.nominal.GoName),
		))
		dispatch = &nominalOwnershipInterfaceDispatch{name: name, site: site, info: info, method: method}
		g.nominalOwnershipInterfaceDispatches[key] = dispatch
	}
	site.dispatchName = dispatch.name
	callArgs := make([]string, 0, len(args)+3)
	callArgs = append(callArgs, receiverExpr)
	callArgs = append(callArgs, args...)
	callArgs = append(callArgs, args[site.sourceArgument])
	if g.executionContextsEnabled() && ctx.executionContextExpr != "" {
		callArgs = append(callArgs, ctx.executionContextExpr)
	}
	return fmt.Sprintf("%s(%s)", dispatch.name, strings.Join(callArgs, ", ")), true
}

func (g *generator) renderNominalOwnershipVariant(buf *bytes.Buffer, info *functionInfo) {
	variant := g.nominalOwnershipVariantInfo(info)
	if g == nil || buf == nil || info == nil || variant == nil || variant.nominal == nil {
		return
	}
	ctx := g.compileBodyContext(info)
	ctx.nominalOwnershipVariantActive = true
	if len(variant.path) == 0 {
		ctx.callerOwnedResultSlot = "__able_owned"
		ctx.nominalOwnershipResultType = "*" + variant.nominal.GoName
	}
	lines, resultExpr, ok := g.compileBody(ctx, info)
	if !ok {
		return
	}
	bodyName := g.compiledBodyName(info)
	entryName := g.compiledEntryName(info)
	if g.executionContextsEnabled() {
		bodyName = g.compiledContextBodyName(info)
		entryName = g.compiledContextEntryName(info)
	}
	bodyName = nominalOwnershipVariantName(bodyName)
	entryName = nominalOwnershipVariantName(entryName)

	g.writeNominalOwnershipVariantSignature(buf, bodyName, info, variant)
	if g.executionContextsEnabled() {
		fmt.Fprintln(buf, "\t_ = __able_exec_ctx")
	}
	for _, line := range lines {
		fmt.Fprintf(buf, "\t%s\n", line)
	}
	if len(variant.path) == 0 {
		fmt.Fprintf(buf, "\t*__able_owned = *%s\n", resultExpr)
		fmt.Fprintln(buf, "\treturn __able_owned, nil")
	} else {
		_, goPath, supported := g.nominalOwnershipResultField(info.ReturnType, variant.path)
		if !supported {
			return
		}
		fmt.Fprintf(buf, "\t__able_owned_outer := %s\n", resultExpr)
		fieldExpr := "__able_owned_outer." + strings.Join(goPath, ".")
		fmt.Fprintf(buf, "\t*__able_owned = *%s\n", fieldExpr)
		outerInfo := g.structInfoByGoName(info.ReturnType)
		if outerInfo == nil || len(goPath) != 1 {
			return
		}
		fmt.Fprintf(buf, "\treturn &%s{", outerInfo.GoName)
		for index, field := range outerInfo.Fields {
			if index > 0 {
				fmt.Fprint(buf, ", ")
			}
			value := "__able_owned_outer." + field.GoName
			if field.GoName == goPath[0] {
				value = "__able_owned"
			}
			fmt.Fprintf(buf, "%s: %s", field.GoName, value)
		}
		fmt.Fprintln(buf, "}, nil")
	}
	fmt.Fprintln(buf, "}")
	fmt.Fprintln(buf)

	g.writeNominalOwnershipVariantSignature(buf, entryName, info, variant)
	if g.executionContextsEnabled() {
		fmt.Fprintln(buf, "\t_ = __able_exec_ctx")
		if envVar, ok := g.packageEnvVar(info.Package); ok {
			writeExecutionContextPackageEnv(buf, "\t", "__able_exec_ctx", "__able_runtime", envVar)
		}
	} else if envVar, ok := g.packageEnvVar(info.Package); ok {
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
	}
	args := make([]string, 0, len(info.Params)+1)
	for _, param := range info.Params {
		args = append(args, param.GoName)
	}
	args = append(args, "__able_owned")
	fmt.Fprintf(buf, "\treturn %s(%s)\n", bodyName, g.compiledCallArgs(&compileContext{executionContextExpr: "__able_exec_ctx"}, args))
	fmt.Fprintln(buf, "}")
	fmt.Fprintln(buf)
}

func (g *generator) writeNominalOwnershipVariantSignature(
	buf *bytes.Buffer,
	name string,
	info *functionInfo,
	variant *nominalOwnershipVariant,
) {
	fmt.Fprintf(buf, "func %s(", name)
	for index, param := range info.Params {
		if index > 0 {
			fmt.Fprint(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	if len(info.Params) > 0 {
		fmt.Fprint(buf, ", ")
	}
	fmt.Fprintf(buf, "__able_owned *%s", variant.nominal.GoName)
	if g.executionContextsEnabled() {
		fmt.Fprintf(buf, ", __able_exec_ctx %s", executionContextType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
}

func (g *generator) renderNominalOwnershipInterfaceDispatches(buf *bytes.Buffer) {
	if g == nil || buf == nil || len(g.nominalOwnershipInterfaceDispatches) == 0 {
		return
	}
	keys := make([]string, 0, len(g.nominalOwnershipInterfaceDispatches))
	for key := range g.nominalOwnershipInterfaceDispatches {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		g.renderNominalOwnershipInterfaceDispatch(buf, g.nominalOwnershipInterfaceDispatches[key])
	}
}

func (g *generator) renderNominalOwnershipInterfaceDispatch(
	buf *bytes.Buffer,
	dispatch *nominalOwnershipInterfaceDispatch,
) {
	if dispatch == nil || dispatch.site == nil || dispatch.site.nominal == nil ||
		dispatch.info == nil || dispatch.method == nil || dispatch.name == "" {
		return
	}
	method := dispatch.method
	fmt.Fprintf(buf, "func %s(receiver %s", dispatch.name, dispatch.info.GoType)
	for index, goType := range method.ParamGoTypes {
		fmt.Fprintf(buf, ", arg%d %s", index, goType)
	}
	fmt.Fprintf(buf, ", __able_owned *%s", dispatch.site.nominal.GoName)
	if g.executionContextsEnabled() {
		fmt.Fprintf(buf, ", __able_exec_ctx %s", executionContextType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", method.ReturnGoType)
	fmt.Fprintln(buf, "\tswitch w := receiver.(type) {")
	for _, adapter := range g.nativeInterfaceKnownAdapters(dispatch.info) {
		if !g.nominalOwnershipAdapterDispatchSupported(dispatch, adapter) {
			continue
		}
		impl := adapter.Methods[method.Name]
		fmt.Fprintf(buf, "\tcase %s:\n", adapter.AdapterType)
		args := make([]string, 0, len(method.ParamGoTypes)+2)
		if method.ExpectsSelf {
			args = append(args, "w.Value")
		}
		for index := range method.ParamGoTypes {
			args = append(args, fmt.Sprintf("arg%d", index))
		}
		args = append(args, "__able_owned")
		target := g.compiledCallTargetName(dispatch.site.caller.ctx.packageName, impl.Info)
		if g.executionContextsEnabled() {
			target = compiledContextName(target)
			args = append(args, "__able_exec_ctx")
		}
		fmt.Fprintf(buf, "\t\treturn %s(%s)\n", nominalOwnershipVariantName(target), strings.Join(args, ", "))
	}
	fmt.Fprintln(buf, "\tdefault:")
	args := make([]string, 0, len(method.ParamGoTypes)+1)
	for index := range method.ParamGoTypes {
		args = append(args, fmt.Sprintf("arg%d", index))
	}
	methodName := method.GoName
	if g.executionContextsEnabled() {
		methodName = nativeInterfaceContextMethodName(method)
		args = append(args, "__able_exec_ctx")
	}
	fmt.Fprintf(buf, "\t\treturn receiver.%s(%s)\n", methodName, strings.Join(args, ", "))
	fmt.Fprintln(buf, "\t}")
	fmt.Fprintln(buf, "}")
	fmt.Fprintln(buf)
}

func (g *generator) nominalOwnershipAdapterDispatchSupported(
	dispatch *nominalOwnershipInterfaceDispatch,
	adapter *nativeInterfaceAdapter,
) bool {
	if dispatch == nil || dispatch.method == nil || adapter == nil || adapter.AdapterType == "" {
		return false
	}
	impl := adapter.Methods[dispatch.method.Name]
	if impl == nil || impl.Info == nil || g.nominalOwnershipVariantInfo(impl.Info) == nil {
		return false
	}
	expected := len(dispatch.method.ParamGoTypes)
	if dispatch.method.ExpectsSelf {
		expected++
	}
	if len(impl.Info.Params) != expected || impl.Info.ReturnType != dispatch.method.ReturnGoType {
		return false
	}
	offset := 0
	if dispatch.method.ExpectsSelf {
		if impl.Info.Params[0].GoType != adapter.GoType {
			return false
		}
		offset = 1
	}
	for index, goType := range dispatch.method.ParamGoTypes {
		if impl.Info.Params[index+offset].GoType != goType {
			return false
		}
	}
	return true
}
