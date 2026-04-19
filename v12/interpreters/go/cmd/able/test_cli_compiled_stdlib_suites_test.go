package main

import "testing"

var compiledStdlibSuiteCases = map[string]compiledStdlibCase{
	"BigintAndBiguintSuites": {
		label:    "stdlib bigint/biguint",
		relPaths: []string{"bigint.test.able", "biguint.test.able"},
		expected: []string{"BigInt", "BigUint", "raises on underflow"},
	},
	"ExtendedNumericSuites": {
		label:    "stdlib int128/uint128/rational",
		relPaths: []string{"int128.test.able", "uint128.test.able", "rational.test.able"},
		expected: []string{"Int128", "UInt128", "Rational", "round-trips through display helpers"},
	},
	"NumbersNumericSuite": {
		label:    "stdlib numbers_numeric",
		relPaths: []string{"numbers_numeric.test.able"},
		expected: []string{"Numeric primitives", "covers f64 fractional helpers"},
	},
	"FoundationalSuites": {
		label:    "foundational stdlib suites",
		relPaths: []string{"simple.test.able", "assertions.test.able", "enumerable.test.able"},
		expected: []string{"simple suite verifies addition works", "able.spec assertions passes equality matcher", "Enumerable helpers maps and filters arrays"},
	},
	"CollectionsListVectorSuites": {
		label:    "collections list/vector suites",
		relPaths: []string{"list.test.able", "vector.test.able"},
		expected: []string{"List supports prepend/head/tail with structural sharing", "Vector supports set without mutating prior versions"},
	},
	"CollectionsTreeMapTreeSetSuites": {
		label:    "collections tree_map/tree_set suites",
		relPaths: []string{"tree_map.test.able", "tree_set.test.able"},
		expected: []string{"TreeMap inserts, updates, and retrieves entries", "TreeSet inserts unique values and iterates in order"},
	},
	"CollectionsPersistentMapPersistentSetSuites": {
		label:    "collections persistent_map/persistent_set suites",
		relPaths: []string{"persistent_map.test.able", "persistent_set.test.able"},
		expected: []string{"PersistentMap stores, reads, and updates entries", "PersistentSet unions and intersects"},
	},
	"CollectionsPersistentSortedSetAndQueueSuites": {
		label:    "collections persistent_sorted_set/persistent_queue suites",
		relPaths: []string{"persistent_sorted_set.test.able", "persistent_queue.test.able"},
		expected: []string{"PersistentSortedSet keeps values ordered and unique", "PersistentQueue iterates values in FIFO order"},
	},
	"CollectionsLinkedListAndLazySeqSuites": {
		label:    "collections linked_list/lazy_seq suites",
		relPaths: []string{"linked_list.test.able", "lazy_seq.test.able"},
		expected: []string{"LinkedList pushes and pops from both ends", "LazySeq iterates with caching and produces arrays"},
	},
	"CollectionsHashMapSmokeAndHashSetSuites": {
		label:    "collections hash_map_smoke/hash_set suites",
		relPaths: []string{"collections/hash_map_smoke.test.able", "collections/hash_set.test.able"},
		expected: []string{"HashSet adds, removes, and checks membership", "HashSet subset, superset, and disjoint checks"},
	},
	"CollectionsDequeAndQueueSmokeSuites": {
		label:    "collections deque_smoke/queue_smoke suites",
		relPaths: []string{"collections/deque_smoke.test.able", "collections/queue_smoke.test.able"},
		expected: []string{"able test: no tests to run"},
	},
	"CollectionsBitSetAndHeapSuites": {
		label:    "collections bit_set/heap suites",
		relPaths: []string{"bit_set.test.able", "heap.test.able"},
		expected: []string{"BitSet sets, checks, and resets bits", "Heap pushes and pops smallest values first"},
	},
	"CollectionsArrayAndRangeSmokeSuites": {
		label:    "collections array_smoke/range_smoke suites",
		relPaths: []string{"collections/array_smoke.test.able", "collections/range_smoke.test.able"},
		expected: []string{"able test: no tests to run"},
	},
	"ConcurrencyChannelMutexAndQueueSuites": {
		label:    "concurrency channel_mutex/concurrent_queue suites",
		relPaths: []string{"concurrency/channel_mutex.test.able", "concurrency/concurrent_queue.test.able"},
		expected: []string{"Channel supports send/receive/close operations", "ConcurrentQueue supports try operations and close"},
	},
	"MathAndCoreNumericSuites": {
		label:    "math/core numeric suites",
		relPaths: []string{"math.test.able", "core/numeric_smoke.test.able"},
		expected: []string{"able.math computes gcd/lcm for integers", "able.math offers rounding helpers"},
	},
	"FsAndPathSmokeSuites": {
		label:    "fs/path smoke suites",
		relPaths: []string{"fs_smoke.test.able", "path_smoke.test.able"},
		expected: []string{"able test: no tests to run"},
	},
	"IoSmokeSuite": {
		label:    "io smoke suite",
		relPaths: []string{"io_smoke.test.able"},
		expected: []string{"able test: no tests to run"},
	},
	"OsSmokeSuite": {
		label:    "os smoke suite",
		relPaths: []string{"os_smoke.test.able"},
		expected: []string{"able test: no tests to run"},
	},
	"ProcessSmokeSuite": {
		label:    "process smoke suite",
		relPaths: []string{"process_smoke.test.able"},
		expected: []string{"able test: no tests to run"},
	},
	"TermSmokeSuite": {
		label:    "term smoke suite",
		relPaths: []string{"term_smoke.test.able"},
		expected: []string{"able test: no tests to run"},
	},
	"HarnessReportersSmokeSuite": {
		label:    "harness/reporters smoke suite",
		relPaths: []string{"harness_reporters_smoke.test.able"},
		expected: []string{"able test: no tests to run"},
	},
	"TextStringSuites": {
		label:    "text/string suites",
		relPaths: []string{"text/string_methods.test.able", "text/string_split.test.able", "text/string_builder.test.able", "text/string_smoke.test.able"},
		expected: []string{"String methods reports lengths and prefixes/suffixes", "String split/join joins and concats strings", "StringBuilder pushes strings and finishes"},
	},
}

