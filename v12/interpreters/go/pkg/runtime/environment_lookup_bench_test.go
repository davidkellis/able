package runtime

import "testing"

var (
	benchmarkEnvironmentLookupValueSink   Value
	benchmarkEnvironmentLookupOwnerSink   *Environment
	benchmarkEnvironmentLookupVersionSink uint64
	benchmarkEnvironmentLookupOKSink      bool
)

var benchmarkEnvironmentLookupNames = [...]string{
	"alpha",
	"beta",
	"gamma",
	"delta",
	"epsilon",
	"zeta",
	"eta",
}

func benchmarkEnvironmentLookupSetup(b *testing.B, valueCapacity int, bindingCount int) (*Environment, string, Value) {
	b.Helper()

	if bindingCount <= 0 || bindingCount > len(benchmarkEnvironmentLookupNames) {
		b.Fatalf("invalid binding count %d", bindingCount)
	}

	env := NewEnvironmentWithValueCapacity(nil, valueCapacity)
	env.SetSingleThread()

	var targetName string
	var targetValue Value
	for idx := 0; idx < bindingCount; idx++ {
		targetName = benchmarkEnvironmentLookupNames[idx]
		targetValue = NewSmallInt(int64(idx+1), IntegerI32)
		env.DefineWithoutMerge(targetName, targetValue)
	}

	return env, targetName, targetValue
}

