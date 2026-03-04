package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func (l *IRLowerer) lowerPatternAssignment(expr *ast.AssignmentExpression, pattern ast.Pattern, right IRValueRef) (IRValueRef, error) {
	l.declarePatternBindingsForAssignment(pattern, expr.Operator == ast.AssignmentDeclare)
	bindings, err := l.collectPatternBindings(pattern)
	if err != nil {
		return nil, err
	}
	errValue := l.newValue("assign_err", l.typeOf(expr), expr)
	if err := l.current.Append(&IRDestructure{
		Pattern:  pattern,
		Value:    right,
		Error:    errValue,
		Source:   expr,
		Bindings: bindings,
	}); err != nil {
		return nil, err
	}

	okBlock := l.newNamedBlock("assign_ok")
	errBlock := l.newNamedBlock("assign_err")
	joinBlock := l.newNamedBlock("assign_join")
	resultSlot := l.newSlot("assign_result", l.typeOf(expr), false, expr)

	if err := l.current.SetTerminator(&IRCheck{
		Error:    IRValueUse{Value: errValue},
		OkLabel:  okBlock.Label,
		ErrLabel: errBlock.Label,
	}); err != nil {
		return nil, err
	}

	l.current = okBlock
	if err := l.current.Append(&IRStore{Slot: resultSlot, Value: right, Source: expr}); err != nil {
		return nil, err
	}
	if err := l.current.SetTerminator(&IRJump{Target: joinBlock.Label}); err != nil {
		return nil, err
	}

	l.current = errBlock
	if err := l.current.Append(&IRStore{Slot: resultSlot, Value: IRValueUse{Value: errValue}, Source: expr}); err != nil {
		return nil, err
	}
	if err := l.current.SetTerminator(&IRJump{Target: joinBlock.Label}); err != nil {
		return nil, err
	}

	l.current = joinBlock
	result := l.newValue("assign_result", l.typeOf(expr), expr)
	if err := l.current.Append(&IRLoad{Dest: result, Slot: resultSlot, Source: expr}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: result}, nil
}

func (l *IRLowerer) declarePatternBindingsForAssignment(pattern ast.Pattern, forceDeclare bool) {
	if pattern == nil {
		return
	}
	switch p := pattern.(type) {
	case *ast.Identifier:
		l.declareBindingNameForAssignment(p.Name, l.typeOf(p), forceDeclare, p)
	case *ast.TypedPattern:
		l.declarePatternBindingsForAssignment(p.Pattern, forceDeclare)
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil {
				l.declareBindingNameForAssignment(field.Binding.Name, l.typeOf(field.Binding), forceDeclare, field)
			}
			l.declarePatternBindingsForAssignment(field.Pattern, forceDeclare)
		}
	case *ast.ArrayPattern:
		for _, elem := range p.Elements {
			l.declarePatternBindingsForAssignment(elem, forceDeclare)
		}
		if p.RestPattern != nil {
			l.declarePatternBindingsForAssignment(p.RestPattern, forceDeclare)
		}
	}
}

func (l *IRLowerer) declareBindingNameForAssignment(name string, typ typechecker.Type, forceDeclare bool, source ast.Node) {
	if name == "" {
		return
	}
	if forceDeclare {
		l.declareSlot(name, typ, true, source)
		return
	}
	if l.lookupSlot(name) == nil {
		l.declareSlot(name, typ, true, source)
	}
}

