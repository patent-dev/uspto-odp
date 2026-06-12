#!/usr/bin/env bash
#
# Verifies that every exported Client method has a per-endpoint integration test
# named TestIntegration<Method> in a //go:build integration test file. This is the
# machine-checkable form of the "one targeted integration test per endpoint"
# requirement (see README_STANDARD.md). Run from the module root: make check-integration
#
# Methods that are not endpoints (helpers like HTTPClient) are listed, one name
# per line, in .integration-coverage-ignore.
set -euo pipefail

ignore_file=".integration-coverage-ignore"

# Exported *Client methods in the ROOT package only (the hand-written wrapper) -
# not generated/ (oapi-codegen emits its own Client type) or demo/ (own module).
methods=$(for f in *.go; do
  [[ "$f" == *_test.go ]] && continue
  grep -hoE '^func \([a-z_]+ \*Client\) [A-Z][A-Za-z0-9]*' "$f" 2>/dev/null
done | sed -E 's/^func \([a-z_]+ \*Client\) //' | sort -u)

# TestIntegration<Name> funcs in root-package //go:build integration test files.
# The `|| true` keeps a non-matching final file from tripping `set -e` inside the
# command substitution (the loop body's exit status leaks out otherwise).
tests=$(for f in *_test.go; do
  [ -f "$f" ] || continue
  grep -qE '^//go:build integration' "$f" && grep -hoE '^func TestIntegration[A-Za-z0-9]+' "$f" || true
done 2>/dev/null | sed -E 's/^func TestIntegration//' | sort -u)

ignored=""
# `|| true` keeps an all-comment/empty ignore file (grep -v matches nothing,
# exit 1) from tripping `set -e` inside the command substitution.
[ -f "$ignore_file" ] && ignored=$( (grep -vE '^[[:space:]]*(#|$)' "$ignore_file" || true) | sort -u)

covered=0
missing=0
for m in $methods; do
  printf '%s\n' "$ignored" | grep -qxF "$m" && continue
  if printf '%s\n' "$tests" | grep -qxF "$m"; then
    covered=$((covered + 1))
  else
    echo "MISSING: TestIntegration$m (no integration test for Client.$m)"
    missing=$((missing + 1))
  fi
done

echo "integration coverage: ${covered} covered, ${missing} missing"
[ "$missing" -eq 0 ]
