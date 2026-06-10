"""Detect scorecard input drift before a no-input current rewrite."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any


def current_source_drift(scoreboard: dict[str, Any], current_path: Path) -> list[str]:
    """Return rows whose prior source or verifier/input contract changed.

    A bare scoreboard rewrite must not relabel old measurements after an
    application, contract, or foreign-reference source edit. A refresh
    supplies explicit new source reports, so it is the only normal path that
    may promote a changed fingerprint.
    """

    try:
        current = json.loads(current_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return []
    old_rows = current.get("rows") if isinstance(current, dict) else None
    if not isinstance(old_rows, list):
        return []
    old_sources: dict[tuple[str, str], dict[str, Any]] = {}
    for row in old_rows:
        if not isinstance(row, dict):
            continue
        benchmark = row.get("benchmark")
        mode = row.get("mode")
        source = row.get("able_source")
        if isinstance(benchmark, str) and isinstance(mode, str) and isinstance(source, dict):
            old_sources[(benchmark, mode)] = source

    drifted: list[str] = []
    for row in scoreboard["rows"]:
        key = (row["benchmark"], row["mode"])
        old_source = old_sources.get(key)
        if old_source is not None and old_source.get("sha256") != row["able_source"]["sha256"]:
            drifted.append(f"{key[0]}/{key[1]}")
        old_row = next(
            (
                candidate
                for candidate in old_rows
                if isinstance(candidate, dict)
                and candidate.get("benchmark") == key[0]
                and candidate.get("mode") == key[1]
            ),
            None,
        )
        if not isinstance(old_row, dict):
            continue
        old_contract = old_row.get("benchmark_contract")
        new_contract = row.get("benchmark_contract")
        if (
            isinstance(old_contract, dict)
            and isinstance(new_contract, dict)
            and old_contract.get("sha256") != new_contract.get("sha256")
        ):
            drifted.append(f"{key[0]}/{key[1]}/verifier-input")
        old_comparisons = old_row.get("comparisons")
        if not isinstance(old_comparisons, dict):
            continue
        for language, comparison in row["comparisons"].items():
            if not isinstance(comparison, dict):
                continue
            new_reference = comparison.get("source")
            old_comparison = old_comparisons.get(language)
            old_reference = old_comparison.get("source") if isinstance(old_comparison, dict) else None
            if (
                isinstance(old_reference, dict)
                and isinstance(new_reference, dict)
                and old_reference.get("sha256") != new_reference.get("sha256")
            ):
                drifted.append(f"{key[0]}/{key[1]}/{language}")
    return drifted
