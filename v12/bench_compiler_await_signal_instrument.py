#!/usr/bin/env python3
"""Build a Go overlay that measures generated Await signal lifecycle."""

from __future__ import annotations

import argparse
import json
from pathlib import Path


def replace_once(source: str, old: str, new: str, label: str) -> str:
    count = source.count(old)
    if count != 1:
        raise ValueError(f"{label}: expected one anchor, found {count}")
    return source.replace(old, new)


DECLARATIONS = r"""
var __able_await_signal_diag struct {
	statesInitialized atomic.Uint64
	payloadsWithState atomic.Uint64
	channelsCreated   atomic.Uint64
	beginSuccess      atomic.Uint64
	beginReentry      atomic.Uint64
	rearms            atomic.Uint64
	clearWaiting      atomic.Uint64
	signals           atomic.Uint64
	signalsWaiting    atomic.Uint64
	signalsNotWaiting atomic.Uint64
	signalsEnqueued   atomic.Uint64
	signalsDropped    atomic.Uint64
	waitCycles        atomic.Uint64
	serialWaits       atomic.Uint64
	goroutineWaits    atomic.Uint64
	cancelledWaits    atomic.Uint64
	overlapEvents     atomic.Uint64
	maxActive         atomic.Uint64
}

var __able_await_signal_payloads sync.Map
var __able_await_signal_active_mu sync.Mutex
var __able_await_signal_active = map[*__able_async_payload]map[*__able_await_state]struct{}{}

func __able_await_signal_record_state(payload *__able_async_payload) {
	__able_await_signal_diag.statesInitialized.Add(1)
	if payload == nil {
		return
	}
	if _, loaded := __able_await_signal_payloads.LoadOrStore(payload, struct{}{}); !loaded {
		__able_await_signal_diag.payloadsWithState.Add(1)
	}
}

func __able_await_signal_record_begin(payload *__able_async_payload, state *__able_await_state) {
	if payload == nil || state == nil {
		return
	}
	__able_await_signal_active_mu.Lock()
	states := __able_await_signal_active[payload]
	if states == nil {
		states = map[*__able_await_state]struct{}{}
		__able_await_signal_active[payload] = states
	}
	if len(states) > 0 {
		__able_await_signal_diag.overlapEvents.Add(1)
	}
	states[state] = struct{}{}
	active := uint64(len(states))
	__able_await_signal_active_mu.Unlock()
	for {
		current := __able_await_signal_diag.maxActive.Load()
		if active <= current || __able_await_signal_diag.maxActive.CompareAndSwap(current, active) {
			break
		}
	}
}

func __able_await_signal_record_end(payload *__able_async_payload, state *__able_await_state) {
	if payload == nil || state == nil {
		return
	}
	__able_await_signal_active_mu.Lock()
	if states := __able_await_signal_active[payload]; states != nil {
		delete(states, state)
		if len(states) == 0 {
			delete(__able_await_signal_active, payload)
		}
	}
	__able_await_signal_active_mu.Unlock()
}

func __able_await_signal_snapshot() map[string]uint64 {
	__able_await_signal_active_mu.Lock()
	active := uint64(0)
	for _, states := range __able_await_signal_active {
		active += uint64(len(states))
	}
	__able_await_signal_active_mu.Unlock()
	return map[string]uint64{
		"states_initialized": __able_await_signal_diag.statesInitialized.Load(),
		"payloads_with_state": __able_await_signal_diag.payloadsWithState.Load(),
		"channels_created": __able_await_signal_diag.channelsCreated.Load(),
		"begin_success": __able_await_signal_diag.beginSuccess.Load(),
		"begin_reentry": __able_await_signal_diag.beginReentry.Load(),
		"rearms": __able_await_signal_diag.rearms.Load(),
		"clear_waiting": __able_await_signal_diag.clearWaiting.Load(),
		"signals": __able_await_signal_diag.signals.Load(),
		"signals_waiting": __able_await_signal_diag.signalsWaiting.Load(),
		"signals_not_waiting": __able_await_signal_diag.signalsNotWaiting.Load(),
		"signals_enqueued": __able_await_signal_diag.signalsEnqueued.Load(),
		"signals_dropped": __able_await_signal_diag.signalsDropped.Load(),
		"wait_cycles": __able_await_signal_diag.waitCycles.Load(),
		"serial_waits": __able_await_signal_diag.serialWaits.Load(),
		"goroutine_waits": __able_await_signal_diag.goroutineWaits.Load(),
		"cancelled_waits": __able_await_signal_diag.cancelledWaits.Load(),
		"overlap_events": __able_await_signal_diag.overlapEvents.Load(),
		"max_active": __able_await_signal_diag.maxActive.Load(),
		"active_at_exit": active,
	}
}

"""


