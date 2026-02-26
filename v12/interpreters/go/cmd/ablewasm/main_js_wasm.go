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
	if len(args) == 0 {
		return string(encodeError("missing request JSON argument"))
	}
	return string(wasmhost.EvaluateRequestJSON([]byte(args[0].String())))
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
