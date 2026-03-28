package compiler

import (
	"bytes"
	"fmt"
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
	g.renderCompiledFunctionBody(buf, info, lines, retExpr)
	rendered[info] = struct{}{}
	return true
}

func (g *generator) renderPendingCompiledFunctionFallbacks(buf *bytes.Buffer, rendered map[*functionInfo]struct{}) {
	for _, info := range g.sortedFunctionInfos() {
		if info == nil || !info.Compileable {
			continue
		}
		if _, ok := rendered[info]; ok {
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
	fmt.Fprintf(buf, "func __able_compiled_%s(", info.GoName)
	for i, param := range info.Params {
		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
	if envVar, ok := g.packageEnvVar(info.Package); ok {
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
	}
	for _, line := range lines {
		fmt.Fprintf(buf, "\t%s\n", line)
	}
	fmt.Fprintf(buf, "\treturn %s, nil\n", retExpr)
	fmt.Fprintf(buf, "}\n\n")
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
	g.renderCompiledMethodBody(buf, info, lines, retExpr)
	rendered[method] = struct{}{}
	return true
}

func (g *generator) renderPendingCompiledMethodFallbacks(buf *bytes.Buffer, rendered map[*methodInfo]struct{}) {
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil || !method.Info.Compileable {
			continue
		}
		if _, ok := rendered[method]; ok {
			continue
		}
		info := method.Info
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
	fmt.Fprintf(buf, "func __able_compiled_%s(", info.GoName)
	for i, param := range info.Params {
		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
	if envVar, ok := g.packageEnvVar(info.Package); ok {
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
	}
	for _, line := range lines {
		fmt.Fprintf(buf, "\t%s\n", line)
	}
	fmt.Fprintf(buf, "\treturn %s, nil\n", retExpr)
	fmt.Fprintf(buf, "}\n\n")
}
