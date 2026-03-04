# Able v12 Spec TODOs

This list tracks the remaining v12 items after audit; completed work should be removed.

## Parser gaps
- None currently tracked (cast `as` line breaks fixed 2026-02-06).

## Compiler AOT gaps
- Evaluate whether no-interpreter static runtimes should add runtime-independent alias/constraint revalidation for generic interface dispatch, or keep this permanently compile-time-only and document it as final.

## Stdlib externalization gaps
- Confirm and document canonical stdlib resolution contract end-to-end (`able setup`, cache layout, lockfile pins, and `able override` precedence).
- Clarify collision/error semantics when multiple `name: able` roots are visible through `ABLE_MODULE_PATHS`, lockfile sources, or overrides.

## Compiler AOT performance / dynamic-carrier staged limits
- `runtime.Value` usage categories are now documented in `spec/full_spec_v12.md` under the AOT boundary section.
- Desired end-state: compiled polymorphism lowers primarily to host-native mechanisms (Go interfaces/concrete dispatch/generic specialization), with `runtime.Value` used only for explicit dynamic boundaries and residual non-representable cases.
- Native-lowering mandate: static compiled code should represent nominal/user-defined program values with host-native concrete structures (not interpreter object-model carriers) and should never invoke interpreter execution paths unless entering explicit dynamic features.
- Stage-0 flag scaffolding landed: `--experimental-mono-arrays` and `ABLE_EXPERIMENTAL_MONO_ARRAYS` flow through compiler options; current CLI default is ON with explicit opt-out.
- Stage-1 partial landed behind the flag: Go runtime now has typed array stores (`i32`, `i64`, `bool`, `u8`) and compiler lowering for typed literals/index plus `push/len/get/set` intrinsics when static element type is known.
- Stage-1 boundary coverage now includes explicit dynamic-call mono-array roundtrip fixtures plus nullable/union/interface callback conversion success/failure fixtures under `--experimental-mono-arrays`.
- Stage-1 index optimization landed: array read/write/get/set lowering now keeps native integer indices as native `int` where safe instead of boxing through `bridge.ToInt` + `bridge.AsInt`.
- Stage-1 propagation/cast optimization landed: mono typed index propagation paths now avoid boxing `i32` reads into `runtime.Value` when a native widening cast is semantically safe (e.g., `i32 -> i64`).
- Stage-1 compatibility fixes landed:
  - `Array` struct converters now accept/synchronize raw `*runtime.ArrayValue` carriers at explicit runtime boundaries.
  - Interface-annotated local assignment now enforces interface coercion via `bridge.MatchType`, preserving interface args for compiled dispatch.
- Stage-1 strict sweep status (2026-02-26): compiler strict fixture audits and `TestCompilerDynamicBoundary*` are green.
- Stage-1 perf snapshot (compiled-only, 5-run avg, 2026-02-26, post compatibility fixes): `bench/noop` default `0.062s` / `3.20` GC vs mono `0.060s` / `3.20`; `bench/sieve_count` default `0.072s` / `5.40` vs mono `0.074s` / `5.20`; `bench/sieve_full` default `0.164s` / `23.20` vs mono `0.164s` / `23.00`.
- Staged Go runtime/compiler limit: generic/dynamic paths still rely on dynamic carrier values by design; remaining mono-array rollout work is staged rollout mechanics, observability, and eventual flag-retirement criteria.
- Staged Go compiler limit: interface/existential dispatch wrappers and extern boundary shims still use dynamic carrier values by design; continue reducing avoidable carrier usage only where static specialization is semantically valid.
- Pending workstream: monomorphized container ABI (`Array<T>` native element storage) behind a gated compiler/runtime rollout plan; staged proposal captured in `v12/design/monomorphized-container-abi.md`.
- Pending workstream: broaden native lowering beyond arrays (struct/union/interface-call-site specialization where statically representable) and add regression guards that fail when new static paths regress to dynamic carrier helpers.
