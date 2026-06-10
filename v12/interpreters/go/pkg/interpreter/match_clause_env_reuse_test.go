package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestMatchPatternForClauseTransientMultiBindingAllocatesLessWhenEligible(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)
	pattern := ast.ArrP(
		[]ast.Pattern{
			ast.PatternFrom("a"),
			ast.PatternFrom("b"),
			ast.PatternFrom("c"),
			ast.PatternFrom("d"),
			ast.PatternFrom("e"),
		},
		nil,
	)
	value := &runtime.ArrayValue{
		Elements: []runtime.Value{
			runtime.NewSmallInt(1, runtime.IntegerI32),
			runtime.NewSmallInt(2, runtime.IntegerI32),
			runtime.NewSmallInt(3, runtime.IntegerI32),
			runtime.NewSmallInt(4, runtime.IntegerI32),
			runtime.NewSmallInt(5, runtime.IntegerI32),
		},
	}
	eligiblePlan := clauseLocalScopePlan(pattern, nil, ast.Bin("+", ast.ID("a"), ast.ID("e")))
	ineligiblePlan := clauseLocalScopePlan(pattern, nil, ast.Lam(nil, ast.ID("a")))

	clauseEnv, matched, transientEnv, transientBindings := interp.matchPatternForClauseTransient(pattern, value, base, eligiblePlan)
	if !matched || clauseEnv == nil {
		t.Fatalf("expected eligible multi-binding clause to match during pool warmup")
	}
	interp.releaseTransientClauseMatch(transientEnv, transientBindings)

	eligibleAllocs := testing.AllocsPerRun(1000, func() {
		clauseEnv, matched, transientEnv, transientBindings := interp.matchPatternForClauseTransient(pattern, value, base, eligiblePlan)
		if !matched || clauseEnv == nil {
			panic("expected eligible multi-binding clause to match")
		}
		interp.releaseTransientClauseMatch(transientEnv, transientBindings)
	})
	ineligibleAllocs := testing.AllocsPerRun(1000, func() {
		clauseEnv, matched, transientEnv, transientBindings := interp.matchPatternForClauseTransient(pattern, value, base, ineligiblePlan)
		if !matched || clauseEnv == nil {
			panic("expected ineligible multi-binding clause to match")
		}
		interp.releaseTransientClauseMatch(transientEnv, transientBindings)
	})
	if eligibleAllocs >= ineligibleAllocs {
		t.Fatalf("expected eligible transient multi-binding clause path to allocate less than fallback path, got eligible=%.2f ineligible=%.2f", eligibleAllocs, ineligibleAllocs)
	}
}

func TestTransientClauseBindingBufferPoolHotPathStaysLowAllocation(t *testing.T) {
	interp := New()

	warm := interp.acquireTransientClauseBindingBuffer(4)
	warm.bindings = append(warm.bindings, patternBinding{Name: "x", Value: runtime.BoolValue{Val: true}})
	interp.releaseTransientClauseBindingBuffer(warm)

	allocs := testing.AllocsPerRun(1000, func() {
		buf := interp.acquireTransientClauseBindingBuffer(4)
		bindings := buf.bindings[:0]
		bindings = append(bindings,
			patternBinding{Name: "x", Value: runtime.BoolValue{Val: true}},
			patternBinding{Name: "y", Value: runtime.BoolValue{Val: false}},
		)
		buf.bindings = bindings
		interp.releaseTransientClauseBindingBuffer(buf)
	})
	if allocs >= 0.01 {
		t.Fatalf("expected pooled clause binding buffer hot path to stay near-zero allocation, got %.4f allocs/run", allocs)
	}
}

