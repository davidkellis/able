package interpreter

import (
	"reflect"
	"sort"
	"strconv"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

const bytecodeLambdaProgramCacheMaxEntries = 128

// bytecodeLambdaProgramCacheKey includes the exact lowering dependencies that
// are visible from the lambda's definition environment. Unrelated call-scope
// topology and ordinary value updates do not change immutable bytecode.
type bytecodeLambdaProgramCacheKey struct {
	expr                *ast.LambdaExpression
	bindingShapeStateID uint64
	dependencyShape     string
}

func canReuseCallableClosureEnvForBytecode(slotProgram *bytecodeProgram, needsRuntimeTypeBindings bool, closure *runtime.Environment) bool {
	if closure == nil || slotProgram == nil || slotProgram.frameLayout == nil {
		return false
	}
	if slotProgram.frameLayout.needsEnvScopes {
		return false
	}
	if needsRuntimeTypeBindings {
		return false
	}
	return true
}

func canReuseLambdaClosureEnvForBytecode(slotProgram *bytecodeProgram, decl *ast.LambdaExpression, call *ast.FunctionCall, closure *runtime.Environment) bool {
	if decl == nil {
		return false
	}
	return canReuseCallableClosureEnvForBytecode(slotProgram, callableNeedsExplicitRuntimeTypeBindings(decl), closure)
}

func lambdaSyntheticFunctionDefinition(expr *ast.LambdaExpression) *ast.FunctionDefinition {
	if expr == nil || expr.Body == nil {
		return nil
	}
	return ast.NewFunctionDefinition(
		nil,
		expr.Params,
		ast.NewBlockExpression([]ast.Statement{expr.Body}),
		expr.ReturnType,
		expr.GenericParams,
		expr.WhereClause,
		false,
		false,
	)
}

func (i *Interpreter) lowerLambdaExpressionBytecodeWithEnv(expr *ast.LambdaExpression, env *runtime.Environment) (*bytecodeProgram, error) {
	if expr == nil || expr.Body == nil {
		return nil, nil
	}
	key := i.bytecodeLambdaProgramCacheKeyForEnv(expr, env)
	if cached := i.lookupCachedLambdaBytecode(key); cached != nil {
		return cached, nil
	}
	program, err := i.lowerFunctionDefinitionBytecodeWithEnv(lambdaSyntheticFunctionDefinition(expr), env)
	if err != nil || program == nil {
		return program, err
	}
	// Complete the metadata before the program is published to the shared
	// cache. Later FunctionValue attachments only store their own pointer.
	setFunctionBytecodeProgram(&runtime.FunctionValue{Declaration: expr}, program)
	// Lowering can create temporary lexical scopes while it analyzes slot
	// eligibility. Recompute the key before publication so only the completed
	// definition environment's dependencies are recorded.
	return i.cacheLambdaBytecode(i.bytecodeLambdaProgramCacheKeyForEnv(expr, env), program), nil
}

func (i *Interpreter) bytecodeLambdaProgramCacheKeyForEnv(expr *ast.LambdaExpression, env *runtime.Environment) bytecodeLambdaProgramCacheKey {
	return bytecodeLambdaProgramCacheKey{
		expr:                expr,
		bindingShapeStateID: env.BindingShapeStateID(),
		dependencyShape:     i.bytecodeLambdaDependencyShape(expr, env),
	}
}

func (i *Interpreter) bytecodeLambdaDependencyShape(expr *ast.LambdaExpression, env *runtime.Environment) string {
	if expr == nil || env == nil {
		return ""
	}
	names := i.bytecodeLambdaDependencyNamesFor(expr)
	var buf strings.Builder
	buf.Grow(len(names) * 12)
	for _, name := range names {
		buf.WriteString(strconv.Itoa(len(name)))
		buf.WriteByte(':')
		buf.WriteString(name)
		if env.Has(name) {
			buf.WriteByte('1')
		} else {
			buf.WriteByte('0')
		}
		if def, ok := env.StructDefinition(name); ok && def != nil {
			buf.WriteByte('@')
			buf.WriteString(strconv.FormatUint(uint64(reflect.ValueOf(def).Pointer()), 16))
		}
		buf.WriteByte(';')
	}
	return buf.String()
}

func (i *Interpreter) bytecodeLambdaDependencyNamesFor(expr *ast.LambdaExpression) []string {
	if i == nil || expr == nil {
		return nil
	}
	i.bytecodeLambdaCacheMu.RLock()
	names, ok := i.bytecodeLambdaDependencyNames[expr]
	i.bytecodeLambdaCacheMu.RUnlock()
	if ok {
		return names
	}
	names = bytecodeLambdaDependencyNames(expr)
	i.bytecodeLambdaCacheMu.Lock()
	if i.bytecodeLambdaDependencyNames == nil {
		i.bytecodeLambdaDependencyNames = make(map[*ast.LambdaExpression][]string)
	}
	if existing, exists := i.bytecodeLambdaDependencyNames[expr]; exists {
		names = existing
	} else {
		i.bytecodeLambdaDependencyNames[expr] = names
	}
	i.bytecodeLambdaCacheMu.Unlock()
	return names
}

func bytecodeLambdaDependencyNames(expr *ast.LambdaExpression) []string {
	set := make(map[string]struct{})
	seen := make(map[uintptr]struct{})
	var visit func(reflect.Value)
	visit = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		if value.Kind() == reflect.Interface {
			if value.IsNil() {
				return
			}
			visit(value.Elem())
			return
		}
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return
			}
			pointer := value.Pointer()
			if _, ok := seen[pointer]; ok {
				return
			}
			seen[pointer] = struct{}{}
			if ident, ok := value.Interface().(*ast.Identifier); ok && ident.Name != "" {
				set[ident.Name] = struct{}{}
			}
			visit(value.Elem())
			return
		}
		switch value.Kind() {
		case reflect.Struct:
			typeInfo := value.Type()
			for idx := 0; idx < value.NumField(); idx++ {
				if typeInfo.Field(idx).PkgPath != "" {
					continue
				}
				visit(value.Field(idx))
			}
		case reflect.Slice, reflect.Array:
			for idx := 0; idx < value.Len(); idx++ {
				visit(value.Index(idx))
			}
		case reflect.Map:
			iter := value.MapRange()
			for iter.Next() {
				visit(iter.Key())
				visit(iter.Value())
			}
		}
	}
	visit(reflect.ValueOf(expr))
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (i *Interpreter) lookupCachedLambdaBytecode(key bytecodeLambdaProgramCacheKey) *bytecodeProgram {
	if i == nil || key.expr == nil {
		return nil
	}
	i.bytecodeLambdaCacheMu.RLock()
	program := i.bytecodeLambdaCache[key]
	i.bytecodeLambdaCacheMu.RUnlock()
	return program
}