def instrument_compiled(source: str) -> str:
    source = replace_once(
        source,
        "type __able_await_arm_state struct {\n",
        DECLARATIONS + "type __able_await_arm_state struct {\n",
        "declarations",
    )
    source = replace_once(
        source,
        "\twaker       runtime.Value\n",
        "\twaker       runtime.Value\n\tdiagBegins  uint64\n",
        "state begin count",
    )
    channel_make = "make(chan struct{}, 1)"
    state_make = f"\t\ts.waitCh = {channel_make}\n"
    source = replace_once(
        source,
        state_make,
        "\t\t__able_await_signal_diag.channelsCreated.Add(1)\n" + state_make,
        "state channel creation",
    )
    payload_make = f"\t\t\ts.payload.awaitSignal = {channel_make}\n"
    if payload_make in source:
        source = replace_once(
            source,
            payload_make,
            "\t\t\t__able_await_signal_diag.channelsCreated.Add(1)\n"
            + payload_make,
            "payload channel creation",
        )
    source = replace_once(
        source,
        "func (s *__able_await_state) signal() bool {\n"
        "\ts.mu.Lock()\n"
        "\tdefer s.mu.Unlock()\n"
        "\tif !s.waiting {\n"
        "\t\treturn false\n"
        "\t}\n",
        "func (s *__able_await_state) signal() bool {\n"
        "\t__able_await_signal_diag.signals.Add(1)\n"
        "\ts.mu.Lock()\n"
        "\tdefer s.mu.Unlock()\n"
        "\tif !s.waiting {\n"
        "\t\t__able_await_signal_diag.signalsNotWaiting.Add(1)\n"
        "\t\treturn false\n"
        "\t}\n"
        "\t__able_await_signal_diag.signalsWaiting.Add(1)\n",
        "signal state",
    )
    source = replace_once(
        source,
        "\tcase ch <- struct{}{}:\n\tdefault:\n",
        "\tcase ch <- struct{}{}:\n"
        "\t\t__able_await_signal_diag.signalsEnqueued.Add(1)\n"
        "\tdefault:\n"
        "\t\t__able_await_signal_diag.signalsDropped.Add(1)\n",
        "signal enqueue",
    )
    source = replace_once(
        source,
        "\tif s.waiting {\n\t\treturn false\n\t}\n"
        "\ts.waiting = true\n\ts.wakePending = false\n\treturn true\n",
        "\tif s.waiting {\n"
        "\t\t__able_await_signal_diag.beginReentry.Add(1)\n"
        "\t\treturn false\n"
        "\t}\n"
        "\ts.waiting = true\n"
        "\ts.wakePending = false\n"
        "\t__able_await_signal_diag.beginSuccess.Add(1)\n"
        "\tif s.diagBegins > 0 {\n"
        "\t\t__able_await_signal_diag.rearms.Add(1)\n"
        "\t}\n"
        "\ts.diagBegins++\n"
        "\t__able_await_signal_record_begin(s.payload, s)\n"
        "\treturn true\n",
        "begin waiting",
    )
    source = replace_once(
        source,
        "\ts.mu.Lock()\n"
        "\ts.waiting = false\n"
        "\ts.wakePending = false\n"
        "\ts.drainSignalLocked()\n"
        "\ts.mu.Unlock()\n",
        "\ts.mu.Lock()\n"
        "\twasWaiting := s.waiting\n"
        "\ts.waiting = false\n"
        "\ts.wakePending = false\n"
        "\ts.drainSignalLocked()\n"
        "\ts.mu.Unlock()\n"
        "\tif wasWaiting {\n"
        "\t\t__able_await_signal_diag.clearWaiting.Add(1)\n"
        "\t\t__able_await_signal_record_end(s.payload, s)\n"
        "\t}\n",
        "clear waiting",
    )
    legacy_state = (
        "\tstate := &__able_await_state{\n"
        "\t\tarms:       arms,\n"
        "\t\tdefaultArm: defaultArm,\n"
        "\t\tpayload:    payload,\n"
        "\t}\n"
    )
    source = replace_once(
        source,
        legacy_state,
        legacy_state + "\t__able_await_signal_record_state(payload)\n",
        "legacy state initialization",
    )
    context_state = (
        "\tstate := &__able_await_state{arms: arms, defaultArm: defaultArm, "
        "payload: payload}\n"
    )
    source = replace_once(
        source,
        context_state,
        context_state + "\t__able_await_signal_record_state(payload)\n",
        "context state initialization",
    )
    source = source.replace(
        "\t\twaitCh := state.ensureWaitCh()\n",
        "\t\t__able_await_signal_diag.waitCycles.Add(1)\n"
        "\t\twaitCh := state.ensureWaitCh()\n",
    )
    if source.count("__able_await_signal_diag.waitCycles.Add(1)") != 2:
        raise ValueError("wait cycles: expected two await helpers")
    source = source.replace(
        "\t\tif payload.yield != nil && payload.resume != nil {\n",
        "\t\tif payload.yield != nil && payload.resume != nil {\n"
        "\t\t\t__able_await_signal_diag.serialWaits.Add(1)\n",
    )
    if source.count("__able_await_signal_diag.serialWaits.Add(1)") != 2:
        raise ValueError("serial waits: expected two await helpers")
    source = source.replace(
        "\t\t} else {\n\t\t\t// Goroutine-backed tasks have no cooperative resume channel.",
        "\t\t} else {\n"
        "\t\t\t__able_await_signal_diag.goroutineWaits.Add(1)\n"
        "\t\t\t// Goroutine-backed tasks have no cooperative resume channel.",
    )
    source = replace_once(
        source,
        "\t\t} else {\n\t\t\tif exec := __able_future_executor(); exec != nil {\n",
        "\t\t} else {\n"
        "\t\t\t__able_await_signal_diag.goroutineWaits.Add(1)\n"
        "\t\t\tif exec := __able_future_executor(); exec != nil {\n",
        "context goroutine wait",
    )
    source = source.replace(
        "\t\t\tcase <-ctx.Done():\n",
        "\t\t\tcase <-ctx.Done():\n"
        "\t\t\t\t__able_await_signal_diag.cancelledWaits.Add(1)\n",
    ).replace(
        "\t\t\tcase <-waitContext.Done():\n",
        "\t\t\tcase <-waitContext.Done():\n"
        "\t\t\t\t__able_await_signal_diag.cancelledWaits.Add(1)\n",
    )
    return source


