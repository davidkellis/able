package compiler

import (
	"bytes"
	"fmt"
	"strings"
)

func (g *generator) renderCompiledBodies(buf *bytes.Buffer) (map[*methodInfo]struct{}, map[*functionInfo]struct{}) {
	renderedMethods := make(map[*methodInfo]struct{})
	renderedFunctions := make(map[*functionInfo]struct{})
	g.renderAdditionalCompiledBodies(buf, renderedMethods, renderedFunctions)
	return renderedMethods, renderedFunctions
}

func (g *generator) renderAdditionalCompiledBodies(buf *bytes.Buffer, renderedMethods map[*methodInfo]struct{}, renderedFunctions map[*functionInfo]struct{}) {
	for {
		progress := false
		specializedCount := len(g.specializedFunctions)
		for _, method := range g.sortedMethodInfos() {
			if g.tryRenderCompiledMethodBody(buf, method, renderedMethods) {
				progress = true
			}
		}
		for _, info := range g.sortedFunctionInfos() {
			if g.tryRenderCompiledFunctionBody(buf, info, renderedFunctions) {
				progress = true
			}
		}
		if progress || len(g.specializedFunctions) != specializedCount {
			continue
		}
		break
	}
}

// renderNativeInterfaceSpecializationBodies closes the dependency loop between
// interface adapters and compiled default-method specializations. Rendering an
// adapter can resolve a concrete inherited method for the first time; that
// method body must then be emitted before the generated adapter references it.
func (g *generator) renderNativeInterfaceSpecializationBodies(buf *bytes.Buffer, renderedMethods map[*methodInfo]struct{}, renderedFunctions map[*functionInfo]struct{}) {
	if g == nil || buf == nil {
		return
	}
	for {
		beforeFunctions := len(renderedFunctions)
		beforeMethods := len(renderedMethods)
		beforeSpecializations := len(g.specializedFunctions)

		g.renderNativeInterfaces(buf)
		g.renderAdditionalCompiledBodies(buf, renderedMethods, renderedFunctions)

		if beforeFunctions == len(renderedFunctions) &&
			beforeMethods == len(renderedMethods) &&
			beforeSpecializations == len(g.specializedFunctions) {
			return
		}
	}
}

func (g *generator) tryRenderCompiledFunctionBody(buf *bytes.Buffer, info *functionInfo, rendered map[*functionInfo]struct{}) bool {
	if info == nil || !info.Compileable {
		return false
	}
	if _, ok := rendered[info]; ok {
		return false
	}
	g.refreshRepresentableFunctionInfo(info)
	if info.ExternBody != nil {
		g.renderCompiledExternFunctionBody(buf, info)
		rendered[info] = struct{}{}
		return true
	}
	ctx := g.compileBodyContext(info)
	lines, retExpr, ok := g.compileBody(ctx, info)
	if !ok {
		if info.Reason == "" {
			reason := ctx.reason
			if reason == "" {
				reason = "unsupported function body"
			}
			info.Reason = reason
		}
		return false
	}
	info.Reason = ""
	if isCanonicalPrimitiveStringUTF8ValidateFunction(info) {
		g.renderCanonicalPrimitiveStringUTF8ValidateFunctionBody(buf, info, lines, retExpr)
	} else if isCanonicalPrimitiveStringUTF8DecodeFunction(info) {
		g.renderCanonicalPrimitiveStringUTF8DecodeFunctionBody(buf, info, lines, retExpr)
	} else {
		g.renderCompiledFunctionBody(buf, info, lines, retExpr)
	}
	g.renderCallerOwnedResultVariant(buf, info)
	g.renderNominalOwnershipVariant(buf, info)
	rendered[info] = struct{}{}
	return true
}

func (g *generator) renderPendingCompiledFunctionFallbacks(buf *bytes.Buffer, rendered map[*functionInfo]struct{}) {
	for _, info := range g.sortedFunctionInfos() {
		if info == nil {
			continue
		}
		if _, ok := rendered[info]; ok {
			continue
		}
		if !info.Compileable {
			if _, required := g.staticCallTargets[info]; required && info.HasOriginal {
				g.renderCompiledFunctionFallback(buf, info)
				rendered[info] = struct{}{}
			}
			continue
		}
		ctx := g.compileBodyContext(info)
		if _, _, ok := g.compileBody(ctx, info); !ok && info.Reason == "" {
			reason := ctx.reason
			if reason == "" {
				reason = "unsupported function body"
			}
			info.Reason = reason
		}
		if info.Reason == "" {
			info.Reason = "unsupported function body"
		}
		info.Compileable = false
		g.renderCompiledFunctionFallback(buf, info)
	}
}

