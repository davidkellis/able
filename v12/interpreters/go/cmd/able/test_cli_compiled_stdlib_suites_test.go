package main

import "testing"

var compiledStdlibSuiteCases = map[string]compiledStdlibCase{
	"BigintSuite": {
		label:    "stdlib bigint",
		relPaths: []string{"bigint.test.able"},
		expected: []string{"BigInt"},
	},
	"BiguintSuite": {
		label:    "stdlib biguint",
		relPaths: []string{"biguint.test.able"},
		expected: []string{"BigUint", "raises on underflow"},
	},
	"Int128Suite": {
		label:    "stdlib int128",
		relPaths: []string{"int128.test.able"},
		expected: []string{"Int128"},
	},
	"UInt128Suite": {
		label:    "stdlib uint128",
		relPaths: []string{"uint128.test.able"},
		expected: []string{"UInt128"},
	},
	"RationalSuite": {
		label:    "stdlib rational",
		relPaths: []string{"rational.test.able"},
		expected: []string{"Rational", "round-trips through display helpers"},
	},
	"NumbersNumericSuite": {
		label:    "stdlib numbers_numeric",
		relPaths: []string{"numbers_numeric.test.able"},
		expected: []string{"Numeric primitives", "covers f64 fractional helpers"},
	},
	"FoundationalSimpleSuite": {
		label:    "foundational stdlib simple suite",
		relPaths: []string{"simple.test.able"},
		expected: []string{"simple suite verifies addition works"},
	},
	"FoundationalAssertionsSuite": {
		label:    "foundational stdlib assertions suite",
		relPaths: []string{"assertions.test.able"},
		expected: []string{"able.spec assertions passes equality matcher"},
	},
	"FoundationalEnumerableArraySuite": {
		label:    "foundational stdlib enumerable Array suite",
		relPaths: []string{"enumerable.test.able"},
		expected: []string{"Enumerable helpers maps and filters arrays"},
	},
	"CollectionsListSuite": {
		label:    "collections list suite",
		relPaths: []string{"list.test.able"},
		expected: []string{"List supports prepend/head/tail with structural sharing"},
	},
	"CollectionsVectorSuite": {
		label:    "collections vector suite",
		relPaths: []string{"vector.test.able"},
		expected: []string{"Vector supports set without mutating prior versions"},
	},
	"CollectionsTreeMapTreeSetSuites": {
		label:    "collections tree_map/tree_set suites",
		relPaths: []string{"tree_map.test.able", "tree_set.test.able"},
		expected: []string{"TreeMap inserts, updates, and retrieves entries", "TreeSet inserts unique values and iterates in order"},
	},
	"CollectionsPersistentMapSuite": {
		label:    "collections persistent_map suite",
		relPaths: []string{"persistent_map.test.able"},
		expected: []string{"PersistentMap stores, reads, and updates entries"},
	},
	"CollectionsPersistentSetSuite": {
		label:    "collections persistent_set suite",
		relPaths: []string{"persistent_set.test.able"},
		expected: []string{"PersistentSet unions and intersects"},
	},
	"CollectionsPersistentSortedSetSuite": {
		label:    "collections persistent_sorted_set suite",
		relPaths: []string{"persistent_sorted_set.test.able"},
		expected: []string{"PersistentSortedSet keeps values ordered and unique"},
	},
	"CollectionsPersistentQueueSuite": {
		label:    "collections persistent_queue suite",
		relPaths: []string{"persistent_queue.test.able"},
		expected: []string{"PersistentQueue iterates values in FIFO order"},
	},
	"CollectionsLinkedListSuite": {
		label:    "collections linked_list suite",
		relPaths: []string{"linked_list.test.able"},
		expected: []string{"LinkedList pushes and pops from both ends"},
	},
	"CollectionsLazySeqSuite": {
		label:    "collections lazy_seq suite",
		relPaths: []string{"lazy_seq.test.able"},
		expected: []string{"LazySeq iterates with caching and produces arrays"},
	},
	"CollectionsHashMapSmokeSuite": {
		label:    "collections hash_map_smoke suite",
		relPaths: []string{"collections/hash_map_smoke.test.able"},
		expected: []string{"able test: no tests to run"},
	},
	"CollectionsHashSetSuite": {
		label:    "collections hash_set suite",
		relPaths: []string{"collections/hash_set.test.able"},
		expected: []string{"HashSet adds, removes, and checks membership", "HashSet subset, superset, and disjoint checks", "HashSet preserves collection type for eager map"},
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

func TestTestCommandCompiledRunsStdlibBigintSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "BigintSuite")
}

func TestTestCommandCompiledRunsStdlibBiguintSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "BiguintSuite")
}

func TestTestCommandCompiledRunsStdlibInt128Suite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "Int128Suite")
}

func TestTestCommandCompiledRunsStdlibUInt128Suite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "UInt128Suite")
}

func TestTestCommandCompiledRunsStdlibRationalSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "RationalSuite")
}

func TestTestCommandCompiledRunsStdlibNumbersNumericSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "NumbersNumericSuite")
}

func TestTestCommandCompiledRunsStdlibFoundationalSimpleSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "FoundationalSimpleSuite")
}

func TestTestCommandCompiledRunsStdlibFoundationalAssertionsSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "FoundationalAssertionsSuite")
}

func TestTestCommandCompiledRunsStdlibFoundationalEnumerableArraySuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "FoundationalEnumerableArraySuite")
}

func TestTestCommandCompiledRunsStdlibCollectionsListSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsListSuite")
}

func TestTestCommandCompiledRunsStdlibCollectionsVectorSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsVectorSuite")
}

func TestTestCommandCompiledRunsStdlibCollectionsTreeMapTreeSetSuites(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsTreeMapTreeSetSuites")
}

func TestTestCommandCompiledRunsStdlibCollectionsPersistentMapSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsPersistentMapSuite")
}

func TestTestCommandCompiledRunsStdlibCollectionsPersistentSetSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsPersistentSetSuite")
}

func TestTestCommandCompiledRunsStdlibCollectionsPersistentSortedSetSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsPersistentSortedSetSuite")
}

func TestTestCommandCompiledRunsStdlibCollectionsPersistentQueueSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsPersistentQueueSuite")
}

func TestTestCommandCompiledRunsStdlibCollectionsLinkedListSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsLinkedListSuite")
}

func TestTestCommandCompiledRunsStdlibCollectionsLazySeqSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsLazySeqSuite")
}

func TestTestCommandCompiledRunsStdlibCollectionsHashMapSmokeSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsHashMapSmokeSuite")
}

func TestTestCommandCompiledRunsStdlibCollectionsHashSetSuite(t *testing.T) {
	runNamedCompiledStdlibCase(t, "CollectionsHashSetSuite")
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
