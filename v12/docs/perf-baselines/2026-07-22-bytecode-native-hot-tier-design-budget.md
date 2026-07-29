# Bytecode native hot-code tier design and target budget

## Decision

**no-go-native-tier-prototype-current-evidence**.

The architecture contract is defined, but no implementation prototype is
admitted. Current evidence has neither a qualifying three-family hot-function
reach class nor a selected portable native backend.

## End-to-end native proxy

Current compiled Able is the only complete semantics-preserving native engine
in the repository, so its full-process time is used as an attainable planning
proxy, not as a promised JIT result or a mathematical bound.

| Measure | Result |
| --- | ---: |
| Common selected applications | 63 |
| Proxy target meets | 34 |
| Proxy target misses | 29 |
| Current bytecode target excess | 221.503684s |
| Proxy target excess | 15.481053s |
| Target excess removed by proxy | 93.01% |

Native-equivalent execution would be transformative, but it is not sufficient
for the product target: 29 of 63 rows still miss when replaced wholesale by
the current compiled engine. Concurrency, regex, text/map, nominal, and several
control-heavy rows therefore also depend on compiler/runtime improvements.

## Known typed-region reach

This model linearly substitutes current compiled time for the observed dynamic
instruction share. Instruction share is not wall-time share, so the result is
only an equal-cost reach-sizing model. It cannot prove a speedup.

| Application | Family | Instruction share | Required fraction to target | Modeled target ratio | Excess removed | Material? |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `monte_carlo_pi` | stochastic-numeric | 41.79% | 47.00% | 1.09x | 88.92% | yes |
| `rms_norm` | float-array | 11.76% | 88.93% | 7.47x | 13.23% | no |
| `fixed_width_128` | wide-numeric | 6.62% | 96.23% | 19.26x | 6.88% | no |
| `future_await_race` | concurrency | 48.20% | 130.20% | 3.11x | 37.02% | yes |

Only Monte Carlo Pi clears the predeclared 25% target-excess reduction gate.
RMS Norm and Fixed Width 128 require roughly 84% and 98% compiled-equivalent
coverage to reach target, far above their known 12% and 7% safe-region shares.
Future Await Race cannot reach target even at whole-application compiled-proxy
speed because its current compiled runtime is itself about 3.03x over the
bytecode interpreter target. Fine-grained region compilation is not admitted.

## Hot-function reach census

The census counts bytecode instructions and program entries, not elapsed time.
Each row is one verifier-backed application run with bootstrap work reset
immediately before `main()`. The compiled-equivalent substitutions remain
equal-cost reach models, not performance measurements.

| Application | Family | Hottest program | Program share | Primitive share | Max primitive span | Hottest-program excess removed | Contract eligible? |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- |
| `fixed_width_128` | wide-numeric | `ordered_select_checksum` | 46.74% | 34.38% | 7 | 48.57% | no |
| `distance_field` | float-numeric | `main` | 56.41% | 35.90% | 4 | 59.77% | no |
| `concurrent_event_routing` | concurrency-text | `split` | 38.50% | 35.16% | 4 | 38.41% | no |
| `word_frequency` | text-map | `split` | 40.23% | 35.09% | 4 | 39.46% | no |
| `array_slice_window` | array-iterator | `rolling_checksum` | 63.49% | 34.71% | 6 | 62.39% | no |
| `reverse_complement` | byte-text | `reverse_complement_fasta` | 53.41% | 43.40% | 4 | 53.01% | no |

Work is coarse enough by source function—the hottest program owns roughly
38% to 63% of instructions—but not by the pointer-free leaf execution unit.
Only about 34% to 43% of all instructions are conservatively primitive-eligible,
and the hottest programs interrupt those operations with runtime effects after
at most four to seven static instructions. Compiling an entire hot function
would therefore require unsupported Go callbacks or dense side-exit/re-entry
traffic. The common candidate class is measured but contract-ineligible.

## Scalar-proof gap census

The retained opt-in census combines lowering-time slot facts with checked
type inference. Counts remain diagnostic and do not alter normal execution.
The table shows the largest proven scalar gap in each verifier-backed row.

| Application | Family | Largest proven opcode/proof | Instructions | Share |
| --- | --- | --- | ---: | ---: |
| `fixed_width_128` | wide-numeric | `LoadSlotStructField` / `primitive-field-integer` | 6816906 | 11.28% |
| `distance_field` | float-numeric | `LoadSlot` / `primitive-slot-float` | 22000024 | 28.21% |
| `concurrent_event_routing` | concurrency-text | `LoadSlot` / `primitive-slot-integer` | 1086208 | 5.08% |
| `word_frequency` | text-map | `LoadSlot` / `primitive-slot-integer` | 944384 | 5.37% |
| `array_slice_window` | array-slice | `LoadSlot` / `primitive-slot-integer` | 1020268 | 11.66% |
| `reverse_complement` | bio-text-array | `LoadSlot` / `primitive-slot-integer` | 6033694 | 9.96% |

Proven integer `LoadSlot` traffic repeats in five unlike applications,
while Distance Field instead has 28.21% proven float `LoadSlot` traffic.
A guarded integer-load hint was tested against three unlike workloads and
rejected: every three-run average regressed (0.95%-17.06%). The candidate
added a hot branch but did not remove the existing carrier type switch.
Disposition: Rejected after all three unlike guard workloads regressed; the extra hint branch cost more than the narrower type switch saved.

