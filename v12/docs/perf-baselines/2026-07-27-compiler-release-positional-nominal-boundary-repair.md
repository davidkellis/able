# Compiler release positional-nominal boundary repair

## Decision

**retain-general-boundary-corrections-and-release-sharding**.

The separate compiler release lane exposed two generated-code correctness
failures after named nominal values moved to shared positional storage:

- the Ratio runtime conversion helper read `num` and `den` directly from
  `StructInstanceValue.Fields`; and
- both map-literal spread emitters read the HashMap `handle` directly from
  `StructInstanceValue.Fields`.

Those assumptions are invalid for a nominal value created by the shared
positional encoding. Both paths now use the existing
`__able_struct_named_field_value` representation boundary. This preserves the
native positional carrier while allowing the semantic field name to be
resolved without converting the value to map-backed storage.

The release lane also moved the full fallback audit out of the ordinary
compiler core and into 24 deterministic shards. The former single test took
168.531 seconds. Every supported shard now completes in less than one minute.

This is a correctness and release-gate tranche, not a performance win. No
benchmark-selected optimization or new lowering special case advanced.

## Failures and root causes

### Ratio

Compiler execution-matrix batch 7 stopped the
`06_12_03_stdlib_numeric_ratio_divmod` fixture after `half` and reported:

```text
runtime: ratio value missing num
```

The generated helper checked the semantic nominal identity correctly but then
assumed named-map storage:

```go
inst.Fields["num"]
inst.Fields["den"]
```

Named compiler values can instead be created by
`runtime.NewStructInstancePositionalSized`. The shared named-field accessor
already understands both representations, so the helper now resolves the raw
value through that accessor before applying its existing integer-value checks.

### Map literal spread

After the Ratio repair, execution-matrix batch 8 exposed the same
representation mismatch in:

- `06_01_literals_array_map_inference`; and
- `06_01_bytecode_map_spread`.

Both reported:

```text
runtime: map literal spread expects HashMap value
```

The direct generator rejected a nil `Fields` map and both the direct and IR
emitters read `inst.Fields["handle"]`. They now accept positional nominal
storage and resolve `handle` through the same shared accessor.

The HashMap identity check remains the language/kernel boundary required by
map-literal spread syntax. This is not a HashMap performance fast path and does
not change the shared lowering of ordinary non-primitive nominal types.

## Retained changes

- `generator_render_runtime_numeric.go`
  - Ratio `num` and `den` reads use the shared named-field accessor.
- `generator_collections.go`
  - map spread no longer requires named-map storage;
  - `handle` uses the shared named-field accessor.
- `ir_codegen.go`
  - IR map spread uses the same shared accessor.
- `compiler_array_value_boundary_test.go`
  - a structural guard requires both Ratio fields to use the shared accessor
    and forbids the representation-specific reads.
- `compiler_hashmap_native_test.go`
  - a compiled map-spread guard requires the shared accessor and forbids the
    direct `Fields` read.
- `run_all_tests.sh`
  - the full fallback audit runs as 24 release shards;
  - the monolithic fallback test is excluded from the ordinary core batches.

No runtime, interpreter, bytecode VM, language, canonical stdlib, dependency,
or WASM source changed.

## Verification

### Focused final-code checks

| Check | Result |
| --- | --- |
| Ratio/map-spread/shared positional-accessor structural guards | pass, 1.060s |
| repaired Ratio and two map-spread execution fixtures | pass, 2.417s |
| `go test ./cmd/ablec -count=1` | pass, 5.751s |

### Compiler release lane

Every final-code release matrix passed:

| Matrix | Coverage | Result |
| --- | ---: | --- |
| fallback audit | 24/24 shards | pass; slowest observed 37.915s |
| compiled execution fixtures | 24/24 shards | pass |
| strict dispatch | 24/24 shards | pass |
| interface-lookup bypass | 24/24 shards | pass |
| boundary fallback markers | 24/24 shards | pass; slowest observed 48.317s |

The 32 ordinary compiler batches and their release outlier checks also passed.
The aggregate long batches took 64.810s, 101.562s, 142.406s, and 82.347s;
their constituent tests remain separately bounded. The concurrency,
diagnostics, and dynamic-boundary outliers took 20.374s, 8.015s, and 3.869s.

The full compiled-CLI release lane passed. Its `cmd/able` aggregate took
1260.300s while compiling and executing the generated canonical-stdlib suite.
That command is an intentionally complete release aggregate rather than an
ordinary individual test; process inspection confirmed sustained build and
execution activity rather than a deadlock.

### Performance evidence state

The scorecard itself remains synchronized and complete:

- 119 selected rows and 126 full-status rows;
- five successful Able/reference samples for every selected row;
- 20 retained source reports and 33 retained reference reports; and
- selection identity
  `e6a6ccacc9620f9e1b89e2510cab52a85114ddbf2f41e33abae7c6d8a70241f8`.

The production compiler hash intentionally invalidates 12 of the 21 closure
snapshots. The checked selector now records exactly those 12 invalidations,
all for `scope-content-drift:compiler-production`; nine bytecode-only closures
remain current. No old performance disposition was silently rebased.

## Interpretation

The failures shared one general cause: a generated semantic boundary treated
the runtime's optional named-field map as the nominal representation. Using
the shared accessor keeps the actual compiler representation positional and
native. It does not add boxing, interpreter fallback, or a container-specific
storage rule.

The full fixture matrices are materially broader than the two failing
fixtures: they verify output, strict dispatch, interface lookup, fallback
markers, and compiled execution across the complete current fixture corpus.
The compiled CLI additionally exercises the canonical stdlib through the
production command surface.

The evidence selector remains conservative because it hashes the whole
compiler production scope. Its 12 invalidations do not imply that all 12
families execute either changed path; they require a reach-aware evidence
refresh before their earlier performance dispositions can be called current.

## Next recommendation

Complete a strict generated-code reach audit for the two corrected paths, then
refresh only the selected compiled closure evidence reached by current
applications.

Why: the compiler-scope hash correctly invalidated 12 closures, but source
inspection currently finds Ratio use in `rational_series` and no portable
benchmark source using map-literal spread. A generated-code census is needed
to distinguish real execution reach from helpers that are merely emitted.

What it entails: strictly compile all 63 portable applications under
disk-backed `/var/tmp`; prove every graph remains interpreter-free; classify
calls to the Ratio conversion helper and emitted map-spread path separately
from helper definitions; rerun five-or-more verifier-backed Able and equivalent
Go processes for every reached compiled application; record arithmetic means;
and advance only closures justified by that reach plus the completed release
matrices. If the census finds broader reach, expand the refresh rather than
assuming it away.

Why it is important: this restores a current performance ledger without
rerunning unrelated closed work or laundering a compiler-wide hash change
through stale measurements. Once the ledger is current, audit the remaining
direct generated `StructInstanceValue.Fields` reads so this representation
class cannot recur.
