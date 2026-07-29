#!/usr/bin/env python3
"""Fast contracts for matched Go benchmark toolchains."""

from __future__ import annotations

import unittest
from unittest.mock import patch

from bench_go_toolchain_contract import (
    current_go_toolchain,
    require_matching_go_toolchains,
    validate_selector,
)


class GoToolchainContractTests(unittest.TestCase):
    def test_selector_must_name_an_exact_patch_release(self) -> None:
        validate_selector("go1.26.5")
        for invalid in ("", "auto", "local", "go1.26", "1.26.5"):
            with self.subTest(invalid=invalid), self.assertRaises(ValueError):
                validate_selector(invalid)

    @patch("bench_go_toolchain_contract.subprocess.run")
    def test_resolved_version_is_recorded(self, run) -> None:
        run.return_value.returncode = 0
        run.return_value.stdout = "go version go1.26.5 linux/amd64\n"
        run.return_value.stderr = ""
        self.assertEqual(
            current_go_toolchain("go1.26.5"),
            {
                "selector": "go1.26.5",
                "go_version": "go version go1.26.5 linux/amd64",
            },
        )
        self.assertEqual(run.call_args.kwargs["env"]["GOTOOLCHAIN"], "go1.26.5")

    @patch("bench_go_toolchain_contract.subprocess.run")
    def test_mislabeled_resolved_version_is_rejected(self, run) -> None:
        run.return_value.returncode = 0
        run.return_value.stdout = "go version go1.26.4 linux/amd64\n"
        run.return_value.stderr = ""
        with self.assertRaisesRegex(ValueError, "does not match selector"):
            current_go_toolchain("go1.26.5")
        with self.assertRaisesRegex(ValueError, "does not match expected version"):
            current_go_toolchain(
                expected_version="go version go1.26.6 linux/amd64"
            )

    def test_matching_comparison_contract_passes(self) -> None:
        contract = {
            "selector": "go1.26.5",
            "go_version": "go version go1.26.5 linux/amd64",
        }
        require_matching_go_toolchains(
            contract,
            contract,
            context="fib/compiled/go",
            expected_selector="go1.26.5",
            expected_version="go version go1.26.5 linux/amd64",
        )

    def test_mixed_version_and_selector_are_rejected(self) -> None:
        able = {
            "selector": "go1.26.5",
            "go_version": "go version go1.26.5 linux/amd64",
        }
        with self.assertRaisesRegex(ValueError, "version mismatch"):
            require_matching_go_toolchains(
                able,
                {
                    "selector": "go1.26.4",
                    "go_version": "go version go1.26.4 linux/amd64",
                },
                context="fib/compiled/go",
                expected_selector="go1.26.5",
            )
        with self.assertRaisesRegex(ValueError, "selector mismatch"):
            require_matching_go_toolchains(
                able,
                {
                    "selector": "go1.26.4",
                    "go_version": "go version go1.26.5 linux/amd64",
                },
                context="fib/compiled/go",
                expected_selector="go1.26.5",
            )
        with self.assertRaisesRegex(ValueError, "selector mismatch"):
            require_matching_go_toolchains(
                able,
                {
                    "selector": None,
                    "go_version": "go version go1.26.5 linux/amd64",
                },
                context="fib/compiled/go",
                expected_selector="go1.26.5",
            )


if __name__ == "__main__":
    unittest.main()
