package compiler

// requiresBootstrapExecution reports whether the generated executable must
// evaluate the source program before it can enter compiled code. Keep this
// decision shared by all generated artifacts whose metadata is only needed by
// that interpreter bootstrap path.
func (g *generator) requiresBootstrapExecution() bool {
	if g == nil {
		return true
	}
	if g.hasDynamicFeature || len(g.collectFallbacks()) > 0 || !g.noBootstrapImportsSeedable() {
		return true
	}
	functionCount := 0
	for _, byName := range g.functions {
		for _, info := range byName {
			if info == nil {
				continue
			}
			functionCount++
			if !info.Compileable {
				return true
			}
		}
	}
	return functionCount == 0
}

// retainsPackageInterfaceDefaultBodies reports whether package-interface
// default AST bodies must be emitted for interpreter registration. A static
// executable calls the separately generated compiled default helpers, so the
// redundant AST bodies would be unreachable. Library output has no launcher
// decision and retains the metadata for its existing consumers.
func (g *generator) retainsPackageInterfaceDefaultBodies() bool {
	return g == nil || !g.opts.EmitMain || g.requiresBootstrapExecution()
}
