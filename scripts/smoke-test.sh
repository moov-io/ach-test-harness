#!/usr/bin/env bash
# Start a built ach-test-harness binary in an empty directory (outside the
# repository) and check that it serves /live. This finds binaries that need
# files from the repository at run time.
#
# Usage: scripts/smoke-test.sh <path to binary>
set -euo pipefail

if [ $# -ne 1 ] || [ ! -f "$1" ]; then
  echo "usage: $0 <path to binary>" >&2
  exit 2
fi

bin="$(cd "$(dirname "$1")" && pwd)/$(basename "$1")"
dir="$(mktemp -d)"
cd "$dir"

"$bin" > harness.log 2>&1 &
pid=$!
trap 'kill "$pid" 2>/dev/null || true; cd /; rm -rf "$dir"' EXIT

for _ in $(seq 1 30); do
  if curl -fsS http://localhost:3333/live > /dev/null 2>&1; then
    echo "smoke test passed: $bin serves /live from $dir"
    exit 0
  fi
  if ! kill -0 "$pid" 2>/dev/null; then
    echo "smoke test failed: $bin stopped" >&2
    cat harness.log >&2
    exit 1
  fi
  sleep 1
done

echo "smoke test failed: $bin did not serve /live in 30 seconds" >&2
cat harness.log >&2
exit 1
