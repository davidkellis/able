package compiler

func (g *generator) compileBodyContext(info *functionInfo) *compileContext {
	if g == nil {
		return newCompileContext(nil, info, nil, nil, "", nil)
	}
	ctx := newCompileContext(g, info, g.functionsForCompileContext(info), g.overloadsForPackage(info.Package), info.Package, g.compileContextGenericNames(info))
	ctx.implSiblings = g.defaultSiblingsForFunction(info)
	return ctx
}

func (g *generator) bodyCompileable(info *functionInfo, retType string) bool {
	if info == nil || info.Definition == nil {
		return false
	}
	if info.ExternBody != nil {
		if ok := g.externBodyCompileable(info); ok {
			info.Reason = ""
		} else if info.Reason == "" {
			info.Reason = "unsupported extern body"
		}
		return info.Reason == ""
	}
	if info.Definition.Body == nil {
		info.Reason = "missing function body"
		return false
	}
	ctx := g.compileBodyContext(info)
	ctx.analysisOnly = true
	_, _, ok := g.compileBody(ctx, info)
	if !ok {
		info.Reason = ctx.reason
		if info.Reason == "" {
			info.Reason = "unsupported function body"
		}
	}
	return ok
}

func (g *generator) resolveCompileableFunctions() {
	for {
		allInfos := g.sortedFunctionInfos()
		pending := make([]*functionInfo, 0, len(allInfos))
		for _, info := range allInfos {
			if info == nil {
				continue
			}
			if !info.SupportedTypes {
				info.Compileable = false
				continue
			}
			if !info.Compileable {
				pending = append(pending, info)
			}
		}
		progress := false
		for _, info := range pending {
			if info == nil {
				continue
			}
			if info.Compileable {
				continue
			}
			if ok := g.bodyCompileable(info, info.ReturnType); ok {
				info.Compileable = true
				info.Reason = ""
				progress = true
			}
		}
		if len(g.allFunctionInfos()) != len(allInfos) {
			continue
		}
		if !progress {
			for _, info := range pending {
				if info == nil {
					continue
				}
				if info.Compileable {
					continue
				}
				if info.Reason == "" {
					info.Reason = "unsupported function body"
				}
				info.Compileable = false
			}
			break
		}
	}
	g.touchNativeInterfaceAdapters()
}

func (g *generator) resolveCompileabilityFixedPoint() {
	if g == nil {
		return
	}
	for {
		functionCount := len(g.allFunctionInfos())
		specializedCount := len(g.specializedFunctions)
		g.resolveCompileableFunctions()
		g.resolveCompileableMethods()
		if len(g.allFunctionInfos()) == functionCount && len(g.specializedFunctions) == specializedCount {
			break
		}
	}
}

func (g *generator) collectFallbacks() []FallbackInfo {
	if g == nil {
		return nil
	}
	fallbacks := make([]FallbackInfo, 0, len(g.fallbacks))
	fallbacks = append(fallbacks, g.fallbacks...)
	for _, info := range g.sortedFunctionInfos() {
		if info == nil || info.Compileable {
			continue
		}
		reason := info.Reason
		if reason == "" {
			reason = "unsupported function body"
		}
		name := info.Name
		if info.QualifiedName != "" {
			name = info.QualifiedName
		}
		fallbacks = append(fallbacks, FallbackInfo{Name: name, Reason: reason})
	}
	return fallbacks
}
