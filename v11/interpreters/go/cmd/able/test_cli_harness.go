package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

func buildDiscoveryRequest(interp *interpreter.Interpreter, cli *testCliModule, config TestCliConfig) (runtime.Value, error) {
	if cli == nil || cli.discoveryDef == nil {
		return nil, fmt.Errorf("missing DiscoveryRequest definition")
	}
	includePaths, _ := makeStringArray(interp, config.Filters.IncludePaths)
	excludePaths, _ := makeStringArray(interp, config.Filters.ExcludePaths)
	includeNames, _ := makeStringArray(interp, config.Filters.IncludeNames)
	excludeNames, _ := makeStringArray(interp, config.Filters.ExcludeNames)
	includeTags, _ := makeStringArray(interp, config.Filters.IncludeTags)
	excludeTags, _ := makeStringArray(interp, config.Filters.ExcludeTags)

	fields := map[string]runtime.Value{
		"include_paths": includePaths,
		"exclude_paths": excludePaths,
		"include_names": includeNames,
		"exclude_names": excludeNames,
		"include_tags":  includeTags,
		"exclude_tags":  excludeTags,
		"list_only":     runtime.BoolValue{Val: config.ListOnly},
	}
	return &runtime.StructInstanceValue{Definition: cli.discoveryDef, Fields: fields}, nil
}

func buildRunOptions(interp *interpreter.Interpreter, cli *testCliModule, run TestRunOptions) (runtime.Value, error) {
	if cli == nil || cli.runOptionsDef == nil {
		return nil, fmt.Errorf("missing RunOptions definition")
	}
	var shuffle runtime.Value = runtime.NilValue{}
	if run.ShuffleSeed != nil {
		shuffle = makeIntegerValue(runtime.IntegerI64, *run.ShuffleSeed)
	}
	fields := map[string]runtime.Value{
		"shuffle_seed": shuffle,
		"fail_fast":    runtime.BoolValue{Val: run.FailFast},
		"parallelism":  makeIntegerValue(runtime.IntegerI32, int64(run.Parallelism)),
		"repeat":       makeIntegerValue(runtime.IntegerI32, int64(run.Repeat)),
	}
	return &runtime.StructInstanceValue{Definition: cli.runOptionsDef, Fields: fields}, nil
}

func buildTestPlan(cli *testCliModule, descriptors runtime.Value) (runtime.Value, error) {
	if cli == nil || cli.testPlanDef == nil {
		return nil, fmt.Errorf("missing TestPlan definition")
	}
	fields := map[string]runtime.Value{
		"descriptors": descriptors,
	}
	return &runtime.StructInstanceValue{Definition: cli.testPlanDef, Fields: fields}, nil
}

func callHarnessDiscover(interp *interpreter.Interpreter, cli *testCliModule, request runtime.Value) (runtime.Value, bool) {
	if cli == nil {
		fmt.Fprintln(os.Stderr, "able test: missing CLI module")
		return nil, false
	}
	result, err := interp.CallFunction(cli.discoverAll, []runtime.Value{request})
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return nil, false
	}
	if failure := extractFailure(interp, result); failure != nil {
		fmt.Fprintf(os.Stderr, "able test: %s\n", formatFailure(failure))
		return nil, false
	}
	if _, err := coerceArrayValue(interp, result, "discovery result"); err != nil {
		fmt.Fprintln(os.Stderr, "able test: discovery returned unexpected result")
		return nil, false
	}
	return result, true
}

func callHarnessRun(
	interp *interpreter.Interpreter,
	cli *testCliModule,
	plan runtime.Value,
	options runtime.Value,
	reporter runtime.Value,
) *harnessFailure {
	if cli == nil {
		fmt.Fprintln(os.Stderr, "able test: missing CLI module")
		return &harnessFailure{message: "missing CLI module"}
	}
	result, err := interp.CallFunction(cli.runPlan, []runtime.Value{plan, options, reporter})
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return &harnessFailure{message: err.Error()}
	}
	if isNilValue(result) {
		return nil
	}
	if failure := extractFailure(interp, result); failure != nil {
		fmt.Fprintf(os.Stderr, "able test: %s\n", formatFailure(failure))
		return failure
	}
	fmt.Fprintln(os.Stderr, "able test: run_plan returned unexpected result")
	return &harnessFailure{message: "run_plan returned unexpected result"}
}

