package testcli

type ReporterFormat string

const (
	ReporterJSON ReporterFormat = "json"
	ReporterTap  ReporterFormat = "tap"
)

type EventState struct {
	Total           int
	Failed          int
	Skipped         int
	FrameworkErrors int
}

type TestDescriptor struct {
	FrameworkID string          `json:"framework_id"`
	ModulePath  string          `json:"module_path"`
	TestID      string          `json:"test_id"`
	DisplayName string          `json:"display_name"`
	Tags        []string        `json:"tags"`
	Metadata    []MetadataEntry `json:"metadata"`
	Location    *SourceLocation `json:"location"`
}

type MetadataEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SourceLocation struct {
	ModulePath string `json:"module_path"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
}

type FailureData struct {
	Message  string          `json:"message"`
	Details  *string         `json:"details"`
	Location *SourceLocation `json:"location"`
}

type TestEvent struct {
	Kind       string          `json:"event"`
	Descriptor *TestDescriptor `json:"descriptor,omitempty"`
	DurationMs int64           `json:"duration_ms,omitempty"`
	Failure    *FailureData    `json:"failure,omitempty"`
	Reason     *string         `json:"reason,omitempty"`
	Message    string          `json:"message,omitempty"`
}
