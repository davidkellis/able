package compiler

import "strings"

// compiledEnvironmentEffect records the deliberately small proof surface used
// to decide whether an imported static call may enter a raw generated body.
// Absence of proof always retains the package-environment entry wrapper.
type compiledEnvironmentEffect struct {
	localIndependent bool
	callees          map[*functionInfo]struct{}
	generatedBody    string
}

func (g *generator) resolveCompiledEnvironmentIndependence() {
	if g == nil {
		return
	}
	effects := make(map[*functionInfo]*compiledEnvironmentEffect)
	for {
		before := len(g.environmentEffectFunctionInfos())
		for _, info := range g.environmentEffectFunctionInfos() {
			if _, analyzed := effects[info]; analyzed || info == nil || !info.Compileable || info.Definition == nil || info.Definition.Body == nil || info.ExternBody != nil {
				continue
			}
			effect := &compiledEnvironmentEffect{
				localIndependent: true,
				callees:          make(map[*functionInfo]struct{}),
			}
			ctx := g.compileBodyContext(info)
			ctx.analysisOnly = true
			ctx.environmentEffect = effect
			lines, retExpr, ok := g.compileBody(ctx, info)
			effect.generatedBody = strings.Join(lines, "\n") + "\n" + retExpr
			if !ok || generatedBodyUsesPackageEnvironment(lines, retExpr) {
				effect.localIndependent = false
			}
			effects[info] = effect
		}
		if len(g.environmentEffectFunctionInfos()) == before {
			break
		}
	}
	for _, effect := range effects {
		g.recordGeneratedCompiledCallees(effect)
	}

	independent := make(map[*functionInfo]bool, len(effects))
	for info, effect := range effects {
		if effect.localIndependent {
			independent[info] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for info := range independent {
			for callee := range effects[info].callees {
				if !independent[callee] {
					delete(independent, info)
					changed = true
					break
				}
			}
		}
	}
	g.environmentIndependent = independent
	g.environmentIndependentGoNames = make(map[string]bool, len(independent))
	for info := range independent {
		if info != nil && info.GoName != "" {
			g.environmentIndependentGoNames[info.GoName] = true
		}
	}
}

// environmentEffectFunctionInfos includes inherent methods in the same
// conservative fixed point as free functions and interface implementations.
// allFunctionInfos deliberately serves other generator passes that do not
// render inherent methods, so broadening it would change unrelated emission.
func (g *generator) environmentEffectFunctionInfos() []*functionInfo {
	if g == nil {
		return nil
	}
	infos := g.sortedFunctionInfos()
	seen := make(map[*functionInfo]struct{}, len(infos)+len(g.methodList))
	for _, info := range infos {
		if info != nil {
			seen[info] = struct{}{}
		}
	}
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil {
			continue
		}
		if _, ok := seen[method.Info]; ok {
			continue
		}
		seen[method.Info] = struct{}{}
		infos = append(infos, method.Info)
	}
	return infos
}

func generatedBodyUsesPackageEnvironment(lines []string, retExpr string) bool {
	body := strings.Join(lines, "\n") + "\n" + retExpr
	for _, fragment := range []string{
		"__able_env_",
		"__able_pkg_env_",
		"__able_call_named",
		"__able_call_value",
		"__able_method_call",
		"__able_member_get",
		"__able_member_set",
		"__able_extern_call",
		"__able_spawn",
		"__able_await",
		"__able_function_call_env",
		"__able_lookup_",
		"bridge.Call",
		".RuntimeData()",
	} {
		if strings.Contains(body, fragment) {
			return true
		}
	}
	return false
}

func (g *generator) recordGeneratedCompiledCallees(effect *compiledEnvironmentEffect) {
	if g == nil || effect == nil {
		return
	}
	for _, info := range g.environmentEffectFunctionInfos() {
		if info == nil || info.GoName == "" {
			continue
		}
		// Entry wrappers establish their own package environment, so only raw
		// generated-body calls propagate a callee dependency to the caller.
		if strings.Contains(effect.generatedBody, "__able_compiled_"+info.GoName) {
			effect.callees[info] = struct{}{}
		}
	}
}

func (g *generator) compiledFunctionEnvironmentIndependentByGoName(goName string) bool {
	return g != nil && g.environmentIndependentGoNames[goName]
}

func (g *generator) compiledEnvironmentEffectIndependent(effect *compiledEnvironmentEffect) bool {
	if g == nil || effect == nil || !effect.localIndependent {
		return false
	}
	for callee := range effect.callees {
		if callee == nil || (!g.environmentIndependent[callee] && !g.compiledFunctionEnvironmentIndependentByGoName(callee.GoName)) {
			return false
		}
	}
	return true
}

func (g *generator) functionInfoByGoName(goName string) *functionInfo {
	if g == nil || strings.TrimSpace(goName) == "" {
		return nil
	}
	for _, info := range g.allFunctionInfos() {
		if info != nil && info.GoName == goName {
			return info
		}
	}
	return nil
}
