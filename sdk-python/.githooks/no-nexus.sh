#! /usr/bin/env bash
set -euo pipefail

# Ensure poetry.lock does not contain Indeed Nexus PyPI source blocks.
# Runs against the sdk-python package directory (script lives in .githooks/).

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

SED=(sed)
if command -v gsed >/dev/null 2>&1; then
  SED=(gsed)
fi

"${SED[@]}" -i'' \
    -e '/./{H;$!d}' \
    -e 'x' \
    -e 's|\[package.source\]\ntype\s*=\s*\"legacy\"\nurl\s*=\s*\"https://nexus.corp.indeed.com/repository/pypi/simple\"\nreference\s*=\s*\"nexus\"||' \
    poetry.lock

"${SED[@]}" -i'' \
    -e '1{/^\s*$/d}' \
    poetry.lock

"${SED[@]}" -i'' \
    -e '/^\s*$/N;/^\s*\n$/D' \
    poetry.lock

if git diff --exit-code -- poetry.lock >/dev/null 2>&1; then
  exit 0
fi

# Only fail if we actually removed a Nexus source block.
if git diff -- poetry.lock | grep -q 'nexus.corp.indeed.com'; then
  exit 1
fi

exit 0
