package wasmhost

import (
	"encoding/json"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

// EvaluateRequest describes a wasm-hosted AST execution request.
type EvaluateRequest struct {
	// ExecMode selects the runtime backend: "treewalker" (default) or "bytecode".
	ExecMode string `json:"execMode,omitempty"`
	// Setup contains optional module JSON payloads evaluated before Module.
	// It is retained for compatibility with the first AST bridge.
	Setup []json.RawMessage `json:"setup,omitempty"`
	// SetupModules carries ordered setup modules with their host source origins.
	SetupModules []SetupModule `json:"setupModules,omitempty"`
	// Module is the entry module JSON payload (fixture AST format).
	Module json.RawMessage `json:"module"`
	// EntryOrigin identifies the entry source in host diagnostics when available.
	EntryOrigin string `json:"entryOrigin,omitempty"`
}

// SetupModule is one ordered, host-supplied dependency AST module.
type SetupModule struct {
	Origin string          `json:"origin,omitempty"`
	Module json.RawMessage `json:"module"`
}

// EvaluateResponse describes the wasm-hosted AST execution result.
type EvaluateResponse struct {
	OK                   bool     `json:"ok"`
	Result               string   `json:"result,omitempty"`
	Error                string   `json:"error,omitempty"`
	TypecheckDiagnostics []string `json:"typecheckDiagnostics,omitempty"`
}

// OutputSink receives observable output from a hosted evaluation. The current
// JS/WASM bridge maps it to the matching able_host methods. It is deliberately
// separate from source loading and other host services.
type OutputSink interface {
	WriteStdout(message string) error
	WriteStderr(message string) error
}

// EvaluateRequestJSON executes an EvaluateRequest and returns a JSON response.
func EvaluateRequestJSON(payload []byte) []byte {
	return evaluateRequestJSON(payload, nil)
}

// EvaluateRequestJSONWithOutput executes a request and forwards host output.
func EvaluateRequestJSONWithOutput(payload []byte, output OutputSink) []byte {
	return evaluateRequestJSON(payload, output)
}

func evaluateRequestJSON(payload []byte, output OutputSink) []byte {
	var req EvaluateRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		resp := EvaluateResponse{
			OK:    false,
			Error: fmt.Sprintf("decode request json: %v", err),
		}
		reportFailure(output, resp)
		return marshalResponse(resp)
	}
	return marshalResponse(evaluate(req, output))
}

// Evaluate executes a request and returns a structured response.
func Evaluate(req EvaluateRequest) EvaluateResponse {
	return evaluate(req, nil)
}

// EvaluateWithOutput executes a request and forwards host output.
func EvaluateWithOutput(req EvaluateRequest, output OutputSink) EvaluateResponse {
	return evaluate(req, output)
}

func evaluate(req EvaluateRequest, output OutputSink) EvaluateResponse {
	interp, err := newInterpreter(req.ExecMode)
	if err != nil {
		resp := EvaluateResponse{OK: false, Error: err.Error()}
		reportFailure(output, resp)
		return resp
	}
	installHostOutput(interp, output)

	setupIndex := 0
	for _, raw := range req.Setup {
		if resp, failed := evaluateSetupModule(interp, raw, "", setupIndex, output); failed {
			return resp
		}
		setupIndex++
	}
	for _, setup := range req.SetupModules {
		if resp, failed := evaluateSetupModule(interp, setup.Module, setup.Origin, setupIndex, output); failed {
			return resp
		}
		setupIndex++
	}

	if len(req.Module) == 0 {
		resp := EvaluateResponse{OK: false, Error: "missing module payload"}
		reportFailure(output, resp)
		return resp
	}

	entry, err := interpreter.DecodeModule(req.Module)
	if err != nil {
		resp := EvaluateResponse{OK: false, Error: fmt.Sprintf("decode %s json: %v", entryModuleLabel(req.EntryOrigin), err)}
		reportFailure(output, resp)
		return resp
	}
	value, env, err := interp.EvaluateModule(entry)
	if err != nil {
		resp := EvaluateResponse{OK: false, Error: fmt.Sprintf("evaluate %s: %v", entryModuleLabel(req.EntryOrigin), err)}
		reportFailure(output, resp)
		return resp
	}
	rendered, err := interp.Stringify(value, env)
	if err != nil {
		resp := EvaluateResponse{OK: false, Error: fmt.Sprintf("stringify result: %v", err)}
		reportFailure(output, resp)
		return resp
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

func evaluateSetupModule(interp *interpreter.Interpreter, raw json.RawMessage, origin string, index int, output OutputSink) (EvaluateResponse, bool) {
	label := setupModuleLabel(index, origin)
	mod, err := interpreter.DecodeModule(raw)
	if err != nil {
		resp := EvaluateResponse{OK: false, Error: fmt.Sprintf("decode %s: %v", label, err)}
		reportFailure(output, resp)
		return resp, true
	}
	if _, _, err := interp.EvaluateModule(mod); err != nil {
		resp := EvaluateResponse{OK: false, Error: fmt.Sprintf("evaluate %s: %v", label, err)}
		reportFailure(output, resp)
		return resp, true
	}
	return EvaluateResponse{}, false
}

func setupModuleLabel(index int, origin string) string {
	if origin == "" {
		return fmt.Sprintf("setup module %d", index)
	}
	return fmt.Sprintf("setup module %d (%s)", index, origin)
}

func entryModuleLabel(origin string) string {
	if origin == "" {
		return "module"
	}
	return fmt.Sprintf("module (%s)", origin)
}

func installHostOutput(interp *interpreter.Interpreter, output OutputSink) {
	if interp == nil || output == nil {
		return
	}
	interp.GlobalEnvironment().Define("print", runtime.NativeFunctionValue{
		Name:  "print",
		Arity: 1,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			parts := make([]string, 0, len(args))
			for _, arg := range args {
				rendered, err := interp.Stringify(arg, ctx.Env)
				if err != nil {
					return nil, fmt.Errorf("wasm host print: %w", err)
				}
				parts = append(parts, rendered)
			}
			if err := output.WriteStdout(strings.Join(parts, " ") + "\n"); err != nil {
				return nil, fmt.Errorf("wasm host stdout: %w", err)
			}
			return runtime.NilValue{}, nil
		},
	})
}

func reportFailure(output OutputSink, resp EvaluateResponse) {
	if output == nil || resp.OK || resp.Error == "" {
		return
	}
	// The structured response remains the source of truth if a host reporter
	// itself fails; a missing/broken stderr callback must not hide evaluation
	// diagnostics from the JavaScript caller.
	_ = output.WriteStderr(resp.Error + "\n")
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
