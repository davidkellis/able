package testcli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

type EventEmitter struct {
	format   ReporterFormat
	stdout   io.Writer
	state    *EventState
	tapIndex int
}

func NewEventEmitter(format ReporterFormat, stdout io.Writer, state *EventState) *EventEmitter {
	return &EventEmitter{
		format: format,
		stdout: stdout,
		state:  state,
	}
}

func (e *EventEmitter) EmitHeader() {
	if e == nil || e.format != ReporterTap || e.stdout == nil {
		return
	}
	fmt.Fprintln(e.stdout, "TAP version 13")
}

func (e *EventEmitter) EmitValue(interp *interpreter.Interpreter, value runtime.Value) error {
	event, err := DecodeTestEvent(interp, value)
	if err != nil || event == nil {
		return err
	}
	return e.Emit(event)
}

func (e *EventEmitter) Emit(event *TestEvent) error {
	if e == nil || event == nil {
		return nil
	}
	recordEvent(e.state, event)
	switch e.format {
	case ReporterJSON:
		payload, err := json.Marshal(event)
		if err != nil {
			return err
		}
		fmt.Fprintln(e.stdout, string(payload))
	case ReporterTap:
		e.emitTap(event)
	}
	return nil
}

func (e *EventEmitter) emitTap(event *TestEvent) {
	if e.stdout == nil || event == nil {
		return
	}
	switch event.Kind {
	case "case_passed":
		e.tapIndex++
		fmt.Fprintf(e.stdout, "ok %d - %s\n", e.tapIndex, event.Descriptor.DisplayName)
	case "case_failed":
		e.tapIndex++
		fmt.Fprintf(e.stdout, "not ok %d - %s\n", e.tapIndex, event.Descriptor.DisplayName)
		emitTapFailure(e.stdout, event.Failure)
	case "case_skipped":
		e.tapIndex++
		reason := "skipped"
		if event.Reason != nil {
			reason = *event.Reason
		}
		fmt.Fprintf(e.stdout, "ok %d - %s # SKIP %s\n", e.tapIndex, event.Descriptor.DisplayName, reason)
	case "framework_error":
		fmt.Fprintf(e.stdout, "Bail out! %s\n", event.Message)
	}
}

func recordEvent(state *EventState, event *TestEvent) {
	if state == nil || event == nil {
		return
	}
	switch event.Kind {
	case "case_passed":
		state.Total++
	case "case_failed":
		state.Total++
		state.Failed++
	case "case_skipped":
		state.Total++
		state.Skipped++
	case "framework_error":
		state.FrameworkErrors++
	}
}

func emitTapFailure(stdout io.Writer, failure *FailureData) {
	if stdout == nil || failure == nil {
		return
	}
	lines := []string{
		"  ---",
		fmt.Sprintf("  message: %s", sanitizeTapValue(failure.Message)),
	}
	if failure.Details != nil {
		lines = append(lines, fmt.Sprintf("  details: %s", sanitizeTapValue(*failure.Details)))
	}
	if failure.Location != nil {
		lines = append(lines, fmt.Sprintf(
			"  location: %s",
			sanitizeTapValue(fmt.Sprintf("%s:%d:%d", failure.Location.ModulePath, failure.Location.Line, failure.Location.Column)),
		))
	}
	lines = append(lines, "  ...")
	for _, line := range lines {
		fmt.Fprintln(stdout, line)
	}
}

func sanitizeTapValue(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\\n"), "\n", "\\n")
}
