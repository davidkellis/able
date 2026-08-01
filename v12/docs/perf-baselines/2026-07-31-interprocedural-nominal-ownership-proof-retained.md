# Interprocedural nominal ownership proof retained

Date: 2026-07-31

## Decision

Retain the opt-in interprocedural ownership-transfer analysis and its full
strict-census integration. Retain the fresh-aggregate effect correction that
distinguishes a newly constructed root from a returned alias while continuing
to classify directly embedded old identities as captures.

Retain no storage reuse, carrier selection, generated execution, runtime,
interpreter, VM, stdlib, language, dependency, benchmark, frozen-workspace, or
WASM change. This tranche was explicitly proof-only, so no runtime A/B result
or performance claim is attached to it.

## Proof boundary

The compiler diagnostic computes fresh-successor summaries through typed
direct calls and fixed points. A summary records the nominal input, fresh
result type, and an optional field path when the successor is embedded in a
fresh result. Conditional implementations are accepted only when every branch
proves the same source and result path.

Caller transfer sites are admitted only when:

- the source generation begins at a locally fresh struct literal;
- every resolved callee parameter is read-only and non-escaping;
- direct dispatch has one proven target, or native-interface dispatch has a
  complete implementation set whose members agree on the result path;
- the result unconditionally replaces the source in the same straight-line
  region, either directly or through the proven fresh field path; and
- no retained alias, capture/storage, returned alias, dynamic call,
  parameter-origin identity, or conditional/nonstraight replacement is found.

Function parameters and arbitrary nominal call results do not establish local
ownership. This deliberately blocks reuse when a caller may still retain the
old identity.

The report is opt-in through `ablec --nominal-ownership-json`. It is collected
after generated files are rendered and does not participate in normal
compilation or code generation.

## Focused semantic guards

New compiler guards cover:

- direct local replacement;
- native-interface replacement with two complete implementations;
- fresh successors embedded in an outer result field;
- retained old aliases;
- callee capture/storage;
- conditional/nonstraight replacement;
- parameter-origin identities; and
- diagnostic opt-in isolation from generated Go.

The effect guard suite also verifies that fresh aggregate roots are not return
aliases and that whole nominal values embedded in aggregates remain captures.

## Five-application gate

The five rows from the prior generated-site join all compiled under
`--no-fallbacks`.

| Application | Prior row | Ownership result |
| --- | --- | --- |
| Binary Event Log | `EventRecord` callback | no transfer claim; read-only callback only |
| Concurrent Event Routing | `EventRecord` callback | no transfer claim; read-only callback only |
| Concurrent Graph Visitors | `VisitState` | eligible direct native-interface replacement |
| Concurrent Packet Codecs | `CursorState` | eligible native-interface successor at `cursor` |
| Concurrent Tree Folds | `FoldState` | eligible native-interface successor at `state` |

Packet also exposes one general direct `PacketStats` replacement. The two
EventRecord rows correctly do not become ownership sites: callable
non-mutation alone does not imply that a callback consumes and replaces an
identity.

## Full strict census

All 66 selected compiled applications generated successfully with zero
failures. No generated dependency report contains `pkg/interpreter`.

The census analyzed 34,885 callable instances and reported 434 fresh-successor
summaries. It found ten structural candidate call sites: nine eligible sites
across six unlike applications and one fail-closed site.

| Application / nominal | Shape | Dispatch | Result |
| --- | --- | --- | --- |
| Concurrent Audio Voices `MixState` | direct replacement | direct | eligible |
| Concurrent Audio Voices `PhaseState` | direct replacement | direct | blocked: source not locally fresh |
| Concurrent Graph Visitors `VisitState` | direct replacement | two interface implementations | eligible |
| Concurrent Packet Codecs `CursorState` | embedded `cursor` field | two interface implementations | eligible |
| Concurrent Packet Codecs `PacketStats` | direct replacement | direct | eligible |
| Concurrent Scene Tiles `TileState` | direct replacement | direct | eligible |
| Concurrent Tree Folds `FoldState` | embedded `state` field | two interface implementations | eligible |
| Dependency Wave Validation `WaveTask` | three direct replacements | direct | eligible |

The final totals are seven eligible direct sites, two eligible embedded-field
sites, three eligible interface sites, and one blocked direct site. Repeated
`WaveTask` sites account for three of the nine eligible call sites.

The full disposable census generated 285,336,889 bytes across 7,060,431 Go
lines and 3,316 generated files before deleting every row module. Per-row
generation ranged from 395 ms to 16.081 seconds, with a 4.405-second mean; no
individual generation approached the 60-second limit.

## Verification

- focused nominal-ownership and nominal-effect compiler guards;
- retained caller-owned-result, old-alias, capture, and conditional guards;
- native-interface execution, self-return, imported generic alias, Result, and
  Option guards;
- `go test ./cmd/ablec`;
- `go test ./cmd/able-generated-boundary-census`;
- shell syntax validation for the extended census driver;
- final 66/66 strict census; and
- direct audit of all ten reported sites.

The aggregate disposable census SHA-256 is
`493a0eab29a93517df5f0779839d6dfce7317af7e6d09159923441f93e41ded8`.
Its compact machine-readable companion is
`2026-07-31-interprocedural-nominal-ownership-proof-retained.json`.
The two exact disk-backed task workspaces totaled 808,376 KiB and were removed
after recording compact evidence. The repository Python cache created by the
census contract test and the 8,180 KiB disposable extern-Go cache were also
removed; no Able task directory remains under `/tmp` or `/var/tmp`.

## Next recommendation

Build an opt-in generated execution prototype that consumes these exact proof
facts; keep normal compilation unchanged until the experiment passes the full
gate.

Why: the missing alias/liveness fact is now proven at nine sites across six
unlike applications, including all three owners where the prior generated-only
ceiling reduced allocations and time materially.

What it entails: make ownership facts available before rendering, add one
structural caller-owned storage path for direct and embedded successors, keep
parameter/dynamic/capture/conditional cases on fresh allocation, rerun focused
identity guards and all 66 strict applications, then collect at least five
balanced baseline/candidate/equivalent-Go pairs on Graph, Packet, Tree, and
additional admitted applications.

Why it matters: this is the first production experiment that can remove the
measured nominal allocation wall without crossing into the interpreter,
boxing static values, weakening Able reference identity, or adding
nominal/container/benchmark-specific rules.