def instrument_main(source: str) -> str:
    source = replace_once(
        source,
        '\t"fmt"\n',
        '\t"encoding/json"\n\t"fmt"\n',
        "JSON import",
    )
    return replace_once(
        source,
        "\tos.Exit(exitCode)\n",
        "\tdiagnostic, _ := json.Marshal(__able_await_signal_snapshot())\n"
        '\tfmt.Fprintf(os.Stderr, "__ABLE_AWAIT_SIGNAL__=%s\\n", diagnostic)\n'
        "\tos.Exit(exitCode)\n",
        "main report",
    )


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--module-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--overlay", type=Path, required=True)
    args = parser.parse_args()

    module_dir = args.module_dir.resolve()
    output_dir = args.output_dir.resolve()
    output_dir.mkdir(parents=True, exist_ok=True)
    paths = {
        module_dir / "compiled.go": instrument_compiled(
            (module_dir / "compiled.go").read_text(encoding="utf-8")
        ),
        module_dir / "main.go": instrument_main(
            (module_dir / "main.go").read_text(encoding="utf-8")
        ),
    }
    replacements: dict[str, str] = {}
    for path, contents in paths.items():
        replacement = output_dir / path.name
        replacement.write_text(contents, encoding="utf-8")
        replacements[str(path)] = str(replacement)
    args.overlay.write_text(
        json.dumps({"Replace": replacements}, sort_keys=True, indent=2) + "\n",
        encoding="utf-8",
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
