package compiler

import (
	"fmt"
	"strings"
)

const executionContextType = "*__able_execution_context"

func (g *generator) executionContextsEnabled() bool {
	return g != nil && (g.opts.ExperimentalExecutionContext || g.schedulerExecutionContextRequired)
}

func (g *generator) callableExecutionContextsEnabled() bool {
	return g != nil && g.executionContextsEnabled() && g.schedulerExecutionContextRequired
}

func (g *generator) compiledCallArgs(ctx *compileContext, args []string) string {
	if g != nil && g.executionContextsEnabled() && ctx != nil && ctx.executionContextExpr != "" {
		args = append(append([]string{}, args...), ctx.executionContextExpr)
	}
	return strings.Join(args, ", ")
}

// compiledContextCallTargetName selects the fixed-context entry point only for
// source-level static calls that already carry a compile context.  Rendered
// runtime adapters intentionally keep using the compatibility entry points:
// they have no lexical execution context to propagate.
func (g *generator) compiledContextCallTargetName(ctx *compileContext, callerPackage string, info *functionInfo) string {
	if g != nil && info != nil {
		if g.staticCallTargets == nil {
			g.staticCallTargets = make(map[*functionInfo]struct{})
		}
		g.staticCallTargets[info] = struct{}{}
	}
	target := g.compiledCallTargetName(callerPackage, info)
	if ctx != nil && ctx.environmentEffect != nil && info != nil && target == g.compiledBodyName(info) {
		// Only a raw-body call inherits the callee's package-environment
		// requirement. An entry wrapper establishes the callee package itself.
		ctx.environmentEffect.callees[info] = struct{}{}
	}
	if g != nil && g.executionContextsEnabled() && info != nil && info.Compileable && ctx != nil && ctx.executionContextExpr != "" {
		return compiledContextName(target)
	}
	return target
}

func (g *generator) compiledContextCallTargetNameForPackage(ctx *compileContext, callerPackage string, calleePackage string, goName string) string {
	target := g.compiledCallTargetNameForPackage(callerPackage, calleePackage, goName)
	if ctx != nil && ctx.environmentEffect != nil {
		rawTarget := "__able_compiled_" + strings.TrimSpace(goName)
		if target == rawTarget {
			if info := g.functionInfoByGoName(goName); info != nil {
				ctx.environmentEffect.callees[info] = struct{}{}
			} else {
				ctx.environmentEffect.localIndependent = false
			}
		}
	}
	if g != nil && g.executionContextsEnabled() && ctx != nil && ctx.executionContextExpr != "" {
		return compiledContextName(target)
	}
	return target
}

func compiledContextName(name string) string {
	if name == "" {
		return ""
	}
	return name + "_ctx"
}

func (g *generator) compiledNativeWrapperCallName(info *functionInfo) string {
	if g != nil && g.callableExecutionContextsEnabled() {
		return compiledContextName(g.compiledCallTargetName("", info))
	}
	return g.compiledBodyName(info)
}

func (g *generator) compiledNativeWrapperCallArgs(info *functionInfo, contextExpr string) string {
	if info == nil {
		return ""
	}
	args := make([]string, 0, len(info.Params)+1)
	for _, param := range info.Params {
		args = append(args, param.GoName)
	}
	if g != nil && g.callableExecutionContextsEnabled() {
		if strings.TrimSpace(contextExpr) == "" {
			contextExpr = "__able_context_from_args()"
		}
		args = append(args, contextExpr)
	}
	return strings.Join(args, ", ")
}

func (g *generator) nativeCallableParamParts(paramGoTypes []string, namePrefix string) []string {
	parts := make([]string, 0, len(paramGoTypes)+1)
	for idx, paramType := range paramGoTypes {
		parts = append(parts, fmt.Sprintf("%s%d %s", namePrefix, idx, paramType))
	}
	if g != nil && g.callableExecutionContextsEnabled() {
		parts = append(parts, "__able_exec_ctx "+executionContextType)
	}
	return parts
}

func (g *generator) nativeCallableInvokeArgs(ctx *compileContext, args []string) []string {
	result := append([]string{}, args...)
	if g == nil || !g.callableExecutionContextsEnabled() {
		return result
	}
	contextExpr := "__able_context_from_args()"
	if ctx != nil && strings.TrimSpace(ctx.executionContextExpr) != "" {
		contextExpr = ctx.executionContextExpr
	}
	return append(result, contextExpr)
}

func (g *generator) inlineExecutionContextEnvLinesForPackage(pkgName string, contextExpr string) []string {
	if g == nil {
		return nil
	}
	if !g.executionContextsEnabled() || strings.TrimSpace(contextExpr) == "" {
		return g.inlineRuntimeEnvSwapLinesForPackage(pkgName)
	}
	envVar, ok := g.packageEnvVar(pkgName)
	if !ok || envVar == "" {
		return nil
	}
	localContext := contextExpr + "_package"
	lines := []string{
		fmt.Sprintf("var %s __able_execution_context", localContext),
		fmt.Sprintf("%s = __able_context_with_environment(%s, %s, &%s)", contextExpr, contextExpr, envVar, localContext),
	}
	return append(lines, g.inlineRuntimeEnvSwapLinesForPackage(pkgName)...)
}

func (g *generator) runtimeHelperCallExpr(ctx *compileContext, helper string, args string) string {
	if g != nil && g.executionContextsEnabled() && ctx != nil && ctx.executionContextExpr != "" {
		return fmt.Sprintf("%s(%s, %s)", runtimeHelperContextName(helper), args, ctx.executionContextExpr)
	}
	return fmt.Sprintf("%s(%s)", helper, args)
}

func runtimeHelperContextName(helper string) string {
	return strings.TrimSuffix(helper, "_impl") + "_ctx"
}

func contextAwareConcurrencyHelper(helper string) bool {
	switch helper {
	case "__able_channel_new_impl",
		"__able_channel_send_impl",
		"__able_channel_receive_impl",
		"__able_channel_try_send_impl",
		"__able_channel_try_receive_impl",
		"__able_channel_await_try_recv_impl",
		"__able_channel_await_try_send_impl",
		"__able_channel_close_impl",
		"__able_channel_is_closed_impl",
		"__able_mutex_new_impl",
		"__able_mutex_lock_impl",
		"__able_mutex_unlock_impl",
		"__able_mutex_await_lock_impl":
		return true
	default:
		return false
	}
}

func (g *generator) runtimeConcurrencyContextParams() string {
	if g.executionContextsEnabled() {
		return ", __able_exec_ctxs ...*__able_execution_context"
	}
	return ""
}
