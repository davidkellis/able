package compiler

import "bytes"

func (g *generator) renderRuntimeFutureDeclarations(buf *bytes.Buffer) {
	buf.WriteString(`
var __able_future_error_def *runtime.StructDefinitionValue
var __able_future_status_defs map[string]*runtime.StructDefinitionValue
var __able_future_status_pending runtime.Value
var __able_future_status_resolved runtime.Value
var __able_future_status_cancelled runtime.Value
var __able_future_status_once sync.Once
var __able_future_defs_mu sync.Mutex

type __able_compiled_yield struct {
	result runtime.Value
	err    error
	done   bool
}
`)
}