func runNamedCompiledStdlibCase(t *testing.T, name string) {
	t.Helper()

	tc, ok := compiledStdlibSuiteCases[name]
	if !ok {
		t.Fatalf("missing compiled stdlib suite case %q", name)
	}
	runCompiledStdlibCase(t, tc)
}

func TestTestCommandCompiledRunsStdlibBigintAndBiguintSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "BigintAndBiguintSuites")
}

func TestTestCommandCompiledRunsStdlibExtendedNumericSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "ExtendedNumericSuites")
}

func TestTestCommandCompiledRunsStdlibNumbersNumericSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "NumbersNumericSuite")
}

func TestTestCommandCompiledRunsStdlibFoundationalSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "FoundationalSuites")
}

func TestTestCommandCompiledRunsStdlibCollectionsListVectorSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsListVectorSuites")
}

func TestTestCommandCompiledRunsStdlibCollectionsTreeMapTreeSetSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsTreeMapTreeSetSuites")
}

func TestTestCommandCompiledRunsStdlibCollectionsPersistentMapPersistentSetSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsPersistentMapPersistentSetSuites")
}

func TestTestCommandCompiledRunsStdlibCollectionsPersistentSortedSetAndQueueSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsPersistentSortedSetAndQueueSuites")
}

func TestTestCommandCompiledRunsStdlibCollectionsLinkedListAndLazySeqSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsLinkedListAndLazySeqSuites")
}

func TestTestCommandCompiledRunsStdlibCollectionsHashMapSmokeAndHashSetSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsHashMapSmokeAndHashSetSuites")
}

func TestTestCommandCompiledRunsStdlibCollectionsDequeAndQueueSmokeSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsDequeAndQueueSmokeSuites")
}

func TestTestCommandCompiledRunsStdlibCollectionsBitSetAndHeapSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsBitSetAndHeapSuites")
}

func TestTestCommandCompiledRunsStdlibCollectionsArrayAndRangeSmokeSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsArrayAndRangeSmokeSuites")
}

func TestTestCommandCompiledRunsStdlibConcurrencyChannelMutexAndQueueSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "ConcurrencyChannelMutexAndQueueSuites")
}

func TestTestCommandCompiledRunsStdlibMathAndCoreNumericSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "MathAndCoreNumericSuites")
}

func TestTestCommandCompiledRunsStdlibFsAndPathSmokeSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "FsAndPathSmokeSuites")
}

func TestTestCommandCompiledRunsStdlibIoSmokeSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "IoSmokeSuite")
}

func TestTestCommandCompiledRunsStdlibOsSmokeSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "OsSmokeSuite")
}

func TestTestCommandCompiledRunsStdlibProcessSmokeSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "ProcessSmokeSuite")
}

func TestTestCommandCompiledRunsStdlibTermSmokeSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "TermSmokeSuite")
}

func TestTestCommandCompiledRunsStdlibHarnessReportersSmokeSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "HarnessReportersSmokeSuite")
}

func TestTestCommandCompiledRunsStdlibTextStringSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "TextStringSuites")
}
