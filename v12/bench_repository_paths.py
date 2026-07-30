"""Resolve repository-owned provenance paths across relocated checkouts."""

from __future__ import annotations

from pathlib import Path


def repository_owned_path(raw_path: str | Path, repo_root: Path) -> Path:
    """Resolve a recorded v12 path against the active repository root.

    Historical benchmark reports may contain absolute paths from the checkout
    that produced them. Repository-owned fields should follow the retained
    `v12/...` suffix when those reports are validated from another checkout.
    """

    path = Path(raw_path)
    active_root = repo_root.resolve()
    if not path.is_absolute():
        candidate = (active_root / path).resolve()
    else:
        resolved = path.resolve()
        try:
            resolved.relative_to(active_root)
            return resolved
        except ValueError:
            pass
        parts = path.parts
        try:
            v12_index = parts.index("v12")
        except ValueError:
            return resolved
        candidate = (active_root / Path(*parts[v12_index:])).resolve()
    try:
        candidate.relative_to(active_root)
    except ValueError:
        raise ValueError(
            f"recorded repository path escapes active root {active_root}: {raw_path}"
        ) from None
    return candidate