func (i *Interpreter) cacheLambdaBytecode(key bytecodeLambdaProgramCacheKey, program *bytecodeProgram) *bytecodeProgram {
	if i == nil || key.expr == nil || program == nil {
		return program
	}
	i.bytecodeLambdaCacheMu.Lock()
	defer i.bytecodeLambdaCacheMu.Unlock()
	if i.bytecodeLambdaCache == nil {
		i.bytecodeLambdaCache = make(map[bytecodeLambdaProgramCacheKey]*bytecodeProgram, bytecodeLambdaProgramCacheMaxEntries)
	}
	if existing := i.bytecodeLambdaCache[key]; existing != nil {
		return existing
	}
	if len(i.bytecodeLambdaCache) >= bytecodeLambdaProgramCacheMaxEntries {
		// Bound retention from long-lived interpreters that repeatedly evaluate
		// the same source under newly shaped dynamic scopes.
		i.bytecodeLambdaCache = make(map[bytecodeLambdaProgramCacheKey]*bytecodeProgram, bytecodeLambdaProgramCacheMaxEntries)
		i.bytecodeLambdaDependencyNames = make(map[*ast.LambdaExpression][]string, bytecodeLambdaProgramCacheMaxEntries)
	}
	i.bytecodeLambdaCache[key] = program
	return program
}