func (l *IRLowerer) collectPatternBindings(pattern ast.Pattern) (map[ast.Node]*IRSlot, error) {
	if pattern == nil {
		return nil, nil
	}
	bindings := make(map[ast.Node]*IRSlot)
	var walk func(pat ast.Pattern) error
	walk = func(pat ast.Pattern) error {
		if pat == nil {
			return nil
		}
		switch p := pat.(type) {
		case *ast.Identifier:
			if p.Name == "" || p.Name == "_" {
				return nil
			}
			slot := l.lookupSlot(p.Name)
			if slot == nil {
				return fmt.Errorf("compiler: missing slot for pattern %s", p.Name)
			}
			bindings[p] = slot
		case *ast.TypedPattern:
			return walk(p.Pattern)
		case *ast.StructPattern:
			for _, field := range p.Fields {
				if field == nil {
					continue
				}
				if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
					slot := l.lookupSlot(field.Binding.Name)
					if slot == nil {
						return fmt.Errorf("compiler: missing slot for pattern %s", field.Binding.Name)
					}
					bindings[field] = slot
				}
				if err := walk(field.Pattern); err != nil {
					return err
				}
			}
		case *ast.ArrayPattern:
			for _, elem := range p.Elements {
				if err := walk(elem); err != nil {
					return err
				}
			}
			if p.RestPattern != nil {
				switch rest := p.RestPattern.(type) {
				case *ast.Identifier:
					if rest.Name == "" || rest.Name == "_" {
						return nil
					}
					slot := l.lookupSlot(rest.Name)
					if slot == nil {
						return fmt.Errorf("compiler: missing slot for pattern %s", rest.Name)
					}
					bindings[rest] = slot
				case *ast.WildcardPattern:
					return nil
				default:
					if err := walk(rest); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	if err := walk(pattern); err != nil {
		return nil, err
	}
	return bindings, nil
}

func (l *IRLowerer) lowerCall(call *ast.FunctionCall) (IRValueRef, error) {
	if call == nil {
		return nil, fmt.Errorf("compiler: call is nil")
	}
	callee, err := l.lowerExpr(call.Callee)
	if err != nil {
		return nil, err
	}
	args := make([]IRValueRef, 0, len(call.Arguments))
	for _, arg := range call.Arguments {
		val, err := l.lowerExpr(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}
	return l.emitInvoke(IROpCall, callee, args, call)
}

func (l *IRLowerer) emitCompute(op IROp, operator string, args []IRValueRef, source ast.Node) (IRValueRef, error) {
	value := l.newValue("tmp", l.typeOf(source), source)
	if err := l.current.Append(&IRCompute{Dest: value, Op: op, Operator: operator, Args: args, Source: source}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) emitInvoke(op IROp, callee IRValueRef, args []IRValueRef, source ast.Node) (IRValueRef, error) {
	value := l.newValue("tmp", l.typeOf(source), source)
	errValue := l.newValue("err", l.typeOf(source), source)
	if err := l.current.Append(&IRInvoke{
		Value:  value,
		Error:  errValue,
		Op:     op,
		Callee: callee,
		Args:   args,
		Source: source,
	}); err != nil {
		return nil, err
	}
	okBlock := l.newNamedBlock("invoke_ok")
	errBlock := l.ensureErrorBlock()
	if handler, ok := l.currentErrorHandler(); ok {
		errBlock = l.blocks[handler.label]
		if errBlock == nil {
			return nil, fmt.Errorf("compiler: missing error handler block %s", handler.label)
		}
		if handler.slot != nil {
			if err := l.current.Append(&IRStore{Slot: handler.slot, Value: IRValueUse{Value: errValue}, Source: source}); err != nil {
				return nil, err
			}
		}
	}
	if err := l.current.SetTerminator(&IRCheck{
		Error:    IRValueUse{Value: errValue},
		OkLabel:  okBlock.Label,
		ErrLabel: errBlock.Label,
	}); err != nil {
		return nil, err
	}
	l.current = okBlock
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) ensureErrorBlock() *IRBlock {
	if l.errorBlock != nil {
		return l.errorBlock
	}
	block := l.newNamedBlock("error")
	_ = block.SetTerminator(&IRUnreachable{})
	l.errorBlock = block
	return block
}

func (l *IRLowerer) loadIdentifier(id *ast.Identifier) (IRValueRef, error) {
	if id == nil {
		return nil, fmt.Errorf("compiler: identifier is nil")
	}
	slot := l.lookupSlot(id.Name)
	if slot == nil {
		return IRGlobal{Name: id.Name, TypeRef: l.typeOf(id)}, nil
	}
	value := l.newValue(id.Name, l.typeOf(id), id)
	if err := l.current.Append(&IRLoad{Dest: value, Slot: slot, Source: id}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) paramName(pattern ast.Pattern) (string, bool) {
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p.Name == "" {
			return "", false
		}
		return p.Name, true
	case *ast.TypedPattern:
		return l.paramName(p.Pattern)
	default:
		return "", false
	}
}

func (l *IRLowerer) declarePatternBindings(pattern ast.Pattern, allowDeclare bool) {
	if pattern == nil {
		return
	}
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p.Name == "" {
			return
		}
		if allowDeclare {
			l.declareSlot(p.Name, l.typeOf(p), true, p)
		}
	case *ast.TypedPattern:
		l.declarePatternBindings(p.Pattern, allowDeclare)
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil && field.Binding.Name != "" && allowDeclare {
				l.declareSlot(field.Binding.Name, l.typeOf(field.Binding), true, field)
			}
			l.declarePatternBindings(field.Pattern, allowDeclare)
		}
	case *ast.ArrayPattern:
		for _, elem := range p.Elements {
			l.declarePatternBindings(elem, allowDeclare)
		}
		if p.RestPattern != nil {
			l.declarePatternBindings(p.RestPattern, allowDeclare)
		}
	}
}

func (l *IRLowerer) pushScope() {
	l.scopes = append(l.scopes, make(map[string]*IRSlot))
}

func (l *IRLowerer) popScope() {
	if len(l.scopes) == 0 {
		return
	}
	l.scopes = l.scopes[:len(l.scopes)-1]
}

func (l *IRLowerer) lookupSlot(name string) *IRSlot {
	for i := len(l.scopes) - 1; i >= 0; i-- {
		if slot, ok := l.scopes[i][name]; ok {
			return slot
		}
	}
	return nil
}

func (l *IRLowerer) declareSlot(name string, typ typechecker.Type, mutable bool, source ast.Node) *IRSlot {
	if name == "" {
		return nil
	}
	scope := l.scopes[len(l.scopes)-1]
	if slot, ok := scope[name]; ok {
		return slot
	}
	slot := &IRSlot{
		Name:    name,
		Type:    typ,
		Mutable: mutable,
		Source:  source,
	}
	scope[name] = slot
	l.function.Locals = append(l.function.Locals, slot)
	return slot
}

func (l *IRLowerer) newValue(name string, typ typechecker.Type, source ast.Node) *IRValue {
	l.nextValueID++
	return &IRValue{
		ID:     l.nextValueID,
		Name:   name,
		Type:   typ,
		Source: source,
	}
}

func (l *IRLowerer) newSlot(name string, typ typechecker.Type, mutable bool, source ast.Node) *IRSlot {
	return l.declareSlot(fmt.Sprintf("%s_%d", name, len(l.function.Locals)), typ, mutable, source)
}

func (l *IRLowerer) newNamedBlock(label string) *IRBlock {
	if _, ok := l.blocks[label]; ok {
		l.nextBlockID++
		label = fmt.Sprintf("%s_%d", label, l.nextBlockID)
	}
	block := &IRBlock{Label: label}
	l.blocks[label] = block
	return block
}

func (l *IRLowerer) pushLoop(ctx loopContext) {
	l.loopStack = append(l.loopStack, ctx)
}

func (l *IRLowerer) popLoop() {
	if len(l.loopStack) == 0 {
		return
	}
	l.loopStack = l.loopStack[:len(l.loopStack)-1]
}

func (l *IRLowerer) currentLoop() (loopContext, bool) {
	if len(l.loopStack) == 0 {
		return loopContext{}, false
	}
	return l.loopStack[len(l.loopStack)-1], true
}

func (l *IRLowerer) pushBreakpoint(ctx breakpointContext) {
	l.breakpoints = append(l.breakpoints, ctx)
}

func (l *IRLowerer) popBreakpoint() {
	if len(l.breakpoints) == 0 {
		return
	}
	l.breakpoints = l.breakpoints[:len(l.breakpoints)-1]
}

func (l *IRLowerer) lookupBreakpoint(label string) (breakpointContext, bool) {
	if label == "" {
		return breakpointContext{}, false
	}
	for i := len(l.breakpoints) - 1; i >= 0; i-- {
		if l.breakpoints[i].label == label {
			return l.breakpoints[i], true
		}
	}
	return breakpointContext{}, false
}

func (l *IRLowerer) pushErrorHandler(label string, slot *IRSlot) {
	l.errorStack = append(l.errorStack, errorContext{label: label, slot: slot})
}

func (l *IRLowerer) popErrorHandler() {
	if len(l.errorStack) == 0 {
		return
	}
	l.errorStack = l.errorStack[:len(l.errorStack)-1]
}

func (l *IRLowerer) currentErrorHandler() (errorContext, bool) {
	if len(l.errorStack) == 0 {
		return errorContext{}, false
	}
	return l.errorStack[len(l.errorStack)-1], true
}

func (l *IRLowerer) propagateError(errValue IRValueRef, source ast.Node) error {
	if handler, ok := l.currentErrorHandler(); ok {
		if handler.slot != nil {
			if err := l.current.Append(&IRStore{Slot: handler.slot, Value: errValue, Source: source}); err != nil {
				return err
			}
		}
		return l.current.SetTerminator(&IRJump{Target: handler.label})
	}
	return l.current.SetTerminator(&IRJump{Target: l.ensureErrorBlock().Label})
}

func (l *IRLowerer) typeOf(node ast.Node) typechecker.Type {
	if l.types == nil || node == nil {
		return nil
	}
	if typ, ok := l.types[node]; ok {
		return typ
	}
	return nil
}
