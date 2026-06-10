package compiler

import (
	"strings"
	"testing"
)

func TestCompilerChannelPayloadRecoveryStartsAtBlockingPath(t *testing.T) {
	result := compileExecFixtureResult(t, "06_12_19_stdlib_concurrency_channel_mutex_queue")

	sendBody, ok := findCompiledFunction(result, "__able_channel_send_impl")
	if !ok {
		t.Fatal("__able_channel_send_impl helper not found")
	}
	sendFastPath := strings.Index(sendBody, "select {")
	sendPayload := strings.Index(sendBody, "payload := __able_current_payload()")
	if sendFastPath < 0 || sendPayload < 0 || sendPayload < sendFastPath {
		t.Fatalf("channel send must recover its payload only after its non-blocking fast path:\n%s", sendBody)
	}
	assertSingleLocalizedChannelPayloadRecovery(t, sendBody)

	receiveBody, ok := findCompiledFunction(result, "__able_channel_receive_impl")
	if !ok {
		t.Fatal("__able_channel_receive_impl helper not found")
	}
	receiveFastPath := strings.Index(receiveBody, "select {")
	receivePayload := strings.Index(receiveBody, "payload := __able_current_payload()")
	if receiveFastPath < 0 || receivePayload < 0 || receivePayload < receiveFastPath {
		t.Fatalf("channel receive must recover its payload only after its non-blocking fast path:\n%s", receiveBody)
	}
	if strings.Contains(receiveBody, "if len(ch) > 0") {
		t.Fatalf("channel receive must not race a length check against multiple receivers:\n%s", receiveBody)
	}
	fastPath := receiveBody[receiveFastPath:receivePayload]
	if !strings.Contains(fastPath, "case value := <-ch:") ||
		!strings.Contains(fastPath, "default:") {
		t.Fatalf("channel receive fast path must use a non-blocking receive:\n%s", receiveBody)
	}
	assertSingleLocalizedChannelPayloadRecovery(t, receiveBody)
}

func assertSingleLocalizedChannelPayloadRecovery(t *testing.T, body string) {
	t.Helper()
	if got := strings.Count(body, "payload := __able_current_payload()"); got != 1 {
		t.Fatalf("channel blocking path must recover its payload exactly once, got %d:\n%s", got, body)
	}
	for _, stale := range []string{
		"__able_mark_current_task_blocked()",
		"__able_mark_current_task_unblocked()",
	} {
		if strings.Contains(body, stale) {
			t.Fatalf("channel blocking path must reuse its recovered payload instead of calling %s:\n%s", stale, body)
		}
	}
	if !strings.Contains(body, "MarkBlocked(payload.handle)") ||
		!strings.Contains(body, "MarkUnblocked(payload.handle)") {
		t.Fatalf("channel blocking path must preserve scheduler block bookkeeping with the recovered payload:\n%s", body)
	}
}
