//go:build js && wasm

package interpreter

// testingT captures the subset of testing.T used by fixture helpers.
type testingT interface {
	Helper()
	Fatalf(format string, args ...any)
}
