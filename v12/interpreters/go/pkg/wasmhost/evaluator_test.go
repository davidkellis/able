package wasmhost

import (
	"encoding/json"
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
