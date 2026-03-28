package compiler

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

const (
	compilerExecBatchIndexEnv            = "ABLE_COMPILER_EXEC_FIXTURE_BATCH_INDEX"
	compilerExecBatchCountEnv            = "ABLE_COMPILER_EXEC_FIXTURE_BATCH_COUNT"
	compilerStrictDispatchBatchIndexEnv  = "ABLE_COMPILER_STRICT_DISPATCH_BATCH_INDEX"
	compilerStrictDispatchBatchCountEnv  = "ABLE_COMPILER_STRICT_DISPATCH_BATCH_COUNT"
	compilerInterfaceLookupBatchIndexEnv = "ABLE_COMPILER_INTERFACE_LOOKUP_BATCH_INDEX"
	compilerInterfaceLookupBatchCountEnv = "ABLE_COMPILER_INTERFACE_LOOKUP_BATCH_COUNT"
	compilerBoundaryAuditBatchIndexEnv   = "ABLE_COMPILER_BOUNDARY_AUDIT_BATCH_INDEX"
	compilerBoundaryAuditBatchCountEnv   = "ABLE_COMPILER_BOUNDARY_AUDIT_BATCH_COUNT"
)

func applyCompilerFixtureBatch(t *testing.T, fixtures []string, indexEnv, countEnv string) []string {
	t.Helper()
	if len(fixtures) == 0 {
		return fixtures
	}
	rawIndex := strings.TrimSpace(os.Getenv(indexEnv))
	rawCount := strings.TrimSpace(os.Getenv(countEnv))
	if rawIndex == "" && rawCount == "" {
		return fixtures
	}
	if rawIndex == "" || rawCount == "" {
		t.Fatalf("fixture batching requires both %s and %s", indexEnv, countEnv)
	}

	batchIndex, err := strconv.Atoi(rawIndex)
	if err != nil {
		t.Fatalf("invalid %s=%q: %v", indexEnv, rawIndex, err)
	}
	batchCount, err := strconv.Atoi(rawCount)
	if err != nil {
		t.Fatalf("invalid %s=%q: %v", countEnv, rawCount, err)
	}
	if batchCount <= 0 {
		t.Fatalf("invalid %s=%d: must be > 0", countEnv, batchCount)
	}
	if batchIndex < 0 || batchIndex >= batchCount {
		t.Fatalf("invalid %s=%d for %s=%d", indexEnv, batchIndex, countEnv, batchCount)
	}

	batch := make([]string, 0, (len(fixtures)+batchCount-1)/batchCount)
	for idx, rel := range fixtures {
		if idx%batchCount == batchIndex {
			batch = append(batch, rel)
		}
	}
	return batch
}