func BenchmarkEnvironmentLookupCurrentValueNoLock(b *testing.B) {
	cases := []struct {
		name          string
		valueCapacity int
		bindingCount  int
		verify        func(*testing.B, *Environment)
	}{
		{
			name:          "inline1",
			valueCapacity: 0,
			bindingCount:  1,
			verify: func(b *testing.B, env *Environment) {
				if env.values != nil || env.spill != nil || env.inlineCount != 1 {
					b.Fatalf("expected inline1 shape, got values=%t spill=%t inline=%d", env.values != nil, env.spill != nil, env.inlineCount)
				}
			},
		},
		{
			name:          "inline2",
			valueCapacity: 0,
			bindingCount:  2,
			verify: func(b *testing.B, env *Environment) {
				if env.values != nil || env.spill != nil || env.inlineCount != 2 {
					b.Fatalf("expected inline2 shape, got values=%t spill=%t inline=%d", env.values != nil, env.spill != nil, env.inlineCount)
				}
			},
		},
		{
			name:          "inline4",
			valueCapacity: 0,
			bindingCount:  4,
			verify: func(b *testing.B, env *Environment) {
				if env.values != nil || env.spill != nil || env.inlineCount != 4 {
					b.Fatalf("expected inline4 shape, got values=%t spill=%t inline=%d", env.values != nil, env.spill != nil, env.inlineCount)
				}
			},
		},
		{
			name:          "spill5",
			valueCapacity: 5,
			bindingCount:  5,
			verify: func(b *testing.B, env *Environment) {
				if env.values != nil || env.spill == nil || env.spill.count != 5 {
					count := uint8(0)
					if env.spill != nil {
						count = env.spill.count
					}
					b.Fatalf("expected spill5 shape, got values=%t spill=%t spill_count=%d", env.values != nil, env.spill != nil, count)
				}
			},
		},
		{
			name:          "map5",
			valueCapacity: 7,
			bindingCount:  5,
			verify: func(b *testing.B, env *Environment) {
				if env.values == nil || env.spill != nil || len(env.values) != 5 {
					size := 0
					if env.values != nil {
						size = len(env.values)
					}
					b.Fatalf("expected map5 shape, got values=%t value_count=%d spill=%t", env.values != nil, size, env.spill != nil)
				}
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			env, targetName, targetValue := benchmarkEnvironmentLookupSetup(b, tc.valueCapacity, tc.bindingCount)
			tc.verify(b, env)

			got, ok := env.lookupCurrentValueNoLock(targetName)
			if !ok || got != targetValue {
				b.Fatalf("lookupCurrentValueNoLock(%q) = (%#v, %t), want (%#v, true)", targetName, got, ok, targetValue)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for idx := 0; idx < b.N; idx++ {
				benchmarkEnvironmentLookupValueSink, benchmarkEnvironmentLookupOKSink = env.lookupCurrentValueNoLock(targetName)
			}
		})
	}
}

func benchmarkEnvironmentLookupChain(singleThread bool) (*Environment, *Environment, *Environment, Value, Value, Value) {
	rootValue := NewSmallInt(1, IntegerI32)
	parentValue := NewSmallInt(2, IntegerI32)
	childValue := NewSmallInt(3, IntegerI32)

	root := NewEnvironment(nil)
	parent := NewEnvironment(root)
	child := NewEnvironment(parent)
	if singleThread {
		child.SetSingleThread()
	}

	root.DefineWithoutMerge("root", rootValue)
	parent.DefineWithoutMerge("outer", parentValue)
	child.DefineWithoutMerge("inner", childValue)

	return root, parent, child, rootValue, parentValue, childValue
}

func BenchmarkEnvironmentLookupWithOwnerAndRevisionHint(b *testing.B) {
	cases := []struct {
		name         string
		singleThread bool
		targetName   string
		wantOwner    func(root *Environment, parent *Environment, child *Environment) *Environment
		wantValue    func(root Value, parent Value, child Value) Value
	}{
		{
			name:         "current_single_thread",
			singleThread: true,
			targetName:   "inner",
			wantOwner:    func(_ *Environment, _ *Environment, child *Environment) *Environment { return child },
			wantValue:    func(_ Value, _ Value, child Value) Value { return child },
		},
		{
			name:         "parent_single_thread",
			singleThread: true,
			targetName:   "outer",
			wantOwner:    func(_ *Environment, parent *Environment, _ *Environment) *Environment { return parent },
			wantValue:    func(_ Value, parent Value, _ Value) Value { return parent },
		},
		{
			name:         "grandparent_single_thread",
			singleThread: true,
			targetName:   "root",
			wantOwner:    func(root *Environment, _ *Environment, _ *Environment) *Environment { return root },
			wantValue:    func(root Value, _ Value, _ Value) Value { return root },
		},
		{
			name:         "current_multi_thread",
			singleThread: false,
			targetName:   "inner",
			wantOwner:    func(_ *Environment, _ *Environment, child *Environment) *Environment { return child },
			wantValue:    func(_ Value, _ Value, child Value) Value { return child },
		},
		{
			name:         "parent_multi_thread",
			singleThread: false,
			targetName:   "outer",
			wantOwner:    func(_ *Environment, parent *Environment, _ *Environment) *Environment { return parent },
			wantValue:    func(_ Value, parent Value, _ Value) Value { return parent },
		},
		{
			name:         "grandparent_multi_thread",
			singleThread: false,
			targetName:   "root",
			wantOwner:    func(root *Environment, _ *Environment, _ *Environment) *Environment { return root },
			wantValue:    func(root Value, _ Value, _ Value) Value { return root },
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			root, parent, child, rootValue, parentValue, childValue := benchmarkEnvironmentLookupChain(tc.singleThread)
			wantOwner := tc.wantOwner(root, parent, child)
			wantValue := tc.wantValue(rootValue, parentValue, childValue)

			got, owner, version, ok := child.LookupWithOwnerAndRevisionHint(tc.targetName, tc.singleThread)
			if !ok || got != wantValue || owner != wantOwner || version != wantOwner.RevisionWithHint(tc.singleThread) {
				b.Fatalf(
					"LookupWithOwnerAndRevisionHint(%q, %t) = (%#v, %p, %d, %t), want (%#v, %p, %d, true)",
					tc.targetName,
					tc.singleThread,
					got,
					owner,
					version,
					ok,
					wantValue,
					wantOwner,
					wantOwner.RevisionWithHint(tc.singleThread),
				)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for idx := 0; idx < b.N; idx++ {
				benchmarkEnvironmentLookupValueSink, benchmarkEnvironmentLookupOwnerSink, benchmarkEnvironmentLookupVersionSink, benchmarkEnvironmentLookupOKSink =
					child.LookupWithOwnerAndRevisionHint(tc.targetName, tc.singleThread)
			}
		})
	}
}
