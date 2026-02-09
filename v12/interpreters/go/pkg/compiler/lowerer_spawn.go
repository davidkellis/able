package compiler

import (
	"fmt"
	"sort"

	"able/interpreter-go/pkg/ast"
)

func (l *IRLowerer) captureSlots() []*IRSlot {
	seen := make(map[string]bool)
	captured := make([]*IRSlot, 0)
	for i := len(l.scopes) - 1; i >= 0; i-- {
		scope := l.scopes[i]
		names := make([]string, 0, len(scope))
		for name := range scope {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if seen[name] {
				continue
			}
			seen[name] = true
			if slot := scope[name]; slot != nil {
				captured = append(captured, slot)
			}
		}
	}
	return captured
}

func (l *IRLowerer) lowerSpawnFunction(expr ast.Expression, captures []*IRSlot) (*IRFunction, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: spawn body is nil")
	}
	l.spawnCount++
	fn := &IRFunction{
		Name:       fmt.Sprintf("%s_spawn_%d", l.function.Name, l.spawnCount),
		Package:    l.function.Package,
		ReturnType: l.typeOf(expr),
		Locals:     nil,
		Blocks:     make(map[string]*IRBlock),
		EntryLabel: "entry",
		Source:     expr,
	}
	sub := &IRLowerer{
		types:  l.types,
		blocks: make(map[string]*IRBlock),
	}
	entry := sub.newNamedBlock("entry")
	sub.function = fn
	sub.current = entry

	sub.pushScope()
	if len(captures) > 0 {
		scope := sub.scopes[len(sub.scopes)-1]
		for _, slot := range captures {
			if slot == nil || slot.Name == "" {
				continue
			}
			scope[slot.Name] = slot
		}
	}
	sub.pushScope()
	for _, slot := range captures {
		if slot == nil || slot.Name == "" {
			continue
		}
		fn.Captured = append(fn.Captured, slot)
	}
	value, err := sub.lowerExpr(expr)
	if err != nil {
		return nil, err
	}
	if sub.current.Terminator == nil {
		if err := sub.current.SetTerminator(&IRReturn{Value: value, Source: expr}); err != nil {
			return nil, err
		}
	}
	sub.popScope()
	sub.popScope()
	fn.Blocks = sub.blocks
	return fn, nil
}
