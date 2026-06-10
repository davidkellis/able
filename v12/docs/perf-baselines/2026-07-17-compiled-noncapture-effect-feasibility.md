# Compiled non-capture/effect feasibility gate

Date: 2026-07-17

## Decision

Keep the expanded identity regression test, but build no interprocedural effect
analysis and no loop-carried result-storage optimization.

The opportunity remains material in the previously measured direct recurrence,
but it does not occur in the required three unlike programs. Adding a compiler
analysis for one user fixture plus the related signed/unsigned accumulation
pair would be speculative and would fail the same broad-program admission rule
used for runtime candidates.

No compiler lowering, runtime, interpreter, canonical stdlib, benchmark,
verifier, reference implementation, spec, or WASM behavior changed.

## Existing proof inventory

The audit inspected the active caller-owned result implementation and adjacent
compiler analyses:

- caller-owned result resolution proves fresh small nominal returns and fresh
  tail-call chains;
- imported environment independence computes a conservative function-level
  fixed point over package/global/runtime dependencies;
- closure, iterator, and spawn lowering records lexical captures for their
  generated functions; and
- Go escape analysis decides placement only after generated Go exists.

None provides a parameter-level guarantee that an Able object is not retained.
More importantly, such a fact would still be insufficient without caller-side
alias/liveness proof. The two analyses must be designed together before object
identities may be merged.

## Corpus result

A deliberately broad source census covered active v12 examples, fixtures, and
`../able-stdlib/src`:

| Syntactic shape | Sites |
| --- | ---: |
| `x = call(x, ...)` | 28 |
| `x = x.method(...)` | 64 |
| Total initial sites | 92 |

Most sites are primitive updates, Arrays/Strings, persistent containers,
large aggregate nominals, or a dynamic generic callback. They do not enter the
current small-scalar nominal result family. Rational's eligible results are
temporary rather than loop-carried and are already allocation-free under the
retained distinct caller-owned ABI.

The surviving direct recurrence cohort is:

| Source | Nominal definition | Classification |
| --- | --- | --- |
| nominal recurrence fixture | `RecurrenceState` | direct, unconditional; positive ceiling case |
| signed accumulation | `Int128` | direct chain; related numeric family |
| unsigned accumulation | `UInt128` | direct chain; related numeric family |
| Fixed Width modular/select workers | `UInt128` | conditional adoption/replacement; unsafe |

Signed and unsigned accumulation do not count as unlike applications for this
gate. The cohort therefore supplies only two independent shapes, and one is a
negative guard.

## Strengthened semantic guard

`TestCompilerLoopCarriedNominalPreservesCalleeCaptureAndConditionalCandidate`
adds two missing failure modes to the existing retained-old-result test:

1. `retain_and_advance` stores its input in an Array before returning a fresh
   `State`. Reusing the input pointer as result storage would mutate history.
2. A loop adopts its first fresh candidate as `best`, retains that pointer, and
   rejects later candidates. Reusing the candidate allocation site would mutate
   the still-live best value.

Both `advance` and `retain_and_advance` qualify for existing `_into` variants,
so the test exercises the exact future optimization boundary rather than an
unrelated unsupported call path. Its expected output is:

```text
10 20 12 26 101 203 101 203
```

## Verification

The focused caller-owned and loop-carried compiler group passes:

```text
go test ./pkg/compiler -run 'TestCompilerCallerOwnedNominalResult|TestCompilerLoopCarriedNominal' -count=1 -timeout 60s
```

The run completed in 3.996 seconds. No performance comparison was run because
the cross-program admission gate failed before candidate construction; timing
an ineligible candidate would only optimize the known fixture family.

## Next direction

Reconcile concurrency benchmarking across Binary Trees and unlike concurrent
applications under equal CPU/executor budgets before selecting another compiler
hotspot. Binary Trees currently compares Able launched under the single-P Go
guard with a Go reference that explicitly resets `GOMAXPROCS` to twice the host
CPU count. That configuration can distort both the 5.4x product ratio and the
priority assigned to recursive nominal allocation.

The bounded next tranche should run Able and Go on the same affinity set with
Able's goroutine executor explicit, include at least two unlike Future/channel
applications, retain five-run samples and verifiers, and profile only if a gap
survives the fair configuration. This is a measurement-contract reconciliation,
not a Binary Trees lowering rule.
