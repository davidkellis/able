package wasmhost

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestEvaluateRequestJSONTreewalker(t *testing.T) {
	resp := decodeResponse(t, EvaluateRequestJSON(mustJSON(t, EvaluateRequest{
		Module: []byte(simpleAdditionModuleJSON),
	})))
	if !resp.OK {
		t.Fatalf("expected success, got error: %s", resp.Error)
	}
	if resp.Result != "3" {
		t.Fatalf("expected result 3, got %q", resp.Result)
	}
}

func TestEvaluateRequestJSONBytecode(t *testing.T) {
	resp := decodeResponse(t, EvaluateRequestJSON(mustJSON(t, EvaluateRequest{
		ExecMode: "bytecode",
		Module:   []byte(simpleAdditionModuleJSON),
	})))
	if !resp.OK {
		t.Fatalf("expected success, got error: %s", resp.Error)
	}
	if resp.Result != "3" {
		t.Fatalf("expected result 3, got %q", resp.Result)
	}
}

func TestEvaluateRequestJSONRejectsUnsupportedExecMode(t *testing.T) {
	resp := decodeResponse(t, EvaluateRequestJSON(mustJSON(t, EvaluateRequest{
		ExecMode: "jit",
		Module:   []byte(simpleAdditionModuleJSON),
	})))
	if resp.OK {
		t.Fatalf("expected failure for invalid exec mode")
	}
	if !strings.Contains(resp.Error, "unsupported execMode") {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
}

func TestEvaluateRequestJSONDecodeError(t *testing.T) {
	resp := decodeResponse(t, EvaluateRequestJSON([]byte(`{"execMode":"treewalker","module":{"type":"Module","imports":[],"body":[{"type":"UnknownNode"}]}}`)))
	if resp.OK {
		t.Fatalf("expected failure for invalid module payload")
	}
	if !strings.Contains(resp.Error, "decode module json") {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
}

func TestEvaluateWithOutputForwardsPrintInBothModes(t *testing.T) {
	for _, execMode := range []string{"treewalker", "bytecode"} {
		t.Run(execMode, func(t *testing.T) {
			output := &recordingOutput{}
			resp := EvaluateWithOutput(EvaluateRequest{
				ExecMode: execMode,
				Module:   []byte(hostOutputModuleJSON),
			}, output)
			if !resp.OK {
				t.Fatalf("expected success, got error: %s", resp.Error)
			}
			if resp.Result != "3" {
				t.Fatalf("result = %q, want 3", resp.Result)
			}
			if got, want := output.stdout, []string{"wasm host output\n"}; !sameStrings(got, want) {
				t.Fatalf("stdout = %#v, want %#v", got, want)
			}
			if len(output.stderr) != 0 {
				t.Fatalf("unexpected stderr: %#v", output.stderr)
			}
		})
	}
}

func TestEvaluateWithoutOutputDoesNotInstallHostPrint(t *testing.T) {
	resp := Evaluate(EvaluateRequest{Module: []byte(hostOutputModuleJSON)})
	if resp.OK {
		t.Fatal("host print should only exist when an output sink is supplied")
	}
}

func TestEvaluateRequestJSONWithOutputForwardsFailure(t *testing.T) {
	output := &recordingOutput{}
	resp := decodeResponse(t, EvaluateRequestJSONWithOutput([]byte(`{"module":{"type":"Module","imports":[],"body":[{"type":"UnknownNode"}]}}`), output))
	if resp.OK {
		t.Fatal("expected failure")
	}
	if len(output.stderr) != 1 || !strings.Contains(output.stderr[0], "decode module json") {
		t.Fatalf("stderr = %#v, want decode error", output.stderr)
	}
}

func TestEvaluateWithSetupModulesUsesDependencyOrder(t *testing.T) {
	for _, execMode := range []string{"treewalker", "bytecode"} {
		t.Run(execMode, func(t *testing.T) {
			resp := Evaluate(EvaluateRequest{
				ExecMode: execMode,
				SetupModules: []SetupModule{
					{Origin: "modules/dep.able", Module: []byte(moduleJSON("dep", nil, `{"type":"AssignmentExpression","operator":":=","left":{"type":"Identifier","name":"base"},"right":{"type":"IntegerLiteral","value":40}}`))},
					{Origin: "modules/math.able", Module: []byte(moduleJSON("math", []string{"dep", "base"}, `{"type":"AssignmentExpression","operator":":=","left":{"type":"Identifier","name":"answer"},"right":{"type":"BinaryExpression","operator":"+","left":{"type":"Identifier","name":"base"},"right":{"type":"IntegerLiteral","value":2}}}`))},
				},
				EntryOrigin: "modules/main.able",
				Module:      []byte(moduleJSON("app", []string{"math", "answer"}, `{"type":"Identifier","name":"answer"}`)),
			})
			if !resp.OK {
				t.Fatalf("expected success, got error: %s", resp.Error)
			}
			if resp.Result != "42" {
				t.Fatalf("result = %q, want 42", resp.Result)
			}
		})
	}
}

func TestEvaluateWithSetupModulesIncludesOriginInFailure(t *testing.T) {
	resp := Evaluate(EvaluateRequest{
		SetupModules: []SetupModule{{
			Origin: "modules/bad.able",
			Module: []byte(`{"type":"Module","imports":[],"body":[{"type":"UnknownNode"}]}`),
		}},
		Module: []byte(simpleAdditionModuleJSON),
	})
	if resp.OK {
		t.Fatal("expected setup failure")
	}
	if !strings.Contains(resp.Error, "setup module 0 (modules/bad.able)") {
		t.Fatalf("error = %q, want setup origin", resp.Error)
	}
}

type recordingOutput struct {
	stdout []string
	stderr []string
}

func (o *recordingOutput) WriteStdout(message string) error {
	o.stdout = append(o.stdout, message)
	return nil
}

func (o *recordingOutput) WriteStderr(message string) error {
	o.stderr = append(o.stderr, message)
	return nil
}

func sameStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for idx := range left {
		if left[idx] != right[idx] {
			return false
		}
	}
	return true
}

func moduleJSON(packageName string, selector []string, body string) string {
	imports := "[]"
	if len(selector) == 2 {
		imports = fmt.Sprintf(`[{"type":"ImportStatement","packagePath":[{"type":"Identifier","name":%q}],"isWildcard":false,"selectors":[{"type":"ImportSelector","name":{"type":"Identifier","name":%q}}]}]`, selector[0], selector[1])
	}
	return fmt.Sprintf(`{"type":"Module","package":{"type":"PackageStatement","namePath":[{"type":"Identifier","name":%q}]},"imports":%s,"body":[%s]}`, packageName, imports, body)
}

func decodeResponse(t *testing.T, payload []byte) EvaluateResponse {
	t.Helper()
	var resp EvaluateResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("decode response: %v\npayload=%s", err, string(payload))
	}
	return resp
}

func mustJSON(t *testing.T, req EvaluateRequest) []byte {
	t.Helper()
	encoded, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	return encoded
}

const simpleAdditionModuleJSON = `{
  "type": "Module",
  "imports": [],
  "body": [
    {
      "type": "AssignmentExpression",
      "operator": ":=",
      "left": { "type": "Identifier", "name": "a" },
      "right": { "type": "IntegerLiteral", "value": 1 }
    },
    {
      "type": "AssignmentExpression",
      "operator": ":=",
      "left": { "type": "Identifier", "name": "b" },
      "right": { "type": "IntegerLiteral", "value": 2 }
    },
    {
      "type": "BinaryExpression",
      "operator": "+",
      "left": { "type": "Identifier", "name": "a" },
      "right": { "type": "Identifier", "name": "b" }
    }
  ]
}`

const hostOutputModuleJSON = `{
  "type": "Module",
  "imports": [],
  "body": [
    {
      "type": "FunctionCall",
      "callee": { "type": "Identifier", "name": "print" },
      "arguments": [{ "type": "StringLiteral", "value": "wasm host output" }],
      "isTrailingLambda": false
    },
    { "type": "IntegerLiteral", "value": 3 }
  ]
}`
