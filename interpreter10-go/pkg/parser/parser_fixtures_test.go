package parser

import (
	"os"
	"testing"
)

func TestParseConcurrencyFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "concurrency"))
}

func TestParseControlFlowFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "control"))
}

func TestParseTypeExpressionFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "types"))
}

func TestParseBasicsFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "basics"))
}

func TestParseErrorFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "errors"))
}

func TestParseExpressionFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "expressions"))
}

func TestParseFunctionFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "functions"))
}

func TestParseImportFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "imports"))
}

func TestParseInterfaceFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "interfaces"))
}

func TestParseMatchFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "match"))
}

func TestParsePatternFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "patterns"))
}

func TestParsePipeFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "pipes"))
}

func TestParsePrivacyFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "privacy"))
}

func TestParseStdlibFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "stdlib"))
}

func TestParseStringFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "strings"))
}

func TestParseStructFixtures(t *testing.T) {
	skipFixtureTests(t)
	runFixtureCases(t, collectFixtureCases(t, "structs"))
}

func skipFixtureTests(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("parser fixture corpus skipped in short test mode")
	}
	if val := os.Getenv("GO_PARSER_FIXTURES"); val == "1" {
		return
	}
	t.Skip("parser fixture corpus temporarily disabled pending shared fixture decoder")
}
