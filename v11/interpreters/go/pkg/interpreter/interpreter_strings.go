package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) stringMember(str runtime.StringValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("String member access expects identifier")
	}
	// When stdlib string methods are not imported, surface a clear diagnostic.
	return nil, fmt.Errorf("string has no member '%s' (import able.text.string for stdlib helpers)", ident.Name)
}
