package testcli

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

func DecodeTestEvent(interp *interpreter.Interpreter, value runtime.Value) (*TestEvent, error) {
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil, nil
	}
	switch structTag(inst) {
	case "case_started":
		descriptor, err := decodeDescriptor(interp, structField(inst, "descriptor"))
		if err != nil {
			return nil, err
		}
		return &TestEvent{Kind: "case_started", Descriptor: descriptor}, nil
	case "case_passed":
		descriptor, err := decodeDescriptor(interp, structField(inst, "descriptor"))
		if err != nil {
			return nil, err
		}
		duration := decodeNumber(structField(inst, "duration_ms"))
		return &TestEvent{Kind: "case_passed", Descriptor: descriptor, DurationMs: duration}, nil
	case "case_failed":
		descriptor, err := decodeDescriptor(interp, structField(inst, "descriptor"))
		if err != nil {
			return nil, err
		}
		duration := decodeNumber(structField(inst, "duration_ms"))
		failure, err := decodeFailure(interp, structField(inst, "failure"))
		if err != nil {
			return nil, err
		}
		return &TestEvent{Kind: "case_failed", Descriptor: descriptor, DurationMs: duration, Failure: failure}, nil
	case "case_skipped":
		descriptor, err := decodeDescriptor(interp, structField(inst, "descriptor"))
		if err != nil {
			return nil, err
		}
		reason := decodeOptionalString(interp, structField(inst, "reason"))
		return &TestEvent{Kind: "case_skipped", Descriptor: descriptor, Reason: reason}, nil
	case "framework_error":
		message := decodeString(interp, structField(inst, "message"))
		return &TestEvent{Kind: "framework_error", Message: message}, nil
	default:
		return nil, nil
	}
}

func DecodeDescriptorArray(interp *interpreter.Interpreter, value runtime.Value) ([]TestDescriptor, error) {
	arrayVal, err := coerceArrayValue(interp, value, "descriptor array")
	if err != nil {
		return nil, err
	}
	out := make([]TestDescriptor, 0, len(arrayVal.Elements))
	for _, entry := range arrayVal.Elements {
		desc, err := decodeDescriptor(interp, entry)
		if err != nil {
			return nil, err
		}
		out = append(out, *desc)
	}
	return out, nil
}

func FormatMetadata(entries []MetadataEntry) string {
	if len(entries) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		parts = append(parts, fmt.Sprintf("%s=%s", entry.Key, entry.Value))
	}
	return strings.Join(parts, ",")
}

func decodeDescriptor(interp *interpreter.Interpreter, value runtime.Value) (*TestDescriptor, error) {
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil, fmt.Errorf("expected TestDescriptor struct")
	}
	return &TestDescriptor{
		FrameworkID: decodeString(interp, structField(inst, "framework_id")),
		ModulePath:  decodeString(interp, structField(inst, "module_path")),
		TestID:      decodeString(interp, structField(inst, "test_id")),
		DisplayName: decodeString(interp, structField(inst, "display_name")),
		Tags:        decodeStringArray(interp, structField(inst, "tags")),
		Metadata:    decodeMetadataArray(interp, structField(inst, "metadata")),
		Location:    decodeLocation(interp, structField(inst, "location")),
	}, nil
}

func decodeFailure(interp *interpreter.Interpreter, value runtime.Value) (*FailureData, error) {
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil, fmt.Errorf("expected Failure struct")
	}
	return &FailureData{
		Message:  decodeString(interp, structField(inst, "message")),
		Details:  decodeOptionalString(interp, structField(inst, "details")),
		Location: decodeLocation(interp, structField(inst, "location")),
	}, nil
}

func decodeLocation(interp *interpreter.Interpreter, value runtime.Value) *SourceLocation {
	if isNilValue(value) {
		return nil
	}
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil
	}
	return &SourceLocation{
		ModulePath: decodeString(interp, structField(inst, "module_path")),
		Line:       int(decodeNumber(structField(inst, "line"))),
		Column:     int(decodeNumber(structField(inst, "column"))),
	}
}

