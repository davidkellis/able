#!/usr/bin/env python3
"""Resolve and validate exact Go toolchain contracts for benchmark evidence."""

from __future__ import annotations

import argparse
import os
import re
import subprocess
from typing import Any


EXACT_SELECTOR = re.compile(r"^go[0-9]+\.[0-9]+\.[0-9]+$")


def validate_selector(selector: str) -> None:
    if not EXACT_SELECTOR.fullmatch(selector):
        raise ValueError(
            "Go toolchain selector must name an exact patch release such as go1.26.5"
        )


def current_go_toolchain(
    selector: str | None = None,
    expected_version: str | None = None,
) -> dict[str, str | None]:
    selected = selector or os.environ.get("GOTOOLCHAIN")
    if selector:
        validate_selector(selector)
    environment = os.environ.copy()
    if selector:
        environment["GOTOOLCHAIN"] = selector
    result = subprocess.run(
        ["go", "version"],
        env=environment,
        text=True,
        capture_output=True,
        check=False,
    )
    if result.returncode != 0:
        raise ValueError(result.stderr.strip() or "cannot identify Go toolchain")
    version = result.stdout.strip()
    if selector and not version.startswith(f"go version {selector} "):
        raise ValueError(
            f"resolved Go version does not match selector {selector!r}: {version!r}"
        )
    expected = expected_version or os.environ.get("ABLE_BENCH_EXPECTED_GO_VERSION")
    if expected and version != expected:
        raise ValueError(
            f"resolved Go version does not match expected version: "
            f"expected {expected!r}, got {version!r}"
        )
    return {
        "selector": selected,
        "go_version": version,
    }


def require_matching_go_toolchains(
    able: Any,
    reference: Any,
    *,
    context: str,
    expected_selector: str | None = None,
    expected_version: str | None = None,
) -> None:
    if not isinstance(able, dict) or not isinstance(reference, dict):
        raise ValueError(f"{context}: missing Go toolchain contract")
    able_version = able.get("go_version")
    reference_version = reference.get("go_version")
    if (
        not isinstance(able_version, str)
        or not isinstance(reference_version, str)
        or able_version != reference_version
    ):
        raise ValueError(
            f"{context}: Go toolchain version mismatch: "
            f"Able={able_version!r}, reference={reference_version!r}"
        )
    expected_version = expected_version or os.environ.get(
        "ABLE_BENCH_EXPECTED_GO_VERSION"
    )
    if expected_version and able_version != expected_version:
        raise ValueError(
            f"{context}: Go toolchain version does not match expected version: "
            f"expected={expected_version!r}, actual={able_version!r}"
        )
    expected_selector = expected_selector or os.environ.get("GOTOOLCHAIN")
    able_selector = able.get("selector")
    reference_selector = reference.get("selector")
    if expected_selector is not None:
        selector_mismatch = (
            able_selector != expected_selector
            or reference_selector != expected_selector
        )
    else:
        selector_mismatch = able_selector != reference_selector
    if selector_mismatch:
        raise ValueError(
            f"{context}: Go toolchain selector mismatch: "
            f"Able={able_selector!r}, reference={reference_selector!r}, "
            f"expected={expected_selector!r}"
        )


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--selector", required=True)
    parser.add_argument("--expected-version")
    parser.add_argument("--print-version", action="store_true")
    args = parser.parse_args()
    try:
        contract = current_go_toolchain(args.selector, args.expected_version)
    except ValueError as error:
        parser.error(str(error))
    if args.print_version:
        print(contract["go_version"])


if __name__ == "__main__":
    main()
