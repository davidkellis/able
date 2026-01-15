package driver

import (
	"fmt"
	"strings"
)

// DiagnosticSeverity captures parser diagnostic levels.
type DiagnosticSeverity string

const (
	SeverityError   DiagnosticSeverity = "error"
	SeverityWarning DiagnosticSeverity = "warning"
)

// DiagnosticLocation references a source span for diagnostics.
type DiagnosticLocation struct {
	Path      string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}

// ParserDiagnostic represents a structured parser diagnostic.
type ParserDiagnostic struct {
	Severity DiagnosticSeverity
	Message  string
	Location DiagnosticLocation
}

// ParserDiagnosticError wraps a diagnostic for error handling.
type ParserDiagnosticError struct {
	Diagnostic ParserDiagnostic
}

func (e *ParserDiagnosticError) Error() string {
	return e.Diagnostic.Message
}

// DescribeParserDiagnostic formats a parser diagnostic for CLI output.
func DescribeParserDiagnostic(diag ParserDiagnostic) string {
	message := strings.TrimSpace(diag.Message)
	if strings.HasPrefix(message, "parser:") {
		message = strings.TrimSpace(strings.TrimPrefix(message, "parser:"))
	}
	location := formatDiagnosticLocation(diag.Location)
	prefix := "parser: "
	if diag.Severity == SeverityWarning {
		prefix = "warning: parser: "
	}
	if location != "" {
		return fmt.Sprintf("%s%s %s", prefix, location, message)
	}
	return fmt.Sprintf("%s%s", prefix, message)
}

func formatDiagnosticLocation(loc DiagnosticLocation) string {
	path := strings.TrimSpace(loc.Path)
	line := loc.Line
	column := loc.Column
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
