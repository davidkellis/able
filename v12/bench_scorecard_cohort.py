"""Read the durable comparison-source manifest from an aggregate scorecard."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any


def display_path(path: Path, repo_root: Path) -> str:
    try:
        return str(path.resolve().relative_to(repo_root))
    except ValueError:
        return str(path)


def is_sha256(value: Any) -> bool:
    return (
        isinstance(value, str)
        and len(value) == 64
        and all(character in "0123456789abcdef" for character in value)
    )


def canonical_stdlib_source_state(value: Any) -> dict[str, Any]:
    """Validate the canonical stdlib runtime-source identity in an aggregate."""

    if not isinstance(value, dict) or value.get("kind") != "canonical-stdlib-source-state":
        raise ValueError("canonical stdlib source state has an invalid kind")
    root = value.get("root")
    source_file_count = value.get("source_file_count")
    source_tree_sha256 = value.get("source_tree_sha256")
    git_head = value.get("git_head")
    git_dirty = value.get("git_dirty")
    if not isinstance(root, str) or not root:
        raise ValueError("canonical stdlib source state needs a root")
    if type(source_file_count) is not int or source_file_count < 1:
        raise ValueError("canonical stdlib source state needs a positive source_file_count")
    if not is_sha256(source_tree_sha256):
        raise ValueError("canonical stdlib source state needs a valid source_tree_sha256")
    if git_head is not None and (not isinstance(git_head, str) or len(git_head) != 40):
        raise ValueError("canonical stdlib source state has an invalid git_head")
    if not isinstance(git_dirty, bool):
        raise ValueError("canonical stdlib source state needs git_dirty")
    return {
        "kind": "canonical-stdlib-source-state",
        "root": root,
        "source_file_count": source_file_count,
        "source_tree_sha256": source_tree_sha256,
        "git_head": git_head,
        "git_dirty": git_dirty,
    }


def aggregate_manifest(path: Path, repo_root: Path) -> tuple[list[Path], dict[str, Any] | None]:
    """Read one aggregate's explicit reports and optional stdlib source state."""

    try:
        aggregate = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        raise ValueError(f"aggregate scorecard not found: {path}") from None
    except json.JSONDecodeError as error:
        raise ValueError(f"invalid aggregate scorecard JSON in {path}: {error}") from None
    if not isinstance(aggregate, dict) or aggregate.get("kind") != "external-benchmark-scoreboard":
        raise ValueError(f"{display_path(path, repo_root)}: expected an external-benchmark-scoreboard")
    sources = aggregate.get("sources")
    if not isinstance(sources, list) or not sources:
        raise ValueError(f"{display_path(path, repo_root)}: aggregate scorecard needs a non-empty sources array")

    inputs: list[Path] = []
    seen: set[Path] = set()
    for source in sources:
        if not isinstance(source, dict):
            raise ValueError(f"{display_path(path, repo_root)}: aggregate source must be an object")
        raw_path = source.get("path")
        if not isinstance(raw_path, str) or not raw_path:
            raise ValueError(f"{display_path(path, repo_root)}: aggregate source needs a non-empty path")
        source_path = Path(raw_path)
        if not source_path.is_absolute():
            source_path = repo_root / source_path
        source_identity = source_path.resolve()
        if source_identity in seen:
            raise ValueError(
                f"{display_path(path, repo_root)}: aggregate repeats {display_path(source_path, repo_root)}"
            )
        seen.add(source_identity)
        inputs.append(source_path)
    state = aggregate.get("canonical_stdlib_source_state")
    if state is not None:
        try:
            state = canonical_stdlib_source_state(state)
        except ValueError as error:
            raise ValueError(f"{display_path(path, repo_root)}: {error}") from None
    return inputs, state


def aggregate_source_inputs(path: Path, repo_root: Path) -> list[Path]:
    """Return the explicit comparison-report cohort recorded by an aggregate."""

    return aggregate_manifest(path, repo_root)[0]