The completed carrier/consumer census accounts for every proven integer
load in six applications. Raw-i32 values feeding
`JumpIfBinaryCompareFalse` crossed four unlike applications, but a balanced
six-process-per-side direct-bool candidate regressed Array Slice Window
8.66% and Reverse Complement 1.65%. Its largest apparent improvement was
matched by the zero-reach Distance Field control, so the candidate was removed.

## Required execution architecture

- **`pointer-free-native-state`:** Native memory contains only primitive scalars, tags, root-table indices, instruction identities, and exit records; it never contains runtime.Value or other Go pointers.
- **`go-owned-boxed-roots`:** All boxed values and identity-bearing nominal objects remain strongly reachable in a Go-owned root table for the entire native activation.
- **`coarse-entry-atomic-commit`:** Entry snapshots proved primitive inputs; successful exit commits primitive slot updates atomically, while a side exit resumes the ordinary VM at the exact original instruction before any uncommitted effect.
- **`no-native-go-callback`:** Initial native code is leaf code: allocation, calls, dynamic lookup, member/index semantics, boxing, externs, and all other effects side-exit to Go instead of calling Go from native code.
- **`source-exact-errors`:** Overflow, division, cast, type, and guard failures return an original bytecode instruction identity so the ordinary VM produces the canonical source-attached Able error.
- **`structured-unwind`:** Raise, rescue, ensure, rethrow, propagation, break/continue cleanup, iterator cleanup, and return coercion remain ordinary-VM boundaries in the first tier.
- **`cooperative-suspension`:** Spawn, await, yield, Future cancellation, serial scheduling, and goroutine scheduling remain ordinary-VM boundaries; native loops have a bounded backedge poll that side-exits for cancellation or fairness.
- **`extern-abi-independence`:** The tier does not embed Go plugin symbols, Go ABI entry points, runtime.Value layouts, or a training-profile identity; extern calls always occur after an ordinary-VM side exit.
- **`portable-fallback`:** Unsupported operating-system/architecture pairs, disabled executable memory, compilation failure, and code-cache eviction retain byte-for-byte ordinary VM behavior.
- **`bounded-code-cache`:** Code is process-local, content-addressed by program and semantic ABI identities, write-xor-execute, bounded by bytes and entries, and never reclaimed while an activation can execute it.

The central GC/ABI rule is intentionally strict: generated code is leaf code
over pointer-free state. Go retains every boxed or identity-bearing value in a
root table; allocation, calls, errors, effects, and suspension side-exit before
execution. This avoids embedding Go interface layouts or Go heap pointers in
unscanned native memory and keeps the ordinary VM semantically authoritative.

## Backend decision

| Backend | Disposition | Reason |
| --- | --- | --- |
| `generated-go-plugin` | `reject-as-hot-tier-backend` | The current extern path invokes the Go toolchain and buildmode=plugin, while the completed PGO gate records exact-profile/toolchain coupling and regression. It is not a portable low-latency in-process tier. |
| `second-go-ir-executor` | `closed-existing-gates` | A Go-level typed-region executor and whole-function register executor already failed broad wall-time gates and do not remove host-language dispatch. |
| `custom-machine-code-emitters` | `defer-until-reach-gate` | No machine-code backend or stable trampoline exists in the repository; per-architecture emitters would add the largest implementation and security burden. |
| `external-c-abi-jit-library` | `defer-until-reach-gate-and-adr` | A C-ABI JIT can satisfy pointer-free leaf execution, but the repository has no selected dependency, distribution policy, compile-latency evidence, or supported-platform decision. |

No backend is selected. Generated Go plugins repeat the current exact-toolchain
and deployment constraints; another Go executor repeats rejected mechanisms;
and choosing custom emitters or an external C-ABI JIT before proving coarse
function reach would commit substantial portability and security work without
an admitted performance unit.

## Prototype admission gate

A prototype requires at least 3 unlike applications, at least 25% modeled target-excess reduction in every row, and a reviewed backend ADR. Current qualifying hot-function
classes: 0; selected backends: 0. Therefore the prototype remains closed.

## Next recommendation

Complete **bytecode-six-application-clean-cpu-owner-refresh**.

Refresh the broad clean CPU owner matrix and choose the largest exact non-closed VM leaf that is material in at least three unlike applications.

Collect:

- bounded runtime-only profiles for Fixed Width 128, Distance Field, Concurrent Event Routing, Word Frequency, Array Slice Window, and Reverse Complement.
- exact source and call-path attribution for owners that repeat across at least three unlike families.
- closure filtering against retained and broadly rejected mechanisms.
- repeated averaged A/B wall-time guards plus a zero-reach control for any admitted candidate.

Admission: Implement a candidate only when one exact non-closed VM leaf is CPU-material in at least three unlike applications and the same removable operation improves repeated averages without regressing established guards.

Why this is next: carrier, consumer, operation, operand-role, trace, and clean
CPU attribution close the shared `CallMemberArraySlot` lead. Its 2,241,334
common push values already take the monomorphic `u8` path, and push CPU is
material only in Reverse Complement. A refreshed broad owner matrix is needed
before another execution candidate can satisfy the three-family gate.

This remains non-WASM work and forbids benchmark, application, named nominal,
and container-specific compilation paths.

## Reproduction

```sh
python3 v12/bench_bytecode_native_hot_tier_budget_test.py
v12/bench_bytecode_native_hot_tier_budget --check
```
