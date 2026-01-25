package ast

import "encoding/json"

// MarshalJSON ensures integer literals serialize with numeric values in fixtures.
func (lit *IntegerLiteral) MarshalJSON() ([]byte, error) {
	if lit == nil {
		return []byte("null"), nil
	}
	value := "0"
	if lit.Value != nil {
		value = lit.Value.String()
	}
	payload := struct {
		Type        NodeType        `json:"type"`
		Value       json.RawMessage `json:"value"`
		IntegerType *IntegerType    `json:"integerType,omitempty"`
	}{
		Type:        lit.Type,
		Value:       json.RawMessage(value),
		IntegerType: lit.IntegerType,
	}
	return json.Marshal(payload)
}
