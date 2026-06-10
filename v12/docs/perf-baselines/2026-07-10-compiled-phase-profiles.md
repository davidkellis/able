# Generated-Binary Phase Profiles: Bootstrap vs Main

This tranche adds attribution rather than an application optimization. Generated
Able binaries can now opt into separate CPU profiles for bootstrap/package
registration and user `main` execution with `ABLE_GO_PHASE_PROFILE_DIR`. The
normal profiling variables and normal program behavior are unchanged. A phase
profile writes `bootstrap.cpu.pprof`, `main.cpu.pprof`,
`{bootstrap,main}-{start,end}.allocs.pprof`, and `phase-stats.json` to the
requested directory. The JSON contains allocation bytes/counts, frees,
live-heap deltas, heap-object deltas, and GC counts for each phase; it
intentionally cannot be combined with `ABLE_GO_CPU_PROFILE`, because Go
permits only one active CPU profiler.

The generated launcher starts the bootstrap phase before package registration
and switches to the main phase immediately before `RunRegisteredMain`. This is
a generated-binary facility, not a benchmark feature.

## Method

The canonical stdlib was pinned at
`/home/david/sync/projects/able-stdlib/src`. Each binary ran with
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and CPU affinity `0`. Independent
normal-process profiles were merged so the short bootstrap segment supplied
useful samples.

| Workload | Launches | Bootstrap samples | Main samples |
| --- | ---: | ---: | ---: |
| Document Audit | 60 | 200 ms | 30 ms |
| Lexical Rollup | 60 | 140 ms | 1.73 s |
| i-before-e | 60 | 180 ms | 50 ms |

The retained merged profiles are
`20260710_{document_audit,lexical_rollup,i_before_e}_compiled_phase_{bootstrap,main}.cpu.pprof`
under `v12/interpreters/go/.profiles/`.

## Evidence

The main phases have the expected divergent user work: Document Audit spends
time in its public iterator/string pipeline, Lexical Rollup in `read_lines`,
splitting, iterator generation, and String containment, and i-before-e in a
short direct file/string loop. Those differences do not authorize an
application-specific lowering path.

Bootstrap instead repeats a concrete compiler metadata boundary in all three
programs. Generated package definitions deserialize default bodies as AST JSON
through `interpreter.DecodeNodeJSON`:

| Workload | `DecodeNodeJSON` bootstrap CPU | Bootstrap share |
| --- | ---: | ---: |
| Document Audit | 60 ms | 30.0% |
| Lexical Rollup | 10 ms | 7.1% |
| i-before-e | 70 ms | 38.9% |

The containing registration path also repeats (`RegisterIn` and generated
`__able_register_compiled_packages`). The current generator creates this work
for block-expression defaults, methods definitions, and implementation
definitions by serializing AST nodes in
`pkg/compiler/generator_export_defs.go` and decoding them while registering the
generated package. This is independent of container type, input corpus, and
the application source.

### Bootstrap allocation refresh

The retained phase hook now records allocation-counter deltas without changing
ordinary generated binaries. Fresh 240-launch merges used the same canonical
stdlib, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and CPU-0 guardrails.
The bootstrap allocation counts repeat strongly:

| Workload | Bytes/launch | Allocations/launch | Heap objects/launch | GC/launch |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 4,332,875 | 16,129 | 14,475 | 0.35 |
| Lexical Rollup | 4,332,013 | 16,126 | 14,472 | 0.34 |
| i-before-e | 4,094,215 | 12,120 | 11,155 | 0.39 |

The corresponding cumulative bootstrap CPU profiles retain `RegisterIn`
(60.3%/53.7%/42.0%) and `DecodeNodeJSON`
(29.4%/17.9%/27.5% including descendants) in all three applications, together
with GC scanning. Map growth and `Environment.Define` appear, but are small or
not repeated as the same concrete descendant. The allocation counters prove a
shared bootstrap allocation wall, but do not assign its bytes to one safe
registry/map leaf; they are insufficient authorization for a capacity or map
change.

Retained refreshed artifacts are
`20260710_{document_audit,lexical_rollup,i_before_e}_phase_stats_{bootstrap,main}.cpu.pprof`
and their matching `_allocations.json` summaries under
`v12/interpreters/go/.profiles/`.

### Exact phase-allocation source refresh

The start/end allocation snapshots now temporarily set Go's allocation sample
rate to one byte, drain the runtime's two-GC profile lag at each boundary, and
restore the original rate when the opt-in profiler stops. Ordinary binaries
never take this path. Twelve independent guarded launches of each verified
binary were merged for this source-attribution pass. The retained profiles are
`20260710_{document_audit,lexical_rollup,i_before_e}_phase_allocs_exact_{bootstrap,main}-{start,end}.pprof`.

The collector itself necessarily allocates CPU-profile writers and serialized
profile data. The bootstrap differences therefore exclude sample paths through
`runtime/pprof` and `compress/flate`; main allocation differences are not used
because their short phases are dominated by that collector footprint. With
that declared instrumentation removed, the same two production boundaries
repeat in all three programs:

