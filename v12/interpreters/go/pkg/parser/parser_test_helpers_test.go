package parser

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"able/interpreter-go/pkg/ast"
)

type fixtureCase struct {
	name   string
	source string
}

func checkSpan(t testing.TB, label string, span ast.Span, startLine, startCol, endLine, endCol int) {
	t.Helper()
	if span.Start.Line != startLine || span.Start.Column != startCol {
		t.Fatalf("%s start span mismatch: got (%d,%d), want (%d,%d)", label, span.Start.Line, span.Start.Column, startLine, startCol)
	}
	if span.End.Line != endLine || span.End.Column != endCol {
		t.Fatalf("%s end span mismatch: got (%d,%d), want (%d,%d)", label, span.End.Line, span.End.Column, endLine, endCol)
	}
}

func assertModulesEqual(t testing.TB, expected interface{}, actual interface{}) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		return
	}
	wantJSON, _ := json.Marshal(expected)
	gotJSON, _ := json.Marshal(actual)
	var wantAny interface{}
	var gotAny interface{}
	_ = json.Unmarshal(wantJSON, &wantAny)
	_ = json.Unmarshal(gotJSON, &gotAny)
	if reflect.DeepEqual(wantAny, gotAny) {
		return
	}
	wantPretty, _ := json.MarshalIndent(wantAny, "", "  ")
	gotPretty, _ := json.MarshalIndent(gotAny, "", "  ")
	t.Fatalf("module mismatch\nexpected: %s\n   actual: %s", wantPretty, gotPretty)
}

func runFixtureCases(t *testing.T, cases []fixtureCase) {
	t.Helper()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			source := tc.source
			if source == "" {
				source = loadFixtureSource(t, tc.name)
			}
			p, err := NewModuleParser()
			if err != nil {
				t.Fatalf("NewModuleParser error: %v", err)
			}
			defer p.Close()

			mod, err := p.ParseModule([]byte(source))
			if err != nil {
				t.Fatalf("ParseModule error for %s: %v", tc.name, err)
			}
			NormalizeFixtureModule(mod)

			expected := loadFixtureModule(t, tc.name)
			assertModulesEqual(t, expected, mod)
		})
	}
}

func collectFixtureCases(t *testing.T, category string) []fixtureCase {
	t.Helper()
	root := filepath.Join(fixturesRoot(), category)
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var cases []fixtureCase
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := filepath.Join(category, entry.Name())
		cases = append(cases, fixtureCase{name: name})
	}
	sort.Slice(cases, func(i, j int) bool {
		return cases[i].name < cases[j].name
	})
	return cases
}

func loadFixtureSource(t testing.TB, fixture string) string {
	t.Helper()
	path := filepath.Join(fixturesRoot(), fixture, "source.able")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ""
		}
		t.Fatalf("read fixture source %s: %v", fixture, err)
	}
	return string(data)
}

func loadFixtureModule(t testing.TB, fixture string) map[string]any {
	t.Helper()
	path := filepath.Join(fixturesRoot(), fixture, "module.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture module %s: %v", fixture, err)
	}
	var module map[string]any
	if err := json.Unmarshal(data, &module); err != nil {
		t.Fatalf("unmarshal fixture module %s: %v", fixture, err)
	}
	return module
}

func fixturesRoot() string {
	return filepath.Join("..", "..", "..", "..", "fixtures", "ast")
}
