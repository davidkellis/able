package profilehook

import "testing"

func TestRegisterStopHookRunsUntilUnregistered(t *testing.T) {
	calls := 0
	unregister := RegisterStopHook(func() { calls++ })
	runStopHooks()
	if calls != 1 {
		t.Fatalf("hook calls = %d, want 1", calls)
	}
	unregister()
	unregister()
	runStopHooks()
	if calls != 1 {
		t.Fatalf("hook calls after unregister = %d, want 1", calls)
	}
}
