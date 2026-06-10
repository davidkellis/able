#!/usr/bin/env python3
"""Fast tests for mode-aware benchmark CPU and executor contracts."""

from __future__ import annotations

import importlib.machinery
import importlib.util
import subprocess
import sys
import unittest
from pathlib import Path
from types import ModuleType


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
CATALOG = SCRIPT_DIR / "bench_external_catalog.sh"
sys.path.insert(0, str(SCRIPT_DIR))

import bench_execution_contract as contracts  # noqa: E402


def load_script(name: str) -> ModuleType:
    loader = importlib.machinery.SourceFileLoader(name, str(SCRIPT_DIR / name))
    spec = importlib.util.spec_from_loader(name, loader)
    if spec is None:
        raise RuntimeError(f"cannot load {name}")
    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)
    return module


scoreboard = load_script("bench_external_scoreboard")


def contract(mode: str, budget: int, affinity: str, executor: str) -> dict[str, object]:
    return {
        "mode": mode,
        "logical_cpu_budget": budget,
        "cpu_affinity": affinity,
        "executor_policy": executor,
    }


class CatalogExecutionContractTests(unittest.TestCase):
    def catalog(self, *arguments: str, ok: bool = True) -> subprocess.CompletedProcess[str]:
        result = subprocess.run(
            [str(CATALOG), *arguments],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        if ok:
            self.assertEqual(result.returncode, 0, result.stderr)
        return result

    def test_compiled_parallel_and_serial_budgets_are_explicit(self) -> None:
        result = self.catalog(
            "execution-contracts", "compiled", "fib", "binarytrees", "channel_rollup"
        )
        self.assertEqual(
            result.stdout.splitlines(),
            [
                "fib\tcompiled\t1\tserial",
                "binarytrees\tcompiled\t4\tgoroutine",
                "channel_rollup\tcompiled\t4\tgoroutine",
            ],
        )

    def test_interpreter_budget_stays_single_cpu_with_executor_policy(self) -> None:
        result = self.catalog(
            "execution-contracts", "bytecode", "fib", "binarytrees", "mutex_ledger"
        )
        self.assertEqual(
            result.stdout.splitlines(),
            [
                "fib\tbytecode\t1\tserial",
                "binarytrees\tbytecode\t1\tgoroutine",
                "mutex_ledger\tbytecode\t1\tgoroutine",
            ],
        )

    def test_cpu_pool_resolution_is_ordered_and_bounded(self) -> None:
        self.assertEqual(self.catalog("resolve-cpus", "3-5,8", "1").stdout.strip(), "3")
        self.assertEqual(self.catalog("resolve-cpus", "3-5,8", "4").stdout.strip(), "3,4,5,8")
        result = self.catalog("resolve-cpus", "3-5", "4", ok=False)
        self.assertNotEqual(result.returncode, 0)


class ScoreboardExecutionContractTests(unittest.TestCase):
    def test_valid_contract_is_normalized(self) -> None:
        value = contract("compiled", 4, "3,4,5,8", "goroutine")
        self.assertEqual(scoreboard.declared_execution_contract(value, "test"), value)

    def test_reference_mode_may_differ_when_resources_match(self) -> None:
        self.assertTrue(
            contracts.execution_contracts_compatible(
                contract("treewalker", 1, "3", "serial"),
                contract("bytecode", 1, "3", "serial"),
            )
        )

    def test_row_rejects_reference_resource_mismatch(self) -> None:
        row = {
            "benchmark": "binarytrees",
            "mode": "compiled",
            "able": {"execution_contract": contract("compiled", 4, "3-6", "goroutine")},
            "execution_contract": contract("compiled", 4, "3-6", "goroutine"),
            "comparisons": {
                "go": {
                    "execution_contract": contract("compiled", 1, "3", "goroutine")
                }
            },
        }
        with self.assertRaisesRegex(ValueError, "reference execution contract differs"):
            scoreboard.row_execution_contract(row, "binarytrees/compiled")

    def test_row_rejects_mode_mismatch(self) -> None:
        row = {
            "benchmark": "fib",
            "mode": "bytecode",
            "able": {},
            "execution_contract": contract("compiled", 1, "3", "serial"),
            "comparisons": {},
        }
        with self.assertRaisesRegex(ValueError, "mode does not match"):
            scoreboard.row_execution_contract(row, "fib/bytecode")


if __name__ == "__main__":
    unittest.main()
