package interpreter

import "able/interpreter-go/pkg/ast"

type bytecodeExpressionProgramCacheKey struct {
	expr                   ast.Expression
	allowPlaceholderLambda bool
}

func (i *Interpreter) lookupCachedExpressionBytecode(expr ast.Expression, allowPlaceholderLambda bool) (*bytecodeProgram, bool) {
	if i == nil || expr == nil {
		return nil, false
	}
	key := bytecodeExpressionProgramCacheKey{
		expr:                   expr,
		allowPlaceholderLambda: allowPlaceholderLambda,
	}
	i.bytecodeExprCacheMu.RLock()
	program, ok := i.bytecodeExprCache[key]
	i.bytecodeExprCacheMu.RUnlock()
	return program, ok
}

func (i *Interpreter) cacheExpressionBytecode(expr ast.Expression, allowPlaceholderLambda bool, program *bytecodeProgram) *bytecodeProgram {
	if i == nil || expr == nil || program == nil {
		return program
	}
	key := bytecodeExpressionProgramCacheKey{
		expr:                   expr,
		allowPlaceholderLambda: allowPlaceholderLambda,
	}
	i.bytecodeExprCacheMu.Lock()
	if existing, ok := i.bytecodeExprCache[key]; ok && existing != nil {
		i.bytecodeExprCacheMu.Unlock()
		return existing
	}
	i.bytecodeExprCache[key] = program
	i.bytecodeExprCacheMu.Unlock()
	return program
}
