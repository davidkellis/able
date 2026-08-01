package compiler

import "able/interpreter-go/pkg/ast"

// resolvedNominalEffectAnalyzer centralizes construction of the opt-in effect
// model so ownership diagnostics can consume the same fail-closed fixed point
// without changing ordinary compilation or generated execution.
func (g *generator) resolvedNominalEffectAnalyzer() *nominalEffectAnalyzer {
	if g == nil {
		return nil
	}
	analyzer := &nominalEffectAnalyzer{
		g:          g,
		byInfo:     make(map[*functionInfo]*nominalEffectCallable),
		byLambda:   make(map[*ast.LambdaExpression]*nominalEffectCallable),
		byLocalDef: make(map[*ast.FunctionDefinition]*nominalEffectCallable),
	}
	analyzer.collectRootCallables()
	for index := 0; index < len(analyzer.callables); index++ {
		analyzer.collectNestedCallables(analyzer.callables[index])
	}
	for _, callable := range analyzer.callables {
		analyzer.collectAliases(callable)
	}
	for _, callable := range analyzer.callables {
		analyzer.collectDirectEffects(callable)
	}
	analyzer.closeParameterTypeFixedPoint()
	analyzer.resetEffectsWithInferredParameterTypes()
	for _, callable := range analyzer.callables {
		analyzer.collectDirectEffects(callable)
	}
	analyzer.closeFixedPoint()
	return analyzer
}
