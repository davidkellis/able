package interpreter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

type runtimeCallFrame struct {
	node *ast.FunctionCall
}

type runtimeDiagnosticContext struct {
	node      ast.Node
	callStack []runtimeCallFrame
}

type runtimeDiagnosticError struct {
	err     error
	context *runtimeDiagnosticContext
}

func (e runtimeDiagnosticError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e runtimeDiagnosticError) Unwrap() error {
	return e.err
}

type RuntimeDiagnosticNote struct {
	Message  string
	Location driver.DiagnosticLocation
}

type RuntimeDiagnostic struct {
	Severity driver.DiagnosticSeverity
	Message  string
	Location driver.DiagnosticLocation
	Notes    []RuntimeDiagnosticNote
}

func (i *Interpreter) BuildRuntimeDiagnostic(err error) RuntimeDiagnostic {
	message := runtimeMessageFromError(err)
	ctx := runtimeContextFromError(err)

	location := runtimeLocationFromNode(i, nil)
	if ctx != nil && ctx.node != nil {
		location = runtimeLocationFromNode(i, ctx.node)
	}
	if location == (driver.DiagnosticLocation{}) && ctx != nil {
		for idx := len(ctx.callStack) - 1; idx >= 0; idx-- {
			if ctx.callStack[idx].node == nil {
				continue
			}
			location = runtimeLocationFromNode(i, ctx.callStack[idx].node)
			if location != (driver.DiagnosticLocation{}) {
				break
			}
		}
	}

	var notes []RuntimeDiagnosticNote
	if ctx != nil && len(ctx.callStack) > 0 {
		for idx := len(ctx.callStack) - 1; idx >= 0 && len(notes) < 8; idx-- {
			node := ctx.callStack[idx].node
			if node == nil {
				continue
			}
			noteLocation := runtimeLocationFromNode(i, node)
			if noteLocation == (driver.DiagnosticLocation{}) || runtimeLocationsEqual(noteLocation, location) {
				continue
			}
			notes = append(notes, RuntimeDiagnosticNote{
				Message:  "called from here",
				Location: noteLocation,
			})
		}
	}

	return RuntimeDiagnostic{
		Severity: driver.SeverityError,
		Message:  message,
		Location: location,
		Notes:    notes,
	}
}

func DescribeRuntimeDiagnostic(diag RuntimeDiagnostic) string {
	message := strings.TrimSpace(diag.Message)
	if strings.HasPrefix(message, "runtime:") {
		message = strings.TrimSpace(strings.TrimPrefix(message, "runtime:"))
	}
	location := formatRuntimeLocation(diag.Location)
	prefix := "runtime: "
	if diag.Severity == driver.SeverityWarning {
		prefix = "warning: runtime: "
	}
	var b strings.Builder
	if location != "" {
		fmt.Fprintf(&b, "%s%s %s", prefix, location, message)
	} else {
		fmt.Fprintf(&b, "%s%s", prefix, message)
	}
	for _, note := range diag.Notes {
		noteLoc := formatRuntimeLocation(note.Location)
		if noteLoc != "" {
			fmt.Fprintf(&b, "\nnote: %s %s", noteLoc, note.Message)
		} else {
			fmt.Fprintf(&b, "\nnote: %s", note.Message)
		}
	}
	return b.String()
}

func (i *Interpreter) attachRuntimeContext(err error, node ast.Node, state *evalState) error {
	if err == nil || node == nil {
		return err
	}
	switch err.(type) {
	case returnSignal, breakSignal, continueSignal, generatorStopSignal:
		return err
	}
	if runtimeContextFromError(err) != nil {
		return err
	}
	context := &runtimeDiagnosticContext{
		node:      node,
		callStack: state.snapshotCallStack(),
	}
	if sig, ok := err.(raiseSignal); ok {
		sig.context = context
		return sig
	}
	if diagErr, ok := err.(runtimeDiagnosticError); ok {
		diagErr.context = context
		return diagErr
	}
	return runtimeDiagnosticError{err: err, context: context}
}

func runtimeContextFromError(err error) *runtimeDiagnosticContext {
	if err == nil {
		return nil
	}
	var sig raiseSignal
	if errors.As(err, &sig) {
		if sig.context != nil {
			return sig.context
		}
	}
	var diagErr runtimeDiagnosticError
	if errors.As(err, &diagErr) {
		if diagErr.context != nil {
			return diagErr.context
		}
	}
	return nil
}

func runtimeMessageFromError(err error) string {
	if err == nil {
		return ""
	}
	var sig raiseSignal
	if errors.As(err, &sig) {
		if msg := runtimeErrorValueMessage(sig.value); msg != "" {
			return msg
		}
	}
	var diagErr runtimeDiagnosticError
	if errors.As(err, &diagErr) {
		if diagErr.err != nil && diagErr.err != err {
			return runtimeMessageFromError(diagErr.err)
		}
	}
	return err.Error()
}

func runtimeErrorValueMessage(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.ErrorValue:
		return v.Message
	case *runtime.ErrorValue:
		if v != nil {
			return v.Message
		}
	}
	return ""
}

func formatRuntimeLocation(loc driver.DiagnosticLocation) string {
	path := strings.TrimSpace(loc.Path)
	line := loc.Line
	column := loc.Column
	if path != "" {
		path = normalizeRuntimePath(path)
	}
	switch {
	case path != "" && line > 0 && column > 0:
		return fmt.Sprintf("%s:%d:%d", path, line, column)
	case path != "" && line > 0:
		return fmt.Sprintf("%s:%d", path, line)
	case path != "":
		return path
	case line > 0 && column > 0:
		return fmt.Sprintf("line %d, column %d", line, column)
	case line > 0:
		return fmt.Sprintf("line %d", line)
	default:
		return ""
	}
}

func runtimeLocationFromNode(i *Interpreter, node ast.Node) driver.DiagnosticLocation {
	if node == nil {
		return driver.DiagnosticLocation{}
	}
	span := node.Span()
	path := ""
	if i != nil && i.nodeOrigins != nil {
		if origin, ok := i.nodeOrigins[node]; ok {
			path = origin
		}
	}
	return driver.DiagnosticLocation{
		Path:      path,
		Line:      span.Start.Line,
		Column:    span.Start.Column,
		EndLine:   span.End.Line,
		EndColumn: span.End.Column,
	}
}

func runtimeLocationsEqual(left, right driver.DiagnosticLocation) bool {
	if left == (driver.DiagnosticLocation{}) || right == (driver.DiagnosticLocation{}) {
		return false
	}
	return left.Path == right.Path && left.Line == right.Line && left.Column == right.Column
}

var (
	runtimeDiagRootOnce sync.Once
	runtimeDiagRootPath string
)

func normalizeRuntimePath(raw string) string {
	if raw == "" {
		return ""
	}
	path := raw
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}
	root := runtimeDiagnosticRoot()
	if root != "" {
		if rel, err := filepath.Rel(root, path); err == nil {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(path)
}

func runtimeDiagnosticRoot() string {
	runtimeDiagRootOnce.Do(func() {
		start := ""
		if _, file, _, ok := goruntime.Caller(0); ok {
			start = filepath.Dir(file)
		} else if wd, err := os.Getwd(); err == nil {
			start = wd
		}
		dir := start
		for i := 0; i < 12 && dir != "" && dir != string(filepath.Separator); i++ {
			if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
				runtimeDiagRootPath = dir
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	})
	return runtimeDiagRootPath
}
