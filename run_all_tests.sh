#!/usr/bin/env bash

set -euo pipefail

VERSION="v11"
ARGS=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      if [[ $# -lt 2 ]]; then
        echo "Missing value for --version" >&2
        exit 1
      fi
      VERSION="$2"
      shift 2
      ;;
    --version=*)
      VERSION="${1#*=}"
      shift
      ;;
    *)
      ARGS+=("$1")
      shift
      ;;
  esac
done

case "$VERSION" in
  v10)
    TARGET="./v10/run_all_tests.sh"
    ;;
  v11)
    TARGET="./v11/run_all_tests.sh"
    ;;
  *)
    echo "Unknown version '$VERSION'. Expected v10 or v11." >&2
    exit 1
    ;;
esac

if [[ ! -x "$TARGET" ]]; then
  echo "Test runner not found at $TARGET" >&2
  exit 1
fi

exec "$TARGET" "${ARGS[@]:-}"
