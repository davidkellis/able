//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"able/interpreter-go/pkg/wasmhost"
)

var evalRequestFunc js.Func

func main() {
	evalRequestFunc = js.FuncOf(evalRequest)
	js.Global().Set("__able_eval_request_json", evalRequestFunc)

	select {}
}

func evalRequest(_ js.Value, args []js.Value) interface{} {
	output := jsHostOutput{}
	if len(args) == 0 {
		_ = output.WriteStderr("missing request JSON argument\n")
		return string(encodeError("missing request JSON argument"))
	}
	return string(wasmhost.EvaluateRequestJSONWithOutput([]byte(args[0].String()), output))
}

// jsHostOutput adapts the Go js/wasm runtime to the named able_host methods.
// The production ABI reserves pointer/length imports for non-Go embeddings;
// Go's js runtime owns that low-level import object, so this prototype passes
// UTF-8 strings through the same host method names instead.
type jsHostOutput struct{}

func (jsHostOutput) WriteStdout(message string) error {
	return callHostOutput("write_stdout", message)
}

func (jsHostOutput) WriteStderr(message string) error {
	return callHostOutput("write_stderr", message)
}

func callHostOutput(method string, message string) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("able_host.%s failed: %v", method, recovered)
		}
	}()
	host := js.Global().Get("able_host")
	if host.Type() != js.TypeObject {
		return fmt.Errorf("able_host.%s is unavailable", method)
	}
	callback := host.Get(method)
	if callback.Type() != js.TypeFunction {
		return fmt.Errorf("able_host.%s is unavailable", method)
	}
	callback.Invoke(message)
	return nil
}

func encodeError(message string) []byte {
	payload, err := json.Marshal(wasmhost.EvaluateResponse{
		OK:    false,
		Error: message,
	})
	if err == nil {
		return payload
	}
	return []byte(fmt.Sprintf(`{"ok":false,"error":"%s"}`, message))
}
