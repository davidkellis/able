package interpreter

import "able/interpreter-go/pkg/ast"

func (i *Interpreter) cachedTypeExpressionTuple(args []ast.TypeExpression) []ast.TypeExpression {
	switch len(args) {
	case 0:
		return nil
	case 1:
		return i.cachedTypeExpressionTuple1(args[0])
	case 2:
		return i.cachedTypeExpressionTuple2(args[0], args[1])
	case 3:
		return i.cachedTypeExpressionTuple3(args[0], args[1], args[2])
	}
	if i == nil {
		return append([]ast.TypeExpression(nil), args...)
	}
	key := makeTypeExpressionSliceKey(args)
	i.typeInfoCacheMu.RLock()
	if cached, ok := i.typeExpressionTupleCache[key]; ok {
		i.typeInfoCacheMu.RUnlock()
		return cached
	}
	i.typeInfoCacheMu.RUnlock()
	created := append([]ast.TypeExpression(nil), args...)
	i.typeInfoCacheMu.Lock()
	if i.typeExpressionTupleCache == nil {
		i.typeExpressionTupleCache = make(map[typeExpressionSliceKey][]ast.TypeExpression)
	}
	if existing, ok := i.typeExpressionTupleCache[key]; ok {
		i.typeInfoCacheMu.Unlock()
		return existing
	}
	i.typeExpressionTupleCache[key] = created
	i.typeInfoCacheMu.Unlock()
	return created
}

func (i *Interpreter) cachedTypeExpressionTuple1(arg0 ast.TypeExpression) []ast.TypeExpression {
	if arg0 == nil {
		return nil
	}
	if i == nil {
		return []ast.TypeExpression{arg0}
	}
	key := typeExpressionSliceKey{count: 1, arg0: arg0}
	i.typeInfoCacheMu.RLock()
	if cached, ok := i.typeExpressionTupleCache[key]; ok {
		i.typeInfoCacheMu.RUnlock()
		return cached
	}
	i.typeInfoCacheMu.RUnlock()
	created := []ast.TypeExpression{arg0}
	i.typeInfoCacheMu.Lock()
	if i.typeExpressionTupleCache == nil {
		i.typeExpressionTupleCache = make(map[typeExpressionSliceKey][]ast.TypeExpression)
	}
	if existing, ok := i.typeExpressionTupleCache[key]; ok {
		i.typeInfoCacheMu.Unlock()
		return existing
	}
	i.typeExpressionTupleCache[key] = created
	i.typeInfoCacheMu.Unlock()
	return created
}

func (i *Interpreter) cachedTypeExpressionTuple2(arg0 ast.TypeExpression, arg1 ast.TypeExpression) []ast.TypeExpression {
	if arg0 == nil && arg1 == nil {
		return nil
	}
	if i == nil {
		return []ast.TypeExpression{arg0, arg1}
	}
	key := typeExpressionSliceKey{count: 2, arg0: arg0, arg1: arg1}
	i.typeInfoCacheMu.RLock()
	if cached, ok := i.typeExpressionTupleCache[key]; ok {
		i.typeInfoCacheMu.RUnlock()
		return cached
	}
	i.typeInfoCacheMu.RUnlock()
	created := []ast.TypeExpression{arg0, arg1}
	i.typeInfoCacheMu.Lock()
	if i.typeExpressionTupleCache == nil {
		i.typeExpressionTupleCache = make(map[typeExpressionSliceKey][]ast.TypeExpression)
	}
	if existing, ok := i.typeExpressionTupleCache[key]; ok {
		i.typeInfoCacheMu.Unlock()
		return existing
	}
	i.typeExpressionTupleCache[key] = created
	i.typeInfoCacheMu.Unlock()
	return created
}

func (i *Interpreter) cachedTypeExpressionTuple3(arg0 ast.TypeExpression, arg1 ast.TypeExpression, arg2 ast.TypeExpression) []ast.TypeExpression {
	if arg0 == nil && arg1 == nil && arg2 == nil {
		return nil
	}
	if i == nil {
		return []ast.TypeExpression{arg0, arg1, arg2}
	}
	key := typeExpressionSliceKey{count: 3, arg0: arg0, arg1: arg1, arg2: arg2}
	i.typeInfoCacheMu.RLock()
	if cached, ok := i.typeExpressionTupleCache[key]; ok {
		i.typeInfoCacheMu.RUnlock()
		return cached
	}
	i.typeInfoCacheMu.RUnlock()
	created := []ast.TypeExpression{arg0, arg1, arg2}
	i.typeInfoCacheMu.Lock()
	if i.typeExpressionTupleCache == nil {
		i.typeExpressionTupleCache = make(map[typeExpressionSliceKey][]ast.TypeExpression)
	}
	if existing, ok := i.typeExpressionTupleCache[key]; ok {
		i.typeInfoCacheMu.Unlock()
		return existing
	}
	i.typeExpressionTupleCache[key] = created
	i.typeInfoCacheMu.Unlock()
	return created
}