| Bootstrap source, per launch | Document Audit | Lexical Rollup | i-before-e |
| --- | ---: | ---: | ---: |
| `newInterpreter` cumulative allocation | 687 KiB | 687 KiB | 687 KiB |
| `DecodeNodeJSON` family allocation | 370 KiB | 370 KiB | 212 KiB |
| JSON object-interface objects | 4,652 | 4,652 | 2,657 |

`newInterpreter`'s 570 KiB direct allocation per launch comes almost entirely
from eager initial capacity for the equality-dispatch, bound-method, and two
method-scope caches. JSON decoding is the same generic AST metadata boundary
already measured by CPU: its concrete leaves are `objectInterface`,
`arrayInterface`, `literalInterface`, unquoting, and AST node construction.
Neither result depends on an application, corpus, package name, or nominal
container.

## Decision

Keep the phase profiler. It has made a repeated, generic compiler boundary
measurable without perturbing ordinary binaries. No production lowering change
is included in this tranche.

### Rejected direct-constructor candidate

The first candidate replaced every generated `DecodeNodeJSON` call with a
recursive Go-constructor renderer for the complete shared AST contract. It
covered expressions, patterns, statements, definitions, generic metadata, and
the fields not accepted by constructors, then passed the focused metadata tests
and the complete `./pkg/compiler` suite.

It is not retained. Building the canonical-stdlib-precompiled Document Audit
application with the candidate remained in generated-source parsing after more
than six CPU minutes, before writing any output file. A non-mutating stack
sample showed `go/parser.(*parser).next`; the emitted constructor graph for the
full stdlib metadata set was too large for an ordinary application build. The
run was stopped before a binary or phase profile existed. This is a broad
compiler-throughput guard failure, not a benchmark-specific timing result, so
the implementation was reverted in full.

### Rejected compact tagged-codec candidate

The second candidate retained bounded source: it encoded the complete exported
AST contract as one tagged binary payload, emitted it as base64, and decoded it
without JSON maps. It round-tripped every current AST-node factory and a rich
nested definition graph, passed generated-Go build coverage, and the ordinary
dependency-resolved Document Audit, Lexical Rollup, and i-before-e binaries
built and produced their expected outputs. Its generated binaries were smaller
by 0.11%–0.14% (31,552–42,032 bytes), but that is not a performance result.

It is not retained. Under the same guarded, interleaved 50-launch process-wall
loops, the candidate was effectively neutral for Document Audit and Lexical
Rollup and regressed i-before-e:

| Workload | Baseline average | Candidate average | Candidate change |
| --- | ---: | ---: | ---: |
| Document Audit | 68,130,221 ns | 67,981,341 ns | -0.22% |
| Lexical Rollup | 106,001,159 ns | 105,942,593 ns | -0.06% |
| i-before-e | 70,401,464 ns | 72,156,366 ns | +2.49% |

The 60-launch bootstrap profiles agree with the broad decision. Document
Audit's new reflective decoder is 50 ms/33.3% of a 150 ms bootstrap merge,
against JSON's prior 60 ms/30.0% of 200 ms; Lexical Rollup is 10 ms/8.3%,
essentially unchanged from JSON's 10 ms/7.1%; i-before-e receives no decoder
sample, but its process-wall guard regresses. The candidate was therefore
reverted before the unrelated numeric control: a representation replacement
that loses one independent application and has no material wins in the other
two does not clear the broad bar.

No further decoder representation work is authorized now. Both generic
alternatives—expanded Go constructors and compact reflection decoding—failed
for different broad reasons. The allocation-statistics hook is retained.

### Rejected lazy method-cache construction candidate

The one remaining startup candidate deferred the eager equality-dispatch,
bound-method, and two method-scope cache maps until their existing first-store
paths. It changed all interpreter modes uniformly and passed focused cache,
rollover, invalidation, scope, and equality regressions. It is not retained.

Two guarded, interleaved 50-launch comparisons gave only small or mixed
process-wall movement: Document Audit improved 1.30% then 0.53%, Lexical
Rollup improved 0.56% then regressed 0.33%, and i-before-e regressed 0.51%
then improved 0.62% when launch order was reversed. More importantly, the
five-run pinned persistent member-call guard regressed from `732,662,013` to
`761,385,355 ns/op` (+3.9%) with identical `B/op` and `allocs/op`. The former
cache capacities therefore remain. A sub-percent startup signal cannot justify
a material long-running generic interpreter regression.

Do not retry cache-allocation timing or lower the same cache capacities without
new cross-workload evidence. The next work should return to refreshed,
bounded bytecode VM profiles across the independent text/iterator/numeric
workloads and require a material shared primitive or dispatch descendant before
another runtime candidate. No package-, collection-, corpus-, or
benchmark-specific path is allowed.
