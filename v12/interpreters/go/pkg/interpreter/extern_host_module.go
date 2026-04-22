package interpreter

import (
	"fmt"
	"reflect"

	"able/interpreter-go/pkg/ast"
)

func (m *externHostModule) lookup(name string) (reflect.Value, error) {
	if m == nil || m.plugin == nil {
		return reflect.Value{}, fmt.Errorf("extern host module not initialized")
	}
	if name == "" {
		return reflect.Value{}, fmt.Errorf("extern function name is empty")
	}
	if m.symbols == nil {
		m.symbols = make(map[string]reflect.Value)
	}
	if sym, ok := m.symbols[name]; ok {
		return sym, nil
	}
	symbolName := externSymbolName(name)
	raw, err := m.plugin.Lookup(symbolName)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("extern lookup %s: %w", name, err)
	}
	fn := reflect.ValueOf(raw)
	if fn.Kind() != reflect.Func {
		return reflect.Value{}, fmt.Errorf("extern symbol %s is not a function", name)
	}
	m.symbols[name] = fn
	return fn, nil
}

func (m *externHostModule) lookupInvoker(def *ast.ExternFunctionBody) (externHostInvoker, error) {
	if m == nil || def == nil || def.Signature == nil || def.Signature.ID == nil {
		return nil, fmt.Errorf("extern invoker missing signature")
	}
	name := def.Signature.ID.Name
	if name == "" {
		return nil, fmt.Errorf("extern function name is empty")
	}
	if m.invokers == nil {
		m.invokers = make(map[string]externHostInvoker)
	}
	if invoker, ok := m.invokers[name]; ok {
		return invoker, nil
	}
	fn, err := m.lookup(name)
	if err != nil {
		return nil, err
	}
	raw := fn.Interface()
	invoker := buildExternFastInvoker(def, raw)
	m.invokers[name] = invoker
	return invoker, nil
}
