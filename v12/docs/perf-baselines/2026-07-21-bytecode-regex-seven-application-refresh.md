# Bytecode regex seven-application refresh

Date: 2026-07-21

## Decision

Keep no VM, compiler, runtime, canonical-stdlib, language, benchmark,
reference, or WASM change. Fresh profiles reproduce the generic Array-slot
read/cache path and small named-struct field checks across Regex Suffix, Regex
Set, Regex Stream, Log Routing/Redaction, and Configuration
Validation/Extraction. Word Frequency and Mandelbrot confirm that those exact
children belong to the regex/NFA carrier rather than a new corpus-wide VM
operation.

Both apparent candidates are already closed by causal evidence. The canonical
Array-slot direct cache is allocation-free and its remaining version checks
preserve dynamic runtime/member invalidation. The small named-field map
candidate was independently retested earlier today across unlike nominal
programs and regressed every row. The current profiles expose no new cache
miss, coherence rule, or representation that invalidates those results, so no
candidate advanced to implementation or timing.

## Reproducibility contract

One current interpreter test binary was frozen at SHA-256
`6758a13355a1adeebe0984098679c3ad344b0f8cf1a8642694e873e3dd12d53e`.
Each application ran twice for main-only CPU profiles and twice in independent
unprofiled processes for exact measured-main allocation counters. Every
process used CPU 0, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, the canonical
external stdlib, skipped repeated setup typechecking, and had a 59-second test
cap. Setup, lowering, output capture, and final forced collections were
outside the measured main window.

All seven sources match the promoted scorecard hashes. The five regex rows
used their full scorecard scale, not their smaller diagnostic `--profile`
scale. Word Frequency and Mandelbrot are unrelated text/map and numeric
discriminators. Active-workstation movement is retained in the arithmetic
means rather than selecting the fastest process.

| Application | Profiled main runs | Mean | Merged samples |
| --- | --- | ---: | ---: |
| Regex Suffix Audit | 2.933 s, 3.769 s | 3.351 s | 6.66 s |
| Regex Set Audit | 3.661 s, 4.295 s | 3.978 s | 7.92 s |
| Regex Stream Audit | 3.114 s, 3.232 s | 3.173 s | 6.30 s |
| Log Routing/Redaction | 3.272 s, 2.846 s | 3.059 s | 6.08 s |
| Configuration Validation/Extraction | 0.950 s, 0.931 s | 0.940 s | 1.85 s |
| Word Frequency | 1.565 s, 1.250 s | 1.408 s | 2.80 s |
| Mandelbrot | 9.931 s, 6.153 s | 8.042 s | 16.01 s |

## Exact CPU intersection

Admission required the same exact symbol to own at least 1% flat CPU in at
least three unlike applications. Related audit programs alone could not
establish a new general candidate, and aggregate dispatcher/runtime parents
were excluded.

| Exact symbol or family | Breadth | Flat shares | Disposition |
| --- | ---: | --- | --- |
| `(*bytecodeVM).runResumable` | 7 | 8.93%-18.06% | aggregate dispatcher parent, closed |
| `(*bytecodeVM).appendSlotStackValueChecked` | 7 | 2.14%-5.11% | closed stack-carrier family |
| `internal/runtime/maps.ctrlGroup.matchH2` | 6 | 1.64%-3.78% | different VM/runtime maps |
| `(*bytecodeVM).execLoadSlotOpcode` | 6 | 1.39%-2.70% | aggregate load parent |
| `runtime.mapaccess2_fast64` | 6 | 1.07%-1.64% | different Array/cache maps |
| canonical Array-slot call/read/cache leaves | 5 | 1.08%-5.95% | retained direct cache; proof path already closed |
| named-struct plan/storage leaves | 5 | 1.08%-2.16% | closed small-definition field family |
| `bytecodeRawIntegerValueInfo` | 5 | 1.32%-2.14% | closed raw-integer family; Word also reaches it |
| `bytecodeStackSnapshotValue` | 5 | 1.25%-3.24% | closed stack-carrier family; controls also reach it |
| `runtime.tryDeferToSpanScan` | 5 | 1.48%-4.31% | Go GC machinery |

