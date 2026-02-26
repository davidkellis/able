package wasmhost

import (
	"encoding/json"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/interpreter"
)

// EvaluateRequest describes a wasm-hosted AST execution request.
type EvaluateRequest struct {
	// ExecMode selects the runtime backend: "treewalker" (default) or "bytecode".
	ExecMode string `json:"execMode,omitempty"`
	// Setup contains optional module JSON payloads evaluated before Module.
	Setup []json.RawMessage `json:"setup,omitempty"`
	// Module is the entry module JSON payload (fixture AST format).
	Module json.RawMessage `json:"module"`
}

// EvaluateResponse describes the wasm-hosted AST execution result.
type EvaluateResponse struct {
	OK                   bool     `json:"ok"`
	Result               string   `json:"result,omitempty"`
	Error                string   `json:"error,omitempty"`
	TypecheckDiagnostics []string `json:"typecheckDiagnostics,omitempty"`
}

// EvaluateRequestJSON executes an EvaluateRequest and returns a JSON response.
func EvaluateRequestJSON(payload []byte) []byte {
	var req EvaluateRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return marshalResponse(EvaluateResponse{
			OK:    false,
			Error: fmt.Sprintf("decode request json: %v", err),
		})
	}
	return marshalResponse(Evaluate(req))
}

// Evaluate executes a request and returns a structured response.
func Evaluate(req EvaluateRequest) EvaluateResponse {
	interp, err := newInterpreter(req.ExecMode)
	if err != nil {
		return EvaluateResponse{OK: false, Error: err.Error()}
	}

	for idx, raw := range req.Setup {
		mod, err := interpreter.DecodeModule(raw)
		if err != nil {
			return EvaluateResponse{
				OK:    false,
				Error: fmt.Sprintf("decode setup module %d: %v", idx, err),
			}
		}
		if _, _, err := interp.EvaluateModule(mod); err != nil {
			return EvaluateResponse{
				OK:    false,
				Error: fmt.Sprintf("evaluate setup module %d: %v", idx, err),
			}
		}
	}

	if len(req.Module) == 0 {
		return EvaluateResponse{OK: false, Error: "missing module payload"}
	}

	entry, err := interpreter.DecodeModule(req.Module)
	if err != nil {
		return EvaluateResponse{OK: false, Error: fmt.Sprintf("decode module json: %v", err)}
	}
	value, env, err := interp.EvaluateModule(entry)
	if err != nil {
		return EvaluateResponse{OK: false, Error: fmt.Sprintf("evaluate module: %v", err)}
	}
	rendered, err := interp.Stringify(value, env)
	if err != nil {
		return EvaluateResponse{OK: false, Error: fmt.Sprintf("stringify result: %v", err)}
	}

	resp := EvaluateResponse{
		OK:     true,
		Result: rendered,
	}
	if diags := interp.TypecheckDiagnostics(); len(diags) > 0 {
		resp.TypecheckDiagnostics = make([]string, 0, len(diags))
		for _, diag := range diags {
			resp.TypecheckDiagnostics = append(resp.TypecheckDiagnostics, diag.Message)
		}
	}
	return resp
}

func newInterpreter(execMode string) (*interpreter.Interpreter, error) {
	switch strings.ToLower(strings.TrimSpace(execMode)) {
	case "", "treewalker":
		return interpreter.New(), nil
	case "bytecode":
		return interpreter.NewBytecode(), nil
	default:
		return nil, fmt.Errorf("unsupported execMode %q (expected treewalker or bytecode)", execMode)
	}
}

func marshalResponse(resp EvaluateResponse) []byte {
	encoded, err := json.Marshal(resp)
	if err == nil {
		return encoded
	}
	fallback := fmt.Sprintf(`{"ok":false,"error":"encode response json: %s"}`, sanitizeJSONError(err.Error()))
	return []byte(fallback)
}

func sanitizeJSONError(message string) string {
	message = strings.ReplaceAll(message, `\`, `\\`)
	message = strings.ReplaceAll(message, `"`, `\"`)
	return message
}
