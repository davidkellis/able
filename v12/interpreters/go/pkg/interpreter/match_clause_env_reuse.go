package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

const transientClauseBindingPoolMaxCapacity = 64

type transientClauseBindingBuffer struct {
	bindings []runtime.EnvironmentBinding
}

func singletonStructPatternValue(name string, value runtime.Value, base *runtime.Environment) bool {
	if name == "" || base == nil {
		return false
	}
	existing, ok := base.Lookup(name)
	if !ok {
		return false
	}
	switch defVal := existing.(type) {
	case *runtime.StructDefinitionValue:
		return defVal != nil && isSingletonStructDef(defVal.Node) && valuesEqual(existing, value)
	case runtime.StructDefinitionValue:
		return isSingletonStructDef(defVal.Node) && valuesEqual(existing, value)
	default:
		return false
	}
}

func (i *Interpreter) acquireTransientClauseEnv(base *runtime.Environment, valueCapacity int, name string, value runtime.Value) *runtime.Environment {
	if i == nil {
		return runtime.NewEnvironmentWithSingleBinding(base, valueCapacity, name, value)
	}
	if pooled := i.transientClauseEnvPool.Get(); pooled != nil {
		if env, ok := pooled.(*runtime.Environment); ok && env != nil {
			env.ResetForSingleBindingReuse(base, valueCapacity, name, value)
			return env
		}
	}
	return runtime.NewEnvironmentWithSingleBinding(base, valueCapacity, name, value)
}

func (i *Interpreter) releaseTransientClauseEnv(env *runtime.Environment) {
	if i == nil || env == nil {
		return
	}
	i.transientClauseEnvPool.Put(env)
}

func (i *Interpreter) acquireTransientClauseEnvForBindingSets(base *runtime.Environment, valueCapacity int, bindings []runtime.EnvironmentBinding) *runtime.Environment {
	if i == nil {
		return runtime.NewEnvironmentWithBindingSets(base, valueCapacity, bindings, nil)
	}
	if pooled := i.transientClauseEnvPool.Get(); pooled != nil {
		if env, ok := pooled.(*runtime.Environment); ok && env != nil {
			env.ResetForBindingSetsReuse(base, valueCapacity, bindings, nil)
			return env
		}
	}
	return runtime.NewEnvironmentWithBindingSets(base, valueCapacity, bindings, nil)
}

func (i *Interpreter) acquireTransientClauseBindingBuffer(minCapacity int) *transientClauseBindingBuffer {
	if minCapacity < 1 {
		minCapacity = 1
	}
	if i == nil {
		return &transientClauseBindingBuffer{
			bindings: make([]runtime.EnvironmentBinding, 0, minCapacity),
		}
	}
	if pooled := i.transientClauseBindingPool.Get(); pooled != nil {
		if buf, ok := pooled.(*transientClauseBindingBuffer); ok && buf != nil {
			if cap(buf.bindings) < minCapacity {
				buf.bindings = make([]runtime.EnvironmentBinding, 0, minCapacity)
			} else {
				buf.bindings = buf.bindings[:0]
			}
			return buf
		}
	}
	return &transientClauseBindingBuffer{
		bindings: make([]runtime.EnvironmentBinding, 0, minCapacity),
	}
}

func (i *Interpreter) releaseTransientClauseBindingBuffer(buf *transientClauseBindingBuffer) {
	if i == nil || buf == nil {
		return
	}
	bindings := buf.bindings
	if bindings == nil {
		i.transientClauseBindingPool.Put(buf)
		return
	}
	if cap(bindings) > transientClauseBindingPoolMaxCapacity {
		buf.bindings = nil
		i.transientClauseBindingPool.Put(buf)
		return
	}
	full := bindings[:cap(bindings)]
	clear(full)
	buf.bindings = full[:0]
	i.transientClauseBindingPool.Put(buf)
}

func (i *Interpreter) releaseTransientClauseMatch(env *runtime.Environment, bindings *transientClauseBindingBuffer) {
	i.releaseTransientClauseEnv(env)
	i.releaseTransientClauseBindingBuffer(bindings)
}

func (i *Interpreter) bindinglessTransientClauseEnv(base *runtime.Environment, plan clauseScopePlan) (*runtime.Environment, *runtime.Environment) {
	if !plan.needsLocalScope {
		return base, nil
	}
	if !plan.transientScopeEnvOK {
		return bindinglessClauseScopeEnv(base, plan), nil
	}
	env := i.acquireTransientClauseEnvForBindingSets(base, plan.localBindingCapacity, nil)
	return env, env
}