func decodeMetadataArray(interp *interpreter.Interpreter, value runtime.Value) []MetadataEntry {
	if isNilValue(value) {
		return nil
	}
	arrayVal, err := coerceArrayValue(interp, value, "metadata array")
	if err != nil {
		return nil
	}
	out := make([]MetadataEntry, 0, len(arrayVal.Elements))
	for _, entry := range arrayVal.Elements {
		inst, ok := entry.(*runtime.StructInstanceValue)
		if !ok || inst == nil {
			continue
		}
		out = append(out, MetadataEntry{
			Key:   decodeString(interp, structField(inst, "key")),
			Value: decodeString(interp, structField(inst, "value")),
		})
	}
	return out
}

func decodeStringArray(interp *interpreter.Interpreter, value runtime.Value) []string {
	if isNilValue(value) {
		return nil
	}
	arrayVal, err := coerceArrayValue(interp, value, "string array")
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(arrayVal.Elements))
	for _, entry := range arrayVal.Elements {
		out = append(out, decodeString(interp, entry))
	}
	return out
}

func decodeString(_ *interpreter.Interpreter, value runtime.Value) string {
	switch v := value.(type) {
	case runtime.StringValue:
		return v.Val
	case *runtime.StringValue:
		if v != nil {
			return v.Val
		}
	}
	return runtimeValueToString(value)
}

func decodeOptionalString(interp *interpreter.Interpreter, value runtime.Value) *string {
	if isNilValue(value) {
		return nil
	}
	decoded := decodeString(interp, value)
	return &decoded
}

func decodeNumber(value runtime.Value) int64 {
	switch v := value.(type) {
	case runtime.IntegerValue:
		if v.Val != nil {
			return v.Val.Int64()
		}
	case *runtime.IntegerValue:
		if v != nil && v.Val != nil {
			return v.Val.Int64()
		}
	}
	return 0
}

func coerceArrayValue(interp *interpreter.Interpreter, value runtime.Value, label string) (*runtime.ArrayValue, error) {
	if value == nil {
		return nil, fmt.Errorf("expected %s", label)
	}
	switch v := value.(type) {
	case *runtime.ArrayValue:
		return v, nil
	default:
		if interp == nil {
			return nil, fmt.Errorf("expected %s", label)
		}
		return interp.CoerceArrayValue(value)
	}
}

func structField(inst *runtime.StructInstanceValue, name string) runtime.Value {
	if inst == nil || name == "" {
		return runtime.NilValue{}
	}
	if inst.Fields != nil {
		if value, ok := inst.Fields[name]; ok && value != nil {
			return value
		}
	}
	if inst.Definition != nil && inst.Definition.Node != nil && len(inst.Positional) > 0 {
		for idx, field := range inst.Definition.Node.Fields {
			if field != nil && field.Name != nil && field.Name.Name == name {
				if idx < len(inst.Positional) {
					return inst.Positional[idx]
				}
				break
			}
		}
	}
	return runtime.NilValue{}
}

func structTag(inst *runtime.StructInstanceValue) string {
	if inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
		return ""
	}
	return inst.Definition.Node.ID.Name
}

func isNilValue(value runtime.Value) bool {
	switch value.(type) {
	case nil:
		return true
	case runtime.NilValue:
		return true
	case *runtime.NilValue:
		return true
	default:
		return false
	}
}

func runtimeValueToString(value runtime.Value) string {
	switch v := value.(type) {
	case nil:
		return ""
	case runtime.StringValue:
		return v.Val
	case *runtime.StringValue:
		if v != nil {
			return v.Val
		}
	case runtime.BoolValue:
		if v.Val {
			return "true"
		}
		return "false"
	case runtime.IntegerValue:
		return v.String()
	case *runtime.IntegerValue:
		if v != nil {
			return v.String()
		}
	case runtime.FloatValue:
		return fmt.Sprintf("%g", v.Val)
	case runtime.CharValue:
		return string(v.Val)
	case runtime.NilValue, *runtime.NilValue:
		return "nil"
	}
	if value == nil {
		return ""
	}
	return fmt.Sprintf("<%s>", value.Kind())
}
