"""Validation shared by benchmark comparison and compact scorecard tooling."""

from __future__ import annotations

from typing import Any


def declared_execution_contract(raw_contract: Any, context: str) -> dict[str, Any]:
    """Validate and normalize one measured CPU/executor contract."""

    if not isinstance(raw_contract, dict):
        raise ValueError(f"{context}: execution_contract must be an object")
    mode = raw_contract.get("mode")
    budget = raw_contract.get("logical_cpu_budget")
    affinity = raw_contract.get("cpu_affinity")
    executor = raw_contract.get("executor_policy")
    if not isinstance(mode, str) or not mode:
        raise ValueError(f"{context}: execution_contract.mode must be a non-empty string")
    if not isinstance(budget, int) or isinstance(budget, bool) or budget <= 0:
        raise ValueError(
            f"{context}: execution_contract.logical_cpu_budget must be a positive integer"
        )
    if not isinstance(affinity, str) or not affinity:
        raise ValueError(
            f"{context}: execution_contract.cpu_affinity must be a non-empty string"
        )
    if executor not in {"serial", "goroutine"}:
        raise ValueError(
            f"{context}: execution_contract.executor_policy must be serial or goroutine"
        )
    return {
        "mode": mode,
        "logical_cpu_budget": budget,
        "cpu_affinity": affinity,
        "executor_policy": executor,
    }


def execution_contracts_compatible(
    measured: dict[str, Any], reference: dict[str, Any]
) -> bool:
    """Return whether two runtimes received the same resource/executor contract."""

    return all(
        measured[field] == reference[field]
        for field in ("logical_cpu_budget", "cpu_affinity", "executor_policy")
    )


def row_execution_contract(row: dict[str, Any], context: str) -> dict[str, Any] | None:
    """Read a fresh row contract while retaining compatibility with legacy reports."""

    raw_row = row.get("execution_contract")
    raw_able = row.get("able", {}).get("execution_contract")
    if raw_row is None and raw_able is None:
        return None
    row_contract = (
        declared_execution_contract(raw_row, context) if raw_row is not None else None
    )
    able_contract = (
        declared_execution_contract(raw_able, f"{context}: able")
        if raw_able is not None
        else None
    )
    if row_contract is not None and able_contract is not None and row_contract != able_contract:
        raise ValueError(f"{context}: row and Able execution contracts differ")
    contract = row_contract or able_contract
    assert contract is not None
    if contract["mode"] != row.get("mode"):
        raise ValueError(f"{context}: execution contract mode does not match row mode")
    for language, comparison in row.get("comparisons", {}).items():
        if not isinstance(comparison, dict) or "execution_contract" not in comparison:
            continue
        reference_contract = declared_execution_contract(
            comparison["execution_contract"], f"{context}/{language}: reference"
        )
        if not execution_contracts_compatible(contract, reference_contract):
            raise ValueError(f"{context}/{language}: reference execution contract differs")
    return contract
