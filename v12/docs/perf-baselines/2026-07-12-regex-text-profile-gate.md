# Regex/text shared-leaf profile gate — 2026-07-12

## Decision

Keep no bytecode VM, compiler, runtime, or canonical-stdlib performance
change from this tranche. The new regex application is slow enough to merit
investigation, but its material VM costs do not recur in both an independent
regex workload and a non-regex text control. `execCallOpcode` is common only
as a dispatcher parent; optimizing it without a shared child would repeat the
previous rejected broad experiments.

## Method

All profiles used a single process with `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, and canonical external `able-stdlib`. The bytecode runtime
benchmark loads and warms the program before its own CPU-profiler hook starts,
so the retained samples exclude parsing and bootstrap work.

The full suffix audit and the default `regex_is_match_small` both exceed the
quick-profile budget in an interpreter. Their source now accepts an explicit
`--profile` argument solely for bounded profiling: it changes the input prefix
to 128 items while retaining the same program, public APIs, and control flow.
Default runs retain their former input sizes and verifier output. The suffix
audit's Go, Python, and Ruby counterparts recognize the same optional mode.

Retained steady-state profiles in `v12/interpreters/go/.profiles/`:

- `20260712_regex_suffix_audit_profile_steady_5x.cpu.pprof`
- `20260712_regex_is_match_profile_steady_5x.cpu.pprof`
- `20260712_string_split_join_regex_control_steady_5x.cpu.pprof`

| Workload | Profile setup | Result / sample |
| --- | --- | --- |
| `regex_suffix_audit` | 128-word `--profile`, 5 calls | 20.16s CPU sample; a separate 1-call calibration measured 3,821,488,072 ns/op, 80,335,472 B/op, 700,949 allocs/op. |
| `regex_is_match_small` | 128-input `--profile`, 5 calls | 1,079,147,288 ns/op, 48,989,291 B/op, 517,411 allocs/op; 5.27s CPU sample. |
| `string_split_join_small` | default input, 5 calls | 1,037,644,064 ns/op, 49,415,608 B/op, 555,692 allocs/op; 5.18s CPU sample. |

## Attribution

| Workload | Material descendants | Decision relevance |
| --- | --- | --- |
| Suffix audit | Cached identifier-name resolution 24.65% cumulative, Array-slot member calls 11.16%, scope-entry validation 9.77%. | Builder/NFA results flow through repeated local-name and struct/Array access. |
| Independent regex | `execCallOpcode` 28.65%, call-name 14.61%, cached identifier resolution 11.76%, Array-slot calls 7.59%. | It shares some dispatch with the audit but not the audit's scope-validation magnitude. |
| Split/join control | `execCallOpcode` 27.61%, call-name 18.73%, inline return 17.95%, then typed-pattern/coercion and raw-integer work. | It has no material cached-identifier leaf and its return/type path is different. |

The safe conclusion is that there is no repeated concrete leaf. In particular:

- cached identifier/scope work repeats only in the two regex shapes;
- inline-return/type-match work is material only in split/join; and
- Array-slot transport is weak in the text control.

No regex-specific VM path, named-container shortcut, or source-shape fusion is
appropriate. The missing `able.text.string` import in `regex_nfa.able` was
fixed because it blocked normal `String.chars()` method resolution; it is a
module-visibility correctness repair, not an optimization.

## Next recommendation

Profile the generated compiled runtime for the suffix audit, independent
regex, and split/join control under the same bounded lane. The compiler is
also far from its external reference on the verified application, while this
tranche rules out a bytecode-only candidate. Keep an AOT change only if the
same bridge/value/dispatch leaf repeats in all three generated binaries.