func createTestReporter(
	interp *interpreter.Interpreter,
	cli *testCliModule,
	format TestReporterFormat,
	state *TestEventState,
) (*reporterBundle, error) {
	if cli == nil {
		return nil, fmt.Errorf("missing CLI module")
	}
	emit := createEventHandler(interp, format, state)
	emitFn := runtime.NativeFunctionValue{
		Name:  "__able_test_cli_emit",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) > 0 && args[0] != nil {
				emit(args[0])
			}
			return runtime.NilValue{}, nil
		},
	}

	if format == reporterJSON || format == reporterTap {
		if format == reporterTap {
			fmt.Fprintln(os.Stdout, "TAP version 13")
		}
		reporter, err := interp.CallFunction(cli.cliReporter, []runtime.Value{emitFn})
		if err != nil {
			return nil, err
		}
		return &reporterBundle{reporter: reporter}, nil
	}

	inner, err := createStdlibReporter(interp, cli, format)
	if err != nil {
		return nil, err
	}
	reporter, err := interp.CallFunction(cli.cliComposite, []runtime.Value{inner, emitFn})
	if err != nil {
		return nil, err
	}
	var finish func()
	if format == reporterProgress {
		finish = func() { finishProgressReporter(interp, inner) }
	}
	return &reporterBundle{reporter: reporter, finish: finish}, nil
}

func createStdlibReporter(
	interp *interpreter.Interpreter,
	cli *testCliModule,
	format TestReporterFormat,
) (runtime.Value, error) {
	writeLine := runtime.NativeFunctionValue{
		Name:  "__able_test_cli_line",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) == 0 || args[0] == nil {
				return runtime.NilValue{}, nil
			}
			fmt.Fprintln(os.Stdout, runtimeValueToString(args[0]))
			return runtime.NilValue{}, nil
		},
	}
	var reporterFn runtime.Value
	switch format {
	case reporterDoc:
		reporterFn = cli.docReporter
	case reporterProgress:
		reporterFn = cli.progressReporter
	default:
		return nil, fmt.Errorf("unsupported reporter format %q", format)
	}
	return interp.CallFunction(reporterFn, []runtime.Value{writeLine})
}

func finishProgressReporter(interp *interpreter.Interpreter, reporter runtime.Value) {
	method, err := interp.LookupStructMethod(reporter, "finish")
	if err != nil || method == nil {
		return
	}
	if _, err := interp.CallFunction(method, []runtime.Value{reporter}); err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
	}
}

func emitTestPlanList(interp *interpreter.Interpreter, descriptors runtime.Value, config TestCliConfig) {
	items, err := decodeDescriptorArray(interp, descriptors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return
	}
	if config.ReporterFormat == reporterJSON {
		payload, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "able test: %v\n", err)
			return
		}
		fmt.Fprintln(os.Stdout, string(payload))
		return
	}
	if len(items) == 0 {
		fmt.Fprintln(os.Stdout, "able test: no tests found")
		return
	}
	for _, item := range items {
		tags := "-"
		if len(item.Tags) > 0 {
			tags = strings.Join(item.Tags, ",")
		}
		modulePath := item.ModulePath
		if modulePath == "" {
			modulePath = "-"
		}
		parts := []string{
			item.FrameworkID,
			modulePath,
			item.TestID,
			item.DisplayName,
			fmt.Sprintf("tags=%s", tags),
		}
		if config.DryRun {
			parts = append(parts, fmt.Sprintf("metadata=%s", formatMetadata(item.Metadata)))
		}
		fmt.Fprintln(os.Stdout, strings.Join(parts, " | "))
	}
}

func createEventHandler(
	interp *interpreter.Interpreter,
	format TestReporterFormat,
	state *TestEventState,
) func(runtime.Value) {
	tapIndex := 0
	return func(event runtime.Value) {
		decoded, err := decodeTestEvent(interp, event)
		if err != nil || decoded == nil {
			if err != nil {
				fmt.Fprintf(os.Stderr, "able test: %v\n", err)
			}
			return
		}
		recordTestEvent(state, decoded)

		switch format {
		case reporterJSON:
			payload, err := json.Marshal(decoded)
			if err != nil {
				fmt.Fprintf(os.Stderr, "able test: %v\n", err)
				return
			}
			fmt.Fprintln(os.Stdout, string(payload))
		case reporterTap:
			switch decoded.Kind {
			case "case_passed":
				tapIndex++
				fmt.Fprintf(os.Stdout, "ok %d - %s\n", tapIndex, decoded.Descriptor.DisplayName)
			case "case_failed":
				tapIndex++
				fmt.Fprintf(os.Stdout, "not ok %d - %s\n", tapIndex, decoded.Descriptor.DisplayName)
				emitTapFailure(decoded.Failure)
			case "case_skipped":
				tapIndex++
				reason := "skipped"
				if decoded.Reason != nil {
					reason = *decoded.Reason
				}
				fmt.Fprintf(os.Stdout, "ok %d - %s # SKIP %s\n", tapIndex, decoded.Descriptor.DisplayName, reason)
			case "framework_error":
				fmt.Fprintf(os.Stdout, "Bail out! %s\n", decoded.Message)
			}
		default:
		}
	}
}

func emitTapFailure(failure *failureData) {
	if failure == nil {
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
		fmt.Fprintln(os.Stdout, line)
	}
}

func sanitizeTapValue(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\\n"), "\n", "\\n")
}

func recordTestEvent(state *TestEventState, event *testEvent) {
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
