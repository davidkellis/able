package interpreter

import "testing"

func TestExecutorKindFromEnvironmentDefaultsToSerial(t *testing.T) {
	t.Setenv(ExecutorEnvVar, "")

	got, err := ExecutorKindFromEnvironment()
	if err != nil {
		t.Fatalf("ExecutorKindFromEnvironment returned error: %v", err)
	}
	if got != "serial" {
		t.Fatalf("ExecutorKindFromEnvironment = %q, want serial", got)
	}
}

func TestExecutorKindFromEnvironmentAcceptsGoroutine(t *testing.T) {
	t.Setenv(ExecutorEnvVar, "goroutine")

	got, err := ExecutorKindFromEnvironment()
	if err != nil {
		t.Fatalf("ExecutorKindFromEnvironment returned error: %v", err)
	}
	if got != "goroutine" {
		t.Fatalf("ExecutorKindFromEnvironment = %q, want goroutine", got)
	}
}

func TestExecutorKindFromEnvironmentRejectsUnknownValue(t *testing.T) {
	t.Setenv(ExecutorEnvVar, "bogus")

	if _, err := ExecutorKindFromEnvironment(); err == nil {
		t.Fatalf("ExecutorKindFromEnvironment accepted invalid executor value")
	}
}

func TestNewExecutorFromEnvironment(t *testing.T) {
	t.Setenv(ExecutorEnvVar, "goroutine")

	exec, err := NewExecutorFromEnvironment()
	if err != nil {
		t.Fatalf("NewExecutorFromEnvironment returned error: %v", err)
	}
	if _, ok := exec.(*GoroutineExecutor); !ok {
		t.Fatalf("NewExecutorFromEnvironment returned %T, want *GoroutineExecutor", exec)
	}
}
