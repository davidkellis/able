package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (l *IRLowerer) pushIterator(binding string) {
	l.iteratorStack = append(l.iteratorStack, binding)
}

func (l *IRLowerer) popIterator() {
	if len(l.iteratorStack) == 0 {
		return
	}
	l.iteratorStack = l.iteratorStack[:len(l.iteratorStack)-1]
}

func (l *IRLowerer) currentIterator() (string, bool) {
	if len(l.iteratorStack) == 0 {
		return "", false
	}
	return l.iteratorStack[len(l.iteratorStack)-1], true
}

func (l *IRLowerer) lowerYield(stmt *ast.YieldStatement) error {
	if stmt == nil {
		return fmt.Errorf("compiler: yield statement is nil")
	}
	binding, ok := l.currentIterator()
	if !ok || binding == "" {
		return fmt.Errorf("compiler: yield may only appear inside iterator literal")
	}
	arg := stmt.Expression
	if arg == nil {
		arg = ast.NewNilLiteral()
	}
	callee := ast.NewMemberAccessExpression(ast.NewIdentifier(binding), ast.NewIdentifier("yield"))
	call := ast.NewFunctionCall(callee, []ast.Expression{arg}, nil, false)
	_, err := l.lowerExpr(call)
	return err
}

func (l *IRLowerer) lowerIteratorFunction(expr *ast.IteratorLiteral, binding string, captures []*IRSlot) (*IRFunction, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: iterator literal is nil")
	}
	l.iterCount++
	fn := &IRFunction{
		Name:       fmt.Sprintf("%s_iter_%d", l.function.Name, l.iterCount),
		Package:    l.function.Package,
		ReturnType: l.typeOf(expr),
		Locals:     nil,
		Blocks:     make(map[string]*IRBlock),
		EntryLabel: "entry",
		Source:     expr,
		Captured:   captures,
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
	// Scope for iterator bindings.
	sub.pushScope()
	genName := binding
	if genName == "" {
		genName = "gen"
	}
	genValue := sub.newValue(genName, nil, expr)
	fn.Params = append(fn.Params, IRParam{Value: genValue, Mutable: false})
	genSlot := sub.declareSlot(genName, nil, false, expr)
	if err := sub.current.Append(&IRStore{Slot: genSlot, Value: IRValueUse{Value: genValue}, Source: expr}); err != nil {
		return nil, err
	}
	if genName != "gen" {
		aliasSlot := sub.declareSlot("gen", nil, false, expr)
		if err := sub.current.Append(&IRStore{Slot: aliasSlot, Value: IRValueUse{Value: genValue}, Source: expr}); err != nil {
			return nil, err
		}
	}

	sub.pushIterator(genName)
	block := ast.NewBlockExpression(expr.Body)
	_, err := sub.lowerBlock(block)
	sub.popIterator()
	if err != nil {
		return nil, err
	}
	if sub.current.Terminator == nil {
		if err := sub.current.SetTerminator(&IRReturn{Value: IRConst{Literal: ast.NewNilLiteral(), TypeRef: nil}, Source: expr}); err != nil {
			return nil, err
		}
	}

	sub.popScope()
	sub.popScope()
	fn.Blocks = sub.blocks
	return fn, nil
}