func (g *generator) renderCompiledFunctionBody(buf *bytes.Buffer, info *functionInfo, lines []string, retExpr string) {
	g.renderCompiledFunctionBodyWithPrefix(buf, info, nil, lines, retExpr)
}

func (g *generator) renderCompiledFunctionBodyWithPrefix(buf *bytes.Buffer, info *functionInfo, prefix []string, lines []string, retExpr string) {
	prefix = append(g.dynamicImportScopePrefix(info), prefix...)
	bodyName := g.compiledBodyName(info)
	entryName := g.compiledEntryName(info)
	if g.executionContextsEnabled() {
		bodyName = g.compiledContextBodyName(info)
		entryName = g.compiledContextEntryName(info)
	}
	fmt.Fprintf(buf, "func %s(", bodyName)
	for i, param := range info.Params {
		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	if g.executionContextsEnabled() {
		if len(info.Params) > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "__able_exec_ctx %s", executionContextType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
	if g.executionContextsEnabled() {
		fmt.Fprintf(buf, "\t_ = __able_exec_ctx\n")
	}
	for _, line := range prefix {
		fmt.Fprintf(buf, "\t%s\n", line)
	}
	for _, line := range lines {
		fmt.Fprintf(buf, "\t%s\n", line)
	}
	fmt.Fprintf(buf, "\treturn %s, nil\n", retExpr)
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(", entryName)
	for i, param := range info.Params {
		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	if g.executionContextsEnabled() {
		if len(info.Params) > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "__able_exec_ctx %s", executionContextType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
	if g.executionContextsEnabled() {
		fmt.Fprintf(buf, "\t_ = __able_exec_ctx\n")
		if envVar, ok := g.packageEnvVar(info.Package); ok {
			writeExecutionContextPackageEnv(buf, "\t", "__able_exec_ctx", "__able_runtime", envVar)
		}
	} else if envVar, ok := g.packageEnvVar(info.Package); ok {
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
	}
	args := make([]string, 0, len(info.Params))
	for _, param := range info.Params {
		args = append(args, param.GoName)
	}
	fmt.Fprintf(buf, "\treturn %s(%s)\n", bodyName, g.compiledCallArgs(&compileContext{executionContextExpr: "__able_exec_ctx"}, args))
	fmt.Fprintf(buf, "}\n\n")
	if g.executionContextsEnabled() {
		g.renderCompiledContextCompatibilityWrappers(buf, info)
	}
}

func (g *generator) tryRenderCompiledMethodBody(buf *bytes.Buffer, method *methodInfo, rendered map[*methodInfo]struct{}) bool {
	if method == nil || method.Info == nil || !method.Info.Compileable {
		return false
	}
	if _, ok := rendered[method]; ok {
		return false
	}
	info := method.Info
	g.refreshRepresentableFunctionInfo(info)
	if isNativeArrayCoreMethod(method) {
		g.renderNativeArrayCoreMethod(buf, method, info)
		rendered[method] = struct{}{}
		return true
	}
	ctx := g.compileBodyContext(info)
	lines, retExpr, ok := g.compileBody(ctx, info)
	if !ok {
		if info.Reason == "" {
			reason := ctx.reason
			if reason == "" {
				reason = "unsupported method body"
			}
			info.Reason = reason
		}
		return false
	}
	info.Reason = ""
	if isCanonicalPrimitiveStringContainsMethod(method) {
		g.renderCanonicalPrimitiveStringContainsMethodBody(buf, info, lines, retExpr)
		rendered[method] = struct{}{}
		return true
	}
	if isCanonicalPrimitiveStringLenBytesMethod(method) {
		g.renderCanonicalPrimitiveStringLenBytesMethodBody(buf, info, lines, retExpr)
		rendered[method] = struct{}{}
		return true
	}
	if isCanonicalPrimitiveStringFromBytesUncheckedMethod(method) {
		g.renderCanonicalPrimitiveStringFromBytesUncheckedMethodBody(buf, info, lines, retExpr)
		rendered[method] = struct{}{}
		return true
	}
	if isCanonicalPrimitiveStringCharsMethod(method) {
		g.renderCanonicalPrimitiveStringCharsMethodBody(buf, info, lines, retExpr)
		rendered[method] = struct{}{}
		return true
	}
	if isCanonicalPrimitiveStringBytesMethod(method) {
		g.renderCanonicalPrimitiveStringBytesMethodBody(buf, info, lines, retExpr)
		rendered[method] = struct{}{}
		return true
	}
	g.renderCompiledMethodBody(buf, info, lines, retExpr)
	g.renderCallerOwnedResultVariant(buf, info)
	g.renderNominalOwnershipVariant(buf, info)
	rendered[method] = struct{}{}
	return true
}

func (g *generator) renderPendingCompiledMethodFallbacks(buf *bytes.Buffer, rendered map[*methodInfo]struct{}) {
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil {
			continue
		}
		if _, ok := rendered[method]; ok {
			continue
		}
		info := method.Info
		if !info.Compileable {
			if _, required := g.staticCallTargets[info]; required && info.HasOriginal {
				g.renderCompiledMethodFallback(buf, method)
				rendered[method] = struct{}{}
			}
			continue
		}
		if isNativeArrayCoreMethod(method) {
			g.renderNativeArrayCoreMethod(buf, method, info)
			continue
		}
		ctx := g.compileBodyContext(info)
		if _, _, ok := g.compileBody(ctx, info); !ok && info.Reason == "" {
			reason := ctx.reason
			if reason == "" {
				reason = "unsupported method body"
			}
			info.Reason = reason
		}
		if info.Reason == "" {
			info.Reason = "unsupported method body"
		}
		info.Compileable = false
		g.renderCompiledMethodFallback(buf, method)
	}
}

func (g *generator) renderCompiledMethodBody(buf *bytes.Buffer, info *functionInfo, lines []string, retExpr string) {
	g.renderCompiledMethodBodyWithPrefix(buf, info, nil, lines, retExpr)
}

func (g *generator) renderCompiledMethodBodyWithPrefix(buf *bytes.Buffer, info *functionInfo, prefix []string, lines []string, retExpr string) {
	prefix = append(g.dynamicImportScopePrefix(info), prefix...)
	bodyName := g.compiledBodyName(info)
	entryName := g.compiledEntryName(info)
	if g.executionContextsEnabled() {
		bodyName = g.compiledContextBodyName(info)
		entryName = g.compiledContextEntryName(info)
	}
	fmt.Fprintf(buf, "func %s(", bodyName)
	for i, param := range info.Params {
		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	if g.executionContextsEnabled() {
		if len(info.Params) > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "__able_exec_ctx %s", executionContextType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
	if g.executionContextsEnabled() {
		fmt.Fprintf(buf, "\t_ = __able_exec_ctx\n")
	}
	for _, line := range prefix {
		fmt.Fprintf(buf, "\t%s\n", line)
	}
	for _, line := range lines {
		fmt.Fprintf(buf, "\t%s\n", line)
	}
	fmt.Fprintf(buf, "\treturn %s, nil\n", retExpr)
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(", entryName)
	for i, param := range info.Params {
		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	if g.executionContextsEnabled() {
		if len(info.Params) > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "__able_exec_ctx %s", executionContextType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
	if g.executionContextsEnabled() {
		fmt.Fprintf(buf, "\t_ = __able_exec_ctx\n")
		if envVar, ok := g.packageEnvVar(info.Package); ok {
			writeExecutionContextPackageEnv(buf, "\t", "__able_exec_ctx", "__able_runtime", envVar)
		}
	} else if envVar, ok := g.packageEnvVar(info.Package); ok {
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
	}
	args := make([]string, 0, len(info.Params))
	for _, param := range info.Params {
		args = append(args, param.GoName)
	}
	fmt.Fprintf(buf, "\treturn %s(%s)\n", bodyName, g.compiledCallArgs(&compileContext{executionContextExpr: "__able_exec_ctx"}, args))
	fmt.Fprintf(buf, "}\n\n")
	if g.executionContextsEnabled() {
		g.renderCompiledContextCompatibilityWrappers(buf, info)
	}
}

// renderCompiledContextCompatibilityWrappers keeps the generated callable
// names stable for runtime adapters and external callers.  Experimental static
// lowering calls the paired fixed-pointer entries above; these wrappers are
// deliberately cold boundary shims.
func (g *generator) renderCompiledContextCompatibilityWrappers(buf *bytes.Buffer, info *functionInfo) {
	if g == nil || buf == nil || info == nil {
		return
	}
	args := make([]string, 0, len(info.Params)+1)
	for _, param := range info.Params {
		args = append(args, param.GoName)
	}
	writeSignature := func(name string) {
		fmt.Fprintf(buf, "func %s(", name)
		for i, param := range info.Params {
			if i > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
		}
		fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
	}
	writeSignature(g.compiledBodyName(info))
	fmt.Fprintf(buf, "\treturn %s(%s)\n", g.compiledContextBodyName(info), strings.Join(append(args, "__able_context_from_args()"), ", "))
	fmt.Fprintf(buf, "}\n\n")
	writeSignature(g.compiledEntryName(info))
	fmt.Fprintf(buf, "\treturn %s(%s)\n", g.compiledContextEntryName(info), strings.Join(append(args, "__able_context_from_args()"), ", "))
	fmt.Fprintf(buf, "}\n\n")
}
