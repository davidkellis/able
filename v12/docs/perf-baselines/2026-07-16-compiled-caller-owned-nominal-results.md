# Compiled caller-owned nominal results

Date: 2026-07-16

## Decision

Retain a generic internal caller-owned result ABI for statically compiled
functions that are proven to return a fresh small nominal value.

Eligible results are ordinary nominal structs occupying at most two Go words
of supported scalar fields. The proof accepts a fresh struct literal or a
tail-call chain leading to another proven-fresh result. Any explicit Able
`return` inside the body currently excludes the function, making early returns
conservative. The proof contains no package, application, benchmark,
`Int128`, `UInt128`, `Rational`, or other nominal name.

Static callers use an internal `_into` entry and provide distinct addressable
storage. Tail calls forward that storage through the proven-fresh chain.
Ordinary compiled entries and runtime/dynamic/interface wrappers retain the
existing pointer-return ABI. Returning `self`, a parameter, or another
existing object is not eligible and continues to return the original pointer.
This preserves Able reference identity and alias-visible mutation.

Go escape analysis remains authoritative. A local result that is stored in an
aggregate or kept beyond its lexical lifetime is promoted to the heap; a
non-escaping transient stays in caller storage. Thus the optimization removes
only allocations Go proves unnecessary and does not pool or reuse
identity-bearing Able values.

## Exact allocation results

Preserved baseline and candidate strict no-fallback binaries received exact
allocation-only main-phase runs under `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, and a 55-second process guard.

| Workload | Baseline bytes / allocations | Candidate bytes / allocations | Allocation change |
| --- | ---: | ---: | ---: |
| signed accumulation | 16,000,888 / 1,000,022 | 3,200,688 / 200,016 | -80.0% |
| unsigned accumulation | 16,000,360 / 1,000,013 | 3,200,160 / 200,007 | -80.0% |
| Fixed Width 128 | 35,536,224 / 2,220,986 | 35,536,192 / 2,220,984 | neutral |
| Rational Series | 4,800,272 / 300,016 | 256 / 15 | -100.0% rounded |

Signed and unsigned accumulation retain one result across each loop
iteration, so Go still promotes 200,000 final-per-iteration slots. Rational
does not retain its transients and loses essentially the entire nominal
allocation wall. Fixed Width retains its result pointers across its hot loops;
all caller slots therefore still escape, correctly producing no material
allocation change.

All outputs retain their established SHA-256 values. Fixed Width and Rational
also pass their external Ruby verifiers.

## Repeated wall-time gate

Each primary workload received 25 independent alternating baseline/candidate
pairs with launch order reversal. Values below are arithmetic process means,
as required for this variable workstation.

| Workload | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: |
| signed accumulation | 126.701 ms | 95.608 ms | -24.54% |
| unsigned accumulation | 111.650 ms | 87.433 ms | -21.69% |
| Fixed Width 128 | 181.484 ms | 182.737 ms | +0.69% |
| Rational Series | 105.451 ms | 101.002 ms | -4.22% |

The short Rational process is increasingly startup-dominated after its main
allocation work disappears, explaining why the wall improvement is smaller
than the allocation reduction. Fixed Width is neutral and does not pay a
material penalty when escape analysis rejects stack placement.

## Generality and semantic guards

- An arbitrary user-defined two-`i64` `Pair` receives the internal ABI, proving
  selection is structural rather than tied to the numeric stdlib definitions.
- Aliasing a returned `Pair` and mutating either reference updates the same
  object.
- Storing a returned `Pair` in an Array preserves identity and mutation through
  the aggregate; Go promotes that slot rather than reusing it.
- A function returning an existing parameter and a function containing an
  explicit early return do not receive `_into` variants.
- Dynamic wrappers continue to call the ordinary pointer-return entry.
- A proven-fresh function that raises before construction propagates control;
  rescue behavior remains correct.
- The same controls pass with the fixed execution-context ABI enabled.
- Existing struct mutation/return, Array alias lifetime, interface callable
  return, raised-string rescue, and native error-union tests pass.
- Focused wide-runtime tests and `go build ./cmd/able ./cmd/ablec` pass.

The unrelated `sum_u32_small` guard first measured candidate +2.91% over 25
pairs, then -0.32% over an independent 50 reversed-order pairs. The combined
75-pair means are baseline 85.057 ms and candidate 85.707 ms (+0.76%), within
the observed startup noise, with identical hashes.

Three alternating full Binary Trees pairs improve from 29.6760 s to 28.5103 s
(-3.93%). A separate candidate run passes the external verifier with SHA-256
`341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1`.
`Node` is not eligible because its nullable nominal fields are not scalar, so
this is an important no-regression guard rather than evidence for a tree rule.

Generated binary size is unchanged for narrow-u32 and Binary Trees, and grows
by only about 17-34 KiB in the wide-numeric programs that actually emit result
variants.

## Implementation hygiene

Return compilation was moved from the oversized generator source into
`generator_returns.go`; all touched source files remain below one thousand
lines. Profiling hooks and temporary analysis counters were not added to the
runtime. No canonical stdlib, benchmark, verifier, reference implementation,
spec, bytecode VM, or WASM source changed.

## Next recommendation

Reconcile the remaining loop-carried nominal result escape across signed
accumulation, unsigned accumulation, and Fixed Width before attempting result
slot reuse.

Why: the retained ABI removed every nonescaping Rational result and 80% of the
accumulation allocations, but exact escape analysis shows that loop-carried
result pointers still force one allocation per iteration. Fixed Width is the
critical guard: all of its nominal results remain loop-carried, so an unsafe
single-slot reuse would appear fast while changing observable identity if an
older result were aliased.

What it entails: classify loop-carried result bindings for old-value uses,
alias creation, aggregate/environment capture, and dynamic escape. Add an
independent user-defined nominal recurrence workload because the signed and
unsigned accumulation fixtures are related; advance only a generic two-slot or
lifetime-disjoint storage rule whose proof applies to at least three unlike
programs and at least two nominal definitions. Test an intentional
retained-old-result alias so storage reuse cannot silently mutate an earlier
value. If that proof does not repeat broadly, close this branch and select the
largest remaining non-wide compiled scorecard miss. Keep bytecode work in the
queue and continue to defer WASM.