The exact Array chain comprises
`execCallMemberArraySlot`,
`lookupCachedCanonicalArraySlotCallForArrayValidatedWithVersions`, and
`finishArrayReadSlotMemberFast`. It clears 1% flat in all five regex rows but
not in either discriminator. Caller traces put it at canonical parallel-array
NFA transition/thread reads, agreeing with the prior source-site counters.
The cache already has direct instruction-indexed entries and previously
measured about 2.2 ns per allocation-free hit. Its larger prefix proves that
runtime data is absent and validates global/method revisions; deleting that
work would change member semantics after mutation.

Named-field plan/storage checks also recur in all three regex workload
families. They cover diverse NFA transition, thread, capture, character, and
builder definitions, so they are general nominal work rather than a named
Regex shortcut. That fact still does not reopen the representation: enabling
the existing name-to-index map for small definitions regressed Wide Records
5.19%, Unicode 1.44%, Log Routing 5.10%, Inventory 0.83%, and Mandelbrot
2.60%. Direct comparison remains faster for one-to-four-field definitions.

Word Frequency is instead led by call/return, hashing, raw integers, and
type-match work. Mandelbrot is led by fused float branching, allocation, and
raw-float storage. Their shared stack/dispatcher/runtime parents do not resolve
to the NFA Array or nominal-field operations.

## Exact measured-main allocation

Object-count pairs were stable to 43 allocations. Byte spans in four regex or
text rows reflect small setup/GC timing differences but remain below 0.25% of
their totals; Mandelbrot differed by only 128 bytes and six allocations.

| Application | Mean bytes | Mean allocations | Mean frees | Mean GCs |
| --- | ---: | ---: | ---: | ---: |
| Regex Suffix Audit | 53,705,988 | 515,260 | 179,803.5 | 3.0 |
| Regex Set Audit | 43,981,260 | 309,064.5 | 119,172 | 2.0 |
| Regex Stream Audit | 39,311,256 | 483,446 | 170,344 | 2.0 |
| Log Routing/Redaction | 23,339,588 | 175,583 | 71,539.5 | 1.5 |
| Configuration Validation/Extraction | 17,208,000 | 147,423.5 | 57,119.5 | 1.0 |
| Word Frequency | 54,148,152 | 625,738 | 350,086.5 | 3.0 |
| Mandelbrot | 615,167,064 | 76,303,228 | 75,379,795.5 | 33.0 |

The allocation shapes remain consistent with the prior ownership profiles:
regex rows allocate NFA thread/capture/codepoint structures around the
already-retained closure scratch and primitive thread carrier; Word Frequency
allocates text/map/call values; Mandelbrot materializes raw-float results. No
one new allocation boundary spans three unlike workload families.

## Verification

The unchanged production tree passes the complete bytecode family:

```text
go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  26.280s
```

The strict 83-row frontier and scorecard-evidence checks pass after recording
the refreshed ownership evidence. Temporary binaries, profiles, output
captures, and allocation reports were removed.

## Next recommendation

Refresh the `compiled-concurrency` ownership group across the eleven selected
concurrency and feature-interaction applications, with NBody and short
non-concurrent startup controls.

Why: after refreshing the four largest bytecode ownership groups, compiled
concurrency is the largest remaining group at 12.904 target-excess seconds.
Its shared `bridge.currentGID`/`runtime.Stack` wall is exceptionally large,
but the old fixed-context replacement regressed NBody 54.7%. Current generated
binaries include subsequent bridge and lazy-value changes, so a same-binary
refresh can determine whether the guard failure still has the same cause or a
new bounded identity propagation design is now visible.

What it entails: build current generated binaries, collect repeated
verifier-backed wall means plus main-only CPU/allocation profiles for distinct
channel, future, mutex, indexing, dependency, routing, and validation shapes,
and trace goroutine-identity calls by semantic consumer. Any candidate must be
a general concurrency-context rule, preserve dynamically entered/native
goroutines, and pass NBody plus current compiled target guards. Do not assume a
fixed context, add application or named-container lowering, or begin WASM.