func (i *Interpreter) tryMatchPatternForTransientSingleBindingClause(pattern ast.Pattern, value runtime.Value, base *runtime.Environment, plan clauseScopePlan) (*runtime.Environment, bool, bool, *runtime.Environment) {
	if !plan.transientSingleBindOK || pattern == nil || base == nil {
		return nil, false, false, nil
	}
	value = bytecodeMaterializeRawValue(bytecodeSlotReadValue(value))
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p == nil || p.Name == "" || p.Name == "_" {
			return nil, false, false, nil
		}
		if singletonStructPatternValue(p.Name, value, base) {
			env, transientEnv := i.bindinglessTransientClauseEnv(base, plan)
			return env, true, true, transientEnv
		}
		env := i.acquireTransientClauseEnv(base, plan.localBindingCapacity, p.Name, value)
		return env, true, true, env
	case *ast.TypedPattern:
		if p == nil {
			return nil, false, false, nil
		}
		coerced, ok := i.matchTypedPatternValueInEnv(p.TypeAnnotation, value, base)
		if !ok {
			return nil, false, true, nil
		}
		return i.tryMatchPatternForTransientSingleBindingClause(p.Pattern, coerced, base, plan)
	default:
		return nil, false, false, nil
	}
}

func (i *Interpreter) tryMatchPatternForTransientBindingSetClause(pattern ast.Pattern, value runtime.Value, base *runtime.Environment, plan clauseScopePlan) (*runtime.Environment, bool, bool, *runtime.Environment, *transientClauseBindingBuffer) {
	if !plan.transientBindingSetOK || pattern == nil || base == nil {
		return nil, false, false, nil, nil
	}
	patternCapacity := plan.patternBindingCount
	if patternCapacity == 0 {
		return nil, false, false, nil, nil
	}
	value = bytecodeMaterializeRawValue(bytecodeSlotReadValue(value))
	buf := i.acquireTransientClauseBindingBuffer(patternCapacity)
	bindings := buf.bindings[:0]
	var err error
	bindings, err = i.collectPatternBindings(pattern, value, base, bindings)
	if err != nil {
		buf.bindings = bindings
		i.releaseTransientClauseBindingBuffer(buf)
		return nil, false, true, nil, nil
	}
	if len(bindings) == 0 {
		buf.bindings = bindings
		i.releaseTransientClauseBindingBuffer(buf)
		env, transientEnv := i.bindinglessTransientClauseEnv(base, plan)
		return env, true, true, transientEnv, nil
	}
	buf.bindings = bindings
	env := i.acquireTransientClauseEnvForBindingSets(base, plan.localBindingCapacity, bindings)
	return env, true, true, env, buf
}

func (i *Interpreter) matchPatternForClauseTransient(pattern ast.Pattern, value runtime.Value, base *runtime.Environment, plan clauseScopePlan) (*runtime.Environment, bool, *runtime.Environment, *transientClauseBindingBuffer) {
	if clauseEnv, matched, handled, transientEnv := i.tryMatchPatternForTransientSingleBindingClause(pattern, value, base, plan); handled {
		return clauseEnv, matched, transientEnv, nil
	}
	if clauseEnv, matched, handled, transientEnv, bindings := i.tryMatchPatternForTransientBindingSetClause(pattern, value, base, plan); handled {
		return clauseEnv, matched, transientEnv, bindings
	}
	if matchEnv, matched, handled := i.matchPatternFastWithScopeReuse(pattern, value, base, plan.capturePatternBinding, true, plan.localBindingCapacity); handled {
		if !matched {
			return nil, false, nil, nil
		}
		if matchEnv == base {
			clauseEnv, transientEnv := i.bindinglessTransientClauseEnv(base, plan)
			return clauseEnv, true, transientEnv, nil
		}
		return matchEnv, true, nil, nil
	}
	matchEnv, err := i.matchPatternIntoClauseEnv(pattern, value, base, nil, plan.capturePatternBinding, plan.localBindingCapacity)
	if err != nil {
		return nil, false, nil, nil
	}
	if matchEnv == nil {
		clauseEnv, transientEnv := i.bindinglessTransientClauseEnv(base, plan)
		return clauseEnv, true, transientEnv, nil
	}
	return matchEnv, true, nil, nil
}
