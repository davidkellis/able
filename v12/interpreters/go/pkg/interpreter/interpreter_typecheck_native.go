//go:build !(js && wasm)

package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

// prepareModuleTypechecking performs the optional standalone-module check.
// Program evaluation installs its own checked facts before calling the shared
// AST evaluator, so this remains only the standalone compatibility path.
func (i *Interpreter) prepareModuleTypechecking(module *ast.Module) (func(), error) {
	i.typecheckDiagnostics = nil
	if !i.typecheckerEnabled {
		return nil, nil
	}
	if module == nil {
		return nil, fmt.Errorf("typechecker: module is nil")
	}
	if i.typechecker == nil {
		i.typechecker = typechecker.New()
	}
	diags, err := i.typechecker.CheckModule(module)
	if err != nil {
		return nil, err
	}
	i.typecheckDiagnostics = append(i.typecheckDiagnostics[:0], diags...)
	if i.typecheckerStrict && len(diags) > 0 {
		msg := diags[0].Message
		if !strings.HasPrefix(msg, "typechecker:") {
			msg = "typechecker: " + msg
		}
		return nil, fmt.Errorf("%s", msg)
	}
	if len(diags) == 0 {
		return i.pushBytecodeInferenceFacts(i.typechecker.Inference()), nil
	}
	return i.pushBytecodeInferenceFacts(nil), nil
}
