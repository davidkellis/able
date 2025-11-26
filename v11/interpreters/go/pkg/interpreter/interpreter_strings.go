package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) stringMember(str runtime.StringValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("String member access expects identifier")
	}
	return nil, fmt.Errorf("String has no member '%s' (import able.text.string for stdlib helpers)", ident.Name)
}
