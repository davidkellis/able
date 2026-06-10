# Post-Quickening Top-Miss Bytecode Profile Reconciliation

Date: 2026-07-18

## Decision

Complete the bounded K-Nucleotide, Fixed Width 128, Reverse Complement,
Distance Field, and Mandelbrot CPU-profile gate with no retained VM, compiler,
runtime, benchmark, fixture, language, or canonical `able-stdlib` change.

The five profiles reproduce real, large bytecode costs, but they do not expose
one new removable semantic operation in at least three unlike applications.
The exact overlaps are either VM dispatcher parents, previously rejected
slot/carrier and return/frame boundaries, or wrappers whose concrete children
split by application. No candidate advanced to an A/B gate.

## Profile contract

Each application ran once in a separate ordinary bytecode process with the
current post-quickening binary, canonical external stdlib, catalog working
directory and arguments, CPU 0, `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, and a 55-second process cap. Every process completed and
passed its public Ruby verifier. These profiles include normal loading and
startup, but generated-main execution owns 87.5%-95.9% of the samples in all
five, so the reconciliation compares exact descendants below
`runResumable(...)` rather than startup parents.

All five profiled executables have SHA-256
`9c893cf6c03c755305cccc52e40f14c0fdb2e6105ea7ae757415ac9bf879b0d5`
and Go build ID `90188e44e89b96ea07662facec614ecb166b4bb1`.

| Application | Wall | User | CPU samples | GC | Verification |
| --- | ---: | ---: | ---: | ---: | :---: |
| K-Nucleotide | 41.510 s | 41.110 s | 41.120 s | 37 | 1/1 |
| Fixed Width 128 | 7.340 s | 7.170 s | 7.160 s | 46 | 1/1 |
| Reverse Complement | 6.100 s | 5.980 s | 6.000 s | 24 | 1/1 |
| Distance Field | 5.510 s | 5.450 s | 5.400 s | 26 | 1/1 |
| Mandelbrot | 6.050 s | 5.980 s | 5.930 s | 38 | 1/1 |

The machine-readable process records are retained in
`2026-07-18-post-quickening-top-miss-profiles/`. They preserve the execution
contract, sample, verifier, stdout hash, and status for each application.

## Exact cross-application reconciliation

At a 2% cumulative materiality threshold, the exact interpreter-owned
overlaps are:

| Exact operation | K-Nucleotide | Fixed Width | Reverse | Distance | Mandelbrot | Reconciliation |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| `runResumable` flat / cumulative | 6.27% / 95.94% | 3.49% / 87.57% | 5.83% / 87.67% | 6.67% / 95.93% | 16.19% / 92.41% | dispatcher parent; opcode owners differ |
| `execLoadSlotOpcode` cumulative | 4.23% | 5.73% | 18.67% | 5.00% | 7.08% | wrapper over slot-to-stack transport |
| `appendSlotStackValueChecked` flat / cumulative | 1.78% / 3.38% | 0.84% / 3.63% | 1.83% / 18.50% | 2.04% / 3.89% | 4.72% / 6.07% | two generic ordering/carrier trials already failed broad guards |
| `execCallOpcode` cumulative | 30.23% | 49.16% | 36.17% | 44.63% | below 2% | call-name, wide-integer member, array-slot, and static/native descendants split |
| `execBinary` cumulative | 16.05% | 5.17% | below 2% | 8.33% | 25.97% | bitwise, wide-integer, float-region, and float-loop owners split |
| `finishInlineReturn` cumulative | 15.59% | 6.84% | below 2% | 8.52% | below 2% | same three-program return parent whose guard reorder and carrier variants were rejected |

The common loop machinery is measurable but not itself a semantic helper.
Summing the instruction-fetch, loop-test, and opcode-switch source lines gives
approximately 3.19% of K-Nucleotide samples, 1.40% of Fixed Width, 2.83% of
Reverse Complement, 2.04% of Distance Field, and 9.44% of Mandelbrot. Below
that switch, the hot opcodes and value types diverge. This does not authorize
changing the dispatcher or adding an application-shaped fused opcode.

The remaining three-program overlaps are also closed evidence:

- `execCallName(...)` and `tryInlineResolvedCallFromStack(...)` recur in
  K-Nucleotide, Fixed Width, and Distance Field, but the recent raw-cell
  preservation candidate failed the broad collection guard;
- `pushCallFrame(...)`, `popCallFrameFields(...)`, and
  `finishInlineReturn(...)` recur in the same three, but repeated frame and
  slotless-return variants were neutral or regressive; and
- generic stores recur in K-Nucleotide, Reverse Complement, and Mandelbrot,
  while their exact owners are integer/map state, byte-array state, and float
  loop state respectively.

## Allocation decision

No allocation profiles were added. CPU attribution is already decisive:

| Application | `runtime.mallocgc` cumulative | Exact Able owner |
| --- | ---: | --- |
| K-Nucleotide | 2.70% | distributed call/type/map work |
| Fixed Width 128 | 18.85% | checked `UInt128` result construction (84% of sampled allocator callers) |
| Reverse Complement | 8.17% | integer snapshots and primitive byte-array values |
| Distance Field | 14.26% | normalized/materialized floats plus Ratio/native boundaries |
| Mandelbrot | 23.44% | normalized raw-float boxing (98.6% of sampled allocator callers) |

Distance Field and Mandelbrot repeat the already-exhaustively rejected
raw-float boxing family, but only two of the five programs do so. Fixed Width
is a non-primitive nominal wide-integer path, Reverse Complement is integer and
array transport, and K-Nucleotide is not allocator-led. A sampled allocation
profile would add precision to different owners without creating a legal
three-program candidate.

## Verification and cleanup

- 5/5 bounded bytecode processes completed under the 55-second cap.
- 5/5 outputs passed their public verifier, with no timeout or failure.
- All five JSON evidence files pass a structural status/verification check.
- All five profiles used the exact same executable fingerprint.
- Raw profiles, executables, and stdout/stderr captures were removed after
  attribution; only the small machine-readable process records and this
  aggregate remain.
- No WASM work was performed.

## Next recommendation

Run a bounded register-form/basic-block bytecode feasibility census before
making another VM implementation change.

Why: these profiles show that the only new cross-family cost is instruction
dispatch plus repeated slot/stack transport, while the semantic helpers below
it diverge or have already failed broad A/B gates. The remaining product gaps
are too large for another branch reorder. A register-form block can potentially
remove dispatches and transient stack movement across ordinary operations
without changing Able values, primitive semantics, or nominal lowering rules.

What it entails: add temporary, opt-in counters that classify dynamically
executed basic blocks and quantify eliminable `LoadSlot`/`StoreSlot`/stack
transport and dispatch in these five applications plus unlike array,
iterator, text, and nominal controls. Define no benchmark or named-type
opcode. Admit an experimental register-form block only if one general block
shape accounts for a material share in at least three unlike verified
programs; its exits must fall back to the existing bytecode path at calls,
dynamic dispatch, exceptions, yields, and type/identity boundaries. Start with
a correctness-only feature flag, then use preserved baseline/candidate
binaries and repeated order-balanced verifier-backed averages. Stop at the
census if coverage is fragmented or the required representation would reopen
the rejected raw carrier/side-lane designs. Continue to defer WASM.