func TestMatchPatternForClauseTransientMultiBindingAllowsNonEscapingOrElse(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)
	pattern := ast.ArrP(
		[]ast.Pattern{
			ast.PatternFrom("a"),
			ast.PatternFrom("b"),
		},
		nil,
	)
	value := &runtime.ArrayValue{
		Elements: []runtime.Value{
			runtime.NewSmallInt(1, runtime.IntegerI32),
			runtime.NewSmallInt(2, runtime.IntegerI32),
		},
	}
	eligiblePlan := clauseLocalScopePlan(
		pattern,
		nil,
		ast.OrElse(ast.Prop(ast.Block(ast.Raise(ast.Str("x")))), nil, ast.ID("a")),
	)
	if !eligiblePlan.transientBindingSetOK {
		t.Fatalf("expected non-escaping or-else clause body to allow transient binding-set reuse")
	}
	ineligiblePlan := clauseLocalScopePlan(pattern, nil, ast.Lam(nil, ast.ID("a")))

	clauseEnv, matched, transientEnv, transientBindings := interp.matchPatternForClauseTransient(pattern, value, base, eligiblePlan)
	if !matched || clauseEnv == nil {
		t.Fatalf("expected eligible multi-binding or-else clause to match during pool warmup")
	}
	interp.releaseTransientClauseMatch(transientEnv, transientBindings)

	eligibleAllocs := testing.AllocsPerRun(1000, func() {
		clauseEnv, matched, transientEnv, transientBindings := interp.matchPatternForClauseTransient(pattern, value, base, eligiblePlan)
		if !matched || clauseEnv == nil {
			panic("expected eligible multi-binding or-else clause to match")
		}
		interp.releaseTransientClauseMatch(transientEnv, transientBindings)
	})
	ineligibleAllocs := testing.AllocsPerRun(1000, func() {
		clauseEnv, matched, transientEnv, transientBindings := interp.matchPatternForClauseTransient(pattern, value, base, ineligiblePlan)
		if !matched || clauseEnv == nil {
			panic("expected ineligible multi-binding clause to match")
		}
		interp.releaseTransientClauseMatch(transientEnv, transientBindings)
	})
	if eligibleAllocs >= ineligibleAllocs {
		t.Fatalf("expected non-escaping or-else clause path to allocate less than fallback path, got eligible=%.2f ineligible=%.2f", eligibleAllocs, ineligibleAllocs)
	}
}

func TestMatchPatternForClauseTransientBindinglessClauseAllocatesLessWhenEligible(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)
	pattern := ast.Wc()
	value := runtime.BoolValue{Val: true}
	eligiblePlan := clauseLocalScopePlan(
		pattern,
		nil,
		ast.Assign(ast.ID("local"), ast.Int(1)),
	)
	if !eligiblePlan.needsLocalScope || !eligiblePlan.transientScopeEnvOK {
		t.Fatalf("expected eligible bindingless clause to require local scope and allow transient scope reuse")
	}
	ineligiblePlan := clauseLocalScopePlan(
		pattern,
		nil,
		ast.Assign(ast.ID("local"), ast.Lam(nil, ast.Int(1))),
	)

	clauseEnv, matched, transientEnv, transientBindings := interp.matchPatternForClauseTransient(pattern, value, base, eligiblePlan)
	if !matched || clauseEnv == nil || transientEnv == nil || transientBindings != nil {
		t.Fatalf("expected eligible bindingless clause to match with transient scope reuse")
	}
	interp.releaseTransientClauseMatch(transientEnv, transientBindings)

	eligibleAllocs := testing.AllocsPerRun(1000, func() {
		clauseEnv, matched, transientEnv, transientBindings := interp.matchPatternForClauseTransient(pattern, value, base, eligiblePlan)
		if !matched || clauseEnv == nil {
			panic("expected eligible bindingless clause to match")
		}
		interp.releaseTransientClauseMatch(transientEnv, transientBindings)
	})
	ineligibleAllocs := testing.AllocsPerRun(1000, func() {
		clauseEnv, matched, transientEnv, transientBindings := interp.matchPatternForClauseTransient(pattern, value, base, ineligiblePlan)
		if !matched || clauseEnv == nil {
			panic("expected ineligible bindingless clause to match")
		}
		interp.releaseTransientClauseMatch(transientEnv, transientBindings)
	})
	if eligibleAllocs >= ineligibleAllocs {
		t.Fatalf("expected eligible bindingless clause path to allocate less than fallback path, got eligible=%.2f ineligible=%.2f", eligibleAllocs, ineligibleAllocs)
	}
}
