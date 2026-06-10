"""Stable compact JSON rendering for the reviewable external scoreboard."""

from __future__ import annotations

import json
from typing import Any


def render_json(scoreboard: dict[str, Any]) -> str:
    """Keep one application row per line so the report stays reviewable."""

    lines = ["{"]
    scalar_keys = [
        "kind",
        "generated_at",
        "measurement_source",
        "targets",
        "sources",
        "benchmarks",
        "modes",
        "languages",
    ]
    if "canonical_stdlib_source_state" in scoreboard:
        scalar_keys.append("canonical_stdlib_source_state")
    if "selection_manifest" in scoreboard:
        scalar_keys.append("selection_manifest")
    for key in scalar_keys:
        rendered = json.dumps(scoreboard[key], separators=(",", ":"), ensure_ascii=False)
        lines.append(f'  "{key}": {rendered},')
    lines.append('  "rows": [')
    for index, row in enumerate(scoreboard["rows"]):
        suffix = "," if index + 1 < len(scoreboard["rows"]) else ""
        lines.append("    " + json.dumps(row, separators=(",", ":"), ensure_ascii=False) + suffix)
    lines.extend(["  ]", "}"])
    return "\n".join(lines) + "\n"
