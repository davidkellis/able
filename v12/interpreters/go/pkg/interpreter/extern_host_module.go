package interpreter

import (
	"fmt"
	"reflect"
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
