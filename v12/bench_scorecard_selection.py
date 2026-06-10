"""Validate explicit mode-aware external benchmark selection manifests."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path
from typing import Any


KIND = "external-benchmark-selection-manifest"
SCHEMA_VERSION = 1
ALLOWED_MODES = ("compiled", "bytecode")


def is_sha256(value: Any) -> bool:
    return (
        isinstance(value, str)
        and len(value) == 64
        and all(character in "0123456789abcdef" for character in value)
    )


def canonical_manifest(value: Any) -> dict[str, Any]:
    """Return the stable semantic form of one selection manifest."""

    if not isinstance(value, dict) or value.get("kind") != KIND:
        raise ValueError("selection manifest has an invalid kind")
    if value.get("schema_version") != SCHEMA_VERSION:
        raise ValueError(f"selection manifest needs schema_version {SCHEMA_VERSION}")
    raw_modes = value.get("modes")
    if not isinstance(raw_modes, dict) or set(raw_modes) != set(ALLOWED_MODES):
        raise ValueError("selection manifest must list compiled and bytecode modes")

    modes: dict[str, list[str]] = {}
    seen_rows: set[tuple[str, str]] = set()
    for mode in ALLOWED_MODES:
        raw_benchmarks = raw_modes.get(mode)
        if (
            not isinstance(raw_benchmarks, list)
            or not raw_benchmarks
            or any(not isinstance(benchmark, str) or not benchmark for benchmark in raw_benchmarks)
        ):
            raise ValueError(f"selection manifest {mode} entries must be non-empty strings")
        benchmarks = sorted(raw_benchmarks)
        if len(benchmarks) != len(set(benchmarks)):
            raise ValueError(f"selection manifest repeats a {mode} benchmark")
        for benchmark in benchmarks:
            key = (benchmark, mode)
            if key in seen_rows:
                raise ValueError(f"selection manifest repeats {benchmark}/{mode}")
            seen_rows.add(key)
        modes[mode] = benchmarks
    return {"kind": KIND, "schema_version": SCHEMA_VERSION, "modes": modes}


def semantic_sha256(manifest: dict[str, Any]) -> str:
    payload = json.dumps(
        canonical_manifest(manifest), sort_keys=True, separators=(",", ":")
    ).encode("utf-8")
    return hashlib.sha256(payload).hexdigest()


def load_manifest(path: Path) -> dict[str, Any]:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        raise ValueError(f"selection manifest not found: {path}") from None
    except json.JSONDecodeError as error:
        raise ValueError(f"invalid selection manifest JSON in {path}: {error}") from None
    return canonical_manifest(value)


def manifest_keys(manifest: dict[str, Any]) -> set[tuple[str, str]]:
    normalized = canonical_manifest(manifest)
    return {
        (benchmark, mode)
        for mode, benchmarks in normalized["modes"].items()
        for benchmark in benchmarks
    }


def display_path(path: Path, repo_root: Path) -> str:
    try:
        return str(path.resolve().relative_to(repo_root))
    except ValueError:
        return str(path)


def manifest_record(path: Path, repo_root: Path) -> dict[str, Any]:
    manifest = load_manifest(path)
    return {
        **manifest,
        "path": display_path(path, repo_root),
        "source_sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
        "selection_sha256": semantic_sha256(manifest),
        "row_count": len(manifest_keys(manifest)),
    }


def canonical_record(value: Any) -> dict[str, Any]:
    manifest = canonical_manifest(value)
    if not isinstance(value, dict):
        raise ValueError("selection manifest record must be an object")
    path = value.get("path")
    source_sha256 = value.get("source_sha256")
    selection_sha256 = value.get("selection_sha256")
    row_count = value.get("row_count")
    if not isinstance(path, str) or not path:
        raise ValueError("selection manifest record needs a path")
    if not is_sha256(source_sha256):
        raise ValueError("selection manifest record needs a source_sha256")
    if selection_sha256 != semantic_sha256(manifest):
        raise ValueError("selection manifest record selection_sha256 does not match its rows")
    expected_count = len(manifest_keys(manifest))
    if type(row_count) is not int or row_count != expected_count:
        raise ValueError("selection manifest record row_count does not match its rows")
    return {
        **manifest,
        "path": path,
        "source_sha256": source_sha256,
        "selection_sha256": selection_sha256,
        "row_count": row_count,
    }
